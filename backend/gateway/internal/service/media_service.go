// Package service 媒体服务
package service

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/life-recorder/gateway/internal/model"
	"github.com/life-recorder/gateway/internal/repository"
	pkguuid "github.com/life-recorder/gateway/internal/pkg/uuid"
	"github.com/life-recorder/gateway/internal/service/storage"
)

// MediaService 媒体服务
type MediaService struct {
	mediaRepo    *repository.MediaRepo
	metaRepo     *repository.MediaMetadataRepo
	extractor    *MetadataExtractor
	storageAdapter storage.Adapter
	maxUploadSize int64
}

// NewMediaService 创建媒体服务
func NewMediaService(
	mediaRepo *repository.MediaRepo,
	metaRepo *repository.MediaMetadataRepo,
	extractor *MetadataExtractor,
	storageAdapter storage.Adapter,
	maxUploadSize int64,
) *MediaService {
	return &MediaService{
		mediaRepo:     mediaRepo,
		metaRepo:      metaRepo,
		extractor:     extractor,
		storageAdapter: storageAdapter,
		maxUploadSize: maxUploadSize,
	}
}

// UploadResult 上传结果
type UploadResult struct {
	Media     *model.Media          `json:"media"`
	Metadata  *model.MediaMetadata  `json:"metadata,omitempty"`
	Suggestions *model.MetadataSuggestions `json:"suggestions,omitempty"`
}

// Upload 上传媒体文件
func (s *MediaService) Upload(ctx context.Context, userID string, file *multipart.FileHeader, eventID *string) (*model.Media, error) {
	if file.Size > s.maxUploadSize {
		return nil, fmt.Errorf("文件大小超过限制 (%d bytes)", s.maxUploadSize)
	}

	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer src.Close()

	// 生成存储路径
	ext := strings.ToLower(filepath.Ext(file.Filename))
	mediaType := detectMediaType(ext)
	datePrefix := time.Now().Format("2006/01/02")
	storageKey := fmt.Sprintf("%s/%s/%s%s", userID, datePrefix, pkguuid.New(), ext)

	// 保存文件
	if err := s.storageAdapter.Save(ctx, storageKey, src, file.Size, file.Header.Get("Content-Type")); err != nil {
		return nil, fmt.Errorf("保存文件失败: %w", err)
	}

	// 获取文件 URL
	fileURL, _ := s.storageAdapter.GetURL(ctx, storageKey)

	// 计算文件哈希
	tmpPath := "" // 本地存储时可以直接读文件
	_ = tmpPath

	// 创建数据库记录
	media := &model.Media{
		ID:             pkguuid.New(),
		UserID:         userID,
		EventID:        eventID,
		Type:           mediaType,
		StorageAdapter: "local",
		StoragePath:    storageKey,
		OriginalName:   file.Filename,
		MimeType:       detectMimeType(ext),
		Size:           file.Size,
	}

	if err := s.mediaRepo.Create(media); err != nil {
		// 清理已上传的文件
		s.storageAdapter.Delete(ctx, storageKey)
		return nil, fmt.Errorf("创建媒体记录失败: %w", err)
	}

	_ = fileURL
	return media, nil
}

// UploadWithMetadata 上传并返回元数据提取结果
func (s *MediaService) UploadWithMetadata(ctx context.Context, userID string, file *multipart.FileHeader, eventID *string) (*UploadResult, error) {
	// 先上传文件
	media, err := s.Upload(ctx, userID, file, eventID)
	if err != nil {
		return nil, err
	}

	// 异步提取元数据（简化实现：同步等待）
	// 在本地存储模式下可以直接通过路径访问文件
	filePath := media.StoragePath // 本地存储时这会是相对路径

	// 启动元数据提取
	go func() {
		bgCtx := context.Background()
		if err := s.extractor.ExtractAndSuggest(bgCtx, media.ID, filePath, media.OriginalName); err != nil {
			// 日志记录错误，不影响上传结果
			_ = err
		}
	}()

	return &UploadResult{
		Media: media,
	}, nil
}

// GetByID 获取媒体详情（含元数据）
func (s *MediaService) GetByID(id, userID string) (*model.Media, error) {
	return s.mediaRepo.FindByID(id, userID)
}

// GetMetadata 获取媒体元数据
func (s *MediaService) GetMetadata(mediaID string) (*model.MediaMetadata, error) {
	return s.metaRepo.FindByMediaID(mediaID)
}

// Delete 删除媒体
func (s *MediaService) Delete(ctx context.Context, id, userID string) error {
	media, err := s.mediaRepo.FindByID(id, userID)
	if err != nil {
		return fmt.Errorf("媒体不存在")
	}

	// 删除存储文件
	if err := s.storageAdapter.Delete(ctx, media.StoragePath); err != nil {
		// 记录但不阻止删除记录
		_ = err
	}

	// 删除缩略图
	if media.ThumbnailPath != "" {
		s.storageAdapter.Delete(ctx, media.ThumbnailPath)
	}

	return s.mediaRepo.Delete(id, userID)
}

// List 媒体列表
func (s *MediaService) List(userID string, page, pageSize int, mediaType string) ([]model.Media, int64, error) {
	return s.mediaRepo.List(userID, page, pageSize, mediaType)
}

// GetFile 获取文件内容
func (s *MediaService) GetFile(ctx context.Context, id, userID string) (io.ReadCloser, *model.Media, error) {
	media, err := s.mediaRepo.FindByID(id, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("媒体不存在")
	}

	reader, err := s.storageAdapter.Get(ctx, media.StoragePath)
	if err != nil {
		return nil, nil, fmt.Errorf("获取文件失败: %w", err)
	}

	return reader, media, nil
}

// detectMediaType 检测媒体类型
func detectMediaType(ext string) string {
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".tiff", ".tif", ".heic", ".heif":
		return "image"
	case ".mp4", ".mov", ".avi", ".mkv", ".wmv", ".flv", ".webm", ".m4v", ".3gp":
		return "video"
	case ".mp3", ".wav", ".ogg", ".flac", ".aac", ".m4a", ".wma":
		return "audio"
	default:
		return "document"
	}
}

// detectMimeType 检测 MIME 类型
func detectMimeType(ext string) string {
	mimes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".bmp":  "image/bmp",
		".tiff": "image/tiff",
		".tif":  "image/tiff",
		".heic": "image/heic",
		".heif": "image/heif",
		".mp4":  "video/mp4",
		".mov":  "video/quicktime",
		".avi":  "video/x-msvideo",
		".mkv":  "video/x-matroska",
		".webm": "video/webm",
		".m4v":  "video/x-m4v",
		".mp3":  "audio/mpeg",
		".wav":  "audio/wav",
		".ogg":  "audio/ogg",
		".flac": "audio/flac",
		".aac":  "audio/aac",
		".m4a":  "audio/mp4",
	}
	if mime, ok := mimes[ext]; ok {
		return mime
	}
	return "application/octet-stream"
}
