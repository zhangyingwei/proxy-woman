package features

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"ProxyWoman/internal/proxycore"
)

// HAR格式结构定义
// 参考: https://w3c.github.io/web-performance/specs/HAR/Overview.html

// HARFile HAR文件根结构
type HARFile struct {
	Log HARLog `json:"log"`
}

// HARLog HAR日志结构
type HARLog struct {
	Version string      `json:"version"`
	Creator HARCreator  `json:"creator"`
	Browser *HARBrowser `json:"browser,omitempty"`
	Pages   []HARPage   `json:"pages,omitempty"`
	Entries []HAREntry  `json:"entries"`
	Comment string      `json:"comment,omitempty"`
}

// HARCreator 创建者信息
type HARCreator struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Comment string `json:"comment,omitempty"`
}

// HARBrowser 浏览器信息
type HARBrowser struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Comment string `json:"comment,omitempty"`
}

// HARPage 页面信息
type HARPage struct {
	StartedDateTime string     `json:"startedDateTime"`
	ID              string     `json:"id"`
	Title           string     `json:"title"`
	PageTimings     HARTimings `json:"pageTimings"`
	Comment         string     `json:"comment,omitempty"`
}

// HAREntry HAR条目
type HAREntry struct {
	PageRef         string      `json:"pageref,omitempty"`
	StartedDateTime string      `json:"startedDateTime"`
	Time            float64     `json:"time"`
	Request         HARRequest  `json:"request"`
	Response        HARResponse `json:"response"`
	Cache           HARCache    `json:"cache"`
	Timings         HARTimings  `json:"timings"`
	ServerIPAddress string      `json:"serverIPAddress,omitempty"`
	Connection      string      `json:"connection,omitempty"`
	Comment         string      `json:"comment,omitempty"`
}

// HARRequest HAR请求
type HARRequest struct {
	Method      string            `json:"method"`
	URL         string            `json:"url"`
	HTTPVersion string            `json:"httpVersion"`
	Cookies     []HARCookie       `json:"cookies"`
	Headers     []HARNameValue    `json:"headers"`
	QueryString []HARNameValue    `json:"queryString"`
	PostData    *HARPostData      `json:"postData,omitempty"`
	HeadersSize int64             `json:"headersSize"`
	BodySize    int64             `json:"bodySize"`
	Comment     string            `json:"comment,omitempty"`
}

// HARResponse HAR响应
type HARResponse struct {
	Status      int            `json:"status"`
	StatusText  string         `json:"statusText"`
	HTTPVersion string         `json:"httpVersion"`
	Cookies     []HARCookie    `json:"cookies"`
	Headers     []HARNameValue `json:"headers"`
	Content     HARContent     `json:"content"`
	RedirectURL string         `json:"redirectURL"`
	HeadersSize int64          `json:"headersSize"`
	BodySize    int64          `json:"bodySize"`
	Comment     string         `json:"comment,omitempty"`
}

// HARNameValue 名称值对
type HARNameValue struct {
	Name    string `json:"name"`
	Value   string `json:"value"`
	Comment string `json:"comment,omitempty"`
}

// HARCookie Cookie信息
type HARCookie struct {
	Name     string    `json:"name"`
	Value    string    `json:"value"`
	Path     string    `json:"path,omitempty"`
	Domain   string    `json:"domain,omitempty"`
	Expires  time.Time `json:"expires,omitempty"`
	HTTPOnly bool      `json:"httpOnly,omitempty"`
	Secure   bool      `json:"secure,omitempty"`
	Comment  string    `json:"comment,omitempty"`
}

// HARPostData POST数据
type HARPostData struct {
	MimeType string           `json:"mimeType"`
	Params   []HARParam       `json:"params,omitempty"`
	Text     string           `json:"text,omitempty"`
	Comment  string           `json:"comment,omitempty"`
}

