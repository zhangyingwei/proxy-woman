#!/bin/bash

# ProxyWoman 调试启动脚本

echo "🐛 启动 ProxyWoman 调试模式..."

# 设置环境变量
export CGO_ENABLED=1
export WAILS_DEBUG=1

# 检查是否安装了必要的工具
if ! command -v wails &> /dev/null; then
    echo "❌ Wails CLI 未安装，正在安装..."
    go install github.com/wailsapp/wails/v2/cmd/wails@latest
fi

# 清理之前的构建
echo "🧹 清理之前的构建..."
rm -rf build/bin/*

# 启动调试模式
echo "🔍 启动调试模式..."
wails dev -loglevel Debug -v 2

echo "✅ 调试会话结束"
