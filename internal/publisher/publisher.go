package publisher

import (
	"context"
	"fmt"
	"math/rand"
	"path/filepath"
	"strings"
	"time"

	"auto-wx-post/internal/cache"
	"auto-wx-post/internal/config"
	"auto-wx-post/internal/logger"
	"auto-wx-post/internal/markdown"
	"auto-wx-post/internal/media"
	"auto-wx-post/internal/wechat"
)

// Publisher 发布器
type Publisher struct {
	cfg          *config.Config
	wechatClient *wechat.Client
	cacheManager *cache.Manager
	mediaManager *media.Manager
	mdParser     *markdown.Parser
	mdBeautifier *markdown.Beautifier
	log          *logger.Logger
}

// NewPublisher 创建发布器
func NewPublisher(
	cfg *config.Config,
	wechatClient *wechat.Client,
	cacheManager *cache.Manager,
	mediaManager *media.Manager,
	log *logger.Logger,
) (*Publisher, error) {
	mdParser := markdown.NewParser()
	
	// 尝试加载CSS模板，如果不存在使用默认
	mdBeautifier, err := markdown.NewBeautifier("./assets")
	if err != nil {
		log.Warn("Failed to load CSS templates, using defaults", "error", err)
		mdBeautifier, _ = markdown.NewBeautifier("")
	}

	return &Publisher{
		cfg:          cfg,
		wechatClient: wechatClient,
		cacheManager: cacheManager,
		mediaManager: mediaManager,
		mdParser:     mdParser,
		mdBeautifier: mdBeautifier,
		log:          log,
	}, nil
}

// PublishArticle 发布单篇文章
func (p *Publisher) PublishArticle(ctx context.Context, filePath string) error {
	p.log.Info("Publishing article", "file", filePath)

	// 检查是否已处理
	processed, err := p.cacheManager.IsFileProcessed(filePath)
	if err != nil {
		return fmt.Errorf("check cache: %w", err)
	}
	if processed {
		p.log.Info("Article already published, skipping", "file", filePath)
		return nil
	}

	// 解析Markdown
	article, err := p.mdParser.ParseFile(filePath)
	if err != nil {
		return fmt.Errorf("parse markdown: %w", err)
	}

	// 处理封面图片
	images := article.Images
	if len(images) == 0 || article.GenCover == "true" {
		// 生成随机封面
		seed := p.randomString(10)
		coverURL := fmt.Sprintf("%s/%s/%s",
			p.cfg.Image.PlaceholderService,
			seed,
			p.cfg.Image.DefaultCoverSize)
		images = append([]string{coverURL}, images...)
	}

	// 并发上传图片
	p.log.Info("Uploading images", "count", len(images))
	imageMap, err := p.mediaManager.UploadImagesConcurrently(ctx, images, p.cfg.Publish.ConcurrentUploads)
	if err != nil {
		p.log.Warn("Some images failed to upload", "error", err)
	}

	// 更新内容中的图片URL
	urlMap := make(map[string]string)
	for originalURL, info := range imageMap {
		urlMap[originalURL] = info.URL
	}
	article.Content = p.mdParser.UpdateImageURLs(article.Content, urlMap)

	// 转换为HTML
	htmlContent := p.mdParser.ToHTML(article.Content)
	
	// 美化HTML
	beautifiedHTML, err := p.mdBeautifier.Beautify(htmlContent)
	if err != nil {
		return fmt.Errorf("beautify html: %w", err)
	}

	// 准备文章数据
	var thumbMediaID string
	if len(images) > 0 {
		if info, ok := imageMap[images[0]]; ok {
			thumbMediaID = info.MediaID
		}
	}

	// 获取作者
	author := article.Author
	if author == "" {
		author = p.cfg.Blog.Author
	}

	// 生成文章链接
	filename := filepath.Base(filePath)
	link := strings.TrimSuffix(filename, filepath.Ext(filename))
	sourceURL := p.cfg.Blog.BaseURL + link

	// 创建微信文章
	wechatArticle := wechat.Article{
		Title:            article.Title,
		ThumbMediaID:     thumbMediaID,
		Author:           author,
		Digest:           article.Subtitle,
		ShowCoverPic:     1,
		Content:          beautifiedHTML,
		ContentSourceURL: sourceURL,
	}

	// 添加到草稿箱
	p.log.Info("Adding to WeChat draft", "title", article.Title)
	mediaID, err := p.wechatClient.AddDraft(ctx, []wechat.Article{wechatArticle})
	if err != nil {
		return fmt.Errorf("add draft: %w", err)
	}

	p.log.Info("Successfully published", "media_id", mediaID)

	// 标记为已处理
	if err := p.cacheManager.MarkFileProcessed(filePath); err != nil {
		p.log.Warn("Failed to mark as processed", "error", err)
	}

	return nil
}

// randomString 生成随机字符串
func (p *Publisher) randomString(length int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz"
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	
	b := make([]byte, length)
	for i := range b {
		b[i] = letters[rnd.Intn(len(letters))]
	}
	return string(b)
}
