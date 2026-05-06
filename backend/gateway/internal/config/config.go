// Package config 配置管理
// 使用 Viper 加载和管理应用配置，支持 YAML 文件和环境变量覆盖
package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config 应用全局配置
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	Storage   StorageConfig   `mapstructure:"storage"`
	AIService AIServiceConfig `mapstructure:"ai_service"`
	Log       LogConfig       `mapstructure:"log"`
	Webhook   WebhookConfig   `mapstructure:"webhook"`
	Setup     SetupConfig     `mapstructure:"setup"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"` // debug / release / test
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type          string `mapstructure:"type"` // sqlite / postgres / mysql
	SQLitePath    string `mapstructure:"sqlite_path"`
	PostgresHost  string `mapstructure:"postgres_host"`
	PostgresPort  int    `mapstructure:"postgres_port"`
	PostgresUser  string `mapstructure:"postgres_user"`
	PostgresPassword string `mapstructure:"postgres_password"`
	PostgresDBName string `mapstructure:"postgres_dbname"`
	PostgresSSLMode string `mapstructure:"postgres_sslmode"`
	MySQLHost     string `mapstructure:"mysql_host"`
	MySQLPort     int    `mapstructure:"mysql_port"`
	MySQLUser     string `mapstructure:"mysql_user"`
	MySQLPassword string `mapstructure:"mysql_password"`
	MySQLDBName   string `mapstructure:"mysql_dbname"`
	MaxIdleConns  int    `mapstructure:"max_idle_conns"`
	MaxOpenConns  int    `mapstructure:"max_open_conns"`
	LogLevel      string `mapstructure:"log_level"`
}

// JWTConfig JWT 认证配置
type JWTConfig struct {
	Secret     string `mapstructure:"secret"`
	AccessTTL  int    `mapstructure:"access_ttl"`   // 秒
	RefreshTTL int    `mapstructure:"refresh_ttl"`  // 秒
	Issuer     string `mapstructure:"issuer"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	DefaultAdapter string            `mapstructure:"default_adapter"`
	Local          LocalStorageConfig `mapstructure:"local"`
	MaxUploadSize  int64             `mapstructure:"max_upload_size"`
}

// LocalStorageConfig 本地存储配置
type LocalStorageConfig struct {
	BasePath string `mapstructure:"base_path"`
	BaseURL  string `mapstructure:"base_url"`
}

// AIServiceConfig AI 服务配置
type AIServiceConfig struct {
	URL     string `mapstructure:"url"`
	Timeout int    `mapstructure:"timeout"` // 秒
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	FilePath   string `mapstructure:"file_path"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
}

// WebhookConfig Webhook 配置
type WebhookConfig struct {
	WorkerCount int   `mapstructure:"worker_count"`
	RetryDelays []int `mapstructure:"retry_delays"`
}

// SetupConfig 初始化配置
type SetupConfig struct {
	Initialized bool `mapstructure:"initialized"`
}

// DSN 返回数据库连接字符串
func (c *DatabaseConfig) DSN() string {
	switch c.Type {
	case "sqlite":
		return c.SQLitePath
	case "postgres":
		return fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			c.PostgresHost, c.PostgresPort, c.PostgresUser,
			c.PostgresPassword, c.PostgresDBName, c.PostgresSSLMode,
		)
	case "mysql":
		return fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			c.MySQLUser, c.MySQLPassword, c.MySQLHost, c.MySQLPort, c.MySQLDBName,
		)
	default:
		return c.SQLitePath
	}
}

// Load 加载配置文件
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// 设置默认值
	setDefaults(v)

	// 从文件读取
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("./configs")
		v.AddConfigPath("/etc/life-recorder")
	}

	// 环境变量覆盖
	v.SetEnvPrefix("LR")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	return &cfg, nil
}

// setDefaults 设置配置默认值
func setDefaults(v *viper.Viper) {
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.mode", "release")
	v.SetDefault("database.type", "sqlite")
	v.SetDefault("database.sqlite_path", "./data/life-recorder.db")
	v.SetDefault("database.max_idle_conns", 10)
	v.SetDefault("database.max_open_conns", 100)
	v.SetDefault("database.log_level", "warn")
	v.SetDefault("jwt.access_ttl", 3600)
	v.SetDefault("jwt.refresh_ttl", 604800)
	v.SetDefault("jwt.issuer", "life-recorder")
	v.SetDefault("storage.default_adapter", "local")
	v.SetDefault("storage.local.base_path", "./data/uploads")
	v.SetDefault("storage.local.base_url", "/uploads")
	v.SetDefault("storage.max_upload_size", 104857600)
	v.SetDefault("ai_service.url", "http://localhost:8081")
	v.SetDefault("ai_service.timeout", 300)
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")
	v.SetDefault("log.output", "stdout")
	v.SetDefault("webhook.worker_count", 5)
	v.SetDefault("webhook.retry_delays", []int{5, 30, 120})
	v.SetDefault("setup.initialized", false)
}
