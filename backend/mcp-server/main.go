// Package main LifeRecorder MCP Server
// 提供 Model Context Protocol 服务，支持 Agent 对接
//
// 工具列表：
//   - create_event: 创建事件
//   - query_events: 查询事件
//   - update_event: 更新事件
//   - delete_event: 删除事件
//   - search_events: 搜索事件
//   - generate_slideshow: 生成幻灯片
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// MCPServer MCP 服务器
type MCPServer struct {
	gatewayURL string
	client     *http.Client
}

// NewMCPServer 创建 MCP 服务器
func NewMCPServer(gatewayURL string) *MCPServer {
	return &MCPServer{
		gatewayURL: gatewayURL,
		client:     &http.Client{Timeout: 30 * time.Second},
	}
}

// Tool MCP 工具定义
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// ToolCall 工具调用
type ToolCall struct {
	Name   string                 `json:"name"`
	Input  map[string]interface{} `json:"input"`
}

// ToolResult 工具结果
type ToolResult struct {
	Content []ContentBlock `json:"content"`
	IsError bool           `json:"isError,omitempty"`
}

// ContentBlock 内容块
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// GetTools 返回可用工具列表
func (s *MCPServer) GetTools() []Tool {
	return []Tool{
		{
			Name:        "create_event",
			Description: "创建一个新的生活事件记录",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"title":       map[string]interface{}{"type": "string", "description": "事件标题"},
					"description": map[string]interface{}{"type": "string", "description": "事件描述"},
					"event_time":  map[string]interface{}{"type": "string", "description": "事件时间 (RFC3339)"},
					"location":    map[string]interface{}{"type": "string", "description": "地点"},
					"tags":        map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "description": "标签"},
					"mood":        map[string]interface{}{"type": "string", "description": "心情"},
					"category":    map[string]interface{}{"type": "string", "description": "分类"},
				},
				"required": []string{"title", "event_time"},
			},
		},
		{
			Name:        "query_events",
			Description: "查询生活事件记录",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"page":       map[string]interface{}{"type": "integer", "description": "页码"},
					"page_size":  map[string]interface{}{"type": "integer", "description": "每页数量"},
					"start_time": map[string]interface{}{"type": "string", "description": "开始时间"},
					"end_time":   map[string]interface{}{"type": "string", "description": "结束时间"},
					"category":   map[string]interface{}{"type": "string", "description": "分类筛选"},
					"keyword":    map[string]interface{}{"type": "string", "description": "关键词搜索"},
				},
			},
		},
		{
			Name:        "update_event",
			Description: "更新事件记录",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id":          map[string]interface{}{"type": "string", "description": "事件 ID"},
					"title":       map[string]interface{}{"type": "string", "description": "事件标题"},
					"description": map[string]interface{}{"type": "string", "description": "事件描述"},
					"event_time":  map[string]interface{}{"type": "string", "description": "事件时间"},
					"location":    map[string]interface{}{"type": "string", "description": "地点"},
					"tags":        map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
					"mood":        map[string]interface{}{"type": "string"},
					"category":    map[string]interface{}{"type": "string"},
				},
				"required": []string{"id"},
			},
		},
		{
			Name:        "delete_event",
			Description: "删除事件记录",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{"type": "string", "description": "事件 ID"},
				},
				"required": []string{"id"},
			},
		},
		{
			Name:        "search_events",
			Description: "搜索事件记录",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"keyword":    map[string]interface{}{"type": "string", "description": "搜索关键词"},
					"start_time": map[string]interface{}{"type": "string"},
					"end_time":   map[string]interface{}{"type": "string"},
					"tags":       map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
					"page":       map[string]interface{}{"type": "integer"},
					"page_size":  map[string]interface{}{"type": "integer"},
				},
				"required": []string{"keyword"},
			},
		},
		{
			Name:        "generate_slideshow",
			Description: "生成幻灯片展示",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"title":    map[string]interface{}{"type": "string", "description": "幻灯片标题"},
					"event_ids": map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
					"style":    map[string]interface{}{"type": "string", "description": "展示风格"},
				},
				"required": []string{"title"},
			},
		},
	}
}

// CallTool 调用工具
func (s *MCPServer) CallTool(ctx context.Context, call ToolCall) (*ToolResult, error) {
	switch call.Name {
	case "create_event":
		return s.callCreateEvent(ctx, call.Input)
	case "query_events":
		return s.callQueryEvents(ctx, call.Input)
	case "update_event":
		return s.callUpdateEvent(ctx, call.Input)
	case "delete_event":
		return s.callDeleteEvent(ctx, call.Input)
	case "search_events":
		return s.callSearchEvents(ctx, call.Input)
	case "generate_slideshow":
		return s.callGenerateSlideshow(ctx, call.Input)
	default:
		return &ToolResult{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("未知工具: %s", call.Name)}},
			IsError: true,
		}, nil
	}
}

