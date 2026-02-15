package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"auto-wx-post/internal/cache"
	"auto-wx-post/internal/config"
	"auto-wx-post/internal/logger"
	"auto-wx-post/internal/markdown"
	"auto-wx-post/internal/media"
	"auto-wx-post/internal/publisher"
	"auto-wx-post/internal/wechat"
)

// Server implements an MCP (Model Context Protocol) server
type Server struct {
	cfg          *config.Config
	wechatClient *wechat.Client
	cacheManager *cache.Manager
	mediaManager *media.Manager
	publisher    *publisher.Publisher
	mdParser     *markdown.Parser
	log          *logger.Logger
}

// NewServer creates a new MCP server
func NewServer(
	cfg *config.Config,
	wechatClient *wechat.Client,
	cacheManager *cache.Manager,
	mediaManager *media.Manager,
	pub *publisher.Publisher,
	log *logger.Logger,
) *Server {
	return &Server{
		cfg:          cfg,
		wechatClient: wechatClient,
		cacheManager: cacheManager,
		mediaManager: mediaManager,
		publisher:    pub,
		mdParser:     markdown.NewParser(),
		log:          log,
	}
}

// GetTools returns the list of available tools
func (s *Server) GetTools() []Tool {
	return []Tool{
		{
			Name:        "list_articles",
			Description: "列出指定日期范围内的所有 Markdown 文章。返回文章路径、标题和发布状态。",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"start_date": {
						Type:        "string",
						Description: "开始日期 (YYYY-MM-DD 格式)，留空使用配置的 days_before",
					},
					"end_date": {
						Type:        "string",
						Description: "结束日期 (YYYY-MM-DD 格式)，留空使用配置的 days_after",
					},
					"show_published": {
						Type:        "boolean",
						Description: "是否显示已发布的文章 (默认: false)",
					},
				},
			},
		},
		{
			Name:        "parse_article",
			Description: "解析指定的 Markdown 文章，返回文章元数据（标题、作者、日期、副标题等）和内容预览。",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"file_path": {
						Type:        "string",
						Description: "Markdown 文件的完整路径",
					},
				},
				Required: []string{"file_path"},
			},
		},
		{
			Name:        "upload_image",
			Description: "上传单张图片到微信公众号，返回图片的 media_id 和 URL。支持本地文件路径或远程 URL。",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"image_path": {
						Type:        "string",
						Description: "图片的本地路径或远程 URL",
					},
				},
				Required: []string{"image_path"},
			},
		},
		{
			Name:        "publish_article",
			Description: "发布文章到微信公众号草稿箱。自动处理图片上传、Markdown 转 HTML、样式美化等。",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"file_path": {
						Type:        "string",
						Description: "要发布的 Markdown 文件路径",
					},
					"force": {
						Type:        "boolean",
						Description: "强制发布，即使文章已经发布过 (默认: false)",
					},
				},
				Required: []string{"file_path"},
			},
		},
		{
			Name:        "get_cache_status",
			Description: "获取缓存状态，包括已发布的文章数量和文件列表。",
			InputSchema: InputSchema{
				Type:       "object",
				Properties: map[string]Property{},
			},
		},
		{
			Name:        "clear_cache",
			Description: "清空缓存。警告：这将清除所有已发布文章的记录，可能导致重复发布。",
			InputSchema: InputSchema{
				Type:       "object",
				Properties: map[string]Property{},
			},
		},
	}
}

// CallTool executes a tool with the given parameters
func (s *Server) CallTool(ctx context.Context, params ToolCallParams) (ToolCallResult, error) {
	s.log.Info("MCP tool called", "tool", params.Name)

	switch params.Name {
	case "list_articles":
		return s.handleListArticles(ctx, params.Arguments)
	case "parse_article":
		return s.handleParseArticle(ctx, params.Arguments)
	case "upload_image":
		return s.handleUploadImage(ctx, params.Arguments)
	case "publish_article":
		return s.handlePublishArticle(ctx, params.Arguments)
	case "get_cache_status":
		return s.handleGetCacheStatus(ctx, params.Arguments)
	case "clear_cache":
		return s.handleClearCache(ctx, params.Arguments)
	default:
		return ToolCallResult{
			IsError: true,
			Content: []Content{{
				Type: "text",
				Text: fmt.Sprintf("Unknown tool: %s", params.Name),
			}},
		}, nil
	}
}

func (s *Server) handleListArticles(ctx context.Context, args map[string]interface{}) (ToolCallResult, error) {
	// Parse arguments
	var startDate, endDate string
	showPublished := false

	if val, ok := args["start_date"].(string); ok && val != "" {
		startDate = val
	}
	if val, ok := args["end_date"].(string); ok && val != "" {
		endDate = val
	}
	if val, ok := args["show_published"].(bool); ok {
		showPublished = val
	}

	// Find articles
	articles, err := s.findArticles(startDate, endDate, showPublished)
	if err != nil {
		return ToolCallResult{
			IsError: true,
			Content: []Content{{
				Type: "text",
				Text: fmt.Sprintf("Failed to find articles: %v", err),
			}},
		}, nil
	}

	// Format result
	result := fmt.Sprintf("Found %d article(s):\n\n", len(articles))
	for i, article := range articles {
		status := "未发布"
		if article.Published {
			status = "已发布"
		}
		result += fmt.Sprintf("%d. %s\n   Path: %s\n   Status: %s\n\n",
			i+1, article.Title, article.Path, status)
	}

	return ToolCallResult{
		Content: []Content{{
			Type: "text",
			Text: result,
		}},
	}, nil
}

