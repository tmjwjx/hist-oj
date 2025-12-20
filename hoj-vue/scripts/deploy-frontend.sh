#!/bin/bash

# 前端部署脚本
# 用法: ./deploy-frontend.sh [server-ip] [ssh-user]

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 配置
SERVER_IP=${1:-"43.143.133.62"}
SSH_USER=${2:-"root"}
WEB_ROOT="/var/www/hoj"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}HOJ 前端部署脚本${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "服务器: ${SERVER_IP}"
echo "用户: ${SSH_USER}"
echo "Web 根目录: ${WEB_ROOT}"
echo ""

# 检查本地环境
echo -e "${YELLOW}[1/5] 检查本地环境...${NC}"
if ! command -v node &> /dev/null; then
    echo -e "${RED}错误: 未安装 Node.js${NC}"
    exit 1
fi
if ! command -v npm &> /dev/null; then
    echo -e "${RED}错误: 未安装 npm${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Node.js 已安装: $(node -v)${NC}"
echo -e "${GREEN}✓ npm 已安装: $(npm -v)${NC}"

# 进入项目目录
cd "$(dirname "$0")/.."
echo "当前目录: $(pwd)"

# 安装依赖
echo -e "${YELLOW}[2/5] 安装依赖...${NC}"
if [ ! -d "node_modules" ]; then
    npm install
else
    echo "依赖已安装，跳过"
fi
echo -e "${GREEN}✓ 依赖安装完成${NC}"

# 构建项目
echo -e "${YELLOW}[3/5] 构建项目...${NC}"
npm run build

if [ ! -d "dist" ]; then
    echo -e "${RED}错误: 构建失败，dist 目录不存在${NC}"
    exit 1
fi
echo -e "${GREEN}✓ 构建完成${NC}"

# 打包文件
echo -e "${YELLOW}[4/5] 打包文件...${NC}"
PACKAGE_NAME="hoj-frontend-$(date +%Y%m%d-%H%M%S).tar.gz"
cd dist
tar -czf "/tmp/${PACKAGE_NAME}" .
cd ..
echo -e "${GREEN}✓ 打包完成: /tmp/${PACKAGE_NAME}${NC}"

# 上传并部署
echo -e "${YELLOW}[5/5] 上传并部署到服务器...${NC}"
scp "/tmp/${PACKAGE_NAME}" "${SSH_USER}@${SERVER_IP}:/tmp/"

ssh "${SSH_USER}@${SERVER_IP}" << EOF
set -e

echo "创建 Web 根目录..."
mkdir -p ${WEB_ROOT}

echo "备份旧版本..."
if [ -d "${WEB_ROOT}" ] && [ "\$(ls -A ${WEB_ROOT})" ]; then
    BACKUP_DIR="${WEB_ROOT}.bak.\$(date +%Y%m%d-%H%M%S)"
    echo "备份到: \${BACKUP_DIR}"
    cp -r ${WEB_ROOT} \${BACKUP_DIR}
fi

echo "清空 Web 根目录..."
rm -rf ${WEB_ROOT}/*

echo "解压新版本..."
tar -xzf /tmp/${PACKAGE_NAME} -C ${WEB_ROOT}

echo "设置权限..."
chown -R www-data:www-data ${WEB_ROOT}
chmod -R 755 ${WEB_ROOT}

echo "清理临时文件..."
rm -f /tmp/${PACKAGE_NAME}

echo "重启 Nginx..."
if command -v nginx &> /dev/null; then
    nginx -t && systemctl reload nginx
    echo "Nginx 重启完成"
else
    echo "警告: Nginx 未安装，请手动配置 Web 服务器"
fi

echo "部署完成！"
EOF

echo -e "${GREEN}✓ 部署完成${NC}"

# 清理本地临时文件
rm -f "/tmp/${PACKAGE_NAME}"

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}前端部署成功！${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "访问地址:"
echo "  http://${SERVER_IP}/"
echo ""
echo "如果无法访问，请检查:"
echo "  1. Nginx 是否正确配置"
echo "  2. 防火墙是否开放 80 端口"
echo "  3. 查看 Nginx 日志: tail -f /var/log/nginx/error.log"
echo ""
