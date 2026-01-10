-- ========================================
-- 对战系统 - 添加挑战者准备状态字段
-- 数据库: hoj
-- 表: battle_room
-- 日期: 2025-01-10
-- ========================================

USE hoj;

-- 1. 添加 challenger_ready 字段
ALTER TABLE `battle_room`
ADD COLUMN `challenger_ready` TINYINT(1) NOT NULL DEFAULT 0
COMMENT '挑战者是否已准备 (0-未准备, 1-已准备)'
AFTER `challenger_username`;

-- 2. 添加索引以提高查询性能
ALTER TABLE `battle_room`
ADD INDEX `idx_challenger_ready` (`challenger_ready`);

-- 3. 验证字段是否添加成功
SELECT '========================================' AS '';
SELECT '✓ 字段添加成功!' AS '状态';
SELECT '========================================' AS '';

-- 4. 查看新字段的详细信息
SELECT
    COLUMN_NAME AS '字段名',
    COLUMN_TYPE AS '类型',
    IS_NULLABLE AS '可空',
    COLUMN_DEFAULT AS '默认值',
    COLUMN_COMMENT AS '注释'
FROM INFORMATION_SCHEMA.COLUMNS
WHERE TABLE_SCHEMA = 'hoj'
AND TABLE_NAME = 'battle_room'
AND COLUMN_NAME = 'challenger_ready';

-- 5. 查看表结构 (包含所有字段)
DESCRIBE battle_room;

-- ========================================
-- 执行完成!
-- 请检查上方输出,确认字段已添加
-- ========================================
