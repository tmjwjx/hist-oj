#!/bin/bash

# HOJ2 + 报名系统 + 代码对战 + 班级管理系统自动化部署脚本
# 用途：一键构建、打包、上传、部署 hist-oj、hoj-frontend
#       hist-oj 已包含：报名系统、Rating计算、代码对战、班级管理等功能
#       hoj-frontend 已包含：报名系统前端

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
PROJECT_DIR="/Users/zhuangqingjia/vscode/histoj/hist-oj"
REMOTE_DIR="/opt"
TARGET_PLATFORM="${TARGET_PLATFORM:-linux/amd64}"
HIST_OJ_NO_CACHE="${HIST_OJ_NO_CACHE:-0}"
FRONTEND_NO_CACHE="${FRONTEND_NO_CACHE:-0}"
PRUNE_BEFORE_BUILD="${PRUNE_BEFORE_BUILD:-0}"
GO_BUILD_FORCE_REBUILD="${GO_BUILD_FORCE_REBUILD:-0}"
PULL_BASE_IMAGES="${PULL_BASE_IMAGES:-0}"
CLEAN_FRONTEND_LOCAL_CACHE="${CLEAN_FRONTEND_LOCAL_CACHE:-0}"
TAR_COMPRESS_LEVEL="${TAR_COMPRESS_LEVEL:-1}"
CLEANUP_REMOTE_AFTER_DEPLOY="${CLEANUP_REMOTE_AFTER_DEPLOY:-0}"

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

# 重试执行命令
retry_cmd() {
    local retries=$1
    local sleep_seconds=$2
    shift 2
    local n=1

    while [ $n -le "$retries" ]; do
        if "$@"; then
            return 0
        fi
        if [ $n -lt "$retries" ]; then
            log_warn "命令失败，${sleep_seconds}s 后重试 (${n}/${retries})"
            sleep "$sleep_seconds"
        fi
        n=$((n + 1))
    done
    return 1
}

# 预拉取基础镜像（规避网络波动导致的构建失败）
pull_base_images() {
    log_info "预拉取基础镜像..."
    retry_cmd 3 5 docker pull --platform "${TARGET_PLATFORM}" node:16-alpine || {
        log_error "node:16-alpine 拉取失败"
        return 1
    }
    retry_cmd 3 5 docker pull --platform "${TARGET_PLATFORM}" nginx:alpine || {
        log_error "nginx:alpine 拉取失败"
        return 1
    }
    retry_cmd 3 5 docker pull --platform "${TARGET_PLATFORM}" golang:1.24-alpine || {
        log_error "golang:1.24-alpine 拉取失败"
        return 1
    }
    retry_cmd 3 5 docker pull --platform "${TARGET_PLATFORM}" alpine:3.18 || {
        log_error "alpine:3.18 拉取失败"
        return 1
    }
    log_info "✓ 基础镜像拉取完成"
}

