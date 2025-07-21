# ProxyWoman 证书问题排查指南

## 常见错误信息

### "TLS handshake failed: remote error: tls: unknown certificate"

这个错误表明客户端不信任 ProxyWoman 生成的证书。

## 解决步骤

### 1. 检查证书是否存在

```bash
# 使用CLI检查
./proxywoman cert path

# 或者检查默认位置
ls ~/.proxywoman/ca-cert.pem
```

### 2. 查看证书安装说明

```bash
# 获取详细安装说明
./proxywoman cert install-help
```

### 3. 测试证书生成

```bash
# 测试证书生成功能
./proxywoman test-cert example.com 8443
```

### 4. 手动安装证书

#### macOS 安装步骤

1. **找到证书文件**
   ```bash
   open ~/.proxywoman/
   ```

2. **安装证书**
   - 双击 `ca-cert.pem` 文件
   - 选择"系统"钥匙串
   - 输入管理员密码

3. **设置信任**
   - 在钥匙串访问中找到"ProxyWoman Root CA"
   - 双击证书
   - 展开"信任"部分
   - 将"使用此证书时"设置为"始终信任"
   - 关闭窗口并再次输入密码

4. **验证安装**
   ```bash
   # 检查证书是否在系统钥匙串中
   security find-certificate -c "ProxyWoman Root CA" /Library/Keychains/System.keychain
   ```

#### Windows 安装步骤

1. **以管理员身份运行**
   - 右键点击命令提示符，选择"以管理员身份运行"

2. **安装证书**
   ```cmd
   certlm.msc
   ```
   - 导航到"受信任的根证书颁发机构" → "证书"
   - 右键 → "所有任务" → "导入"
   - 选择 `ca-cert.pem` 文件

3. **或使用命令行**
   ```cmd
   certutil -addstore -f "ROOT" "%USERPROFILE%\.proxywoman\ca-cert.pem"
   ```

#### Linux 安装步骤

```bash
# 复制证书到系统目录
sudo cp ~/.proxywoman/ca-cert.pem /usr/local/share/ca-certificates/proxywoman.crt

# 更新证书存储
sudo update-ca-certificates

# 对于某些发行版
sudo trust anchor ~/.proxywoman/ca-cert.pem
```

### 5. 浏览器特定设置

#### Chrome/Chromium
- 设置 → 隐私设置和安全性 → 安全 → 管理证书
- 导入到"受信任的根证书颁发机构"

#### Firefox
- 设置 → 隐私与安全 → 证书 → 查看证书
- 证书颁发机构 → 导入
- 勾选"信任此CA来标识网站"

#### Safari
- 使用系统钥匙串设置（与macOS步骤相同）

### 6. 重新生成证书

如果证书损坏或过期：

```bash
# 删除旧证书
rm ~/.proxywoman/ca-cert.pem
rm ~/.proxywoman/ca-key.pem

# 重新生成
./proxywoman cert regenerate

# 重新安装新证书
./proxywoman cert install-help
```

## 高级诊断

### 检查证书详情

```bash
# 查看证书信息
openssl x509 -in ~/.proxywoman/ca-cert.pem -text -noout

# 检查证书有效期
openssl x509 -in ~/.proxywoman/ca-cert.pem -dates -noout
```

### 测试TLS连接

```bash
# 测试特定域名的证书
openssl s_client -connect example.com:443 -CAfile ~/.proxywoman/ca-cert.pem

# 测试本地代理
openssl s_client -connect localhost:8080 -CAfile ~/.proxywoman/ca-cert.pem
```

### 检查系统代理设置

```bash
# macOS
scutil --proxy

# Linux
echo $http_proxy
echo $https_proxy

# Windows
netsh winhttp show proxy
```

## 常见问题

### Q: 安装证书后仍然出现错误
A: 
1. 重启浏览器
2. 清除浏览器缓存
3. 检查证书是否正确安装在"受信任的根证书颁发机构"
4. 确认证书没有过期

### Q: 某些网站仍然无法访问
A:
1. 检查网站是否使用证书钉扎(Certificate Pinning)
2. 尝试在隐私/无痕模式下访问
3. 检查是否有其他安全软件干扰

### Q: 移动设备如何安装证书
A:
1. 将证书文件传输到设备
2. iOS: 设置 → 通用 → VPN与设备管理 → 安装描述文件
3. Android: 设置 → 安全 → 加密与凭据 → 从存储设备安装

### Q: 企业环境中的证书问题
A:
1. 联系IT管理员
2. 可能需要将ProxyWoman证书添加到企业证书策略
3. 考虑使用企业CA签发的证书

## 预防措施

1. **定期检查证书有效期**
   ```bash
   ./proxywoman cert path | xargs openssl x509 -dates -noout -in
   ```

2. **备份证书文件**
   ```bash
   cp ~/.proxywoman/ca-cert.pem ~/backup/
   ```

3. **监控证书状态**
   - 在应用中添加证书状态检查
   - 设置证书过期提醒

## 自动化脚本

### 证书安装脚本 (macOS)

```bash
#!/bin/bash
CERT_PATH="$HOME/.proxywoman/ca-cert.pem"

if [ -f "$CERT_PATH" ]; then
    echo "Installing ProxyWoman CA certificate..."
    sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain "$CERT_PATH"
    echo "Certificate installed successfully"
else
    echo "Certificate file not found: $CERT_PATH"
    exit 1
fi
```

### 证书验证脚本

```bash
#!/bin/bash
CERT_PATH="$HOME/.proxywoman/ca-cert.pem"

echo "Checking certificate..."
if openssl x509 -checkend 86400 -noout -in "$CERT_PATH"; then
    echo "✅ Certificate is valid"
else
    echo "❌ Certificate will expire within 24 hours"
fi

echo "Certificate details:"
openssl x509 -subject -dates -noout -in "$CERT_PATH"
```

## 联系支持

如果问题仍然存在：

1. 收集诊断信息：
   ```bash
   ./proxywoman test-cert example.com > debug.log 2>&1
   ```

2. 提供以下信息：
   - 操作系统版本
   - 浏览器版本
   - 错误信息截图
   - 证书详情
   - 代理设置

3. 在GitHub Issues中报告问题，附上诊断信息
