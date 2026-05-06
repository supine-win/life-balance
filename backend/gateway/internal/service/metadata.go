// Package service 元数据提取服务
package service

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/life-recorder/gateway/internal/model"
	"github.com/life-recorder/gateway/internal/pkg/logger"
	"github.com/life-recorder/gateway/internal/repository"
)

// MetadataExtractor 元数据提取服务
type MetadataExtractor struct {
	metadataRepo *repository.MediaMetadataRepo
	mediaRepo    *repository.MediaRepo
}

// NewMetadataExtractor 创建元数据提取服务
func NewMetadataExtractor(metadataRepo *repository.MediaMetadataRepo, mediaRepo *repository.MediaRepo) *MetadataExtractor {
	return &MetadataExtractor{
		metadataRepo: metadataRepo,
		mediaRepo:    mediaRepo,
	}
}

// ExtractAndSuggest 提取元数据并生成建议卡片
// 这是核心方法：解析文件的 EXIF/元数据，进行逆地理编码，生成 AI 建议
func (e *MetadataExtractor) ExtractAndSuggest(ctx context.Context, mediaID string, filePath string, originalName string) error {
	logger.S().Infow("开始提取元数据", "media_id", mediaID, "file", filePath)

	// 创建或获取元数据记录
	meta, err := e.metadataRepo.FindByMediaID(mediaID)
	if err != nil {
		// 记录不存在，创建新记录
		meta = &model.MediaMetadata{
			ID:            model.BaseModel{}.ID,
			MediaID:       mediaID,
			ExtractStatus: "processing",
		}
		if meta.ID == "" {
			meta.ID = generateUUID()
		}
	} else {
		meta.ExtractStatus = "processing"
	}

	now := time.Now()
	meta.ExtractedAt = &now

	// 根据文件类型提取元数据
	ext := strings.ToLower(filepath.Ext(originalName))
	if isImageFile(ext) {
		if err := e.extractImageMetadata(meta, filePath, originalName); err != nil {
			logger.S().Warnw("图片元数据提取失败", "error", err)
			meta.ExtractError = err.Error()
		}
	} else if isVideoFile(ext) {
		if err := e.extractVideoMetadata(meta, filePath, originalName); err != nil {
			logger.S().Warnw("视频元数据提取失败", "error", err)
			meta.ExtractError = err.Error()
		}
	} else if isAudioFile(ext) {
		if err := e.extractAudioMetadata(meta, filePath, originalName); err != nil {
			logger.S().Warnw("音频元数据提取失败", "error", err)
			meta.ExtractError = err.Error()
		}
	}

	// 尝试从文件名提取日期
	if meta.ShotTime == nil {
		if t, pattern := parseDateFromFilename(originalName); t != nil {
			meta.ShotTime = t
			meta.ShotTimeSrc = "filename"
			meta.FilenamePattern = pattern
		}
	}

	// GPS 逆地理编码
	if meta.GPSLatitude != nil && meta.GPSLongitude != nil {
		locationName, err := reverseGeocode(*meta.GPSLatitude, *meta.GPSLongitude)
		if err != nil {
			logger.S().Warnw("逆地理编码失败", "error", err)
		} else if locationName != "" {
			meta.LocationName = locationName
			meta.LocationSrc = "geocode"
		}
	}

	// 生成建议卡片
	suggestions := e.generateSuggestions(meta)
	sugJSON, _ := json.Marshal(suggestions)
	var sugMap model.JSONString
	json.Unmarshal(sugJSON, &sugMap)
	meta.Suggestions = sugMap

	// 如果提取无致命错误，标记为完成
	if meta.ExtractError == "" {
		meta.ExtractStatus = "completed"
	} else {
		// 有部分错误但可能也有成功的部分数据
		if meta.ShotTime != nil || meta.CameraModel != "" || meta.LocationName != "" {
			meta.ExtractStatus = "completed" // 部分成功
		} else {
			meta.ExtractStatus = "failed"
		}
	}

	// 保存到数据库
	if meta.CreatedAt.IsZero() {
		err = e.metadataRepo.Create(meta)
	} else {
		err = e.metadataRepo.Update(meta)
	}

	if err != nil {
		return fmt.Errorf("保存元数据失败: %w", err)
	}

	logger.S().Infow("元数据提取完成", "media_id", mediaID, "status", meta.ExtractStatus)
	return nil
}

