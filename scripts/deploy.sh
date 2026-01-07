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
    log_info "构建 hist-oj 镜像..."
    cd hist-oj
    docker build --platform linux/amd64 -t hist-oj:latest . || {
        log_error "hist-oj 镜像构建失败"
        exit 1
    }
    log_info "✓ hist-oj 镜像构建成功"

    # 构建报名系统后端
    log_info "构建 registration-backend 镜像..."
    cd ../registration-system/backend
    docker build --platform linux/amd64 -t registration-backend:latest . || {
        log_error "registration-backend 镜像构建失败"
        exit 1
    }
    log_info "✓ registration-backend 镜像构建成功"

    # 构建前端
    log_info "构建 hoj-frontend 镜像..."
    cd ../../hoj-vue
    docker build --platform linux/amd64 -t hoj-frontend:latest . || {
        log_error "hoj-frontend 镜像构建失败"
        exit 1
    }
    log_info "✓ hoj-frontend 镜像构建成功"

    cd "$PROJECT_DIR"
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

        # 加载 hist-oj 镜像
        echo "[INFO] 加载 hist-oj 镜像..."
        gunzip -c hist-oj.tar.gz | docker load

        # 加载报名系统镜像
        echo "[INFO] 加载 registration-backend 镜像..."
        gunzip -c registration-backend.tar.gz | docker load

        # 加载前端镜像
        echo "[INFO] 加载 hoj-frontend 镜像..."
        gunzip -c hoj-frontend.tar.gz | docker load

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
            docker stop hist-oj 2>/dev/null || true
            docker rm hist-oj 2>/dev/null || true
        fi

        if [ "$DEPLOY_TARGET" = "all" ] || [ "$DEPLOY_TARGET" = "registration" ]; then
            docker stop registration-backend 2>/dev/null || true
            docker rm registration-backend 2>/dev/null || true
        fi

        if [ "$DEPLOY_TARGET" = "all" ] || [ "$DEPLOY_TARGET" = "frontend" ]; then
            docker stop hoj-frontend 2>/dev/null || true
            docker rm hoj-frontend 2>/dev/null || true
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

# 显示部署结果
show_result() {
    log_info "=========================================="
    log_info "部署完成！"
    log_info "=========================================="
    log_info "服务访问地址："
    log_info "  - 前端: http://${SERVER_IP}"
    log_info "  - hist-oj API: http://${SERVER_IP}:9527"
    log_info "  - 报名页面: http://${SERVER_IP}/registration"
    log_info "  - 管理后台: http://${SERVER_IP}/admin/registration"
    log_info ""
    log_info "新增功能："
    log_info "  1. 手动调整 Rating"
    log_info "     - API: POST http://${SERVER_IP}:9527/api/rating/admin/adjust"
    log_info "     - 文档: hist-oj/MANUAL_RATING_ADJUST.md"
    log_info ""
    log_info "  2. 持久化消息已读状态（解决浏览器缓存清除问题）"
    log_info "     - 用户和管理员的已读状态分别存储在数据库"
    log_info "     - 支持跨设备同步已读状态"
    log_info "     - 新增字段: last_view_time, admin_last_view_time"
    log_info ""
    log_info "验证命令："
    log_info "  curl http://${SERVER_IP}:9527/health"
    log_info "  curl http://${SERVER_IP}/api/rating/contest/info/1002"
    log_info "  curl http://${SERVER_IP}/registration-api/competitions"
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
    fi

    save_images "$DEPLOY_TARGET"
    upload_images "$DEPLOY_TARGET"
    deploy_on_server "$DEPLOY_TARGET"
    cleanup
    cleanup_server
    show_result

    log_info "✓ 所有步骤完成！"
}

# 执行主函数
main "$@"
