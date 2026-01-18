# SSL 证书配置说明

## 概述

本项目已配置 SSL 证书，支持 HTTPS 访问，特别是为了支持班级签到功能的摄像头扫码功能。

## 文件结构

```
scripts/
├── nginx/
│   ├── bingoj.conf          # Nginx SSL 配置文件
│   ├── ssl/                 # SSL 证书目录（已加入 .gitignore）
│   │   ├── bingoj.cn.crt   # SSL 证书文件
│   │   └── bingoj.cn.key   # SSL 私钥文件
│   └── README.md           # 本文档
├── deploy.sh               # 主部署脚本（包含 SSL 配置）
└── deploy-ssl.sh           # SSL 快速部署脚本（仅更新证书）
```

## 证书信息

- **域名**: bingoj.cn, www.bingoj.cn
- **证书类型**: TrustAsia C1 DV Free (RSA)
- **有效期**: 90天（需要在到期前续期）
- **颁发机构**: Certum Trusted Network CA

## 部署方式

### 方式 1：完整部署（包含 SSL）

```bash
./scripts/deploy.sh
```

这将自动：
1. 构建 Docker 镜像
2. 上传镜像到服务器
3. **上传并配置 SSL 证书**
4. **安装并配置 Nginx**
5. 启动所有服务

### 方式 2：仅部署 SSL 证书（快速更新）

如果只是更新证书而不重新构建镜像：

```bash
./scripts/deploy-ssl.sh
```

这将：
1. 上传 SSL 证书到服务器
2. 配置 Nginx
3. 重载 Nginx 配置
4. 测试 HTTPS 连接

## 验证 SSL 配置

### 1. 检查 HTTPS 访问

```bash
curl -I https://bingoj.cn
```

期望返回：`HTTP/2 200` 或 `HTTP/1.1 200`

### 2. 检查 HTTP 重定向

```bash
curl -I http://bingoj.cn
```

期望返回：`HTTP/1.1 301 Moved Permanently`（重定向到 HTTPS）

### 3. 查看 SSL 证书详情

```bash
echo | openssl s_client -connect bingoj.cn:443 -servername bingoj.cn
```

### 4. 查看证书有效期

```bash
echo | openssl s_client -connect bingoj.cn:443 -servername bingoj.cn 2>/dev/null | openssl x509 -noout -dates
```

## 架构说明

### Nginx 反向代理架构

```
用户浏览器
    ↓
    ↓ HTTPS (443端口)
    ↓
宿主机 Nginx (SSL终端)
    ↓
    ↓ HTTP (80端口)
    ↓
Docker 容器 (hoj-frontend)
    ↓
    ↓ 内部网络
    ↓
hist-oj 容器 (9527端口)
```

### 端口映射

| 服务 | 端口 | 说明 |
|------|------|------|
| Nginx (HTTPS) | 443 | 对外提供 HTTPS 服务 |
| Nginx (HTTP) | 80 | HTTP 重定向到 HTTPS |
| hoj-frontend | 80 (容器内部) | 前端静态文件服务 |
| hist-oj API | 9527 | 后端 API 服务 |
| registration-backend | 8080 | 报名系统 API |

## 重要提示

### ⚠️ 安全警告

1. **私钥保护**：
   - 私钥文件 `bingoj.cn.key` 已加入 `.gitignore`
   - 永远不要将私钥文件提交到 Git 仓库
   - 私钥文件权限应为 `600`（仅所有者可读写）

2. **证书更新**：
   - 证书有效期为 90 天
   - 请在到期前 7-10 天续期
   - 续期后使用 `./scripts/deploy-ssl.sh` 快速更新

3. **配置备份**：
   - 旧配置会自动备份为 `bingoj.conf.bak`
   - 如需回滚，使用备份文件

### 班级签到功能

配置 SSL 后，班级签到功能的摄像头扫码可以正常使用：

**学生端签到页面**：`https://bingoj.cn/classroom/student`

**测试摄像头权限**：
```javascript
// 在浏览器控制台执行
navigator.mediaDevices.getUserMedia({ video: true })
  .then(stream => {
    console.log('✅ 摄像头权限正常');
    stream.getTracks().forEach(track => track.stop());
  })
  .catch(err => {
    console.error('❌ 摄像头错误:', err);
  });
```

## 故障排查

### 问题 1：HTTPS 无法访问

```bash
# 检查 Nginx 状态
ssh root@43.143.133.62 "systemctl status nginx"

# 检查端口监听
ssh root@43.143.133.62 "netstat -tlnp | grep :443"

# 查看 Nginx 日志
ssh root@43.143.133.62 "tail -f /var/log/nginx/bingoj_error.log"
```

### 问题 2：证书错误

```bash
# 检查证书文件
ssh root@43.143.133.62 "ls -la /etc/nginx/ssl/"

# 检查证书有效期
ssh root@43.143.133.62 "openssl x509 -in /etc/nginx/ssl/bingoj.cn.crt -noout -dates"

# 测试 Nginx 配置
ssh root@43.143.133.62 "nginx -t"
```

### 问题 3：摄像头无法访问

1. 确认使用 HTTPS 访问（HTTP 不支持摄像头）
2. 检查浏览器控制台是否有错误
3. 确认 Nginx 配置中包含：
   ```nginx
   add_header Permissions-Policy "camera=self, microphone=self, geolocation=(self)";
   ```

## 更新证书步骤

### 1. 从腾讯云下载新证书

下载 Nginx 格式的证书包

### 2. 解压并复制

```bash
# 解压证书包
unzip bingoj.cn_nginx.zip

# 复制到项目目录
cp bingoj.cn_nginx/bingoj.cn_bundle.crt scripts/nginx/ssl/bingoj.cn.crt
cp bingoj.cn_nginx/bingoj.cn.key scripts/nginx/ssl/bingoj.cn.key
```

### 3. 部署到服务器

```bash
./scripts/deploy-ssl.sh
```

### 4. 验证

```bash
curl -I https://bingoj.cn
```

## 相关文档

- [腾讯云 SSL 证书文档](https://cloud.tencent.com/document/product/400)
- [Nginx SSL 配置官方文档](https://nginx.org/en/docs/http/ngx_http_ssl_module.html)
- [Let's Encrypt (免费 SSL 证书)](https://letsencrypt.org/)

## 联系支持

如有问题，请查看：
- 项目 Issue: [GitHub Issues](https://github.com/your-repo/issues)
- 技术文档: `/docs` 目录