// HARParam 参数
type HARParam struct {
	Name        string `json:"name"`
	Value       string `json:"value,omitempty"`
	FileName    string `json:"fileName,omitempty"`
	ContentType string `json:"contentType,omitempty"`
	Comment     string `json:"comment,omitempty"`
}

// HARContent 内容
type HARContent struct {
	Size        int64  `json:"size"`
	Compression int64  `json:"compression,omitempty"`
	MimeType    string `json:"mimeType"`
	Text        string `json:"text,omitempty"`
	Encoding    string `json:"encoding,omitempty"`
	Comment     string `json:"comment,omitempty"`
}

// HARCache 缓存信息
type HARCache struct {
	BeforeRequest *HARCacheEntry `json:"beforeRequest,omitempty"`
	AfterRequest  *HARCacheEntry `json:"afterRequest,omitempty"`
	Comment       string         `json:"comment,omitempty"`
}

// HARCacheEntry 缓存条目
type HARCacheEntry struct {
	Expires    string `json:"expires,omitempty"`
	LastAccess string `json:"lastAccess"`
	ETag       string `json:"eTag"`
	HitCount   int    `json:"hitCount"`
	Comment    string `json:"comment,omitempty"`
}

// HARTimings 时间信息
type HARTimings struct {
	Blocked float64 `json:"blocked,omitempty"`
	DNS     float64 `json:"dns,omitempty"`
	Connect float64 `json:"connect,omitempty"`
	Send    float64 `json:"send"`
	Wait    float64 `json:"wait"`
	Receive float64 `json:"receive"`
	SSL     float64 `json:"ssl,omitempty"`
	Comment string  `json:"comment,omitempty"`
}

// HARManager HAR管理器
type HARManager struct{}

// NewHARManager 创建HAR管理器
func NewHARManager() *HARManager {
	return &HARManager{}
}

// ExportFlowsToHAR 导出Flows到HAR格式
func (hm *HARManager) ExportFlowsToHAR(flows []*proxycore.Flow, filePath string) error {
	harFile := &HARFile{
		Log: HARLog{
			Version: "1.2",
			Creator: HARCreator{
				Name:    "ProxyWoman",
				Version: "1.0.0",
				Comment: "Network debugging proxy",
			},
			Entries: make([]HAREntry, 0, len(flows)),
		},
	}

	// 转换每个Flow到HAR Entry
	for _, flow := range flows {
		entry := hm.flowToHAREntry(flow)
		harFile.Log.Entries = append(harFile.Log.Entries, entry)
	}

	// 序列化为JSON
	data, err := json.MarshalIndent(harFile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal HAR data: %v", err)
	}

	// 写入文件
	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write HAR file: %v", err)
	}

	return nil
}

// ImportHARToFlows 从HAR文件导入Flows
func (hm *HARManager) ImportHARToFlows(filePath string) ([]*proxycore.Flow, error) {
	// 读取文件
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read HAR file: %v", err)
	}

	// 解析JSON
	var harFile HARFile
	err = json.Unmarshal(data, &harFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HAR file: %v", err)
	}

	// 转换HAR Entries到Flows
	flows := make([]*proxycore.Flow, 0, len(harFile.Log.Entries))
	for i, entry := range harFile.Log.Entries {
		flow := hm.harEntryToFlow(entry, fmt.Sprintf("imported_%d", i))
		flows = append(flows, flow)
	}

	return flows, nil
}

// flowToHAREntry 将Flow转换为HAR Entry
func (hm *HARManager) flowToHAREntry(flow *proxycore.Flow) HAREntry {
	entry := HAREntry{
		StartedDateTime: flow.StartTime.Format(time.RFC3339),
		Time:            float64(flow.Duration.Nanoseconds()) / 1e6, // 转换为毫秒
		Request:         hm.flowRequestToHAR(flow),
		Response:        hm.flowResponseToHAR(flow),
		Cache:           HARCache{},
		Timings: HARTimings{
			Send:    0,
			Wait:    float64(flow.Duration.Nanoseconds()) / 1e6,
			Receive: 0,
		},
	}

	return entry
}

