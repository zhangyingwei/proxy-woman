package main

import (
	"context"
	"fmt"

	"ProxyWoman/internal/certmanager"
	"ProxyWoman/internal/config"
	"ProxyWoman/internal/export"
	"ProxyWoman/internal/features"
	"ProxyWoman/internal/logger"
	"ProxyWoman/internal/proxycore"
	"ProxyWoman/internal/storage"
	"ProxyWoman/internal/system"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx            context.Context
	config         *config.Config
	proxyServer    *proxycore.ProxyServer
	certManager    *certmanager.CertManager
	systemManager  *system.SystemManager
	featureManager *features.FeatureManager
	exportService  *export.ExportService
	database       *storage.Database
	isRunning      bool
}

// NewApp creates a new App application struct
func NewApp() *App {
	systemManager := system.NewSystemManager()

	// 加载配置
	cfg, err := config.LoadConfig(systemManager.GetConfigDir())
	if err != nil {
		fmt.Printf("Failed to load config, using defaults: %v\n", err)
		cfg = config.DefaultConfig()
		cfg.ConfigDir = systemManager.GetConfigDir()
	}

	// 初始化日志
	if err := logger.InitLogger(cfg.ConfigDir, cfg.LogLevel); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
	}

	// 初始化数据库
	database, err := storage.NewDatabase()
	if err != nil {
		fmt.Printf("Failed to initialize database: %v\n", err)
		// 继续运行，但功能会受限
	}

	certManager := certmanager.NewCertManager(cfg.ConfigDir)
	featureManager := features.NewFeatureManager(database)
	exportService := export.NewExportService()

	app := &App{
		config:         cfg,
		certManager:    certManager,
		systemManager:  systemManager,
		featureManager: featureManager,
		exportService:  exportService,
		database:       database,
		isRunning:      false,
	}

	return app
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// 初始化证书管理器
	if err := a.certManager.InitCA(); err != nil {
		logger.Error("Failed to initialize CA: %v", err)
		fmt.Printf("Failed to initialize CA: %v\n", err)
	}

	// 创建代理服务器
	a.proxyServer = proxycore.NewProxyServer(a.config.ProxyPort, a.certManager)

	// 设置拦截器
	allowBlockInterceptor := features.NewAllowBlockInterceptor(a.featureManager.AllowBlock)
	mapLocalInterceptor := features.NewMapLocalInterceptor(a.featureManager.MapLocal)
	breakpointInterceptor := features.NewBreakpointInterceptor(a.featureManager.Breakpoint)
	scriptInterceptor := features.NewScriptInterceptor(a.featureManager.Scripting)

	// 设置断点事件处理器
	breakpointInterceptor.SetEventHandler(func(session *features.BreakpointSession) {
		// 通过Wails事件系统发送断点事件到前端
		runtime.EventsEmit(ctx, "breakpoint-hit", session)
	})

	// 创建更多拦截器
	upstreamInterceptor := features.NewUpstreamInterceptor(a.featureManager.Upstream)
	reverseProxyInterceptor := features.NewReverseProxyInterceptor(a.featureManager.ReverseProxy)

	// 按顺序添加拦截器（顺序很重要）
	a.proxyServer.AddRequestInterceptor(allowBlockInterceptor)     // 首先检查允许/阻止
	a.proxyServer.AddRequestInterceptor(reverseProxyInterceptor)   // 然后检查反向代理
	a.proxyServer.AddRequestInterceptor(upstreamInterceptor)       // 然后检查上游代理
	a.proxyServer.AddRequestInterceptor(mapLocalInterceptor)       // 然后检查Map Local
	a.proxyServer.AddRequestInterceptor(breakpointInterceptor)     // 然后检查断点
	a.proxyServer.AddRequestInterceptor(scriptInterceptor)        // 最后执行脚本

	a.proxyServer.AddResponseInterceptor(breakpointInterceptor)    // 响应断点
	a.proxyServer.AddResponseInterceptor(scriptInterceptor)        // 响应脚本

	// 设置流量处理回调
	a.proxyServer.SetFlowHandler(func(flow *proxycore.Flow) {
		// 通过Wails事件系统发送新的流量到前端
		runtime.EventsEmit(ctx, "new-flow", flow)
	})
}

