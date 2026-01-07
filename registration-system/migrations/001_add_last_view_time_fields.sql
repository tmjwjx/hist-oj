-- 添加 last_view_time 和 admin_last_view_time 字段到报名表
-- 用于持久化存储用户和管理员的最后查看时间，解决浏览器缓存清除导致的消息未读问题

-- 为 histcontest_register_registrations 表添加字段
-- 用于 MySQL 生产环境

-- 添加 last_view_time（用户最后查看通信记录的时间）
ALTER TABLE histcontest_register_registrations
ADD COLUMN IF NOT EXISTS last_view_time DATETIME DEFAULT NULL COMMENT '用户最后查看通信记录的时间';

-- 添加 admin_last_view_time（管理员最后查看通信记录的时间）
ALTER TABLE histcontest_register_registrations
ADD COLUMN IF NOT EXISTS admin_last_view_time DATETIME DEFAULT NULL COMMENT '管理员最后查看通信记录的时间';

-- 创建索引以提高查询性能
CREATE INDEX IF NOT EXISTS idx_last_view_time ON histcontest_register_registrations(last_view_time);
CREATE INDEX IF NOT EXISTS idx_admin_last_view_time ON histcontest_register_registrations(admin_last_view_time);

-- 验证字段是否添加成功
SELECT
    COLUMN_NAME,
    DATA_TYPE,
    IS_NULLABLE,
    COLUMN_DEFAULT,
    COLUMN_COMMENT
FROM
    INFORMATION_SCHEMA.COLUMNS
WHERE
    TABLE_SCHEMA = 'hoj'
    AND TABLE_NAME = 'histcontest_register_registrations'
    AND COLUMN_NAME IN ('last_view_time', 'admin_last_view_time');
