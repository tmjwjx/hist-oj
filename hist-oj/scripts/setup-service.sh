#!/bin/bash

# 在服务器上创建 systemd 服务的脚本
# 用法: 在服务器上运行此脚本

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 配置
SERVICE_NAME="hist-oj"
DEPLOY_DIR="/opt/hist-oj"
SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}创建 hist-oj systemd 服务${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# 检查是否为 root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}错误: 请使用 root 用户运行此脚本${NC}"
    exit 1
fi

# 检查部署目录
if [ ! -d "${DEPLOY_DIR}" ]; then
    echo -e "${RED}错误: 部署目录不存在: ${DEPLOY_DIR}${NC}"
    exit 1
fi

if [ ! -f "${DEPLOY_DIR}/hist-oj-server" ]; then
    echo -e "${RED}错误: 可执行文件不存在: ${DEPLOY_DIR}/hist-oj-server${NC}"
    exit 1
fi

# 创建服务文件
echo -e "${YELLOW}创建服务文件: ${SERVICE_FILE}${NC}"
cat > "${SERVICE_FILE}" << 'EOF'
[Unit]
Description=HOJ Rating Service
After=network.target mysql.service
Wants=mysql.service

[Service]
Type=simple
User=root
WorkingDirectory=/opt/hist-oj
ExecStart=/opt/hist-oj/hist-oj-server
Restart=on-failure
RestartSec=5s

# 日志配置
StandardOutput=journal
StandardError=journal
SyslogIdentifier=hist-oj

# 资源限制
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF

echo -e "${GREEN}✓ 服务文件创建完成${NC}"

# 重载 systemd
echo -e "${YELLOW}重载 systemd...${NC}"
systemctl daemon-reload
echo -e "${GREEN}✓ systemd 重载完成${NC}"

# 启动服务
echo -e "${YELLOW}启动服务...${NC}"
systemctl start ${SERVICE_NAME}
sleep 2

# 检查服务状态
if systemctl is-active --quiet ${SERVICE_NAME}; then
    echo -e "${GREEN}✓ 服务启动成功${NC}"
else
    echo -e "${RED}✗ 服务启动失败${NC}"
    systemctl status ${SERVICE_NAME} --no-pager
    exit 1
fi

# 设置开机自启
echo -e "${YELLOW}设置开机自启...${NC}"
systemctl enable ${SERVICE_NAME}
echo -e "${GREEN}✓ 开机自启设置完成${NC}"

# 显示服务状态
echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}服务创建成功！${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "服务状态:"
systemctl status ${SERVICE_NAME} --no-pager
echo ""
echo "常用命令:"
echo "  启动服务: systemctl start ${SERVICE_NAME}"
echo "  停止服务: systemctl stop ${SERVICE_NAME}"
echo "  重启服务: systemctl restart ${SERVICE_NAME}"
echo "  查看状态: systemctl status ${SERVICE_NAME}"
echo "  查看日志: journalctl -u ${SERVICE_NAME} -f"
echo ""
