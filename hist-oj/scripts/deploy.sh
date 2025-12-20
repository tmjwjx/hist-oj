#!/bin/bash

# HOJ Rating 系统部署脚本
# 用法: ./deploy.sh [server-ip] [ssh-user]

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 配置
SERVER_IP=${1:-"43.143.133.62"}
SSH_USER=${2:-"root"}
DEPLOY_DIR="/opt/hist-oj"
SERVICE_NAME="hist-oj"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}HOJ Rating 系统部署脚本${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "服务器: ${SERVER_IP}"
echo "用户: ${SSH_USER}"
echo "部署目录: ${DEPLOY_DIR}"
echo ""

# 检查本地环境
echo -e "${YELLOW}[1/6] 检查本地环境...${NC}"
if ! command -v go &> /dev/null; then
    echo -e "${RED}错误: 未安装 Go${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Go 已安装: $(go version)${NC}"

# 编译项目
echo -e "${YELLOW}[2/6] 编译项目...${NC}"
cd "$(dirname "$0")/.."
echo "当前目录: $(pwd)"

# 清理旧的构建
rm -f hist-oj-server

# 编译 Linux 版本
echo "编译 Linux amd64 版本..."
GOOS=linux GOARCH=amd64 go build -o hist-oj-server ./cmd/server/main.go

if [ ! -f "hist-oj-server" ]; then
    echo -e "${RED}错误: 编译失败${NC}"
    exit 1
fi
echo -e "${GREEN}✓ 编译成功${NC}"

# 打包文件
echo -e "${YELLOW}[3/6] 打包文件...${NC}"
PACKAGE_NAME="hist-oj-$(date +%Y%m%d-%H%M%S).tar.gz"
tar -czf "/tmp/${PACKAGE_NAME}" \
    hist-oj-server \
    configs/ \
    --exclude='*.log' \
    --exclude='.git'

echo -e "${GREEN}✓ 打包完成: /tmp/${PACKAGE_NAME}${NC}"

# 上传到服务器
echo -e "${YELLOW}[4/6] 上传到服务器...${NC}"
scp "/tmp/${PACKAGE_NAME}" "${SSH_USER}@${SERVER_IP}:/tmp/"
echo -e "${GREEN}✓ 上传完成${NC}"

# 在服务器上部署
echo -e "${YELLOW}[5/6] 在服务器上部署...${NC}"
ssh "${SSH_USER}@${SERVER_IP}" << EOF
set -e

echo "创建部署目录..."
mkdir -p ${DEPLOY_DIR}
cd ${DEPLOY_DIR}

echo "备份旧版本..."
if [ -f "hist-oj-server" ]; then
    mv hist-oj-server hist-oj-server.bak.\$(date +%Y%m%d-%H%M%S)
fi

echo "解压新版本..."
tar -xzf /tmp/${PACKAGE_NAME} -C ${DEPLOY_DIR}

echo "设置执行权限..."
chmod +x ${DEPLOY_DIR}/hist-oj-server

echo "清理临时文件..."
rm -f /tmp/${PACKAGE_NAME}

echo "部署完成！"
EOF

echo -e "${GREEN}✓ 部署完成${NC}"

# 重启服务
echo -e "${YELLOW}[6/6] 重启服务...${NC}"
ssh "${SSH_USER}@${SERVER_IP}" << EOF
set -e

# 检查服务是否存在
if systemctl list-units --full -all | grep -q "${SERVICE_NAME}.service"; then
    echo "重启 ${SERVICE_NAME} 服务..."
    systemctl restart ${SERVICE_NAME}
    sleep 2
    systemctl status ${SERVICE_NAME} --no-pager
else
    echo "服务不存在，需要手动创建 systemd 服务"
    echo "请参考部署文档创建服务文件"
fi
EOF

echo -e "${GREEN}✓ 服务重启完成${NC}"

# 清理本地临时文件
rm -f "/tmp/${PACKAGE_NAME}"

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}部署成功！${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "验证部署:"
echo "  curl http://${SERVER_IP}:9527/health"
echo ""
echo "查看日志:"
echo "  ssh ${SSH_USER}@${SERVER_IP} 'journalctl -u ${SERVICE_NAME} -f'"
echo ""
