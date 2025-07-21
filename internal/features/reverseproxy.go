package features

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"ProxyWoman/internal/proxycore"
)

// ReverseProxyRule 反向代理规则
type ReverseProxyRule struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	ListenPath  string `json:"listenPath"`  // 监听路径模式
	TargetURL   string `json:"targetUrl"`   // 目标URL
	Enabled     bool   `json:"enabled"`
	IsRegex     bool   `json:"isRegex"`
	StripPath   bool   `json:"stripPath"`   // 是否去除路径前缀
	AddHeaders  map[string]string `json:"addHeaders"`  // 添加的请求头
	Description string `json:"description"`
}

// ReverseProxyManager 反向代理管理器
type ReverseProxyManager struct {
	rules      map[string]*ReverseProxyRule
	rulesMutex sync.RWMutex
	proxies    map[string]*httputil.ReverseProxy
}

// NewReverseProxyManager 创建反向代理管理器
func NewReverseProxyManager() *ReverseProxyManager {
	return &ReverseProxyManager{
		rules:   make(map[string]*ReverseProxyRule),
		proxies: make(map[string]*httputil.ReverseProxy),
	}
}

// AddRule 添加反向代理规则
func (rpm *ReverseProxyManager) AddRule(rule *ReverseProxyRule) error {
	rpm.rulesMutex.Lock()
	defer rpm.rulesMutex.Unlock()

	// 验证目标URL
	targetURL, err := url.Parse(rule.TargetURL)
	if err != nil {
		return fmt.Errorf("invalid target URL: %v", err)
	}

	// 创建反向代理
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	
	// 自定义Director函数
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		
		// 添加自定义头部
		for key, value := range rule.AddHeaders {
			req.Header.Set(key, value)
		}
		
		// 处理路径
		if rule.StripPath && rule.ListenPath != "/" {
			// 去除监听路径前缀
			if strings.HasPrefix(req.URL.Path, rule.ListenPath) {
				req.URL.Path = strings.TrimPrefix(req.URL.Path, rule.ListenPath)
				if !strings.HasPrefix(req.URL.Path, "/") {
					req.URL.Path = "/" + req.URL.Path
				}
			}
		}
	}

	rpm.rules[rule.ID] = rule
	rpm.proxies[rule.ID] = proxy

	return nil
}

// RemoveRule 移除反向代理规则
func (rpm *ReverseProxyManager) RemoveRule(ruleID string) {
	rpm.rulesMutex.Lock()
	defer rpm.rulesMutex.Unlock()
	
	delete(rpm.rules, ruleID)
	delete(rpm.proxies, ruleID)
}

