package features

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"ProxyWoman/internal/proxycore"
)

// MapLocalInterceptor Map Local拦截器
type MapLocalInterceptor struct {
	manager *MapLocalManager
}

// NewMapLocalInterceptor 创建Map Local拦截器
func NewMapLocalInterceptor(manager *MapLocalManager) *MapLocalInterceptor {
	return &MapLocalInterceptor{
		manager: manager,
	}
}

// InterceptRequest 拦截请求
func (mli *MapLocalInterceptor) InterceptRequest(flow *proxycore.Flow, w http.ResponseWriter, r *http.Request) (bool, error) {
	rule, err := mli.manager.MatchRule(flow.URL)
	if err != nil {
		return false, err
	}

	if rule != nil {
		// 处理Map Local
		err = mli.manager.HandleMapLocal(w, r, rule)
		if err != nil {
			return false, err
		}

		// 标记Flow为Map Local处理
		flow.AddTag("map-local")
		return true, nil
	}

	return false, nil
}

// BreakpointInterceptor 断点拦截器
type BreakpointInterceptor struct {
	manager      *BreakpointManager
	eventHandler func(session *BreakpointSession)
}

// NewBreakpointInterceptor 创建断点拦截器
func NewBreakpointInterceptor(manager *BreakpointManager) *BreakpointInterceptor {
	return &BreakpointInterceptor{
		manager: manager,
	}
}

// SetEventHandler 设置事件处理器
func (bi *BreakpointInterceptor) SetEventHandler(handler func(session *BreakpointSession)) {
	bi.eventHandler = handler
	bi.manager.SetEventHandler(handler)
}

// InterceptRequest 拦截请求
func (bi *BreakpointInterceptor) InterceptRequest(flow *proxycore.Flow, w http.ResponseWriter, r *http.Request) (bool, error) {
	session, hasBreakpoint := bi.manager.CheckBreakpoint(flow, "request")
	if !hasBreakpoint {
		return false, nil
	}

	// 等待断点恢复
	_, err := bi.manager.WaitForBreakpoint(session, 5*time.Minute)
	if err != nil {
		return false, err
	}

	// 如果有修改的请求，使用修改后的请求
	if session.ModifiedRequest != nil {
		*r = *session.ModifiedRequest
	}

	flow.AddTag("breakpoint-request")
	return false, nil // 继续处理请求
}

// InterceptResponse 拦截响应
func (bi *BreakpointInterceptor) InterceptResponse(flow *proxycore.Flow, resp *http.Response) (*http.Response, error) {
	session, hasBreakpoint := bi.manager.CheckBreakpoint(flow, "response")
	if !hasBreakpoint {
		return resp, nil
	}

	// 等待断点恢复
	modifiedResp, err := bi.manager.WaitForBreakpoint(session, 5*time.Minute)
	if err != nil {
		return resp, err
	}

	flow.AddTag("breakpoint-response")

	if modifiedResp != nil {
		return modifiedResp, nil
	}

	return resp, nil
}

// ScriptInterceptor 脚本拦截器
type ScriptInterceptor struct {
	manager *ScriptManager
}

// NewScriptInterceptor 创建脚本拦截器
func NewScriptInterceptor(manager *ScriptManager) *ScriptInterceptor {
	return &ScriptInterceptor{
		manager: manager,
	}
}

