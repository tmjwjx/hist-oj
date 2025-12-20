# HOJ2 Rating 系统部署指南

## 一、部署架构

### 服务组成
- **hist-oj**: Rating 计算服务（Go）- 端口 9527
- **hoj-frontend**: 前端服务（Vue + Nginx）- 端口 80/443
- **hoj-backend**: 后端服务（Spring Boot）- 端口 6688（已存在）
- **hoj-mysql**: 数据库服务（MySQL）- 端口 3306（已存在）

### 网络配置
- 所有服务必须在 `hoj_hoj-network` 网络中
- hist-oj 需要能被前端 Nginx 通过容器名访问

## 二、快速部署（推荐）

### 方式一：使用部署脚本（最简单）

```bash
# 在项目根目录执行
./scripts/deploy.sh
```

脚本会自动完成：
1. 构建 hist-oj 和 hoj-frontend 镜像
2. 保存镜像为 tar.gz 文件
3. 上传到服务器
4. 在服务器上加载镜像
5. 停止并删除旧容器
6. 启动新容器
7. 确保网络配置正确

### 方式二：使用 Docker Compose（本地测试）

```bash
# 在项目根目录执行
docker-compose up -d
```

**注意**：此方式仅适用于本地测试，生产环境请使用部署脚本。

## 三、手动部署步骤

### 3.1 构建镜像

```bash
# 构建 hist-oj 镜像
cd hist-oj
docker build --platform linux/amd64 -t hist-oj:latest .

# 构建前端镜像
cd ../hoj-vue
docker build --platform linux/amd64 -t hoj-frontend:latest .
```

### 3.2 保存并上传镜像

```bash
# 保存镜像
cd ..
docker save hist-oj:latest | gzip > hist-oj.tar.gz
docker save hoj-frontend:latest | gzip > hoj-frontend.tar.gz

# 上传到服务器
sshpass -p "n208966737" scp hist-oj.tar.gz root@43.143.133.62:/opt/
sshpass -p "n208966737" scp hoj-frontend.tar.gz root@43.143.133.62:/opt/
```

### 3.3 在服务器上部署

```bash
# SSH 到服务器
ssh root@43.143.133.62

# 加载镜像
cd /opt
gunzip -c hist-oj.tar.gz | docker load
gunzip -c hoj-frontend.tar.gz | docker load

# 停止并删除旧容器
docker stop hist-oj hoj-frontend
docker rm hist-oj hoj-frontend

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

# 验证服务状态
docker ps | grep -E "hist-oj|hoj-frontend"
docker logs --tail 50 hist-oj
docker logs --tail 50 hoj-frontend
```

### 3.4 验证部署

```bash
# 测试 hist-oj 健康检查
curl http://localhost:9527/health

# 测试 Rating API
curl http://localhost:9527/api/rating/contest/info/1002

# 从前端容器测试访问 hist-oj
docker exec hoj-frontend curl -s http://hist-oj:9527/health
```

## 四、常见问题排查

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

# 如果 hist-oj 不在 hoj_hoj-network 中，添加网络
docker network connect hoj_hoj-network hist-oj

# 验证
docker exec hoj-frontend curl -s http://hist-oj:9527/health
```

### 问题 2：Rating 计算结果不正确

**可能原因**：
1. 时间范围过滤问题
2. 使用了错误的字段（UseTime vs Time）
3. 排名计算逻辑错误

**排查步骤**：
```bash
# 查看 hist-oj 日志
docker logs --tail 200 hist-oj | grep -E "contest_id|Rating"

# 检查数据库中的 Rating 数据
docker exec hoj-mysql mysql -uroot -phist2025 -e "
USE hoj;
SELECT r.rank, u.username, r.old_rating, r.new_rating, r.rating_change
FROM rating_history r
JOIN user_info u ON r.uid = u.uuid
WHERE r.contest_id = 1002
ORDER BY r.rank LIMIT 10;
"

# 重新计算 Rating
curl -X POST http://localhost:9527/api/rating/calculate/1002
```

### 问题 3：前端显示不正常

**检查前端 Nginx 配置**：
```bash
docker exec hoj-frontend cat /etc/nginx/nginx.conf
```

**确认配置包含**：
- `/api/rating` 代理到 `hist-oj:9527`
- `/api` 代理到 `hoj-backend:6688`
- DNS resolver 配置：`resolver 127.0.0.11 valid=30s;`

## 五、配置文件说明

### hist-oj 配置文件

位置：`hist-oj/configs/config.yaml`

关键配置项：
```yaml
server:
  port: 9527

database:
  host: hoj-mysql
  port: 3306
  user: root
  password: hist2025
  dbname: hoj

rating:
  k_factor: 32
  min_rating: 800
  default_rating: 1200
```

### 前端 Nginx 配置

位置：`hoj-vue/nginx.conf`

关键配置：
```nginx
# Rating API 代理
location /api/rating {
    resolver 127.0.0.11 valid=30s;
    set $rating_backend "hist-oj:9527";
    proxy_pass http://$rating_backend;
    # ... 其他配置
}

# HOJ API 代理
location /api {
    resolver 127.0.0.11 valid=30s;
    set $backend "hoj-backend:6688";
    proxy_pass http://$backend;
    # ... 其他配置
}
```

## 六、回滚方案

如果部署出现问题，可以快速回滚：

```bash
# 停止新容器
docker stop hist-oj hoj-frontend

# 启动旧容器（如果还存在）
docker start hist-oj-old hoj-frontend-old

# 或者重新加载旧镜像
docker load < /opt/backup/hist-oj-old.tar.gz
docker load < /opt/backup/hoj-frontend-old.tar.gz
```

## 七、监控和日志

### 查看服务状态
```bash
docker ps | grep -E "hist-oj|hoj-frontend"
```

### 查看日志
```bash
# 实时日志
docker logs -f hist-oj
docker logs -f hoj-frontend

# 最近日志
docker logs --tail 100 hist-oj
docker logs --tail 100 hoj-frontend

# 搜索错误
docker logs hist-oj 2>&1 | grep -i error
```

### 健康检查
```bash
# hist-oj 健康检查
curl http://43.143.133.62:9527/health

# 前端健康检查
curl http://43.143.133.62/

# 从服务器内部检查
ssh root@43.143.133.62 "docker exec hoj-frontend curl -s http://hist-oj:9527/health"
```

## 八、性能优化建议

1. **数据库连接池**：hist-oj 已配置连接池，默认最大连接数 10
2. **缓存策略**：Rating 数据会缓存在内存中，减少数据库查询
3. **日志轮转**：Docker 日志已配置自动轮转（max-size: 10m, max-file: 3）
4. **健康检查**：所有服务都配置了健康检查，自动重启异常容器

## 九、安全注意事项

1. **密码管理**：生产环境应使用环境变量或密钥管理服务
2. **端口暴露**：9527 端口仅用于内部通信，不应对外暴露
3. **网络隔离**：使用 Docker 网络隔离不同服务
4. **日志脱敏**：确保日志中不包含敏感信息

## 十、更新记录

- 2025-12-19: 修复 Rating 计算逻辑，使用 Time 字段而非 UseTime
- 2025-12-19: 修复时间范围过滤问题，添加 submit_time 过滤
- 2025-12-19: 修复 Docker 网络配置问题，确保 hist-oj 在 hoj_hoj-network 中
- 2025-12-18: 初始版本，实现基础 Rating 功能
