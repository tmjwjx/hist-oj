#!/bin/bash

# HOJ Rating & Classroom & Plagiarism 系统部署脚本
# 功能: Rating 计算 + 班级管理 + 代码查重
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
echo -e "${GREEN}HOJ Rating & Classroom 系统部署${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "服务器: ${SERVER_IP}"
echo "用户: ${SSH_USER}"
echo "部署目录: ${DEPLOY_DIR}"
echo ""
echo -e "${YELLOW}功能模块:${NC}"
echo "  - Rating 计算系统"
echo "  - 班级管理系统（班级、签到、题库、作业、资料库、随机选人、即时通讯）"
echo "  - 代码查重系统（基于 sim 工具）"
echo ""
echo -e "${YELLOW}部署前检查:${NC}"
echo "  1. 确保 configs/config.yaml 已正确配置数据库连接"
echo ""

# 检查本地环境
echo -e "${YELLOW}[1/7] 检查本地环境...${NC}"
if ! command -v go &> /dev/null; then
    echo -e "${RED}错误: 未安装 Go${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Go 已安装: $(go version)${NC}"

# 编译项目
echo -e "${YELLOW}[2/7] 编译项目...${NC}"
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
echo -e "${YELLOW}[3/7] 打包文件...${NC}"
PACKAGE_NAME="hist-oj-$(date +%Y%m%d-%H%M%S).tar.gz"
tar -czf "/tmp/${PACKAGE_NAME}" \
    hist-oj-server \
    configs/

echo -e "${GREEN}✓ 打包完成: /tmp/${PACKAGE_NAME}${NC}"

# 上传到服务器
echo -e "${YELLOW}[4/7] 上传到服务器...${NC}"
if ! scp "/tmp/${PACKAGE_NAME}" "${SSH_USER}@${SERVER_IP}:/tmp/"; then
    echo -e "${RED}错误: 上传失败，请检查网络连接和 SSH 配置${NC}"
    exit 1
fi
echo -e "${GREEN}✓ 上传完成${NC}"

# 在服务器上部署
echo -e "${YELLOW}[5/7] 在服务器上部署...${NC}"
if ! ssh "${SSH_USER}@${SERVER_IP}" << EOF
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

echo "✓ 部署完成！"
EOF
then
    echo -e "${RED}错误: 服务器部署失败${NC}"
    exit 1
fi
echo -e "${GREEN}✓ 部署完成${NC}"

# 安装 sim 查重工具
echo -e "${YELLOW}[6/7] 安装 sim 查重工具...${NC}"
if ! ssh "${SSH_USER}@${SERVER_IP}" << 'EOF'
set -e

# 颜色定义（用于子 shell）
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "检查 sim 工具是否已安装..."
if command -v sim_c &> /dev/null; then
    echo -e "${GREEN}✓ sim 工具已存在${NC}"
    sim_c --help | head -3
    exit 0
fi

echo -e "${YELLOW}sim 工具未安装，开始安装...${NC}"

# 检测系统类型并安装依赖
if [ -f /etc/redhat-release ]; then
    echo "检测到 CentOS/RHEL 系统"
    PKG_MANAGER="yum"
    PKG_INSTALL="yum install -y"
elif [ -f /etc/debian_version ]; then
    echo "检测到 Debian/Ubuntu 系统"
    PKG_MANAGER="apt-get"
    PKG_INSTALL="apt-get install -y"
else
    echo -e "${RED}错误: 不支持的操作系统${NC}"
    exit 1
fi

echo "安装编译依赖（gcc, make, flex, bison, git）..."
if ! $PKG_INSTALL gcc make flex bison git; then
    echo -e "${RED}错误: 依赖安装失败${NC}"
    exit 1
fi
echo -e "${GREEN}✓ 依赖安装完成${NC}"

# 克隆 sim 源码
SIM_DIR="/tmp/sim-install-$$"
echo "克隆 sim 源码到 ${SIM_DIR}..."
if ! git clone git@github.com:aspxcor/Code-Repetition-Check-sim.git "$SIM_DIR" 2>/dev/null; then
    echo "SSH 克隆失败，尝试 HTTPS 克隆..."
    if ! git clone https://github.com/aspxcor/Code-Repetition-Check-sim.git "$SIM_DIR"; then
        echo -e "${RED}错误: sim 源码克隆失败${NC}"
        rm -rf "$SIM_DIR"
        exit 1
    fi
fi

cd "${SIM_DIR}/sim_src"

