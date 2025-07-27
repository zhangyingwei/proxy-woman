#!/bin/bash

# ProxyWoman 提交准备脚本
# 用于整理和提交当前的修改内容

echo "🚀 ProxyWoman 提交准备脚本"
echo "================================"

# 检查当前分支
current_branch=$(git branch --show-current)
echo "📍 当前分支: $current_branch"

# 显示修改统计
echo ""
echo "📊 修改统计:"
git diff --stat HEAD

echo ""
echo "📁 文件状态:"
git status --porcelain

echo ""
echo "🔍 主要修改内容:"
echo "1. 响应体解码架构重构 - 后端自动解码，前端即时显示"
echo "2. 应用图标设计 - 现代化网络主题图标，支持多平台"
echo "3. 布局恢复优化 - 左右分栏布局，保留解码功能"
echo "4. 脚本日志查看 - 完整的脚本执行记录和日志"
echo "5. 16进制视图 - Chrome风格的hex dump显示"

echo ""
echo "📋 提交建议:"
echo "git add ."
echo "git commit -m \"feat: 响应体解码优化、应用图标设计、布局恢复\""
echo ""
echo "或者分别提交:"
echo "git add internal/proxycore/decoder.go internal/proxycore/flow.go app.go"
echo "git commit -m \"feat: 实现后端响应体自动解码和16进制视图\""
echo ""
echo "git add build/icon* scripts/*icon* wails.json"
echo "git commit -m \"feat: 设计专业应用图标，支持多平台格式\""
echo ""
echo "git add frontend/src/components/DetailViewSimplified.svelte frontend/src/App.svelte"
echo "git commit -m \"feat: 恢复左右分栏布局，优化用户体验\""

echo ""
read -p "是否要查看详细的修改差异? (y/n): " show_diff
if [ "$show_diff" = "y" ] || [ "$show_diff" = "Y" ]; then
    echo ""
    echo "📝 详细修改差异:"
    git diff HEAD --name-only | while read file; do
        echo ""
        echo "=== $file ==="
        git diff HEAD -- "$file" | head -20
        echo "..."
    done
fi

echo ""
read -p "是否要执行 git add . ? (y/n): " do_add
if [ "$do_add" = "y" ] || [ "$do_add" = "Y" ]; then
    git add .
    echo "✅ 已执行 git add ."
    
    echo ""
    echo "📋 建议的提交信息:"
    echo "feat: 响应体解码优化、应用图标设计、布局恢复"
    echo ""
    echo "主要改进:"
    echo "- 🔧 后端自动解码: 将解码操作前移到代理端，提升性能"
    echo "- 🔢 16进制视图: Chrome风格的hex dump显示"
    echo "- 🎨 应用图标: 现代化网络主题图标设计"
    echo "- 📱 布局恢复: 左右分栏布局，更好的空间利用"
    echo "- 📜 脚本日志: 完整的脚本执行记录查看"
    echo "- 🎯 用户体验: 简化操作，自动处理，即时显示"
    echo ""
    
    read -p "是否要执行提交? (y/n): " do_commit
    if [ "$do_commit" = "y" ] || [ "$do_commit" = "Y" ]; then
        git commit -m "feat: 响应体解码优化、应用图标设计、布局恢复

主要改进:
- 🔧 后端自动解码: 将解码操作前移到代理端，提升性能
- 🔢 16进制视图: Chrome风格的hex dump显示  
- 🎨 应用图标: 现代化网络主题图标设计
- 📱 布局恢复: 左右分栏布局，更好的空间利用
- 📜 脚本日志: 完整的脚本执行记录查看
- 🎯 用户体验: 简化操作，自动处理，即时显示

技术改进:
- 新增ResponseDecoder解码器，支持Gzip等压缩格式
- 扩展FlowResponse数据结构，增加解码字段
- 创建DetailViewSimplified组件，恢复左右分栏布局
- 设计SVG矢量图标，生成多平台图标文件
- 集成脚本执行日志，支持右键菜单查看
- 优化请求类型检测，提升过滤精度"
        
        echo "✅ 提交完成!"
        echo ""
        echo "📋 提交信息:"
        git log --oneline -1
    fi
fi

echo ""
echo "🎉 脚本执行完成!"
