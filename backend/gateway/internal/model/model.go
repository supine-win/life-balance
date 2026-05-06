// Package model 数据模型定义
// 定义所有数据库实体和 GORM 模型
package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// JSONString 自定义 JSON 类型，兼容 SQLite
type JSONString map[string]interface{}

func (j JSONString) Value() (driver.Value, error) {
	if j == nil {
		return "{}", nil
	}
	bytes, err := json.Marshal(j)
	return string(bytes), err
}

func (j *JSONString) Scan(value interface{}) error {
	if value == nil {
		*j = JSONString{}
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case string:
		bytes = []byte(v)
	case []byte:
		bytes = v
	}
	return json.Unmarshal(bytes, j)
}

// StringArray 自定义字符串数组类型
type StringArray []string

func (s StringArray) Value() (driver.Value, error) {
	if s == nil {
		return "[]", nil
	}
	bytes, err := json.Marshal(s)
	return string(bytes), err
}

func (s *StringArray) Scan(value interface{}) error {
	if value == nil {
		*s = StringArray{}
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case string:
		bytes = []byte(v)
	case []byte:
		bytes = v
	}
	return json.Unmarshal(bytes, s)
}

// BaseModel 基础模型
type BaseModel struct {
	ID        string         `gorm:"type:varchar(36);primaryKey" json:"id"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate 创建前生成 UUID
func (b *BaseModel) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	return nil
}

// ==================== 用户与权限 ====================

// User 用户模型
type User struct {
	BaseModel
	Username  string `gorm:"type:varchar(64);uniqueIndex;not null" json:"username"`
	Email     string `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Password  string `gorm:"type:varchar(255);not null" json:"-"`
	Nickname  string `gorm:"type:varchar(64);default:''" json:"nickname"`
	Avatar    string `gorm:"type:varchar(512);default:''" json:"avatar"`
	Status    int    `gorm:"type:tinyint;default:1" json:"status"` // 1=正常 0=禁用
	LastLogin *time.Time `json:"last_login"`
	Roles     []Role `gorm:"many2many:user_roles" json:"roles,omitempty"`
}

func (User) TableName() string { return "users" }

