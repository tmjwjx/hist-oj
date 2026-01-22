-- 添加 is_excluded 字段到 battle_record 表
-- 用于管理员标记不计本场对决
-- 迁移文件: 011_add_battle_record_excluded.sql
-- 创建时间: 2026-01-22
--
-- 安全说明:
-- 1. 使用 TINYINT(1) 而不是 BOOLEAN,兼容性更好
-- 2. 默认值 0 表示正常计入
-- 3. 添加在 battle_time 字段之后,保持逻辑顺序
-- 4. 添加索引提高查询性能

-- 添加字段
ALTER TABLE `battle_record`
ADD COLUMN `is_excluded` TINYINT(1) NOT NULL DEFAULT 0
COMMENT '是否不计本场对决: 0-正常计入, 1-不计入排名和历史记录'
AFTER `battle_time`;

-- 添加索引
ALTER TABLE `battle_record`
ADD INDEX `idx_is_excluded` (`is_excluded`);
