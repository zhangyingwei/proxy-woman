package features

import (
	"fmt"
	"sync"
	"time"

	"github.com/dop251/goja"
	"ProxyWoman/internal/proxycore"
)

// Script 脚本结构
type Script struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Content     string `json:"content"`
	Enabled     bool   `json:"enabled"`
	Type        string `json:"type"` // "request", "response", "both"
	Description string `json:"description"`
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

// Log 记录日志
func (sc *ScriptConsole) Log(args ...interface{}) {
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
	scripts     map[string]*Script
	scriptsMutex sync.RWMutex
	vm          *goja.Runtime
	storage     ScriptStorage
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
		err := sm.executeScript(script, flow, "request")
		if err != nil {
			// 记录错误但继续执行其他脚本
			fmt.Printf("Script execution error (%s): %v\n", script.Name, err)
		} else {
			fmt.Printf("Script '%s' executed successfully\n", script.Name)
			executed = true
		}
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

		err := sm.executeScript(script, flow, "response")
		if err != nil {
			// 记录错误但继续执行其他脚本
			fmt.Printf("Script execution error (%s): %v\n", script.Name, err)
		} else {
			executed = true
		}
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
func (sm *ScriptManager) executeScript(script *Script, flow *proxycore.Flow, phase string) error {
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
	vm.Set("console", console)
	
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
		return fmt.Errorf("script execution failed: %v", err)
	}
	
	// 获取修改后的值并应用到Flow
	if context.Request != nil && flow.Request != nil {
		// 应用请求修改
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
		// 应用响应修改
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

	return nil
}

// ValidateScript 验证脚本语法
func (sm *ScriptManager) ValidateScript(content string) error {
	vm := goja.New()
	_, err := vm.RunString(content)
	return err
}
