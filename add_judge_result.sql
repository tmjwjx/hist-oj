-- 添加 judge_result 字段到 homework_submit 表
-- 执行方式：mysql -h43.143.133.62 -uroot -phist2025 hoj < add_judge_result.sql

USE hoj;

-- 检查并添加字段（如果不存在）
SET @dbname = DATABASE();
SET @tablename = 'homework_submit';
SET @columnname = 'judge_result';

SET @preparedStatement = (SELECT IF(
  (
    SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
    WHERE TABLE_SCHEMA = @dbname
    AND TABLE_NAME = @tablename
    AND COLUMN_NAME = @columnname
  ) > 0,
  'SELECT 1',
  CONCAT('ALTER TABLE ', @tablename, ' ADD COLUMN ', @columnname, ' VARCHAR(50) DEFAULT NULL COMMENT ''评测结果（编程题）: AC/WA/CE/TLE/MLE/RE等'' AFTER `is_officially_submitted`')
));

PREPARE alterIfNotExists FROM @preparedStatement;
EXECUTE alterIfNotExists;
DEALLOCATE PREPARE alterIfNotExists;

-- 检查并添加索引（如果不存在）
SET @indexname = 'idx_judge_result';
SET @preparedStatement = (SELECT IF(
  (
    SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
    WHERE TABLE_SCHEMA = @dbname
    AND TABLE_NAME = @tablename
    AND INDEX_NAME = @indexname
  ) > 0,
  'SELECT 1',
  CONCAT('CREATE INDEX ', @indexname, ' ON ', @tablename, ' (`judge_result`)')
));

PREPARE createIndexIfNotExists FROM @preparedStatement;
EXECUTE createIndexIfNotExists;
DEALLOCATE PREPARE createIndexIfNotExists;

-- 验证字段是否添加成功
SELECT
    COLUMN_NAME,
    COLUMN_TYPE,
    IS_NULLABLE,
    COLUMN_DEFAULT,
    COLUMN_COMMENT
FROM INFORMATION_SCHEMA.COLUMNS
WHERE TABLE_SCHEMA = 'hoj'
  AND TABLE_NAME = 'homework_submit'
  AND COLUMN_NAME = 'judge_result';

-- 验证索引是否添加成功
SELECT
    INDEX_NAME,
    COLUMN_NAME,
    NON_UNIQUE
FROM INFORMATION_SCHEMA.STATISTICS
WHERE TABLE_SCHEMA = 'hoj'
  AND TABLE_NAME = 'homework_submit'
  AND INDEX_NAME = 'idx_judge_result';
