# Docker 镜像部署指令

## 🚀 推荐方式：使用自动化脚本

```bash
cd /Users/tianjiajie/Desktop/hoj2

# 完整部署
./scripts/deploy.sh

# 只部署 hist-oj
./scripts/quick-deploy.sh hist-oj

# 只部署前端
./scripts/quick-deploy.sh frontend
```

**自动化脚本会自动处理**：
- ✅ 构建 amd64 架构镜像
- ✅ 保存并压缩镜像
- ✅ 上传到服务器
- ✅ 加载镜像并启动容器
- ✅ 确保网络配置正确（hoj_hoj-network）
- ✅ 验证服务状态

---

## 📦 手动部署流程

### 1. 本地构建并保存镜像

```bash
cd /Users/tianjiajie/Desktop/hoj2

# 构建 hist-oj 镜像
cd hist-oj
docker build --platform linux/amd64 -t hist-oj:latest .

# 构建前端镜像
cd ../hoj-vue
docker build --platform linux/amd64 -t hoj-frontend:latest .

# 保存镜像为 tar.gz 文件
cd ..
docker save hist-oj:latest | gzip > hist-oj.tar.gz
docker save hoj-frontend:latest | gzip > hoj-frontend.tar.gz

# 查看文件大小
ls -lh *.tar.gz
```

### 2. 上传镜像到服务器

```bash
# 上传 hist-oj 镜像
sshpass -p "n208966737" scp hist-oj.tar.gz root@43.143.133.62:/opt/

# 上传前端镜像
sshpass -p "n208966737" scp hoj-frontend.tar.gz root@43.143.133.62:/opt/

# 如果需要，上传配置文件
sshpass -p "n208966737" scp -r hist-oj/configs root@43.143.133.62:/workspace/hoj-deploy/distributed/main/hist-oj/
```

### 3. 服务器端操作

#### 3.1 加载镜像

```bash
# SSH 登录服务器
ssh root@43.143.133.62

# 进入目录
cd /opt

# 加载 hist-oj 镜像
gunzip -c hist-oj.tar.gz | docker load

# 加载前端镜像
gunzip -c hoj-frontend.tar.gz | docker load

# 验证镜像已加载
docker images | grep -E "hist-oj|hoj-frontend"
```

#### 3.2 停止并删除旧容器

```bash
# 停止旧容器
docker stop hist-oj hoj-frontend 2>/dev/null

# 删除旧容器
docker rm hist-oj hoj-frontend 2>/dev/null
```

#### 3.3 启动新容器

```bash
# 启动 hist-oj（确保在 hoj_hoj-network 网络中）
docker run -d \
  --name hist-oj \
  --network hoj_hoj-network \
  -p 9527:9527 \
  -v /workspace/hoj-deploy/distributed/main/hist-oj/configs:/app/configs \
  --restart unless-stopped \
  hist-oj:latest

# 启动前端（确保在 hoj_hoj-network 网络中）
docker run -d \
  --name hoj-frontend \
  --network hoj_hoj-network \
  -p 80:80 \
  -p 443:443 \
  --restart unless-stopped \
  hoj-frontend:latest

# 等待服务启动
sleep 5
```

#### 3.4 验证服务

```bash
# 查看容器状态
docker ps | grep -E "hist-oj|hoj-frontend"

# 测试 hist-oj 健康检查
curl -s http://localhost:9527/health

# 测试前端访问 hist-oj
docker exec hoj-frontend curl -s http://hist-oj:9527/health

# 查看日志
docker logs --tail 50 hist-oj
docker logs --tail 50 hoj-frontend
```

---

## 🔧 常用管理命令

### 容器管理

```bash
# 重启服务
docker restart hist-oj hoj-frontend

# 停止服务
docker stop hist-oj hoj-frontend

# 启动服务
docker start hist-oj hoj-frontend

# 删除容器（需要先停止）
docker stop hist-oj hoj-frontend
docker rm hist-oj hoj-frontend
```

### 日志查看

```bash
# 查看实时日志
docker logs -f hist-oj
docker logs -f hoj-frontend

# 查看最近日志
docker logs --tail 100 hist-oj
docker logs --tail 100 hoj-frontend

# 搜索错误日志
docker logs hist-oj 2>&1 | grep -i error
docker logs hoj-frontend 2>&1 | grep -i error
```

### 容器调试

```bash
# 进入容器
docker exec -it hist-oj sh
docker exec -it hoj-frontend sh

# 查看容器详细信息
docker inspect hist-oj
docker inspect hoj-frontend

# 查看容器资源使用
docker stats hist-oj hoj-frontend
```

### 镜像管理

```bash
# 查看镜像
docker images | grep -E "hist-oj|hoj-frontend"

# 删除镜像（需要先删除容器）
docker rmi hist-oj:latest hoj-frontend:latest

# 清理未使用的镜像
docker image prune -a
```

---

## 🚨 故障排查

### 问题 1：前端无法访问 hist-oj（502 错误）

**症状**：
```
hist-oj could not be resolved (2: Server failure)
```

**原因**：hist-oj 和 hoj-frontend 不在同一个 Docker 网络中