// UpdateRule 更新反向代理规则
func (rpm *ReverseProxyManager) UpdateRule(rule *ReverseProxyRule) error {
	rpm.rulesMutex.Lock()
	defer rpm.rulesMutex.Unlock()
	
	if _, exists := rpm.rules[rule.ID]; !exists {
		return fmt.Errorf("rule not found: %s", rule.ID)
	}
	
	// 移除旧的代理
	delete(rpm.proxies, rule.ID)
	
	// 创建新的代理
	targetURL, err := url.Parse(rule.TargetURL)
	if err != nil {
		return fmt.Errorf("invalid target URL: %v", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		
		for key, value := range rule.AddHeaders {
			req.Header.Set(key, value)
		}
		
		if rule.StripPath && rule.ListenPath != "/" {
			if strings.HasPrefix(req.URL.Path, rule.ListenPath) {
				req.URL.Path = strings.TrimPrefix(req.URL.Path, rule.ListenPath)
				if !strings.HasPrefix(req.URL.Path, "/") {
					req.URL.Path = "/" + req.URL.Path
				}
			}
		}
	}

	rpm.rules[rule.ID] = rule
	rpm.proxies[rule.ID] = proxy

	return nil
}

// GetAllRules 获取所有反向代理规则
func (rpm *ReverseProxyManager) GetAllRules() []*ReverseProxyRule {
	rpm.rulesMutex.RLock()
	defer rpm.rulesMutex.RUnlock()
	
	rules := make([]*ReverseProxyRule, 0, len(rpm.rules))
	for _, rule := range rpm.rules {
		rules = append(rules, rule)
	}
	return rules
}

// MatchRule 匹配反向代理规则
func (rpm *ReverseProxyManager) MatchRule(path string) (*ReverseProxyRule, *httputil.ReverseProxy) {
	rpm.rulesMutex.RLock()
	defer rpm.rulesMutex.RUnlock()
	
	for _, rule := range rpm.rules {
		if !rule.Enabled {
			continue
		}
		
		matched, err := rpm.matchPath(path, rule)
		if err != nil {
			continue
		}
		
		if matched {
			proxy := rpm.proxies[rule.ID]
			return rule, proxy
		}
	}
	
	return nil, nil
}

// matchPath 匹配路径
func (rpm *ReverseProxyManager) matchPath(path string, rule *ReverseProxyRule) (bool, error) {
	if rule.IsRegex {
		regex, err := regexp.Compile(rule.ListenPath)
		if err != nil {
			return false, err
		}
		return regex.MatchString(path), nil
	} else {
		// 简单前缀匹配
		return strings.HasPrefix(path, rule.ListenPath), nil
	}
}

// ReverseProxyInterceptor 反向代理拦截器
type ReverseProxyInterceptor struct {
	manager *ReverseProxyManager
}

// NewReverseProxyInterceptor 创建反向代理拦截器
func NewReverseProxyInterceptor(manager *ReverseProxyManager) *ReverseProxyInterceptor {
	return &ReverseProxyInterceptor{
		manager: manager,
	}
}

// InterceptRequest 拦截请求
func (rpi *ReverseProxyInterceptor) InterceptRequest(flow *proxycore.Flow, w http.ResponseWriter, r *http.Request) (bool, error) {
	rule, proxy := rpi.manager.MatchRule(r.URL.Path)
	if rule == nil || proxy == nil {
		return false, nil // 不处理，继续正常代理
	}
	
	// 使用反向代理处理请求
	flow.AddTag("reverse-proxy")
	flow.AddTag(fmt.Sprintf("reverse-proxy-%s", rule.Name))
	
	// 设置响应头标识
	w.Header().Set("X-ProxyWoman-ReverseProxy", "true")
	w.Header().Set("X-ProxyWoman-Rule", rule.Name)
	
	// 执行反向代理
	proxy.ServeHTTP(w, r)
	
	return true, nil // 请求已处理
}

// ReverseProxyServer 反向代理服务器
type ReverseProxyServer struct {
	manager *ReverseProxyManager
	server  *http.Server
	port    int
	running bool
}

// NewReverseProxyServer 创建反向代理服务器
func NewReverseProxyServer(port int, manager *ReverseProxyManager) *ReverseProxyServer {
	return &ReverseProxyServer{
		manager: manager,
		port:    port,
		running: false,
	}
}

// Start 启动反向代理服务器
func (rps *ReverseProxyServer) Start() error {
	if rps.running {
		return fmt.Errorf("reverse proxy server is already running")
	}

	mux := http.NewServeMux()
	
	// 处理所有请求
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		rule, proxy := rps.manager.MatchRule(r.URL.Path)
		if rule != nil && proxy != nil {
			// 设置响应头
			w.Header().Set("X-ProxyWoman-ReverseProxy", "true")
			w.Header().Set("X-ProxyWoman-Rule", rule.Name)
			
			// 执行反向代理
			proxy.ServeHTTP(w, r)
		} else {
			// 没有匹配的规则，返回404
			http.NotFound(w, r)
		}
	})

	rps.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", rps.port),
		Handler: mux,
	}

	rps.running = true

	go func() {
		if err := rps.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Reverse proxy server error: %v\n", err)
		}
	}()

	return nil
}

// Stop 停止反向代理服务器
func (rps *ReverseProxyServer) Stop() error {
	if !rps.running {
		return nil
	}

	rps.running = false
	if rps.server != nil {
		return rps.server.Close()
	}
	return nil
}

// IsRunning 检查是否正在运行
func (rps *ReverseProxyServer) IsRunning() bool {
	return rps.running
}
