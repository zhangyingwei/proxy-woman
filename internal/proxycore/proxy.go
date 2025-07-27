package proxycore

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"ProxyWoman/internal/certmanager"
)

// RequestInterceptor 请求拦截器接口
type RequestInterceptor interface {
	InterceptRequest(flow *Flow, w http.ResponseWriter, r *http.Request) (handled bool, err error)
}

// ResponseInterceptor 响应拦截器接口
type ResponseInterceptor interface {
	InterceptResponse(flow *Flow, resp *http.Response) (modified *http.Response, err error)
}

// ProxyServer 代理服务器
type ProxyServer struct {
	port                int
	certManager         *certmanager.CertManager
	requestInterceptors []RequestInterceptor
	responseInterceptors []ResponseInterceptor
	server              *http.Server
	flows               map[string]*Flow
	flowsMutex          sync.RWMutex
	flowHandler         func(*Flow)
	running             bool
}

// NewProxyServer 创建新的代理服务器
func NewProxyServer(port int, certManager *certmanager.CertManager) *ProxyServer {
	return &ProxyServer{
		port:                 port,
		certManager:          certManager,
		requestInterceptors:  make([]RequestInterceptor, 0),
		responseInterceptors: make([]ResponseInterceptor, 0),
		flows:                make(map[string]*Flow),
		running:              false,
	}
}

// AddRequestInterceptor 添加请求拦截器
func (ps *ProxyServer) AddRequestInterceptor(interceptor RequestInterceptor) {
	ps.requestInterceptors = append(ps.requestInterceptors, interceptor)
}

// AddResponseInterceptor 添加响应拦截器
func (ps *ProxyServer) AddResponseInterceptor(interceptor ResponseInterceptor) {
	ps.responseInterceptors = append(ps.responseInterceptors, interceptor)
}

// SetFlowHandler 设置流量处理回调函数
func (ps *ProxyServer) SetFlowHandler(handler func(*Flow)) {
	ps.flowHandler = handler
}

// Start 启动代理服务器
func (ps *ProxyServer) Start() error {
	if ps.running {
		return fmt.Errorf("proxy server is already running")
	}

	ps.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", ps.port),
		Handler: ps,
	}

	ps.running = true

	go func() {
		if err := ps.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Proxy server error: %v\n", err)
		}
	}()

	return nil
}

// Stop 停止代理服务器
func (ps *ProxyServer) Stop() error {
	if !ps.running {
		return nil
	}

	ps.running = false
	if ps.server != nil {
		return ps.server.Close()
	}
	return nil
}

// ServeHTTP 实现 http.Handler 接口
func (ps *ProxyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		ps.handleConnect(w, r)
	} else {
		ps.handleHTTP(w, r)
	}
}

