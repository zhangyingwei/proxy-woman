#!/bin/bash

# ProxyWoman 开发环境启动脚本

echo "🚀 启动 ProxyWoman 开发环境..."

# 检查是否安装了必要的工具
if ! command -v go &> /dev/null; then
    echo "❌ Go 未安装，请先安装 Go 1.23+"
    exit 1
fi

if ! command -v wails &> /dev/null; then
    echo "❌ Wails CLI 未安装，正在安装..."
    go install github.com/wailsapp/wails/v2/cmd/wails@latest
fi

# 检查 Go 模块依赖
echo "📦 检查 Go 依赖..."
go mod tidy

# 检查前端依赖
if [ -d "frontend/node_modules" ]; then
    echo "✅ 前端依赖已安装"
else
    echo "📦 安装前端依赖需要 Node.js，请手动运行: cd frontend && npm install"
fi

# 启动开发服务器
echo "🔥 启动开发服务器..."
wails dev