// Role 角色模型
type Role struct {
	ID          string    `gorm:"type:varchar(36);primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(64);uniqueIndex;not null" json:"name"`
	Description string    `gorm:"type:varchar(255);default:''" json:"description"`
	SortOrder   int       `gorm:"default:0" json:"sort_order"`
	Status      int       `gorm:"type:tinyint;default:1" json:"status"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	Permissions []Permission `gorm:"many2many:role_permissions" json:"permissions,omitempty"`
}

func (Role) TableName() string { return "roles" }

// Permission 权限模型
type Permission struct {
	ID          string    `gorm:"type:varchar(36);primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(128);uniqueIndex;not null" json:"name"`
	Resource    string    `gorm:"type:varchar(64);not null;index" json:"resource"`
	Action      string    `gorm:"type:varchar(32);not null" json:"action"`
	Description string    `gorm:"type:varchar(255);default:''" json:"description"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (Permission) TableName() string { return "permissions" }

// RefreshToken 刷新令牌模型
type RefreshToken struct {
	ID         string     `gorm:"type:varchar(36);primaryKey" json:"id"`
	UserID     string     `gorm:"type:varchar(36);not null;index" json:"user_id"`
	Token      string     `gorm:"type:varchar(512);uniqueIndex;not null" json:"-"`
	DeviceInfo string     `gorm:"type:varchar(255);default:''" json:"device_info"`
	ExpiresAt  time.Time  `gorm:"not null;index" json:"expires_at"`
	RevokedAt  *time.Time `json:"revoked_at"`
	CreatedAt  time.Time  `gorm:"autoCreateTime" json:"created_at"`
}

func (RefreshToken) TableName() string { return "refresh_tokens" }

// ==================== 核心业务 ====================

// Event 事件模型
type Event struct {
	BaseModel
	UserID       string      `gorm:"type:varchar(36);not null;index" json:"user_id"`
	Title        string      `gorm:"type:varchar(255);not null" json:"title"`
	Description  string      `gorm:"type:text" json:"description"`
	EventTime    time.Time   `gorm:"not null;index" json:"event_time"`
	EndTime      *time.Time  `json:"end_time"`
	Location     string      `gorm:"type:varchar(255);default:''" json:"location"`
	Latitude     *float64    `gorm:"type:decimal(10,7)" json:"latitude"`
	Longitude    *float64    `gorm:"type:decimal(10,7)" json:"longitude"`
	Mood         string      `gorm:"type:varchar(32);default:''" json:"mood"`
	Tags         StringArray `gorm:"type:json" json:"tags"`
	Participants StringArray `gorm:"type:json" json:"participants"`
	Category     string      `gorm:"type:varchar(64);default:''" json:"category"`
	Color        string      `gorm:"type:varchar(16);default:''" json:"color"`
	Priority     int         `gorm:"type:tinyint;default:0" json:"priority"`
	IsConfirmed  bool        `gorm:"default:false" json:"is_confirmed"`
	Source       string      `gorm:"type:varchar(32);default:'text'" json:"source"`
	SourceData   JSONString  `gorm:"type:json" json:"source_data"`
	IsDeleted    bool        `gorm:"default:false" json:"is_deleted"`
	Media        []Media     `gorm:"foreignKey:EventID" json:"media,omitempty"`
	User         User        `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (Event) TableName() string { return "events" }

// Media 媒体模型
type Media struct {
	ID              string     `gorm:"type:varchar(36);primaryKey" json:"id"`
	UserID          string     `gorm:"type:varchar(36);not null;index" json:"user_id"`
	EventID         *string    `gorm:"type:varchar(36);index" json:"event_id"`
	Type            string     `gorm:"type:varchar(16);not null" json:"type"` // image/video/audio/document
	StorageAdapter  string     `gorm:"type:varchar(32);not null;default:'local'" json:"storage_adapter"`
	StoragePath     string     `gorm:"type:varchar(512);not null" json:"storage_path"`
	OriginalName    string     `gorm:"type:varchar(255);default:''" json:"original_name"`
	ThumbnailPath   string     `gorm:"type:varchar(512);default:''" json:"thumbnail_path"`
	MimeType        string     `gorm:"type:varchar(128);not null" json:"mime_type"`
	Size            int64      `gorm:"not null;default:0" json:"size"`
	Width           *int       `json:"width"`
	Height          *int       `json:"height"`
	Duration        *float64   `gorm:"type:decimal(10,2)" json:"duration"`
	FileHash        string     `gorm:"type:varchar(64);index" json:"file_hash"`
	Metadata        JSONString `gorm:"type:json" json:"metadata"`
	SortOrder       int        `gorm:"default:0" json:"sort_order"`
	CreatedAt       time.Time  `gorm:"autoCreateTime" json:"created_at"`
}

func (Media) TableName() string { return "media" }

// Tag 标签模型
type Tag struct {
	ID         string    `gorm:"type:varchar(36);primaryKey" json:"id"`
	UserID     string    `gorm:"type:varchar(36);not null" json:"user_id"`
	Name       string    `gorm:"type:varchar(64);not null" json:"name"`
	Color      string    `gorm:"type:varchar(16);default:''" json:"color"`
	UsageCount int       `gorm:"default:0" json:"usage_count"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (Tag) TableName() string { return "tags" }

// Category 分类模型
type Category struct {
	ID        string    `gorm:"type:varchar(36);primaryKey" json:"id"`
	UserID    string    `gorm:"type:varchar(36);not null" json:"user_id"`
	Name      string    `gorm:"type:varchar(64);not null" json:"name"`
	Icon      string    `gorm:"type:varchar(64);default:''" json:"icon"`
	Color     string    `gorm:"type:varchar(16);default:''" json:"color"`
	SortOrder int       `gorm:"default:0" json:"sort_order"`
	ParentID  *string   `gorm:"type:varchar(36)" json:"parent_id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (Category) TableName() string { return "categories" }

// ==================== 配置系统 ====================

// Config 配置模型
type Config struct {
	ID          string    `gorm:"type:varchar(36);primaryKey" json:"id"`
	UserID      *string   `gorm:"type:varchar(36)" json:"user_id"`
	Scope       string    `gorm:"type:varchar(16);not null;default:'system'" json:"scope"` // system / user
	Category    string    `gorm:"type:varchar(64);not null;index" json:"category"`
	Key         string    `gorm:"type:varchar(128);not null" json:"key"`
	Value       string    `gorm:"type:text;not null" json:"value"`
	ValueType   string    `gorm:"type:varchar(16);default:'string'" json:"value_type"`
	Description string    `gorm:"type:varchar(255);default:''" json:"description"`
	IsPublic    bool      `gorm:"default:false" json:"is_public"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Config) TableName() string { return "configs" }

// ==================== 展示系统 ====================

// DisplayTemplate 展示模板
type DisplayTemplate struct {
	ID        string    `gorm:"type:varchar(36);primaryKey" json:"id"`
	UserID    string    `gorm:"type:varchar(36);not null" json:"user_id"`
	Name      string    `gorm:"type:varchar(128);not null" json:"name"`
	Type      string    `gorm:"type:varchar(32);not null" json:"type"` // timeline/calendar/slideshow/grid/map
	Config    JSONString `gorm:"type:json;not null" json:"config"`
	IsDefault bool      `gorm:"default:false" json:"is_default"`
	IsBuiltin bool      `gorm:"default:false" json:"is_builtin"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (DisplayTemplate) TableName() string { return "display_templates" }

// Slideshow 幻灯片
type Slideshow struct {
	ID           string     `gorm:"type:varchar(36);primaryKey" json:"id"`
	UserID       string     `gorm:"type:varchar(36);not null" json:"user_id"`
	Title        string     `gorm:"type:varchar(255);not null" json:"title"`
	Description  string     `gorm:"type:text" json:"description"`
	TemplateID   *string    `gorm:"type:varchar(36)" json:"template_id"`
	EventIDs     StringArray `gorm:"type:json" json:"event_ids"`
	MediaIDs     StringArray `gorm:"type:json" json:"media_ids"`
	BgMusicID    *string    `gorm:"type:varchar(36)" json:"bg_music_id"`
	Config       JSONString `gorm:"type:json" json:"config"`
	Duration     int        `gorm:"default:0" json:"duration"`
	IsPublished  bool       `gorm:"default:false" json:"is_published"`
	PublishedURL string     `gorm:"type:varchar(255);default:''" json:"published_url"`
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Slideshow) TableName() string { return "slideshows" }

// ==================== Webhook ====================

// Webhook Webhook 配置
type Webhook struct {
	ID         string     `gorm:"type:varchar(36);primaryKey" json:"id"`
	UserID     string     `gorm:"type:varchar(36);not null" json:"user_id"`
	Name       string     `gorm:"type:varchar(128);not null" json:"name"`
	URL        string     `gorm:"type:varchar(512);not null" json:"url"`
	Secret     string     `gorm:"type:varchar(255);default:''" json:"-"`
	Events     StringArray `gorm:"type:json;not null" json:"events"`
	Headers    JSONString `gorm:"type:json" json:"headers"`
	IsActive   bool       `gorm:"default:true" json:"is_active"`
	RetryCount int        `gorm:"type:tinyint;default:3" json:"retry_count"`
	TimeoutMs  int        `gorm:"default:5000" json:"timeout_ms"`
	CreatedAt  time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Webhook) TableName() string { return "webhooks" }

// WebhookDelivery Webhook 投递日志
type WebhookDelivery struct {
	ID             string     `gorm:"type:varchar(36);primaryKey" json:"id"`
	WebhookID      string     `gorm:"type:varchar(36);not null;index" json:"webhook_id"`
	EventType      string     `gorm:"type:varchar(64);not null" json:"event_type"`
	Payload        JSONString `gorm:"type:json;not null" json:"payload"`
	RequestHeaders JSONString `gorm:"type:json" json:"request_headers"`
	ResponseStatus *int       `json:"response_status"`
	ResponseBody   string     `gorm:"type:text" json:"response_body"`
	DurationMs     *int       `json:"duration_ms"`
	Success        bool       `gorm:"default:false" json:"success"`
	Attempts       int        `gorm:"type:tinyint;default:0" json:"attempts"`
	NextRetry      *time.Time `gorm:"index" json:"next_retry"`
	CreatedAt      time.Time  `gorm:"autoCreateTime;index" json:"created_at"`
}

func (WebhookDelivery) TableName() string { return "webhook_deliveries" }

// ==================== 系统管理 ====================

// AuditLog 审计日志
type AuditLog struct {
	ID         string     `gorm:"type:varchar(36);primaryKey" json:"id"`
	UserID     *string    `gorm:"type:varchar(36);index" json:"user_id"`
	Username   string     `gorm:"type:varchar(64)" json:"username"`
	Action     string     `gorm:"type:varchar(64);not null" json:"action"`
	Resource   string     `gorm:"type:varchar(64);not null;index" json:"resource"`
	ResourceID *string    `gorm:"type:varchar(36)" json:"resource_id"`
	Detail     JSONString `gorm:"type:json" json:"detail"`
	IPAddress  string     `gorm:"type:varchar(45)" json:"ip_address"`
	UserAgent  string     `gorm:"type:varchar(512)" json:"user_agent"`
	CreatedAt  time.Time  `gorm:"autoCreateTime;index" json:"created_at"`
}

func (AuditLog) TableName() string { return "audit_logs" }

// Backup 备份记录
type Backup struct {
	ID          string     `gorm:"type:varchar(36);primaryKey" json:"id"`
	UserID      *string    `gorm:"type:varchar(36)" json:"user_id"`
	Filename    string     `gorm:"type:varchar(255);not null" json:"filename"`
	FilePath    string     `gorm:"type:varchar(512);not null" json:"file_path"`
	FileSize    int64      `gorm:"default:0" json:"file_size"`
	BackupType  string     `gorm:"type:varchar(32);default:'full'" json:"backup_type"`
	Scope       JSONString `gorm:"type:json" json:"scope"`
	Status      string     `gorm:"type:varchar(16);default:'pending'" json:"status"`
	Error       string     `gorm:"type:text" json:"error"`
	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
}

func (Backup) TableName() string { return "backups" }

// AITask AI 任务
type AITask struct {
	ID          string     `gorm:"type:varchar(36);primaryKey" json:"id"`
	UserID      string     `gorm:"type:varchar(36);not null" json:"user_id"`
	Type        string     `gorm:"type:varchar(32);not null" json:"type"`
	Status      string     `gorm:"type:varchar(16);default:'pending'" json:"status"`
	InputData   JSONString `gorm:"type:json;not null" json:"input_data"`
	OutputData  JSONString `gorm:"type:json" json:"output_data"`
	Error       string     `gorm:"type:text" json:"error"`
	Progress    int        `gorm:"type:tinyint;default:0" json:"progress"`
	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
}

func (AITask) TableName() string { return "ai_tasks" }
