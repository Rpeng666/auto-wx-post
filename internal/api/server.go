package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"auto-wx-post/internal/cache"
	"auto-wx-post/internal/config"
	"auto-wx-post/internal/logger"
	"auto-wx-post/internal/markdown"
	"auto-wx-post/internal/media"
	"auto-wx-post/internal/publisher"
	"auto-wx-post/internal/wechat"
)

// Server implements HTTP REST API server
type Server struct {
	cfg          *config.Config
	wechatClient *wechat.Client
	cacheManager *cache.Manager
	mediaManager *media.Manager
	publisher    *publisher.Publisher
	mdParser     *markdown.Parser
	log          *logger.Logger
	apiKey       string // API authentication key
}

// NewServer creates a new HTTP API server
func NewServer(
	cfg *config.Config,
	wechatClient *wechat.Client,
	cacheManager *cache.Manager,
	mediaManager *media.Manager,
	pub *publisher.Publisher,
	log *logger.Logger,
	apiKey string,
) *Server {
	return &Server{
		cfg:          cfg,
		wechatClient: wechatClient,
		cacheManager: cacheManager,
		mediaManager: mediaManager,
		publisher:    pub,
		mdParser:     markdown.NewParser(),
		log:          log,
		apiKey:       apiKey,
	}
}

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// ListArticlesRequest represents the request for listing articles
type ListArticlesRequest struct {
	StartDate     string `json:"start_date,omitempty"`
	EndDate       string `json:"end_date,omitempty"`
	ShowPublished bool   `json:"show_published,omitempty"`
}

// ParseArticleRequest represents the request for parsing an article
type ParseArticleRequest struct {
	FilePath string `json:"file_path"`
}

// UploadImageRequest represents the request for uploading an image
type UploadImageRequest struct {
	ImagePath string `json:"image_path"`
}

// PublishArticleRequest represents the request for publishing an article
type PublishArticleRequest struct {
	FilePath string `json:"file_path"`
	Force    bool   `json:"force,omitempty"`
}

// ArticleInfo represents article information
type ArticleInfo struct {
	Path      string `json:"path"`
	Title     string `json:"title"`
	Author    string `json:"author"`
	Date      string `json:"date"`
	Subtitle  string `json:"subtitle"`
	Published bool   `json:"published"`
}

// ImageInfo represents uploaded image information
type ImageInfo struct {
	MediaID string `json:"media_id"`
	URL     string `json:"url"`
}

// CacheStatus represents cache status
type CacheStatus struct {
	Size  int `json:"size"`
	Count int `json:"count"`
}

// SetupRoutes sets up HTTP routes
func (s *Server) SetupRoutes() http.Handler {
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/health", s.handleHealth)

	// API routes
	mux.HandleFunc("/api/articles/list", s.authMiddleware(s.handleListArticles))
	mux.HandleFunc("/api/articles/parse", s.authMiddleware(s.handleParseArticle))
	mux.HandleFunc("/api/articles/publish", s.authMiddleware(s.handlePublishArticle))
	mux.HandleFunc("/api/images/upload", s.authMiddleware(s.handleUploadImage))
	mux.HandleFunc("/api/cache/status", s.authMiddleware(s.handleCacheStatus))
	mux.HandleFunc("/api/cache/clear", s.authMiddleware(s.handleClearCache))

	return s.corsMiddleware(s.loggingMiddleware(mux))
}

// authMiddleware checks API key authentication
func (s *Server) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Skip auth if no API key is configured
		if s.apiKey == "" {
			next(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			s.respondError(w, http.StatusUnauthorized, "Missing Authorization header")
			return
		}

		// Support both "Bearer <token>" and "<token>" formats
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token != s.apiKey {
			s.respondError(w, http.StatusUnauthorized, "Invalid API key")
			return
		}

		next(w, r)
	}
}

// corsMiddleware adds CORS headers
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware logs HTTP requests
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		s.log.Info("HTTP request",
			"method", r.Method,
			"path", r.URL.Path,
			"remote", r.RemoteAddr)

		next.ServeHTTP(w, r)

		s.log.Info("HTTP response",
			"method", r.Method,
			"path", r.URL.Path,
			"duration", time.Since(start))
	})
}

