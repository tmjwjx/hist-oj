#!/bin/bash

# HOJ2 + 报名系统自动化部署脚本
# 用途：一键构建、打包、上传、部署 hist-oj、hoj-frontend 和 registration-backend 服务

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
build_images() {
    log_info "开始构建 Docker 镜像..."

    cd "$PROJECT_DIR"

    # 构建 hist-oj
    log_info "构建 hist-oj 镜像（不使用缓存）..."
    cd hist-oj
    docker build --no-cache --platform linux/amd64 -t hist-oj:latest . || {
        log_error "hist-oj 镜像构建失败"
        exit 1
    }
    log_info "✓ hist-oj 镜像构建成功"

    # 构建报名系统后端
    log_info "构建 registration-backend 镜像（不使用缓存）..."
    cd ../registration-system/backend
    docker build --no-cache --platform linux/amd64 -t registration-backend:latest . || {
        log_error "registration-backend 镜像构建失败"
        exit 1
    }
    log_info "✓ registration-backend 镜像构建成功"

    # 构建前端
    log_info "构建 hoj-frontend 镜像（注意：这将清理旧的构建缓存）..."
    cd ../../hoj-vue

    # 清理旧的构建文件和 Docker 缓存
    log_info "清理旧的构建文件和 Docker 缓存..."
    rm -rf dist node_modules/.cache
    docker builder prune -af
    docker system prune -af --volumes 2>/dev/null || true

    # 记录构建前的镜像 ID（用于验证）
    FRONTEND_IMAGE_BEFORE=$(docker images hoj-frontend:latest --format "{{.ID}}" 2>/dev/null || echo "")
    if [ -n "$FRONTEND_IMAGE_BEFORE" ]; then
        log_info "构建前镜像 ID: $FRONTEND_IMAGE_BEFORE"
    fi

    # 强制重新构建（不使用任何缓存）
    log_info "开始重新构建前端镜像（不使用缓存）..."
    docker build \
        --no-cache \
        --pull \
        --platform linux/amd64 \
        --progress=plain \
        -t hoj-frontend:latest \
        . || {
        log_error "hoj-frontend 镜像构建失败"
        exit 1
    }
    log_info "✓ hoj-frontend 镜像构建成功"

    # 显示新镜像信息
    NEW_IMAGE_ID=$(docker images hoj-frontend:latest --format "{{.ID}}")
    NEW_IMAGE_SIZE=$(docker images hoj-frontend:latest --format "{{.Size}}")
    log_info "新镜像 ID: $NEW_IMAGE_ID, 大小: $NEW_IMAGE_SIZE"

    # 验证镜像是否真的更新了
    if [ "$FRONTEND_IMAGE_BEFORE" = "$NEW_IMAGE_ID" ]; then
        log_error "⚠️  警告：镜像 ID 未变化！可能使用了缓存。"
        log_error "建议：手动删除镜像后重新构建"
        log_error "命令：docker rmi $NEW_IMAGE_ID && ./scripts/deploy.sh"
    else
        log_info "✓ 镜像已更新（旧: $FRONTEND_IMAGE_BEFORE -> 新: $NEW_IMAGE_ID）"
    fi

    cd "$PROJECT_DIR"
}

# 验证镜像内容
verify_image_content() {
    log_info "验证镜像内容..."
    docker run --rm hoj-frontend:latest sh -c '
        # 检查 app.js 文件
        APP_JS=$(find /usr/share/nginx/html/assets/js -name "app.*.js" -type f | head -1)
        if [ -z "$APP_JS" ]; then
            echo "[ERROR] 未找到 app.js 文件"
            exit 1
        fi

        echo "[INFO] 找到 app.js: $(basename $APP_JS)"

        # 检查关键功能是否存在
        if grep -q "fa-briefcase" "$APP_JS"; then
            COUNT=$(grep -o "fa-briefcase" "$APP_JS" | wc -l)
            echo "[INFO] ✓ fa-briefcase 图标已包含 ($COUNT 处)"
        else
            echo "[ERROR] fa-briefcase 图标未找到！"
            exit 1
        fi

        # 检查导航栏图标
        if grep -q "NavBar" "$APP_JS" && grep -q "toolbox" "$APP_JS"; then
            echo "[INFO] ✓ 导航栏工具箱代码已包含"
        else
            echo "[WARN] 导航栏代码可能有问题"
        fi
    ' || {
        log_error "镜像内容验证失败！"
        log_error "这表明构建没有正确包含最新代码。"
        exit 1
    }
    log_info "✓ 镜像内容验证通过"
}

