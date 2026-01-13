-- 添加 judge_result 字段到 homework_submit 表
-- 用于存储编程题的评测结果（AC、WA、CE等）
-- 此迁移可安全重复执行

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
