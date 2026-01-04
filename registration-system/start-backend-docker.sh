#!/bin/bash
# 报名系统 Docker 快速启动脚本

set -e

echo "=========================================="
echo "   报名系统 Docker 部署脚本"
echo "=========================================="
echo ""

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 检查 Docker 是否安装
echo -e "${YELLOW}[1/6] 检查 Docker...${NC}"
if ! command -v docker &> /dev/null; then
    echo -e "${RED}✗ Docker 未安装${NC}"
    echo "  请先安装 Docker: https://docs.docker.com/get-docker/"
    exit 1
fi
echo -e "${GREEN}✓ Docker 已安装${NC}"
echo ""

# 检查 Docker Compose
echo -e "${YELLOW}[2/6] 检查 Docker Compose...${NC}"
if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo -e "${RED}✗ Docker Compose 未安装${NC}"
    echo "  请先安装 Docker Compose"
    exit 1
fi
echo -e "${GREEN}✓ Docker Compose 已安装${NC}"
echo ""

# 停止旧容器（如果存在）
echo -e "${YELLOW}[3/6] 停止旧容器...${NC}"
if docker ps -a | grep -q "registration-backend"; then
    docker stop registration-backend 2>/dev/null || true
    docker rm registration-backend 2>/dev/null || true
    echo "  已停止旧容器"
else
    echo "  没有旧容器需要停止"
fi
echo ""

# 构建镜像
echo -e "${YELLOW}[4/6] 构建 Docker 镜像...${NC}"
docker build -t registration-backend:latest ./backend
echo -e "${GREEN}✓ 镜像构建完成${NC}"
echo ""

# 运行容器
echo -e "${YELLOW}[5/6] 启动容器...${NC}"
docker run -d \
  --name registration-backend \
  --restart unless-stopped \
  -p 127.0.0.1:8080:8080 \
  -v $(pwd)/backend/uploads:/app/uploads \
  -e DATABASE_HOST=43.143.133.62 \
  -e DATABASE_PORT=3306 \
  -e DATABASE_USER=root \
  -e DATABASE_PASSWORD=hist2025 \
  -e DATABASE_DBNAME=hoj \
  -e SERVER_PORT=8080 \
  registration-backend:latest

echo -e "${GREEN}✓ 容器已启动${NC}"
echo ""

# 等待服务启动
echo -e "${YELLOW}[6/6] 等待服务启动...${NC}"
sleep 5

# 检查容器状态
if docker ps | grep -q "registration-backend"; then
    echo -e "${GREEN}✓ 容器正在运行${NC}"
    echo ""
    echo "容器信息:"
    docker ps --filter "name=registration-backend" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
    echo ""

    # 测试服务
    echo "测试服务连接:"
    if curl -s http://127.0.0.1:8080/api/competitions > /dev/null 2>&1; then
        echo -e "${GREEN}✓ 服务响应正常${NC}"
    else
        echo -e "${YELLOW}⚠ 服务可能还在启动中，请稍后测试${NC}"
    fi
else
    echo -e "${RED}✗ 容器启动失败${NC}"
    echo ""
    echo "查看日志:"
    docker logs registration-backend
    exit 1
fi

echo ""
echo "=========================================="
echo -e "${GREEN}   部署完成！${NC}"
echo "=========================================="
echo ""
echo "常用命令:"
echo "  查看日志: docker logs -f registration-backend"
echo "  停止服务: docker stop registration-backend"
echo "  启动服务: docker start registration-backend"
echo "  重启服务: docker restart registration-backend"
echo "  进入容器: docker exec -it registration-backend sh"
echo ""
echo "测试命令:"
echo "  curl http://127.0.0.1:8080/api/competitions"
echo "  curl http://bingoj.cn/registration-api/api/competitions"
echo ""
