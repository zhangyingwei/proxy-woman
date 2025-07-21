# ProxyWoman

一个现代化的跨平台网络调试代理工具，使用 Go + Wails + Svelte 构建。

## 功能特性

- 🔍 **HTTP/HTTPS 流量拦截** - 支持完整的 MITM (中间人攻击) 代理
- 📊 **实时流量监控** - 实时显示网络请求和响应
- 🔒 **自动证书管理** - 自动生成和管理 CA 证书
- 🎨 **现代化界面** - 基于 Svelte 的响应式用户界面
- ⚡ **高性能** - Go 语言后端，高并发处理能力
- 🖥️ **跨平台** - 支持 macOS、Windows、Linux

## 项目结构

```
ProxyWoman/
├── internal/                 # Go 后端模块
│   ├── certmanager/         # 证书管理
│   ├── proxycore/           # 代理核心引擎
│   ├── system/              # 系统集成
│   └── features/            # 高级功能 (未来扩展)
├── frontend/                # Svelte 前端
│   ├── src/
│   │   ├── components/      # UI 组件
│   │   ├── stores/          # 状态管理
│   │   └── services/        # 服务层
│   └── ...
├── app.go                   # Wails 应用主文件
└── main.go                  # 程序入口
```

## 开发环境要求

- Go 1.23+
- Node.js 16+ (用于前端开发)
- Wails CLI v2.10.1+

## 快速开始

### 1. 安装依赖

```bash
# 安装 Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 进入项目目录
cd ProxyWoman

# 安装 Go 依赖
go mod tidy

# 安装前端依赖 (需要 Node.js)
cd frontend && npm install
```

### 2. 开发模式运行

```bash
# 在项目根目录运行
wails dev
```

### 3. 构建生产版本

```bash
# 构建 macOS 应用
wails build -platform darwin

# 构建 Windows 应用
wails build -platform windows

# 构建 Linux 应用
wails build -platform linux
```

## 使用说明

### 启动代理

1. 启动应用后，点击工具栏中的"启动代理"按钮
2. 应用会自动设置系统代理到本地端口 8080
3. 开始浏览网页，流量将自动显示在界面中

### 证书安装

1. 点击工具栏中的"证书"按钮查看 CA 证书路径
2. 将证书安装到系统钥匙串中并设置为信任
3. 这样才能正确拦截 HTTPS 流量

### 流量查看

- 左侧面板显示所有拦截的请求列表
- 点击任意请求查看详细的请求/响应信息
- 支持按 URL、方法、域名等进行过滤

## 开发计划

这是项目的 MVP (最小可行产品) 版本，包含了基础的代理功能。未来计划添加：

- [ ] 断点调试功能
- [ ] Map Local (本地映射)
- [ ] 请求重放和编辑
- [ ] JavaScript 脚本支持
- [ ] HAR 格式导入/导出
- [ ] 更多过滤和搜索选项

## 技术栈

- **后端**: Go 1.23
- **前端**: Svelte + TypeScript + Vite
- **桌面框架**: Wails v2
- **代理引擎**: 基于 Go net/http 的自定义实现
- **证书管理**: Go crypto/x509
- **脚本引擎**: goja (计划中)

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License
