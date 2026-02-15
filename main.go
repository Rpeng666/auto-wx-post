package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"auto-wx-post/internal/cache"
	"auto-wx-post/internal/config"
	"auto-wx-post/internal/logger"
	"auto-wx-post/internal/mcp"
	"auto-wx-post/internal/media"
	"auto-wx-post/internal/publisher"
	"auto-wx-post/internal/wechat"
)

var (
	configPath = flag.String("config", "config.yaml", "配置文件路径")
	clearCache = flag.Bool("clear-cache", false, "清空缓存")
	dryRun     = flag.Bool("dry-run", false, "模拟运行(不实际发布)")
	mcpServer  = flag.Bool("mcp", false, "启动 MCP (Model Context Protocol) 服务器")
)

func main() {
	flag.Parse()

	// 加载配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	log, err := logger.NewLogger(&cfg.Log)
	if err != nil {
		fmt.Fprintf(os.Stderr, "初始化日志失败: %v\n", err)
		os.Exit(1)
	}

	log.Info("启动微信公众号自动发布工具")
	startTime := time.Now()

	// 初始化缓存
	cacheManager, err := cache.NewManager(cfg.Cache.StoreFile)
	if err != nil {
		log.Error("初始化缓存失败", "error", err)
		os.Exit(1)
	}

	if *clearCache {
		if err := cacheManager.Clear(); err != nil {
			log.Error("清空缓存失败", "error", err)
			os.Exit(1)
		}
		log.Info("缓存已清空")
		return
	}

	log.Info("缓存加载完成", "size", cacheManager.Size())

	// 初始化微信客户端
	timeout := time.Duration(cfg.Publish.Timeout) * time.Second
	wechatClient := wechat.NewClient(&cfg.WeChat, timeout, cfg.Publish.MaxRetries)

	// 初始化媒体管理器
	mediaManager, err := media.NewManager(wechatClient, cacheManager, &cfg.Image)
	if err != nil {
		log.Error("初始化媒体管理器失败", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := mediaManager.Cleanup(); err != nil {
			log.Warn("清理临时文件失败", "error", err)
		}
	}()

	// 初始化发布器
	pub, err := publisher.NewPublisher(cfg, wechatClient, cacheManager, mediaManager, log)
	if err != nil {
		log.Error("初始化发布器失败", "error", err)
		os.Exit(1)
	}

	// MCP 服务器模式
	if *mcpServer {
		log.Info("启动 MCP 服务器模式")
		mcpSrv := mcp.NewServer(cfg, wechatClient, cacheManager, mediaManager, pub, log)
		handler := mcp.NewHandler(mcpSrv)

		ctx := context.Background()
		if err := handler.Run(ctx); err != nil {
			log.Error("MCP 服务器错误", "error", err)
			os.Exit(1)
		}
		return
	}

	// 扫描并发布文章
	ctx := context.Background()

	// 计算日期范围
	now := time.Now()
	startDate := now.AddDate(0, 0, -cfg.Publish.DaysBefore)
	endDate := now.AddDate(0, 0, cfg.Publish.DaysAfter)

	log.Info("开始扫描文章",
		"start_date", startDate.Format("2006-01-02"),
		"end_date", endDate.Format("2006-01-02"))

	// 遍历日期范围
	successCount := 0
	errorCount := 0
	skipCount := 0

	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")

		// 查找匹配日期的文章
		articles, err := findArticlesByDate(cfg.Blog.SourcePath, dateStr)
		if err != nil {
			log.Error("查找文章失败", "date", dateStr, "error", err)
			continue
		}

		if len(articles) == 0 {
			continue
		}

		log.Info("找到文章", "date", dateStr, "count", len(articles))

		// 发布文章
		for _, article := range articles {
			// 检查是否已处理
			processed, _ := cacheManager.IsFileProcessed(article)
			if processed {
				log.Info("文章已发布，跳过", "file", article)
				skipCount++
				continue
			}

			if *dryRun {
				log.Info("模拟运行模式，跳过实际发布", "file", article)
				continue
			}

			if err := pub.PublishArticle(ctx, article); err != nil {
				log.Error("发布文章失败", "file", article, "error", err)
				errorCount++
			} else {
				successCount++
			}

			// 避免频繁请求
			time.Sleep(2 * time.Second)
		}
	}

	elapsed := time.Since(startTime)
	log.Info("任务完成",
		"duration", elapsed,
		"success", successCount,
		"error", errorCount,
		"skipped", skipCount)
}

// findArticlesByDate 查找指定日期的文章
func findArticlesByDate(sourcePath, dateStr string) ([]string, error) {
	var articles []string

	err := filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 只处理.md文件
		if info.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}

		// 读取文件内容检查日期
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		// 简单检查是否包含日期
		if strings.Contains(string(content), fmt.Sprintf("date: %s", dateStr)) ||
			strings.Contains(string(content), fmt.Sprintf("date: '%s'", dateStr)) ||
			strings.Contains(string(content), fmt.Sprintf("date: \"%s\"", dateStr)) {
			articles = append(articles, path)
		}

		return nil
	})

	return articles, err
}