# 保存镜像
save_images() {
    local DEPLOY_TARGET="${1:-all}"
    log_info "保存 Docker 镜像（目标: $DEPLOY_TARGET）..."

    cd "$PROJECT_DIR"

    if [ "$DEPLOY_TARGET" = "all" ] || [ "$DEPLOY_TARGET" = "backend" ]; then
        # 保存 hist-oj
        log_info "保存 hist-oj 镜像..."
        docker save hist-oj:latest | gzip > hist-oj.tar.gz || {
            log_error "hist-oj 镜像保存失败"
            exit 1
        }
        log_info "✓ hist-oj 镜像已保存 ($(du -h hist-oj.tar.gz | cut -f1))"
    fi

    if [ "$DEPLOY_TARGET" = "all" ] || [ "$DEPLOY_TARGET" = "registration" ]; then
        # 保存报名系统
        log_info "保存 registration-backend 镜像..."
        docker save registration-backend:latest | gzip > registration-backend.tar.gz || {
            log_error "registration-backend 镜像保存失败"
            exit 1
        }
        log_info "✓ registration-backend 镜像已保存 ($(du -h registration-backend.tar.gz | cut -f1))"
    fi

    if [ "$DEPLOY_TARGET" = "all" ] || [ "$DEPLOY_TARGET" = "frontend" ]; then
        # 保存前端
        log_info "保存 hoj-frontend 镜像..."
        docker save hoj-frontend:latest | gzip > hoj-frontend.tar.gz || {
            log_error "hoj-frontend 镜像保存失败"
            exit 1
        }
        log_info "✓ hoj-frontend 镜像已保存 ($(du -h hoj-frontend.tar.gz | cut -f1))"
    fi
}

# 上传镜像到服务器
upload_images() {
    local DEPLOY_TARGET="${1:-all}"
    log_info "上传镜像到服务器（目标: $DEPLOY_TARGET）..."

    cd "$PROJECT_DIR"

    if [ "$DEPLOY_TARGET" = "all" ] || [ "$DEPLOY_TARGET" = "backend" ]; then
        # 上传 hist-oj
        log_info "上传 hist-oj 镜像..."
        sshpass -p "$SERVER_PASS" scp hist-oj.tar.gz ${SERVER_USER}@${SERVER_IP}:${REMOTE_DIR}/ || {
            log_error "hist-oj 镜像上传失败"
            exit 1
        }
        log_info "✓ hist-oj 镜像上传成功"
    fi

    if [ "$DEPLOY_TARGET" = "all" ] || [ "$DEPLOY_TARGET" = "registration" ]; then
        # 上传报名系统
        log_info "上传 registration-backend 镜像..."
        sshpass -p "$SERVER_PASS" scp registration-backend.tar.gz ${SERVER_USER}@${SERVER_IP}:${REMOTE_DIR}/ || {
            log_error "registration-backend 镜像上传失败"
            exit 1
        }
        log_info "✓ registration-backend 镜像上传成功"
    fi

    if [ "$DEPLOY_TARGET" = "all" ] || [ "$DEPLOY_TARGET" = "frontend" ]; then
        # 上传前端
        log_info "上传 hoj-frontend 镜像..."
        sshpass -p "$SERVER_PASS" scp hoj-frontend.tar.gz ${SERVER_USER}@${SERVER_IP}:${REMOTE_DIR}/ || {
            log_error "hoj-frontend 镜像上传失败"
            exit 1
        }
        log_info "✓ hoj-frontend 镜像上传成功"
    fi

    # 上传数据库迁移脚本（只在完整部署时上传）
    if [ "$DEPLOY_TARGET" = "all" ]; then
        log_info "上传数据库迁移脚本..."

        # 上传 Rating 系统迁移脚本
        sshpass -p "$SERVER_PASS" scp hist-oj/migrations/005_add_manual_rating_fields.sql ${SERVER_USER}@${SERVER_IP}:${REMOTE_DIR}/ 2>/dev/null || {
            log_warn "Rating 迁移脚本上传失败（可能不存在）"
        }

        # 上传报名系统迁移脚本（新增）
        sshpass -p "$SERVER_PASS" scp registration-system/migrations/001_add_last_view_time_fields.sql ${SERVER_USER}@${SERVER_IP}:${REMOTE_DIR}/ 2>/dev/null || {
            log_warn "报名系统迁移脚本上传失败（可能不存在）"
        }

        log_info "✓ 数据库迁移脚本上传完成"
    fi
}

