// Package storage 存储适配器接口和实现
package storage

import (
	"context"
	"io"
)

// Adapter 存储适配器接口
type Adapter interface {
	// Save 保存文件
	Save(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error
	// Get 获取文件
	Get(ctx context.Context, key string) (io.ReadCloser, error)
	// Delete 删除文件
	Delete(ctx context.Context, key string) error
	// GetURL 获取公开 URL
	GetURL(ctx context.Context, key string) (string, error)
	// Exists 检查文件是否存在
	Exists(ctx context.Context, key string) (bool, error)
	// GetSize 获取文件大小
	GetSize(ctx context.Context, key string) (int64, error)
}

// AdapterFactory 适配器工厂类型
type AdapterFactory func(cfg map[string]interface{}) (Adapter, error)

// 注册的适配器工厂
var adapterFactories = map[string]AdapterFactory{
	"local": NewLocalStorage,
}

// RegisterAdapter 注册存储适配器
func RegisterAdapter(name string, factory AdapterFactory) {
	adapterFactories[name] = factory
}

// CreateAdapter 创建存储适配器
func CreateAdapter(name string, cfg map[string]interface{}) (Adapter, error) {
	factory, ok := adapterFactories[name]
	if !ok {
		return nil, ErrAdapterNotFound
	}
	return factory(cfg)
}
