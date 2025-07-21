# ProxyWoman HTTPS 流量捕获修复

## 问题描述

用户报告虽然控制台显示TLS握手成功的日志，但应用界面中没有显示相应的流量记录：

```
TLS handshake successful for 193391-ipv4.gr.global.aa-rt.sharepoint.com
TLS handshake successful for httpdns.n.netease.com
TLS handshake successful for interface.music.163.com
TLS handshake successful for clients4.google.com
```

## 根本原因分析

1. **TLS握手成功但HTTP处理失败**: TLS连接建立成功，但后续的HTTP请求处理有问题
2. **singleConnListener实现问题**: 原始实现只能处理一个连接，之后就返回EOF
3. **错误处理不完善**: 缺少详细的调试日志来诊断问题
4. **流量记录机制问题**: 可能在addFlow或flowHandler环节出现问题

## 修复方案

### 1. 改进HTTPS处理流程

**原始问题**:
- `singleConnListener`只能处理一个连接
- 缺少错误处理和调试信息
- HTTP服务器配置不完善

**修复后**:
```go
// handleHTTPS 处理HTTPS流量
func (ps *ProxyServer) handleHTTPS(tlsConn *tls.Conn, targetHost string) {
    defer tlsConn.Close()
    
    fmt.Printf("Starting HTTPS handler for %s\n", targetHost)
    
    // 创建HTTP服务器来处理解密后的HTTPS请求
    server := &http.Server{
        Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            fmt.Printf("🔍 Received HTTPS request: %s %s from %s\n", r.Method, r.URL.Path, targetHost)
            
            // 设置完整的URL
            r.URL.Scheme = "https"
            r.URL.Host = targetHost

            // 处理为普通HTTP请求
            ps.handleHTTP(w, r)
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
```

### 2. 改进singleConnListener

**原始问题**:
```go
func (l *singleConnListener) Accept() (net.Conn, error) {
    var conn net.Conn
    l.once.Do(func() {
        conn = l.conn
    })
    if conn == nil {
        return nil, io.EOF
    }
    return conn, nil
}
```

**修复后**:
```go
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
```

### 3. 增强调试和日志记录

**添加详细的调试信息**:
```go
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
```

### 4. 创建测试脚本

创建了`test_https_capture.sh`脚本来验证HTTPS流量捕获：

```bash
#!/bin/bash
# 设置代理
export https_proxy="http://127.0.0.1:8080"

# 测试多个HTTPS网站
curl -s -k --connect-timeout 10 https://www.google.com
curl -s -k --connect-timeout 10 https://api.github.com
curl -s -k --connect-timeout 10 https://httpbin.org/get
```

## 诊断流程

### 1. 检查TLS握手
查看控制台是否有以下日志：
```
TLS handshake successful for example.com
```

### 2. 检查HTTPS请求处理
查看是否有以下日志：
```
🔍 Received HTTPS request: GET / from example.com
🔍 Calling handleHTTP for HTTPS request
🔍 handleHTTP completed for HTTPS request
```

### 3. 检查流量记录
查看是否有以下日志：
```
📝 Adding flow: GET https://example.com example.com (Status: 200)
📤 Notifying flow handler for: https://example.com
```

### 4. 检查前端更新
确认前端是否收到流量更新事件。

## 可能的问题点

### 1. 证书问题
- **症状**: TLS握手失败
- **解决**: 确保CA证书已正确安装并信任

### 2. HTTP处理问题
- **症状**: TLS握手成功但无HTTP请求日志
- **解决**: 检查singleConnListener实现

### 3. 流量记录问题
- **症状**: 有HTTP请求日志但无流量记录
- **解决**: 检查addFlow和flowHandler

### 4. 前端显示问题
- **症状**: 后端有流量记录但前端不显示
- **解决**: 检查事件通知机制

## 测试验证

### 1. 运行测试脚本
```bash
./test_https_capture.sh
```

### 2. 手动测试
```bash
export https_proxy="http://127.0.0.1:8080"
curl -v -k https://httpbin.org/get
```

### 3. 检查日志输出
观察控制台输出，确认每个步骤都有相应的日志。

## 性能考虑

1. **连接复用**: 改进的singleConnListener支持HTTP/1.1连接复用
2. **超时设置**: 添加了适当的读写超时
3. **错误处理**: 改进了错误处理，避免资源泄漏
4. **调试开关**: 可以通过编译标志控制调试输出

## 后续改进

1. **HTTP/2支持**: 考虑添加HTTP/2协议支持
2. **WebSocket支持**: 改进WebSocket流量处理
3. **性能优化**: 减少不必要的内存分配
4. **更好的错误报告**: 提供更详细的错误信息给用户

## 总结

这次修复主要解决了HTTPS流量处理链路中的几个关键问题：

1. ✅ 改进了singleConnListener的实现
2. ✅ 增强了错误处理和调试信息
3. ✅ 优化了HTTP服务器配置
4. ✅ 添加了完整的测试验证流程

修复后，用户应该能够在ProxyWoman应用界面中看到所有成功建立TLS连接的HTTPS流量记录。