# 在服务器上部署
deploy_on_server() {
    local DEPLOY_TARGET="${1:-all}"
    log_info "在服务器上部署服务（目标: $DEPLOY_TARGET）..."

    # 传递参数到远程脚本
    sshpass -p "$SERVER_PASS" ssh -o StrictHostKeyChecking=no ${SERVER_USER}@${SERVER_IP} "DEPLOY_TARGET='$DEPLOY_TARGET'" << 'ENDSSH'
        set -e

        echo "[INFO] 加载 Docker 镜像..."
        cd /opt

        # 记录加载前的镜像 ID
        FRONTEND_IMAGE_BEFORE=$(docker images hoj-frontend:latest --format "{{.ID}}" 2>/dev/null || echo "")

        # 加载 hist-oj 镜像
        echo "[INFO] 加载 hist-oj 镜像..."
        gunzip -c hist-oj.tar.gz | docker load

        # 加载报名系统镜像
        echo "[INFO] 加载 registration-backend 镜像..."
        gunzip -c registration-backend.tar.gz | docker load

        # 加载前端镜像
        echo "[INFO] 加载 hoj-frontend 镜像..."
        gunzip -c hoj-frontend.tar.gz | docker load

        # 验证前端镜像是否更新
        FRONTEND_IMAGE_AFTER=$(docker images hoj-frontend:latest --format "{{.ID}}")
        if [ "$FRONTEND_IMAGE_BEFORE" != "$FRONTEND_IMAGE_AFTER" ]; then
            echo "[INFO] ✓ hoj-frontend 镜像已更新"
            echo "[INFO]   旧镜像: $FRONTEND_IMAGE_BEFORE"
            echo "[INFO]   新镜像: $FRONTEND_IMAGE_AFTER"
        else
            echo "[WARN] hoj-frontend 镜像 ID 未变化（可能使用了相同的基础层）"
        fi

        # 显示新镜像的创建时间和大小
        docker images hoj-frontend:latest --format "[INFO] 镜像信息: 创建于 {{.CreatedAt}}, 大小: {{.Size}}"

        # 执行数据库迁移（如果迁移脚本存在且字段未添加）
        echo "[INFO] 检查数据库迁移..."

        # 检查 last_view_time 和 admin_last_view_time 字段是否已存在
        FIELD_EXISTS=$(mysql -h43.143.133.62 -uroot -phist2025 -sN -e \
            "SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS \
             WHERE TABLE_SCHEMA='hoj' \
             AND TABLE_NAME='histcontest_register_registrations' \
             AND COLUMN_NAME IN ('last_view_time', 'admin_last_view_time')" 2>/dev/null || echo "0")

        if [ "$FIELD_EXISTS" -ge "2" ]; then
            echo "[INFO] ✓ 数据库字段已存在，跳过迁移"
        else
            echo "[INFO] 需要执行数据库迁移..."

            # Rating 系统迁移
            if [ -f "/opt/005_add_manual_rating_fields.sql" ]; then
                echo "[INFO] 执行 Rating 系统数据库迁移..."
                mysql -h43.143.133.62 -uroot -phist2025 hoj < /opt/005_add_manual_rating_fields.sql && echo "[INFO] ✓ Rating 迁移成功" || echo "[WARN] Rating 迁移失败"
            fi

            # 报名系统迁移
            if [ -f "/opt/001_add_last_view_time_fields.sql" ]; then
                echo "[INFO] 执行报名系统数据库迁移..."
                mysql -h43.143.133.62 -uroot -phist2025 hoj < /opt/001_add_last_view_time_fields.sql && echo "[INFO] ✓ 报名系统迁移成功" || echo "[WARN] 报名系统迁移失败"
            fi
        fi

        echo "[INFO] 停止并删除旧容器..."

        # 根据部署目标选择性停止容器
        if [ "$DEPLOY_TARGET" = "all" ] || [ "$DEPLOY_TARGET" = "backend" ]; then
            echo "[INFO] 停止 hist-oj 容器..."
            docker stop hist-oj 2>/dev/null || true
            docker rm hist-oj 2>/dev/null || true
        fi

        if [ "$DEPLOY_TARGET" = "all" ] || [ "$DEPLOY_TARGET" = "registration" ]; then
            echo "[INFO] 停止 registration-backend 容器..."
            docker stop registration-backend 2>/dev/null || true
            docker rm registration-backend 2>/dev/null || true
        fi

        if [ "$DEPLOY_TARGET" = "all" ] || [ "$DEPLOY_TARGET" = "frontend" ]; then
            echo "[INFO] 停止 hoj-frontend 容器..."
            docker stop hoj-frontend 2>/dev/null || true
            docker rm hoj-frontend 2>/dev/null || true

            # 等待容器完全停止
            sleep 2

            # 验证容器已被删除
            if docker ps -a | grep -q hoj-frontend; then
                echo "[WARN] hoj-frontend 容器仍然存在，强制删除..."
                docker stop hoj-frontend 2>/dev/null || true
                docker rm hoj-frontend 2>/dev/null || true
            fi
            echo "[INFO] ✓ hoj-frontend 容器已停止并删除"
        fi

        echo "[INFO] 创建上传文件目录..."
        mkdir -p /opt/registration-uploads

        # 根据部署目标选择性启动容器
        if [ "$DEPLOY_TARGET" = "all" ] || [ "$DEPLOY_TARGET" = "backend" ]; then
            echo "[INFO] 启动 hist-oj 容器..."
            docker run -d \
                --name hist-oj \
                --network hoj_hoj-network \
                -p 9527:9527 \
                -v /workspace/hoj-deploy/distributed/main/hist-oj/configs:/app/configs \
                --restart unless-stopped \
                hist-oj:latest
        fi

        if [ "$DEPLOY_TARGET" = "all" ] || [ "$DEPLOY_TARGET" = "registration" ]; then
            echo "[INFO] 启动 registration-backend 容器..."
            docker run -d \
                --name registration-backend \
                --network hoj_hoj-network \
                -v /opt/registration-uploads:/app/uploads \
                -e DATABASE_HOST=43.143.133.62 \
                -e DATABASE_PORT=3306 \
                -e DATABASE_USER=root \
                -e DATABASE_PASSWORD=hist2025 \
                -e DATABASE_DBNAME=hoj \
                -e SERVER_PORT=8080 \
                -e TZ=Asia/Shanghai \
                --restart unless-stopped \
                --health-cmd="curl -f http://localhost:8080/api/competitions || exit 1" \
                --health-interval=30s \
                --health-timeout=10s \
                --health-retries=3 \
                --health-start-period=40s \
                registration-backend:latest
        fi

        if [ "$DEPLOY_TARGET" = "all" ] || [ "$DEPLOY_TARGET" = "frontend" ]; then
            echo "[INFO] 启动 hoj-frontend 容器..."
            docker run -d \
                --name hoj-frontend \
                --network hoj_hoj-network \
                -p 80:80 \
                -p 443:443 \
                --restart unless-stopped \
                --health-cmd="wget --no-verbose --tries=1 --spider http://127.0.0.1/ || exit 1" \
                --health-interval=30s \
                --health-timeout=3s \
                --health-retries=3 \
                --health-start-period=10s \
                hoj-frontend:latest
        fi

        echo "[INFO] 等待服务启动..."
        sleep 5

        echo "[INFO] 验证服务状态..."
        docker ps | grep -E "hist-oj|registration-backend|hoj-frontend"

        echo "[INFO] 测试 hist-oj 健康检查..."
        curl -s http://localhost:9527/health || echo "健康检查失败"

        echo "[INFO] 测试手动调整 Rating API（需要管理员权限）..."
        # 注释掉自动测试，避免每次部署都给 test 用户 +10 分
        # curl -s -X POST http://localhost:9527/api/rating/admin/adjust \
        #     -H "Content-Type: application/json" \
        #     -H "X-Operator-UID: admin" \
        #     -d '{"username": "test", "ratingChange": 10, "reason": "部署测试"}' \
        #     || echo "手动调整 API 测试失败（预期行为，因为用户可能不存在）"
        echo "[INFO] 已禁用自动 Rating 测试（如需测试，请手动执行）"

        echo "[INFO] 测试报名系统 API..."
        docker exec hoj-frontend curl -s http://registration-backend:8080/api/competitions || echo "报名系统 API 测试失败"

        echo "[INFO] 测试前端访问 hist-oj..."
        docker exec hoj-frontend curl -s http://hist-oj:9527/health || echo "前端访问 hist-oj 失败"

        echo "[INFO] 测试工具箱路由（用户端）..."
        docker exec hoj-frontend curl -s -o /dev/null -w "%{http_code}" http://localhost/toolbox | grep -q "200" && echo "[INFO] ✓ 用户工具箱路由正常" || echo "[WARN] 用户工具箱路由异常"

        echo "[INFO] 测试工具箱路由（管理端）..."
        docker exec hoj-frontend curl -s -o /dev/null -w "%{http_code}" http://localhost/admin/toolbox | grep -q "200" && echo "[INFO] ✓ 管理员工具箱路由正常" || echo "[WARN] 管理员工具箱路由异常"

        echo "[INFO] 验证工具箱组件文件是否在容器中..."
        # 查找实际的 app.js 文件（文件名带有 hash）
        APP_JS=$(docker exec hoj-frontend find /usr/share/nginx/html/assets/js -name "app.*.js" -type f 2>/dev/null | head -1)
        if [ -n "$APP_JS" ]; then
            echo "[INFO] 找到 app.js 文件: $(basename $APP_JS)"

            # 检查文件修改时间
            FILE_TIME=$(docker exec hoj-frontend stat -c "%Y" "$APP_JS" 2>/dev/null)
            CURRENT_TIME=$(date +%s)
            TIME_DIFF=$((CURRENT_TIME - FILE_TIME))

            if [ $TIME_DIFF -lt 300 ]; then
                echo "[INFO] ✓ app.js 文件是最近创建的（$TIME_DIFF 秒前）"
            else
                echo "[WARN] app.js 文件较旧（$TIME_DIFF 秒前），可能不是最新版本"
            fi

            if docker exec hoj-frontend grep -q "toolbox" "$APP_JS" 2>/dev/null; then
                COUNT=$(docker exec hoj-frontend grep -o "toolbox" "$APP_JS" 2>/dev/null | wc -l)
                echo "[INFO] ✓ 工具箱代码已包含在构建文件中（找到 $COUNT 处引用）"
            else
                echo "[WARN] 工具箱代码未找到，可能构建有问题"
            fi

            # 检查新图标
            if docker exec hoj-frontend grep -q "el-icon-s-grid" "$APP_JS" 2>/dev/null; then
                echo "[INFO] ✓ 新图标 el-icon-s-grid 已包含"
            else
                echo "[WARN] 新图标未找到"
            fi

            # 检查 Rating 管理（在 chunk 文件中）
            CHUNK_JS=$(docker exec hoj-frontend find /usr/share/nginx/html/assets/js -name "chunk-*.js" -type f -exec grep -l "ToolboxAdmin" {} \; 2>/dev/null | head -1)
            if [ -n "$CHUNK_JS" ]; then
                echo "[INFO] 找到 ToolboxAdmin chunk: $(basename $CHUNK_JS)"
                if docker exec hoj-frontend grep -q "Rating 管理" "$CHUNK_JS" 2>/dev/null; then
                    echo "[INFO] ✓ Rating 管理功能已包含"
                fi
                if docker exec hoj-frontend grep -q "admin-rating" "$CHUNK_JS" 2>/dev/null; then
                    echo "[INFO] ✓ Rating 路由已正确配置"
                fi
            fi
        else
            echo "[WARN] 未找到 app.js 文件"
        fi

        echo "[INFO] 部署完成！"
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
    rm -f hist-oj.tar.gz registration-backend.tar.gz hoj-frontend.tar.gz
    log_info "✓ 清理完成"
}

