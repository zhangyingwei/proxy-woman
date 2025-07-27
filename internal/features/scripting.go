package features

import (
	"fmt"
	"sync"
	"time"

	"ProxyWoman/internal/proxycore"

	"github.com/dop251/goja"
)

// Script 脚本结构
type Script struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Content     string    `json:"content"`
	Enabled     bool      `json:"enabled"`
	Type        string    `json:"type"` // "request", "response", "both"
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// ScriptContext 脚本执行上下文
type ScriptContext struct {
	Flow     *proxycore.Flow `json:"flow"`
	Request  *ScriptRequest  `json:"request"`
	Response *ScriptResponse `json:"response"`
	Console  *ScriptConsole  `json:"console"`
}

// ScriptRequest 脚本中的请求对象
type ScriptRequest struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

// ScriptResponse 脚本中的响应对象
type ScriptResponse struct {
	StatusCode int               `json:"statusCode"`
	Status     string            `json:"status"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
}

// ScriptConsole 脚本控制台
type ScriptConsole struct {
	logs []string
}

// Log 记录日志（大写L，用于Go代码调用）
func (sc *ScriptConsole) Log(args ...interface{}) {
	message := fmt.Sprint(args...)
	sc.logs = append(sc.logs, message)
}

// log 记录日志（小写l，用于JavaScript调用）
func (sc *ScriptConsole) LogJS(args ...interface{}) {
	message := fmt.Sprint(args...)
	sc.logs = append(sc.logs, message)
}

// GetLogs 获取日志
func (sc *ScriptConsole) GetLogs() []string {
	return sc.logs
}

// ScriptStorage 脚本存储接口
type ScriptStorage interface {
	SaveScript(script *Script) error
	GetScripts() ([]*Script, error)
	DeleteScript(id string) error
	UpdateScriptStatus(id string, enabled bool) error
}

// ScriptManager 脚本管理器
type ScriptManager struct {
	scripts      map[string]*Script
	scriptsMutex sync.RWMutex
	vm           *goja.Runtime
	storage      ScriptStorage
}

// NewScriptManager 创建脚本管理器
func NewScriptManager(storage ScriptStorage) *ScriptManager {
	manager := &ScriptManager{
		scripts: make(map[string]*Script),
		vm:      goja.New(),
		storage: storage,
	}

	// 从数据库加载脚本
	manager.loadScriptsFromStorage()

	return manager
}

// loadScriptsFromStorage 从存储加载脚本
func (sm *ScriptManager) loadScriptsFromStorage() {
	if sm.storage == nil {
		return
	}

	scripts, err := sm.storage.GetScripts()
	if err != nil {
		fmt.Printf("Failed to load scripts from storage: %v\n", err)
		return
	}

	sm.scriptsMutex.Lock()
	defer sm.scriptsMutex.Unlock()

	for _, script := range scripts {
		sm.scripts[script.ID] = script
	}
}

// AddScript 添加脚本
func (sm *ScriptManager) AddScript(script *Script) error {
	sm.scriptsMutex.Lock()
	defer sm.scriptsMutex.Unlock()

	script.CreatedAt = time.Now()
	script.UpdatedAt = time.Now()

	// 保存到数据库
	if sm.storage != nil {
		if err := sm.storage.SaveScript(script); err != nil {
			return fmt.Errorf("failed to save script: %v", err)
		}
	}

	sm.scripts[script.ID] = script
	return nil
}

// RemoveScript 移除脚本
func (sm *ScriptManager) RemoveScript(scriptID string) error {
	sm.scriptsMutex.Lock()
	defer sm.scriptsMutex.Unlock()

	// 从数据库删除
	if sm.storage != nil {
		if err := sm.storage.DeleteScript(scriptID); err != nil {
			return fmt.Errorf("failed to delete script: %v", err)
		}
	}

	delete(sm.scripts, scriptID)
	return nil
}

// UpdateScript 更新脚本
func (sm *ScriptManager) UpdateScript(script *Script) error {
	sm.scriptsMutex.Lock()
	defer sm.scriptsMutex.Unlock()

	if _, exists := sm.scripts[script.ID]; !exists {
		return fmt.Errorf("script not found: %s", script.ID)
	}

	script.UpdatedAt = time.Now()

	// 保存到数据库
	if sm.storage != nil {
		if err := sm.storage.SaveScript(script); err != nil {
			return fmt.Errorf("failed to update script: %v", err)
		}
	}

	sm.scripts[script.ID] = script
	return nil
}

// UpdateScriptStatus 更新脚本状态
func (sm *ScriptManager) UpdateScriptStatus(scriptID string, enabled bool) error {
	sm.scriptsMutex.Lock()
	defer sm.scriptsMutex.Unlock()

	script, exists := sm.scripts[scriptID]
	if !exists {
		return fmt.Errorf("script not found: %s", scriptID)
	}

	// 更新数据库
	if sm.storage != nil {
		if err := sm.storage.UpdateScriptStatus(scriptID, enabled); err != nil {
			return fmt.Errorf("failed to update script status: %v", err)
		}
	}

	script.Enabled = enabled
	script.UpdatedAt = time.Now()
	return nil
}

// GetScript 获取脚本
func (sm *ScriptManager) GetScript(scriptID string) (*Script, bool) {
	sm.scriptsMutex.RLock()
	defer sm.scriptsMutex.RUnlock()
	script, exists := sm.scripts[scriptID]
	return script, exists
}

// GetAllScripts 获取所有脚本
func (sm *ScriptManager) GetAllScripts() []*Script {
	sm.scriptsMutex.RLock()
	defer sm.scriptsMutex.RUnlock()

	scripts := make([]*Script, 0, len(sm.scripts))
	for _, script := range sm.scripts {
		scripts = append(scripts, script)
	}
	return scripts
}

// ExecuteRequestScripts 执行请求脚本
func (sm *ScriptManager) ExecuteRequestScripts(flow *proxycore.Flow) error {
	sm.scriptsMutex.RLock()
	defer sm.scriptsMutex.RUnlock()

	fmt.Printf("ExecuteRequestScripts: Found %d scripts\n", len(sm.scripts))
	executed := false
	for _, script := range sm.scripts {
		if !script.Enabled {
			fmt.Printf("Script '%s' is disabled, skipping\n", script.Name)
			continue
		}

		if script.Type != "request" && script.Type != "both" {
			fmt.Printf("Script '%s' type '%s' not applicable for request phase\n", script.Name, script.Type)
			continue
		}

		fmt.Printf("Executing request script: %s\n", script.Name)
		logs, err := sm.executeScript(script, flow, "request")

		// 记录脚本执行信息到Flow
		execution := proxycore.ScriptExecution{
			ScriptID:   script.ID,
			ScriptName: script.Name,
			Phase:      "request",
			Success:    err == nil,
			Logs:       logs,
			ExecutedAt: time.Now(),
		}
		if err != nil {
			execution.Error = err.Error()
			fmt.Printf("Script execution error (%s): %v\n", script.Name, err)
		} else {
			fmt.Printf("Script '%s' executed successfully\n", script.Name)
			executed = true
		}

		// 添加执行记录到Flow
		if flow.ScriptExecutions == nil {
			flow.ScriptExecutions = make([]proxycore.ScriptExecution, 0)
		}
		flow.ScriptExecutions = append(flow.ScriptExecutions, execution)
	}

	// 只有在实际执行了脚本时才添加标签
	if executed {
		if flow.Tags == nil {
			flow.Tags = make([]string, 0)
		}

		// 检查是否已经有脚本标签，避免重复添加
		hasScriptTag := false
		for _, tag := range flow.Tags {
			if tag == "script-processed" {
				hasScriptTag = true
				break
			}
		}

		if !hasScriptTag {
			flow.Tags = append(flow.Tags, "script-processed")
		}
	}

	return nil
}

// ExecuteResponseScripts 执行响应脚本
func (sm *ScriptManager) ExecuteResponseScripts(flow *proxycore.Flow) error {
	sm.scriptsMutex.RLock()
	defer sm.scriptsMutex.RUnlock()

	executed := false
	for _, script := range sm.scripts {
		if !script.Enabled {
			continue
		}

		if script.Type != "response" && script.Type != "both" {
			continue
		}

		logs, err := sm.executeScript(script, flow, "response")

		// 记录脚本执行信息到Flow
		execution := proxycore.ScriptExecution{
			ScriptID:   script.ID,
			ScriptName: script.Name,
			Phase:      "response",
			Success:    err == nil,
			Logs:       logs,
			ExecutedAt: time.Now(),
		}
		if err != nil {
			execution.Error = err.Error()
			fmt.Printf("Script execution error (%s): %v\n", script.Name, err)
		} else {
			executed = true
		}

		// 添加执行记录到Flow
		if flow.ScriptExecutions == nil {
			flow.ScriptExecutions = make([]proxycore.ScriptExecution, 0)
		}
		flow.ScriptExecutions = append(flow.ScriptExecutions, execution)
	}

	// 只有在实际执行了脚本时才添加标签
	if executed {
		if flow.Tags == nil {
			flow.Tags = make([]string, 0)
		}

		// 检查是否已经有脚本标签，避免重复添加
		hasScriptTag := false
		for _, tag := range flow.Tags {
			if tag == "script-processed" {
				hasScriptTag = true
				break
			}
		}

		if !hasScriptTag {
			flow.Tags = append(flow.Tags, "script-processed")
		}
	}

	return nil
}

// executeScript 执行单个脚本
func (sm *ScriptManager) executeScript(script *Script, flow *proxycore.Flow, phase string) ([]string, error) {
	// 创建新的VM实例以避免状态污染
	vm := goja.New()

	// 创建脚本上下文
	console := &ScriptConsole{logs: make([]string, 0)}
	context := &ScriptContext{
		Flow:    flow,
		Console: console,
	}

	// 设置请求对象
	if flow.Request != nil {
		context.Request = &ScriptRequest{
			Method:  flow.Request.Method,
			URL:     flow.Request.URL,
			Headers: flow.Request.Headers,
			Body:    string(flow.Request.Body),
		}
	}

	// 设置响应对象（如果存在）
	if flow.Response != nil {
		context.Response = &ScriptResponse{
			StatusCode: flow.Response.StatusCode,
			Status:     flow.Response.Status,
			Headers:    flow.Response.Headers,
			Body:       string(flow.Response.Body),
		}
	}

	// 将上下文对象注入到VM中
	vm.Set("flow", context.Flow)
	vm.Set("request", context.Request)
	vm.Set("response", context.Response)

	// 创建console对象，支持JavaScript的console.log调用
	consoleObj := vm.NewObject()
	consoleObj.Set("log", console.LogJS)
	vm.Set("console", consoleObj)

	// 创建JavaScript友好的context对象
	contextObj := vm.NewObject()
	contextObj.Set("flow", context.Flow)

	// 设置request对象
	if context.Request != nil {
		requestObj := vm.NewObject()
		requestObj.Set("method", context.Request.Method)
		requestObj.Set("url", context.Request.URL)
		requestObj.Set("headers", context.Request.Headers)
		requestObj.Set("body", context.Request.Body)
		contextObj.Set("request", requestObj)
		console.LogJS(fmt.Sprintf("Created request object: method=%s, url=%s", context.Request.Method, context.Request.URL))
	} else {
		console.LogJS("No request object available")
	}

	// 设置response对象
	if context.Response != nil {
		responseObj := vm.NewObject()
		responseObj.Set("statusCode", context.Response.StatusCode)
		responseObj.Set("status", context.Response.Status)
		responseObj.Set("headers", context.Response.Headers)
		responseObj.Set("body", context.Response.Body)
		contextObj.Set("response", responseObj)
		console.LogJS(fmt.Sprintf("Created response object: statusCode=%d, status=%s", context.Response.StatusCode, context.Response.Status))
	} else {
		console.LogJS("No response object available")
	}

	vm.Set("context", contextObj)

	// 添加一些实用函数
	vm.Set("setTimeout", func(callback func(), delay int) {
		go func() {
			time.Sleep(time.Duration(delay) * time.Millisecond)
			callback()
		}()
	})

	// 执行脚本
	_, err := vm.RunString(script.Content)
	if err != nil {
		return console.GetLogs(), fmt.Errorf("script execution failed: %v", err)
	}

	// 检查并调用特定的生命周期函数
	if phase == "request" {
		if onRequestFunc := vm.Get("onRequest"); onRequestFunc != nil {
			if callable, ok := goja.AssertFunction(onRequestFunc); ok {
				// 传递JavaScript友好的context对象
				_, err := callable(goja.Undefined(), vm.Get("context"))
				if err != nil {
					console.LogJS(fmt.Sprintf("onRequest function error: %v", err))
				}
			}
		}
	} else if phase == "response" {
		if onResponseFunc := vm.Get("onResponse"); onResponseFunc != nil {
			if callable, ok := goja.AssertFunction(onResponseFunc); ok {
				// 传递JavaScript友好的context对象
				_, err := callable(goja.Undefined(), vm.Get("context"))
				if err != nil {
					console.LogJS(fmt.Sprintf("onResponse function error: %v", err))
				}
			}
		}
	}

	// 获取修改后的值并应用到Flow
	// 直接从VM中获取修改后的context对象
	if contextVal := vm.Get("context"); contextVal != nil && !goja.IsUndefined(contextVal) {
		console.LogJS("Found context object in VM")

		// 将context对象转换为Go对象
		if contextObj := contextVal.ToObject(vm); contextObj != nil {
			// 应用请求修改
			if requestVal := contextObj.Get("request"); requestVal != nil && !goja.IsUndefined(requestVal) && flow.Request != nil {
				console.LogJS("Processing request modifications")
				if requestObj := requestVal.ToObject(vm); requestObj != nil {
					if method := requestObj.Get("method"); method != nil && !goja.IsUndefined(method) {
						if methodStr := method.String(); methodStr != "" {
							flow.Request.Method = methodStr
							console.LogJS(fmt.Sprintf("Updated request method to: %s", methodStr))
						}
					}
					if url := requestObj.Get("url"); url != nil && !goja.IsUndefined(url) {
						if urlStr := url.String(); urlStr != "" {
							flow.Request.URL = urlStr
							console.LogJS(fmt.Sprintf("Updated request URL to: %s", urlStr))
						}
					}
					if headers := requestObj.Get("headers"); headers != nil && !goja.IsUndefined(headers) {
						if headersObj := headers.Export(); headersObj != nil {
							if headerMap, ok := headersObj.(map[string]interface{}); ok {
								newHeaders := make(map[string]string)
								for k, v := range headerMap {
									if strVal, ok := v.(string); ok {
										newHeaders[k] = strVal
									}
								}
								flow.Request.Headers = newHeaders
								console.LogJS(fmt.Sprintf("Updated request headers: %d headers", len(newHeaders)))
							}
						}
					}
					if body := requestObj.Get("body"); body != nil && !goja.IsUndefined(body) {
						if bodyStr := body.String(); bodyStr != "" {
							flow.Request.Body = []byte(bodyStr)
							console.LogJS(fmt.Sprintf("Updated request body: %d bytes", len(bodyStr)))
						}
					}
				}
			}

			// 应用响应修改
			if responseVal := contextObj.Get("response"); responseVal != nil && !goja.IsUndefined(responseVal) && flow.Response != nil {
				console.LogJS("Processing response modifications")
				if responseObj := responseVal.ToObject(vm); responseObj != nil {
					if statusCode := responseObj.Get("statusCode"); statusCode != nil && !goja.IsUndefined(statusCode) {
						if statusCodeInt := statusCode.ToInteger(); statusCodeInt != 0 {
							flow.Response.StatusCode = int(statusCodeInt)
							console.LogJS(fmt.Sprintf("Updated response status code to: %d", statusCodeInt))
						}
					}
					if status := responseObj.Get("status"); status != nil && !goja.IsUndefined(status) {
						if statusStr := status.String(); statusStr != "" {
							flow.Response.Status = statusStr
							console.LogJS(fmt.Sprintf("Updated response status to: %s", statusStr))
						}
					}
					if headers := responseObj.Get("headers"); headers != nil && !goja.IsUndefined(headers) {
						if headersObj := headers.Export(); headersObj != nil {
							if headerMap, ok := headersObj.(map[string]interface{}); ok {
								newHeaders := make(map[string]string)
								for k, v := range headerMap {
									if strVal, ok := v.(string); ok {
										newHeaders[k] = strVal
									}
								}
								flow.Response.Headers = newHeaders
								console.LogJS(fmt.Sprintf("Updated response headers: %d headers", len(newHeaders)))
							}
						}
					}
					if body := responseObj.Get("body"); body != nil && !goja.IsUndefined(body) {
						if bodyStr := body.String(); bodyStr != "" {
							originalLen := len(flow.Response.Body)
							flow.Response.Body = []byte(bodyStr)
							console.LogJS(fmt.Sprintf("Updated response body: %d -> %d bytes", originalLen, len(bodyStr)))
						}
					}
				}
			}
		}
	} else {
		console.LogJS("No context object found in VM")
	}

	// 备用方案：直接从VM中获取修改后的request和response对象
	if context.Request != nil && flow.Request != nil {
		if val := vm.Get("request"); val != nil {
			if reqObj := val.Export(); reqObj != nil {
				if req, ok := reqObj.(*ScriptRequest); ok {
					flow.Request.Method = req.Method
					flow.Request.URL = req.URL
					flow.Request.Headers = req.Headers
					flow.Request.Body = []byte(req.Body)
				}
			}
		}
	}

	if context.Response != nil && flow.Response != nil {
		if val := vm.Get("response"); val != nil {
			if respObj := val.Export(); respObj != nil {
				if resp, ok := respObj.(*ScriptResponse); ok {
					flow.Response.StatusCode = resp.StatusCode
					flow.Response.Status = resp.Status
					flow.Response.Headers = resp.Headers
					flow.Response.Body = []byte(resp.Body)
				}
			}
		}
	}

	return console.GetLogs(), nil
}

// ValidateScript 验证脚本语法
func (sm *ScriptManager) ValidateScript(content string) error {
	vm := goja.New()
	_, err := vm.RunString(content)
	return err
}
