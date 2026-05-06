// Package storage 存储适配器错误定义
package storage

import "errors"

var (
	// ErrAdapterNotFound 适配器未找到
	ErrAdapterNotFound = errors.New("storage adapter not found")
	// ErrFileNotFound 文件未找到
	ErrFileNotFound = errors.New("file not found")
	// ErrFileExists 文件已存在
	ErrFileExists = errors.New("file already exists")
)
