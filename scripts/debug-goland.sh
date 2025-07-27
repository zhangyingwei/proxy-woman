#!/bin/bash

# GoLand 调试脚本
# 设置必要的环境变量来解决链接问题

echo "🐛 Setting up debug environment for GoLand..."

# 设置环境变量
export CGO_ENABLED=1
export WAILS_DEBUG=1
export CGO_LDFLAGS="-framework UniformTypeIdentifiers"
export GOOS=darwin
export GOARCH=arm64

echo "✅ Environment variables set:"
echo "   CGO_ENABLED=$CGO_ENABLED"
echo "   WAILS_DEBUG=$WAILS_DEBUG"
echo "   CGO_LDFLAGS=$CGO_LDFLAGS"
echo "   GOOS=$GOOS"
echo "   GOARCH=$GOARCH"

echo ""
echo "🚀 Now you can run debug in GoLand with these settings:"
echo "   1. Open Run/Debug Configurations"
echo "   2. Select 'Debug Wails App' configuration"
echo "   3. Make sure CGO_LDFLAGS is set to: -framework UniformTypeIdentifiers"
echo "   4. Click Debug button"
echo ""
echo "Or run directly:"
echo "   # For development (with debug symbols):"
echo "   CGO_ENABLED=1 CGO_LDFLAGS=\"-framework UniformTypeIdentifiers\" go run -tags dev -gcflags \"all=-N -l\" ."
echo "   # For production:"
echo "   CGO_ENABLED=1 CGO_LDFLAGS=\"-framework UniformTypeIdentifiers\" go run -tags desktop,production -ldflags \"-w -s\" ."
echo ""

# 可选：直接启动调试
if [ "$1" = "run" ]; then
    echo "🔍 Starting debug session..."
    CGO_ENABLED=1 CGO_LDFLAGS="-framework UniformTypeIdentifiers" go run -tags dev -gcflags "all=-N -l" .
fi
