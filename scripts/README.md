# 部署脚本使用说明

## 快速开始

### 方式一：完整部署（推荐首次部署）

```bash
cd /Users/tianjiajie/Desktop/hoj2
./scripts/deploy.sh
```

这个脚本会：
1. ✅ 构建 hist-oj 和 hoj-frontend 镜像
2. ✅ 保存镜像为 tar.gz 文件
3. ✅ 上传到服务器
4. ✅ 在服务器上加载并启动容器
5. ✅ 验证服务状态
6. ✅ 清理临时文件

### 方式二：快速部署（只部署单个服务）

```bash
cd /Users/tianjiajie/Desktop/hoj2

# 只部署 hist-oj
./scripts/quick-deploy.sh hist-oj

# 只部署前端
./scripts/quick-deploy.sh frontend

# 部署所有服务
./scripts/quick-deploy.sh all
```

## 常用命令

### 1. 完整部署（包含构建）
```bash
./scripts/deploy.sh
```

### 2. 跳过构建，直接部署（如果镜像已存在）
```bash
./scripts/deploy.sh --skip-build
```

### 3. 只部署 hist-oj
```bash
./scripts/quick-deploy.sh hist-oj
```

### 4. 只部署前端
```bash
./scripts/quick-deploy.sh frontend
```

## 部署后验证

### 验证 hist-oj 服务
```bash
# 健康检查
curl http://43.143.133.62:9527/health

# 测试 Rating API
curl http://43.143.133.62:9527/api/rating/contest/info/1002
```

### 验证前端服务
```bash
# 访问首页
curl http://43.143.133.62/

# 测试前端访问 hist-oj
ssh root@43.143.133.62 "docker exec hoj-frontend curl -s http://hist-oj:9527/health"
```

### 查看服务状态
```bash
ssh root@43.143.133.62 "docker ps | grep -E 'hist-oj|hoj-frontend'"
```

### 查看日志
```bash
# hist-oj 日志
ssh root@43.143.133.62 "docker logs --tail 100 hist-oj"

# 前端日志
ssh root@43.143.133.62 "docker logs --tail 100 hoj-frontend"
```

## 故障排查

### 问题 1：前端无法访问 hist-oj（502 错误）

**解决方案**：
```bash
ssh root@43.143.133.62 "docker network connect hoj_hoj-network hist-oj"
```

### 问题 2：服务启动失败

**查看日志**：
```bash
ssh root@43.143.133.62 "docker logs hist-oj"
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
'"
```

## 回滚操作

如果部署出现问题，可以快速回滚：

```bash
ssh root@43.143.133.62 << 'EOF'
# 停止新容器
docker stop hist-oj hoj-frontend

# 删除新容器
docker rm hist-oj hoj-frontend

# 重新加载旧镜像（如果有备份）
docker load < /opt/backup/hist-oj-old.tar.gz
docker load < /opt/backup/hoj-frontend-old.tar.gz

# 启动旧容器
docker run -d --name hist-oj --network hoj_hoj-network -p 9527:9527 \
  -v /workspace/hoj-deploy/distributed/main/hist-oj/configs:/app/configs \
  --restart unless-stopped hist-oj:old

docker run -d --name hoj-frontend --network hoj_hoj-network -p 80:80 -p 443:443 \
  --restart unless-stopped hoj-frontend:old
EOF
```

## 注意事项

1. **网络配置**：确保 hist-oj 和 hoj-frontend 都在 `hoj_hoj-network` 网络中
2. **配置文件**：hist-oj 的配置文件位于 `/workspace/hoj-deploy/distributed/main/hist-oj/configs/config.yaml`
3. **端口占用**：确保 9527、80、443 端口未被占用
4. **权限问题**：脚本需要 Docker 和 SSH 权限

## 更新日志

- 2025-12-19: 创建自动化部署脚本
- 2025-12-19: 添加快速部署脚本
- 2025-12-19: 修复网络配置问题
