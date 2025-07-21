# ProxyWoman API 文档

## 概述

ProxyWoman 提供了完整的 Go API 和 Wails 绑定，支持所有核心功能的编程访问。

## 核心模块

### 代理服务器 (ProxyServer)

```go
import "ProxyWoman/internal/proxycore"

// 创建代理服务器
server := proxycore.NewProxyServer(8080, certManager)

// 添加拦截器
server.AddRequestInterceptor(interceptor)
server.AddResponseInterceptor(interceptor)

// 启动/停止
server.Start()
server.Stop()

// 流量管理
flows := server.GetFlows()
server.ClearFlows()
server.PinFlow(flowID)
```

### 证书管理 (CertManager)

```go
import "ProxyWoman/internal/certmanager"

// 创建证书管理器
certManager := certmanager.NewCertManager(configDir)

// 初始化 CA
err := certManager.InitCA()

// 生成服务器证书
cert, err := certManager.GenerateServerCert(hostname)

// 获取 CA 证书路径
path := certManager.GetCACertPath()
```

### 功能管理 (FeatureManager)

```go
import "ProxyWoman/internal/features"

// 创建功能管理器
fm := features.NewFeatureManager()

// Map Local
rule := &features.MapLocalRule{
    ID: "rule1",
    Name: "Test Rule",
    URLPattern: "example.com",
    LocalPath: "/path/to/file",
    Enabled: true,
}
fm.MapLocal.AddRule(rule)

// 断点
breakRule := &features.BreakpointRule{
    ID: "break1",
    Name: "API Break",
    URLPattern: "/api/",
    BreakOnRequest: true,
    Enabled: true,
}
fm.Breakpoint.AddRule(breakRule)

// 脚本
script := &features.Script{
    ID: "script1",
    Name: "Header Modifier",
    Content: "request.headers['X-Test'] = 'value';",
    Type: "request",
    Enabled: true,
}
fm.Scripting.AddScript(script)
```

## Wails 绑定方法

### 代理控制

```typescript
// 启动代理
await StartProxy()

// 停止代理
await StopProxy()

// 检查状态
const isRunning = await IsProxyRunning()

// 获取端口
const port = await GetProxyPort()
```

### 流量管理

```typescript
// 获取所有流量
const flows = await GetFlows()

// 清空流量
await ClearFlows()

// 钉住流量
await PinFlow(flowId)

// 获取钉住的流量
const pinnedFlows = await GetPinnedFlows()

// 根据ID获取流量
const flow = await GetFlowByID(flowId)
```

### Map Local

```typescript
// 添加规则
await AddMapLocalRule({
  id: "rule1",
  name: "Test Rule",
  urlPattern: "example.com",
  localPath: "/path/to/file",
  enabled: true
})

// 获取所有规则
const rules = await GetMapLocalRules()

// 更新规则
await UpdateMapLocalRule(rule)

// 删除规则
await RemoveMapLocalRule(ruleId)
```

### 断点管理

```typescript
// 添加断点规则
await AddBreakpointRule({
  id: "break1",
  name: "API Break",
  urlPattern: "/api/",
  breakOnRequest: true,
  enabled: true
})

// 获取断点规则
const rules = await GetBreakpointRules()

// 恢复断点
await ResumeBreakpoint(sessionId)

// 取消断点
await CancelBreakpoint(sessionId)

// 获取活跃断点
const active = await GetActiveBreakpoints()
```

### 脚本管理

```typescript
// 添加脚本
await AddScript({
  id: "script1",
  name: "Header Modifier",
  content: "request.headers['X-Test'] = 'value';",
  type: "request",
  enabled: true
})

// 获取所有脚本
const scripts = await GetAllScripts()

// 验证脚本
await ValidateScript(scriptContent)

// 更新脚本
await UpdateScript(script)

// 删除脚本
await RemoveScript(scriptId)
```

### 允许/阻止列表

```typescript
// 添加规则
await AddAllowBlockRule({
  id: "rule1",
  name: "Block Ads",
  urlPattern: "ads.example.com",
  type: "block",
  enabled: true
})

// 设置模式
await SetAllowBlockMode("blacklist")

// 获取模式
const mode = await GetAllowBlockMode()

// 获取规则
const rules = await GetAllowBlockRules()
```

### 请求重放

```typescript
// 重放流量
const response = await ReplayFlow(flowId)

// 发送自定义请求
const response = await SendCustomRequest({
  method: "GET",
  url: "https://api.example.com",
  headers: {"Authorization": "Bearer token"},
  body: ""
})

// 修改并重放
const response = await ModifyAndReplayFlow(flowId, {
  method: "POST",
  headers: {"Content-Type": "application/json"}
})
```

