// Package service 事件服务
package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/life-recorder/gateway/internal/model"
	"github.com/life-recorder/gateway/internal/repository"
)

// EventService 事件服务
type EventService struct {
	eventRepo *repository.EventRepo
	mediaRepo *repository.MediaRepo
}

// NewEventService 创建事件服务
func NewEventService(eventRepo *repository.EventRepo, mediaRepo *repository.MediaRepo) *EventService {
	return &EventService{
		eventRepo: eventRepo,
		mediaRepo: mediaRepo,
	}
}

// CreateInput 创建事件输入
type CreateInput struct {
	Title        string   `json:"title" binding:"required,max=255"`
	Description  string   `json:"description"`
	EventTime    string   `json:"event_time" binding:"required"`
	EndTime      string   `json:"end_time"`
	Location     string   `json:"location"`
	Latitude     *float64 `json:"latitude"`
	Longitude    *float64 `json:"longitude"`
	Mood         string   `json:"mood"`
	Tags         []string `json:"tags"`
	Participants []string `json:"participants"`
	Category     string   `json:"category"`
	Color        string   `json:"color"`
	Priority     int      `json:"priority"`
	MediaIDs     []string `json:"media_ids"`
	IsConfirmed  bool     `json:"is_confirmed"`
	Source       string   `json:"source"`
	SourceData   map[string]interface{} `json:"source_data"`
}

// Create 创建事件
func (s *EventService) Create(userID string, input *CreateInput) (*model.Event, error) {
	eventTime, err := parseTime(input.EventTime)
	if err != nil {
		return nil, errors.New("无效的事件时间格式")
	}
	eventTimeTyped := eventTime.(time.Time)

	event := &model.Event{
		UserID:      userID,
		Title:       input.Title,
		Description: input.Description,
		EventTime:   eventTimeTyped,
		Location:    input.Location,
		Latitude:    input.Latitude,
		Longitude:   input.Longitude,
		Mood:        input.Mood,
		Tags:        model.StringArray(input.Tags),
		Participants: model.StringArray(input.Participants),
		Category:    input.Category,
		Color:       input.Color,
		Priority:    input.Priority,
		IsConfirmed: input.IsConfirmed,
		Source:      input.Source,
		SourceData:  model.JSONString(input.SourceData),
		IsDeleted:   false,
	}

	var endTimeTyped time.Time
	if input.EndTime != "" {
		endTimeVal, err := parseTime(input.EndTime)
		if err == nil {
			endTimeTyped = endTimeVal.(time.Time)
			event.EndTime = &endTimeTyped
		}
	}

	if event.Source == "" {
		event.Source = "text"
	}

	if err := s.eventRepo.Create(event); err != nil {
		return nil, fmt.Errorf("创建事件失败: %w", err)
	}

	// 关联媒体
	if len(input.MediaIDs) > 0 {
		for i, mediaID := range input.MediaIDs {
			media, err := s.mediaRepo.FindByID(mediaID, userID)
			if err == nil {
				media.EventID = &event.ID
				media.SortOrder = i
				s.mediaRepo.Update(media)
			}
		}
	}

	return s.eventRepo.FindByID(event.ID, userID)
}

// GetByID 获取事件详情
func (s *EventService) GetByID(id, userID string) (*model.Event, error) {
	return s.eventRepo.FindByID(id, userID)
}

// Update 更新事件
func (s *EventService) Update(id, userID string, input *CreateInput) (*model.Event, error) {
	event, err := s.eventRepo.FindByID(id, userID)
	if err != nil {
		return nil, errors.New("事件不存在")
	}

	event.Title = input.Title
	event.Description = input.Description
	event.Location = input.Location
	event.Latitude = input.Latitude
	event.Longitude = input.Longitude
	event.Mood = input.Mood
	event.Tags = model.StringArray(input.Tags)
	event.Participants = model.StringArray(input.Participants)
	event.Category = input.Category
	event.Color = input.Color
	event.Priority = input.Priority

	if input.EventTime != "" {
		if tVal, err := parseTime(input.EventTime); err == nil {
			event.EventTime = tVal.(time.Time)
		}
	}
	if input.EndTime != "" {
		if tVal, err := parseTime(input.EndTime); err == nil {
			tTyped := tVal.(time.Time)
			event.EndTime = &tTyped
		}
	}

	if err := s.eventRepo.Update(event); err != nil {
		return nil, fmt.Errorf("更新事件失败: %w", err)
	}

	return event, nil
}

// Delete 删除事件
func (s *EventService) Delete(id, userID string) error {
	return s.eventRepo.Delete(id, userID)
}

// List 查询事件列表
func (s *EventService) List(userID string, page, pageSize int, filters map[string]interface{}) ([]model.Event, int64, error) {
	return s.eventRepo.List(userID, page, pageSize, filters)
}

// Confirm 确认事件
func (s *EventService) Confirm(id, userID string, modifications map[string]interface{}) (*model.Event, error) {
	event, err := s.eventRepo.FindByID(id, userID)
	if err != nil {
		return nil, errors.New("事件不存在")
	}

	// 应用修改
	if title, ok := modifications["title"].(string); ok && title != "" {
		event.Title = title
	}
	if desc, ok := modifications["description"].(string); ok {
		event.Description = desc
	}

	event.IsConfirmed = true
	if err := s.eventRepo.Update(event); err != nil {
		return nil, fmt.Errorf("确认事件失败: %w", err)
	}
	return event, nil
}

// CalendarMonth 获取月份日历数据
func (s *EventService) CalendarMonth(userID string, year, month int) ([]model.Event, error) {
	return s.eventRepo.CalendarMonth(userID, year, month)
}

// ApplySuggestions 将元数据建议应用到事件
func (s *EventService) ApplySuggestions(eventID, userID string, mediaID string, fields []string) (*model.Event, error) {
	event, err := s.eventRepo.FindByID(eventID, userID)
	if err != nil {
		return nil, errors.New("事件不存在")
	}

	// 获取媒体元数据
	media, err := s.mediaRepo.FindByID(mediaID, userID)
	if err != nil || media.MediaMetadata == nil {
		return nil, errors.New("媒体元数据不存在")
	}

	meta := media.MediaMetadata
	suggestions := meta.Suggestions

	for _, field := range fields {
		switch field {
		case "event_time":
			if meta.ShotTime != nil {
				event.EventTime = *meta.ShotTime
			}
		case "location":
			if meta.LocationName != "" {
				event.Location = meta.LocationName
				if meta.GPSLatitude != nil {
					event.Latitude = meta.GPSLatitude
				}
				if meta.GPSLongitude != nil {
					event.Longitude = meta.GPSLongitude
				}
			}
		case "tags":
			if tags, ok := suggestions["tags"].([]interface{}); ok {
				var tagStrs []string
				for _, t := range tags {
					if s, ok := t.(string); ok {
						tagStrs = append(tagStrs, s)
					}
				}
				// 合并标签，去重
				existingTags := make(map[string]bool)
				for _, t := range event.Tags {
					existingTags[t] = true
				}
				for _, t := range tagStrs {
					if !existingTags[t] {
						event.Tags = append(event.Tags, t)
					}
				}
			}
		}
	}

	if err := s.eventRepo.Update(event); err != nil {
		return nil, fmt.Errorf("应用建议失败: %w", err)
	}
	return event, nil
}

// parseTime 解析时间字符串
func parseTime(s string) (interface{}, error) {
	// 尝试多种时间格式
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return nil, fmt.Errorf("无法解析时间: %s", s)
}
