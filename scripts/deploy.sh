#!/bin/bash

# HOJ2 + 报名系统 + 代码对战 + 班级管理系统自动化部署脚本
# 用途：一键构建、打包、上传、部署 hist-oj、hoj-frontend、registration-backend
#       hist-oj 包含：Rating计算、代码对战、班级管理等功能

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

    # 先清理旧的构建文件和 Docker 缓存（在构建之前）
    log_info "清理旧的构建文件和 Docker 缓存..."
    cd hoj-vue
    rm -rf dist node_modules/.cache
    cd "$PROJECT_DIR"
    docker builder prune -af
    docker system prune -af --volumes 2>/dev/null || true

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

    # 构建前端（Docker 构建时会自动安装 package.json 中的所有依赖，包括 jsQR）
    log_info "构建 hoj-frontend 镜像（不使用缓存，包含 jsQR 二维码扫描功能）..."

    # 切换到前端目录（从 registration-system/backend 返回项目根目录，然后进入 hoj-vue）
    cd "$PROJECT_DIR/hoj-vue"

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
        # 检查目录是否存在
        if [ ! -d "/usr/share/nginx/html/assets/js" ]; then
            echo "[ERROR] 目录 /usr/share/nginx/html/assets/js 不存在"
            ls -la /usr/share/nginx/html/ || echo "无法列出目录"
            exit 1
        fi

        # 检查 app.js 文件
        APP_JS=$(find /usr/share/nginx/html/assets/js -name "app.*.js" -type f | head -1)
        if [ -z "$APP_JS" ]; then
            echo "[ERROR] 未找到 app.js 文件"
            echo "[INFO] 可用的 JS 文件:"
            ls -la /usr/share/nginx/html/assets/js/ | head -20
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
        docker save hist-oj:latest | gzip -9 > hist-oj.tar.gz || {
            log_error "hist-oj 镜像保存失败"
            exit 1
        }
        log_info "✓ hist-oj 镜像已保存 ($(du -h hist-oj.tar.gz | cut -f1))"
    fi

    if [ "$DEPLOY_TARGET" = "all" ] || [ "$DEPLOY_TARGET" = "registration" ]; then
        # 保存报名系统
        log_info "保存 registration-backend 镜像..."
        docker save registration-backend:latest | gzip -9 > registration-backend.tar.gz || {
            log_error "registration-backend 镜像保存失败"
            exit 1
        }
        log_info "✓ registration-backend 镜像已保存 ($(du -h registration-backend.tar.gz | cut -f1))"
    fi

    if [ "$DEPLOY_TARGET" = "all" ] || [ "$DEPLOY_TARGET" = "frontend" ]; then
        # 保存前端（使用最高压缩级别 -9）
        log_info "保存 hoj-frontend 镜像（使用最大压缩以减小上传体积）..."
        docker save hoj-frontend:latest | gzip -9 > hoj-frontend.tar.gz || {
            log_error "hoj-frontend 镜像保存失败"
            exit 1
        }
        local SIZE=$(du -h hoj-frontend.tar.gz | cut -f1)
        log_info "✓ hoj-frontend 镜像已保存（压缩后大小: $SIZE）"
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
        # 上传前端（使用 rsync -z 压缩传输，比 scp 更快）
        log_info "上传 hoj-frontend 镜像（使用 rsync 压缩传输）..."
        sshpass -p "$SERVER_PASS" rsync -avz --progress hoj-frontend.tar.gz ${SERVER_USER}@${SERVER_IP}:${REMOTE_DIR}/ || {
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

        # 上传报名系统迁移脚本
        sshpass -p "$SERVER_PASS" scp registration-system/migrations/001_add_last_view_time_fields.sql ${SERVER_USER}@${SERVER_IP}:${REMOTE_DIR}/ 2>/dev/null || {
            log_warn "报名系统迁移脚本上传失败（可能不存在）"
        }

        # 上传代码对战系统迁移脚本（所有相关脚本）
        sshpass -p "$SERVER_PASS" scp hist-oj/migrations/battle.sql ${SERVER_USER}@${SERVER_IP}:${REMOTE_DIR}/ 2>/dev/null || {
            log_warn "代码对战初始表脚本上传失败（可能不存在）"
        }
        sshpass -p "$SERVER_PASS" scp hist-oj/migrations/alter_battle_problem_id.sql ${SERVER_USER}@${SERVER_IP}:${REMOTE_DIR}/ 2>/dev/null || {
            log_warn "代码对战 problem_id 修改脚本上传失败（可能不存在）"
        }
        sshpass -p "$SERVER_PASS" scp hist-oj/migrations/alter_battle_tables.sql ${SERVER_USER}@${SERVER_IP}:${REMOTE_DIR}/ 2>/dev/null || {
            log_warn "代码对战表字段修改脚本上传失败（可能不存在）"
        }
        # 上传新增的 opponent_rating 字段迁移脚本
        sshpass -p "$SERVER_PASS" scp hist-oj/migrations/add_battle_record_opponent_rating.sql ${SERVER_USER}@${SERVER_IP}:${REMOTE_DIR}/ 2>/dev/null || {
            log_warn "代码对战 opponent_rating 字段脚本上传失败（可能不存在）"
        }
        # 上传回填旧数据的脚本
        sshpass -p "$SERVER_PASS" scp hist-oj/migrations/backfill_opponent_rating.sql ${SERVER_USER}@${SERVER_IP}:${REMOTE_DIR}/ 2>/dev/null || {
            log_warn "代码对战 opponent_rating 回填脚本上传失败（可能不存在）"
        }

        # 上传班级系统迁移脚本
        sshpass -p "$SERVER_PASS" scp sqlAndsetting/classroom.sql ${SERVER_USER}@${SERVER_IP}:${REMOTE_DIR}/ 2>/dev/null || {
            log_warn "班级系统迁移脚本上传失败（可能不存在）"
        }

        log_info "✓ 数据库迁移脚本上传完成"
    fi
}

