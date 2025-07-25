package features

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"ProxyWoman/internal/proxycore"
)

// BreakpointRule 断点规则
type BreakpointRule struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	URLPattern  string `json:"urlPattern"`
	Method      string `json:"method"`
	Enabled     bool   `json:"enabled"`
	IsRegex     bool   `json:"isRegex"`
	BreakOnRequest  bool `json:"breakOnRequest"`
	BreakOnResponse bool `json:"breakOnResponse"`
}

// BreakpointSession 断点会话
type BreakpointSession struct {
	ID          string                 `json:"id"`
	Flow        *proxycore.Flow        `json:"flow"`
	Rule        *BreakpointRule        `json:"rule"`
	Type        string                 `json:"type"` // "request" or "response"
	StartTime   time.Time              `json:"startTime"`
	ResponseChan chan *http.Response   `json:"-"`
	ErrorChan   chan error             `json:"-"`
	ModifiedRequest *http.Request      `json:"-"`
	ModifiedResponse *http.Response    `json:"-"`
}

// BreakpointStorage 断点存储接口
type BreakpointStorage interface {
	SaveBreakpointRule(rule *BreakpointRule) error
	GetBreakpointRules() ([]*BreakpointRule, error)
	DeleteBreakpointRule(id string) error
	UpdateBreakpointRuleStatus(id string, enabled bool) error
}

// BreakpointManager 断点管理器
type BreakpointManager struct {
	rules        map[string]*BreakpointRule
	sessions     map[string]*BreakpointSession
	rulesMutex   sync.RWMutex
	sessionsMutex sync.RWMutex
	eventHandler func(session *BreakpointSession)
	storage      BreakpointStorage
}

// NewBreakpointManager 创建新的断点管理器
func NewBreakpointManager(storage BreakpointStorage) *BreakpointManager {
	manager := &BreakpointManager{
		rules:    make(map[string]*BreakpointRule),
		sessions: make(map[string]*BreakpointSession),
		storage:  storage,
	}

	// 从数据库加载规则
	manager.loadRulesFromStorage()

	return manager
}

// loadRulesFromStorage 从存储加载规则
func (bm *BreakpointManager) loadRulesFromStorage() {
	if bm.storage == nil {
		return
	}

	rules, err := bm.storage.GetBreakpointRules()
	if err != nil {
		fmt.Printf("Failed to load breakpoint rules from storage: %v\n", err)
		return
	}

	bm.rulesMutex.Lock()
	defer bm.rulesMutex.Unlock()

	for _, rule := range rules {
		bm.rules[rule.ID] = rule
	}
}

// SetEventHandler 设置事件处理器
func (bm *BreakpointManager) SetEventHandler(handler func(session *BreakpointSession)) {
	bm.eventHandler = handler
}

// AddRule 添加断点规则
func (bm *BreakpointManager) AddRule(rule *BreakpointRule) error {
	bm.rulesMutex.Lock()
	defer bm.rulesMutex.Unlock()

	// 保存到数据库
	if bm.storage != nil {
		if err := bm.storage.SaveBreakpointRule(rule); err != nil {
			return fmt.Errorf("failed to save breakpoint rule: %v", err)
		}
	}

	bm.rules[rule.ID] = rule
	return nil
}

// RemoveRule 移除断点规则
func (bm *BreakpointManager) RemoveRule(ruleID string) error {
	bm.rulesMutex.Lock()
	defer bm.rulesMutex.Unlock()

	// 从数据库删除
	if bm.storage != nil {
		if err := bm.storage.DeleteBreakpointRule(ruleID); err != nil {
			return fmt.Errorf("failed to delete breakpoint rule: %v", err)
		}
	}

	delete(bm.rules, ruleID)
	return nil
}

// UpdateRuleStatus 更新断点规则状态
func (bm *BreakpointManager) UpdateRuleStatus(ruleID string, enabled bool) error {
	bm.rulesMutex.Lock()
	defer bm.rulesMutex.Unlock()

	rule, exists := bm.rules[ruleID]
	if !exists {
		return fmt.Errorf("breakpoint rule not found: %s", ruleID)
	}

	// 更新数据库
	if bm.storage != nil {
		if err := bm.storage.UpdateBreakpointRuleStatus(ruleID, enabled); err != nil {
			return fmt.Errorf("failed to update breakpoint rule status: %v", err)
		}
	}

	rule.Enabled = enabled
	return nil
}

