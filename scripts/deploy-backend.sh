#!/bin/bash

# HOJ2 后端服务部署脚本
# 用途：仅构建、打包、上传、部署 hist-oj 服务

set -e  # 遇到错误立即退出

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 配置变量
SERVER_IP="43.143.133.62"
SERVER_USER="root"
SERVER_PASS="n208966737"
PROJECT_DIR="/Users/tianjiajie/Desktop/hoj2"
REMOTE_DIR="/opt"

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

# 检查必要的命令
check_requirements() {
    log_info "检查必要的命令..."

    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装，请先安装 Docker"
        exit 1
    fi

    if ! command -v sshpass &> /dev/null; then
        log_error "sshpass 未安装，请先安装: brew install sshpass"
        exit 1
    fi

    log_info "✓ 所有必要命令已就绪"
}

# 构建镜像
build_backend() {
    log_info "开始构建 hist-oj Docker 镜像..."

    cd "$PROJECT_DIR/hist-oj"

    log_info "构建 hist-oj 镜像..."
    docker build --platform linux/amd64 -t hist-oj:latest . || {
        log_error "hist-oj 镜像构建失败"
        exit 1
    }
    log_info "✓ hist-oj 镜像构建成功"

    cd "$PROJECT_DIR"
}

# 保存镜像
save_backend() {
    log_info "保存 hist-oj Docker 镜像..."

    cd "$PROJECT_DIR"

    log_info "保存 hist-oj 镜像..."
    docker save hist-oj:latest | gzip > hist-oj.tar.gz || {
        log_error "hist-oj 镜像保存失败"
        exit 1
    }
    log_info "✓ hist-oj 镜像已保存 ($(du -h hist-oj.tar.gz | cut -f1))"
}

# 上传镜像到服务器
upload_backend() {
    log_info "上传 hist-oj 镜像到服务器..."

    cd "$PROJECT_DIR"

    log_info "上传 hist-oj 镜像..."
    sshpass -p "$SERVER_PASS" scp hist-oj.tar.gz ${SERVER_USER}@${SERVER_IP}:${REMOTE_DIR}/ || {
        log_error "hist-oj 镜像上传失败"
        exit 1
    }
    log_info "✓ hist-oj 镜像上传成功"
}

# 在服务器上部署
deploy_backend_on_server() {
    log_info "在服务器上部署 hist-oj 服务..."

    sshpass -p "$SERVER_PASS" ssh -o StrictHostKeyChecking=no ${SERVER_USER}@${SERVER_IP} << 'ENDSSH'
        set -e

        echo "[INFO] 加载 hist-oj Docker 镜像..."
        cd /opt

        # 加载 hist-oj 镜像
        echo "[INFO] 加载 hist-oj 镜像..."
        gunzip -c hist-oj.tar.gz | docker load

        echo "[INFO] 停止并删除旧容器..."
        docker stop hist-oj 2>/dev/null || true
        docker rm hist-oj 2>/dev/null || true

        echo "[INFO] 启动 hist-oj 容器..."
        docker run -d \
            --name hist-oj \
            --network hoj_hoj-network \
            -p 9527:9527 \
            -v /workspace/hoj-deploy/distributed/main/hist-oj/configs:/app/configs \
            --restart unless-stopped \
            hist-oj:latest

        echo "[INFO] 等待服务启动..."
        sleep 5

        echo "[INFO] 验证服务状态..."
        docker ps | grep hist-oj

        echo "[INFO] 测试 hist-oj 健康检查..."
        curl -s http://localhost:9527/health || echo "健康检查失败"

        echo "[INFO] 测试前端访问 hist-oj..."
        docker exec hoj-frontend curl -s http://hist-oj:9527/health || echo "前端访问 hist-oj 失败"

        echo "[INFO] 清理服务器上的临时文件..."
        rm -f /opt/hist-oj.tar.gz

        echo "[INFO] hist-oj 部署完成！"
ENDSSH

    if [ $? -eq 0 ]; then
        log_info "✓ 服务器部署成功"
    else
        log_error "服务器部署失败"
        exit 1
    fi
}

# 清理本地临时文件
cleanup() {
    log_info "清理本地临时文件..."
    cd "$PROJECT_DIR"
    rm -f hist-oj.tar.gz
    log_info "✓ 清理完成"
}

# 显示部署结果
show_result() {
    log_info "=========================================="
    log_info "hist-oj 后端部署完成！"
    log_info "=========================================="
    log_info "服务访问地址："
    log_info "  - hist-oj API: http://${SERVER_IP}:9527"
    log_info ""
    log_info "验证命令："
    log_info "  curl http://${SERVER_IP}:9527/health"
    log_info "  curl http://${SERVER_IP}:9527/api/rating/user/rating/<uid>"
    log_info "=========================================="
}

# 主函数
main() {
    log_info "=========================================="
    log_info "HOJ2 后端服务部署"
    log_info "=========================================="

    # 执行部署流程
    check_requirements
    build_backend
    save_backend
    upload_backend
    deploy_backend_on_server
    cleanup
    show_result

    log_info "✓ 所有步骤完成！"
}

# 执行主函数
main "$@"
