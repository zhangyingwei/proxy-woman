# ProxyWoman 开发文档

## 项目概述

ProxyWoman 是一个现代化的网络调试代理工具，采用 Go + Wails + Svelte 技术栈构建。本文档记录了项目的开发过程和技术细节。

## 已完成的功能

### 后端 (Go)

#### 1. 证书管理模块 (`internal/certmanager`)
- ✅ 自动生成根 CA 证书
- ✅ 动态生成服务器证书
- ✅ 证书缓存机制
- ✅ 证书文件管理

#### 2. 代理核心模块 (`internal/proxycore`)
- ✅ HTTP/HTTPS 代理服务器
- ✅ MITM (中间人攻击) 支持
- ✅ 流量拦截和解析
- ✅ Flow 数据结构定义
- ✅ 实时流量处理

#### 3. 系统集成模块 (`internal/system`)
- ✅ macOS 系统代理设置
- ✅ 网络服务管理
- ✅ 配置目录管理

#### 4. 配置管理模块 (`internal/config`)
- ✅ JSON 配置文件支持
- ✅ 默认配置管理
- ✅ 配置加载和保存

#### 5. 日志模块 (`internal/logger`)
- ✅ 多级别日志记录
- ✅ 文件日志输出
- ✅ 日志轮转支持

#### 6. 功能框架 (`internal/features`)
- ✅ 基础框架搭建
- 🔄 为未来高级功能预留空间

### 前端 (Svelte + TypeScript)

#### 1. 状态管理 (`src/stores`)
- ✅ 流量数据存储 (`flowStore.ts`)
- ✅ 选择状态管理 (`selectionStore.ts`)
- ✅ 代理状态管理 (`proxyStore.ts`)
- ✅ 过滤功能支持

#### 2. 服务层 (`src/services`)
- ✅ 事件服务 (`EventService.ts`)
- ✅ 代理服务 (`ProxyService.ts`)
- ✅ Go 后端通信

#### 3. UI 组件 (`src/components`)
- ✅ 工具栏组件 (`Toolbar.svelte`)
- ✅ 流量表格组件 (`FlowTable.svelte`)
- ✅ 详情视图组件 (`DetailView.svelte`)
- ✅ 过滤栏组件 (`FilterBar.svelte`)

#### 4. 主应用
- ✅ 三栏式布局
- ✅ 深色主题
- ✅ 响应式设计

### 集成和配置

#### 1. Wails 集成
- ✅ Go 后端与前端通信
- ✅ 事件系统集成
- ✅ 方法绑定
- ✅ 应用生命周期管理

#### 2. 构建系统
- ✅ 开发脚本 (`scripts/dev.sh`)
- ✅ 构建脚本 (`scripts/build.sh`)
- ✅ Go 模块管理
- ✅ 前端依赖管理

#### 3. 测试
- ✅ 基础单元测试
- ✅ 代理核心功能测试

## 技术架构

### 数据流
```
用户浏览器 → 系统代理 → ProxyWoman代理服务器 → 目标服务器
                              ↓
                         流量解析和存储
                              ↓
                         Wails事件系统
                              ↓
                         Svelte前端更新
```

### 模块依赖关系
```
main.go
├── app.go (主应用)
│   ├── config (配置管理)
│   ├── logger (日志记录)
│   ├── certmanager (证书管理)
│   ├── proxycore (代理核心)
│   ├── system (系统集成)
│   └── features (功能模块)
└── frontend (Svelte前端)
    ├── stores (状态管理)
    ├── services (服务层)
    └── components (UI组件)
```

## 开发环境

### 必需工具
- Go 1.23+
- Wails CLI v2.10.1+
- Node.js 16+ (可选，用于前端开发)

### 快速启动
```bash
# 开发模式
./scripts/dev.sh

# 或手动启动
wails dev
```

### 构建
```bash
# 构建 macOS 版本
./scripts/build.sh darwin

# 构建其他平台
./scripts/build.sh windows
./scripts/build.sh linux
```

## 已完成的完整功能

### 阶段 1: 基础架构 ✅
- [x] 证书管理模块
- [x] 代理核心引擎
- [x] 系统集成
- [x] 基础前端界面

### 阶段 2: UI 优化和核心功能 ✅
- [x] 流量过滤增强
- [x] 请求钉住 (Pinning) 功能
- [x] JSON 树状视图优化
- [x] 侧边栏域名分组

### 阶段 3: 高级调试工具 ✅
- [x] 断点功能
- [x] Map Local (本地映射)
- [x] 请求重放和编辑
- [x] 请求构造器

### 阶段 4: 可扩展性功能 ✅
- [x] JavaScript 脚本引擎
- [x] 允许/阻止列表
- [x] HAR 格式导入/导出
- [x] 设置窗口

### 阶段 5: 最终完善 ✅
- [x] CLI 命令行界面
- [x] 反向代理支持
- [x] 上游代理配置
- [x] 完整文档

## 已知问题

1. **Node.js 依赖**: 前端开发需要 Node.js，但运行时不需要
2. **证书信任**: 用户需要手动安装和信任 CA 证书
3. **平台支持**: 系统代理设置目前仅支持 macOS

## 贡献指南

1. Fork 项目
2. 创建功能分支
3. 提交更改
4. 创建 Pull Request

## 许可证

MIT License