// handleHTTP 处理普通HTTP请求
func (ps *ProxyServer) handleHTTP(w http.ResponseWriter, r *http.Request) {
	// 生成Flow ID
	flowID := ps.generateFlowID()
	flow := NewFlow(flowID, r)

	// 读取请求体
	if r.Body != nil {
		body, err := io.ReadAll(r.Body)
		if err == nil {
			flow.SetRequestBody(body)
		}
		r.Body.Close()
	}

	// 执行请求拦截器
	for _, interceptor := range ps.requestInterceptors {
		handled, err := interceptor.InterceptRequest(flow, w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if handled {
			// 请求已被拦截器处理，直接返回
			ps.addFlow(flow)
			return
		}
	}

	// 创建新的请求
	targetURL := r.URL
	if !targetURL.IsAbs() {
		targetURL = &url.URL{
			Scheme:   "http",
			Host:     r.Host,
			Path:     r.URL.Path,
			RawQuery: r.URL.RawQuery,
		}
	}

	proxyReq, err := http.NewRequest(r.Method, targetURL.String(), strings.NewReader(string(flow.Request.Body)))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 复制请求头
	for name, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Add(name, value)
		}
	}

	// 发送请求
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(proxyReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 设置响应信息
	flow.SetResponse(resp, respBody)

	// 执行响应拦截器
	modifiedResp := resp
	for _, interceptor := range ps.responseInterceptors {
		modifiedResp, err = interceptor.InterceptResponse(flow, modifiedResp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if modifiedResp == nil {
			modifiedResp = resp // 如果拦截器返回nil，使用原始响应
		}
	}

	// 复制响应头到客户端
	for name, values := range modifiedResp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	w.WriteHeader(modifiedResp.StatusCode)

	// 如果响应被修改，需要重新读取响应体
	if modifiedResp != resp && modifiedResp.Body != nil {
		modifiedBody, err := io.ReadAll(modifiedResp.Body)
		if err == nil {
			w.Write(modifiedBody)
		} else {
			w.Write(respBody)
		}
	} else {
		w.Write(respBody)
	}

	// 存储并通知Flow
	ps.addFlow(flow)
}

// handleConnect 处理HTTPS CONNECT请求
func (ps *ProxyServer) handleConnect(w http.ResponseWriter, r *http.Request) {
	// 响应200 OK
	w.WriteHeader(http.StatusOK)

	// 获取底层连接
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer clientConn.Close()

	// 获取目标主机名
	host := r.Host
	if !strings.Contains(host, ":") {
		host += ":443"
	}

	hostname := strings.Split(host, ":")[0]

	// 获取服务器证书
	cert, err := ps.certManager.GenerateServerCert(hostname)
	if err != nil {
		fmt.Printf("Failed to generate certificate for %s: %v\n", hostname, err)
		return
	}

	// 创建TLS配置
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{*cert},
		ServerName:   hostname,
		MinVersion:   tls.VersionTLS12,
		MaxVersion:   tls.VersionTLS13,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
		},
		PreferServerCipherSuites: true,
	}

	// 与客户端建立TLS连接
	tlsConn := tls.Server(clientConn, tlsConfig)
	if err := tlsConn.Handshake(); err != nil {
		fmt.Printf("TLS handshake failed for %s: %v\n", hostname, err)
		fmt.Printf("Certificate details: Subject=%s, DNSNames=%v\n",
			cert.Leaf.Subject.CommonName, cert.Leaf.DNSNames)
		return
	}

	fmt.Printf("TLS handshake successful for %s\n", hostname)

	// 开始处理HTTPS流量
	ps.handleHTTPS(tlsConn, host)
}

// addFlow 添加Flow到存储并通知处理器
func (ps *ProxyServer) addFlow(flow *Flow) {
	fmt.Printf("📝 Adding flow: %s %s %s (Status: %d)\n", flow.Method, flow.URL, flow.Domain, flow.StatusCode)

	ps.flowsMutex.Lock()
	ps.flows[flow.ID] = flow
	ps.flowsMutex.Unlock()

	if ps.flowHandler != nil {
		fmt.Printf("📤 Notifying flow handler for: %s\n", flow.URL)
		ps.flowHandler(flow)
	} else {
		fmt.Printf("⚠️  No flow handler registered!\n")
	}
}

// GetFlows 获取所有Flow
func (ps *ProxyServer) GetFlows() []*Flow {
	ps.flowsMutex.RLock()
	defer ps.flowsMutex.RUnlock()

	flows := make([]*Flow, 0, len(ps.flows))
	for _, flow := range ps.flows {
		flows = append(flows, flow)
	}
	return flows
}

// GetFlow 根据ID获取单个Flow
func (ps *ProxyServer) GetFlow(flowID string) (*Flow, bool) {
	ps.flowsMutex.RLock()
	defer ps.flowsMutex.RUnlock()

	flow, exists := ps.flows[flowID]
	return flow, exists
}

// ClearFlows 清空所有Flow
func (ps *ProxyServer) ClearFlows() {
	ps.flowsMutex.Lock()
	defer ps.flowsMutex.Unlock()
	ps.flows = make(map[string]*Flow)
}

