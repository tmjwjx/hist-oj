#!/bin/bash

# ========================================
# 对战系统 - 准备功能部署脚本
# 在服务器上直接执行此脚本
# ========================================

# 配置信息
DB_HOST="localhost"
DB_PORT="3306"
DB_USER="root"
DB_PASS="n208966737"
DB_NAME="hoj"

echo "========================================="
echo "对战系统 - 准备功能部署"
echo "========================================="
echo ""

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 1. 测试数据库连接
echo "1. 测试数据库连接..."
if mysql -h$DB_HOST -P$DB_PORT -u$DB_USER -p$DB_PASS -e "USE $DB_NAME; SELECT '✓ 数据库连接成功' AS status;" 2>/dev/null; then
    echo -e "${GREEN}✓ 数据库连接成功${NC}"
else
    echo -e "${RED}✗ 数据库连接失败${NC}"
    echo "请检查以下配置:"
    echo "  主机: $DB_HOST"
    echo "  端口: $DB_PORT"
    echo "  用户: $DB_USER"
    echo "  数据库: $DB_NAME"
    exit 1
fi
echo ""

# 2. 检查字段是否已存在
echo "2. 检查 challenger_ready 字段..."
FIELD_EXISTS=$(mysql -h$DB_HOST -P$DB_PORT -u$DB_USER -p$DB_PASS $DB_NAME -sN -e "
    SELECT COUNT(*)
    FROM INFORMATION_SCHEMA.COLUMNS
    WHERE TABLE_SCHEMA = '$DB_NAME'
    AND TABLE_NAME = 'battle_room'
    AND COLUMN_NAME = 'challenger_ready';
" 2>/dev/null)

if [ "$FIELD_EXISTS" = "1" ]; then
    echo -e "${YELLOW}⚠ 字段已存在,跳过添加${NC}"
    echo ""
else
    echo "  字段不存在,开始添加..."
    echo ""

    # 3. 添加字段
    echo "3. 添加 challenger_ready 字段..."
    mysql -h$DB_HOST -P$DB_PORT -u$DB_USER -p$DB_PASS $DB_NAME << 'EOF' 2>/dev/null
ALTER TABLE battle_room
ADD COLUMN challenger_ready TINYINT(1) NOT NULL DEFAULT 0
COMMENT '挑战者是否已准备 (0-未准备, 1-已准备)'
AFTER challenger_username;
EOF

    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ 字段添加成功${NC}"
    else
        echo -e "${RED}✗ 字段添加失败${NC}"
        exit 1
    fi
    echo ""

    # 4. 添加索引
    echo "4. 添加索引..."
    mysql -h$DB_HOST -P$DB_PORT -u$DB_USER -p$DB_PASS $DB_NAME << 'EOF' 2>/dev/null
ALTER TABLE battle_room
ADD INDEX idx_challenger_ready (challenger_ready);
EOF

    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ 索引添加成功${NC}"
    else
        echo -e "${YELLOW}⚠ 索引添加失败(可能已存在)${NC}"
    fi
    echo ""
fi

# 5. 验证字段
echo "5. 验证字段信息..."
mysql -h$DB_HOST -P$DB_PORT -u$DB_USER -p$DB_PASS $DB_NAME -t << 'EOF' 2>/dev/null
SELECT
    COLUMN_NAME AS '字段名',
    COLUMN_TYPE AS '类型',
    IS_NULLABLE AS '可空',
    COLUMN_DEFAULT AS '默认值',
    COLUMN_COMMENT AS '注释'
FROM INFORMATION_SCHEMA.COLUMNS
WHERE TABLE_SCHEMA = '$DB_NAME'
AND TABLE_NAME = 'battle_room'
AND COLUMN_NAME = 'challenger_ready';
EOF

echo ""

# 6. 查看表结构
echo "6. 查看 battle_room 表结构..."
mysql -h$DB_HOST -P$DB_PORT -u$DB_USER -p$DB_PASS $DB_NAME -e "DESCRIBE battle_room;" 2>/dev/null | grep challenger_ready

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ 字段验证成功${NC}"
else
    echo -e "${RED}✗ 字段验证失败${NC}"
fi
echo ""

# 7. 统计现有数据
echo "7. 统计现有数据..."
TOTAL_ROOMS=$(mysql -h$DB_HOST -P$DB_PORT -u$DB_USER -p$DB_PASS $DB_NAME -sN -e "SELECT COUNT(*) FROM battle_room;" 2>/dev/null)
echo "  总房间数: $TOTAL_ROOMS"

ROOMS_WITH_CHALLENGER=$(mysql -h$DB_HOST -P$DB_PORT -u$DB_USER -p$DB_PASS $DB_NAME -sN -e "SELECT COUNT(*) FROM battle_room WHERE challenger_id IS NOT NULL;" 2>/dev/null)
echo "  有挑战者的房间: $ROOMS_WITH_CHALLENGER"

ROOMS_READY=$(mysql -h$DB_HOST -P$DB_PORT -u$DB_USER -p$DB_PASS $DB_NAME -sN -e "SELECT COUNT(*) FROM battle_room WHERE challenger_ready = 1;" 2>/dev/null)
echo "  已准备的房间: $ROOMS_READY"
echo ""

echo "========================================="
echo -e "${GREEN}✓ 部署完成!${NC}"
echo "========================================="
echo ""
echo "下一步操作:"
echo "1. 重启后端服务: cd hist-oj && make run"
echo "2. 清除浏览器缓存 (Ctrl+Shift+R)"
echo "3. 测试准备功能"
echo ""