// flowRequestToHAR 将Flow请求转换为HAR请求
func (hm *HARManager) flowRequestToHAR(flow *proxycore.Flow) HARRequest {
	if flow.Request == nil {
		return HARRequest{}
	}

	headers := make([]HARNameValue, 0, len(flow.Request.Headers))
	for name, value := range flow.Request.Headers {
		headers = append(headers, HARNameValue{
			Name:  name,
			Value: value,
		})
	}

	req := HARRequest{
		Method:      flow.Request.Method,
		URL:         flow.Request.URL,
		HTTPVersion: "HTTP/1.1",
		Cookies:     []HARCookie{},
		Headers:     headers,
		QueryString: []HARNameValue{},
		HeadersSize: -1,
		BodySize:    int64(len(flow.Request.Body)),
	}

	// 添加POST数据
	if len(flow.Request.Body) > 0 {
		req.PostData = &HARPostData{
			MimeType: flow.ContentType,
			Text:     string(flow.Request.Body),
		}
	}

	return req
}

// flowResponseToHAR 将Flow响应转换为HAR响应
func (hm *HARManager) flowResponseToHAR(flow *proxycore.Flow) HARResponse {
	if flow.Response == nil {
		return HARResponse{
			Status:     0,
			StatusText: "",
			Headers:    []HARNameValue{},
			Content:    HARContent{},
		}
	}

	headers := make([]HARNameValue, 0, len(flow.Response.Headers))
	for name, value := range flow.Response.Headers {
		headers = append(headers, HARNameValue{
			Name:  name,
			Value: value,
		})
	}

	return HARResponse{
		Status:      flow.Response.StatusCode,
		StatusText:  flow.Response.Status,
		HTTPVersion: "HTTP/1.1",
		Cookies:     []HARCookie{},
		Headers:     headers,
		Content: HARContent{
			Size:     int64(len(flow.Response.Body)),
			MimeType: flow.ContentType,
			Text:     string(flow.Response.Body),
		},
		RedirectURL: "",
		HeadersSize: -1,
		BodySize:    int64(len(flow.Response.Body)),
	}
}

// harEntryToFlow 将HAR Entry转换为Flow
func (hm *HARManager) harEntryToFlow(entry HAREntry, id string) *proxycore.Flow {
	startTime, _ := time.Parse(time.RFC3339, entry.StartedDateTime)
	duration := time.Duration(entry.Time * 1e6) // 转换为纳秒

	flow := &proxycore.Flow{
		ID:        id,
		URL:       entry.Request.URL,
		Method:    entry.Request.Method,
		StartTime: startTime,
		EndTime:   startTime.Add(duration),
		Duration:  duration,
		Tags:      []string{"imported"},
	}

	// 转换请求
	if entry.Request.URL != "" {
		flow.Request = &proxycore.FlowRequest{
			Method:  entry.Request.Method,
			URL:     entry.Request.URL,
			Headers: make(map[string]string),
		}

		for _, header := range entry.Request.Headers {
			flow.Request.Headers[header.Name] = header.Value
		}

		if entry.Request.PostData != nil {
			flow.Request.Body = []byte(entry.Request.PostData.Text)
		}
	}

	// 转换响应
	if entry.Response.Status > 0 {
		flow.StatusCode = entry.Response.Status
		flow.Response = &proxycore.FlowResponse{
			StatusCode: entry.Response.Status,
			Status:     entry.Response.StatusText,
			Headers:    make(map[string]string),
			Body:       []byte(entry.Response.Content.Text),
		}

		for _, header := range entry.Response.Headers {
			flow.Response.Headers[header.Name] = header.Value
		}

		flow.ContentType = entry.Response.Content.MimeType
	}

	return flow
}