func (s *Server) handleParseArticle(ctx context.Context, args map[string]interface{}) (ToolCallResult, error) {
	filePath, ok := args["file_path"].(string)
	if !ok || filePath == "" {
		return ToolCallResult{
			IsError: true,
			Content: []Content{{
				Type: "text",
				Text: "file_path is required",
			}},
		}, nil
	}

	// Parse article
	article, err := s.mdParser.ParseFile(filePath)
	if err != nil {
		return ToolCallResult{
			IsError: true,
			Content: []Content{{
				Type: "text",
				Text: fmt.Sprintf("Failed to parse article: %v", err),
			}},
		}, nil
	}

	// Format result
	result := fmt.Sprintf(`Article Details:
Title: %s
Author: %s
Date: %s
Subtitle: %s
Generate Cover: %s
Number of Images: %d

Content Preview (first 500 chars):
%s
`,
		article.Title,
		article.Author,
		article.Date,
		article.Subtitle,
		article.GenCover,
		len(article.Images),
		truncateString(article.Content, 500),
	)

	return ToolCallResult{
		Content: []Content{{
			Type: "text",
			Text: result,
		}},
	}, nil
}

func (s *Server) handleUploadImage(ctx context.Context, args map[string]interface{}) (ToolCallResult, error) {
	imagePath, ok := args["image_path"].(string)
	if !ok || imagePath == "" {
		return ToolCallResult{
			IsError: true,
			Content: []Content{{
				Type: "text",
				Text: "image_path is required",
			}},
		}, nil
	}

	// Upload image
	imageInfo, err := s.mediaManager.UploadImage(ctx, imagePath)
	if err != nil {
		return ToolCallResult{
			IsError: true,
			Content: []Content{{
				Type: "text",
				Text: fmt.Sprintf("Failed to upload image: %v", err),
			}},
		}, nil
	}

	result := fmt.Sprintf(`Image uploaded successfully:
Media ID: %s
URL: %s
`,
		imageInfo.MediaID,
		imageInfo.URL,
	)

	return ToolCallResult{
		Content: []Content{{
			Type: "text",
			Text: result,
		}},
	}, nil
}

func (s *Server) handlePublishArticle(ctx context.Context, args map[string]interface{}) (ToolCallResult, error) {
	filePath, ok := args["file_path"].(string)
	if !ok || filePath == "" {
		return ToolCallResult{
			IsError: true,
			Content: []Content{{
				Type: "text",
				Text: "file_path is required",
			}},
		}, nil
	}

	force := false
	if val, ok := args["force"].(bool); ok {
		force = val
	}

	// Check if already published
	if !force {
		published, _ := s.cacheManager.IsFileProcessed(filePath)
		if published {
			return ToolCallResult{
				Content: []Content{{
					Type: "text",
					Text: "Article already published. Use force=true to republish.",
				}},
			}, nil
		}
	}

	// Publish article
	err := s.publisher.PublishArticle(ctx, filePath)
	if err != nil {
		return ToolCallResult{
			IsError: true,
			Content: []Content{{
				Type: "text",
				Text: fmt.Sprintf("Failed to publish article: %v", err),
			}},
		}, nil
	}

	return ToolCallResult{
		Content: []Content{{
			Type: "text",
			Text: fmt.Sprintf("Article published successfully: %s", filePath),
		}},
	}, nil
}

func (s *Server) handleGetCacheStatus(ctx context.Context, args map[string]interface{}) (ToolCallResult, error) {
	size := s.cacheManager.Size()
	result := fmt.Sprintf("Cache contains %d processed article(s).\n", size)

	return ToolCallResult{
		Content: []Content{{
			Type: "text",
			Text: result,
		}},
	}, nil
}

func (s *Server) handleClearCache(ctx context.Context, args map[string]interface{}) (ToolCallResult, error) {
	err := s.cacheManager.Clear()
	if err != nil {
		return ToolCallResult{
			IsError: true,
			Content: []Content{{
				Type: "text",
				Text: fmt.Sprintf("Failed to clear cache: %v", err),
			}},
		}, nil
	}

	return ToolCallResult{
		Content: []Content{{
			Type: "text",
			Text: "Cache cleared successfully.",
		}},
	}, nil
}

// ArticleInfo holds information about an article
type ArticleInfo struct {
	Path      string
	Title     string
	Published bool
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

// SerializeResult serializes a result to JSON
func SerializeResult(result interface{}) (json.RawMessage, error) {
	return json.Marshal(result)
}