// StartProxy 启动代理服务器
func (a *App) StartProxy() error {
	if a.isRunning {
		return fmt.Errorf("proxy is already running")
	}

	// 启动代理服务器
	if err := a.proxyServer.Start(); err != nil {
		logger.Error("Failed to start proxy server: %v", err)
		return fmt.Errorf("failed to start proxy server: %v", err)
	}

	// 设置系统代理
	if err := a.systemManager.SetSystemProxy(a.config.ProxyPort); err != nil {
		// 如果设置系统代理失败，仍然继续运行，但记录错误
		logger.Warn("Failed to set system proxy: %v", err)
		fmt.Printf("Warning: failed to set system proxy: %v\n", err)
	}

	a.isRunning = true
	return nil
}

// StopProxy 停止代理服务器
func (a *App) StopProxy() error {
	if !a.isRunning {
		return nil
	}

	// 禁用系统代理
	if err := a.systemManager.DisableSystemProxy(); err != nil {
		logger.Warn("Failed to disable system proxy: %v", err)
		fmt.Printf("Warning: failed to disable system proxy: %v\n", err)
	}

	// 停止代理服务器
	if err := a.proxyServer.Stop(); err != nil {
		logger.Error("Failed to stop proxy server: %v", err)
		return fmt.Errorf("failed to stop proxy server: %v", err)
	}

	a.isRunning = false
	return nil
}

// IsProxyRunning 检查代理是否正在运行
func (a *App) IsProxyRunning() bool {
	return a.isRunning
}

// GetFlows 获取所有流量记录
func (a *App) GetFlows() []*proxycore.Flow {
	if a.proxyServer == nil {
		return []*proxycore.Flow{}
	}
	return a.proxyServer.GetFlows()
}

// ClearFlows 清空所有流量记录
func (a *App) ClearFlows() {
	if a.proxyServer != nil {
		a.proxyServer.ClearFlows()
	}
}

// GetCACertPath 获取根证书路径
func (a *App) GetCACertPath() string {
	return a.certManager.GetCACertPath()
}

// GetCACertInstallInstructions 获取CA证书安装说明
func (a *App) GetCACertInstallInstructions() string {
	return a.certManager.GetCACertInstallInstructions()
}

// IsCACertInstalled 检查CA证书是否已安装
func (a *App) IsCACertInstalled() bool {
	return a.certManager.IsCACertInstalled()
}

// GetProxyPort 获取代理端口
func (a *App) GetProxyPort() int {
	return a.config.ProxyPort
}

// PinFlow 钉住/取消钉住流量
func (a *App) PinFlow(flowID string) error {
	if a.proxyServer == nil {
		return fmt.Errorf("proxy server not initialized")
	}
	return a.proxyServer.PinFlow(flowID)
}

// GetPinnedFlows 获取所有钉住的流量
func (a *App) GetPinnedFlows() []*proxycore.Flow {
	if a.proxyServer == nil {
		return []*proxycore.Flow{}
	}
	return a.proxyServer.GetPinnedFlows()
}

// GetFlowByID 根据ID获取流量
func (a *App) GetFlowByID(flowID string) (*proxycore.Flow, error) {
	if a.proxyServer == nil {
		return nil, fmt.Errorf("proxy server not initialized")
	}
	return a.proxyServer.GetFlowByID(flowID)
}

// Map Local 相关方法

// AddMapLocalRule 添加Map Local规则
func (a *App) AddMapLocalRule(rule *features.MapLocalRule) {
	a.featureManager.MapLocal.AddRule(rule)
}

// RemoveMapLocalRule 移除Map Local规则
func (a *App) RemoveMapLocalRule(ruleID string) {
	a.featureManager.MapLocal.RemoveRule(ruleID)
}

// GetMapLocalRules 获取所有Map Local规则
func (a *App) GetMapLocalRules() []*features.MapLocalRule {
	return a.featureManager.MapLocal.GetAllRules()
}

// UpdateMapLocalRule 更新Map Local规则
func (a *App) UpdateMapLocalRule(rule *features.MapLocalRule) error {
	return a.featureManager.MapLocal.UpdateRule(rule)
}

// 断点相关方法

// AddBreakpointRule 添加断点规则
func (a *App) AddBreakpointRule(rule *features.BreakpointRule) error {
	return a.featureManager.Breakpoint.AddRule(rule)
}

// RemoveBreakpointRule 移除断点规则
func (a *App) RemoveBreakpointRule(ruleID string) error {
	return a.featureManager.Breakpoint.RemoveRule(ruleID)
}

// UpdateBreakpointRuleStatus 更新断点规则状态
func (a *App) UpdateBreakpointRuleStatus(ruleID string, enabled bool) error {
	return a.featureManager.Breakpoint.UpdateRuleStatus(ruleID, enabled)
}

