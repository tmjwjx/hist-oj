-- 修复慢查询 - 添加索引
-- 执行方法: mysql -h127.0.0.1 -P3306 -uroot -pHoj123456 hoj < fix_slow_query.sql

USE hoj;

-- 1. classroom_homework 表添加联合索引
-- 慢查询: SELECT * FROM classroom_homework WHERE is_exam_mode = 1 AND end_time > '...'
-- 需要: (is_exam_mode, end_time) 联合索引
ALTER TABLE `classroom_homework`
ADD INDEX `idx_exam_end_time` (`is_exam_mode`, `end_time`);

-- 2. contest_rating_status 表添加索引（如果还没有）
-- 慢查询: SELECT * FROM contest_rating_status WHERE contest_id = 1007
-- 需要: contest_id 索引
ALTER TABLE `contest_rating_status`
ADD UNIQUE INDEX `idx_contest_id` (`contest_id`);

-- 3. plagiarism_check 表添加索引（优化查重功能）
ALTER TABLE `plagiarism_check`
ADD INDEX `idx_cid_status` (`cid`, `status`),
ADD INDEX `idx_status_started_at` (`status`, `started_at`);

-- 4. plagiarism_result 表添加索引（优化结果查询）
ALTER TABLE `plagiarism_result`
ADD INDEX `idx_check_id` (`check_id`),
ADD INDEX `idx_cid_over_threshold` (`cid`, `is_over_threshold`);

-- 查看索引是否添加成功
SHOW INDEX FROM `classroom_homework`;
SHOW INDEX FROM `contest_rating_status`;
SHOW INDEX FROM `plagiarism_check`;
SHOW INDEX FROM `plagiarism_result`;