### HAR 导入/导出

```typescript
// 导出到 HAR
await ExportFlowsToHAR("/path/to/export.har")

// 从 HAR 导入
const flows = await ImportHARToFlows("/path/to/import.har")
```

### 反向代理

```typescript
// 添加反向代理规则
await AddReverseProxyRule({
  id: "reverse1",
  name: "API Proxy",
  listenPath: "/api/",
  targetUrl: "http://localhost:3000",
  enabled: true
})

// 获取规则
const rules = await GetReverseProxyRules()
```

### 上游代理

```typescript
// 添加上游代理
await AddUpstreamProxy({
  id: "upstream1",
  name: "Corporate Proxy",
  proxyUrl: "http://proxy.company.com:8080",
  urlPattern: "*.company.com",
  enabled: true
})

// 测试连接
await TestUpstreamProxy(proxyId)

// 获取代理列表
const proxies = await GetUpstreamProxies()
```

## 事件系统

### 监听事件

```typescript
import { EventsOn } from '@wailsapp/runtime'

// 监听新流量
EventsOn('new-flow', (flow) => {
  console.log('New flow:', flow)
})

// 监听断点事件
EventsOn('breakpoint-hit', (session) => {
  console.log('Breakpoint hit:', session)
})
```

### 自定义拦截器

```go
type CustomInterceptor struct{}

func (ci *CustomInterceptor) InterceptRequest(flow *proxycore.Flow, w http.ResponseWriter, r *http.Request) (bool, error) {
    // 自定义请求处理逻辑
    if strings.Contains(flow.URL, "example.com") {
        // 修改请求
        r.Header.Set("X-Custom", "value")
    }
    return false, nil // 继续处理
}

func (ci *CustomInterceptor) InterceptResponse(flow *proxycore.Flow, resp *http.Response) (*http.Response, error) {
    // 自定义响应处理逻辑
    return resp, nil
}

// 添加到代理服务器
server.AddRequestInterceptor(&CustomInterceptor{})
```

## 配置 API

```go
import "ProxyWoman/internal/config"

// 加载配置
cfg, err := config.LoadConfig(configDir)

// 修改配置
cfg.ProxyPort = 9090
cfg.AutoStart = true

// 保存配置
err = cfg.SaveConfig()
```

## 日志 API

```go
import "ProxyWoman/internal/logger"

// 初始化日志
logger.InitLogger(configDir, "info")

// 记录日志
logger.Info("Proxy started on port %d", port)
logger.Error("Failed to start: %v", err)
logger.Debug("Debug information")
```

## 错误处理

所有 API 方法都返回适当的错误信息：

```typescript
try {
  await StartProxy()
} catch (error) {
  console.error('Failed to start proxy:', error)
}
```

```go
if err := server.Start(); err != nil {
    log.Printf("Failed to start server: %v", err)
    return err
}
```

## 类型定义

### Flow 结构

```typescript
interface Flow {
  id: string
  url: string
  method: string
  statusCode: number
  domain: string
  path: string
  scheme: string
  startTime: string
  endTime: string
  duration: number
  requestSize: number
  responseSize: number
  request: FlowRequest
  response: FlowResponse
  isPinned: boolean
  isBlocked: boolean
  contentType: string
  tags: string[]
}
```

### 规则结构

```typescript
interface MapLocalRule {
  id: string
  name: string
  urlPattern: string
  localPath: string
  contentType: string
  enabled: boolean
  isRegex: boolean
}

interface BreakpointRule {
  id: string
  name: string
  urlPattern: string
  method: string
  enabled: boolean
  isRegex: boolean
  breakOnRequest: boolean
  breakOnResponse: boolean
}
```

## 最佳实践

1. **错误处理**: 始终处理 API 调用的错误
2. **资源清理**: 适当清理流量记录和会话
3. **性能考虑**: 避免频繁的 API 调用
4. **并发安全**: 所有 API 都是线程安全的
5. **配置持久化**: 重要配置应保存到文件

## 扩展开发

创建自定义功能的步骤：

1. 实现拦截器接口
2. 注册到代理服务器
3. 添加 Wails 绑定方法
4. 创建前端界面

示例插件结构：

```go
type MyPlugin struct {
    // 插件状态
}

func (p *MyPlugin) InterceptRequest(flow *proxycore.Flow, w http.ResponseWriter, r *http.Request) (bool, error) {
    // 插件逻辑
    return false, nil
}

// 在 App 中注册
app.proxyServer.AddRequestInterceptor(&MyPlugin{})
```
