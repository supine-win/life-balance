// Package uuid UUID 生成工具
package uuid

import "github.com/google/uuid"

// New 生成新的 UUID 字符串
func New() string {
	return uuid.New().String()
}