# 清理服务器上的临时文件
cleanup_server() {
    log_info "清理服务器上的临时文件..."
    sshpass -p "$SERVER_PASS" ssh -o StrictHostKeyChecking=no ${SERVER_USER}@${SERVER_IP} << 'ENDSSH'
        cd /opt
        rm -f hist-oj.tar.gz registration-backend.tar.gz hoj-frontend.tar.gz 005_add_manual_rating_fields.sql 001_add_last_view_time_fields.sql
        echo "[INFO] ✓ 服务器清理完成"
ENDSSH
}

# 清理服务器上的旧镜像
cleanup_old_images() {
    log_info "清理服务器上的旧镜像..."
    sshpass -p "$SERVER_PASS" ssh -o StrictHostKeyChecking=no ${SERVER_USER}@${SERVER_IP} << 'ENDSSH'
        echo "[INFO] 检查并清理旧版本的 Docker 镜像..."

        # 获取当前运行的容器使用的镜像 ID
        RUNNING_HIST_OJ=$(docker inspect hist-oj 2>/dev/null | grep -A 1 '"Image"' | tail -1 | awk -F'"' '{print $2}' || echo "")
        RUNNING_REGISTRATION=$(docker inspect registration-backend 2>/dev/null | grep -A 1 '"Image"' | tail -1 | awk -F'"' '{print $2}' || echo "")
        RUNNING_FRONTEND=$(docker inspect hoj-frontend 2>/dev/null | grep -A 1 '"Image"' | tail -1 | awk -F'"' '{print $2}' || echo "")

        # 清理 hoj-frontend 旧镜像（保留最新的 2 个）
        echo "[INFO] 清理 hoj-frontend 旧镜像..."
        docker images hoj-frontend --format "{{.ID}} {{.CreatedAt}}" | sort -k2 -r | tail -n +3 | while read IMAGE_ID CREATED; do
            if [ "$IMAGE_ID" != "$RUNNING_FRONTEND" ]; then
                echo "[INFO]   删除旧镜像: $IMAGE_ID ($CREATED)"
                docker rmi $IMAGE_ID 2>/dev/null || echo "[WARN]   无法删除镜像 $IMAGE_ID (可能被使用)"
            fi
        done

        # 清理 hist-oj 旧镜像（保留最新的 2 个）
        echo "[INFO] 清理 hist-oj 旧镜像..."
        docker images hist-oj --format "{{.ID}} {{.CreatedAt}}" | sort -k2 -r | tail -n +3 | while read IMAGE_ID CREATED; do
            if [ "$IMAGE_ID" != "$RUNNING_HIST_OJ" ]; then
                echo "[INFO]   删除旧镜像: $IMAGE_ID ($CREATED)"
                docker rmi $IMAGE_ID 2>/dev/null || echo "[WARN]   无法删除镜像 $IMAGE_ID (可能被使用)"
            fi
        done

        # 清理 registration-backend 旧镜像（保留最新的 2 个）
        echo "[INFO] 清理 registration-backend 旧镜像..."
        docker images registration-backend --format "{{.ID}} {{.CreatedAt}}" | sort -k2 -r | tail -n +3 | while read IMAGE_ID CREATED; do
            if [ "$IMAGE_ID" != "$RUNNING_REGISTRATION" ]; then
                echo "[INFO]   删除旧镜像: $IMAGE_ID ($CREATED)"
                docker rmi $IMAGE_ID 2>/dev/null || echo "[WARN]   无法删除镜像 $IMAGE_ID (可能被使用)"
            fi
        done

        # 清理悬空镜像（dangling images）
        DANGLING=$(docker images -f "dangling=true" -q | head -20)
        if [ -n "$DANGLING" ]; then
            echo "[INFO] 清理悬空镜像..."
            docker rmi $DANGLING 2>/dev/null || echo "[WARN]   无悬空镜像需要清理"
        fi

        echo "[INFO] ✓ 旧镜像清理完成"

        # 显示剩余镜像
        echo ""
        echo "[INFO] 当前保留的镜像："
        docker images | grep -E "REPOSITORY|hist-oj|registration-backend|hoj-frontend"
ENDSSH

    if [ $? -eq 0 ]; then
        log_info "✓ 旧镜像清理成功"
    else
        log_warn "旧镜像清理部分失败（非致命错误）"
    fi
}