// GetBreakpointRules 获取所有断点规则
func (a *App) GetBreakpointRules() []*features.BreakpointRule {
	return a.featureManager.Breakpoint.GetAllRules()
}

// ResumeBreakpoint 恢复断点
func (a *App) ResumeBreakpoint(sessionID string) error {
	return a.featureManager.Breakpoint.ResumeBreakpoint(sessionID, nil, nil)
}

// CancelBreakpoint 取消断点
func (a *App) CancelBreakpoint(sessionID string) error {
	return a.featureManager.Breakpoint.CancelBreakpoint(sessionID)
}

// GetActiveBreakpoints 获取活跃的断点
func (a *App) GetActiveBreakpoints() []*features.BreakpointSession {
	return a.featureManager.Breakpoint.GetActiveBreakpoints()
}

// 重放相关方法

// ReplayFlow 重放Flow
func (a *App) ReplayFlow(flowID string) (*features.ReplayResponse, error) {
	flow, err := a.GetFlowByID(flowID)
	if err != nil {
		return nil, err
	}
	return a.featureManager.Replay.ReplayFlow(flow)
}

// SendCustomRequest 发送自定义请求
func (a *App) SendCustomRequest(request *features.ReplayRequest) (*features.ReplayResponse, error) {
	return a.featureManager.Replay.SendRequest(request)
}

// ModifyAndReplayFlow 修改并重放Flow
func (a *App) ModifyAndReplayFlow(flowID string, modifications map[string]interface{}) (*features.ReplayResponse, error) {
	flow, err := a.GetFlowByID(flowID)
	if err != nil {
		return nil, err
	}
	return a.featureManager.Replay.ModifyAndSendRequest(flow, modifications)
}

// 脚本相关方法

// AddScript 添加脚本
func (a *App) AddScript(script *features.Script) error {
	return a.featureManager.Scripting.AddScript(script)
}

// RemoveScript 移除脚本
func (a *App) RemoveScript(scriptID string) error {
	return a.featureManager.Scripting.RemoveScript(scriptID)
}

// UpdateScript 更新脚本
func (a *App) UpdateScript(script *features.Script) error {
	return a.featureManager.Scripting.UpdateScript(script)
}

// UpdateScriptStatus 更新脚本状态
func (a *App) UpdateScriptStatus(scriptID string, enabled bool) error {
	return a.featureManager.Scripting.UpdateScriptStatus(scriptID, enabled)
}

// GetAllScripts 获取所有脚本
func (a *App) GetAllScripts() []*features.Script {
	return a.featureManager.Scripting.GetAllScripts()
}

// ValidateScript 验证脚本语法
func (a *App) ValidateScript(content string) error {
	return a.featureManager.Scripting.ValidateScript(content)
}

// 允许/阻止列表相关方法

// AddAllowBlockRule 添加允许/阻止规则
func (a *App) AddAllowBlockRule(rule *features.AllowBlockRule) {
	a.featureManager.AllowBlock.AddRule(rule)
}

// RemoveAllowBlockRule 移除允许/阻止规则
func (a *App) RemoveAllowBlockRule(ruleID string) {
	a.featureManager.AllowBlock.RemoveRule(ruleID)
}

// UpdateAllowBlockRule 更新允许/阻止规则
func (a *App) UpdateAllowBlockRule(rule *features.AllowBlockRule) error {
	return a.featureManager.AllowBlock.UpdateRule(rule)
}

// GetAllowBlockRules 获取所有允许/阻止规则
func (a *App) GetAllowBlockRules() []*features.AllowBlockRule {
	return a.featureManager.AllowBlock.GetAllRules()
}

// SetAllowBlockMode 设置允许/阻止模式
func (a *App) SetAllowBlockMode(mode string) error {
	return a.featureManager.AllowBlock.SetMode(mode)
}

// GetAllowBlockMode 获取允许/阻止模式
func (a *App) GetAllowBlockMode() string {
	return a.featureManager.AllowBlock.GetMode()
}

// HAR相关方法

// ExportFlowsToHAR 导出Flows到HAR文件
func (a *App) ExportFlowsToHAR(filePath string) error {
	flows := a.GetFlows()
	return a.featureManager.HAR.ExportFlowsToHAR(flows, filePath)
}

// ImportHARToFlows 从HAR文件导入Flows
func (a *App) ImportHARToFlows(filePath string) ([]*proxycore.Flow, error) {
	return a.featureManager.HAR.ImportHARToFlows(filePath)
}

