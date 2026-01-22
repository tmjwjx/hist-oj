-- 添加 is_excluded 字段到 battle_record 表
-- 用于管理员标记不计本场对决
-- 迁移文件: 011_add_battle_record_excluded.sql
-- 创建时间: 2026-01-22

-- 检查字段是否已存在,避免重复执行报错
SET @exist := (SELECT COUNT(*) FROM information_schema.COLUMNS
               WHERE TABLE_SCHEMA = DATABASE()
               AND TABLE_NAME = 'battle_record'
               AND COLUMN_NAME = 'is_excluded');

-- 只有字段不存在时才添加
SET @sql := IF(@exist = 0,
    'ALTER TABLE `battle_record` ADD COLUMN `is_excluded` TINYINT(1) NOT NULL DEFAULT 0 COMMENT ''是否不计本场对决: 0-正常计入, 1-不计入'' AFTER `battle_time`',
    'SELECT ''字段 is_excluded 已存在,跳过添加'' AS info');

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 添加索引以提高查询性能(检查索引是否存在)
SET @index_exist := (SELECT COUNT(*) FROM information_schema.STATISTICS
                     WHERE TABLE_SCHEMA = DATABASE()
                     AND TABLE_NAME = 'battle_record'
                     AND INDEX_NAME = 'idx_is_excluded');

SET @sql := IF(@index_exist = 0,
    'ALTER TABLE `battle_record` ADD INDEX `idx_is_excluded` (`is_excluded`)',
    'SELECT ''索引 idx_is_excluded 已存在,跳过添加'' AS info');

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

