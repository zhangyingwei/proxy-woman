package features

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"ProxyWoman/internal/proxycore"
)

// UpstreamProxy 上游代理配置
type UpstreamProxy struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	ProxyURL    string `json:"proxyUrl"`    // 上游代理URL (http://proxy:port 或 socks5://proxy:port)
	URLPattern  string `json:"urlPattern"` // 匹配的URL模式
	Enabled     bool   `json:"enabled"`
	IsRegex     bool   `json:"isRegex"`
	Username    string `json:"username,omitempty"`
	Password    string `json:"password,omitempty"`
	Description string `json:"description"`
}

// UpstreamManager 上游代理管理器
type UpstreamManager struct {
	proxies     map[string]*UpstreamProxy
	proxiesMutex sync.RWMutex
	clients     map[string]*http.Client
}

// NewUpstreamManager 创建上游代理管理器
func NewUpstreamManager() *UpstreamManager {
	return &UpstreamManager{
		proxies: make(map[string]*UpstreamProxy),
		clients: make(map[string]*http.Client),
	}
}

// AddProxy 添加上游代理
func (um *UpstreamManager) AddProxy(proxy *UpstreamProxy) error {
	um.proxiesMutex.Lock()
	defer um.proxiesMutex.Unlock()

	// 验证代理URL
	proxyURL, err := url.Parse(proxy.ProxyURL)
	if err != nil {
		return fmt.Errorf("invalid proxy URL: %v", err)
	}

	// 创建HTTP客户端
	client, err := um.createHTTPClient(proxy, proxyURL)
	if err != nil {
		return fmt.Errorf("failed to create HTTP client: %v", err)
	}

	um.proxies[proxy.ID] = proxy
	um.clients[proxy.ID] = client

	return nil
}

// RemoveProxy 移除上游代理
func (um *UpstreamManager) RemoveProxy(proxyID string) {
	um.proxiesMutex.Lock()
	defer um.proxiesMutex.Unlock()
	
	delete(um.proxies, proxyID)
	delete(um.clients, proxyID)
}

// UpdateProxy 更新上游代理
func (um *UpstreamManager) UpdateProxy(proxy *UpstreamProxy) error {
	um.proxiesMutex.Lock()
	defer um.proxiesMutex.Unlock()
	
	if _, exists := um.proxies[proxy.ID]; !exists {
		return fmt.Errorf("proxy not found: %s", proxy.ID)
	}

	// 验证代理URL
	proxyURL, err := url.Parse(proxy.ProxyURL)
	if err != nil {
		return fmt.Errorf("invalid proxy URL: %v", err)
	}

	// 创建新的HTTP客户端
	client, err := um.createHTTPClient(proxy, proxyURL)
	if err != nil {
		return fmt.Errorf("failed to create HTTP client: %v", err)
	}

	um.proxies[proxy.ID] = proxy
	um.clients[proxy.ID] = client

	return nil
}

// GetAllProxies 获取所有上游代理
func (um *UpstreamManager) GetAllProxies() []*UpstreamProxy {
	um.proxiesMutex.RLock()
	defer um.proxiesMutex.RUnlock()
	
	proxies := make([]*UpstreamProxy, 0, len(um.proxies))
	for _, proxy := range um.proxies {
		proxies = append(proxies, proxy)
	}
	return proxies
}

// MatchProxy 匹配上游代理
func (um *UpstreamManager) MatchProxy(targetURL string) (*UpstreamProxy, *http.Client) {
	um.proxiesMutex.RLock()
	defer um.proxiesMutex.RUnlock()
	
	for _, proxy := range um.proxies {
		if !proxy.Enabled {
			continue
		}
		
		matched, err := um.matchURL(targetURL, proxy)
		if err != nil {
			continue
		}
		
		if matched {
			client := um.clients[proxy.ID]
			return proxy, client
		}
	}
	
	return nil, nil
}

