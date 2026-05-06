// Package middleware HTTP 中间件
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/life-recorder/gateway/internal/pkg/jwt"
	"github.com/life-recorder/gateway/internal/pkg/response"
)

// AuthRequired JWT 认证中间件
func AuthRequired(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "缺少认证令牌")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "认证令牌格式错误")
			c.Abort()
			return
		}

		claims, err := jwt.ParseToken(parts[1], secret)
		if err != nil {
			response.Unauthorized(c, "认证令牌无效或已过期")
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("roles", claims.Roles)
		c.Set("permissions", claims.Permissions)
		c.Next()
	}
}

// OptionalAuth 可选认证中间件（有 token 则解析，无 token 也放行）
func OptionalAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				if claims, err := jwt.ParseToken(parts[1], secret); err == nil {
					c.Set("user_id", claims.UserID)
					c.Set("username", claims.Username)
					c.Set("roles", claims.Roles)
					c.Set("permissions", claims.Permissions)
				}
			}
		}
		c.Next()
	}
}

// RequirePermission 权限检查中间件
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		permissions, exists := c.Get("permissions")
		if !exists {
			response.Forbidden(c, "无访问权限")
			c.Abort()
			return
		}

		permList, ok := permissions.([]string)
		if !ok {
			response.Forbidden(c, "无访问权限")
			c.Abort()
			return
		}

		// admin 角色拥有所有权限
		roles, _ := c.Get("roles")
		if roleList, ok := roles.([]string); ok {
			for _, r := range roleList {
				if r == "admin" {
					c.Next()
					return
				}
			}
		}

		for _, p := range permList {
			if p == permission || p == "*" {
				c.Next()
				return
			}
		}

		response.Forbidden(c, "无此操作权限")
		c.Abort()
	}
}

// CORS 跨域中间件配置
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, Accept, X-Requested-With")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Disposition")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RequestLogger 请求日志中间件
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Gin 自带的 Logger 已足够，这里可以添加自定义逻辑
		c.Next()
	}
}
