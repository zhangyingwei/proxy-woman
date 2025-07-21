#!/bin/bash

# ProxyWoman 构建脚本

echo "🔨 构建 ProxyWoman..."

# 检查平台参数
PLATFORM=${1:-darwin}

case $PLATFORM in
    darwin)
        echo "🍎 构建 macOS 版本..."
        ;;
    windows)
        echo "🪟 构建 Windows 版本..."
        ;;
    linux)
        echo "🐧 构建 Linux 版本..."
        ;;
    *)
        echo "❌ 不支持的平台: $PLATFORM"
        echo "支持的平台: darwin, windows, linux"
        exit 1
        ;;
esac

# 清理之前的构建
echo "🧹 清理之前的构建..."
rm -rf build/bin

# 确保依赖是最新的
echo "📦 更新依赖..."
go mod tidy

# 构建应用
echo "⚙️ 开始构建..."
wails build -platform $PLATFORM

if [ $? -eq 0 ]; then
    echo "✅ 构建成功！"
    echo "📁 构建文件位置: build/bin/"
    ls -la build/bin/
else
    echo "❌ 构建失败！"
    exit 1
fi
