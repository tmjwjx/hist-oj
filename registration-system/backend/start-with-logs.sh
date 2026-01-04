#!/bin/bash
# 启动报名系统后端（带详细日志）

echo "=========================================="
echo "   启动报名系统后端服务"
echo "=========================================="
echo ""

# 检查当前目录
echo "[1] 当前工作目录:"
pwd
echo ""

# 检查文件是否存在
echo "[2] 检查必要文件:"
if [ ! -f "main.go" ]; then
    echo "❌ 错误: main.go 不存在"
    echo ""
    echo "当前目录内容:"
    ls -la
    exit 1
fi
echo "✓ main.go 存在"
echo ""

# 检查 go.mod
if [ ! -f "go.mod" ]; then
    echo "⚠ 警告: go.mod 不存在"
    echo "  尝试初始化 Go module..."
    go mod init registration-system
fi
echo "✓ go.mod 存在"
echo ""

# 检查端口占用
echo "[3] 检查端口 8080:"
if netstat -tlnp 2>/dev/null | grep -q ":8080 "; then
    echo "❌ 错误: 端口 8080 已被占用"
    echo ""
    echo "占用端口的进程:"
    netstat -tlnp 2>/dev/null | grep ":8080 "
    echo ""
    echo "请先停止占用端口的进程:"
    echo "  kill <PID>"
    exit 1
fi
echo "✓ 端口 8080 可用"
echo ""

# 检查数据库连接
echo "[4] 测试数据库连接:"
if command -v mysql &> /dev/null; then
    if mysql -h43.143.133.62 -uroot -phist2025 -e "SELECT 1;" &> /dev/null; then
        echo "✓ MySQL 数据库连接正常"
    else
        echo "⚠ 警告: 无法连接到 MySQL 数据库"
        echo "  服务可能会在启动时失败"
    fi
else
    echo "⚠ mysql 客户端未安装，跳过数据库连接测试"
fi
echo ""

# 安装依赖
echo "[5] 安装 Go 依赖:"
go mod tidy
if [ $? -eq 0 ]; then
    echo "✓ 依赖安装成功"
else
    echo "❌ 依赖安装失败"
    exit 1
fi
echo ""

# 创建必要的目录
echo "[6] 创建必要的目录:"
mkdir -p uploads logs
echo "✓ 目录创建完成"
echo ""

# 编译程序
echo "[7] 编译程序:"
go build -o registration-server main.go
if [ $? -eq 0 ]; then
    echo "✓ 编译成功"
else
    echo "❌ 编译失败"
    echo "  请检查代码错误"
    exit 1
fi
echo ""

# 启动服务
echo "[8] 启动服务:"
echo "=========================================="
./registration-server 2>&1 | tee logs/startup.log
STARTUP_EXIT_CODE=${PIPESTATUS[0]}

echo ""
echo "=========================================="
echo "服务已停止，退出代码: $STARTUP_EXIT_CODE"
echo "=========================================="
echo ""

if [ $STARTUP_EXIT_CODE -ne 0 ]; then
    echo "❌ 服务启动失败"
    echo ""
    echo "查看详细日志:"
    echo "  cat logs/startup.log"
    echo ""
    echo "常见失败原因:"
    echo "  1. 端口被占用"
    echo "  2. 数据库连接失败"
    echo "  3. 权限不足"
    echo "  4. 依赖包缺失"
    exit 1
fi
