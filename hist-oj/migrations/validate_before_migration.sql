-- 迁移前验证脚本
-- 在执行 009_add_show_answer.sql 之前运行此脚本进行验证

-- 1. 检查表是否存在
SELECT CONCAT('✓ 表 classroom_homework 存在') AS status
FROM information_schema.TABLES
WHERE TABLE_SCHEMA = 'hoj' AND TABLE_NAME = 'classroom_homework';

-- 2. 检查 show_homework 字段是否存在（新字段将添加到它后面）
SELECT CONCAT('✓ 字段 show_homework 存在') AS status
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = 'hoj'
  AND TABLE_NAME = 'classroom_homework'
  AND COLUMN_NAME = 'show_homework';

-- 3. 检查 show_score 字段是否存在
SELECT CONCAT('✓ 字段 show_score 存在') AS status
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = 'hoj'
  AND TABLE_NAME = 'classroom_homework'
  AND COLUMN_NAME = 'show_score';

-- 4. 检查 show_answer 字段是否已存在（如果存在则无需迁移）
SELECT
  CASE
    WHEN COUNT(*) > 0 THEN '⚠ 字段 show_answer 已存在，无需迁移'
    ELSE '✓ 字段 show_answer 不存在，可以添加'
  END AS status
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = 'hoj'
  AND TABLE_NAME = 'classroom_homework'
  AND COLUMN_NAME = 'show_answer';

-- 5. 查看当前字段顺序
SELECT
  COLUMN_NAME AS '字段名',
  COLUMN_TYPE AS '类型',
  COLUMN_DEFAULT AS '默认值',
  COLUMN_COMMENT AS '注释'
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = 'hoj'
  AND TABLE_NAME = 'classroom_homework'
  AND COLUMN_NAME IN ('id', 'title', 'show_homework', 'show_score', 'created_at', 'updated_at')
ORDER BY ORDINAL_POSITION;

-- 6. 统计现有作业数量
SELECT
  COUNT(*) AS '作业总数',
  SUM(CASE WHEN show_homework = 1 THEN 1 ELSE 0 END) AS '允许查看作业内容',
  SUM(CASE WHEN show_score = 1 THEN 1 ELSE 0 END) AS '允许查看成绩'
FROM classroom_homework;

-- 7. 检查是否有外键约束
SELECT
  CASE
    WHEN COUNT(*) = 0 THEN '✓ 无外键约束，可以安全添加字段'
    ELSE CONCAT('⚠ 存在 ', COUNT(*), ' 个外键约束')
  END AS status
FROM information_schema.KEY_COLUMN_USAGE
WHERE TABLE_SCHEMA = 'hoj'
  AND TABLE_NAME = 'classroom_homework'
  AND REFERENCED_TABLE_NAME IS NOT NULL;

-- 8. 检查表的存储引擎（InnoDB 支持在线 DDL）
SELECT
  CONCAT('✓ 存储引擎: ', ENGINE, CASE
    WHEN ENGINE = 'InnoDB' THEN ' (支持在线 DDL，添加字段不会锁表)'
    ELSE ' (添加字段可能锁表)'
  END) AS status
FROM information_schema.TABLES
WHERE TABLE_SCHEMA = 'hoj' AND TABLE_NAME = 'classroom_homework';

-- 9. 评估表大小（预估添加字段所需时间）
SELECT
  CONCAT('✓ 表大小: ',
    ROUND(ROUND((DATA_LENGTH + INDEX_LENGTH) / 1024 / 1024, 2)) AS 'MB'
  ) AS table_size
FROM information_schema.TABLES
WHERE TABLE_SCHEMA = 'hoj' AND TABLE_NAME = 'classroom_homework';

-- 10. 迁移建议
SELECT
  '=== 迁移建议 ===' AS '';

SELECT
  CASE
    WHEN (SELECT COUNT(*) FROM information_schema.COLUMNS
          WHERE TABLE_SCHEMA = 'hoj'
          AND TABLE_NAME = 'classroom_homework'
          AND COLUMN_NAME = 'show_answer') > 0
    THEN '⚠ 警告: show_answer 字段已存在，无需执行迁移'
    ELSE '✓ 可以安全执行迁移: 009_add_show_answer.sql'
  END AS migration_status;

SELECT
  '⚡ 建议: 在业务低峰期执行迁移（虽然添加字段很快，但谨慎为佳）' AS tip_1;

SELECT
  '💡 建议: 执行前备份数据库: mysqldump -u root -p hoj > backup_$(date +%Y%m%d).sql' AS tip_2;

SELECT
  '📝 建议: 执行后运行验证脚本: test_migration.sql' AS tip_3;
