#!/bin/bash

# ProxyWoman HTTPS 流量捕获测试脚本

echo "🔍 ProxyWoman HTTPS 流量捕获测试"
echo "=================================="

# 检查ProxyWoman是否在运行
if ! pgrep -f "ProxyWoman" > /dev/null; then
    echo "❌ ProxyWoman 未运行，请先启动应用"
    exit 1
fi

echo "✅ ProxyWoman 正在运行"

# 设置代理
export http_proxy="http://127.0.0.1:8080"
export https_proxy="http://127.0.0.1:8080"

echo "🔧 已设置代理: $https_proxy"

# 测试HTTPS请求
echo ""
echo "📡 测试HTTPS请求..."

# 测试1: Google
echo "测试 1: Google"
curl -s -k --connect-timeout 10 https://www.google.com > /dev/null
if [ $? -eq 0 ]; then
    echo "✅ Google 请求成功"
else
    echo "❌ Google 请求失败"
fi

# 测试2: GitHub
echo "测试 2: GitHub API"
curl -s -k --connect-timeout 10 https://api.github.com > /dev/null
if [ $? -eq 0 ]; then
    echo "✅ GitHub API 请求成功"
else
    echo "❌ GitHub API 请求失败"
fi

# 测试3: HTTPBin
echo "测试 3: HTTPBin"
curl -s -k --connect-timeout 10 https://httpbin.org/get > /dev/null
if [ $? -eq 0 ]; then
    echo "✅ HTTPBin 请求成功"
else
    echo "❌ HTTPBin 请求失败"
fi

# 测试4: JSON API
echo "测试 4: JSON API"
curl -s -k --connect-timeout 10 https://jsonplaceholder.typicode.com/posts/1 > /dev/null
if [ $? -eq 0 ]; then
    echo "✅ JSON API 请求成功"
else
    echo "❌ JSON API 请求失败"
fi

echo ""
echo "🔍 检查ProxyWoman日志..."
echo "请查看ProxyWoman控制台输出，应该看到类似以下的日志："
echo "  - TLS handshake successful for www.google.com"
echo "  - Received HTTPS request: GET / from www.google.com"
echo "  - TLS handshake successful for api.github.com"
echo "  - Received HTTPS request: GET / from api.github.com"
echo ""
echo "📱 请检查ProxyWoman应用界面，确认以上域名的流量记录已显示"

# 清理代理设置
unset http_proxy
unset https_proxy

echo ""
echo "✅ 测试完成！代理设置已清理"
echo ""
echo "🔧 故障排除提示："
echo "1. 如果看到TLS握手成功但应用中无流量记录，可能是handleHTTPS方法的问题"
echo "2. 如果完全没有TLS握手日志，检查证书是否正确安装"
echo "3. 如果请求失败，检查代理端口是否正确(默认8080)"
echo "4. 确保已安装并信任ProxyWoman的CA证书"