// 反向代理相关方法

// AddReverseProxyRule 添加反向代理规则
func (a *App) AddReverseProxyRule(rule *features.ReverseProxyRule) error {
	return a.featureManager.ReverseProxy.AddRule(rule)
}

// RemoveReverseProxyRule 移除反向代理规则
func (a *App) RemoveReverseProxyRule(ruleID string) {
	a.featureManager.ReverseProxy.RemoveRule(ruleID)
}

// UpdateReverseProxyRule 更新反向代理规则
func (a *App) UpdateReverseProxyRule(rule *features.ReverseProxyRule) error {
	return a.featureManager.ReverseProxy.UpdateRule(rule)
}

// GetReverseProxyRules 获取所有反向代理规则
func (a *App) GetReverseProxyRules() []*features.ReverseProxyRule {
	return a.featureManager.ReverseProxy.GetAllRules()
}

// 上游代理相关方法

// AddUpstreamProxy 添加上游代理
func (a *App) AddUpstreamProxy(proxy *features.UpstreamProxy) error {
	return a.featureManager.Upstream.AddProxy(proxy)
}

// RemoveUpstreamProxy 移除上游代理
func (a *App) RemoveUpstreamProxy(proxyID string) {
	a.featureManager.Upstream.RemoveProxy(proxyID)
}

// UpdateUpstreamProxy 更新上游代理
func (a *App) UpdateUpstreamProxy(proxy *features.UpstreamProxy) error {
	return a.featureManager.Upstream.UpdateProxy(proxy)
}

// GetUpstreamProxies 获取所有上游代理
func (a *App) GetUpstreamProxies() []*features.UpstreamProxy {
	return a.featureManager.Upstream.GetAllProxies()
}

// TestUpstreamProxy 测试上游代理连接
func (a *App) TestUpstreamProxy(proxyID string) error {
	return a.featureManager.Upstream.TestUpstreamProxy(proxyID)
}

// Shutdown 应用关闭时的清理工作
func (a *App) Shutdown(ctx context.Context) {
	// 停止代理
	if a.isRunning {
		a.StopProxy()
	}

	// 保存配置
	if err := a.config.SaveConfig(); err != nil {
		logger.Error("Failed to save config: %v", err)
	}

	// 关闭日志
	logger.Close()
}

// ExportFlows 导出流量数据
func (a *App) ExportFlows(options export.ExportOptions) (*export.ExportResult, error) {
	logger.Info("Starting export with type: %s, scope: %s, flows count: %d",
		options.Type, options.Scope, len(options.Flows))

	result, zipData, err := a.exportService.ExportToZip(options)
	if err != nil {
		logger.Error("Export failed: %v", err)
		return nil, err
	}

	if result.Success && len(zipData) > 0 {
		// 使用Wails的文件保存对话框
		filename, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
			DefaultFilename: options.Filename,
			Title:           "保存导出文件",
			Filters: []runtime.FileFilter{
				{
					DisplayName: "ZIP文件",
					Pattern:     "*.zip",
				},
			},
		})

		if err != nil {
			logger.Error("Failed to show save dialog: %v", err)
			return nil, fmt.Errorf("文件保存对话框失败: %w", err)
		}

		if filename == "" {
			// 用户取消了保存
			result.Success = false
			result.Message = "用户取消了保存"
			return result, nil
		}

		// 写入文件
		if err := a.systemManager.WriteFile(filename, zipData); err != nil {
			logger.Error("Failed to write export file: %v", err)
			return nil, fmt.Errorf("文件写入失败: %w", err)
		}

		result.Filename = filename
		logger.Info("Export completed successfully: %s", filename)
	}

	return result, nil
}

// DecryptRequestBody 解密请求体 (保留用于导出功能)
func (a *App) DecryptRequestBody(body []byte, headers map[string]string) ([]byte, error) {
	return a.exportService.DecryptBody(body, headers)
}

// DecryptResponseBody 解密响应体 (保留用于导出功能)
func (a *App) DecryptResponseBody(body []byte, headers map[string]string) ([]byte, error) {
	return a.exportService.DecryptBody(body, headers)
}

// GetResponseHexView 获取响应体的16进制视图
func (a *App) GetResponseHexView(flowID string) (string, error) {
	flow, exists := a.proxyServer.GetFlow(flowID)
	if !exists {
		return "", fmt.Errorf("flow not found: %s", flowID)
	}

	if flow.Response == nil {
		return "", fmt.Errorf("no response data")
	}

	return flow.Response.HexView, nil
}
