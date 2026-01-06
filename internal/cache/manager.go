package cache

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Manager 缓存管理器 (线程安全)
type Manager struct {
	store     map[string]*CacheEntry
	storePath string
	mutex     sync.RWMutex
}

// CacheEntry 缓存条目
type CacheEntry struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	Timestamp time.Time `json:"timestamp"`
}

// NewManager 创建缓存管理器
func NewManager(storePath string) (*Manager, error) {
	m := &Manager{
		store:     make(map[string]*CacheEntry),
		storePath: storePath,
	}

	// 尝试加载现有缓存
	if err := m.load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("load cache: %w", err)
	}

	return m, nil
}

// Get 获取缓存
func (m *Manager) Get(key string) (string, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	entry, exists := m.store[key]
	if !exists {
		return "", false
	}
	return entry.Value, true
}

// Set 设置缓存
func (m *Manager) Set(key, value string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.store[key] = &CacheEntry{
		Key:       key,
		Value:     value,
		Timestamp: time.Now(),
	}

	return m.save()
}

// FileDigest 计算文件MD5
func FileDigest(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("calculate md5: %w", err)
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// IsFileProcessed 检查文件是否已处理
func (m *Manager) IsFileProcessed(filePath string) (bool, error) {
	digest, err := FileDigest(filePath)
	if err != nil {
		return false, err
	}

	_, exists := m.Get(digest)
	return exists, nil
}

// MarkFileProcessed 标记文件为已处理
func (m *Manager) MarkFileProcessed(filePath string) error {
	digest, err := FileDigest(filePath)
	if err != nil {
		return err
	}

	value := fmt.Sprintf("%s:%s", filePath, time.Now().Format(time.RFC3339))
	return m.Set(digest, value)
}

// load 从文件加载缓存
func (m *Manager) load() error {
	data, err := os.ReadFile(m.storePath)
	if err != nil {
		return err
	}

	var entries []*CacheEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return fmt.Errorf("unmarshal cache: %w", err)
	}

	for _, entry := range entries {
		m.store[entry.Key] = entry
	}

	return nil
}

// save 保存缓存到文件
func (m *Manager) save() error {
	entries := make([]*CacheEntry, 0, len(m.store))
	for _, entry := range m.store {
		entries = append(entries, entry)
	}

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal cache: %w", err)
	}

	if err := os.WriteFile(m.storePath, data, 0644); err != nil {
		return fmt.Errorf("write cache file: %w", err)
	}

	return nil
}

// Clear 清空缓存
func (m *Manager) Clear() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.store = make(map[string]*CacheEntry)
	return m.save()
}

// Size 获取缓存大小
func (m *Manager) Size() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.store)
}
