// Package repository 数据访问层
package repository

import (
	"fmt"

	"github.com/life-recorder/gateway/internal/config"
	"github.com/life-recorder/gateway/internal/model"
	"github.com/life-recorder/gateway/internal/pkg/logger"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// DB 全局数据库实例
var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch cfg.Type {
	case "sqlite":
		dialector = sqlite.Open(cfg.DSN())
	case "postgres":
		dialector = postgres.Open(cfg.DSN())
	case "mysql":
		dialector = mysql.Open(cfg.DSN())
	default:
		dialector = sqlite.Open(cfg.DSN())
	}

	var logLevel gormlogger.LogLevel
	switch cfg.LogLevel {
	case "silent":
		logLevel = gormlogger.Silent
	case "error":
		logLevel = gormlogger.Error
	case "warn":
		logLevel = gormlogger.Warn
	case "info":
		logLevel = gormlogger.Info
	default:
		logLevel = gormlogger.Warn
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger:                 gormlogger.Default.LogMode(logLevel),
		SkipDefaultTransaction: true,
		PrepareStmt:           true,
	})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取数据库实例失败: %w", err)
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)

	DB = db
	return db, nil
}

// AutoMigrate 自动迁移数据库表结构
func AutoMigrate(db *gorm.DB) error {
	models := []interface{}{
		&model.User{},
		&model.Role{},
		&model.Permission{},
		&model.RefreshToken{},
		&model.Event{},
		&model.Media{},
		&model.MediaMetadata{},
		&model.Tag{},
		&model.Category{},
		&model.Config{},
		&model.DisplayTemplate{},
		&model.Slideshow{},
		&model.Webhook{},
		&model.WebhookDelivery{},
		&model.AuditLog{},
		&model.Backup{},
		&model.AITask{},
		&model.SystemSetting{},
	}

	if err := db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("数据库迁移失败: %w", err)
	}

	logger.S().Info("数据库迁移完成")
	return nil
}

// ==================== 通用查询 ====================

// PaginationResult 分页结果
type PaginationResult struct {
	Items    interface{}
	Total    int64
	Page     int
	PageSize int
}

