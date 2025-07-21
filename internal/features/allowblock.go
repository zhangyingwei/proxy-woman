package features

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"ProxyWoman/internal/proxycore"
)

// AllowBlockRule 允许/阻止规则
type AllowBlockRule struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	URLPattern  string `json:"urlPattern"`
	Method      string `json:"method"`
	Type        string `json:"type"` // "allow" or "block"
	Enabled     bool   `json:"enabled"`
	IsRegex     bool   `json:"isRegex"`
	Description string `json:"description"`
}

// AllowBlockManager 允许/阻止管理器
type AllowBlockManager struct {
	rules      map[string]*AllowBlockRule
	rulesMutex sync.RWMutex
	mode       string // "whitelist" (只允许匹配的), "blacklist" (阻止匹配的), "mixed" (混合模式)
}

// NewAllowBlockManager 创建允许/阻止管理器
func NewAllowBlockManager() *AllowBlockManager {
	return &AllowBlockManager{
		rules: make(map[string]*AllowBlockRule),
		mode:  "mixed", // 默认混合模式
	}
}

// SetMode 设置模式
func (abm *AllowBlockManager) SetMode(mode string) error {
	if mode != "whitelist" && mode != "blacklist" && mode != "mixed" {
		return fmt.Errorf("invalid mode: %s", mode)
	}
	abm.mode = mode
	return nil
}

// GetMode 获取当前模式
func (abm *AllowBlockManager) GetMode() string {
	return abm.mode
}

// AddRule 添加规则
func (abm *AllowBlockManager) AddRule(rule *AllowBlockRule) {
	abm.rulesMutex.Lock()
	defer abm.rulesMutex.Unlock()
	abm.rules[rule.ID] = rule
}

// RemoveRule 移除规则
func (abm *AllowBlockManager) RemoveRule(ruleID string) {
	abm.rulesMutex.Lock()
	defer abm.rulesMutex.Unlock()
	delete(abm.rules, ruleID)
}

// UpdateRule 更新规则
func (abm *AllowBlockManager) UpdateRule(rule *AllowBlockRule) error {
	abm.rulesMutex.Lock()
	defer abm.rulesMutex.Unlock()
	
	if _, exists := abm.rules[rule.ID]; !exists {
		return fmt.Errorf("rule not found: %s", rule.ID)
	}
	
	abm.rules[rule.ID] = rule
	return nil
}

// GetRule 获取规则
func (abm *AllowBlockManager) GetRule(ruleID string) (*AllowBlockRule, bool) {
	abm.rulesMutex.RLock()
	defer abm.rulesMutex.RUnlock()
	rule, exists := abm.rules[ruleID]
	return rule, exists
}

// GetAllRules 获取所有规则
func (abm *AllowBlockManager) GetAllRules() []*AllowBlockRule {
	abm.rulesMutex.RLock()
	defer abm.rulesMutex.RUnlock()
	
	rules := make([]*AllowBlockRule, 0, len(abm.rules))
	for _, rule := range abm.rules {
		rules = append(rules, rule)
	}
	return rules
}

// CheckRequest 检查请求是否应该被允许
func (abm *AllowBlockManager) CheckRequest(flow *proxycore.Flow) (allowed bool, rule *AllowBlockRule) {
	abm.rulesMutex.RLock()
	defer abm.rulesMutex.RUnlock()
	
	var matchedAllowRule *AllowBlockRule
	var matchedBlockRule *AllowBlockRule
	
	// 检查所有规则
	for _, rule := range abm.rules {
		if !rule.Enabled {
			continue
		}
		
		// 检查方法匹配
		if rule.Method != "" && rule.Method != "*" && rule.Method != flow.Method {
			continue
		}
		
		// 检查URL匹配
		matched, err := abm.matchURL(flow.URL, rule)
		if err != nil || !matched {
			continue
		}
		
		// 记录匹配的规则
		if rule.Type == "allow" {
			matchedAllowRule = rule
		} else if rule.Type == "block" {
			matchedBlockRule = rule
		}
	}
	
	// 根据模式和匹配的规则决定是否允许
	switch abm.mode {
	case "whitelist":
		// 白名单模式：只有匹配允许规则的请求才被允许
		if matchedAllowRule != nil {
			return true, matchedAllowRule
		}
		return false, nil
		
	case "blacklist":
		// 黑名单模式：匹配阻止规则的请求被阻止，其他都允许
		if matchedBlockRule != nil {
			return false, matchedBlockRule
		}
		return true, nil
		
	case "mixed":
		// 混合模式：阻止规则优先，然后是允许规则，最后默认允许
		if matchedBlockRule != nil {
			return false, matchedBlockRule
		}
		if matchedAllowRule != nil {
			return true, matchedAllowRule
		}
		return true, nil // 默认允许
		
	default:
		return true, nil
	}
}

// matchURL 匹配URL
func (abm *AllowBlockManager) matchURL(url string, rule *AllowBlockRule) (bool, error) {
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

// AllowBlockInterceptor 允许/阻止拦截器
type AllowBlockInterceptor struct {
	manager *AllowBlockManager
}

// NewAllowBlockInterceptor 创建允许/阻止拦截器
func NewAllowBlockInterceptor(manager *AllowBlockManager) *AllowBlockInterceptor {
	return &AllowBlockInterceptor{
		manager: manager,
	}
}

// InterceptRequest 拦截请求
func (abi *AllowBlockInterceptor) InterceptRequest(flow *proxycore.Flow, w http.ResponseWriter, r *http.Request) (bool, error) {
	allowed, rule := abi.manager.CheckRequest(flow)
	
	if !allowed {
		// 请求被阻止
		flow.IsBlocked = true
		if rule != nil {
			flow.AddTag(fmt.Sprintf("blocked-by-%s", rule.Name))
		} else {
			flow.AddTag("blocked")
		}
		
		// 返回403 Forbidden
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("X-ProxyWoman-Blocked", "true")
		if rule != nil {
			w.Header().Set("X-ProxyWoman-Rule", rule.Name)
		}
		
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
    <title>Request Blocked - ProxyWoman</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background: #f5f5f5; }
        .container { background: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .error { color: #d32f2f; }
        .url { background: #f0f0f0; padding: 10px; border-radius: 4px; word-break: break-all; }
    </style>
</head>
<body>
    <div class="container">
        <h1 class="error">🚫 Request Blocked</h1>
        <p>This request has been blocked by ProxyWoman.</p>
        <div class="url">` + flow.URL + `</div>
        <p><small>ProxyWoman - Network Debugging Proxy</small></p>
    </div>
</body>
</html>`))
		
		return true, nil // 请求已处理
	}
	
	// 请求被允许，添加标签
	if rule != nil {
		flow.AddTag(fmt.Sprintf("allowed-by-%s", rule.Name))
	}
	
	return false, nil // 继续处理请求
}

// GetBlockedRequestsCount 获取被阻止的请求数量
func (abm *AllowBlockManager) GetBlockedRequestsCount() int {
	// 这个方法需要与代理服务器集成来统计
	// 暂时返回0
	return 0
}

// GetAllowedRequestsCount 获取被允许的请求数量
func (abm *AllowBlockManager) GetAllowedRequestsCount() int {
	// 这个方法需要与代理服务器集成来统计
	// 暂时返回0
	return 0
}
