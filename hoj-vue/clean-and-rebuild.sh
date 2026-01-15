#!/bin/bash

echo "========================================="
echo "清理 Webpack 缓存并重启开发服务器"
echo "========================================="

# 停止当前的开发服务器（如果在运行）
echo "1. 停止开发服务器..."
pkill -f "vue-cli-service serve" 2>/dev/null || true
pkill -f "npm.*serve" 2>/dev/null || true
sleep 2

# 进入 vue 目录
cd "$(dirname "$0")"

# 清理构建缓存
echo "2. 清理构建缓存..."
rm -rf dist
rm -rf node_modules/.cache
rm -rf .cache
rm -rf src/dist

# 清理 webpack 热更新缓存
echo "3. 清理 Webpack HMR 缓存..."
find . -name ".hot-cache" -type d -exec rm -rf {} + 2>/dev/null || true
find . -name ".webpack-cache" -type d -exec rm -rf {} + 2>/dev/null || true

echo "4. 清理完成！"
echo ""
echo "========================================="
echo "现在请运行: npm run serve"
echo "========================================="
