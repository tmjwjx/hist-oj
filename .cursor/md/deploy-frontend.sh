#!/bin/bash

# 前端自动化部署脚本
# 用途：构建前端镜像并部署到服务器

set -e  # 遇到错误立即退出

# 配置变量
SERVER_IP="43.143.133.62"
SERVER_USER="root"
SERVER_PASSWORD="n208966737"
PROJECT_DIR="/Users/tianjiajie/Desktop/hoj2"
IMAGE_NAME="hoj2-hoj-frontend:latest"
TAR_FILE="hoj-frontend.tar"

echo "=========================================="
echo "前端自动化部署脚本"
echo "=========================================="
echo ""

# 1. 构建前端镜像
echo "步骤 1/5: 构建前端镜像（AMD64 架构）..."
cd "$PROJECT_DIR"
docker buildx build --platform linux/amd64 -t "$IMAGE_NAME" --load ./hoj-vue
echo "✅ 镜像构建完成"
echo ""

# 2. 保存镜像为 tar 文件
echo "步骤 2/5: 保存镜像为 tar 文件..."
docker save -o "$TAR_FILE" "$IMAGE_NAME"
echo "✅ 镜像已保存为 $TAR_FILE"
echo ""

# 3. 上传镜像到服务器
echo "步骤 3/5: 上传镜像到服务器..."
sshpass -p "$SERVER_PASSWORD" scp "$TAR_FILE" "$SERVER_USER@$SERVER_IP:/opt/"
echo "✅ 镜像已上传到服务器"
echo ""

# 4. 在服务器上部署
echo "步骤 4/5: 在服务器上部署..."
sshpass -p "$SERVER_PASSWORD" ssh "$SERVER_USER@$SERVER_IP" << 'EOF'
cd /opt

# 加载新镜像
echo "  - 加载镜像..."
docker load -i hoj-frontend.tar

# 停止并删除旧容器
echo "  - 停止旧容器..."
docker stop hoj-frontend 2>/dev/null || true
docker rm hoj-frontend 2>/dev/null || true

# 启动新容器
echo "  - 启动新容器..."
docker run -d \
  --name hoj-frontend \
  --restart unless-stopped \
  -p 80:80 \
  -p 443:443 \
  -e TZ=Asia/Shanghai \
  --network hoj_hoj-network \
  hoj2-hoj-frontend:latest

# 等待容器启动
sleep 3

# 验证容器状态
echo "  - 验证容器状态..."
docker ps | grep hoj-frontend
EOF
echo "✅ 部署完成"
echo ""

# 5. 验证部署
echo "步骤 5/5: 验证部署..."
echo "  - 测试 Nginx 配置..."
sshpass -p "$SERVER_PASSWORD" ssh "$SERVER_USER@$SERVER_IP" "docker exec hoj-frontend nginx -t"

echo "  - 测试 Rating API..."
sshpass -p "$SERVER_PASSWORD" ssh "$SERVER_USER@$SERVER_IP" "curl -s http://localhost/api/rating/contest/info/1008 | head -c 100"
echo ""
echo "✅ 验证通过"
echo ""

# 6. 清理本地 tar 文件
echo "清理本地临时文件..."
rm -f "$TAR_FILE"
echo "✅ 清理完成"
echo ""

echo "=========================================="
echo "🎉 部署成功！"
echo "=========================================="
echo ""
echo "访问地址: http://bingoj.cn"
echo "Rating API: http://bingoj.cn/api/rating/contest/info/1008"
echo ""
