// Package handler HTTP 处理器
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/life-recorder/gateway/internal/handler/middleware"
	"github.com/life-recorder/gateway/internal/pkg/response"
	"github.com/life-recorder/gateway/internal/pkg/validator"
	"github.com/life-recorder/gateway/internal/service"
)

// ==================== Auth Handler ====================

// AuthHandler 认证处理器
type AuthHandler struct {
	authService *service.AuthService
	jwtSecret   string
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(authService *service.AuthService, jwtSecret string) *AuthHandler {
	return &AuthHandler{authService: authService, jwtSecret: jwtSecret}
}

// Register 注册
// POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var input service.RegisterInput
	if !validator.ValidateJSON(c, &input) {
		return
	}

	result, err := h.authService.Register(&input)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, result)
}

// Login 登录
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var input service.LoginInput
	if !validator.ValidateJSON(c, &input) {
		return
	}

	result, err := h.authService.Login(&input)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	response.OK(c, result)
}

// Refresh 刷新令牌
// POST /api/v1/auth/refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	var input struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if !validator.ValidateJSON(c, &input) {
		return
	}

	result, err := h.authService.RefreshToken(input.RefreshToken)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	response.OK(c, result)
}

// Logout 登出
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// TODO: 撤销 refresh_token
	response.OKWithMessage(c, "已登出", nil)
}

// Me 获取当前用户信息
// GET /api/v1/auth/me
func (h *AuthHandler) Me(c *gin.Context) {
	userID := validator.GetUserID(c)
	user, err := h.authService.GetCurrentUser(userID)
	if err != nil {
		response.NotFound(c, "用户不存在")
		return
	}

	// 构建响应（不返回密码等敏感信息）
	type UserResponse struct {
		ID          string   `json:"id"`
		Username    string   `json:"username"`
		Email       string   `json:"email"`
		Nickname    string   `json:"nickname"`
		Avatar      string   `json:"avatar"`
		Roles       []string `json:"roles"`
		Permissions []string `json:"permissions"`
	}

	var roles []string
	var permissions []string
	for _, r := range user.Roles {
		roles = append(roles, r.Name)
		for _, p := range r.Permissions {
			permissions = append(permissions, p.Name)
		}
	}

	resp := UserResponse{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		Nickname:    user.Nickname,
		Avatar:      user.Avatar,
		Roles:       roles,
		Permissions: permissions,
	}

	response.OK(c, resp)
}

// UpdateProfile 更新个人资料
// PUT /api/v1/auth/profile
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	// TODO: 实现资料更新
	response.OK(c, nil)
}

// ==================== Event Handler ====================

// EventHandler 事件处理器
type EventHandler struct {
	eventService *service.EventService
}

// NewEventHandler 创建事件处理器
func NewEventHandler(eventService *service.EventService) *EventHandler {
	return &EventHandler{eventService: eventService}
}

// List 事件列表
// GET /api/v1/events
func (h *EventHandler) List(c *gin.Context) {
	userID := validator.GetUserID(c)
	page, pageSize := validator.NormalizePagination(
		queryIntDefault(c, "page", 1),
		queryIntDefault(c, "page_size", 20),
	)

	filters := make(map[string]interface{})
	if v := c.Query("start_time"); v != "" {
		filters["start_time"] = v
	}
	if v := c.Query("end_time"); v != "" {
		filters["end_time"] = v
	}
	if v := c.Query("category"); v != "" {
		filters["category"] = v
	}
	if v := c.Query("mood"); v != "" {
		filters["mood"] = v
	}
	if v := c.Query("keyword"); v != "" {
		filters["keyword"] = v
	}
	if v := c.Query("confirmed"); v != "" {
		filters["confirmed"] = v == "true"
	}
	if v := c.Query("sort"); v != "" {
		filters["sort"] = v
	} else {
		filters["sort"] = "event_time"
	}
	if v := c.Query("order"); v != "" {
		filters["order"] = v
	} else {
		filters["order"] = "DESC"
	}

	events, total, err := h.eventService.List(userID, page, pageSize, filters)
	if err != nil {
		response.InternalError(c, "查询事件失败")
		return
	}

	response.OKPage(c, events, total, page, pageSize)
}