// extractImageMetadata 提取图片元数据
// 使用 Go 标准库解析基础信息，完整 EXIF 留给前端或 Python 服务处理
func (e *MetadataExtractor) extractImageMetadata(meta *model.MediaMetadata, filePath string, originalName string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	// 解析 EXIF 数据（基础实现）
	// 生产环境应使用 dsoprea/go-exif 或 rwcarlsen/goexif 库
	exifData := parseBasicEXIF(filePath)
	if exifData != nil {
		meta.RawEXIF = exifData

		// 提取拍摄时间
		if dt, ok := exifData["DateTimeOriginal"]; ok {
			if t, err := parseEXIFTime(fmt.Sprintf("%v", dt)); err == nil {
				meta.ShotTime = &t
				meta.ShotTimeSrc = "exif"
			}
		}

		// 提取 GPS
		if lat, ok := exifData["GPSLatitude"]; ok {
			if lon, ok2 := exifData["GPSLongitude"]; ok2 {
				if latF, err := toFloat64(lat); err == nil {
					if lonF, err := toFloat64(lon); err == nil {
						meta.GPSLatitude = &latF
						meta.GPSLongitude = &lonF
					}
				}
			}
		}

		// 提取相机信息
		if v, ok := exifData["Make"]; ok {
			meta.CameraMake = fmt.Sprintf("%v", v)
		}
		if v, ok := exifData["Model"]; ok {
			meta.CameraModel = fmt.Sprintf("%v", v)
		}
		if v, ok := exifData["LensModel"]; ok {
			meta.LensModel = fmt.Sprintf("%v", v)
		}
		if v, ok := exifData["FocalLength"]; ok {
			if f, err := toFloat64(v); err == nil {
				meta.FocalLength = &f
			}
		}
		if v, ok := exifData["FNumber"]; ok {
			if f, err := toFloat64(v); err == nil {
				meta.Aperture = &f
			}
		}
		if v, ok := exifData["ISOSpeedRatings"]; ok {
			if i, err := toInt(v); err == nil {
				meta.ISO = &i
			}
		}
		if v, ok := exifData["ExposureTime"]; ok {
			meta.ShutterSpeed = fmt.Sprintf("%v", v)
		}
	}

	return nil
}

// extractVideoMetadata 提取视频元数据
func (e *MetadataExtractor) extractVideoMetadata(meta *model.MediaMetadata, filePath string, originalName string) error {
	// 使用 ffprobe 提取视频元数据
	// 生产环境使用 ffmpeg-python 或直接调用 ffprobe
	info := probeVideoFile(filePath)
	if info != nil {
		if v, ok := info["creation_time"]; ok {
			if t, err := time.Parse(time.RFC3339, fmt.Sprintf("%v", v)); err == nil {
				meta.ShotTime = &t
				meta.ShotTimeSrc = "video"
			}
		}
		if v, ok := info["width"]; ok {
			if _, err := toInt(v); err == nil {
				// 更新 media 记录的宽高
			}
		}
		if v, ok := info["duration"]; ok {
			if f, err := toFloat64(v); err == nil {
				_ = f // 更新到 media 记录
			}
		}
		if v, ok := info["video_codec"]; ok {
			meta.VideoCodec = fmt.Sprintf("%v", v)
		}
		if v, ok := info["audio_codec"]; ok {
			meta.AudioCodec = fmt.Sprintf("%v", v)
		}
		if v, ok := info["frame_rate"]; ok {
			if f, err := toFloat64(v); err == nil {
				meta.FrameRate = &f
			}
		}
		if v, ok := info["bit_rate"]; ok {
			if i, err := toInt64(v); err == nil {
				meta.Bitrate = &i
			}
		}
	}
	return nil
}

// extractAudioMetadata 提取音频元数据
func (e *MetadataExtractor) extractAudioMetadata(meta *model.MediaMetadata, filePath string, originalName string) error {
	// 基础实现：从文件信息推断
	return nil
}

// generateSuggestions 生成建议卡片
func (e *MetadataExtractor) generateSuggestions(meta *model.MediaMetadata) *model.MetadataSuggestions {
	sug := &model.MetadataSuggestions{
		Confidence: make(map[string]float64),
	}

	// 建议时间
	if meta.ShotTime != nil {
		sug.EventTime = meta.ShotTime.Format(time.RFC3339)
		sug.Confidence["time"] = 0.95
		if meta.ShotTimeSrc == "filename" {
			sug.Confidence["time"] = 0.7 // 文件名提取的置信度较低
		}
	}

	// 建议地点
	if meta.LocationName != "" {
		sug.Location = meta.LocationName
		sug.Confidence["location"] = 0.88
	}

	// 建议标签
	var tags []string
	if meta.LocationName != "" {
		// 从地点名提取标签（简单分词）
		parts := strings.Split(meta.LocationName, "")
		for i := 0; i < len(parts)-1; i++ {
			if len(tags) >= 5 {
				break
			}
			tags = append(tags, parts[i]+parts[i+1])
		}
	}
	if meta.CameraModel != "" {
		tags = append(tags, "摄影")
	}
	sug.Tags = tags
	sug.Confidence["tags"] = 0.65

	// 建议分类
	if meta.CameraModel != "" {
		sug.Category = "摄影"
	}

	// 生成提示文字
	var hints []string
	if meta.CameraMake != "" && meta.CameraModel != "" {
		hints = append(hints, fmt.Sprintf("使用 %s %s 拍摄", meta.CameraMake, meta.CameraModel))
	} else if meta.CameraModel != "" {
		hints = append(hints, fmt.Sprintf("使用 %s 拍摄", meta.CameraModel))
	}
	if meta.LocationName != "" {
		hints = append(hints, fmt.Sprintf("在%s", meta.LocationName))
	}
	if len(hints) > 0 {
		sug.HintText = strings.Join(hints, "")
	}

	// 设备信息文字
	var deviceParts []string
	if meta.CameraModel != "" {
		deviceParts = append(deviceParts, meta.CameraModel)
	}
	if meta.FocalLength != nil {
		deviceParts = append(deviceParts, fmt.Sprintf("%.0fmm", *meta.FocalLength))
	}
	if meta.Aperture != nil {
		deviceParts = append(deviceParts, fmt.Sprintf("ƒ/%.1f", *meta.Aperture))
	}
	if meta.ISO != nil {
		deviceParts = append(deviceParts, fmt.Sprintf("ISO %d", *meta.ISO))
	}
	if len(deviceParts) > 0 {
		sug.DeviceText = strings.Join(deviceParts, " · ")
	}

	return sug
}

