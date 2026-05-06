// Package validator 请求参数验证
package validator

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// ValidateJSON 验证 JSON 请求体
func ValidateJSON(c *gin.Context, obj interface{}) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		c.JSON(400, gin.H{
			"code":    40000,
			"message": "参数错误: " + err.Error(),
			"data":    nil,
		})
		return false
	}
	return true
}

// ValidateQuery 验证查询参数
func ValidateQuery(c *gin.Context, obj interface{}) bool {
	if err := c.ShouldBindQuery(obj); err != nil {
		c.JSON(400, gin.H{
			"code":    40000,
			"message": "参数错误: " + err.Error(),
			"data":    nil,
		})
		return false
	}
	return true
}

// ValidateURI 验证 URI 参数
func ValidateURI(c *gin.Context, obj interface{}) bool {
	if err := c.ShouldBindUri(obj); err != nil {
		c.JSON(400, gin.H{
			"code":    40000,
			"message": "参数错误: " + err.Error(),
			"data":    nil,
		})
		return false
	}
	return true
}

// BindForm 绑定表单数据
func BindForm(c *gin.Context, obj interface{}) bool {
	if err := c.MustBindWith(obj, binding.Form); err != nil {
		c.JSON(400, gin.H{
			"code":    40000,
			"message": "参数错误: " + err.Error(),
			"data":    nil,
		})
		return false
	}
	return true
}

// GetUserID 从上下文获取用户 ID
func GetUserID(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	return ""
}

// GetUsername 从上下文获取用户名
func GetUsername(c *gin.Context) string {
	if username, exists := c.Get("username"); exists {
		if name, ok := username.(string); ok {
			return name
		}
	}
	return ""
}

// NormalizePagination 规范化分页参数
func NormalizePagination(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

// SanitizeString 清理字符串
func SanitizeString(s string) string {
	return strings.TrimSpace(s)
}
