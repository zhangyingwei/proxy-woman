package proxycore

import (
	"net/http"
	"time"
)

// Flow 表示一个完整的HTTP请求/响应流
type Flow struct {
	ID           string            `json:"id"`
	URL          string            `json:"url"`
	Method       string            `json:"method"`
	StatusCode   int               `json:"statusCode"`
	Client       string            `json:"client"`
	Domain       string            `json:"domain"`
	Path         string            `json:"path"`
	Scheme       string            `json:"scheme"`
	StartTime    time.Time         `json:"startTime"`
	EndTime      time.Time         `json:"endTime"`
	Duration     time.Duration     `json:"duration"`
	RequestSize  int64             `json:"requestSize"`
	ResponseSize int64             `json:"responseSize"`
	Request      *FlowRequest      `json:"request"`
	Response     *FlowResponse     `json:"response"`
	IsPinned     bool              `json:"isPinned"`
	IsBlocked    bool              `json:"isBlocked"`
	ContentType  string            `json:"contentType"`
	Tags         []string          `json:"tags"`
}

// FlowRequest 表示HTTP请求
type FlowRequest struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    []byte            `json:"body"`
	Raw     string            `json:"raw"`
}

// FlowResponse 表示HTTP响应
type FlowResponse struct {
	StatusCode int               `json:"statusCode"`
	Status     string            `json:"status"`
	Headers    map[string]string `json:"headers"`
	Body       []byte            `json:"body"`
	Raw        string            `json:"raw"`
}

// NewFlow 创建新的Flow对象
func NewFlow(id string, req *http.Request) *Flow {
	scheme := "http"
	if req.TLS != nil {
		scheme = "https"
	}

	flow := &Flow{
		ID:        id,
		URL:       req.URL.String(),
		Method:    req.Method,
		Client:    req.RemoteAddr,
		Domain:    req.Host,
		Path:      req.URL.Path,
		Scheme:    scheme,
		StartTime: time.Now(),
		Tags:      make([]string, 0),
		Request: &FlowRequest{
			Method:  req.Method,
			URL:     req.URL.String(),
			Headers: make(map[string]string),
		},
	}

	// 复制请求头
	for name, values := range req.Header {
		if len(values) > 0 {
			flow.Request.Headers[name] = values[0]
		}
	}

	// 设置内容类型
	if contentType := req.Header.Get("Content-Type"); contentType != "" {
		flow.ContentType = contentType
	}

	return flow
}

// SetResponse 设置响应信息
func (f *Flow) SetResponse(resp *http.Response, body []byte) {
	f.EndTime = time.Now()
	f.Duration = f.EndTime.Sub(f.StartTime)
	f.StatusCode = resp.StatusCode
	f.ResponseSize = int64(len(body))

	f.Response = &FlowResponse{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Headers:    make(map[string]string),
		Body:       body,
	}

	// 复制响应头
	for name, values := range resp.Header {
		if len(values) > 0 {
			f.Response.Headers[name] = values[0]
		}
	}

	// 更新内容类型（如果响应中有）
	if contentType := resp.Header.Get("Content-Type"); contentType != "" && f.ContentType == "" {
		f.ContentType = contentType
	}
}

// AddTag 添加标签
func (f *Flow) AddTag(tag string) {
	for _, existingTag := range f.Tags {
		if existingTag == tag {
			return // 标签已存在
		}
	}
	f.Tags = append(f.Tags, tag)
}

// RemoveTag 移除标签
func (f *Flow) RemoveTag(tag string) {
	for i, existingTag := range f.Tags {
		if existingTag == tag {
			f.Tags = append(f.Tags[:i], f.Tags[i+1:]...)
			return
		}
	}
}

// SetRequestBody 设置请求体
func (f *Flow) SetRequestBody(body []byte) {
	f.Request.Body = body
	f.RequestSize = int64(len(body))
}