// Create 创建事件
// POST /api/v1/events
func (h *EventHandler) Create(c *gin.Context) {
	userID := validator.GetUserID(c)
	var input service.CreateInput
	if !validator.ValidateJSON(c, &input) {
		return
	}

	event, err := h.eventService.Create(userID, &input)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, response.Response{
		Code:    0,
		Message: "success",
		Data:    event,
	})
}

// GetByID 获取事件详情
// GET /api/v1/events/:id
func (h *EventHandler) GetByID(c *gin.Context) {
	userID := validator.GetUserID(c)
	id := c.Param("id")

	event, err := h.eventService.GetByID(id, userID)
	if err != nil {
		response.NotFound(c, "事件不存在")
		return
	}

	response.OK(c, event)
}

// Update 更新事件
// PUT /api/v1/events/:id
func (h *EventHandler) Update(c *gin.Context) {
	userID := validator.GetUserID(c)
	id := c.Param("id")

	var input service.CreateInput
	if !validator.ValidateJSON(c, &input) {
		return
	}

	event, err := h.eventService.Update(id, userID, &input)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, event)
}

// Delete 删除事件
// DELETE /api/v1/events/:id
func (h *EventHandler) Delete(c *gin.Context) {
	userID := validator.GetUserID(c)
	id := c.Param("id")

	if err := h.eventService.Delete(id, userID); err != nil {
		response.InternalError(c, "删除事件失败")
		return
	}

	response.OKWithMessage(c, "已删除", nil)
}

// Confirm 确认事件
// POST /api/v1/events/confirm/:id
func (h *EventHandler) Confirm(c *gin.Context) {
	userID := validator.GetUserID(c)
	id := c.Param("id")

	var input struct {
		Modifications map[string]interface{} `json:"modifications"`
	}
	validator.ValidateJSON(c, &input)

	event, err := h.eventService.Confirm(id, userID, input.Modifications)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, event)
}

// CalendarMonth 日历视图
// GET /api/v1/events/calendar/:year/:month
func (h *EventHandler) CalendarMonth(c *gin.Context) {
	userID := validator.GetUserID(c)
	year := queryIntDefault(c, "year", 0)
	month := queryIntDefault(c, "month", 0)

	if year == 0 || month == 0 {
		response.BadRequest(c, "无效的年份或月份")
		return
	}

	events, err := h.eventService.CalendarMonth(userID, year, month)
	if err != nil {
		response.InternalError(c, "查询日历数据失败")
		return
	}

	response.OK(c, gin.H{"events": events})
}

// ApplySuggestions 应用元数据建议
// POST /api/v1/events/:id/apply-suggestions
func (h *EventHandler) ApplySuggestions(c *gin.Context) {
	userID := validator.GetUserID(c)
	eventID := c.Param("id")

	var input struct {
		MediaID string   `json:"media_id" binding:"required"`
		Fields  []string `json:"fields"`
	}
	if !validator.ValidateJSON(c, &input) {
		return
	}

	event, err := h.eventService.ApplySuggestions(eventID, userID, input.MediaID, input.Fields)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, event)
}

// ==================== Media Handler ====================

// MediaHandler 媒体处理器
type MediaHandler struct {
	mediaService *service.MediaService
}

// NewMediaHandler 创建媒体处理器
func NewMediaHandler(mediaService *service.MediaService) *MediaHandler {
	return &MediaHandler{mediaService: mediaService}
}

// Upload 上传文件
// POST /api/v1/media/upload
func (h *MediaHandler) Upload(c *gin.Context) {
	userID := validator.GetUserID(c)
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请选择文件")
		return
	}

	eventIDStr := c.PostForm("event_id")
	var eventID *string
	if eventIDStr != "" {
		eventID = &eventIDStr
	}

	media, err := h.mediaService.Upload(c.Request.Context(), userID, file, eventID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, response.Response{
		Code:    0,
		Message: "success",
		Data:    media,
	})
}

// UploadWithMetadata 上传并提取元数据
// POST /api/v1/media/upload-with-metadata
func (h *MediaHandler) UploadWithMetadata(c *gin.Context) {
	userID := validator.GetUserID(c)
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请选择文件")
		return
	}

	eventIDStr := c.PostForm("event_id")
	var eventID *string
	if eventIDStr != "" {
		eventID = &eventIDStr
	}

	result, err := h.mediaService.UploadWithMetadata(c.Request.Context(), userID, file, eventID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, response.Response{
		Code:    0,
		Message: "success",
		Data:    result,
	})
}

