#!/bin/bash

# GoLand 专用调试脚本
# 解决 CGO 环境变量和构建标签问题

echo "🐛 Starting GoLand debug session..."

# 设置必要的环境变量
export CGO_ENABLED=1
export CGO_LDFLAGS="-framework UniformTypeIdentifiers"
export WAILS_DEBUG=1

# 检查是否在正确的目录
if [ ! -f "go.mod" ]; then
    echo "❌ Error: go.mod not found. Please run this script from the project root directory."
    exit 1
fi

# 检查是否有 main_dev.go
if [ ! -f "main_dev.go" ]; then
    echo "❌ Error: main_dev.go not found."
    exit 1
fi

echo "✅ Environment setup complete"
echo "   CGO_ENABLED=$CGO_ENABLED"
echo "   CGO_LDFLAGS=$CGO_LDFLAGS"
echo "   WAILS_DEBUG=$WAILS_DEBUG"
echo ""

# 构建并运行应用（开发模式，带调试符号）
echo "🔨 Building and running application..."
go run -tags dev -gcflags "all=-N -l" .

echo "🏁 Debug session ended"
