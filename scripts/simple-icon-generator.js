#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

// 简化的图标生成器，使用Canvas API（如果可用）
// 或者提供SVG到PNG转换的指导

const svgPath = path.join(__dirname, '../build/icon.svg');
const buildDir = path.join(__dirname, '../build');

console.log('🎨 ProxyWoman Icon Generator');
console.log('============================\n');

// 检查SVG文件是否存在
if (!fs.existsSync(svgPath)) {
  console.error('❌ SVG icon file not found:', svgPath);
  process.exit(1);
}

console.log('✓ SVG icon found:', svgPath);

// 读取SVG内容
const svgContent = fs.readFileSync(svgPath, 'utf8');
console.log('✓ SVG content loaded');

// 创建一个简化的HTML文件用于预览
const previewHtml = `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>ProxyWoman Icon Preview</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            margin: 0;
            padding: 40px;
            min-height: 100vh;
            display: flex;
            flex-direction: column;
            align-items: center;
            color: white;
        }
        
        .container {
            background: rgba(255, 255, 255, 0.1);
            backdrop-filter: blur(10px);
            border-radius: 20px;
            padding: 40px;
            text-align: center;
            box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
            border: 1px solid rgba(255, 255, 255, 0.2);
        }
        
        h1 {
            margin: 0 0 30px 0;
            font-size: 2.5em;
            font-weight: 300;
            text-shadow: 0 2px 10px rgba(0, 0, 0, 0.3);
        }
        
        .icon-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
            gap: 30px;
            margin: 30px 0;
            max-width: 800px;
        }
        
        .icon-item {
            background: rgba(255, 255, 255, 0.1);
            border-radius: 15px;
            padding: 20px;
            border: 1px solid rgba(255, 255, 255, 0.2);
            transition: transform 0.3s ease, box-shadow 0.3s ease;
        }
        
        .icon-item:hover {
            transform: translateY(-5px);
            box-shadow: 0 15px 35px rgba(0, 0, 0, 0.3);
        }
        
        .icon-item svg {
            width: 100%;
            height: auto;
            max-width: 80px;
            filter: drop-shadow(0 4px 8px rgba(0, 0, 0, 0.3));
        }
        
        .icon-size {
            margin-top: 10px;
            font-size: 0.9em;
            opacity: 0.8;
        }
        
        .instructions {
            background: rgba(255, 255, 255, 0.1);
            border-radius: 15px;
            padding: 30px;
            margin-top: 40px;
            max-width: 600px;
            text-align: left;
            border: 1px solid rgba(255, 255, 255, 0.2);
        }
        
        .instructions h3 {
            margin-top: 0;
            color: #f093fb;
        }
        
        .instructions code {
            background: rgba(0, 0, 0, 0.3);
            padding: 2px 6px;
            border-radius: 4px;
            font-family: 'Monaco', 'Menlo', monospace;
        }
        
        .instructions ol {
            padding-left: 20px;
        }
        
        .instructions li {
            margin-bottom: 10px;
            line-height: 1.6;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🎨 ProxyWoman Icon</h1>
        
        <div class="icon-grid">
            <div class="icon-item">
                ${svgContent}
                <div class="icon-size">512x512</div>
            </div>
            <div class="icon-item">
                <div style="width: 64px; height: 64px; margin: 0 auto;">
                    ${svgContent}
                </div>
                <div class="icon-size">64x64</div>
            </div>
            <div class="icon-item">
                <div style="width: 32px; height: 32px; margin: 0 auto;">
                    ${svgContent}
                </div>
                <div class="icon-size">32x32</div>
            </div>
            <div class="icon-item">
                <div style="width: 16px; height: 16px; margin: 0 auto;">
                    ${svgContent}
                </div>
                <div class="icon-size">16x16</div>
            </div>
        </div>
    </div>
    
    <div class="instructions">
        <h3>📋 如何生成PNG图标</h3>
        <p>由于环境限制，请按以下步骤手动生成PNG图标：</p>
        <ol>
            <li>在浏览器中打开此HTML文件查看图标效果</li>
            <li>使用在线SVG转PNG工具（如 <code>convertio.co</code> 或 <code>cloudconvert.com</code>）</li>
            <li>上传 <code>build/icon.svg</code> 文件</li>
            <li>生成以下尺寸的PNG文件：
                <ul>
                    <li>16x16, 32x32, 48x48, 64x64</li>
                    <li>128x128, 256x256, 512x512, 1024x1024</li>
                </ul>
            </li>
            <li>将生成的PNG文件保存到 <code>build/</code> 目录</li>
            <li>将 <code>512x512</code> 的PNG重命名为 <code>appicon.png</code></li>
        </ol>
        
        <h3>🔧 或者使用命令行工具</h3>
        <p>如果你有ImageMagick或Inkscape，可以使用：</p>
        <code>convert icon.svg -resize 512x512 appicon.png</code><br>
        <code>inkscape icon.svg -w 512 -h 512 -o appicon.png</code>
    </div>
</body>
</html>
`;

// 保存预览HTML文件
const previewPath = path.join(buildDir, 'icon-preview.html');
fs.writeFileSync(previewPath, previewHtml);

console.log('✓ Icon preview created:', previewPath);
console.log('\n📖 Instructions:');
console.log('1. Open icon-preview.html in your browser to see the icon');
console.log('2. Use online tools or ImageMagick to convert SVG to PNG');
console.log('3. Generate PNG files in sizes: 16, 32, 48, 64, 128, 256, 512, 1024');
console.log('4. Save the 512x512 PNG as appicon.png');

// 创建一个简单的图标配置文件
const iconConfig = {
  name: "ProxyWoman",
  description: "A modern network proxy analysis tool with elegant design",
  source: "icon.svg",
  sizes: [16, 32, 48, 64, 128, 256, 512, 1024],
  formats: ["png", "ico"],
  platforms: {
    windows: {
      ico: [16, 24, 32, 48, 64, 128, 256],
      png: [16, 32, 48, 64, 128, 256, 512]
    },
    macos: {
      png: [16, 32, 64, 128, 256, 512, 1024],
      icns: true
    },
    linux: {
      png: [16, 32, 48, 64, 128, 256, 512],
      svg: true
    }
  },
  design: {
    theme: "Network proxy with feminine elegance",
    colors: {
      primary: "#667eea to #764ba2",
      accent: "#f093fb to #f5576c",
      nodes: "#4facfe to #00f2fe"
    },
    elements: [
      "Central proxy node",
      "Network connections",
      "Data flow animation",
      "Elegant curves",
      "Modern gradients"
    ]
  }
};

fs.writeFileSync(
  path.join(buildDir, 'icon-config.json'),
  JSON.stringify(iconConfig, null, 2)
);

console.log('✓ Icon configuration saved');
console.log('\n🎉 Icon generation setup complete!');
console.log(`📁 Files created in: ${buildDir}`);
console.log('   - icon.svg (source)');
console.log('   - icon-preview.html (preview)');
console.log('   - icon-config.json (configuration)');