// GetByID 获取媒体信息
// GET /api/v1/media/:id
func (h *MediaHandler) GetByID(c *gin.Context) {
	userID := validator.GetUserID(c)
	id := c.Param("id")

	media, err := h.mediaService.GetByID(id, userID)
	if err != nil {
		response.NotFound(c, "媒体不存在")
		return
	}

	response.OK(c, media)
}

// GetMetadata 获取媒体元数据
// GET /api/v1/media/:id/metadata
func (h *MediaHandler) GetMetadata(c *gin.Context) {
	id := c.Param("id")

	metadata, err := h.mediaService.GetMetadata(id)
	if err != nil {
		response.NotFound(c, "元数据不存在")
		return
	}

	response.OK(c, metadata)
}

// GetFile 获取/下载文件
// GET /api/v1/media/:id/file
func (h *MediaHandler) GetFile(c *gin.Context) {
	userID := validator.GetUserID(c)
	id := c.Param("id")

	reader, media, err := h.mediaService.GetFile(c.Request.Context(), id, userID)
	if err != nil {
		response.NotFound(c, "文件不存在")
		return
	}
	defer reader.Close()

	c.Header("Content-Type", media.MimeType)
	c.Header("Content-Disposition", "inline; filename=\""+media.OriginalName+"\"")
	c.DataFromReader(http.StatusOK, media.Size, media.MimeType, reader, nil)
}

// Delete 删除媒体
// DELETE /api/v1/media/:id
func (h *MediaHandler) Delete(c *gin.Context) {
	userID := validator.GetUserID(c)
	id := c.Param("id")

	if err := h.mediaService.Delete(c.Request.Context(), id, userID); err != nil {
		response.InternalError(c, "删除失败")
		return
	}

	response.OKWithMessage(c, "已删除", nil)
}

// List 媒体列表
// GET /api/v1/media
func (h *MediaHandler) List(c *gin.Context) {
	userID := validator.GetUserID(c)
	page, pageSize := validator.NormalizePagination(
		queryIntDefault(c, "page", 1),
		queryIntDefault(c, "page_size", 20),
	)
	mediaType := c.Query("type")

	medias, total, err := h.mediaService.List(userID, page, pageSize, mediaType)
	if err != nil {
		response.InternalError(c, "查询媒体列表失败")
		return
	}

	response.OKPage(c, medias, total, page, pageSize)
}

// ==================== Setup Handler ====================

// SetupHandler 安装初始化处理器
type SetupHandler struct {
	authService *service.AuthService
}

// NewSetupHandler 创建安装处理器
func NewSetupHandler(authService *service.AuthService) *SetupHandler {
	return &SetupHandler{authService: authService}
}

// Status 检查初始化状态
// GET /api/v1/setup/status
func (h *SetupHandler) Status(c *gin.Context) {
	response.OK(c, gin.H{
		"initialized": false, // TODO: 从数据库读取
	})
}

// Init 执行初始化
// POST /api/v1/setup/init
func (h *SetupHandler) Init(c *gin.Context) {
	var input service.SetupInitInput
	if !validator.ValidateJSON(c, &input) {
		return
	}

	result, err := h.authService.SetupInit(&input)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, result)
}

// ==================== System Handler ====================

// SystemHandler 系统处理器
type SystemHandler struct{}

// NewSystemHandler 创建系统处理器
func NewSystemHandler() *SystemHandler {
	return &SystemHandler{}
}

// Health 健康检查
// GET /api/v1/system/health
func (h *SystemHandler) Health(c *gin.Context) {
	response.OK(c, gin.H{
		"status":  "healthy",
		"version": "1.0.0",
	})
}

// ==================== 辅助函数 ====================

// queryIntDefault 获取查询参数整数，带默认值
func queryIntDefault(c *gin.Context, key string, defaultVal int) int {
	val := c.Query(key)
	if val == "" {
		return defaultVal
	}
	var result int
	if _, err := validator.NormalizePagination(1, 20); err != nil {
		_ = result
	}
	// 简单实现
	var i int
	for _, ch := range val {
		if ch >= '0' && ch <= '9' {
			i = i*10 + int(ch-'0')
		} else {
			return defaultVal
		}
	}
	if i == 0 {
		return defaultVal
	}
	return i
}

// Ensure middleware import is used
var _ = middleware.AuthRequired
