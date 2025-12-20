#!/bin/bash

# 快速部署脚本 - 只部署 hist-oj 或 hoj-frontend
# 用法: ./quick-deploy.sh [hist-oj|frontend|all]

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

SERVER_IP="43.143.133.62"
SERVER_USER="root"
SERVER_PASS="n208966737"
PROJECT_DIR="/Users/tianjiajie/Desktop/hoj2"

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# 部署 hist-oj
deploy_hist_oj() {
    log_info "部署 hist-oj..."

    cd "$PROJECT_DIR/hist-oj"
    docker build --platform linux/amd64 -t hist-oj:latest .

    cd "$PROJECT_DIR"
    docker save hist-oj:latest | gzip > hist-oj.tar.gz

    sshpass -p "$SERVER_PASS" scp hist-oj.tar.gz ${SERVER_USER}@${SERVER_IP}:/opt/

    sshpass -p "$SERVER_PASS" ssh -o StrictHostKeyChecking=no ${SERVER_USER}@${SERVER_IP} << 'ENDSSH'
        cd /opt
        gunzip -c hist-oj.tar.gz | docker load
        docker stop hist-oj 2>/dev/null || true
        docker rm hist-oj 2>/dev/null || true
        docker run -d \
            --name hist-oj \
            --network hoj_hoj-network \
            -p 9527:9527 \
            -v /workspace/hoj-deploy/distributed/main/hist-oj/configs:/app/configs \
            --restart unless-stopped \
            hist-oj:latest
        sleep 3
        docker ps | grep hist-oj
        curl -s http://localhost:9527/health
ENDSSH

    rm -f hist-oj.tar.gz
    log_info "✓ hist-oj 部署完成"
}

# 部署前端
deploy_frontend() {
    log_info "部署 hoj-frontend..."

    cd "$PROJECT_DIR/hoj-vue"
    docker build --platform linux/amd64 -t hoj-frontend:latest .

    cd "$PROJECT_DIR"
    docker save hoj-frontend:latest | gzip > hoj-frontend.tar.gz

    sshpass -p "$SERVER_PASS" scp hoj-frontend.tar.gz ${SERVER_USER}@${SERVER_IP}:/opt/

    sshpass -p "$SERVER_PASS" ssh -o StrictHostKeyChecking=no ${SERVER_USER}@${SERVER_IP} << 'ENDSSH'
        cd /opt
        gunzip -c hoj-frontend.tar.gz | docker load
        docker stop hoj-frontend 2>/dev/null || true
        docker rm hoj-frontend 2>/dev/null || true
        docker run -d \
            --name hoj-frontend \
            --network hoj_hoj-network \
            -p 80:80 \
            -p 443:443 \
            --restart unless-stopped \
            hoj-frontend:latest
        sleep 3
        docker ps | grep hoj-frontend
ENDSSH

    rm -f hoj-frontend.tar.gz
    log_info "✓ hoj-frontend 部署完成"
}

# 主函数
main() {
    case "${1:-all}" in
        hist-oj)
            deploy_hist_oj
            ;;
        frontend)
            deploy_frontend
            ;;
        all)
            deploy_hist_oj
            deploy_frontend
            ;;
        *)
            log_error "用法: $0 [hist-oj|frontend|all]"
            exit 1
            ;;
    esac

    log_info "=========================================="
    log_info "部署完成！访问地址："
    log_info "  - 前端: http://${SERVER_IP}"
    log_info "  - hist-oj: http://${SERVER_IP}:9527"
    log_info "=========================================="
}

main "$@"