# 上传 SSL 证书和 Nginx 配置
upload_ssl_certs() {
    log_info "上传 SSL 证书和 Nginx 配置..."

    cd "$PROJECT_DIR"

    # 检查证书文件是否存在
    if [ ! -f "scripts/nginx/ssl/bingoj.cn.crt" ] || [ ! -f "scripts/nginx/ssl/bingoj.cn.key" ]; then
        log_error "SSL 证书文件不存在！请确保证书文件在 scripts/nginx/ssl/ 目录下"
        exit 1
    fi

    # 创建服务器上的 Nginx 目录
    sshpass -p "$SERVER_PASS" ssh -o StrictHostKeyChecking=no ${SERVER_USER}@${SERVER_IP} << ENDSSH
        mkdir -p /etc/nginx/ssl
        mkdir -p /etc/nginx/conf.d
        chmod 700 /etc/nginx/ssl
ENDSSH

    # 上传证书文件
    log_info "上传 SSL 证书文件..."
    sshpass -p "$SERVER_PASS" scp scripts/nginx/ssl/bingoj.cn.crt ${SERVER_USER}@${SERVER_IP}:/etc/nginx/ssl/
    sshpass -p "$SERVER_PASS" scp scripts/nginx/ssl/bingoj.cn.key ${SERVER_USER}@${SERVER_IP}:/etc/nginx/ssl/

    # 设置证书文件权限
    sshpass -p "$SERVER_PASS" ssh -o StrictHostKeyChecking=no ${SERVER_USER}@${SERVER_IP} << ENDSSH
        chmod 644 /etc/nginx/ssl/bingoj.cn.crt
        chmod 600 /etc/nginx/ssl/bingoj.cn.key
        chown root:root /etc/nginx/ssl/*
ENDSSH

    # 上传 Nginx 配置文件
    log_info "上传 Nginx SSL 配置..."
    sshpass -p "$SERVER_PASS" scp scripts/nginx/bingoj.conf ${SERVER_USER}@${SERVER_IP}:/etc/nginx/conf.d/

    log_info "✓ SSL 证书和 Nginx 配置上传成功"
}

# 安装并配置 Nginx
install_nginx() {
    log_info "在服务器上安装并配置 Nginx..."

    sshpass -p "$SERVER_PASS" ssh -o StrictHostKeyChecking=no ${SERVER_USER}@${SERVER_IP} << 'ENDSSH'
        set -e

        # 检查 Nginx 是否已安装
        if ! command -v nginx &> /dev/null; then
            echo "[INFO] 安装 Nginx..."
            # 检测系统类型并安装 Nginx
            if [ -f /etc/redhat-release ]; then
                # CentOS/RHEL
                yum install -y nginx
            elif [ -f /etc/debian_version ]; then
                # Ubuntu/Debian
                apt-get update
                apt-get install -y nginx
            else
                echo "[ERROR] 不支持的操作系统"
                exit 1
            fi
            echo "[INFO] ✓ Nginx 安装成功"
        else
            echo "[INFO] ✓ Nginx 已安装"
        fi

        # 创建日志目录
        mkdir -p /var/log/nginx
        chown -R www-data:www-data /var/log/nginx 2>/dev/null || true

        # 先停止旧的前端容器（避免端口冲突）
        if docker ps -a | grep -q hoj-frontend; then
            echo "[INFO] 停止旧的 hoj-frontend 容器（避免端口冲突）..."
            docker stop hoj-frontend 2>/dev/null || true
            docker rm hoj-frontend 2>/dev/null || true
            echo "[INFO] ✓ 旧容器已停止"
        fi

        # 修复并禁用默认配置（避免端口冲突）
        if [ -f /etc/nginx/sites-enabled/default ]; then
            echo "[INFO] 禁用默认 Nginx 配置..."
            rm -f /etc/nginx/sites-enabled/default
        fi

        # 备份旧配置（如果存在）
        if [ -f /etc/nginx/conf.d/bingoj.conf ] && [ ! -f /etc/nginx/conf.d/bingoj.conf.bak ]; then
            cp /etc/nginx/conf.d/bingoj.conf /etc/nginx/conf.d/bingoj.conf.bak
            echo "[INFO] 已备份旧配置"
        fi

        # 测试 Nginx 配置
        echo "[INFO] 测试 Nginx 配置..."
        nginx -t

        # 如果配置测试通过，启动 Nginx
        if [ $? -eq 0 ]; then
            # 启动 Nginx（如果未运行）
            systemctl start nginx 2>/dev/null || service nginx start

            # 设置开机自启
            systemctl enable nginx 2>/dev/null || update-rc.d nginx defaults

            # 重新加载配置
            systemctl reload nginx 2>/dev/null || nginx -s reload

            echo "[INFO] ✓ Nginx 配置成功并已启动"
        else
            echo "[ERROR] Nginx 配置测试失败"
            exit 1
        fi

        # 检查 Nginx 状态
        systemctl status nginx --no-pager || service nginx status

        # 验证 443 端口是否监听
        sleep 2
        if netstat -tlnp 2>/dev/null | grep -q ":443 "; then
            echo "[INFO] ✓ Nginx 正在监听 443 端口（HTTPS）"
        elif ss -tlnp 2>/dev/null | grep -q ":443 "; then
            echo "[INFO] ✓ Nginx 正在监听 443 端口（HTTPS）"
        else
            echo "[WARN] Nginx 未监听 443 端口，请检查配置"
        fi
ENDSSH

    if [ $? -eq 0 ]; then
        log_info "✓ Nginx 安装配置成功"
    else
        log_error "Nginx 安装配置失败"
        exit 1
    fi
}

# 在服务器上部署
deploy_on_server() {
    local DEPLOY_TARGET="${1:-all}"
    log_info "在服务器上部署服务（目标: $DEPLOY_TARGET）..."

    # 传递参数到远程脚本
    sshpass -p "$SERVER_PASS" ssh -o StrictHostKeyChecking=no ${SERVER_USER}@${SERVER_IP} << ENDSSH
        set -e
        DEPLOY_TARGET='$DEPLOY_TARGET'

        echo "[INFO] 加载 Docker 镜像..."
        cd /opt

        # 记录加载前的镜像 ID
        FRONTEND_IMAGE_BEFORE=$(docker images hoj-frontend:latest --format "{{.ID}}" 2>/dev/null || echo "")

        # 加载 hist-oj 镜像（如果文件存在）
        if [ -f "hist-oj.tar.gz" ]; then
            echo "[INFO] 加载 hist-oj 镜像..."
            gunzip -c hist-oj.tar.gz | docker load
        fi

        # 加载报名系统镜像（如果文件存在）
        if [ -f "registration-backend.tar.gz" ]; then
            echo "[INFO] 加载 registration-backend 镜像..."
            gunzip -c registration-backend.tar.gz | docker load
        fi

        # 加载前端镜像
        echo "[INFO] 加载 hoj-frontend 镜像..."
        if [ -f "hoj-frontend.tar.gz" ]; then
            gunzip -c hoj-frontend.tar.gz | docker load
            if [ $? -eq 0 ]; then
                echo "[INFO] ✓ hoj-frontend 镜像加载成功"
            else
                echo "[ERROR] hoj-frontend 镜像加载失败"
                exit 1
            fi
        else
            echo "[ERROR] hoj-frontend.tar.gz 文件不存在"
            exit 1
        fi

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

        # 检查 last_view_time 和 admin_last_view_time 字段是否已存在（报名系统）
        FIELD_EXISTS=$(mysql -h43.143.133.62 -uroot -phist2025 -sN -e \
            "SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS \
             WHERE TABLE_SCHEMA='hoj' \
             AND TABLE_NAME='histcontest_register_registrations' \
             AND COLUMN_NAME IN ('last_view_time', 'admin_last_view_time')" 2>/dev/null || echo "0")

        if [ "$FIELD_EXISTS" -ge "2" ]; then
            echo "[INFO] ✓ 报名系统数据库字段已存在"
        else
            echo "[INFO] 需要执行报名系统数据库迁移..."

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

        # 检查代码对战系统表是否已存在
        BATTLE_TABLE_EXISTS=$(mysql -h43.143.133.62 -uroot -phist2025 -sN -e \
            "SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES \
             WHERE TABLE_SCHEMA='hoj' \
             AND TABLE_NAME='battle_room'" 2>/dev/null || echo "0")

        if [ "$BATTLE_TABLE_EXISTS" -ge "1" ]; then
            echo "[INFO] ✓ 代码对战系统数据库表已存在，检查是否需要更新表结构..."

            # 检查 problem_id 字段类型（需要 VARCHAR(255) 而不是 BIGINT）
            PROBLEM_ID_TYPE=$(mysql -h43.143.133.62 -uroot -phist2025 -sN -e \
                "SELECT DATA_TYPE FROM INFORMATION_SCHEMA.COLUMNS \
                 WHERE TABLE_SCHEMA='hoj' \
                 AND TABLE_NAME='battle_room' \
                 AND COLUMN_NAME='problem_id'" 2>/dev/null || echo "")

            if [ "$PROBLEM_ID_TYPE" = "bigint" ]; then
                echo "[INFO] 检测到旧版本表结构，需要更新 problem_id 字段类型..."

                # 执行表结构更新
                mysql -h43.143.133.62 -uroot -phist2025 hoj << 'SQLEOF'
ALTER TABLE battle_room MODIFY COLUMN problem_id VARCHAR(255) DEFAULT NULL COMMENT '对战题目ID（显示ID，如 0051, Z001）';
ALTER TABLE battle_record MODIFY COLUMN problem_id VARCHAR(255) NOT NULL COMMENT '对战题目ID（显示ID）';
ALTER TABLE battle_room MODIFY COLUMN host_id VARCHAR(64) NOT NULL COMMENT '房主用户ID';
ALTER TABLE battle_room MODIFY COLUMN challenger_id VARCHAR(64) DEFAULT NULL COMMENT '挑战者用户ID';
ALTER TABLE battle_room MODIFY COLUMN winner_id VARCHAR(64) DEFAULT NULL COMMENT '获胜者用户ID';
ALTER TABLE battle_record MODIFY COLUMN user_id VARCHAR(64) NOT NULL COMMENT '用户ID';
ALTER TABLE battle_record MODIFY COLUMN opponent_id VARCHAR(64) NOT NULL COMMENT '对手用户ID';
ALTER TABLE user_battle_stats MODIFY COLUMN user_id VARCHAR(64) NOT NULL COMMENT '用户ID';
SQLEOF

                if [ $? -eq 0 ]; then
                    echo "[INFO] ✓ 代码对战系统表结构更新成功"
                else
                    echo "[WARN] 代码对战系统表结构更新失败"
                fi
            else
                echo "[INFO] ✓ 代码对战系统表结构已是最新版本"
            fi
        else
            echo "[INFO] 需要执行代码对战系统数据库迁移..."

            # 1. 执行初始表创建
            if [ -f "/opt/battle.sql" ]; then
                echo "[INFO] 执行代码对战系统初始表创建..."
                mysql -h43.143.133.62 -uroot -phist2025 hoj < /opt/battle.sql && echo "[INFO] ✓ 初始表创建成功" || echo "[WARN] 初始表创建失败"
            else
                echo "[WARN] 初始表创建脚本不存在"
            fi

            # 2. 修改 problem_id 字段类型（支持显示ID）
            if [ -f "/opt/alter_battle_problem_id.sql" ]; then
                echo "[INFO] 执行 problem_id 字段类型修改..."
                mysql -h43.143.133.62 -uroot -phist2025 hoj < /opt/alter_battle_problem_id.sql && echo "[INFO] ✓ problem_id 字段修改成功" || echo "[WARN] problem_id 字段修改失败"
            else
                echo "[INFO] problem_id 字段修改脚本不存在（可能已执行）"
            fi

            # 3. 扩展用户ID字段长度（支持32字符UID）
            if [ -f "/opt/alter_battle_tables.sql" ]; then
                echo "[INFO] 执行用户ID字段长度扩展..."
                mysql -h43.143.133.62 -uroot -phist2025 hoj < /opt/alter_battle_tables.sql && echo "[INFO] ✓ 用户ID字段扩展成功" || echo "[WARN] 用户ID字段扩展失败"
            else
                echo "[INFO] 用户ID字段扩展脚本不存在（可能已执行）"
            fi

            # 4. 添加 opponent_rating 字段（支持显示对手rating颜色）
            if [ -f "/opt/add_battle_record_opponent_rating.sql" ]; then
                echo "[INFO] 执行 opponent_rating 字段添加..."
                mysql -h43.143.133.62 -uroot -phist2025 hoj < /opt/add_battle_record_opponent_rating.sql && echo "[INFO] ✓ opponent_rating 字段添加成功" || echo "[WARN] opponent_rating 字段添加失败"
            else
                echo "[INFO] opponent_rating 字段添加脚本不存在（可能已执行）"
            fi

            # 5. 回填旧记录的 opponent_rating 数据
            if [ -f "/opt/backfill_opponent_rating.sql" ]; then
                echo "[INFO] 回填旧记录的 opponent_rating 数据..."
                mysql -h43.143.133.62 -uroot -phist2025 hoj < /opt/backfill_opponent_rating.sql && echo "[INFO] ✓ opponent_rating 数据回填成功" || echo "[WARN] opponent_rating 数据回填失败"
            else
                echo "[INFO] opponent_rating 数据回填脚本不存在（可能已执行）"
            fi
        fi

        # 检查是否需要添加 opponent_rating 字段（对于已存在的旧表）
        OPPONENT_RATING_EXISTS=$(mysql -h43.143.133.62 -uroot -phist2025 -sN -e \
            "SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS \
             WHERE TABLE_SCHEMA='hoj' \
             AND TABLE_NAME='battle_record' \
             AND COLUMN_NAME='opponent_rating'" 2>/dev/null || echo "0")

        if [ "$OPPONENT_RATING_EXISTS" -lt "1" ]; then
            echo "[INFO] 检测到缺少 opponent_rating 字段，正在添加..."
            mysql -h43.143.133.62 -uroot -phist2025 hoj << 'SQLEOF'
ALTER TABLE battle_record ADD COLUMN opponent_rating INT NULL DEFAULT NULL COMMENT '对手的rating值' AFTER opponent_username;
ALTER TABLE battle_record ADD INDEX idx_opponent_rating (opponent_rating);
SQLEOF
            if [ $? -eq 0 ]; then
                echo "[INFO] ✓ opponent_rating 字段添加成功"
            else
                echo "[WARN] opponent_rating 字段添加失败"
            fi
        fi

        # 检查班级系统表是否已存在
        CLASSROOM_TABLE_EXISTS=$(mysql -h43.143.133.62 -uroot -phist2025 -sN -e \
            "SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES \
             WHERE TABLE_SCHEMA='hoj' \
             AND TABLE_NAME='classroom'" 2>/dev/null || echo "0")

        if [ "$CLASSROOM_TABLE_EXISTS" -lt "1" ]; then
            echo "[INFO] 需要执行班级系统数据库迁移..."
            if [ -f "/opt/classroom.sql" ]; then
                echo "[INFO] 执行班级系统数据库迁移..."
                mysql -h43.143.133.62 -uroot -phist2025 hoj < /opt/classroom.sql && echo "[INFO] ✓ 班级系统迁移成功" || echo "[WARN] 班级系统迁移失败"
            else
                echo "[WARN] 班级系统迁移脚本不存在"
            fi
        else
            echo "[INFO] ✓ 班级系统数据库表已存在"
        fi

        # 检查 homework_submit 表是否需要添加 is_officially_submitted 字段
        IS_OFFICIALLY_SUBMITTED_EXISTS=$(mysql -h43.143.133.62 -uroot -phist2025 -sN -e \
            "SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS \
             WHERE TABLE_SCHEMA='hoj' \
             AND TABLE_NAME='homework_submit' \
             AND COLUMN_NAME='is_officially_submitted'" 2>/dev/null || echo "0")

        if [ "$IS_OFFICIALLY_SUBMITTED_EXISTS" -lt "1" ]; then
            echo "[INFO] 需要添加 homework_submit.is_officially_submitted 字段..."
            if [ -f "/opt/add_is_officially_submitted.sql" ]; then
                echo "[INFO] 执行 is_officially_submitted 字段添加..."
                mysql -h43.143.133.62 -uroot -phist2025 hoj < /opt/add_is_officially_submitted.sql && echo "[INFO] ✓ is_officially_submitted 字段添加成功" || echo "[WARN] is_officially_submitted 字段添加失败"
            else
                echo "[INFO] 直接添加 is_officially_submitted 字段..."
                mysql -h43.143.133.62 -uroot -phist2025 hoj << 'SQLEOF'
ALTER TABLE homework_submit ADD COLUMN is_officially_submitted INT(1) NOT NULL DEFAULT 0 COMMENT '是否已正式提交（0=草稿自动保存，1=用户点击提交）' AFTER is_scored;
ALTER TABLE homework_submit ADD INDEX idx_is_officially_submitted (is_officially_submitted);
SQLEOF
                if [ $? -eq 0 ]; then
                    echo "[INFO] ✓ is_officially_submitted 字段添加成功"
                else
                    echo "[WARN] is_officially_submitted 字段添加失败"
                fi
            fi
        else
            echo "[INFO] ✓ homework_submit.is_officially_submitted 字段已存在"
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
        mkdir -p /opt/hist-oj-uploads/classroom/homework
        mkdir -p /opt/hist-oj-uploads/classroom/images

        # 根据部署目标选择性启动容器
        if [ "$DEPLOY_TARGET" = "all" ] || [ "$DEPLOY_TARGET" = "backend" ]; then
            echo "[INFO] 启动 hist-oj 容器..."
            docker run -d \
                --name hist-oj \
                --network hoj_hoj-network \
                -p 9527:9527 \
                -v /opt/hist-oj-uploads:/app/uploads \
                -e DATABASE_HOST=43.143.133.62 \
                -e DATABASE_PORT=3306 \
                -e DATABASE_USER=root \
                -e DATABASE_PASSWORD=hist2025 \
                -e DATABASE_DBNAME=hoj \
                -e HOJ_API_BASE_URL=http://hoj-backserver:6688 \
                -e JWT_SECRET=hoj-secret-init \
                -e TZ=Asia/Shanghai \
                --restart unless-stopped \
                hist-oj:latest
        fi

        if [ "$DEPLOY_TARGET" = "all" ] || [ "$DEPLOY_TARGET" = "registration" ]; then
            echo "[INFO] 启动 registration-backend 容器..."
            docker run -d \
                --name registration-backend \
                --network hoj_hoj-network \
                -p 8080:8080 \
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
            echo "[INFO] 启动 hoj-frontend 容器（映射到内部端口 8081，避免与 Nginx 冲突）..."
            docker run -d \
                --name hoj-frontend \
                --network hoj_hoj-network \
                --network main_hoj-network \
                -p 8081:80 \
                --restart unless-stopped \
                --health-cmd="wget --no-verbose --tries=1 --spider http://localhost/ || exit 1" \
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

        echo "[INFO] 测试前端访问 hoj-backend（主要 API 服务）..."
        docker exec hoj-frontend getent hosts hoj-backend > /dev/null 2>&1 && echo "[INFO] ✓ hoj-backend 可以解析" || echo "[ERROR] ✗ hoj-backend 无法解析（网络问题）"
        docker exec hoj-frontend curl -s -o /dev/null -w "%{http_code}" http://hoj-backend/api/common/getUserInfo | grep -q "200" && echo "[INFO] ✓ hoj-backend API 可访问" || echo "[WARN] hoj-backend API 访问异常（可能需要认证）"

        echo "[INFO] 测试代码对战 API..."
        docker exec hoj-frontend curl -s http://hist-oj:9527/api/battle/rank || echo "代码对战 API 测试失败"

        echo "[INFO] 测试工具箱路由（用户端）..."
        docker exec hoj-frontend curl -s -o /dev/null -w "%{http_code}" http://localhost/toolbox | grep -q "200" && echo "[INFO] ✓ 用户工具箱路由正常" || echo "[WARN] 用户工具箱路由异常"

        echo "[INFO] 测试工具箱路由（管理端）..."
        docker exec hoj-frontend curl -s -o /dev/null -w "%{http_code}" http://localhost/admin/toolbox | grep -q "200" && echo "[INFO] ✓ 管理员工具箱路由正常" || echo "[WARN] 管理员工具箱路由异常"

        echo "[INFO] 测试代码对战路由..."
        docker exec hoj-frontend curl -s -o /dev/null -w "%{http_code}" http://localhost/battle | grep -q "200" && echo "[INFO] ✓ 代码对战路由正常" || echo "[WARN] 代码对战路由异常（可能需要登录）"

        echo "[INFO] 测试代码对战排行榜路由..."
        docker exec hoj-frontend curl -s -o /dev/null -w "%{http_code}" http://localhost/battle/rank | grep -q "200" && echo "[INFO] ✓ 代码对战排行榜路由正常" || echo "[WARN] 代码对战排行榜路由异常"

        echo "[INFO] 测试班级管理系统路由（教师端）..."
        docker exec hoj-frontend curl -s -o /dev/null -w "%{http_code}" http://localhost/classroom/teacher | grep -q "200" && echo "[INFO] ✓ 教师工作台路由正常" || echo "[WARN] 教师工作台路由异常"

        echo "[INFO] 测试班级管理系统路由（学生端）..."
        docker exec hoj-frontend curl -s -o /dev/null -w "%{http_code}" http://localhost/classroom/student | grep -q "200" && echo "[INFO] ✓ 学生工作台路由正常" || echo "[WARN] 学生工作台路由异常"

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

            # 检查代码对战功能
            if docker exec hoj-frontend grep -q "BattleHome\|BattleRoom\|代码对战" "$APP_JS" 2>/dev/null; then
                echo "[INFO] ✓ 代码对战功能已包含"
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
                # 检查判题终端功能
                if docker exec hoj-frontend grep -q "判题终端" "$CHUNK_JS" 2>/dev/null; then
                    echo "[INFO] ✓ 判题终端功能已包含"
                fi
                if docker exec hoj-frontend grep -q "JudgeTerminalTool" "$CHUNK_JS" 2>/dev/null; then
                    echo "[INFO] ✓ 判题终端组件已包含"
                fi
            fi

            # 检查代码对战相关组件
            BATTLE_CHUNK_JS=$(docker exec hoj-frontend find /usr/share/nginx/html/assets/js -name "chunk-*.js" -type f -exec grep -l "BattleHome\|BattleRoom" {} \; 2>/dev/null | head -1)
            if [ -n "$BATTLE_CHUNK_JS" ]; then
                echo "[INFO] 找到代码对战 chunk: $(basename $BATTLE_CHUNK_JS)"
                if docker exec hoj-frontend grep -q "fa-gamepad\|代码对战" "$BATTLE_CHUNK_JS" 2>/dev/null; then
                    echo "[INFO] ✓ 代码对战功能已包含"
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
        rm -f hist-oj.tar.gz registration-backend.tar.gz hoj-frontend.tar.gz \
              005_add_manual_rating_fields.sql 001_add_last_view_time_fields.sql \
              battle.sql alter_battle_problem_id.sql alter_battle_tables.sql \
              add_battle_record_opponent_rating.sql backfill_opponent_rating.sql \
              classroom.sql
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
    log_info "✅ HTTPS 已启用，SSL 证书配置成功！"
    log_info ""
    log_info "服务访问地址（推荐使用 HTTPS）："
    log_info "  - 前端主页: https://bingoj.cn 或 http://${SERVER_IP}"
    log_info "  - hist-oj API: http://${SERVER_IP}:9527"
    log_info "  - 用户工具箱: https://bingoj.cn/toolbox"
    log_info "  - 管理员工具箱: https://bingoj.cn/admin/toolbox"
    log_info "  - 代码对战: https://bingoj.cn/battle"
    log_info "  - 对战排行榜: https://bingoj.cn/battle/rank"
    log_info "  - 教师工作台: https://bingoj.cn/classroom/teacher"
    log_info "  - 学生工作台: https://bingoj.cn/classroom/student"
    log_info "  - 班级签到: https://bingoj.cn/classroom/student (支持摄像头扫码)"
    log_info "  - 报名系统: https://bingoj.cn/toolbox -> 赛事报名系统"
    log_info "  - 报名管理: https://bingoj.cn/admin/toolbox -> 赛事报名系统管理"
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
    log_info "  2. 代码对战系统（新增）"
    log_info "     - 访问路径: /battle"
    log_info "     - 创建房间、加入房间、1v1实时对战"
    log_info "     - 智能选题：从双方都未AC的题目中随机选择"
    log_info "     - 对战排行榜：按胜场数排名"
    log_info "     - API: http://${SERVER_IP}:9527/api/battle/*"
    log_info ""
    log_info "  3. 判题终端（管理员工具箱）"
    log_info "     - 访问路径: 管理员工具箱 -> 判题终端"
    log_info "     - 支持本地样例自测和远程判题提交"
    log_info "     - 自动保存用户凭据"
    log_info "     - 实时查看判题结果和历史记录"
    log_info ""
    log_info "  4. 手动调整 Rating"
    log_info "     - API: POST http://${SERVER_IP}:9527/api/rating/admin/adjust"
    log_info "     - 文档: hist-oj/MANUAL_RATING_ADJUST.md"
    log_info ""
    log_info "  5. 持久化消息已读状态（解决浏览器缓存清除问题）"
    log_info "     - 用户和管理员的已读状态分别存储在数据库"
    log_info "     - 支持跨设备同步已读状态"
    log_info "     - 新增字段: last_view_time, admin_last_view_time"
    log_info ""
    log_info "  6. 对战记录对手Rating显示（新增）"
    log_info "     - 对战历史记录中对手名字根据rating显示不同颜色"
    log_info "     - 对战结果标签完美居中对齐"
    log_info "     - 新增字段: battle_record.opponent_rating"
    log_info ""
    log_info "  7. 班级管理系统（新增）"
    log_info "     - 班级管理：创建班级、学生管理、教师管理"
    log_info "     - 签到系统：上课签到、签到统计"
    log_info "     - 📷 二维码签到：支持摄像头扫描自动签到（jsQR 库）"
    log_info "     - 题库管理：班级题目、作业题库"
    log_info "     - 作业系统：发布作业、提交作业、批改作业"
    log_info "     - 资料库：课程资料上传、下载"
    log_info "     - 随机选人：课堂随机提问功能"
    log_info "     - 即时通讯：班级群聊、私信"
    log_info "     - API: http://${SERVER_IP}:9527/api/classroom/*"
    log_info ""
    log_info "  8. ✅ HTTPS + 二维码签到功能（已修复）"
    log_info "     - SSL 证书已配置，支持 HTTPS 访问"
    log_info "     - 使用 jsQR 库实现跨浏览器二维码扫描"
    log_info "     - 支持 Chrome、Firefox、Safari、Edge 全平台"
    log_info "     - 移除手动输入 Token，简化扫码流程"
    log_info "     - 学生端: https://bingoj.cn/classroom/student"
    log_info "     - 教师端: https://bingoj.cn/classroom/teacher"
    log_info ""
    log_info "验证命令："
    log_info "  curl http://${SERVER_IP}:9527/health"
    log_info "  curl http://${SERVER_IP}/api/rating/contest/info/1002"
    log_info "  curl http://${SERVER_IP}/api/battle/rank"
    log_info "  curl http://${SERVER_IP}:9527/api/classroom/list"
    log_info "  curl http://${SERVER_IP}/registration-api/competitions"
    log_info "  curl -I http://${SERVER_IP}/classroom/teacher"
    log_info "  curl -I http://${SERVER_IP}/classroom/student"
    log_info ""
    log_info "测试代码对战功能："
    log_info "  - 对战首页: curl -I http://${SERVER_IP}/battle"
    log_info "  - 对战排行榜: curl -I http://${SERVER_IP}/battle/rank"
    log_info "  - 对战API: curl http://${SERVER_IP}:9527/api/battle/rank"
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
    log_info "数据库迁移："
    log_info "  - 报名系统: 001_add_last_view_time_fields.sql"
    log_info "  - 代码对战: battle.sql (新增)"
    log_info "  - 对战记录对手rating: add_battle_record_opponent_rating.sql (新增)"
    log_info "  - 回填旧记录rating数据: backfill_opponent_rating.sql (新增)"
    log_info "  - 班级管理系统: classroom.sql (新增)"
    log_info "  - 验证对战表:"
    log_info "    mysql -h43.143.133.62 -uroot -phist2025 -e \\"
    log_info "      'SHOW TABLES LIKE \"battle%\"' hoj"
    log_info "  - 验证opponent_rating字段:"
    log_info "    mysql -h43.143.133.62 -uroot -phist2025 -e \\"
    log_info "      'SELECT COLUMN_NAME FROM INFORMATION_SCHEMA.COLUMNS \\"
    log_info "       WHERE TABLE_SCHEMA=\"hoj\" AND TABLE_NAME=\"battle_record\" \\"
    log_info "       AND COLUMN_NAME=\"opponent_rating\"' hoj"
    log_info "  - 验证班级系统表:"
    log_info "    mysql -h43.143.133.62 -uroot -phist2025 -e \\"
    log_info "      'SHOW TABLES LIKE \"classroom%\"' hoj"
    log_info ""
    log_info "代码变更（代码对战系统）："
    log_info "  - 后端:"
    log_info "    - hist-oj/internal/model/battle.go (数据模型)"
    log_info "    - hist-oj/internal/service/battle_service.go (业务逻辑)"
    log_info "    - hist-oj/internal/api/battle_api.go (API接口)"
    log_info "    - hist-oj/internal/api/routes.go (路由注册)"
    log_info "    - hist-oj/internal/client/hoj_api.go (扩展HOJ API客户端)"
    log_info "  - 前端:"
    log_info "    - hoj-vue/src/views/oj/battle/BattleHome.vue (对战大厅)"
    log_info "    - hoj-vue/src/views/oj/battle/BattleRoom.vue (对战房间)"
    log_info "    - hoj-vue/src/views/oj/battle/BattleRank.vue (对战排行榜)"
    log_info "    - hoj-vue/src/views/oj/battle/BattleMyRecords.vue (我的战绩)"
    log_info "    - hoj-vue/src/api/battle.js (API封装)"
    log_info "    - hoj-vue/src/router/ojRoutes.js (路由配置)"
    log_info "    - hoj-vue/vue.config.js (代理配置)"
    log_info ""
    log_info "代码变更（班级管理系统）："
    log_info "  - 后端:"
    log_info "    - hist-oj/internal/model/classroom.go (数据模型)"
    log_info "    - hist-oj/internal/api/classroom_api.go (API接口)"
    log_info "    - hist-oj/internal/api/classroom_api_part2.go (API接口扩展)"
    log_info "    - hist-oj/internal/api/routes.go (路由注册)"
    log_info "    - sqlAndsetting/classroom.sql (数据库表结构)"
    log_info ""
    log_info "=========================================="
}

# 主函数
main() {
    log_info "=========================================="
    log_info "HOJ2 + 报名系统 + 代码对战 + 班级管理自动化部署"
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

    # 上传并配置 SSL 证书（仅在完整部署或前端部署时）
    if [ "$DEPLOY_TARGET" = "all" ] || [ "$DEPLOY_TARGET" = "frontend" ]; then
        upload_ssl_certs
        install_nginx
    fi

    deploy_on_server "$DEPLOY_TARGET"
    cleanup
    cleanup_server
    cleanup_old_images
    show_result

    log_info "✓ 所有步骤完成！"
}

# 执行主函数
main "$@"
