package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config 应用配置
type Config struct {
	ProxyPort    int    `json:"proxyPort"`
	ConfigDir    string `json:"configDir"`
	AutoStart    bool   `json:"autoStart"`
	Theme        string `json:"theme"`
	LogLevel     string `json:"logLevel"`
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		ProxyPort: 8080,
		AutoStart: false,
		Theme:     "dark",
		LogLevel:  "info",
	}
}

// LoadConfig 加载配置文件
func LoadConfig(configDir string) (*Config, error) {
	configPath := filepath.Join(configDir, "config.json")
	
	// 如果配置文件不存在，返回默认配置
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		config := DefaultConfig()
		config.ConfigDir = configDir
		return config, nil
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	// 解析配置
	config := DefaultConfig()
	if err := json.Unmarshal(data, config); err != nil {
		return nil, err
	}

	config.ConfigDir = configDir
	return config, nil
}

// SaveConfig 保存配置文件
func (c *Config) SaveConfig() error {
	configPath := filepath.Join(c.ConfigDir, "config.json")
	
	// 确保配置目录存在
	if err := os.MkdirAll(c.ConfigDir, 0755); err != nil {
		return err
	}

	// 序列化配置
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	// 写入文件
	return os.WriteFile(configPath, data, 0644)
}