# 显示部署结果
show_result() {
    log_info "=========================================="
    log_info "部署完成！"
    log_info "=========================================="
    log_info "服务访问地址："
    log_info "  - 前端主页: http://${SERVER_IP}"
    log_info "  - hist-oj API: http://${SERVER_IP}:9527"
    log_info "  - 用户工具箱: http://${SERVER_IP}/toolbox"
    log_info "  - 管理员工具箱: http://${SERVER_IP}/admin/toolbox"
    log_info "  - 报名系统: http://${SERVER_IP}/toolbox -> 赛事报名系统"
    log_info "  - 报名管理: http://${SERVER_IP}/admin/toolbox -> 赛事报名系统管理"
    log_info ""
    log_info "架构更新："
    log_info "  - 已将赛事报名系统封装到工具箱中"
    log_info "  - 用户端导航: '赛事报名系统' -> '工具箱'"
    log_info "  - 管理端导航: '赛事报名系统' -> '工具箱'"
    log_info "  - 支持未来扩展更多工具到工具箱"
    log_info ""
    log_info "新增功能："
    log_info "  1. 工具箱系统"
    log_info "     - 用户工具箱: /toolbox"
    log_info "     - 管理员工具箱: /admin/toolbox"
    log_info "     - 卡片式布局，易于扩展"
    log_info ""
    log_info "  2. 手动调整 Rating"
    log_info "     - API: POST http://${SERVER_IP}:9527/api/rating/admin/adjust"
    log_info "     - 文档: hist-oj/MANUAL_RATING_ADJUST.md"
    log_info ""
    log_info "  3. 持久化消息已读状态（解决浏览器缓存清除问题）"
    log_info "     - 用户和管理员的已读状态分别存储在数据库"
    log_info "     - 支持跨设备同步已读状态"
    log_info "     - 新增字段: last_view_time, admin_last_view_time"
    log_info ""
    log_info "验证命令："
    log_info "  curl http://${SERVER_IP}:9527/health"
    log_info "  curl http://${SERVER_IP}/api/rating/contest/info/1002"
    log_info "  curl http://${SERVER_IP}/registration-api/competitions"
    log_info ""
    log_info "测试工具箱访问："
    log_info "  - 用户端: curl -I http://${SERVER_IP}/toolbox"
    log_info "  - 管理端: curl -I http://${SERVER_IP}/admin/toolbox"
    log_info ""
    log_warn "⚠️  重要提示："
    log_warn "  如果页面没有更新，请强制刷新浏览器缓存："
    log_warn "  - Windows/Linux: Ctrl + Shift + R 或 Ctrl + F5"
    log_warn "  - Mac: Cmd + Shift + R"
    log_warn "  - 或清除浏览器缓存后重新访问"
    log_warn ""
    log_info ""
    log_info "测试手动调整功能："
    log_info "  curl -X POST http://${SERVER_IP}:9527/api/rating/admin/adjust \\"
    log_info "    -H 'Content-Type: application/json' \\"
    log_info "    -H 'X-Operator-UID: admin' \\"
    log_info "    -d '{\"username\": \"user\", \"ratingChange\": -100, \"reason\": \"测试\"}'"
    log_info ""
    log_info "数据库迁移："
    log_info "  - 已执行: 001_add_last_view_time_fields.sql"
    log_info "  - 新增字段验证:"
    log_info "    mysql -h43.143.133.62 -uroot -phist2025 -e \\"
    log_info "      'SELECT COLUMN_NAME FROM INFORMATION_SCHEMA.COLUMNS \\"
    log_info "       WHERE TABLE_NAME=\"histcontest_register_registrations\" \\"
    log_info "       AND COLUMN_NAME IN (\"last_view_time\", \"admin_last_view_time\")' hoj"
    log_info ""
    log_info "代码变更："
    log_info "  - 新增: hoj-vue/src/views/oj/toolbox/Toolbox.vue"
    log_info "  - 新增: hoj-vue/src/views/admin/toolbox/ToolboxAdmin.vue"
    log_info "  - 修改: hoj-vue/src/components/oj/common/NavBar.vue (导航菜单)"
    log_info "  - 修改: hoj-vue/src/views/admin/Home.vue (管理菜单)"
    log_info "  - 修改: hoj-vue/src/router/ojRoutes.js (用户路由)"
    log_info "  - 修改: hoj-vue/src/router/adminRoutes.js (管理员路由)"
    log_info "=========================================="
}

