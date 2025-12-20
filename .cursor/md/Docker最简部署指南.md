# Docker 最简部署指南

## 🚀 一键自动化部署（推荐）

```bash
cd /Users/tianjiajie/Desktop/hoj2

# 完整部署（构建 + 上传 + 部署）
./scripts/deploy.sh

# 只部署 hist-oj
./scripts/quick-deploy.sh hist-oj

# 只部署前端
./scripts/quick-deploy.sh frontend
```

**自动化脚本会完成**：
1. ✅ 构建 Docker 镜像（amd64 架构）
2. ✅ 保存并压缩镜像
3. ✅ 上传到服务器
4. ✅ 在服务器上加载镜像
5. ✅ 停止并删除旧容器
6. ✅ 启动新容器（确保在 hoj_hoj-network 网络中）
7. ✅ 验证服务状态

---

## 📦 手动部署（备用方案）

### 方式一：源码打包部署

```bash
cd /Users/tianjiajie/Desktop/hoj2

# 打包源码
tar -czf hoj2.tar.gz \
    --exclude='node_modules' \
    --exclude='dist' \
    --exclude='.git' \
    --exclude='*.log' \
    docker-compose.yml \
    hist-oj/ \
    hoj-vue/

# 上传到服务器
sshpass -p "n208966737" scp hoj2.tar.gz root@43.143.133.62:/opt/
```

**服务器端操作**：
```bash
cd /opt

# 解压
mkdir -p hoj2
tar -xzf hoj2.tar.gz -C hoj2
cd hoj2

# 停止旧容器
docker stop hoj-frontend hist-oj 2>/dev/null
docker rm hoj-frontend hist-oj 2>/dev/null

# 启动服务（会自动构建镜像）
docker-compose up -d --build

# 确保 hist-oj 在正确的网络中
docker network connect hoj_hoj-network hist-oj 2>/dev/null

# 查看状态
docker ps | grep -E "hist-oj|hoj-frontend"
```

### 方式二：镜像打包部署

```bash
cd /Users/tianjiajie/Desktop/hoj2

# 构建镜像
cd hist-oj
docker build --platform linux/amd64 -t hist-oj:latest .

cd ../hoj-vue
docker build --platform linux/amd64 -t hoj-frontend:latest .

# 保存镜像
cd ..
docker save hist-oj:latest | gzip > hist-oj.tar.gz
docker save hoj-frontend:latest | gzip > hoj-frontend.tar.gz

# 上传到服务器
sshpass -p "n208966737" scp hist-oj.tar.gz root@43.143.133.62:/opt/
sshpass -p "n208966737" scp hoj-frontend.tar.gz root@43.143.133.62:/opt/
```

**服务器端操作**：
```bash
cd /opt

# 加载镜像
gunzip -c hist-oj.tar.gz | docker load
gunzip -c hoj-frontend.tar.gz | docker load

# 停止并删除旧容器
docker stop hist-oj hoj-frontend 2>/dev/null
docker rm hist-oj hoj-frontend 2>/dev/null

# 启动 hist-oj
docker run -d \
  --name hist-oj \
  --network hoj_hoj-network \
  -p 9527:9527 \
  -v /workspace/hoj-deploy/distributed/main/hist-oj/configs:/app/configs \
  --restart unless-stopped \
  hist-oj:latest

# 启动前端
docker run -d \
  --name hoj-frontend \
  --network hoj_hoj-network \
  -p 80:80 \
  -p 443:443 \
  --restart unless-stopped \
  hoj-frontend:latest

# 验证
docker ps | grep -E "hist-oj|hoj-frontend"
curl http://localhost:9527/health
```

---

## ✅ 验证部署

```bash
# 测试 hist-oj 健康检查
curl http://43.143.133.62:9527/health

# 测试 Rating API
curl http://43.143.133.62:9527/api/rating/contest/info/1002

# 测试前端访问 hist-oj
ssh root@43.143.133.62 "docker exec hoj-frontend curl -s http://hist-oj:9527/health"

# 浏览器访问
open http://43.143.133.62
```

---

## 🔧 常用命令

```bash
# 查看日志
ssh root@43.143.133.62 "docker logs --tail 100 hist-oj"
ssh root@43.143.133.62 "docker logs --tail 100 hoj-frontend"

# 重启服务
ssh root@43.143.133.62 "docker restart hist-oj hoj-frontend"

# 查看容器状态
ssh root@43.143.133.62 "docker ps | grep -E 'hist-oj|hoj-frontend'"

# 进入容器调试
ssh root@43.143.133.62 "docker exec -it hist-oj sh"
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
ssh root@43.143.133.62 "docker network connect hoj_hoj-network hist-oj"
```

### 问题 2：服务启动失败

**查看日志**：
```bash
ssh root@43.143.133.62 "docker logs --tail 200 hist-oj"
```

**重启服务**：
```bash
ssh root@43.143.133.62 "docker restart hist-oj"
```

### 问题 3：Rating 计算不正确

**重新计算 Rating**：
```bash
curl -X POST http://43.143.133.62:9527/api/rating/calculate/1002
```

**查看 Rating 数据**：
```bash
ssh root@43.143.133.62 "docker exec hoj-mysql mysql -uroot -phist2025 -e '
USE hoj;
SELECT r.rank, u.username, r.old_rating, r.new_rating, r.rating_change
FROM rating_history r
JOIN user_info u ON r.uid = u.uuid
WHERE r.contest_id = 1002
ORDER BY r.rank LIMIT 10;
' 2>/dev/null"
```

### 问题 4：检查网络配置

```bash
# 查看容器网络
ssh root@43.143.133.62 "docker inspect hist-oj --format='{{range \$key, \$value := .NetworkSettings.Networks}}{{\$key}} {{end}}'"

# 查看网络详情
ssh root@43.143.133.62 "docker network inspect hoj_hoj-network"
```

---

## 📚 相关文档

- **详细部署指南**: [deploy-guide.md](./deploy-guide.md)
- **镜像部署指令**: [Docker镜像部署指令.md](./Docker镜像部署指令.md)
- **脚本使用说明**: [../scripts/README.md](../scripts/README.md)

---

## 🎯 部署检查清单

**部署前**：
- [ ] Docker 已安装并运行
- [ ] sshpass 已安装（`brew install sshpass`）
- [ ] 服务器 SSH 连接正常
- [ ] 服务器 `hoj_hoj-network` 网络已存在

**部署后**：
- [ ] hist-oj 容器运行正常
- [ ] hoj-frontend 容器运行正常
- [ ] hist-oj 健康检查通过
- [ ] 前端可以访问 hist-oj
- [ ] Rating API 返回正确数据
- [ ] 前端页面正常显示

---

## 🔄 更新记录

- **2025-12-19**: 添加自动化部署脚本
- **2025-12-19**: 修复 Docker 网络配置问题
- **2025-12-19**: 修复 Rating 计算逻辑（使用 Time 字段）
- **2025-12-19**: 修复时间范围过滤问题
