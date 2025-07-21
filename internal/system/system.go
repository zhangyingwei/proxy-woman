package system

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// SystemManager 系统集成管理器
type SystemManager struct {
	configDir string
}

// NewSystemManager 创建新的系统管理器
func NewSystemManager() *SystemManager {
	configDir := getConfigDir()
	return &SystemManager{
		configDir: configDir,
	}
}

// GetConfigDir 获取配置目录
func (sm *SystemManager) GetConfigDir() string {
	return sm.configDir
}

// SetSystemProxy 设置系统代理
func (sm *SystemManager) SetSystemProxy(port int) error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("system proxy setting is only supported on macOS")
	}

	proxyURL := fmt.Sprintf("127.0.0.1:%d", port)

	// 获取网络服务列表
	services, err := sm.getNetworkServices()
	if err != nil {
		return fmt.Errorf("failed to get network services: %v", err)
	}

	// 为每个网络服务设置代理
	for _, service := range services {
		// 设置HTTP代理
		if err := sm.setWebProxy(service, proxyURL); err != nil {
			fmt.Printf("Warning: failed to set HTTP proxy for %s: %v\n", service, err)
		}

		// 设置HTTPS代理
		if err := sm.setSecureWebProxy(service, proxyURL); err != nil {
			fmt.Printf("Warning: failed to set HTTPS proxy for %s: %v\n", service, err)
		}
	}

	return nil
}

// DisableSystemProxy 禁用系统代理
func (sm *SystemManager) DisableSystemProxy() error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("system proxy setting is only supported on macOS")
	}

	// 获取网络服务列表
	services, err := sm.getNetworkServices()
	if err != nil {
		return fmt.Errorf("failed to get network services: %v", err)
	}

	// 为每个网络服务禁用代理
	for _, service := range services {
		// 禁用HTTP代理
		if err := sm.disableWebProxy(service); err != nil {
			fmt.Printf("Warning: failed to disable HTTP proxy for %s: %v\n", service, err)
		}

		// 禁用HTTPS代理
		if err := sm.disableSecureWebProxy(service); err != nil {
			fmt.Printf("Warning: failed to disable HTTPS proxy for %s: %v\n", service, err)
		}
	}

	return nil
}

// getNetworkServices 获取网络服务列表
func (sm *SystemManager) getNetworkServices() ([]string, error) {
	cmd := exec.Command("networksetup", "-listallnetworkservices")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := splitLines(string(output))
	var services []string

	// 跳过第一行（标题行）
	for i := 1; i < len(lines); i++ {
		line := lines[i]
		if line != "" && line[0] != '*' { // 跳过禁用的服务
			services = append(services, line)
		}
	}

	return services, nil
}

// setWebProxy 设置HTTP代理
func (sm *SystemManager) setWebProxy(service, proxyURL string) error {
	cmd := exec.Command("networksetup", "-setwebproxy", service, "127.0.0.1", fmt.Sprintf("%s", getPortFromURL(proxyURL)))
	return cmd.Run()
}

// setSecureWebProxy 设置HTTPS代理
func (sm *SystemManager) setSecureWebProxy(service, proxyURL string) error {
	cmd := exec.Command("networksetup", "-setsecurewebproxy", service, "127.0.0.1", fmt.Sprintf("%s", getPortFromURL(proxyURL)))
	return cmd.Run()
}

// disableWebProxy 禁用HTTP代理
func (sm *SystemManager) disableWebProxy(service string) error {
	cmd := exec.Command("networksetup", "-setwebproxystate", service, "off")
	return cmd.Run()
}

// disableSecureWebProxy 禁用HTTPS代理
func (sm *SystemManager) disableSecureWebProxy(service string) error {
	cmd := exec.Command("networksetup", "-setsecurewebproxystate", service, "off")
	return cmd.Run()
}

// getConfigDir 获取应用配置目录
func getConfigDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ".proxywoman"
	}

	configDir := filepath.Join(homeDir, ".proxywoman")
	os.MkdirAll(configDir, 0755)
	return configDir
}

// splitLines 分割字符串为行
func splitLines(s string) []string {
	var lines []string
	var line string

	for _, char := range s {
		if char == '\n' || char == '\r' {
			if line != "" {
				lines = append(lines, line)
				line = ""
			}
		} else {
			line += string(char)
		}
	}

	if line != "" {
		lines = append(lines, line)
	}

	return lines
}

// getPortFromURL 从URL中提取端口
func getPortFromURL(url string) string {
	// 简单的端口提取，假设格式为 "127.0.0.1:port"
	for i := len(url) - 1; i >= 0; i-- {
		if url[i] == ':' {
			return url[i+1:]
		}
	}
	return "8080" // 默认端口
}

// WriteFile 写入文件
func (sm *SystemManager) WriteFile(filename string, data []byte) error {
	// 确保目录存在
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
