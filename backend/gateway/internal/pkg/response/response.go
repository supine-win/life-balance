// Package response 统一响应格式
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// PageData 分页数据
type PageData struct {
	Items    interface{} `json:"items"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

// 成功响应码
const (
	CodeSuccess = 0
)

// 错误响应码
const (
	CodeBadRequest    = 40000
	CodeUnauthorized  = 40100
	CodeForbidden     = 40300
	CodeNotFound      = 40400
	CodeConflict      = 40900
	CodeInternalError = 50000
)

// OK 成功响应
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: "success",
		Data:    data,
	})
}

// OKWithMessage 成功响应（自定义消息）
func OKWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: message,
		Data:    data,
	})
}

// OKPage 分页成功响应
func OKPage(c *gin.Context, items interface{}, total int64, page, pageSize int) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: "success",
		Data: PageData{
			Items:    items,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		},
	})
}

// Fail 错误响应
func Fail(c *gin.Context, httpStatus, code int, message string) {
	c.JSON(httpStatus, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

// BadRequest 参数错误
func BadRequest(c *gin.Context, message string) {
	Fail(c, http.StatusBadRequest, CodeBadRequest, message)
}

// Unauthorized 未认证
func Unauthorized(c *gin.Context, message string) {
	Fail(c, http.StatusUnauthorized, CodeUnauthorized, message)
}

// Forbidden 无权限
func Forbidden(c *gin.Context, message string) {
	Fail(c, http.StatusForbidden, CodeForbidden, message)
}

// NotFound 未找到
func NotFound(c *gin.Context, message string) {
	Fail(c, http.StatusNotFound, CodeNotFound, message)
}

// Conflict 冲突
func Conflict(c *gin.Context, message string) {
	Fail(c, http.StatusConflict, CodeConflict, message)
}

// InternalError 内部错误
func InternalError(c *gin.Context, message string) {
	Fail(c, http.StatusInternalServerError, CodeInternalError, message)
}