// Paginate 通用分页查询
func Paginate(db *gorm.DB, page, pageSize int, result interface{}) (*PaginationResult, error) {
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	offset := (page - 1) * pageSize
	if err := db.Offset(offset).Limit(pageSize).Find(result).Error; err != nil {
		return nil, err
	}

	return &PaginationResult{
		Items:    result,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// ==================== User Repository ====================

// UserRepo 用户数据访问
type UserRepo struct {
	db *gorm.DB
}

// NewUserRepo 创建用户仓库
func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

// Create 创建用户
func (r *UserRepo) Create(user *model.User) error {
	return r.db.Create(user).Error
}

// FindByID 根据 ID 查找用户
func (r *UserRepo) FindByID(id string) (*model.User, error) {
	var user model.User
	if err := r.db.Preload("Roles.Permissions").First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByUsername 根据用户名查找
func (r *UserRepo) FindByUsername(username string) (*model.User, error) {
	var user model.User
	if err := r.db.Preload("Roles.Permissions").First(&user, "username = ? OR email = ?", username, username).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// Update 更新用户
func (r *UserRepo) Update(user *model.User) error {
	return r.db.Save(user).Error
}

// List 用户列表
func (r *UserRepo) List(page, pageSize int, keyword string) ([]model.User, int64, error) {
	var users []model.User
	var total int64
	query := r.db.Model(&model.User{})
	if keyword != "" {
		query = query.Where("username LIKE ? OR email LIKE ? OR nickname LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	if err := query.Preload("Roles").Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

// ==================== Event Repository ====================

// EventRepo 事件数据访问
type EventRepo struct {
	db *gorm.DB
}

// NewEventRepo 创建事件仓库
func NewEventRepo(db *gorm.DB) *EventRepo {
	return &EventRepo{db: db}
}

// Create 创建事件
func (r *EventRepo) Create(event *model.Event) error {
	return r.db.Create(event).Error
}

// FindByID 根据 ID 查找事件
func (r *EventRepo) FindByID(id, userID string) (*model.Event, error) {
	var event model.Event
	if err := r.db.Preload("Media").Preload("Media.MediaMetadata").First(&event, "id = ? AND user_id = ?", id, userID).Error; err != nil {
		return nil, err
	}
	return &event, nil
}

// Update 更新事件
func (r *EventRepo) Update(event *model.Event) error {
	return r.db.Save(event).Error
}

// Delete 软删除事件
func (r *EventRepo) Delete(id, userID string) error {
	return r.db.Model(&model.Event{}).Where("id = ? AND user_id = ?", id, userID).Update("is_deleted", true).Error
}

// List 查询事件列表
func (r *EventRepo) List(userID string, page, pageSize int, filters map[string]interface{}) ([]model.Event, int64, error) {
	var events []model.Event
	var total int64
	query := r.db.Model(&model.Event{}).Where("user_id = ? AND is_deleted = ?", userID, false)

	if startTime, ok := filters["start_time"]; ok {
		query = query.Where("event_time >= ?", startTime)
	}
	if endTime, ok := filters["end_time"]; ok {
		query = query.Where("event_time <= ?", endTime)
	}
	if category, ok := filters["category"]; ok {
		query = query.Where("category = ?", category)
	}
	if mood, ok := filters["mood"]; ok {
		query = query.Where("mood = ?", mood)
	}
	if confirmed, ok := filters["confirmed"]; ok {
		query = query.Where("is_confirmed = ?", confirmed)
	}
	if keyword, ok := filters["keyword"]; ok {
		query = query.Where("title LIKE ? OR description LIKE ?", "%"+keyword.(string)+"%", "%"+keyword.(string)+"%")
	}

	sortField := "event_time"
	sortOrder := "DESC"
	if sf, ok := filters["sort"]; ok {
		sortField = sf.(string)
	}
	if so, ok := filters["order"]; ok {
		sortOrder = so.(string)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Preload("Media").Offset(offset).Limit(pageSize).
		Order(sortField + " " + sortOrder).Find(&events).Error; err != nil {
		return nil, 0, err
	}
	return events, total, nil
}

// CalendarMonth 获取月份日历数据
func (r *EventRepo) CalendarMonth(userID string, year, month int) ([]model.Event, error) {
	var events []model.Event
	startDate := fmt.Sprintf("%04d-%02d-01", year, month)
	var endDate string
	if month == 12 {
		endDate = fmt.Sprintf("%04d-01-01", year+1)
	} else {
		endDate = fmt.Sprintf("%04d-%02d-01", year, month+1)
	}
	err := r.db.Where("user_id = ? AND is_deleted = ? AND event_time >= ? AND event_time < ?",
		userID, false, startDate, endDate).
		Select("id, title, event_time, color, category, mood").
		Order("event_time ASC").
		Find(&events).Error
	return events, err
}

// ==================== Media Repository ====================

// MediaRepo 媒体数据访问
type MediaRepo struct {
	db *gorm.DB
}

// NewMediaRepo 创建媒体仓库
func NewMediaRepo(db *gorm.DB) *MediaRepo {
	return &MediaRepo{db: db}
}

// Create 创建媒体记录
func (r *MediaRepo) Create(media *model.Media) error {
	return r.db.Create(media).Error
}

// FindByID 根据 ID 查找
func (r *MediaRepo) FindByID(id, userID string) (*model.Media, error) {
	var media model.Media
	if err := r.db.Preload("MediaMetadata").First(&media, "id = ? AND user_id = ?", id, userID).Error; err != nil {
		return nil, err
	}
	return &media, nil
}

// FindByEventID 根据事件 ID 查找媒体列表
func (r *MediaRepo) FindByEventID(eventID string) ([]model.Media, error) {
	var medias []model.Media
	if err := r.db.Where("event_id = ?", eventID).Order("sort_order ASC").Find(&medias).Error; err != nil {
		return nil, err
	}
	return medias, nil
}

// Delete 删除媒体
func (r *MediaRepo) Delete(id, userID string) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.Media{}).Error
}

// List 媒体列表
func (r *MediaRepo) List(userID string, page, pageSize int, mediaType string) ([]model.Media, int64, error) {
	var medias []model.Media
	var total int64
	query := r.db.Model(&model.Media{}).Where("user_id = ?", userID)
	if mediaType != "" {
		query = query.Where("type = ?", mediaType)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&medias).Error; err != nil {
		return nil, 0, err
	}
	return medias, total, nil
}

// Update 更新媒体
func (r *MediaRepo) Update(media *model.Media) error {
	return r.db.Save(media).Error
}

// ==================== MediaMetadata Repository ====================

// MediaMetadataRepo 媒体元数据访问
type MediaMetadataRepo struct {
	db *gorm.DB
}

// NewMediaMetadataRepo 创建媒体元数据仓库
func NewMediaMetadataRepo(db *gorm.DB) *MediaMetadataRepo {
	return &MediaMetadataRepo{db: db}
}

// Create 创建元数据记录
func (r *MediaMetadataRepo) Create(metadata *model.MediaMetadata) error {
	return r.db.Create(metadata).Error
}

// FindByMediaID 根据媒体 ID 查找
func (r *MediaMetadataRepo) FindByMediaID(mediaID string) (*model.MediaMetadata, error) {
	var metadata model.MediaMetadata
	if err := r.db.First(&metadata, "media_id = ?", mediaID).Error; err != nil {
		return nil, err
	}
	return &metadata, nil
}

// Update 更新元数据
func (r *MediaMetadataRepo) Update(metadata *model.MediaMetadata) error {
	return r.db.Save(metadata).Error
}

// ==================== Config Repository ====================

// ConfigRepo 配置数据访问
type ConfigRepo struct {
	db *gorm.DB
}

// NewConfigRepo 创建配置仓库
func NewConfigRepo(db *gorm.DB) *ConfigRepo {
	return &ConfigRepo{db: db}
}

// Get 获取配置
func (r *ConfigRepo) Get(category, key string, userID *string) (*model.Config, error) {
	var cfg model.Config
	query := r.db.Where("category = ? AND key = ?", category, key)
	if userID != nil {
		query = query.Where("user_id = ? AND scope = 'user'", *userID)
	} else {
		query = query.Where("user_id IS NULL AND scope = 'system'")
	}
	if err := query.First(&cfg).Error; err != nil {
		return nil, err
	}
	return &cfg, nil
}

// List 列出配置
func (r *ConfigRepo) List(scope string, userID *string, category string) ([]model.Config, error) {
	var configs []model.Config
	query := r.db.Where("scope = ?", scope)
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	} else {
		query = query.Where("user_id IS NULL")
	}
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if err := query.Order("category, key").Find(&configs).Error; err != nil {
		return nil, err
	}
	return configs, nil
}

// Upsert 创建或更新配置
func (r *ConfigRepo) Upsert(cfg *model.Config) error {
	var existing model.Config
	result := r.db.Where("category = ? AND key = ? AND (user_id = ? OR (user_id IS NULL AND ? IS NULL))",
		cfg.Category, cfg.Key, cfg.UserID, cfg.UserID).First(&existing)

	if result.Error == gorm.ErrRecordNotFound {
		return r.db.Create(cfg).Error
	}
	return r.db.Model(&existing).Updates(map[string]interface{}{
		"value":       cfg.Value,
		"value_type":  cfg.ValueType,
		"description": cfg.Description,
	}).Error
}

// ==================== Webhook Repository ====================

// WebhookRepo Webhook 数据访问
type WebhookRepo struct {
	db *gorm.DB
}

// NewWebhookRepo 创建 Webhook 仓库
func NewWebhookRepo(db *gorm.DB) *WebhookRepo {
	return &WebhookRepo{db: db}
}

// Create 创建 Webhook
func (r *WebhookRepo) Create(webhook *model.Webhook) error {
	return r.db.Create(webhook).Error
}

// FindByID 根据 ID 查找
func (r *WebhookRepo) FindByID(id, userID string) (*model.Webhook, error) {
	var webhook model.Webhook
	if err := r.db.First(&webhook, "id = ? AND user_id = ?", id, userID).Error; err != nil {
		return nil, err
	}
	return &webhook, nil
}

// List 列出 Webhook
func (r *WebhookRepo) List(userID string) ([]model.Webhook, error) {
	var webhooks []model.Webhook
	if err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&webhooks).Error; err != nil {
		return nil, err
	}
	return webhooks, nil
}

// FindActiveByEvent 查找匹配事件类型的活跃 Webhook
func (r *WebhookRepo) FindActiveByEvent(userID, eventType string) ([]model.Webhook, error) {
	var webhooks []model.Webhook
	if err := r.db.Where("user_id = ? AND is_active = ?", userID, true).Find(&webhooks).Error; err != nil {
		return nil, err
	}
	// 过滤匹配事件类型的 Webhook
	var result []model.Webhook
	for _, w := range webhooks {
		for _, e := range w.Events {
			if e == eventType || e == "*" {
				result = append(result, w)
				break
			}
		}
	}
	return result, nil
}

// Update 更新 Webhook
func (r *WebhookRepo) Update(webhook *model.Webhook) error {
	return r.db.Save(webhook).Error
}

// Delete 删除 Webhook
func (r *WebhookRepo) Delete(id, userID string) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.Webhook{}).Error
}

// CreateDelivery 创建投递记录
func (r *WebhookRepo) CreateDelivery(delivery *model.WebhookDelivery) error {
	return r.db.Create(delivery).Error
}

// ListDeliveries 列出投递记录
func (r *WebhookRepo) ListDeliveries(webhookID string, page, pageSize int) ([]model.WebhookDelivery, int64, error) {
	var deliveries []model.WebhookDelivery
	var total int64
	query := r.db.Model(&model.WebhookDelivery{}).Where("webhook_id = ?", webhookID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&deliveries).Error; err != nil {
		return nil, 0, err
	}
	return deliveries, total, nil
}

// ==================== AuditLog Repository ====================

// AuditLogRepo 审计日志数据访问
type AuditLogRepo struct {
	db *gorm.DB
}

// NewAuditLogRepo 创建审计日志仓库
func NewAuditLogRepo(db *gorm.DB) *AuditLogRepo {
	return &AuditLogRepo{db: db}
}

// Create 创建审计日志
func (r *AuditLogRepo) Create(log *model.AuditLog) error {
	return r.db.Create(log).Error
}

// List 列出审计日志
func (r *AuditLogRepo) List(page, pageSize int, filters map[string]interface{}) ([]model.AuditLog, int64, error) {
	var logs []model.AuditLog
	var total int64
	query := r.db.Model(&model.AuditLog{})
	if userID, ok := filters["user_id"]; ok {
		query = query.Where("user_id = ?", userID)
	}
	if action, ok := filters["action"]; ok {
		query = query.Where("action = ?", action)
	}
	if resource, ok := filters["resource"]; ok {
		query = query.Where("resource = ?", resource)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}

// ==================== Role Repository ====================

// RoleRepo 角色数据访问
type RoleRepo struct {
	db *gorm.DB
}

// NewRoleRepo 创建角色仓库
func NewRoleRepo(db *gorm.DB) *RoleRepo {
	return &RoleRepo{db: db}
}

// List 列出所有角色
func (r *RoleRepo) List() ([]model.Role, error) {
	var roles []model.Role
	if err := r.db.Preload("Permissions").Order("sort_order ASC").Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

// FindByID 根据 ID 查找
func (r *RoleRepo) FindByID(id string) (*model.Role, error) {
	var role model.Role
	if err := r.db.Preload("Permissions").First(&role, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

// Create 创建角色
func (r *RoleRepo) Create(role *model.Role) error {
	return r.db.Create(role).Error
}

// AssignRoles 为用户分配角色
func (r *RoleRepo) AssignRoles(userID string, roleIDs []string) error {
	var roles []model.Role
	if err := r.db.Where("id IN ?", roleIDs).Find(&roles).Error; err != nil {
		return err
	}
	return r.db.Model(&model.User{BaseModel: model.BaseModel{ID: userID}}).Association("Roles").Replace(roles)
}