**解决方案**：
```bash
# 检查网络配置
docker inspect hist-oj --format='{{range $key, $value := .NetworkSettings.Networks}}{{$key}} {{end}}'
docker inspect hoj-frontend --format='{{range $key, $value := .NetworkSettings.Networks}}{{$key}} {{end}}'

# 将 hist-oj 加入到 hoj_hoj-network 网络
docker network connect hoj_hoj-network hist-oj

# 验证
docker exec hoj-frontend curl -s http://hist-oj:9527/health
```

### 问题 2：容器启动失败

**查看详细日志**：
```bash
docker logs --tail 200 hist-oj
docker logs --tail 200 hoj-frontend
```

**检查容器状态**：
```bash
docker ps -a | grep -E "hist-oj|hoj-frontend"
```

**重新启动**：
```bash
docker restart hist-oj hoj-frontend
```

### 问题 3：端口冲突

**检查端口占用**：
```bash
netstat -tlnp | grep -E "9527|80|443"
lsof -i :9527
lsof -i :80
```

**解决方案**：
- 停止占用端口的进程
- 或修改容器端口映射

### 问题 4：网络问题

**检查网络**：
```bash
# 查看所有网络
docker network ls

# 查看 hoj_hoj-network 详情
docker network inspect hoj_hoj-network

# 查看容器网络配置
docker inspect hist-oj --format='{{json .NetworkSettings.Networks}}' | jq
```

**重新配置网络**：
```bash
# 断开旧网络
docker network disconnect main_hoj-network hist-oj 2>/dev/null

# 连接到正确的网络
docker network connect hoj_hoj-network hist-oj
```

### 问题 5：配置文件问题

**检查配置文件**：
```bash
# 查看配置文件是否存在
ls -la /workspace/hoj-deploy/distributed/main/hist-oj/configs/

# 查看配置内容
cat /workspace/hoj-deploy/distributed/main/hist-oj/configs/config.yaml

# 从容器内查看
docker exec hist-oj ls -la /app/configs/
docker exec hist-oj cat /app/configs/config.yaml
```

---

## 🔄 更新部署

### 快速更新（推荐）

```bash
# 本地执行
cd /Users/tianjiajie/Desktop/hoj2
./scripts/quick-deploy.sh hist-oj
```

### 手动更新

```bash
# 1. 本地重新构建镜像
cd /Users/tianjiajie/Desktop/hoj2/hist-oj
docker build --platform linux/amd64 -t hist-oj:latest .

# 2. 保存并上传
cd ..
docker save hist-oj:latest | gzip > hist-oj.tar.gz
sshpass -p "n208966737" scp hist-oj.tar.gz root@43.143.133.62:/opt/

# 3. 服务器端重新部署
ssh root@43.143.133.62 << 'EOF'
cd /opt
gunzip -c hist-oj.tar.gz | docker load
docker stop hist-oj
docker rm hist-oj
docker run -d \
  --name hist-oj \
  --network hoj_hoj-network \
  -p 9527:9527 \
  -v /workspace/hoj-deploy/distributed/main/hist-oj/configs:/app/configs \
  --restart unless-stopped \
  hist-oj:latest
docker ps | grep hist-oj
curl -s http://localhost:9527/health
EOF
```

---

## 📊 监控和健康检查

### 健康检查

```bash
# hist-oj 健康检查
curl http://43.143.133.62:9527/health

# Rating API 测试
curl http://43.143.133.62:9527/api/rating/contest/info/1002

# 前端健康检查
curl http://43.143.133.62/

# 前端访问 hist-oj 测试
ssh root@43.143.133.62 "docker exec hoj-frontend curl -s http://hist-oj:9527/health"
```

### 性能监控

```bash
# 查看容器资源使用
docker stats hist-oj hoj-frontend --no-stream

# 查看容器进程
docker top hist-oj
docker top hoj-frontend

# 查看容器网络统计
docker exec hist-oj netstat -s
```

---

## 📝 注意事项

### 关键配置

1. **网络配置**：hist-oj 和 hoj-frontend 必须在 `hoj_hoj-network` 网络中
2. **配置文件**：hist-oj 配置文件位于 `/workspace/hoj-deploy/distributed/main/hist-oj/configs/config.yaml`
3. **端口映射**：
   - hist-oj: 9527
   - hoj-frontend: 80, 443
4. **数据库连接**：hist-oj 需要连接到 hoj-mysql（在同一网络中）

### 安全建议

1. **密码管理**：生产环境应使用环境变量或密钥管理服务
2. **端口暴露**：9527 端口仅用于内部通信，不应对外暴露
3. **日志管理**：定期清理日志文件，避免磁盘占满
4. **备份策略**：定期备份配置文件和数据库

### 性能优化

1. **资源限制**：可以通过 `--memory` 和 `--cpus` 限制容器资源
2. **日志轮转**：使用 `--log-opt max-size=10m --log-opt max-file=3`
3. **健康检查**：容器已配置健康检查，自动重启异常容器

---

## 📚 相关文档

- **最简部署指南**: [Docker最简部署指南.md](./Docker最简部署指南.md)
- **详细部署指南**: [deploy-guide.md](./deploy-guide.md)
- **脚本使用说明**: [../scripts/README.md](../scripts/README.md)

---

## 🔄 更新记录

- **2025-12-19**: 添加自动化部署脚本说明
- **2025-12-19**: 修复网络配置问题（确保在 hoj_hoj-network 中）
- **2025-12-19**: 更新部署流程，添加网络验证步骤
- **2025-12-18**: 初始版本