# 主函数
main() {
    log_info "=========================================="
    log_info "HOJ2 + 报名系统自动化部署"
    log_info "=========================================="

    # 解析参数
    SKIP_BUILD=false
    SKIP_DB_MIGRATION=false
    DEPLOY_TARGET="all"

    while [[ $# -gt 0 ]]; do
        case $1 in
            --skip-build)
                log_warn "跳过镜像构建步骤"
                SKIP_BUILD=true
                shift
                ;;
            --skip-db)
                log_warn "跳过数据库迁移"
                SKIP_DB_MIGRATION=true
                shift
                ;;
            --only-frontend)
                log_warn "仅部署前端"
                DEPLOY_TARGET="frontend"
                shift
                ;;
            --only-backend)
                log_warn "仅部署后端服务"
                DEPLOY_TARGET="backend"
                shift
                ;;
            --only-registration)
                log_warn "仅部署报名系统"
                DEPLOY_TARGET="registration"
                shift
                ;;
            -h|--help)
                echo "用法: $0 [选项]"
                echo ""
                echo "选项:"
                echo "  --skip-build         跳过镜像构建步骤（使用已有镜像）"
                echo "  --skip-db            跳过数据库迁移（适用于已部署环境）"
                echo "  --only-frontend      仅部署前端服务"
                echo "  --only-backend       仅部署后端服务"
                echo "  --only-registration  仅部署报名系统"
                echo "  -h, --help           显示帮助信息"
                echo ""
                echo "示例:"
                echo "  $0                  # 完整部署"
                echo "  $0 --skip-db        # 跳过数据库迁移（快速部署）"
                echo "  $0 --only-frontend  # 仅更新前端"
                exit 0
                ;;
            *)
                log_error "未知参数: $1"
                echo "使用 -h 或 --help 查看帮助"
                exit 1
                ;;
        esac
    done

    # 显示部署计划
    log_info "部署计划："
    log_info "  - 构建镜像: $([ "$SKIP_BUILD" = true ] && echo '否' || echo '是')"
    log_info "  - 数据库迁移: $([ "$SKIP_DB_MIGRATION" = true ] && echo '跳过' || echo '自动检测')"
    log_info "  - 部署目标: $DEPLOY_TARGET"
    echo ""

    # 执行部署流程
    check_requirements

    if [ "$SKIP_BUILD" = false ]; then
        build_images
        verify_image_content
    fi

    save_images "$DEPLOY_TARGET"
    upload_images "$DEPLOY_TARGET"
    deploy_on_server "$DEPLOY_TARGET"
    cleanup
    cleanup_server
    cleanup_old_images
    show_result

    log_info "✓ 所有步骤完成！"
}

# 执行主函数
main "$@"