// generateFlowID 生成Flow ID
func (ps *ProxyServer) generateFlowID() string {
	return fmt.Sprintf("flow_%d", time.Now().UnixNano())
}

// PinFlow 钉住/取消钉住流量
func (ps *ProxyServer) PinFlow(flowID string) error {
	ps.flowsMutex.Lock()
	defer ps.flowsMutex.Unlock()

	flow, exists := ps.flows[flowID]
	if !exists {
		return fmt.Errorf("flow not found: %s", flowID)
	}

	flow.IsPinned = !flow.IsPinned
	return nil
}

// GetPinnedFlows 获取所有钉住的流量
func (ps *ProxyServer) GetPinnedFlows() []*Flow {
	ps.flowsMutex.RLock()
	defer ps.flowsMutex.RUnlock()

	var pinnedFlows []*Flow
	for _, flow := range ps.flows {
		if flow.IsPinned {
			pinnedFlows = append(pinnedFlows, flow)
		}
	}
	return pinnedFlows
}

// FilterFlows 根据条件过滤流量
func (ps *ProxyServer) FilterFlows(filter func(*Flow) bool) []*Flow {
	ps.flowsMutex.RLock()
	defer ps.flowsMutex.RUnlock()

	var filteredFlows []*Flow
	for _, flow := range ps.flows {
		if filter(flow) {
			filteredFlows = append(filteredFlows, flow)
		}
	}
	return filteredFlows
}

// GetFlowByID 根据ID获取流量
func (ps *ProxyServer) GetFlowByID(flowID string) (*Flow, error) {
	ps.flowsMutex.RLock()
	defer ps.flowsMutex.RUnlock()

	flow, exists := ps.flows[flowID]
	if !exists {
		return nil, fmt.Errorf("flow not found: %s", flowID)
	}
	return flow, nil
}

// handleHTTPS 处理HTTPS流量
func (ps *ProxyServer) handleHTTPS(tlsConn *tls.Conn, targetHost string) {
	defer tlsConn.Close()

	fmt.Printf("Starting HTTPS handler for %s\n", targetHost)

	// 创建HTTP服务器来处理解密后的HTTPS请求
	server := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Printf("🔍 Received HTTPS request: %s %s from %s\n", r.Method, r.URL.Path, targetHost)
			fmt.Printf("🔍 Request headers: %v\n", r.Header)

			// 设置完整的URL
			r.URL.Scheme = "https"
			r.URL.Host = targetHost

			// 处理为普通HTTP请求
			fmt.Printf("🔍 Calling handleHTTP for HTTPS request\n")
			ps.handleHTTP(w, r)
			fmt.Printf("🔍 handleHTTP completed for HTTPS request\n")
		}),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// 使用TLS连接处理HTTP请求
	listener := &singleConnListener{conn: tlsConn}
	err := server.Serve(listener)
	if err != nil && err != io.EOF && !strings.Contains(err.Error(), "use of closed network connection") {
		fmt.Printf("HTTPS server error for %s: %v\n", targetHost, err)
	}

	fmt.Printf("HTTPS handler finished for %s\n", targetHost)
}



// singleConnListener 单连接监听器
type singleConnListener struct {
	conn   net.Conn
	once   sync.Once
	closed chan struct{}
}

func (l *singleConnListener) Accept() (net.Conn, error) {
	var conn net.Conn

	l.once.Do(func() {
		if l.closed == nil {
			l.closed = make(chan struct{})
		}
		conn = l.conn
	})

	if conn != nil {
		return conn, nil
	}

	// 等待关闭信号
	<-l.closed
	return nil, io.EOF
}

func (l *singleConnListener) Close() error {
	if l.closed != nil {
		close(l.closed)
	}
	if l.conn != nil {
		return l.conn.Close()
	}
	return nil
}

func (l *singleConnListener) Addr() net.Addr {
	if l.conn != nil {
		return l.conn.LocalAddr()
	}
	return nil
}
