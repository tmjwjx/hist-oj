#!/bin/bash

# 报名系统快速启动脚本
# 用于启动Go后端服务和检查HOJ前端状态

echo "======================================================"
echo "           报名系统启动脚本"
echo "======================================================"
echo ""

# 检查Go是否安装
if ! command -v go &> /dev/null; then
    echo "❌ 错误: Go未安装,请先安装Go 1.21或更高版本"
    exit 1
fi

# 进入后端目录
cd "$(dirname "$0")/backend"

echo "📦 安装Go依赖..."
go mod tidy

if [ $? -ne 0 ]; then
    echo "❌ 错误: Go依赖安装失败"
    exit 1
fi

echo ""
echo "✅ Go依赖安装成功"
echo ""

# 检查数据库连接
echo "🔍 检查数据库连接..."
mysql -h 43.143.133.62 -u root -phist2025 hoj -e "SELECT 1;" &> /dev/null

if [ $? -ne 0 ]; then
    echo "⚠️  警告: 无法连接到数据库 43.143.133.62:3306"
    echo "   请检查数据库是否正常运行"
    echo ""
fi

echo ""
echo "======================================================"
echo "🚀 启动Go后端服务 (端口: 8080)"
echo "======================================================"
echo ""

# 启动Go后端
go run main.go
