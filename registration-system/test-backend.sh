#!/bin/bash

echo "=========================================="
echo "  报名系统Go后端测试脚本"
echo "=========================================="
echo ""

# 检查Go后端是否运行
echo "1. 检查Go后端服务状态..."
if curl -s http://localhost:8080/api/competitions > /dev/null 2>&1; then
    echo "   ✅ Go后端正在运行 (端口8080)"

    echo ""
    echo "2. 测试API连接..."
    response=$(curl -s http://localhost:8080/api/competitions)
    echo "   API响应: $response"

    if [ $? -eq 0 ]; then
        echo "   ✅ API连接成功"
    fi
else
    echo "   ❌ Go后端未运行 (端口8080)"
    echo ""
    echo "请先启动Go后端服务："
    echo "  cd /Users/zhuangqingjia/vscode/histoj/hist-oj/registration-system/backend"
    echo "  go run main.go"
    echo ""
    exit 1
fi

echo ""
echo "3. 检查数据库连接..."
mysql -h 43.143.133.62 -u root -phist2025 hoj -e "SHOW TABLES LIKE 'histcontest_register%';" 2>/dev/null

if [ $? -eq 0 ]; then
    echo "   ✅ 数据库连接成功"
else
    echo "   ⚠️  无法连接到数据库或表未创建"
fi

echo ""
echo "=========================================="
echo "  测试完成"
echo "=========================================="