// handleHealth handles health check requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	s.respondSuccess(w, map[string]interface{}{
		"status":  "ok",
		"version": "1.0.0",
		"time":    time.Now().Format(time.RFC3339),
	})
}

// handleListArticles handles listing articles
func (s *Server) handleListArticles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req ListArticlesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	articles, err := s.findArticles(req.StartDate, req.EndDate, req.ShowPublished)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to find articles: %v", err))
		return
	}

	s.respondSuccess(w, map[string]interface{}{
		"count":    len(articles),
		"articles": articles,
	})
}

// handleParseArticle handles parsing an article
func (s *Server) handleParseArticle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req ParseArticleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	if req.FilePath == "" {
		s.respondError(w, http.StatusBadRequest, "file_path is required")
		return
	}

	article, err := s.mdParser.ParseFile(req.FilePath)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to parse article: %v", err))
		return
	}

	s.respondSuccess(w, map[string]interface{}{
		"title":        article.Title,
		"author":       article.Author,
		"date":         article.Date,
		"subtitle":     article.Subtitle,
		"gen_cover":    article.GenCover,
		"image_count":  len(article.Images),
		"content_size": len(article.Content),
		"content":      truncateString(article.Content, 500),
	})
}

// handleUploadImage handles uploading an image
func (s *Server) handleUploadImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req UploadImageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	if req.ImagePath == "" {
		s.respondError(w, http.StatusBadRequest, "image_path is required")
		return
	}

	ctx := r.Context()
	imageInfo, err := s.mediaManager.UploadImage(ctx, req.ImagePath)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to upload image: %v", err))
		return
	}

	s.respondSuccess(w, ImageInfo{
		MediaID: imageInfo.MediaID,
		URL:     imageInfo.URL,
	})
}

// handlePublishArticle handles publishing an article
func (s *Server) handlePublishArticle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req PublishArticleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	if req.FilePath == "" {
		s.respondError(w, http.StatusBadRequest, "file_path is required")
		return
	}

	// Check if already published
	if !req.Force {
		published, _ := s.cacheManager.IsFileProcessed(req.FilePath)
		if published {
			s.respondError(w, http.StatusConflict, "Article already published. Use force=true to republish.")
			return
		}
	}

	ctx := r.Context()
	err := s.publisher.PublishArticle(ctx, req.FilePath)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to publish article: %v", err))
		return
	}

	s.respondSuccess(w, map[string]interface{}{
		"file_path": req.FilePath,
		"message":   "Article published successfully",
	})
}

// handleCacheStatus handles getting cache status
func (s *Server) handleCacheStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	size := s.cacheManager.Size()
	s.respondSuccess(w, CacheStatus{
		Size:  size,
		Count: size,
	})
}

// handleClearCache handles clearing cache
func (s *Server) handleClearCache(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	err := s.cacheManager.Clear()
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to clear cache: %v", err))
		return
	}

	s.respondSuccess(w, map[string]interface{}{
		"message": "Cache cleared successfully",
	})
}

// Helper methods

func (s *Server) respondSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{
		Success: true,
		Data:    data,
	})
}

func (s *Server) respondError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(Response{
		Success: false,
		Error:   message,
	})
}

func (s *Server) findArticles(startDate, endDate string, showPublished bool) ([]ArticleInfo, error) {
	var articles []ArticleInfo

	sourcePath := s.cfg.Blog.SourcePath
	err := filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}

		// Parse article to get metadata
		article, err := s.mdParser.ParseFile(path)
		if err != nil {
			s.log.Warn("Failed to parse article", "path", path, "error", err)
			return nil
		}

		// Check date range if specified
		if startDate != "" && article.Date < startDate {
			return nil
		}
		if endDate != "" && article.Date > endDate {
			return nil
		}

		// Check published status
		published, _ := s.cacheManager.IsFileProcessed(path)
		if !showPublished && published {
			return nil
		}

		title := article.Title
		if title == "" {
			title = filepath.Base(path)
		}

		articles = append(articles, ArticleInfo{
			Path:      path,
			Title:     title,
			Author:    article.Author,
			Date:      article.Date,
			Subtitle:  article.Subtitle,
			Published: published,
		})

		return nil
	})

	return articles, err
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
