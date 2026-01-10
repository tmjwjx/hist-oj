#!/bin/bash

# 远程服务器数据库迁移脚本
# 用于在服务器上执行数据库迁移

set -e

# 服务器配置
SERVER_HOST="43.143.133.62"
DB_USER="root"
DB_PASS="jia13579.."
DB_NAME="hoj"
DB_PORT="3306"

echo "========================================="
echo "对战系统 - 数据库迁移"
echo "========================================="
echo ""

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "准备在服务器 $SERVER_HOST 上执行数据库迁移..."
echo ""

# SQL 语句
SQL_QUERY="
USE $DB_NAME;

-- 添加 challenger_ready 字段
ALTER TABLE \`battle_room\`
ADD COLUMN \`challenger_ready\` TINYINT(1) NOT NULL DEFAULT 0
COMMENT '挑战者是否已准备 (0-未准备, 1-已准备)'
AFTER \`challenger_username\`;

-- 添加索引
ALTER TABLE \`battle_room\`
ADD INDEX \`idx_challenger_ready\` (\`challenger_ready\`);

-- 验证字段
SELECT '✓ 字段添加成功' AS status;

-- 查看字段信息
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
"

echo "正在尝试连接到远程服务器..."
echo ""

# 尝试使用 sshpass (如果有安装)
if command -v sshpass &> /dev/null; then
    echo -e "${YELLOW}使用 sshpass 连接...${NC}"
    sshpass -p "${DB_PASS}" ssh -o StrictHostKeyChecking=no root@${SERVER_HOST} "mysql -u${DB_USER} -p'${DB_PASS}' ${DB_NAME}" << EOF
${SQL_QUERY}
EOF
else
    # 尝试使用 expect (如果有安装)
    if command -v expect &> /dev/null; then
        echo -e "${YELLOW}使用 expect 连接...${NC}"
        expect << EOF
set timeout 30
spawn ssh root@${SERVER_HOST}
expect {
    "password:" {
        send "${DB_PASS}\r"
        expect "root@"
    }
    "yes/no" {
        send "yes\r"
        exp_continue
    }
}
expect "root@*"
send "mysql -u${DB_USER} -p'${DB_PASS}' ${DB_NAME}\r"
expect "mysql>"
send "ALTER TABLE battle_room ADD COLUMN challenger_ready TINYINT(1) NOT NULL DEFAULT 0 COMMENT '挑战者是否已准备' AFTER challenger_username;\r"
expect "mysql>"
send "ALTER TABLE battle_room ADD INDEX idx_challenger_ready (challenger_ready);\r"
expect "mysql>"
send "SELECT CONCAT('✓ 迁移完成') AS status;\r"
expect "mysql>"
send "exit\r"
expect "root@*"
send "exit\r"
expect eof
EOF
    else
        # 没有自动化工具,生成手动执行脚本
        echo -e "${RED}无法自动连接到服务器${NC}"
        echo ""
        echo "请手动执行以下步骤:"
        echo ""
        echo "1. SSH 连接到服务器:"
        echo "   ssh root@${SERVER_HOST}"
        echo ""
        echo "2. 连接到 MySQL:"
        echo "   mysql -u${DB_USER} -p'${DB_PASS}' ${DB_NAME}"
        echo ""
        echo "3. 执行以下 SQL:"
        echo ""
        echo -e "${GREEN}ALTER TABLE battle_room${NC}"
        echo "ADD COLUMN challenger_ready TINYINT(1) NOT NULL DEFAULT 0 "
        echo "COMMENT '挑战者是否已准备' AFTER challenger_username;"
        echo ""
        echo -e "${GREEN}ALTER TABLE battle_room${NC}"
        echo "ADD INDEX idx_challenger_ready (challenger_ready);"
        echo ""
        echo "4. 验证字段:"
        echo "   DESCRIBE battle_room;"
        echo ""
        exit 1
    fi
fi

echo ""
echo "========================================="
echo "迁移完成"
echo "========================================="
