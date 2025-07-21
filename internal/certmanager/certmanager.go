package certmanager

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// CertManager 管理根证书和动态生成的服务器证书
type CertManager struct {
	caCert     *x509.Certificate
	caKey      *rsa.PrivateKey
	certCache  map[string]*tls.Certificate
	cacheMutex sync.RWMutex
	configDir  string
}

// NewCertManager 创建新的证书管理器
func NewCertManager(configDir string) *CertManager {
	return &CertManager{
		certCache: make(map[string]*tls.Certificate),
		configDir: configDir,
	}
}

// InitCA 初始化根证书，如果不存在则创建
func (cm *CertManager) InitCA() error {
	caPath := filepath.Join(cm.configDir, "ca.crt")
	keyPath := filepath.Join(cm.configDir, "ca.key")

	// 检查证书是否已存在
	if _, err := os.Stat(caPath); err == nil {
		return cm.LoadCA()
	}

	// 创建配置目录
	if err := os.MkdirAll(cm.configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	// 生成根证书
	return cm.generateCA(caPath, keyPath)
}

// LoadCA 加载已存在的根证书
func (cm *CertManager) LoadCA() error {
	caPath := filepath.Join(cm.configDir, "ca.crt")
	keyPath := filepath.Join(cm.configDir, "ca.key")

	// 读取证书文件
	certPEM, err := os.ReadFile(caPath)
	if err != nil {
		return fmt.Errorf("failed to read CA certificate: %v", err)
	}

	// 读取私钥文件
	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("failed to read CA private key: %v", err)
	}

	// 解析证书
	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil {
		return fmt.Errorf("failed to decode CA certificate")
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse CA certificate: %v", err)
	}

	// 解析私钥
	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		return fmt.Errorf("failed to decode CA private key")
	}

	key, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse CA private key: %v", err)
	}

	cm.caCert = cert
	cm.caKey = key

	return nil
}

// GenerateServerCert 为指定主机名生成服务器证书
func (cm *CertManager) GenerateServerCert(hostname string) (*tls.Certificate, error) {
	cm.cacheMutex.RLock()
	if cert, exists := cm.certCache[hostname]; exists {
		cm.cacheMutex.RUnlock()
		return cert, nil
	}
	cm.cacheMutex.RUnlock()

	cm.cacheMutex.Lock()
	defer cm.cacheMutex.Unlock()

	// 双重检查
	if cert, exists := cm.certCache[hostname]; exists {
		return cert, nil
	}

	// 生成新的服务器证书
	cert, err := cm.generateServerCert(hostname)
	if err != nil {
		return nil, err
	}

	cm.certCache[hostname] = cert
	return cert, nil
}

// GetCACertPath 获取根证书文件路径
func (cm *CertManager) GetCACertPath() string {
	return filepath.Join(cm.configDir, "ca.crt")
}

// generateCA 生成根证书和私钥
func (cm *CertManager) generateCA(caPath, keyPath string) error {
	// 生成私钥
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate CA private key: %v", err)
	}

	// 创建证书模板
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:       []string{"ProxyWoman CA"},
			OrganizationalUnit: []string{"ProxyWoman Root CA"},
			Country:            []string{"US"},
			CommonName:         "ProxyWoman Root CA",
		},
		NotBefore:             time.Now().Add(-24 * time.Hour), // 提前1天生效，避免时钟偏差
		NotAfter:              time.Now().Add(365 * 24 * time.Hour * 10), // 10年有效期
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            2,
		MaxPathLenZero:        false,
	}

	// 生成证书
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return fmt.Errorf("failed to create CA certificate: %v", err)
	}

	// 保存证书
	certOut, err := os.Create(caPath)
	if err != nil {
		return fmt.Errorf("failed to create CA certificate file: %v", err)
	}
	defer certOut.Close()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certDER}); err != nil {
		return fmt.Errorf("failed to write CA certificate: %v", err)
	}

	// 保存私钥
	keyOut, err := os.Create(keyPath)
	if err != nil {
		return fmt.Errorf("failed to create CA private key file: %v", err)
	}
	defer keyOut.Close()

	keyBytes := x509.MarshalPKCS1PrivateKey(key)
	if err := pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: keyBytes}); err != nil {
		return fmt.Errorf("failed to write CA private key: %v", err)
	}

	// 解析生成的证书
	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return fmt.Errorf("failed to parse generated CA certificate: %v", err)
	}

	cm.caCert = cert
	cm.caKey = key

	return nil
}