// GetAllRules 获取所有断点规则
func (bm *BreakpointManager) GetAllRules() []*BreakpointRule {
	bm.rulesMutex.RLock()
	defer bm.rulesMutex.RUnlock()
	
	rules := make([]*BreakpointRule, 0, len(bm.rules))
	for _, rule := range bm.rules {
		rules = append(rules, rule)
	}
	return rules
}

// CheckBreakpoint 检查是否需要断点
func (bm *BreakpointManager) CheckBreakpoint(flow *proxycore.Flow, breakType string) (*BreakpointSession, bool) {
	bm.rulesMutex.RLock()
	defer bm.rulesMutex.RUnlock()
	
	for _, rule := range bm.rules {
		if !rule.Enabled {
			continue
		}
		
		// 检查断点类型
		if breakType == "request" && !rule.BreakOnRequest {
			continue
		}
		if breakType == "response" && !rule.BreakOnResponse {
			continue
		}
		
		// 检查方法匹配
		if rule.Method != "" && rule.Method != "*" && rule.Method != flow.Method {
			continue
		}
		
		// 检查URL匹配
		matched, err := bm.matchURL(flow.URL, rule)
		if err != nil || !matched {
			continue
		}
		
		// 创建断点会话
		session := &BreakpointSession{
			ID:           fmt.Sprintf("bp_%d", time.Now().UnixNano()),
			Flow:         flow,
			Rule:         rule,
			Type:         breakType,
			StartTime:    time.Now(),
			ResponseChan: make(chan *http.Response, 1),
			ErrorChan:    make(chan error, 1),
		}
		
		bm.sessionsMutex.Lock()
		bm.sessions[session.ID] = session
		bm.sessionsMutex.Unlock()
		
		// 通知前端
		if bm.eventHandler != nil {
			go bm.eventHandler(session)
		}
		
		return session, true
	}
	
	return nil, false
}

// matchURL 匹配URL
func (bm *BreakpointManager) matchURL(url string, rule *BreakpointRule) (bool, error) {
	if rule.IsRegex {
		regex, err := regexp.Compile(rule.URLPattern)
		if err != nil {
			return false, err
		}
		return regex.MatchString(url), nil
	} else {
		return strings.Contains(url, rule.URLPattern), nil
	}
}

// ResumeBreakpoint 恢复断点
func (bm *BreakpointManager) ResumeBreakpoint(sessionID string, modifiedRequest *http.Request, modifiedResponse *http.Response) error {
	bm.sessionsMutex.Lock()
	defer bm.sessionsMutex.Unlock()
	
	session, exists := bm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("breakpoint session not found: %s", sessionID)
	}
	
	session.ModifiedRequest = modifiedRequest
	session.ModifiedResponse = modifiedResponse
	
	// 发送恢复信号
	if modifiedResponse != nil {
		session.ResponseChan <- modifiedResponse
	} else {
		session.ResponseChan <- nil
	}
	
	// 清理会话
	delete(bm.sessions, sessionID)
	
	return nil
}

// CancelBreakpoint 取消断点
func (bm *BreakpointManager) CancelBreakpoint(sessionID string) error {
	bm.sessionsMutex.Lock()
	defer bm.sessionsMutex.Unlock()
	
	session, exists := bm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("breakpoint session not found: %s", sessionID)
	}
	
	// 发送错误信号
	session.ErrorChan <- fmt.Errorf("breakpoint cancelled")
	
	// 清理会话
	delete(bm.sessions, sessionID)
	
	return nil
}

// GetActiveBreakpoints 获取活跃的断点会话
func (bm *BreakpointManager) GetActiveBreakpoints() []*BreakpointSession {
	bm.sessionsMutex.RLock()
	defer bm.sessionsMutex.RUnlock()
	
	sessions := make([]*BreakpointSession, 0, len(bm.sessions))
	for _, session := range bm.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

// WaitForBreakpoint 等待断点恢复
func (bm *BreakpointManager) WaitForBreakpoint(session *BreakpointSession, timeout time.Duration) (*http.Response, error) {
	select {
	case response := <-session.ResponseChan:
		return response, nil
	case err := <-session.ErrorChan:
		return nil, err
	case <-time.After(timeout):
		// 超时，自动恢复
		bm.sessionsMutex.Lock()
		delete(bm.sessions, session.ID)
		bm.sessionsMutex.Unlock()
		return nil, fmt.Errorf("breakpoint timeout")
	}
}
