package features

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"ProxyWoman/internal/proxycore"
)

// ReplayRequest 重放请求结构
type ReplayRequest struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

// ReplayResponse 重放响应结构
type ReplayResponse struct {
	StatusCode int               `json:"statusCode"`
	Status     string            `json:"status"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	Duration   int64             `json:"duration"` // 毫秒
	Error      string            `json:"error,omitempty"`
}

// ReplayManager 重放管理器
type ReplayManager struct {
	client *http.Client
}

// NewReplayManager 创建重放管理器
func NewReplayManager() *ReplayManager {
	return &ReplayManager{
		client: &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				// 不自动跟随重定向
				return http.ErrUseLastResponse
			},
		},
	}
}

// ReplayFlow 重放Flow
func (rm *ReplayManager) ReplayFlow(flow *proxycore.Flow) (*ReplayResponse, error) {
	if flow.Request == nil {
		return nil, fmt.Errorf("flow has no request data")
	}

	replayReq := &ReplayRequest{
		Method:  flow.Request.Method,
		URL:     flow.Request.URL,
		Headers: flow.Request.Headers,
		Body:    string(flow.Request.Body),
	}

	return rm.SendRequest(replayReq)
}

// SendRequest 发送自定义请求
func (rm *ReplayManager) SendRequest(replayReq *ReplayRequest) (*ReplayResponse, error) {
	startTime := time.Now()

	// 解析URL
	parsedURL, err := url.Parse(replayReq.URL)
	if err != nil {
		return &ReplayResponse{
			Error: fmt.Sprintf("Invalid URL: %v", err),
		}, nil
	}

	// 创建请求体
	var bodyReader io.Reader
	if replayReq.Body != "" {
		bodyReader = strings.NewReader(replayReq.Body)
	}

	// 创建HTTP请求
	req, err := http.NewRequest(replayReq.Method, parsedURL.String(), bodyReader)
	if err != nil {
		return &ReplayResponse{
			Error: fmt.Sprintf("Failed to create request: %v", err),
		}, nil
	}

	// 设置请求头
	for name, value := range replayReq.Headers {
		// 跳过一些自动设置的头部
		if strings.ToLower(name) == "host" ||
			strings.ToLower(name) == "content-length" ||
			strings.ToLower(name) == "connection" {
			continue
		}
		req.Header.Set(name, value)
	}

	// 发送请求
	resp, err := rm.client.Do(req)
	duration := time.Since(startTime).Milliseconds()

	if err != nil {
		return &ReplayResponse{
			Duration: duration,
			Error:    fmt.Sprintf("Request failed: %v", err),
		}, nil
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return &ReplayResponse{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Duration:   duration,
			Error:      fmt.Sprintf("Failed to read response body: %v", err),
		}, nil
	}

	// 构建响应头
	headers := make(map[string]string)
	for name, values := range resp.Header {
		if len(values) > 0 {
			headers[name] = values[0]
		}
	}

	return &ReplayResponse{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Headers:    headers,
		Body:       string(respBody),
		Duration:   duration,
	}, nil
}

// ModifyAndSendRequest 修改并发送请求
func (rm *ReplayManager) ModifyAndSendRequest(originalFlow *proxycore.Flow, modifications map[string]interface{}) (*ReplayResponse, error) {
	if originalFlow.Request == nil {
		return nil, fmt.Errorf("flow has no request data")
	}

	// 创建修改后的请求
	replayReq := &ReplayRequest{
		Method:  originalFlow.Request.Method,
		URL:     originalFlow.Request.URL,
		Headers: make(map[string]string),
		Body:    string(originalFlow.Request.Body),
	}

	// 复制原始头部
	for name, value := range originalFlow.Request.Headers {
		replayReq.Headers[name] = value
	}

	// 应用修改
	if method, ok := modifications["method"].(string); ok {
		replayReq.Method = method
	}
	if url, ok := modifications["url"].(string); ok {
		replayReq.URL = url
	}
	if body, ok := modifications["body"].(string); ok {
		replayReq.Body = body
	}
	if headers, ok := modifications["headers"].(map[string]string); ok {
		for name, value := range headers {
			replayReq.Headers[name] = value
		}
	}

	return rm.SendRequest(replayReq)
}

// CreateRequestFromTemplate 从模板创建请求
func (rm *ReplayManager) CreateRequestFromTemplate(template *ReplayRequest) (*ReplayResponse, error) {
	return rm.SendRequest(template)
}

// ValidateRequest 验证请求
func (rm *ReplayManager) ValidateRequest(replayReq *ReplayRequest) error {
	// 验证方法
	validMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
	methodValid := false
	for _, method := range validMethods {
		if replayReq.Method == method {
			methodValid = true
			break
		}
	}
	if !methodValid {
		return fmt.Errorf("invalid HTTP method: %s", replayReq.Method)
	}

	// 验证URL
	_, err := url.Parse(replayReq.URL)
	if err != nil {
		return fmt.Errorf("invalid URL: %v", err)
	}

	return nil
}
