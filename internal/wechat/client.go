package wechat

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"auto-wx-post/internal/config"
)

// Client 微信API客户端 (单例模式)
type Client struct {
	cfg         *config.WeChatConfig
	httpClient  *http.Client
	token       *Token
	tokenMutex  sync.RWMutex
	retryConfig RetryConfig
}

// Token 访问令牌
type Token struct {
	AccessToken string
	ExpiresAt   time.Time
}

// RetryConfig 重试配置
type RetryConfig struct {
	MaxRetries int
	BaseDelay  time.Duration
}

var (
	clientInstance *Client
	clientOnce     sync.Once
)

// NewClient 创建微信客户端 (单例)
func NewClient(cfg *config.WeChatConfig, timeout time.Duration, maxRetries int) *Client {
	clientOnce.Do(func() {
		clientInstance = &Client{
			cfg: cfg,
			httpClient: &http.Client{
				Timeout: timeout,
			},
			retryConfig: RetryConfig{
				MaxRetries: maxRetries,
				BaseDelay:  time.Second,
			},
		}
	})
	return clientInstance
}

// GetClient 获取客户端实例
func GetClient() *Client {
	return clientInstance
}

// GetAccessToken 获取访问令牌 (自动刷新)
func (c *Client) GetAccessToken(ctx context.Context) (string, error) {
	c.tokenMutex.RLock()
	if c.token != nil && time.Now().Before(c.token.ExpiresAt) {
		token := c.token.AccessToken
		c.tokenMutex.RUnlock()
		return token, nil
	}
	c.tokenMutex.RUnlock()

	// 需要刷新token
	c.tokenMutex.Lock()
	defer c.tokenMutex.Unlock()

	// 双重检查
	if c.token != nil && time.Now().Before(c.token.ExpiresAt) {
		return c.token.AccessToken, nil
	}

	return c.refreshToken(ctx)
}

// refreshToken 刷新访问令牌
func (c *Client) refreshToken(ctx context.Context) (string, error) {
	url := fmt.Sprintf(
		"https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s",
		c.cfg.AppID,
		c.cfg.AppSecret,
	)

	var response struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
	}

	if err := c.doRequestWithRetry(ctx, "GET", url, nil, &response); err != nil {
		return "", fmt.Errorf("fetch access token: %w", err)
	}

	if response.ErrCode != 0 {
		return "", fmt.Errorf("wechat api error: %d - %s", response.ErrCode, response.ErrMsg)
	}

	// 提前5分钟过期，避免边界情况
	expiresAt := time.Now().Add(time.Duration(response.ExpiresIn-300) * time.Second)
	c.token = &Token{
		AccessToken: response.AccessToken,
		ExpiresAt:   expiresAt,
	}

	return c.token.AccessToken, nil
}

// doRequestWithRetry 执行HTTP请求并支持重试
func (c *Client) doRequestWithRetry(ctx context.Context, method, url string, body io.Reader, result interface{}) error {
	var lastErr error

	for i := 0; i <= c.retryConfig.MaxRetries; i++ {
		if i > 0 {
			// 指数退避
			delay := c.retryConfig.BaseDelay * time.Duration(1<<uint(i-1))
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}

		req, err := http.NewRequestWithContext(ctx, method, url, body)
		if err != nil {
			return fmt.Errorf("create request: %w", err)
		}

		if method == "POST" {
			req.Header.Set("Content-Type", "application/json; charset=utf-8")
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		defer resp.Body.Close()
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("server error: %d", resp.StatusCode)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("http error: %d - %s", resp.StatusCode, string(respBody))
		}

		if result != nil {
			if err := json.Unmarshal(respBody, result); err != nil {
				return fmt.Errorf("parse response: %w", err)
			}
		}

		return nil
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

// DoRequest 执行微信API请求 (自动附加token)
func (c *Client) DoRequest(ctx context.Context, method, endpoint string, body io.Reader, result interface{}) error {
	token, err := c.GetAccessToken(ctx)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s?access_token=%s", endpoint, token)
	return c.doRequestWithRetry(ctx, method, url, body, result)
}
