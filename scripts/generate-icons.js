#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

// 检查是否安装了sharp
let sharp;
try {
  sharp = require('sharp');
} catch (error) {
  console.error('Sharp not found. Installing...');
  const { execSync } = require('child_process');
  try {
    execSync('npm install sharp', { stdio: 'inherit' });
    sharp = require('sharp');
  } catch (installError) {
    console.error('Failed to install sharp:', installError.message);
    process.exit(1);
  }
}

const svgPath = path.join(__dirname, '../build/icon.svg');
const buildDir = path.join(__dirname, '../build');

// 确保build目录存在
if (!fs.existsSync(buildDir)) {
  fs.mkdirSync(buildDir, { recursive: true });
}

// 读取SVG文件
const svgBuffer = fs.readFileSync(svgPath);

// 定义需要生成的图标尺寸
const iconSizes = [
  { size: 16, name: 'icon-16.png' },
  { size: 32, name: 'icon-32.png' },
  { size: 48, name: 'icon-48.png' },
  { size: 64, name: 'icon-64.png' },
  { size: 128, name: 'icon-128.png' },
  { size: 256, name: 'icon-256.png' },
  { size: 512, name: 'icon-512.png' },
  { size: 1024, name: 'icon-1024.png' },
  // 特殊用途的图标
  { size: 512, name: 'appicon.png' }, // 替换现有的appicon.png
  { size: 256, name: 'icon.png' },    // 通用图标
];

// macOS特定的图标尺寸
const macOSIconSizes = [
  { size: 16, name: 'icon_16x16.png' },
  { size: 32, name: 'icon_16x16@2x.png' },
  { size: 32, name: 'icon_32x32.png' },
  { size: 64, name: 'icon_32x32@2x.png' },
  { size: 128, name: 'icon_128x128.png' },
  { size: 256, name: 'icon_128x128@2x.png' },
  { size: 256, name: 'icon_256x256.png' },
  { size: 512, name: 'icon_256x256@2x.png' },
  { size: 512, name: 'icon_512x512.png' },
  { size: 1024, name: 'icon_512x512@2x.png' },
];

// Windows特定的图标尺寸
const windowsIconSizes = [
  { size: 16, name: 'icon-16.ico.png' },
  { size: 24, name: 'icon-24.ico.png' },
  { size: 32, name: 'icon-32.ico.png' },
  { size: 48, name: 'icon-48.ico.png' },
  { size: 64, name: 'icon-64.ico.png' },
  { size: 128, name: 'icon-128.ico.png' },
  { size: 256, name: 'icon-256.ico.png' },
];

// 生成图标的函数
async function generateIcon(size, outputPath) {
  try {
    await sharp(svgBuffer)
      .resize(size, size, {
        fit: 'contain',
        background: { r: 0, g: 0, b: 0, alpha: 0 }
      })
      .png({
        quality: 100,
        compressionLevel: 9,
        adaptiveFiltering: true
      })
      .toFile(outputPath);
    
    console.log(`✓ Generated ${path.basename(outputPath)} (${size}x${size})`);
  } catch (error) {
    console.error(`✗ Failed to generate ${path.basename(outputPath)}:`, error.message);
  }
}

// 生成所有图标
async function generateAllIcons() {
  console.log('🎨 Generating ProxyWoman icons...\n');
  
  // 创建平台特定的目录
  const darwinDir = path.join(buildDir, 'darwin');
  const windowsDir = path.join(buildDir, 'windows');
  
  if (!fs.existsSync(darwinDir)) {
    fs.mkdirSync(darwinDir, { recursive: true });
  }
  if (!fs.existsSync(windowsDir)) {
    fs.mkdirSync(windowsDir, { recursive: true });
  }
  
  // 生成通用图标
  console.log('📱 Generating universal icons...');
  for (const icon of iconSizes) {
    const outputPath = path.join(buildDir, icon.name);
    await generateIcon(icon.size, outputPath);
  }
  
  console.log('\n🍎 Generating macOS icons...');
  for (const icon of macOSIconSizes) {
    const outputPath = path.join(darwinDir, icon.name);
    await generateIcon(icon.size, outputPath);
  }
  
  console.log('\n🪟 Generating Windows icons...');
  for (const icon of windowsIconSizes) {
    const outputPath = path.join(windowsDir, icon.name);
    await generateIcon(icon.size, outputPath);
  }
  
  console.log('\n✨ Icon generation completed!');
  console.log(`📁 Icons saved to: ${buildDir}`);
  
  // 生成图标清单
  const manifest = {
    generated: new Date().toISOString(),
    universal: iconSizes.map(icon => ({ size: icon.size, file: icon.name })),
    macOS: macOSIconSizes.map(icon => ({ size: icon.size, file: `darwin/${icon.name}` })),
    windows: windowsIconSizes.map(icon => ({ size: icon.size, file: `windows/${icon.name}` }))
  };
  
  fs.writeFileSync(
    path.join(buildDir, 'icon-manifest.json'),
    JSON.stringify(manifest, null, 2)
  );
  
  console.log('📋 Icon manifest saved to icon-manifest.json');
}

// 运行图标生成
generateAllIcons().catch(error => {
  console.error('❌ Icon generation failed:', error);
  process.exit(1);
});
