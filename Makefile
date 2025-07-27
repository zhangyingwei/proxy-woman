# ProxyWoman Makefile

.PHONY: dev debug build clean test

# 设置环境变量
export CGO_ENABLED=1
export WAILS_DEBUG=1
export CGO_LDFLAGS=-framework UniformTypeIdentifiers

# 开发模式
dev:
	@echo "🚀 Starting development mode..."
	wails dev -loglevel Debug -v 2

# 开发模式（带浏览器）
dev-browser:
	@echo "🚀 Starting development mode in browser..."
	wails dev -browser -loglevel Debug -v 2

# 调试模式 - 直接运行开发版本
debug:
	@echo "🐛 Starting debug mode..."
	CGO_ENABLED=1 CGO_LDFLAGS="-framework UniformTypeIdentifiers" go run -tags dev -gcflags "all=-N -l" .

# 调试开发版本（使用 wails dev）
debug-dev:
	@echo "🐛 Starting debug dev mode with wails dev..."
	wails dev -loglevel Debug -v 2

# 构建调试版本
build-debug:
	@echo "🔨 Building debug version..."
	wails build -ldflags "-w -s" -tags desktop

# 构建生产版本
build:
	@echo "🔨 Building production version..."
	wails build

# 清理构建文件
clean:
	@echo "🧹 Cleaning build files..."
	rm -rf build/bin/*
	rm -rf frontend/dist/*

# 运行测试
test:
	@echo "🧪 Running tests..."
	go test ./...

# 检查环境
doctor:
	@echo "🩺 Checking environment..."
	wails doctor

# 安装依赖
deps:
	@echo "📦 Installing dependencies..."
	go mod tidy
	cd frontend && npm install

# 生成绑定
generate:
	@echo "🔄 Generating bindings..."
	wails generate module

# 帮助信息
help:
	@echo "Available commands:"
	@echo "  dev         - Start development server"
	@echo "  debug       - Run in debug mode"
	@echo "  debug-dev   - Run dev version in debug mode"
	@echo "  build-debug - Build debug version"
	@echo "  build       - Build production version"
	@echo "  clean       - Clean build files"
	@echo "  test        - Run tests"
	@echo "  doctor      - Check environment"
	@echo "  deps        - Install dependencies"
	@echo "  generate    - Generate bindings"
