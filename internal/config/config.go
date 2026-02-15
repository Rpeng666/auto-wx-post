package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config 全局配置结构
type Config struct {
	WeChat  WeChatConfig  `yaml:"wechat"`
	Blog    BlogConfig    `yaml:"blog"`
	Cache   CacheConfig   `yaml:"cache"`
	Image   ImageConfig   `yaml:"image"`
	Publish PublishConfig `yaml:"publish"`
	Log     LogConfig     `yaml:"log"`
}

// WeChatConfig 微信配置
type WeChatConfig struct {
	AppID     string `yaml:"app_id"`
	AppSecret string `yaml:"app_secret"`
}

// BlogConfig 博客配置
type BlogConfig struct {
	SourcePath string `yaml:"source_path"`
	BaseURL    string `yaml:"base_url"`
	Author     string `yaml:"author"`
}

// CacheConfig 缓存配置
type CacheConfig struct {
	StoreFile string `yaml:"store_file"`
}

// ImageConfig 图片配置
type ImageConfig struct {
	TempDir            string `yaml:"temp_dir"`
	PlaceholderService string `yaml:"placeholder_service"`
	DefaultCoverSize   string `yaml:"default_cover_size"`
}

// PublishConfig 发布配置
type PublishConfig struct {
	DaysBefore        int `yaml:"days_before"`
	DaysAfter         int `yaml:"days_after"`
	ConcurrentUploads int `yaml:"concurrent_uploads"`
	MaxRetries        int `yaml:"max_retries"`
	Timeout           int `yaml:"timeout"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level    string `yaml:"level"`
	Format   string `yaml:"format"`
	Output   string `yaml:"output"`
	FilePath string `yaml:"file_path"`
}

var globalConfig *Config

// Load 加载配置文件
func Load(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	// 展开环境变量
	configStr := os.ExpandEnv(string(data))

	var cfg Config
	if err := yaml.Unmarshal([]byte(configStr), &cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	// 验证必需配置
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	globalConfig = &cfg
	return &cfg, nil
}

// Get 获取全局配置
func Get() *Config {
	return globalConfig
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.WeChat.AppID == "" || strings.Contains(c.WeChat.AppID, "${") {
		return fmt.Errorf("WECHAT_APP_ID is required")
	}
	if c.WeChat.AppSecret == "" || strings.Contains(c.WeChat.AppSecret, "${") {
		return fmt.Errorf("WECHAT_APP_SECRET is required")
	}
	if c.Blog.SourcePath == "" {
		return fmt.Errorf("blog.source_path is required")
	}
	return nil
}