# 构建前端镜像（BuildKit 失败时自动降级 legacy builder）
build_frontend_image() {
    local i=1
    local max_try=3
    local npm_registries="${NPM_REGISTRIES:-https://registry.npmmirror.com https://registry.npmjs.org}"
    local npm_install_retries="${NPM_INSTALL_RETRIES:-3}"
    local frontend_no_cache_args=()
    if [ "${FRONTEND_NO_CACHE}" = "1" ]; then
        frontend_no_cache_args+=(--no-cache)
    fi

    while [ $i -le $max_try ]; do
        if [ $i -eq 1 ]; then
            log_info "前端构建尝试 ${i}/${max_try}（BuildKit）..."
            if DOCKER_BUILDKIT=1 docker build "${frontend_no_cache_args[@]}" --progress=plain --platform "${TARGET_PLATFORM}" \
                --build-arg NPM_REGISTRIES="$npm_registries" \
                --build-arg NPM_INSTALL_RETRIES="$npm_install_retries" \
                -t hoj-frontend:latest .; then
                return 0
            fi
        else
            log_warn "前端构建尝试 ${i}/${max_try}（legacy builder，规避 short read/EOF）..."
            if DOCKER_BUILDKIT=0 docker build "${frontend_no_cache_args[@]}" --platform "${TARGET_PLATFORM}" \
                --build-arg NPM_REGISTRIES="$npm_registries" \
                --build-arg NPM_INSTALL_RETRIES="$npm_install_retries" \
                -t hoj-frontend:latest .; then
                return 0
            fi
        fi

        docker builder prune -af >/dev/null 2>&1 || true
        sleep 5
        i=$((i + 1))
    done

    return 1
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
build_images() {
    log_info "开始构建 Docker 镜像..."

    cd "$PROJECT_DIR"

    # 可选：清理前端本地缓存（默认关闭，避免每次都做慢清理）
    if [ "${CLEAN_FRONTEND_LOCAL_CACHE}" = "1" ]; then
        log_info "清理前端本地缓存..."
        cd hoj-vue
        rm -rf dist node_modules/.cache
        cd "$PROJECT_DIR"
    else
        log_info "跳过前端本地缓存清理（如需清理可设置 CLEAN_FRONTEND_LOCAL_CACHE=1）"
    fi

    if [ "${PRUNE_BEFORE_BUILD}" = "1" ]; then
        docker builder prune -af 2>/dev/null || true
    else
        log_info "跳过 Docker builder prune（保留缓存以加速构建）"
    fi

    # 可选：预拉取基础镜像（默认关闭，避免每次耗时拉取）
    if [ "${PULL_BASE_IMAGES}" = "1" ]; then
        pull_base_images || exit 1
    else
        log_info "跳过基础镜像预拉取（如需启用可设置 PULL_BASE_IMAGES=1）"
    fi

    # 构建 hist-oj（包含报名系统、sim 代码查重工具）
    log_info "构建 hist-oj 镜像（默认使用缓存，包含 sim 查重工具和报名系统）..."
    cd hist-oj
    # 删除旧镜像以避免冲突
    docker rmi hist-oj:latest 2>/dev/null || true
    local hist_no_cache_args=()
    if [ "${HIST_OJ_NO_CACHE}" = "1" ]; then
        hist_no_cache_args+=(--no-cache)
    fi
    # BuildKit 失败时自动降级 legacy builder（规避交叉编译 EOF 问题）
    log_info "hist-oj 构建尝试 1/3（BuildKit）..."
    if ! DOCKER_BUILDKIT=1 docker build "${hist_no_cache_args[@]}" --progress=plain --platform "${TARGET_PLATFORM}" \
        --build-arg GO_BUILD_FORCE_REBUILD="${GO_BUILD_FORCE_REBUILD}" \
        -t hist-oj:latest .; then
        docker builder prune -af >/dev/null 2>&1 || true
        sleep 5
        log_warn "hist-oj 构建尝试 2/3（legacy builder）..."
        if ! DOCKER_BUILDKIT=0 docker build "${hist_no_cache_args[@]}" --platform "${TARGET_PLATFORM}" \
            --build-arg GO_BUILD_FORCE_REBUILD="${GO_BUILD_FORCE_REBUILD}" \
            -t hist-oj:latest .; then
            docker builder prune -af >/dev/null 2>&1 || true
            sleep 5
            log_warn "hist-oj 构建尝试 3/3（legacy builder）..."
            if ! DOCKER_BUILDKIT=0 docker build "${hist_no_cache_args[@]}" --platform "${TARGET_PLATFORM}" \
                --build-arg GO_BUILD_FORCE_REBUILD="${GO_BUILD_FORCE_REBUILD}" \
                -t hist-oj:latest .; then
                log_error "hist-oj 镜像构建失败"
                exit 1
            fi
        fi
    fi
    log_info "✓ hist-oj 镜像构建成功（包含报名系统、sim_c, sim_java 查重工具）"

    # 构建前端（Docker 构建时会自动安装 package.json 中的所有依赖，包括 jsQR）
    if [ "${FRONTEND_NO_CACHE}" = "1" ]; then
        log_info "构建 hoj-frontend 镜像（不使用缓存，包含 jsQR 二维码扫描功能和报名系统前端）..."
    else
        log_info "构建 hoj-frontend 镜像（使用缓存加速，包含 jsQR 二维码扫描功能和报名系统前端）..."
    fi

    # 切换到前端目录
    cd "$PROJECT_DIR/hoj-vue"

    # 记录构建前的镜像 ID（用于验证）
    FRONTEND_IMAGE_BEFORE=$(docker images hoj-frontend:latest --format "{{.ID}}" 2>/dev/null || echo "")
    if [ -n "$FRONTEND_IMAGE_BEFORE" ]; then
        log_info "构建前镜像 ID: $FRONTEND_IMAGE_BEFORE"
    fi

    if [ "${FRONTEND_NO_CACHE}" = "1" ]; then
        log_info "开始重新构建前端镜像（不使用缓存）..."
    else
        log_info "开始构建前端镜像（使用缓存）..."
    fi
    build_frontend_image || {
        log_error "hoj-frontend 镜像构建失败"
        exit 1
    }

    # 记录构建后的镜像 ID
    FRONTEND_IMAGE_AFTER=$(docker images hoj-frontend:latest --format "{{.ID}}" 2>/dev/null || echo "")
    if [ -n "$FRONTEND_IMAGE_AFTER" ]; then
        log_info "构建后镜像 ID: $FRONTEND_IMAGE_AFTER"
    fi

    # 验证镜像确实更新了
    if [ -n "$FRONTEND_IMAGE_BEFORE" ] && [ -n "$FRONTEND_IMAGE_AFTER" ]; then
        if [ "$FRONTEND_IMAGE_BEFORE" = "$FRONTEND_IMAGE_AFTER" ]; then
            log_warn "⚠️  警告：前端镜像 ID 未改变，可能使用了缓存"
        else
            log_info "✓ 前端镜像已成功更新"
        fi
    fi

    log_info "✓ hoj-frontend 镜像构建成功（包含报名系统前端、jsQR 扫码功能）"

    cd "$PROJECT_DIR"
}

# 保存单个镜像（支持 pigz 多核压缩）
save_single_image() {
    local image_name="$1"
    local output_file="$2"
    local level="$TAR_COMPRESS_LEVEL"

    if ! [[ "$level" =~ ^[1-9]$ ]]; then
        log_warn "TAR_COMPRESS_LEVEL=${level} 非法，自动回退到 1"
        level=1
    fi

    if command -v pigz >/dev/null 2>&1; then
        local threads
        threads="$(getconf _NPROCESSORS_ONLN 2>/dev/null || echo 4)"
        docker save "$image_name" | pigz -"${level}" -p "${threads}" > "${output_file}"
    else
        docker save "$image_name" | gzip -"${level}" > "${output_file}"
    fi
}

# 保存镜像
save_images() {
    log_info "保存 Docker 镜像..."
    log_info "压缩级别: ${TAR_COMPRESS_LEVEL}（1最快，9最小）"

    log_info "保存 hist-oj 镜像..."
    save_single_image hist-oj:latest hist-oj.tar.gz || {
        log_error "hist-oj 镜像保存失败"
        exit 1
    }
    log_info "✓ hist-oj 镜像已保存 ($(du -h hist-oj.tar.gz | cut -f1))"

    log_info "保存 hoj-frontend 镜像..."
    save_single_image hoj-frontend:latest hoj-frontend.tar.gz || {
        log_error "hoj-frontend 镜像保存失败"
        exit 1
    }
    log_info "✓ hoj-frontend 镜像已保存 ($(du -h hoj-frontend.tar.gz | cut -f1))"
}

# 上传镜像和配置文件
upload_files() {
    log_info "上传文件到服务器..."

    # 上传 hist-oj 镜像
    log_info "上传 hist-oj 镜像..."
    sshpass -p "$SERVER_PASS" scp hist-oj.tar.gz ${SERVER_USER}@${SERVER_IP}:${REMOTE_DIR}/ || {
        log_error "hist-oj 镜像上传失败"
        exit 1
    }
    log_info "✓ hist-oj 镜像上传成功"

    # 上传 hoj-frontend 镜像
    log_info "上传 hoj-frontend 镜像..."
    sshpass -p "$SERVER_PASS" scp hoj-frontend.tar.gz ${SERVER_USER}@${SERVER_IP}:${REMOTE_DIR}/ || {
        log_error "hoj-frontend 镜像上传失败"
        exit 1
    }
    log_info "✓ hoj-frontend 镜像上传成功"

    # 上传 docker-compose.prod.yml（生产环境配置，使用预构建镜像）
    log_info "上传 docker-compose.prod.yml..."
    sshpass -p "$SERVER_PASS" scp docker-compose.prod.yml ${SERVER_USER}@${SERVER_IP}:${REMOTE_DIR}/docker-compose.yml || {
        log_error "docker-compose.prod.yml 上传失败"
        exit 1
    }
    log_info "✓ docker-compose.prod.yml 上传成功"

    # 上传静态资源文件（logo、favicon等）
    log_info "上传静态资源文件..."
    sshpass -p "$SERVER_PASS" scp ${PROJECT_DIR}/logo.png ${PROJECT_DIR}/backstage.png ${PROJECT_DIR}/favicon.ico ${SERVER_USER}@${SERVER_IP}:${REMOTE_DIR}/ || {
        log_warn "静态资源文件上传失败（部分文件不存在，继续部署）"
    }
    log_info "✓ 静态资源文件上传完成"
}

# 远程部署
remote_deploy() {
    log_info "开始远程部署..."

    sshpass -p "$SERVER_PASS" ssh ${SERVER_USER}@${SERVER_IP} << 'ENDSSH'
set -e

echo "======================================"
echo "开始在远程服务器上部署"
echo "======================================"

cd /opt

# 1. 停止旧容器
echo ""
echo "[INFO] 停止旧容器..."
docker stop hist-oj hoj-frontend 2>/dev/null || true
docker rm hist-oj hoj-frontend 2>/dev/null || true

# 2. 加载新镜像
echo ""
echo "[INFO] 加载 hist-oj 镜像..."
gunzip -c hist-oj.tar.gz | docker load

echo ""
echo "[INFO] 加载 hoj-frontend 镜像..."
gunzip -c hoj-frontend.tar.gz | docker load

# 3. 启动新容器
echo ""
echo "[INFO] 启动新容器..."
docker-compose up -d

# 4. 等待容器启动
echo ""
echo "[INFO] 等待容器启动（10秒）..."
sleep 10

# 5. 检查容器状态
echo ""
echo "[INFO] 检查容器状态..."
docker ps | grep -E "hist-oj|hoj-frontend"

# 6. 配置网络连接（确保前端容器能连接到hoj-backend）
echo ""
echo "[INFO] 配置网络连接..."
if docker network inspect main_hoj-network >/dev/null 2>&1; then
    # 连接前端容器到main_hoj-network（hoj-backend所在网络）
    if ! docker network inspect main_hoj-network --format '{{range .Containers}}{{.Name}}{{end}}' | grep -q hoj-frontend; then
        echo "[INFO] 将 hoj-frontend 连接到 main_hoj-network..."
        docker network connect main_hoj-network hoj-frontend
        echo "[INFO] ✓ 网络连接成功"
    else
        echo "[INFO] hoj-frontend 已连接到 main_hoj-network"
    fi
else
    echo "[WARN] main_hoj-network 不存在，跳过网络连接"
fi

# 6. 验证服务
echo ""
echo "======================================"
echo "验证服务"
echo "======================================"

echo ""
echo "[INFO] 测试 hist-oj 健康检查..."
docker exec hist-oj curl -s http://localhost:9527/health || echo "⚠️  hist-oj 健康检查失败"

echo ""
echo "[INFO] 测试报名系统 API..."
docker exec hist-oj curl -s http://localhost:9527/api/registration/competitions || echo "⚠️  报名系统 API 测试失败"

echo ""
echo "[INFO] 测试前端服务..."
docker exec hoj-frontend curl -s http://127.0.0.1/ || echo "⚠️  前端服务测试失败"

echo ""
echo "[INFO] 测试 hist-oj 到报名系统的连接..."
docker exec hoj-frontend curl -s http://hist-oj:9527/api/registration/competitions || echo "⚠️  前端到hist-oj连接测试失败"

echo ""
echo "[INFO] 测试 hoj-frontend 到 hoj-backend 的连接..."
if docker network inspect main_hoj-network >/dev/null 2>&1 && docker network inspect main_hoj-network --format '{{range .Containers}}{{.Name}}{{end}}' | grep -q hoj-frontend; then
    docker exec hoj-frontend ping -c 1 hoj-backend >/dev/null 2>&1 && echo "✓ hoj-backend 网络连接正常" || echo "⚠️  hoj-backend 网络连接失败"
else
    echo "⚠️  hoj-frontend 未连接到 main_hoj-network"
fi

echo ""
echo "======================================"
echo "✅ 部署完成！"
echo "======================================"
echo ""
echo "访问地址："
echo "  - 前端: http://$SERVER_IP/"
echo "  - 后端: http://$SERVER_IP:9527"
echo "  - 报名系统: http://$SERVER_IP/registration/list"
echo ""
echo "端口说明："
echo "  - 前端容器使用 8081 端口"
echo "  - 宿主机 Nginx 监听 80/443 端口并代理到 8081"
echo "  - 如需 HTTPS，请运行: bash scripts/deploy-ssl.sh"
echo ""

ENDSSH

    if [ $? -eq 0 ]; then
        log_info "✓ 远程部署成功"
    else
        log_error "远程部署失败"
        exit 1
    fi
}

# 清理本地临时文件
cleanup_local() {
    log_info "清理本地临时文件..."
    rm -f hist-oj.tar.gz hoj-frontend.tar.gz
    log_info "✓ 临时文件已清理"
}

# 清理远程旧镜像
cleanup_remote() {
    log_info "清理远程旧镜像..."

    sshpass -p "$SERVER_PASS" ssh ${SERVER_USER}@${SERVER_IP} << 'ENDSSH'
cd /opt

# 清理本地临时文件
rm -f hist-oj.tar.gz hoj-frontend.tar.gz

# 列出当前镜像
echo ""
echo "[INFO] 当前镜像列表："
docker images | grep -E "REPOSITORY|hist-oj|hoj-frontend"

# 清理 hist-oj 旧镜像（保留最新的 2 个）
echo ""
echo "[INFO] 清理 hist-oj 旧镜像..."
docker images hist-oj --format "{{.ID}} {{.CreatedAt}}" | sort -k2 -r | tail -n +3 | while read IMAGE_ID CREATED; do
    echo "[INFO] 删除旧镜像: $IMAGE_ID"
    docker rmi $IMAGE_ID 2>/dev/null || true
done

# 清理 hoj-frontend 旧镜像（保留最新的 2 个）
echo ""
echo "[INFO] 清理 hoj-frontend 旧镜像..."
docker images hoj-frontend --format "{{.ID}} {{.CreatedAt}}" | sort -k2 -r | tail -n +3 | while read IMAGE_ID CREATED; do
    echo "[INFO] 删除旧镜像: $IMAGE_ID"
    docker rmi $IMAGE_ID 2>/dev/null || true
done

echo ""
echo "[INFO] 清理后的镜像列表："
docker images | grep -E "REPOSITORY|hist-oj|hoj-frontend"

ENDSSH

    log_info "✓ 远程旧镜像已清理"
}

# 显示部署信息
show_deployment_info() {
    log_info "======================================"
    log_info "部署信息"
    log_info "======================================"
    log_info "  服务器: ${SERVER_IP}"
    log_info "  用户: ${SERVER_USER}"
    log_info "  项目目录: ${PROJECT_DIR}"
    log_info "  远程目录: ${REMOTE_DIR}"
    log_info "  前端无缓存构建: ${FRONTEND_NO_CACHE}"
    log_info "  预拉基础镜像: ${PULL_BASE_IMAGES}"
    log_info "  压缩级别: ${TAR_COMPRESS_LEVEL}"
    log_info "  部署后远端清理: ${CLEANUP_REMOTE_AFTER_DEPLOY}"
    log_info ""
    log_info "部署的服务："
    log_info "  1. hist-oj（后端 + 报名系统 + 对战 + Rating + 班级管理）"
    log_info "  2. hoj-frontend（前端）"
    log_info ""
    log_info "端口映射："
    log_info "  - hist-oj: 9527"
    log_info "  - hoj-frontend 容器: 8081"
    log_info "  - 宿主机 Nginx: 80, 443 (SSL)"
    log_info "======================================"
}

# 部署 SSL 证书
deploy_ssl() {
    log_info ""
    log_info "======================================"
    log_info "部署 SSL 证书"
    log_info "======================================"

    # 检查证书文件
    log_info "检查 SSL 证书文件..."
    if [ ! -f "scripts/nginx/ssl/bingoj.cn.crt" ]; then
        log_error "证书文件不存在: scripts/nginx/ssl/bingoj.cn.crt"
        log_warn "跳过 SSL 部署"
        return 1
    fi

    if [ ! -f "scripts/nginx/ssl/bingoj.cn.key" ]; then
        log_error "私钥文件不存在: scripts/nginx/ssl/bingoj.cn.key"
        log_warn "跳过 SSL 部署"
        return 1
    fi

    if [ ! -f "scripts/nginx/bingoj.conf" ]; then
        log_error "Nginx 配置文件不存在: scripts/nginx/bingoj.conf"
        log_warn "跳过 SSL 部署"
        return 1
    fi

    log_info "✓ 所有必需文件都存在"

    # 上传证书和配置
    log_info "上传 SSL 证书和配置到服务器..."

    # 创建 SSL 目录
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

    # 配置并启动 Nginx
    log_info "配置并启动 Nginx..."

    sshpass -p "$SERVER_PASS" ssh -o StrictHostKeyChecking=no ${SERVER_USER}@${SERVER_IP} << 'ENDSSL'
        set -e

        # 备份旧配置
        if [ -f /etc/nginx/conf.d/bingoj.conf ] && [ ! -f /etc/nginx/conf.d/bingoj.conf.bak ]; then
            cp /etc/nginx/conf.d/bingoj.conf /etc/nginx/conf.d/bingoj.conf.bak
            echo "[INFO] 已备份旧配置"
        fi

        # 确保 Docker 容器使用正确的端口（8081）
        echo "[INFO] 检查 Docker 容器端口配置..."
        cd /opt
        if grep -q '"80:80"' docker-compose.yml; then
            echo "[INFO] 更新 Docker 容器端口配置..."
            sed -i 's/"80:80"/"8081:80"/g' docker-compose.yml
            sed -i '/.*443:443.*/d' docker-compose.yml
            echo "[INFO] 重启前端容器..."
            docker stop hoj-frontend
            docker rm hoj-frontend
            docker-compose up -d hoj-frontend
            sleep 5
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
ENDSSL

    if [ $? -eq 0 ]; then
        log_info "✓ Nginx 配置成功"
    else
        log_error "Nginx 配置失败"
        return 1
    fi

    # 确保网络连接正确
    log_info "确保 Docker 网络连接..."
    sshpass -p "$SERVER_PASS" ssh -o StrictHostKeyChecking=no ${SERVER_USER}@${SERVER_IP} << 'ENDSSH'
        # 连接前端容器到 main_hoj-network
        if docker network inspect main_hoj-network >/dev/null 2>&1; then
            if ! docker network inspect main_hoj-network --format '{{range .Containers}}{{.Name}}{{end}}' | grep -q hoj-frontend; then
                echo "[INFO] 将 hoj-frontend 连接到 main_hoj-network..."
                docker network connect main_hoj-network hoj-frontend
                echo "[INFO] ✓ 网络连接成功"
            else
                echo "[INFO] hoj-frontend 已连接到 main_hoj-network"
            fi
        fi
ENDSSH

    # 测试 HTTPS
    log_info "测试 HTTPS 连接..."
    HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" https://bingoj.cn 2>/dev/null || echo "000")

    if [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "301" ] || [ "$HTTP_CODE" = "302" ]; then
        log_info "✓ HTTPS 访问成功 (HTTP $HTTP_CODE)"
    else
        log_warn "HTTPS 返回状态码: $HTTP_CODE"
    fi

    log_info "✓ SSL 部署完成！"
    return 0
}

# 主函数
main() {
    log_info "======================================"
    log_info "HOJ + 报名系统一键部署"
    log_info "======================================"
    log_info ""

    show_deployment_info

    # 询问是否继续
    echo ""
    read -p "是否继续部署？(y/n): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_warn "部署已取消"
        exit 0
    fi

    # 询问是否部署 SSL
    echo ""
    read -p "是否部署 SSL 证书（HTTPS）？(y/n，默认 y): " -n 1 -r
    echo
    DEPLOY_SSL=true
    if [[ ! $REPLY =~ ^[Yy]$ ]] && [[ ! -z $REPLY ]]; then
        DEPLOY_SSL=false
        log_warn "跳过 SSL 部署，仅使用 HTTP"
    fi

    # 执行部署步骤
    check_requirements
    build_images
    save_images
    upload_files
    remote_deploy
    cleanup_local
    if [ "${CLEANUP_REMOTE_AFTER_DEPLOY}" = "1" ]; then
        cleanup_remote
    else
        log_info "跳过远端镜像清理（如需清理可设置 CLEANUP_REMOTE_AFTER_DEPLOY=1）"
    fi

    # 部署 SSL（如果选择）
    if [ "$DEPLOY_SSL" = true ]; then
        deploy_ssl
        SSL_STATUS=$?
        if [ $SSL_STATUS -eq 0 ]; then
            SSL_MESSAGE="✓ HTTPS 已启用"
        else
            SSL_MESSAGE="⚠️  SSL 部署失败，仅使用 HTTP"
        fi
    else
        SSL_MESSAGE="ℹ️  未部署 SSL，使用 HTTP 访问"
    fi

    log_info ""
    log_info "======================================"
    log_info "🎉 部署成功完成！"
    log_info "======================================"
    log_info ""
    log_info "访问地址："
    if [ "$DEPLOY_SSL" = true ] && [ $SSL_STATUS -eq 0 ]; then
        log_info "  - 前端: https://bingoj.cn/"
        log_info "  - 报名系统: https://bingoj.cn/registration/list"
        log_info "  - 管理后台: https://bingoj.cn/admin/registration"
        log_info ""
        log_info "  $SSL_MESSAGE"
    else
        log_info "  - 前端: http://${SERVER_IP}/"
        log_info "  - 报名系统: http://${SERVER_IP}/registration/list"
        log_info "  - 管理后台: http://${SERVER_IP}/admin/registration"
        log_info ""
        log_info "  $SSL_MESSAGE"
        log_info ""
        log_info "  如需启用 HTTPS，请运行："
        log_info "  bash scripts/deploy-ssl.sh"
    fi
    log_info ""
}

# 运行主函数
main "$@"
