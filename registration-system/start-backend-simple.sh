#!/bin/bash
# 报名系统简单启动脚本（不使用 Docker）

set -e

echo "=========================================="
echo "   报名系统启动脚本"
echo "=========================================="
echo ""

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 1. 检查 Go 环境
echo -e "${YELLOW}[1/5] 检查 Go 环境...${NC}"
if ! command -v go &> /dev/null; then
    echo -e "${RED}✗ Go 未安装${NC}"
    echo "  请先安装 Go 1.21 或更高版本"
    exit 1
fi
echo -e "${GREEN}✓ Go 已安装: $(go version)${NC}"
echo ""

# 2. 进入后端目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$SCRIPT_DIR/backend"

if [ ! -d "$BACKEND_DIR" ]; then
    echo -e "${RED}✗ 后端目录不存在: $BACKEND_DIR${NC}"
    exit 1
fi

cd "$BACKEND_DIR"
echo -e "${GREEN}✓ 工作目录: $(pwd)${NC}"
echo ""

# 3. 检查并停止旧进程
echo -e "${YELLOW}[2/5] 检查旧进程...${NC}"
OLD_PID=$(pgrep -f "go run main.go" || true)
if [ -n "$OLD_PID" ]; then
    echo "  发现旧进程 PID: $OLD_PID"
    kill $OLD_PID 2>/dev/null || true
    sleep 2
    echo -e "${GREEN}✓ 已停止旧进程${NC}"
else
    echo "  没有旧进程在运行"
fi
echo ""

# 4. 安装依赖
echo -e "${YELLOW}[3/5] 安装 Go 依赖...${NC}"
if [ -f "go.mod" ]; then
    go mod tidy
    echo -e "${GREEN}✓ 依赖安装完成${NC}"
else
    echo -e "${RED}✗ 找不到 go.mod 文件${NC}"
    exit 1
fi
echo ""

# 5. 启动服务
echo -e "${YELLOW}[4/5] 启动后端服务...${NC}"

# 创建 logs 目录
mkdir -p logs

# 后台启动服务
nohup go run main.go > logs/backend.log 2>&1 &
NEW_PID=$!

echo "  服务正在启动，PID: $NEW_PID"
echo ""

# 等待服务启动
echo -e "${YELLOW}[5/5] 等待服务启动...${NC}"
for i in {1..10}; do
    sleep 1
    if ps -p $NEW_PID > /dev/null 2>&1; then
        if netstat -tlnp 2>/dev/null | grep -q ":8080 "; then
            echo -e "${GREEN}✓ 服务启动成功！${NC}"
            echo ""
            echo "服务信息:"
            echo "  PID: $NEW_PID"
            echo "  端口: 8080"
            echo "  日志: $BACKEND_DIR/logs/backend.log"
            echo ""

            # 测试服务
            echo "测试服务连接:"
            sleep 2
            if curl -s http://127.0.0.1:8080/api/competitions > /dev/null 2>&1; then
                echo -e "${GREEN}✓ 服务响应正常${NC}"
            else
                echo -e "${YELLOW}⚠ 服务可能还在启动中${NC}"
            fi
            echo ""
            echo "=========================================="
            echo -e "${GREEN}   启动完成！${NC}"
            echo "=========================================="
            echo ""
            echo "常用命令:"
            echo "  查看日志: tail -f $BACKEND_DIR/logs/backend.log"
            echo "  停止服务: kill $NEW_PID"
            echo "  查看状态: ps -p $NEW_PID"
            echo ""
            exit 0
        fi
    else
        echo -e "${RED}✗ 服务启动失败！${NC}"
        echo ""
        echo "查看错误日志:"
        tail -20 logs/backend.log
        exit 1
    fi
done

echo -e "${YELLOW}⚠ 服务可能需要更多时间启动${NC}"
echo "  请稍后手动检查: tail -f $BACKEND_DIR/logs/backend.log"