// generateServerCert 为指定主机名生成服务器证书
func (cm *CertManager) generateServerCert(hostname string) (*tls.Certificate, error) {
	if cm.caCert == nil || cm.caKey == nil {
		return nil, fmt.Errorf("CA certificate not initialized")
	}

	// 生成服务器私钥
	serverKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate server private key: %v", err)
	}

	// 准备DNS名称和IP地址
	var dnsNames []string
	var ipAddresses []net.IP

	// 添加主机名
	dnsNames = append(dnsNames, hostname)

	// 如果是IP地址，添加到IP列表
	if ip := net.ParseIP(hostname); ip != nil {
		ipAddresses = append(ipAddresses, ip)
	} else {
		// 如果是域名，添加通配符版本
		if !strings.HasPrefix(hostname, "*.") {
			wildcardDomain := "*." + hostname
			dnsNames = append(dnsNames, wildcardDomain)
		}

		// 添加常见的子域名
		subdomains := []string{"www." + hostname}
		for _, subdomain := range subdomains {
			if subdomain != hostname {
				dnsNames = append(dnsNames, subdomain)
			}
		}
	}

	// 添加localhost相关
	if hostname == "localhost" || hostname == "127.0.0.1" {
		dnsNames = append(dnsNames, "localhost")
		ipAddresses = append(ipAddresses, net.ParseIP("127.0.0.1"))
		ipAddresses = append(ipAddresses, net.ParseIP("::1"))
	}

	// 创建服务器证书模板
	template := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject: pkix.Name{
			Organization:       []string{"ProxyWoman"},
			OrganizationalUnit: []string{"ProxyWoman Server"},
			Country:            []string{"US"},
			CommonName:         hostname,
		},
		DNSNames:              dnsNames,
		IPAddresses:           ipAddresses,
		NotBefore:             time.Now().Add(-24 * time.Hour), // 提前1天生效
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // 1年有效期
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
	}

	// 生成服务器证书
	certDER, err := x509.CreateCertificate(rand.Reader, &template, cm.caCert, &serverKey.PublicKey, cm.caKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create server certificate: %v", err)
	}

	// 创建 TLS 证书
	cert := &tls.Certificate{
		Certificate: [][]byte{certDER},
		PrivateKey:  serverKey,
	}

	return cert, nil
}

// GetCACertInstallInstructions 获取CA证书安装说明
func (cm *CertManager) GetCACertInstallInstructions() string {
	certPath := cm.GetCACertPath()
	return fmt.Sprintf(`
CA证书安装说明:

证书路径: %s

macOS 安装步骤:
1. 双击证书文件打开钥匙串访问
2. 选择"系统"钥匙串
3. 找到"ProxyWoman Root CA"证书
4. 双击证书，展开"信任"部分
5. 将"使用此证书时"设置为"始终信任"
6. 关闭窗口并输入管理员密码

Windows 安装步骤:
1. 右键点击证书文件，选择"安装证书"
2. 选择"本地计算机"
3. 选择"将所有的证书都放入下列存储"
4. 点击"浏览"，选择"受信任的根证书颁发机构"
5. 点击"下一步"和"完成"

Linux 安装步骤:
1. 复制证书到系统证书目录:
   sudo cp %s /usr/local/share/ca-certificates/proxywoman.crt
2. 更新证书存储:
   sudo update-ca-certificates

注意: 安装证书后需要重启浏览器才能生效
`, certPath, certPath)
}

// IsCACertInstalled 检查CA证书是否已安装（简单检查）
func (cm *CertManager) IsCACertInstalled() bool {
	// 这里只是检查证书文件是否存在
	// 实际的信任状态检查需要平台特定的代码
	_, err := os.Stat(cm.GetCACertPath())
	return err == nil
}