// InterceptRequest 拦截请求
func (si *ScriptInterceptor) InterceptRequest(flow *proxycore.Flow, w http.ResponseWriter, r *http.Request) (bool, error) {
	// 保存原始请求信息
	originalMethod := r.Method
	originalURL := r.URL.String()
	originalHeaders := make(map[string]string)
	for k, v := range r.Header {
		if len(v) > 0 {
			originalHeaders[k] = v[0]
		}
	}

	err := si.manager.ExecuteRequestScripts(flow)
	if err != nil {
		return false, err
	}

	// 检查脚本是否修改了请求，如果有修改则应用到实际请求中
	if flow.Request != nil {
		modified := false

		// 检查方法是否被修改
		if flow.Request.Method != originalMethod {
			r.Method = flow.Request.Method
			modified = true
		}

		// 检查URL是否被修改
		if flow.Request.URL != originalURL {
			newURL, err := url.Parse(flow.Request.URL)
			if err == nil {
				r.URL = newURL
				modified = true
			}
		}

		// 检查请求头是否被修改
		for k, v := range flow.Request.Headers {
			if originalHeaders[k] != v {
				r.Header.Set(k, v)
				modified = true
			}
		}

		// 检查是否有新增的请求头
		for k, v := range flow.Request.Headers {
			if _, exists := originalHeaders[k]; !exists {
				r.Header.Set(k, v)
				modified = true
			}
		}

		// 检查请求体是否被修改
		if len(flow.Request.Body) > 0 {
			r.Body = io.NopCloser(strings.NewReader(string(flow.Request.Body)))
			r.ContentLength = int64(len(flow.Request.Body))
			modified = true
		}

		if modified {
			// 添加脚本修改标签
			flow.AddTag("script-modified-request")
		}
	}

	return false, nil // 继续处理请求
}

// InterceptResponse 拦截响应
func (si *ScriptInterceptor) InterceptResponse(flow *proxycore.Flow, resp *http.Response) (*http.Response, error) {
	// 保存原始响应信息
	originalStatusCode := resp.StatusCode
	originalStatus := resp.Status
	originalHeaders := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			originalHeaders[k] = v[0]
		}
	}

	// 读取原始响应体
	var originalBody []byte
	if resp.Body != nil {
		var err error
		originalBody, err = io.ReadAll(resp.Body)
		if err != nil {
			return resp, err
		}
		resp.Body.Close()
		// 重新设置响应体供脚本使用
		resp.Body = io.NopCloser(bytes.NewReader(originalBody))
	}

	err := si.manager.ExecuteResponseScripts(flow)
	if err != nil {
		// 如果脚本执行失败，恢复原始响应体
		if originalBody != nil {
			resp.Body = io.NopCloser(bytes.NewReader(originalBody))
		}
		return resp, err
	}

	// 检查脚本是否修改了响应，如果有修改则创建新的响应
	if flow.Response != nil {
		modified := false
		newResp := &http.Response{
			Status:           resp.Status,
			StatusCode:       resp.StatusCode,
			Proto:            resp.Proto,
			ProtoMajor:       resp.ProtoMajor,
			ProtoMinor:       resp.ProtoMinor,
			Header:           make(http.Header),
			ContentLength:    resp.ContentLength,
			TransferEncoding: resp.TransferEncoding,
			Close:            resp.Close,
			Uncompressed:     resp.Uncompressed,
			Trailer:          resp.Trailer,
			Request:          resp.Request,
			TLS:              resp.TLS,
		}

		// 复制原始响应头
		for k, v := range resp.Header {
			newResp.Header[k] = v
		}

		// 检查状态码是否被修改
		if flow.Response.StatusCode != originalStatusCode {
			newResp.StatusCode = flow.Response.StatusCode
			modified = true
		}

		// 检查状态文本是否被修改
		if flow.Response.Status != originalStatus {
			newResp.Status = flow.Response.Status
			modified = true
		}

		// 检查响应头是否被修改
		for k, v := range flow.Response.Headers {
			if originalHeaders[k] != v {
				newResp.Header.Set(k, v)
				modified = true
			}
		}

		// 检查是否有新增的响应头
		for k, v := range flow.Response.Headers {
			if _, exists := originalHeaders[k]; !exists {
				newResp.Header.Set(k, v)
				modified = true
			}
		}

		// 检查响应体是否被修改
		if len(flow.Response.Body) > 0 && string(flow.Response.Body) != string(originalBody) {
			newResp.Body = io.NopCloser(bytes.NewReader(flow.Response.Body))
			newResp.ContentLength = int64(len(flow.Response.Body))
			modified = true
		} else {
			// 使用原始响应体
			newResp.Body = io.NopCloser(bytes.NewReader(originalBody))
		}

		if modified {
			// 添加脚本修改标签
			flow.AddTag("script-modified-response")
			return newResp, nil
		}
	}

	// 如果没有修改，恢复原始响应体
	if originalBody != nil {
		resp.Body = io.NopCloser(bytes.NewReader(originalBody))
	}

	return resp, nil
}