// callCreateEvent 调用创建事件 API
func (s *MCPServer) callCreateEvent(ctx context.Context, input map[string]interface{}) (*ToolResult, error) {
	resp, err := s.doAPIRequest(ctx, "POST", "/api/v1/events", input)
	if err != nil {
		return &ToolResult{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("创建事件失败: %v", err)}},
			IsError: true,
		}, nil
	}
	text, _ := json.MarshalIndent(resp, "", "  ")
	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: string(text)}},
	}, nil
}

// callQueryEvents 调用查询事件 API
func (s *MCPServer) callQueryEvents(ctx context.Context, input map[string]interface{}) (*ToolResult, error) {
	resp, err := s.doAPIRequest(ctx, "GET", "/api/v1/events", input)
	if err != nil {
		return &ToolResult{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("查询事件失败: %v", err)}},
			IsError: true,
		}, nil
	}
	text, _ := json.MarshalIndent(resp, "", "  ")
	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: string(text)}},
	}, nil
}

// callUpdateEvent 调用更新事件 API
func (s *MCPServer) callUpdateEvent(ctx context.Context, input map[string]interface{}) (*ToolResult, error) {
	id, _ := input["id"].(string)
	if id == "" {
		return &ToolResult{
			Content: []ContentBlock{{Type: "text", Text: "缺少事件 ID"}},
			IsError: true,
		}, nil
	}
	delete(input, "id")
	resp, err := s.doAPIRequest(ctx, "PUT", fmt.Sprintf("/api/v1/events/%s", id), input)
	if err != nil {
		return &ToolResult{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("更新事件失败: %v", err)}},
			IsError: true,
		}, nil
	}
	text, _ := json.MarshalIndent(resp, "", "  ")
	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: string(text)}},
	}, nil
}

// callDeleteEvent 调用删除事件 API
func (s *MCPServer) callDeleteEvent(ctx context.Context, input map[string]interface{}) (*ToolResult, error) {
	id, _ := input["id"].(string)
	if id == "" {
		return &ToolResult{
			Content: []ContentBlock{{Type: "text", Text: "缺少事件 ID"}},
			IsError: true,
		}, nil
	}
	resp, err := s.doAPIRequest(ctx, "DELETE", fmt.Sprintf("/api/v1/events/%s", id), nil)
	if err != nil {
		return &ToolResult{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("删除事件失败: %v", err)}},
			IsError: true,
		}, nil
	}
	text, _ := json.MarshalIndent(resp, "", "  ")
	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: string(text)}},
	}, nil
}

// callSearchEvents 调用搜索 API
func (s *MCPServer) callSearchEvents(ctx context.Context, input map[string]interface{}) (*ToolResult, error) {
	resp, err := s.doAPIRequest(ctx, "POST", "/api/v1/events/search", input)
	if err != nil {
		return &ToolResult{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("搜索事件失败: %v", err)}},
			IsError: true,
		}, nil
	}
	text, _ := json.MarshalIndent(resp, "", "  ")
	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: string(text)}},
	}, nil
}

// callGenerateSlideshow 调用生成幻灯片 API
func (s *MCPServer) callGenerateSlideshow(ctx context.Context, input map[string]interface{}) (*ToolResult, error) {
	resp, err := s.doAPIRequest(ctx, "POST", "/api/v1/display/slideshow", input)
	if err != nil {
		return &ToolResult{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("生成幻灯片失败: %v", err)}},
			IsError: true,
		}, nil
	}
	text, _ := json.MarshalIndent(resp, "", "  ")
	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: string(text)}},
	}, nil
}

// doAPIRequest 发送 API 请求到网关
func (s *MCPServer) doAPIRequest(ctx context.Context, method, path string, body interface{}) (interface{}, error) {
	// TODO: 实现 HTTP 请求到 Go 网关
	return map[string]interface{}{
		"message": "MCP Server 调用成功（占位实现）",
		"method":  method,
		"path":    path,
	}, nil
}

func main() {
	gatewayURL := os.Getenv("GATEWAY_URL")
	if gatewayURL == "" {
		gatewayURL = "http://localhost:8080"
	}

	port := os.Getenv("MCP_PORT")
	if port == "" {
		port = "8082"
	}

	server := NewMCPServer(gatewayURL)

	// MCP 协议 HTTP 端点
	http.HandleFunc("/tools", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(server.GetTools())
	})

	http.HandleFunc("/call", func(w http.ResponseWriter, r *http.Request) {
		var call ToolCall
		if err := json.NewDecoder(r.Body).Decode(&call); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		result, err := server.CallTool(r.Context(), call)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	})

	log.Printf("MCP Server 启动在 :%s, 网关: %s", port, gatewayURL)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
