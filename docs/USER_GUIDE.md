# ProxyWoman 用户指南

## 目录

1. [快速开始](#快速开始)
2. [基础功能](#基础功能)
3. [高级功能](#高级功能)
4. [CLI 使用](#cli-使用)
5. [故障排除](#故障排除)

## 快速开始

### 安装和启动

1. **下载应用**
   - 从 GitHub Releases 下载适合您操作系统的版本
   - 或使用源码编译：`wails build`

2. **首次启动**
   ```bash
   # GUI 模式
   ./ProxyWoman

   # CLI 模式
   ./proxywoman start
   ```

3. **安装证书**
   - 启动后点击工具栏的"证书"按钮
   - 将 CA 证书安装到系统钥匙串并设置为信任
   - 这是拦截 HTTPS 流量的必要步骤

### 基本使用流程

1. 点击"启动代理"按钮
2. 应用会自动设置系统代理
3. 开始浏览网页，流量将显示在界面中
4. 点击任意请求查看详细信息

## 基础功能

### 流量监控

- **实时显示**: 所有 HTTP/HTTPS 请求实时显示在左侧列表
- **详细信息**: 点击请求查看完整的请求/响应详情
- **过滤搜索**: 使用底部搜索框过滤特定请求
- **钉住功能**: 点击📌图标收藏重要请求

### 界面布局

- **侧边栏**: 显示域名分组和收藏的请求
- **流量表格**: 显示所有拦截的网络请求
- **详情面板**: 显示选中请求的详细信息
- **工具栏**: 代理控制和状态显示

### 数据查看

- **JSON 树状视图**: 自动格式化 JSON 响应
- **HTML 预览**: 支持 HTML 内容预览
- **图片显示**: 直接显示图片响应
- **原始数据**: 查看原始请求/响应数据

## 高级功能

### Map Local (本地映射)

将网络请求映射到本地文件：

1. 打开设置 → Map Local 标签
2. 添加新规则：
   - **规则名称**: 便于识别的名称
   - **URL 模式**: 要匹配的 URL 模式
   - **本地路径**: 本地文件的完整路径
   - **Content-Type**: 可选，指定响应类型

**使用场景**:
- 前端开发时替换 JS/CSS 文件
- 测试不同的 API 响应
- 模拟服务器错误

### 断点调试

在请求或响应时暂停并允许修改：

1. 打开设置 → 断点标签
2. 添加断点规则：
   - **URL 模式**: 要断点的 URL
   - **方法**: HTTP 方法 (GET, POST 等)
   - **断点类型**: 请求、响应或两者

**使用场景**:
- 修改请求参数
- 模拟不同的响应
- 调试 API 交互

### JavaScript 脚本

使用 JavaScript 自动处理请求/响应：

1. 打开设置 → 脚本标签
2. 编写脚本：
   ```javascript
   // 修改请求头
   request.headers['X-Custom-Header'] = 'MyValue';
   
   // 修改响应
   if (response.statusCode === 404) {
     response.statusCode = 200;
     response.body = '{"error": false}';
   }
   
   // 记录日志
   console.log('Processing:', request.url);
   ```

**可用对象**:
- `request`: 请求对象 (method, url, headers, body)
- `response`: 响应对象 (statusCode, headers, body)
- `flow`: 完整的流量对象
- `console`: 日志输出

### 允许/阻止列表

控制哪些请求被允许或阻止：

1. 打开设置 → 允许/阻止标签
2. 选择模式：
   - **混合模式**: 阻止规则优先，其他允许
   - **白名单模式**: 只允许匹配的请求
   - **黑名单模式**: 阻止匹配的请求

### 请求重放

重新发送已捕获的请求：

1. 右键点击任意请求
2. 选择"重放请求"
3. 可选择修改请求参数后发送

### HAR 导入/导出

与其他工具交换数据：

```bash
# 导出当前流量
proxywoman export traffic.har

# 导入 HAR 文件
proxywoman import traffic.har
```

## CLI 使用

### 基本命令

```bash
# 启动代理
proxywoman start --port=8080

# 停止代理
proxywoman stop

# 查看状态
proxywoman status

# 显示帮助
proxywoman help
```

### 配置管理

```bash
# 查看配置
proxywoman config show

# 修改配置
proxywoman config set port 9090
proxywoman config set autostart true
```

### 证书管理

```bash
# 显示证书路径
proxywoman cert path

# 重新生成证书
proxywoman cert regenerate
```

### 规则管理

```bash
# 列出所有规则
proxywoman rules list

# 导出/导入
proxywoman export flows.har
proxywoman import flows.har
```

## 故障排除

### 常见问题

**Q: HTTPS 网站无法访问**
A: 确保已安装并信任 CA 证书

**Q: 系统代理设置失败**
A: 在 macOS 上需要管理员权限，手动设置代理到 127.0.0.1:8080

**Q: 某些应用无法抓包**
A: 某些应用可能绕过系统代理，需要单独配置

**Q: 性能问题**
A: 
- 使用过滤器减少显示的请求数量
- 定期清空流量记录
- 禁用不需要的脚本和规则

### 日志查看

日志文件位置：
- macOS: `~/.proxywoman/logs/`
- Windows: `%USERPROFILE%\.proxywoman\logs\`
- Linux: `~/.proxywoman/logs/`

### 配置文件

配置文件位置：
- macOS: `~/.proxywoman/config.json`
- Windows: `%USERPROFILE%\.proxywoman\config.json`
- Linux: `~/.proxywoman/config.json`

### 重置设置

```bash
# 删除配置目录重置所有设置
rm -rf ~/.proxywoman
```

## 高级配置

### 自定义端口

```json
{
  "proxyPort": 8080,
  "autoStart": false,
  "theme": "dark",
  "logLevel": "info"
}
```

### 性能优化

- 启用请求过滤减少内存使用
- 定期清理流量记录
- 合理使用脚本功能

### 安全注意事项

- CA 证书具有完全的 HTTPS 拦截能力
- 不要在不信任的网络环境中使用
- 定期更新和重新生成证书
- 谨慎分享包含敏感信息的 HAR 文件

## 技术支持

- GitHub Issues: [项目地址]
- 文档: [文档地址]
- 社区: [社区地址]
