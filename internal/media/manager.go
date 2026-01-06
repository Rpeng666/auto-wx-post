package media

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sync"

	"auto-wx-post/internal/cache"
	"auto-wx-post/internal/config"
	"auto-wx-post/internal/wechat"
)

// Manager 媒体管理器
type Manager struct {
	client       *wechat.Client
	cacheManager *cache.Manager
	cfg          *config.ImageConfig
	tempFiles    []string
	mutex        sync.Mutex
}

// ImageInfo 图片信息
type ImageInfo struct {
	MediaID string
	URL     string
}

// NewManager 创建媒体管理器
func NewManager(client *wechat.Client, cacheManager *cache.Manager, cfg *config.ImageConfig) (*Manager, error) {
	// 创建临时目录
	if err := os.MkdirAll(cfg.TempDir, 0755); err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}

	return &Manager{
		client:       client,
		cacheManager: cacheManager,
		cfg:          cfg,
		tempFiles:    make([]string, 0),
	}, nil
}

// UploadImage 上传图片 (支持URL和本地路径)
func (m *Manager) UploadImage(ctx context.Context, imagePath string) (*ImageInfo, error) {
	// 检查缓存
	if cached, exists := m.cacheManager.Get(m.imageDigest(imagePath)); exists {
		return m.parseCachedInfo(cached)
	}

	var localPath string
	var err error

	// 判断是URL还是本地路径
	if isURL(imagePath) {
		localPath, err = m.downloadImage(ctx, imagePath)
		if err != nil {
			return nil, fmt.Errorf("download image: %w", err)
		}
		m.trackTempFile(localPath)
	} else {
		localPath = imagePath
	}

	// 上传到微信
	result, err := m.client.UploadPermanentMedia(ctx, wechat.MediaTypeImage, localPath)
	if err != nil {
		return nil, fmt.Errorf("upload to wechat: %w", err)
	}

	info := &ImageInfo{
		MediaID: result.MediaID,
		URL:     result.URL,
	}

	// 缓存结果
	cacheValue := fmt.Sprintf("%s|%s", info.MediaID, info.URL)
	if err := m.cacheManager.Set(m.imageDigest(imagePath), cacheValue); err != nil {
		// 缓存失败不影响主流程
		fmt.Printf("warning: failed to cache image: %v\n", err)
	}

	return info, nil
}

// UploadImagesConcurrently 并发上传多个图片
func (m *Manager) UploadImagesConcurrently(ctx context.Context, imagePaths []string, maxConcurrent int) (map[string]*ImageInfo, error) {
	results := make(map[string]*ImageInfo)
	var resultMutex sync.Mutex
	var wg sync.WaitGroup

	semaphore := make(chan struct{}, maxConcurrent)
	errChan := make(chan error, len(imagePaths))

	for _, imagePath := range imagePaths {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			info, err := m.UploadImage(ctx, path)
			if err != nil {
				errChan <- fmt.Errorf("upload %s: %w", path, err)
				return
			}

			resultMutex.Lock()
			results[path] = info
			resultMutex.Unlock()
		}(imagePath)
	}

	wg.Wait()
	close(errChan)

	// 收集错误
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return results, fmt.Errorf("upload errors: %v", errs)
	}

	return results, nil
}

// downloadImage 下载图片到临时目录
func (m *Manager) downloadImage(ctx context.Context, imgURL string) (string, error) {
	// 解析URL以获取干净的扩展名
	u, err := url.Parse(imgURL)
	if err != nil {
		return "", fmt.Errorf("parse url: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", imgURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("http error: %d", resp.StatusCode)
	}

	// 生成临时文件名
	// 使用完整的imgURL进行哈希，确保不同参数的图片被视为不同文件
	hash := md5.Sum([]byte(imgURL))

	// 使用 path.Ext 获取不带查询参数的扩展名
	ext := path.Ext(u.Path)
	if ext == "" {
		ext = ".png"
	}

	filename := fmt.Sprintf("%x%s", hash, ext)
	tempPath := filepath.Join(m.cfg.TempDir, filename)

	// 保存文件
	file, err := os.Create(tempPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		return "", err
	}

	return tempPath, nil
}

// trackTempFile 记录临时文件
func (m *Manager) trackTempFile(path string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.tempFiles = append(m.tempFiles, path)
}

// Cleanup 清理临时文件
func (m *Manager) Cleanup() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	var errs []error
	for _, path := range m.tempFiles {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			errs = append(errs, err)
		}
	}

	m.tempFiles = m.tempFiles[:0]

	if len(errs) > 0 {
		return fmt.Errorf("cleanup errors: %v", errs)
	}
	return nil
}

// imageDigest 计算图片标识
func (m *Manager) imageDigest(imagePath string) string {
	hash := md5.Sum([]byte(imagePath))
	return fmt.Sprintf("img_%x", hash)
}

// parseCachedInfo 解析缓存信息
func (m *Manager) parseCachedInfo(cached string) (*ImageInfo, error) {
	var mediaID, url string
	if _, err := fmt.Sscanf(cached, "%s|%s", &mediaID, &url); err != nil {
		return nil, fmt.Errorf("parse cached info: %w", err)
	}
	return &ImageInfo{MediaID: mediaID, URL: url}, nil
}

// isURL 判断是否为URL
func isURL(path string) bool {
	return len(path) > 7 && (path[:7] == "http://" || path[:8] == "https://")
}
