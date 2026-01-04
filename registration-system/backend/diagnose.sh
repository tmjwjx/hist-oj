#!/bin/bash
# 快速诊断脚本 - 找出为什么服务无法启动

echo "=========================================="
echo "   报名系统故障诊断"
echo "=========================================="
echo ""

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 1. 检查 main.go 是否存在
echo -e "${YELLOW}[1] 检查项目文件${NC}"
if [ -f "main.go" ]; then
    echo -e "${GREEN}✓ main.go 存在${NC}"
else
    echo -e "${RED}✗ main.go 不存在${NC}"
    echo ""
    echo "请先确认是否在正确的目录"
    echo "当前目录: $(pwd)"
    echo "目录内容:"
    ls -la
    exit 1
fi
echo ""

# 2. 检查端口占用
echo -e "${YELLOW}[2] 检查端口占用${NC}"
PORT_8080_PROCESS=$(netstat -tlnp 2>/dev/null | grep ":8080 " || echo "")
if [ -n "$PORT_8080_PROCESS" ]; then
    echo -e "${RED}✗ 端口 8080 已被占用${NC}"
    echo "$PORT_8080_PROCESS"
    echo ""
    echo "解决方法:"
    echo "  方法1: 停止占用端口的进程"
    PID=$(echo "$PORT_8080_PROCESS" | awk '{print $7}' | cut -d'/' -f1)
    echo "  sudo kill $PID"
    echo ""
    echo "  方法2: 使用其他端口"
    echo "  修改 main.go 第 1058 行的端口号"
else
    echo -e "${GREEN}✓ 端口 8080 未被占用${NC}"
fi
echo ""

# 3. 检查 Go 环境
echo -e "${YELLOW}[3] 检查 Go 环境${NC}"
if command -v go &> /dev/null; then
    echo -e "${GREEN}✓ Go 已安装: $(go version)${NC}"
else
    echo -e "${RED}✗ Go 未安装${NC}"
    exit 1
fi
echo ""

# 4. 检查依赖
echo -e "${YELLOW}[4] 检查 Go 依赖${NC}"
if [ -f "go.mod" ]; then
    echo "✓ go.mod 存在"

    # 检查依赖是否已下载
    if [ -d "vendor" ] || go list -m all &> /dev/null; then
        echo -e "${GREEN}✓ 依赖已安装${NC}"
    else
        echo -e "${YELLOW}⚠ 依赖未完全安装${NC}"
        echo "  正在安装依赖..."
        go mod tidy
    fi
else
    echo -e "${RED}✗ go.mod 不存在${NC}"
    echo "  正在初始化..."
    go mod init registration-system
    go get github.com/go-sql-driver/mysql
    go get github.com/gorilla/mux
fi
echo ""

# 5. 测试数据库连接
echo -e "${YELLOW}[5] 测试数据库连接${NC}"
DB_OK=0
if command -v mysql &> /dev/null; then
    if mysql -h43.143.133.62 -uroot -phist2025 -e "USE hoj; SHOW TABLES;" &> /dev/null; then
        echo -e "${GREEN}✓ MySQL 连接正常${NC}"
        DB_OK=1
    else
        echo -e "${RED}✗ MySQL 连接失败${NC}"
        echo ""
        echo "可能的原因:"
        echo "  1. MySQL 服务未运行"
        echo "  2. 数据库地址/端口/用户名/密码错误"
        echo "  3. 数据库 'hoj' 不存在"
        echo "  4. 网络连接问题"
    fi
else
    echo -e "${YELLOW}⚠ mysql 客户端未安装，跳过测试${NC}"
fi
echo ""

# 6. 尝试编译
echo -e "${YELLOW}[6] 尝试编译程序${NC}"
COMPILE_OUTPUT=$(go build -o registration-server main.go 2>&1)
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ 编译成功${NC}"
    rm -f registration-server
else
    echo -e "${RED}✗ 编译失败${NC}"
    echo "$COMPILE_OUTPUT"
    exit 1
fi
echo ""

# 总结
echo "=========================================="
echo "   诊断总结"
echo "=========================================="
echo ""

if [ -n "$PORT_8080_PROCESS" ]; then
    echo -e "${RED}主要问题: 端口被占用${NC}"
    echo ""
    echo "建议操作:"
    PID=$(echo "$PORT_8080_PROCESS" | awk '{print $7}' | cut -d'/' -f1)
    echo "  1. 停止占用进程: sudo kill $PID"
    echo "  2. 重新启动服务"
elif [ $DB_OK -eq 0 ]; then
    echo -e "${RED}主要问题: 数据库连接失败${NC}"
    echo ""
    echo "建议操作:"
    echo "  1. 检查数据库是否运行"
    echo "  2. 验证数据库配置: main.go 第 96 行"
    echo "  3. 测试连接: mysql -h43.143.133.62 -uroot -phist2025"
else
    echo -e "${GREEN}✓ 所有检查通过${NC}"
    echo ""
    echo "可以尝试启动服务:"
    echo "  go run main.go"
    echo ""
    echo "或使用启动脚本:"
    echo "  ./start-with-logs.sh"
fi
echo ""
