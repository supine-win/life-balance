// Package main LifeRecorder 网关服务入口
// 启动 HTTP API 服务，提供事件管理、媒体处理、认证等核心功能
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/life-recorder/gateway/internal/config"
	"github.com/life-recorder/gateway/internal/handler"
	"github.com/life-recorder/gateway/internal/handler/middleware"
	"github.com/life-recorder/gateway/internal/pkg/logger"
	"github.com/life-recorder/gateway/internal/repository"
	"github.com/life-recorder/gateway/internal/service"
	"github.com/life-recorder/gateway/internal/service/storage"
)

func main() {
	// 加载配置
	configPath := ""
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化日志
	logger.Init(cfg.Log.Level, cfg.Log.Format, cfg.Log.Output, cfg.Log.FilePath)
	logger.S().Info("LifeRecorder 网关服务启动中...")

	// 确保数据目录存在
	ensureDir(filepath.Dir(cfg.Database.SQLitePath))
	ensureDir(cfg.Storage.Local.BasePath)

	// 初始化数据库
	db, err := repository.InitDB(&cfg.Database)
	if err != nil {
		logger.S().Fatalw("数据库初始化失败", "error", err)
	}

	// 自动迁移
	if err := repository.AutoMigrate(db); err != nil {
		logger.S().Fatalw("数据库迁移失败", "error", err)
	}

	// 种子数据
	if err := seedInitialData(db); err != nil {
		logger.S().Warnw("种子数据初始化失败", "error", err)
	}

	// 初始化仓库
	userRepo := repository.NewUserRepo(db)
	roleRepo := repository.NewRoleRepo(db)
	eventRepo := repository.NewEventRepo(db)
	mediaRepo := repository.NewMediaRepo(db)
	metaRepo := repository.NewMediaMetadataRepo(db)

	// 初始化存储适配器
	storageAdapter, err := storage.CreateAdapter(cfg.Storage.DefaultAdapter, map[string]interface{}{
		"base_path": cfg.Storage.Local.BasePath,
		"base_url":  cfg.Storage.Local.BaseURL,
	})
	if err != nil {
		logger.S().Fatalw("存储适配器初始化失败", "error", err)
	}

	// 初始化服务
	authService := service.NewAuthService(userRepo, roleRepo, &cfg.JWT)
	metadataExtractor := service.NewMetadataExtractor(metaRepo, mediaRepo)
	mediaService := service.NewMediaService(mediaRepo, metaRepo, metadataExtractor, storageAdapter, cfg.Storage.MaxUploadSize)
	eventService := service.NewEventService(eventRepo, mediaRepo)

	// 初始化处理器
	authHandler := handler.NewAuthHandler(authService, cfg.JWT.Secret)
	eventHandler := handler.NewEventHandler(eventService)
	mediaHandler := handler.NewMediaHandler(mediaService)
	setupHandler := handler.NewSetupHandler(authService)
	systemHandler := handler.NewSystemHandler()

	// 设置 Gin
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())
	r.Use(middleware.RequestLogger())

	// 静态文件服务（上传目录）
	r.Static(cfg.Storage.Local.BaseURL, cfg.Storage.Local.BasePath)

	// ==================== 公开路由 ====================
	public := r.Group("/api/v1")
	{
		// 安装初始化
		setup := public.Group("/setup")
		{
			setup.GET("/status", setupHandler.Status)
			setup.POST("/init", setupHandler.Init)
		}

		// 认证
		auth := public.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.Refresh)
		}

		// 健康检查
		public.GET("/system/health", systemHandler.Health)
	}

	// ==================== 需要认证的路由 ====================
	authRequired := r.Group("/api/v1")
	authRequired.Use(middleware.AuthRequired(cfg.JWT.Secret))
	{
		// 认证相关
		authRequired.POST("/auth/logout", authHandler.Logout)
		authRequired.GET("/auth/me", authHandler.Me)
		authRequired.PUT("/auth/profile", authHandler.UpdateProfile)

		// 事件
		events := authRequired.Group("/events")
		{
			events.GET("", eventHandler.List)
			events.POST("", eventHandler.Create)
			events.GET("/calendar/:year/:month", eventHandler.CalendarMonth)
			events.GET("/:id", eventHandler.GetByID)
			events.PUT("/:id", eventHandler.Update)
			events.DELETE("/:id", eventHandler.Delete)
			events.POST("/confirm/:id", eventHandler.Confirm)
			events.POST("/:id/apply-suggestions", eventHandler.ApplySuggestions)
		}

		// 媒体
		mediaGroup := authRequired.Group("/media")
		{
			mediaGroup.GET("", mediaHandler.List)
			mediaGroup.POST("/upload", mediaHandler.Upload)
			mediaGroup.POST("/upload-with-metadata", mediaHandler.UploadWithMetadata)
			mediaGroup.GET("/:id", mediaHandler.GetByID)
			mediaGroup.GET("/:id/file", mediaHandler.GetFile)
			mediaGroup.GET("/:id/metadata", mediaHandler.GetMetadata)
			mediaGroup.DELETE("/:id", mediaHandler.Delete)
		}

		// Webhook（占位）
		webhooks := authRequired.Group("/webhooks")
		{
			_ = webhooks // TODO: 实现 Webhook handler
		}

		// 配置（占位）
		configGroup := authRequired.Group("/config")
		{
			_ = configGroup // TODO: 实现 Config handler
		}

		// 展示（占位）
		display := authRequired.Group("/display")
		{
			_ = display // TODO: 实现 Display handler
		}

		// AI 服务代理（占位）
		ai := authRequired.Group("/ai")
		{
			_ = ai // TODO: 代理到 Python AI 服务
		}

		// 管理（需要 admin 权限）
		admin := authRequired.Group("/admin")
		admin.Use(middleware.RequirePermission("admin:users"))
		{
			_ = admin // TODO: 实现 Admin handler
		}

		// 系统
		system := authRequired.Group("/system")
		{
			_ = system // TODO: 实现 Backup handler
		}
	}

	// 启动服务
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	logger.S().Infof("服务启动在 %s", addr)
	if err := r.Run(addr); err != nil {
		logger.S().Fatalw("服务启动失败", "error", err)
	}
}

// ensureDir 确保目录存在
func ensureDir(path string) {
	if path != "" && path != "." {
		os.MkdirAll(path, 0755)
	}
}

// seedInitialData 初始化种子数据
func seedInitialData(db interface{}) error {
	// 创建默认角色和权限
	// 在 AutoMigrate 后由种子脚本处理
	return nil
}