// matchURL 匹配URL
func (um *UpstreamManager) matchURL(targetURL string, proxy *UpstreamProxy) (bool, error) {
	if proxy.IsRegex {
		regex, err := regexp.Compile(proxy.URLPattern)
		if err != nil {
			return false, err
		}
		return regex.MatchString(targetURL), nil
	} else {
		return strings.Contains(targetURL, proxy.URLPattern), nil
	}
}

// createHTTPClient 创建HTTP客户端
func (um *UpstreamManager) createHTTPClient(proxy *UpstreamProxy, proxyURL *url.URL) (*http.Client, error) {
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}

	// 设置认证
	if proxy.Username != "" && proxy.Password != "" {
		proxyURL.User = url.UserPassword(proxy.Username, proxy.Password)
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	return client, nil
}

// UpstreamInterceptor 上游代理拦截器
type UpstreamInterceptor struct {
	manager *UpstreamManager
}

// NewUpstreamInterceptor 创建上游代理拦截器
func NewUpstreamInterceptor(manager *UpstreamManager) *UpstreamInterceptor {
	return &UpstreamInterceptor{
		manager: manager,
	}
}

// InterceptRequest 拦截请求
func (ui *UpstreamInterceptor) InterceptRequest(flow *proxycore.Flow, w http.ResponseWriter, r *http.Request) (bool, error) {
	proxy, client := ui.manager.MatchProxy(flow.URL)
	if proxy == nil || client == nil {
		return false, nil // 不使用上游代理，继续正常处理
	}

	// 使用上游代理发送请求
	flow.AddTag("upstream-proxy")
	flow.AddTag(fmt.Sprintf("upstream-%s", proxy.Name))

	// 创建新的请求
	proxyReq, err := http.NewRequest(r.Method, flow.URL, r.Body)
	if err != nil {
		return false, fmt.Errorf("failed to create upstream request: %v", err)
	}

	// 复制请求头
	for name, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Add(name, value)
		}
	}

	// 发送请求到上游代理
	resp, err := client.Do(proxyReq)
	if err != nil {
		return false, fmt.Errorf("upstream proxy request failed: %v", err)
	}
	defer resp.Body.Close()

	// 复制响应头
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	// 设置标识头
	w.Header().Set("X-ProxyWoman-Upstream", proxy.Name)

	w.WriteHeader(resp.StatusCode)

	// 复制响应体
	_, err = w.(http.ResponseWriter).Write([]byte{}) // 这里需要实际读取和写入响应体
	if err != nil {
		return false, fmt.Errorf("failed to write upstream response: %v", err)
	}

	return true, nil // 请求已通过上游代理处理
}

// TestUpstreamProxy 测试上游代理连接
func (um *UpstreamManager) TestUpstreamProxy(proxyID string) error {
	um.proxiesMutex.RLock()
	_, exists := um.proxies[proxyID]
	client, clientExists := um.clients[proxyID]
	um.proxiesMutex.RUnlock()

	if !exists || !clientExists {
		return fmt.Errorf("proxy not found: %s", proxyID)
	}

	// 创建测试请求
	testURL := "http://httpbin.org/ip"
	req, err := http.NewRequest("GET", testURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create test request: %v", err)
	}

	// 设置较短的超时时间用于测试
	testClient := &http.Client{
		Transport: client.Transport,
		Timeout:   10 * time.Second,
	}

	// 发送测试请求
	resp, err := testClient.Do(req)
	if err != nil {
		return fmt.Errorf("test request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("test request returned status: %d", resp.StatusCode)
	}

	return nil
}

// GetProxyStats 获取代理统计信息
func (um *UpstreamManager) GetProxyStats(proxyID string) map[string]interface{} {
	// 这里可以实现代理使用统计
	// 例如：请求数量、成功率、平均响应时间等
	return map[string]interface{}{
		"requests": 0,
		"success":  0,
		"errors":   0,
	}
}
