-- 添加 skip_data_changed_at 字段到 contest_rating_status 表
-- 用于追踪 Skip 数据的最后修改时间，判断是否需要重算 Rating

ALTER TABLE `contest_rating_status`
ADD COLUMN `skip_data_changed_at` datetime DEFAULT NULL COMMENT 'Skip数据最后修改时间（用于判断是否需要重算）' AFTER `recalculate_lock`;

-- 添加 updated_at 字段到 contest_skip_users 表
-- 用于追踪 Skip 记录的修改时间

ALTER TABLE `contest_skip_users`
ADD COLUMN `updated_at` datetime DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间' AFTER `created_at`;

-- ========== 数据迁移 ==========

-- 1. 为已有 skip 用户的比赛设置 skipDataChangedAt
-- 使用 skip 用户的最后创建时间作为初始值
UPDATE `contest_rating_status` crs
SET `skip_data_changed_at` = (
    SELECT MAX(`created_at`)
    FROM `contest_skip_users` csu
    WHERE csu.`contest_id` = crs.`contest_id`
)
WHERE EXISTS (
    SELECT 1
    FROM `contest_skip_users` csu
    WHERE csu.`contest_id` = crs.`contest_id`
);

-- 2. 设置现有 skip 记录的 updated_at
-- 将 updated_at 设置为与 created_at 相同
UPDATE `contest_skip_users`
SET `updated_at` = `created_at`
WHERE `updated_at` IS NULL;

-- 3. 验证迁移结果
SELECT
    crs.`contest_id`,
    crs.`calculated_at` AS '上次计算时间',
    crs.`skip_data_changed_at` AS 'Skip最后修改时间',
    COUNT(csu.`uid`) AS 'Skip用户数'
FROM `contest_rating_status` crs
LEFT JOIN `contest_skip_users` csu ON csu.`contest_id` = crs.`contest_id`
GROUP BY crs.`contest_id`
HAVING COUNT(csu.`uid`) > 0
ORDER BY crs.`contest_id`;
