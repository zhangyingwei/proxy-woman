package features

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

// MapLocalRule Map Local规则
type MapLocalRule struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	URLPattern  string `json:"urlPattern"`
	LocalPath   string `json:"localPath"`
	ContentType string `json:"contentType"`
	Enabled     bool   `json:"enabled"`
	IsRegex     bool   `json:"isRegex"`
}

// MapLocalManager Map Local管理器
type MapLocalManager struct {
	rules      map[string]*MapLocalRule
	rulesMutex sync.RWMutex
}

// NewMapLocalManager 创建新的Map Local管理器
func NewMapLocalManager() *MapLocalManager {
	return &MapLocalManager{
		rules: make(map[string]*MapLocalRule),
	}
}

// AddRule 添加规则
func (mlm *MapLocalManager) AddRule(rule *MapLocalRule) {
	mlm.rulesMutex.Lock()
	defer mlm.rulesMutex.Unlock()
	mlm.rules[rule.ID] = rule
}

// RemoveRule 移除规则
func (mlm *MapLocalManager) RemoveRule(ruleID string) {
	mlm.rulesMutex.Lock()
	defer mlm.rulesMutex.Unlock()
	delete(mlm.rules, ruleID)
}

// GetRule 获取规则
func (mlm *MapLocalManager) GetRule(ruleID string) (*MapLocalRule, bool) {
	mlm.rulesMutex.RLock()
	defer mlm.rulesMutex.RUnlock()
	rule, exists := mlm.rules[ruleID]
	return rule, exists
}

// GetAllRules 获取所有规则
func (mlm *MapLocalManager) GetAllRules() []*MapLocalRule {
	mlm.rulesMutex.RLock()
	defer mlm.rulesMutex.RUnlock()
	
	rules := make([]*MapLocalRule, 0, len(mlm.rules))
	for _, rule := range mlm.rules {
		rules = append(rules, rule)
	}
	return rules
}

// UpdateRule 更新规则
func (mlm *MapLocalManager) UpdateRule(rule *MapLocalRule) error {
	mlm.rulesMutex.Lock()
	defer mlm.rulesMutex.Unlock()
	
	if _, exists := mlm.rules[rule.ID]; !exists {
		return fmt.Errorf("rule not found: %s", rule.ID)
	}
	
	mlm.rules[rule.ID] = rule
	return nil
}

// MatchRule 匹配规则
func (mlm *MapLocalManager) MatchRule(url string) (*MapLocalRule, error) {
	mlm.rulesMutex.RLock()
	defer mlm.rulesMutex.RUnlock()
	
	for _, rule := range mlm.rules {
		if !rule.Enabled {
			continue
		}
		
		matched, err := mlm.matchURL(url, rule)
		if err != nil {
			continue // 忽略匹配错误，继续下一个规则
		}
		
		if matched {
			return rule, nil
		}
	}
	
	return nil, nil // 没有匹配的规则
}

// matchURL 匹配URL
func (mlm *MapLocalManager) matchURL(url string, rule *MapLocalRule) (bool, error) {
	if rule.IsRegex {
		regex, err := regexp.Compile(rule.URLPattern)
		if err != nil {
			return false, err
		}
		return regex.MatchString(url), nil
	} else {
		// 简单字符串匹配
		return strings.Contains(url, rule.URLPattern), nil
	}
}

// HandleMapLocal 处理Map Local请求
func (mlm *MapLocalManager) HandleMapLocal(w http.ResponseWriter, r *http.Request, rule *MapLocalRule) error {
	// 检查本地文件是否存在
	if _, err := os.Stat(rule.LocalPath); os.IsNotExist(err) {
		return fmt.Errorf("local file not found: %s", rule.LocalPath)
	}
	
	// 打开本地文件
	file, err := os.Open(rule.LocalPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %v", err)
	}
	defer file.Close()
	
	// 设置Content-Type
	if rule.ContentType != "" {
		w.Header().Set("Content-Type", rule.ContentType)
	} else {
		// 根据文件扩展名推断Content-Type
		ext := filepath.Ext(rule.LocalPath)
		contentType := getContentTypeByExt(ext)
		if contentType != "" {
			w.Header().Set("Content-Type", contentType)
		}
	}
	
	// 设置其他响应头
	w.Header().Set("X-ProxyWoman-MapLocal", "true")
	w.Header().Set("X-ProxyWoman-Rule-ID", rule.ID)
	
	// 复制文件内容到响应
	w.WriteHeader(http.StatusOK)
	_, err = io.Copy(w, file)
	return err
}

// getContentTypeByExt 根据文件扩展名获取Content-Type
func getContentTypeByExt(ext string) string {
	switch strings.ToLower(ext) {
	case ".html", ".htm":
		return "text/html; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".js":
		return "application/javascript; charset=utf-8"
	case ".json":
		return "application/json; charset=utf-8"
	case ".xml":
		return "application/xml; charset=utf-8"
	case ".txt":
		return "text/plain; charset=utf-8"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".ico":
		return "image/x-icon"
	case ".pdf":
		return "application/pdf"
	default:
		return ""
	}
}