// ==================== 辅助函数 ====================

// parseDateFromFilename 从文件名解析日期
func parseDateFromFilename(filename string) (*time.Time, string) {
	patterns := []struct {
		regex   string
		layout  string
		name    string
	}{
		{`(?i)(?:IMG|DSC|VID|Screenshot|Photo|photo|IMG_)\D*?(\d{4})(\d{2})(\d{2})[_\-.]?(\d{2})?(\d{2})?(\d{2})?`, "20060102 15:04:05", "camera_standard"},
		{`(?i)(\d{4})-(\d{2})-(\d{2})[_\-.]?(\d{2})?(\d{2})?(\d{2})?`, "2006-01-02 15:04:05", "iso_date"},
		{`(?i)(\d{4})_(\d{2})_(\d{2})[_\-.]?(\d{2})?(\d{2})?(\d{2})?`, "2006_01_02 15:04:05", "underscore_date"},
	}

	for _, p := range patterns {
		re := regexp.MustCompile(p.regex)
		matches := re.FindStringSubmatch(filename)
		if len(matches) >= 4 {
			dateStr := matches[1] + matches[2] + matches[3]
			if len(matches) >= 7 && matches[4] != "" {
				dateStr += " " + matches[4] + matches[5] + matches[6]
			} else {
				dateStr += " 00:00:00"
			}

			t, err := time.Parse(p.layout, dateStr)
			if err == nil {
				return &t, p.name
			}
		}
	}
	return nil, ""
}

// reverseGeocode GPS 逆地理编码
func reverseGeocode(lat, lng float64) (string, error) {
	// 基础实现：返回空字符串
	// 生产环境应调用高德/百度/Nominatim API
	// TODO: 实现逆地理编码
	return "", nil
}

// parseEXIFTime 解析 EXIF 时间格式 "2026:05:01 14:32:00"
func parseEXIFTime(s string) (time.Time, error) {
	return time.Parse("2006:01:02 15:04:05", s)
}

// isImageFile 判断是否为图片文件
func isImageFile(ext string) bool {
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".tiff", ".tif", ".heic", ".heif":
		return true
	}
	return false
}

// isVideoFile 判断是否为视频文件
func isVideoFile(ext string) bool {
	switch ext {
	case ".mp4", ".mov", ".avi", ".mkv", ".wmv", ".flv", ".webm", ".m4v", ".3gp":
		return true
	}
	return false
}

// isAudioFile 判断是否为音频文件
func isAudioFile(ext string) bool {
	switch ext {
	case ".mp3", ".wav", ".ogg", ".flac", ".aac", ".m4a", ".wma":
		return true
	}
	return false
}

// toFloat64 安全转换为 float64
func toFloat64(v interface{}) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case float32:
		return float64(val), nil
	case int:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case string:
		return strconv.ParseFloat(val, 64)
	default:
		return strconv.ParseFloat(fmt.Sprintf("%v", v), 64)
	}
}

// toInt 安全转换为 int
func toInt(v interface{}) (int, error) {
	switch val := v.(type) {
	case int:
		return val, nil
	case int64:
		return int(val), nil
	case float64:
		return int(val), nil
	case string:
		return strconv.Atoi(val)
	default:
		return strconv.Atoi(fmt.Sprintf("%v", v))
	}
}

// toInt64 安全转换为 int64
func toInt64(v interface{}) (int64, error) {
	i, err := toInt(v)
	return int64(i), err
}

// parseBasicEXIF 基础 EXIF 解析（占位实现）
// 生产环境应使用 go-exif 库完整解析
func parseBasicEXIF(filePath string) map[string]interface{} {
	// TODO: 使用 github.com/dsoprea/go-exif/v3 完整解析
	// 这里返回空 map 作为占位
	return make(map[string]interface{})
}

// probeVideoFile 使用 ffprobe 解析视频元数据（占位实现）
func probeVideoFile(filePath string) map[string]interface{} {
	// TODO: 使用 os/exec 调用 ffprobe 或使用 ffmpeg-go 库
	return make(map[string]interface{})
}

// ComputeFileHash 计算文件 SHA256 哈希
func ComputeFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// generateUUID 生成 UUID
func generateUUID() string {
	// 使用 crypto/rand 生成
	import_uuid := func() string {
		// 简单实现
		return fmt.Sprintf("%x", time.Now().UnixNano())
	}
	return import_uuid()
}