echo "修复 Makefile 配置..."
# 使用 awk 修复 Makefile:
# 1. 替换 /home/dick 为 /usr/local
# 2. 注释掉 MSDOS 配置段，避免覆盖 UNIX 配置
# 3. 添加 -std=gnu89 以兼容旧式 C 代码
awk '
BEGIN { in_msdos_section = 0 }
{
  # 替换 /home/dick 为 /usr/local
  gsub(/\/home\/dick/, "/usr/local")

  # 检测并注释 MSDOS 段
  if (/^#.*For MSDOS/) {
    in_msdos_section = 1
  }
  if (in_msdos_section) {
    print "#" $0
    # 检测 MSDOS 段结束(遇到下一个大段注释或空行后的一般编译配置)
    if (/^# General/) {
      in_msdos_section = 0
    }
    next
  }

  # 修改 CC 定义，添加 -std=gnu89 标志以兼容旧式 C 代码
  if (/^CC[[:space:]]*=[[:space:]]*gcc/) {
    print "CC = gcc -std=gnu89 -D$(SYSTEM) -D$(SUBSYSTEM)"
    next
  }

  print
}' Makefile > Makefile.new && mv Makefile.new Makefile

echo "编译 sim 工具（仅安装 sim_c 和 sim_java）..."
# 只编译我们需要的工具
if ! make sim_c sim_java; then
    echo -e "${RED}错误: sim 编译失败${NC}"
    echo "查看相关配置行..."
    grep -n "BINDIR\|MAN1DIR" Makefile | head -10
    cd /
    rm -rf "$SIM_DIR"
    exit 1
fi

# 手动安装到系统目录
echo "安装 sim 工具到 /usr/local/bin..."
cp -f sim_c sim_java /usr/local/bin/ 2>/dev/null || cp -f sim_c sim_java /usr/bin/
chmod +x /usr/local/bin/sim_c /usr/local/bin/sim_java 2>/dev/null || chmod +x /usr/bin/sim_c /usr/bin/sim_java
echo -e "${GREEN}✓ sim 编译完成${NC}"

# 验证安装
echo "验证 sim 工具安装..."
if ! command -v sim_c &> /dev/null; then
    echo -e "${RED}错误: sim_c 未找到，安装失败${NC}"
    cd /
    rm -rf "$SIM_DIR"
    exit 1
fi

if ! command -v sim_java &> /dev/null; then
    echo -e "${RED}错误: sim_java 未找到，安装失败${NC}"
    cd /
    rm -rf "$SIM_DIR"
    exit 1
fi

# 清理临时文件
cd /
rm -rf "$SIM_DIR"

echo -e "${GREEN}✓ sim 工具安装完成${NC}"
echo "已安装的 sim 工具:"
echo "  - sim_c: $(which sim_c)"
echo "  - sim_java: $(which sim_java)"
echo "  - sim_pasc: $(which sim_pasc)"

# 测试 sim 工具
echo ""
echo "测试 sim 工具..."
echo "  sim_c 版本: $(sim_c --help 2>&1 | head -1 || echo '无法获取版本')"
EOF
then
    echo -e "${RED}错误: sim 工具安装失败${NC}"
    echo -e "${YELLOW}建议: 请手动在服务器上安装 sim 工具${NC}"
    echo "  1. SSH 登录服务器: ssh ${SSH_USER}@${SERVER_IP}"
    echo "  2. 执行安装命令:"
    echo "     git clone git@github.com:aspxcor/Code-Repetition-Check-sim.git /tmp/sim"
    echo "     cd /tmp/sim/sim_src && make install"
    exit 1
fi
echo -e "${GREEN}✓ sim 工具安装完成${NC}"

# 重启服务
echo -e "${YELLOW}[7/7] 重启服务...${NC}"
if ! ssh "${SSH_USER}@${SERVER_IP}" << EOF
set -e

# 检查服务是否存在
if systemctl list-units --full -all | grep -q "${SERVICE_NAME}.service"; then
    echo "重启 ${SERVICE_NAME} 服务..."
    systemctl restart ${SERVICE_NAME}
    sleep 3

    # 检查服务状态
    if systemctl is-active --quiet ${SERVICE_NAME}; then
        echo -e "${GREEN}✓ 服务启动成功${NC}"
    else
        echo -e "${RED}错误: 服务启动失败${NC}"
        systemctl status ${SERVICE_NAME} --no-pager
        exit 1
    fi

    systemctl status ${SERVICE_NAME} --no-pager | head -10
else
    echo "服务不存在，需要手动创建 systemd 服务"
    echo "请参考部署文档创建服务文件"
fi
EOF
then
    echo -e "${RED}错误: 服务重启失败${NC}"
    exit 1
fi
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
echo "验证 sim 工具:"
echo "  ssh ${SSH_USER}@${SERVER_IP} 'sim_c --help'"
echo ""
echo "查看日志:"
echo "  ssh ${SSH_USER}@${SERVER_IP} 'journalctl -u ${SERVICE_NAME} -f'"
echo ""
echo -e "${YELLOW}代码查重功能使用:${NC}"
echo "  1. 比赛结束后，管理员进入比赛详情页，点击'代码查重'标签"
echo "  2. 为每个题目设置查重率阈值，保存后开始查重"
echo ""
