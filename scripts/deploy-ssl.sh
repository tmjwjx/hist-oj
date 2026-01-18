#!/bin/bash

# SSL 证书快速部署脚本
# 用途：仅部署 SSL 证书和 Nginx 配置（不重新构建镜像）

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 配置变量
SERVER_IP="43.143.133.62"
SERVER_USER="root"
SERVER_PASS="n208966737"
PROJECT_DIR="/Users/zhuangqingjia/vscode/histoj/hist-oj"

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查证书文件
check_certs() {
    log_info "检查 SSL 证书文件..."

    if [ ! -f "scripts/nginx/ssl/bingoj.cn.crt" ]; then
        log_error "证书文件不存在: scripts/nginx/ssl/bingoj.cn.crt"
        exit 1
    fi

    if [ ! -f "scripts/nginx/ssl/bingoj.cn.key" ]; then
        log_error "私钥文件不存在: scripts/nginx/ssl/bingoj.cn.key"
        exit 1
    fi

    if [ ! -f "scripts/nginx/bingoj.conf" ]; then
        log_error "Nginx 配置文件不存在: scripts/nginx/bingoj.conf"
        exit 1
    fi

    log_info "✓ 所有必需文件都存在"
}

# 上传证书
upload_certs() {
    log_info "上传 SSL 证书到服务器..."

    cd "$PROJECT_DIR"

    # 创建服务器上的 SSL 目录
    sshpass -p "$SERVER_PASS" ssh -o StrictHostKeyChecking=no ${SERVER_USER}@${SERVER_IP} << ENDSSH
        mkdir -p /etc/nginx/ssl
        chmod 700 /etc/nginx/ssl
ENDSSH

    # 上传证书文件
    log_info "上传证书文件..."
    sshpass -p "$SERVER_PASS" scp scripts/nginx/ssl/bingoj.cn.crt ${SERVER_USER}@${SERVER_IP}:/etc/nginx/ssl/
    sshpass -p "$SERVER_PASS" scp scripts/nginx/ssl/bingoj.cn.key ${SERVER_USER}@${SERVER_IP}:/etc/nginx/ssl/

    # 设置证书文件权限
    sshpass -p "$SERVER_PASS" ssh -o StrictHostKeyChecking=no ${SERVER_USER}@${SERVER_IP} << ENDSSH
        chmod 644 /etc/nginx/ssl/bingoj.cn.crt
        chmod 600 /etc/nginx/ssl/bingoj.cn.key
        chown root:root /etc/nginx/ssl/*
ENDSSH

    # 上传 Nginx 配置文件
    log_info "上传 Nginx 配置..."
    sshpass -p "$SERVER_PASS" scp scripts/nginx/bingoj.conf ${SERVER_USER}@${SERVER_IP}:/etc/nginx/conf.d/

    log_info "✓ 证书和配置上传成功"
}

# 配置 Nginx
configure_nginx() {
    log_info "配置并启动 Nginx..."

    sshpass -p "$SERVER_PASS" ssh -o StrictHostKeyChecking=no ${SERVER_USER}@${SERVER_IP} << 'ENDSSH'
        set -e

        # 备份旧配置
        if [ -f /etc/nginx/conf.d/bingoj.conf ] && [ ! -f /etc/nginx/conf.d/bingoj.conf.bak ]; then
            cp /etc/nginx/conf.d/bingoj.conf /etc/nginx/conf.d/bingoj.conf.bak
            echo "[INFO] 已备份旧配置"
        fi

        # 测试配置
        echo "[INFO] 测试 Nginx 配置..."
        nginx -t

        # 启动/重载 Nginx
        if systemctl is-active --quiet nginx; then
            echo "[INFO] 重新加载 Nginx..."
            systemctl reload nginx
        else
            echo "[INFO] 启动 Nginx..."
            systemctl start nginx
            systemctl enable nginx
        fi

        # 检查状态
        systemctl status nginx --no-pager -l

        # 验证 443 端口
        sleep 2
        if netstat -tlnp | grep -q ":443 "; then
            echo "[INFO] ✓ Nginx 正在监听 443 端口"
        else
            echo "[WARN] Nginx 未监听 443 端口"
        fi
ENDSSH

    if [ $? -eq 0 ]; then
        log_info "✓ Nginx 配置成功"
    else
        log_error "Nginx 配置失败"
        exit 1
    fi
}

# 测试 HTTPS
test_https() {
    log_info "测试 HTTPS 连接..."

    # 测试 HTTPS 访问
    HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" https://bingoj.cn 2>/dev/null || echo "000")

    if [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "301" ] || [ "$HTTP_CODE" = "302" ]; then
        log_info "✓ HTTPS 访问成功 (HTTP $HTTP_CODE)"
    else
        log_warn "HTTPS 返回状态码: $HTTP_CODE"
    fi

    # 测试 HTTP 重定向到 HTTPS
    HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" http://bingoj.cn 2>/dev/null || echo "000")

    if [ "$HTTP_CODE" = "301" ] || [ "$HTTP_CODE" = "302" ]; then
        log_info "✓ HTTP 自动重定向到 HTTPS (HTTP $HTTP_CODE)"
    else
        log_warn "HTTP 状态码: $HTTP_CODE (可能未配置重定向)"
    fi

    # 显示证书信息
    log_info "SSL 证书信息："
    echo | openssl s_client -connect bingoj.cn:443 -servername bingoj.cn 2>/dev/null | openssl x509 -noout -subject -dates 2>/dev/null || echo "无法获取证书信息"
}

# 显示结果
show_result() {
    log_info "=========================================="
    log_info "SSL 证书部署完成！"
    log_info "=========================================="
    log_info ""
    log_info "访问地址："
    log_info "  - HTTPS: https://bingoj.cn"
    log_info "  - HTTP: http://bingoj.cn (自动重定向到 HTTPS)"
    log_info ""
    log_info "班级签到功能（支持摄像头）："
    log_info "  - 学生端: https://bingoj.cn/classroom/student"
    log_info ""
    log_info "验证命令："
    log_info "  curl -I https://bingoj.cn"
    log_info "  curl -I http://bingoj.cn"
    log_info "  echo | openssl s_client -connect bingoj.cn:443 -servername bingoj.cn"
    log_info ""
    log_info "测试摄像头权限："
    log_info "  在浏览器访问 https://bingoj.cn/classroom/student"
    log_info "  打开控制台执行："
    log_info "  navigator.mediaDevices.getUserMedia({ video: true })"
    log_info ""
    log_info "证书有效期：90天（请在到期前续期）"
    log_info "=========================================="
}

# 主函数
main() {
    cd "$PROJECT_DIR"

    log_info "=========================================="
    log_info "SSL 证书快速部署"
    log_info "=========================================="

    check_certs
    upload_certs
    configure_nginx
    test_https
    show_result

    log_info "✓ SSL 部署完成！"
}

main "$@"
