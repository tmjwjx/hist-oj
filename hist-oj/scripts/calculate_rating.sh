#!/bin/bash

# Rating 计算脚本
# 用法: ./calculate_rating.sh <contest_id> [admin_token]

CONTEST_ID=$1
ADMIN_TOKEN=$2
BASE_URL="http://localhost:9527"

if [ -z "$CONTEST_ID" ]; then
    echo "用法: $0 <contest_id> [admin_token]"
    echo "示例: $0 1002"
    echo ""
    echo "如果不提供 admin_token，将尝试不带认证的请求（仅用于测试环境）"
    exit 1
fi

echo "正在触发比赛 $CONTEST_ID 的 Rating 计算..."

if [ -z "$ADMIN_TOKEN" ]; then
    # 不带 token 的请求
    RESPONSE=$(curl -s -X POST "$BASE_URL/api/rating/calculate/$CONTEST_ID")
else
    # 带 token 的请求
    RESPONSE=$(curl -s -X POST "$BASE_URL/api/rating/calculate/$CONTEST_ID" \
        -H "Authorization: Bearer $ADMIN_TOKEN")
fi

echo "响应: $RESPONSE"

# 检查响应
if echo "$RESPONSE" | grep -q '"code":200'; then
    echo "✅ Rating 计算成功！"
    exit 0
elif echo "$RESPONSE" | grep -q '"code":401'; then
    echo "❌ 需要管理员权限。请提供 admin_token 或在浏览器中以管理员身份登录后执行。"
    echo ""
    echo "浏览器执行方法："
    echo "1. 访问 http://localhost:8066 并登录管理员账号"
    echo "2. 打开开发者工具（F12），切换到 Console"
    echo "3. 执行以下代码："
    echo ""
    echo "fetch('/rating-api/rating/calculate/$CONTEST_ID', {"
    echo "  method: 'POST',"
    echo "  credentials: 'include'"
    echo "}).then(r => r.json()).then(console.log)"
    exit 1
else
    echo "❌ 计算失败，请检查错误信息"
    exit 1
fi
