package certmanager

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"time"
)

// CertTester 证书测试器
type CertTester struct {
	certManager *CertManager
}

// NewCertTester 创建证书测试器
func NewCertTester(cm *CertManager) *CertTester {
	return &CertTester{
		certManager: cm,
	}
}

// loadCACert 加载CA证书
func (ct *CertTester) loadCACert() (*x509.Certificate, error) {
	caPath := ct.certManager.GetCACertPath()

	// 读取证书文件
	certPEM, err := os.ReadFile(caPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %v", err)
	}

	// 解析PEM
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode CA certificate PEM")
	}

	// 解析证书
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CA certificate: %v", err)
	}

	return cert, nil
}

// TestCertificateGeneration 测试证书生成
func (ct *CertTester) TestCertificateGeneration(hostname string) error {
	fmt.Printf("Testing certificate generation for: %s\n", hostname)
	
	// 生成证书
	cert, err := ct.certManager.GenerateServerCert(hostname)
	if err != nil {
		return fmt.Errorf("failed to generate certificate: %v", err)
	}
	
	// 验证证书
	if cert.Leaf == nil {
		return fmt.Errorf("certificate leaf is nil")
	}
	
	fmt.Printf("Certificate generated successfully:\n")
	fmt.Printf("  Subject: %s\n", cert.Leaf.Subject.CommonName)
	fmt.Printf("  DNS Names: %v\n", cert.Leaf.DNSNames)
	fmt.Printf("  IP Addresses: %v\n", cert.Leaf.IPAddresses)
	fmt.Printf("  Valid From: %s\n", cert.Leaf.NotBefore.Format(time.RFC3339))
	fmt.Printf("  Valid To: %s\n", cert.Leaf.NotAfter.Format(time.RFC3339))
	
	return nil
}

// TestTLSHandshake 测试TLS握手
func (ct *CertTester) TestTLSHandshake(hostname string, port int) error {
	fmt.Printf("Testing TLS handshake for: %s:%d\n", hostname, port)
	
	// 创建TLS配置，跳过证书验证（用于测试）
	config := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         hostname,
	}
	
	// 连接到服务器
	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", hostname, port), config)
	if err != nil {
		return fmt.Errorf("TLS dial failed: %v", err)
	}
	defer conn.Close()
	
	// 获取连接状态
	state := conn.ConnectionState()
	fmt.Printf("TLS handshake successful:\n")
	fmt.Printf("  Version: %x\n", state.Version)
	fmt.Printf("  Cipher Suite: %x\n", state.CipherSuite)
	fmt.Printf("  Server Certificates: %d\n", len(state.PeerCertificates))
	
	if len(state.PeerCertificates) > 0 {
		cert := state.PeerCertificates[0]
		fmt.Printf("  Server Cert Subject: %s\n", cert.Subject.CommonName)
		fmt.Printf("  Server Cert DNS Names: %v\n", cert.DNSNames)
	}
	
	return nil
}

// TestCertificateChain 测试证书链验证
func (ct *CertTester) TestCertificateChain(hostname string) error {
	fmt.Printf("Testing certificate chain for: %s\n", hostname)
	
	// 生成服务器证书
	serverCert, err := ct.certManager.GenerateServerCert(hostname)
	if err != nil {
		return fmt.Errorf("failed to generate server certificate: %v", err)
	}
	
	// 加载CA证书
	caCert, err := ct.loadCACert()
	if err != nil {
		return fmt.Errorf("failed to load CA certificate: %v", err)
	}
	
	// 创建证书池
	roots := x509.NewCertPool()
	roots.AddCert(caCert)
	
	// 验证证书链
	opts := x509.VerifyOptions{
		Roots:     roots,
		DNSName:   hostname,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	
	chains, err := serverCert.Leaf.Verify(opts)
	if err != nil {
		return fmt.Errorf("certificate chain verification failed: %v", err)
	}
	
	fmt.Printf("Certificate chain verification successful:\n")
	fmt.Printf("  Chains found: %d\n", len(chains))
	for i, chain := range chains {
		fmt.Printf("  Chain %d length: %d\n", i, len(chain))
		for j, cert := range chain {
			fmt.Printf("    Cert %d: %s\n", j, cert.Subject.CommonName)
		}
	}
	
	return nil
}

// TestLocalTLSServer 测试本地TLS服务器
func (ct *CertTester) TestLocalTLSServer(hostname string, port int) error {
	fmt.Printf("Testing local TLS server for: %s:%d\n", hostname, port)
	
	// 生成证书
	cert, err := ct.certManager.GenerateServerCert(hostname)
	if err != nil {
		return fmt.Errorf("failed to generate certificate: %v", err)
	}
	
	// 创建TLS配置
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{*cert},
		ServerName:   hostname,
	}
	
	// 创建监听器
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to create listener: %v", err)
	}
	defer listener.Close()
	
	// 包装为TLS监听器
	tlsListener := tls.NewListener(listener, tlsConfig)
	
	fmt.Printf("TLS server started on %s:%d\n", hostname, port)
	fmt.Printf("Test with: curl -k https://%s:%d/\n", hostname, port)
	
	// 接受一个连接进行测试
	go func() {
		conn, err := tlsListener.Accept()
		if err != nil {
			fmt.Printf("Failed to accept connection: %v\n", err)
			return
		}
		defer conn.Close()
		
		fmt.Printf("TLS connection accepted from: %s\n", conn.RemoteAddr())
		
		// 发送简单的HTTP响应
		response := "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: 25\r\n\r\nProxyWoman TLS Test OK!\r\n"
		conn.Write([]byte(response))
	}()
	
	// 等待一段时间让测试完成
	time.Sleep(5 * time.Second)
	
	return nil
}

// RunAllTests 运行所有测试
func (ct *CertTester) RunAllTests(hostname string, port int) {
	fmt.Printf("=== Running Certificate Tests for %s ===\n\n", hostname)
	
	tests := []struct {
		name string
		fn   func() error
	}{
		{"Certificate Generation", func() error { return ct.TestCertificateGeneration(hostname) }},
		{"Certificate Chain", func() error { return ct.TestCertificateChain(hostname) }},
		{"Local TLS Server", func() error { return ct.TestLocalTLSServer(hostname, port) }},
	}
	
	for _, test := range tests {
		fmt.Printf("--- %s ---\n", test.name)
		if err := test.fn(); err != nil {
			fmt.Printf("❌ FAILED: %v\n\n", err)
		} else {
			fmt.Printf("✅ PASSED\n\n")
		}
	}
	
	fmt.Printf("=== Tests Completed ===\n")
}
