// Package storage 本地存储适配器
package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// LocalStorage 本地文件系统存储
type LocalStorage struct {
	basePath string
	baseURL  string
}

// NewLocalStorage 创建本地存储适配器
func NewLocalStorage(cfg map[string]interface{}) (Adapter, error) {
	basePath, _ := cfg["base_path"].(string)
	baseURL, _ := cfg["base_url"].(string)
	if basePath == "" {
		basePath = "./data/uploads"
	}
	if baseURL == "" {
		baseURL = "/uploads"
	}
	// 确保目录存在
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("创建存储目录失败: %w", err)
	}
	return &LocalStorage{basePath: basePath, baseURL: baseURL}, nil
}

func (s *LocalStorage) fullPath(key string) string {
	return filepath.Join(s.basePath, key)
}

// Save 保存文件
func (s *LocalStorage) Save(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error {
	fullPath := s.fullPath(key)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, reader); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}
	return nil
}

// Get 获取文件
func (s *LocalStorage) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	fullPath := s.fullPath(key)
	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrFileNotFound
		}
		return nil, err
	}
	return file, nil
}

// Delete 删除文件
func (s *LocalStorage) Delete(ctx context.Context, key string) error {
	fullPath := s.fullPath(key)
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// GetURL 获取文件 URL
func (s *LocalStorage) GetURL(ctx context.Context, key string) (string, error) {
	return s.baseURL + "/" + strings.TrimPrefix(key, "/"), nil
}

// Exists 检查文件是否存在
func (s *LocalStorage) Exists(ctx context.Context, key string) (bool, error) {
	fullPath := s.fullPath(key)
	_, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetSize 获取文件大小
func (s *LocalStorage) GetSize(ctx context.Context, key string) (int64, error) {
	fullPath := s.fullPath(key)
	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, ErrFileNotFound
		}
		return 0, err
	}
	return info.Size(), nil
}
