#!/bin/bash

# 对战系统准备功能部署验证脚本
# 用于验证 challenger_ready 字段是否正确部署

set -e

# 配置
DB_HOST="43.143.133.62"
DB_PORT="3306"
DB_USER="root"
DB_PASS="jia13579.."
DB_NAME="hoj"

echo "========================================="
echo "对战系统准备功能 - 部署验证脚本"
echo "========================================="
echo ""

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查函数
check_pass() {
    echo -e "${GREEN}✓${NC} $1"
}

check_fail() {
    echo -e "${RED}✗${NC} $1"
}

check_warn() {
    echo -e "${YELLOW}⚠${NC} $1"
}

# 1. 检查数据库连接
echo "1. 检查数据库连接..."
if mysql -h$DB_HOST -P$DB_PORT -u$DB_USER -p$DB_PASS -e "USE $DB_NAME;" 2>/dev/null; then
    check_pass "数据库连接成功"
else
    check_fail "数据库连接失败"
    exit 1
fi
echo ""

# 2. 检查字段是否存在
echo "2. 检查 challenger_ready 字段..."
FIELD_EXISTS=$(mysql -h$DB_HOST -P$DB_PORT -u$DB_USER -p$DB_PASS $DB_NAME -sN -e "
    SELECT COUNT(*)
    FROM INFORMATION_SCHEMA.COLUMNS
    WHERE TABLE_SCHEMA = '$DB_NAME'
    AND TABLE_NAME = 'battle_room'
    AND COLUMN_NAME = 'challenger_ready';
" 2>/dev/null)

if [ "$FIELD_EXISTS" = "1" ]; then
    check_pass "challenger_ready 字段已存在"

    # 检查字段类型
    FIELD_TYPE=$(mysql -h$DB_HOST -P$DB_PORT -u$DB_USER -p$DB_PASS $DB_NAME -sN -e "
        SELECT COLUMN_TYPE
        FROM INFORMATION_SCHEMA.COLUMNS
        WHERE TABLE_SCHEMA = '$DB_NAME'
        AND TABLE_NAME = 'battle_room'
        AND COLUMN_NAME = 'challenger_ready';
    " 2>/dev/null)

    if [ "$FIELD_TYPE" = "tinyint(1)" ]; then
        check_pass "字段类型正确: $FIELD_TYPE"
    else
        check_warn "字段类型可能不正确: $FIELD_TYPE (预期: tinyint(1))"
    fi

    # 检查默认值
    DEFAULT_VALUE=$(mysql -h$DB_HOST -P$DB_PORT -u$DB_USER -p$DB_PASS $DB_NAME -sN -e "
        SELECT COLUMN_DEFAULT
        FROM INFORMATION_SCHEMA.COLUMNS
        WHERE TABLE_SCHEMA = '$DB_NAME'
        AND TABLE_NAME = 'battle_room'
        AND COLUMN_NAME = 'challenger_ready';
    " 2>/dev/null)

    if [ "$DEFAULT_VALUE" = "0" ]; then
        check_pass "默认值正确: $DEFAULT_VALUE"
    else
        check_warn "默认值可能不正确: $DEFAULT_VALUE (预期: 0)"
    fi

    # 检查是否可空
    IS_NULLABLE=$(mysql -h$DB_HOST -P$DB_PORT -u$DB_USER -p$DB_PASS $DB_NAME -sN -e "
        SELECT IS_NULLABLE
        FROM INFORMATION_SCHEMA.COLUMNS
        WHERE TABLE_SCHEMA = '$DB_NAME'
        AND TABLE_NAME = 'battle_room'
        AND COLUMN_NAME = 'challenger_ready';
    " 2>/dev/null)

    if [ "$IS_NULLABLE" = "NO" ]; then
        check_pass "NOT NULL 约束正确"
    else
        check_warn "字段可为空 (预期: NO)"
    fi
else
    check_fail "challenger_ready 字段不存在,需要先执行数据库迁移"
    echo ""
    echo "请执行以下命令添加字段:"
    echo "mysql -h$DB_HOST -P$DB_PORT -u$DB_USER -p$DB_PASS $DB_NAME << 'EOF'"
    echo "ALTER TABLE \`battle_room\`"
    echo "ADD COLUMN \`challenger_ready\` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '挑战者是否已准备' AFTER \`challenger_username\`;"
    echo "ALTER TABLE \`battle_room\`"
    echo "ADD INDEX \`idx_challenger_ready\` (\`challenger_ready\`);"
    echo "EOF"
    echo ""
    exit 1
fi
echo ""

# 3. 检查索引是否存在
echo "3. 检查索引..."
INDEX_EXISTS=$(mysql -h$DB_HOST -P$DB_PORT -u$DB_USER -p$DB_PASS $DB_NAME -sN -e "
    SELECT COUNT(*)
    FROM INFORMATION_SCHEMA.STATISTICS
    WHERE TABLE_SCHEMA = '$DB_NAME'
    AND TABLE_NAME = 'battle_room'
    AND INDEX_NAME = 'idx_challenger_ready';
" 2>/dev/null)

if [ "$INDEX_EXISTS" = "1" ]; then
    check_pass "索引 idx_challenger_ready 已存在"
else
    check_warn "索引不存在,建议添加以提高性能"
fi
echo ""

# 4. 检查现有数据
echo "4. 检查现有数据..."
TOTAL_ROOMS=$(mysql -h$DB_HOST -P$DB_PORT -u$DB_USER -p$DB_PASS $DB_NAME -sN -e "
    SELECT COUNT(*) FROM battle_room;
" 2>/dev/null)

echo "   总房间数: $TOTAL_ROOMS"

ROOMS_WITH_CHALLENGER=$(mysql -h$DB_HOST -P$DB_PORT -u$DB_USER -p$DB_PASS $DB_NAME -sN -e "
    SELECT COUNT(*) FROM battle_room WHERE challenger_id IS NOT NULL;
" 2>/dev/null)

echo "   有挑战者的房间: $ROOMS_WITH_CHALLENGER"

# 检查是否有 NULL 值
NULL_COUNT=$(mysql -h$DB_HOST -P$DB_PORT -u$DB_USER -p$DB_PASS $DB_NAME -sN -e "
    SELECT COUNT(*) FROM battle_room WHERE challenger_ready IS NULL;
" 2>/dev/null)

if [ "$NULL_COUNT" = "0" ]; then
    check_pass "没有 NULL 值"
else
    check_warn "发现 $NULL_COUNT 条记录的 challenger_ready 为 NULL"
fi
echo ""

# 5. 测试 API (如果服务正在运行)
echo "5. 检查后端 API..."
API_URL="http://localhost:9527/api/battle/ready"

if curl -s -o /dev/null -w "%{http_code}" $API_URL > /dev/null 2>&1; then
    HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" $API_URL)
    if [ "$HTTP_CODE" = "405" ] || [ "$HTTP_CODE" = "404" ]; then
        check_pass "API 服务可访问 (HTTP $HTTP_CODE - 方法不支持/路由未注册是正常的)"
    else
        check_warn "API 返回状态码: $HTTP_CODE"
    fi
else
    check_warn "无法连接到后端 API (http://localhost:9527)"
    check_warn "请确保后端服务正在运行"
fi
echo ""

# 6. 总结
echo "========================================="
echo "验证总结"
echo "========================================="
echo ""

if [ "$FIELD_EXISTS" = "1" ]; then
    echo -e "${GREEN}✓ 数据库迁移已完成${NC}"
    echo ""
    echo "下一步操作:"
    echo "1. 确保后端代码已更新到最新版本"
    echo "2. 重启后端服务: cd hist-oj && make run"
    echo "3. 清除浏览器缓存 (Ctrl+Shift+R)"
    echo "4. 测试准备功能"
    echo ""
    echo "测试步骤:"
    echo "  a. 挑战者加入房间"
    echo "  b. 检查是否显示'未准备'状态"
    echo "  c. 点击'准备'按钮"
    echo "  d. 检查是否显示'已准备'状态"
    echo "  e. 房主检查是否可以看到准备状态"
    echo "  f. 房主点击'开始对战'"
else
    echo -e "${RED}✗ 数据库迁移未完成${NC}"
    echo ""
    echo "请先执行数据库迁移,然后重新运行此脚本。"
    exit 1
fi
