package features

import (
	"net/http"
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
	err := si.manager.ExecuteRequestScripts(flow)
	if err != nil {
		return false, err
	}

	return false, nil // 继续处理请求
}

// InterceptResponse 拦截响应
func (si *ScriptInterceptor) InterceptResponse(flow *proxycore.Flow, resp *http.Response) (*http.Response, error) {
	err := si.manager.ExecuteResponseScripts(flow)
	if err != nil {
		return resp, err
	}

	return resp, nil
}
