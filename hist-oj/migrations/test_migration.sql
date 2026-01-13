-- 测试脚本 - 验证 show_answer 字段添加
-- 使用方式：在 MySQL 中执行此脚本以测试迁移

-- 1. 首先检查表结构（执行迁移前）
SELECT CONCAT('当前表结构中的字段数量: ', COUNT(*)) AS info
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = 'hoj' AND TABLE_NAME = 'classroom_homework';

-- 2. 查看当前 show 相关字段
SELECT COLUMN_NAME, COLUMN_TYPE, COLUMN_DEFAULT, COLUMN_COMMENT
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = 'hoj'
  AND TABLE_NAME = 'classroom_homework'
  AND COLUMN_NAME LIKE 'show%'
ORDER BY ORDINAL_POSITION;

-- 3. 执行迁移（如果字段不存在）
-- 注意：如果字段已存在，会报错，可以忽略
SET @sql = (
    SELECT IF(
        COUNT(*) = 0,
        'ALTER TABLE `classroom_homework`
         ADD COLUMN `show_answer` INT DEFAULT 0 COMMENT ''学生提交后是否可以查看答案（0: 否, 1: 是）'' AFTER `show_homework`',
        'SELECT ''字段 show_answer 已存在，跳过添加'' AS message'
    )
    FROM information_schema.COLUMNS
    WHERE TABLE_SCHEMA = 'hoj'
      AND TABLE_NAME = 'classroom_homework'
      AND COLUMN_NAME = 'show_answer'
);

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 4. 验证字段是否添加成功
SELECT COLUMN_NAME, COLUMN_TYPE, COLUMN_DEFAULT, COLUMN_COMMENT
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = 'hoj'
  AND TABLE_NAME = 'classroom_homework'
  AND COLUMN_NAME = 'show_answer';

-- 5. 查看更新后的字段顺序
SELECT COLUMN_NAME, COLUMN_TYPE, COLUMN_DEFAULT
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = 'hoj'
  AND TABLE_NAME = 'classroom_homework'
  AND COLUMN_NAME IN ('show_homework', 'show_answer', 'show_score')
ORDER BY ORDINAL_POSITION;

-- 6. 测试默认值（插入一条测试记录，然后回滚）
START TRANSACTION;
INSERT INTO classroom_homework (classroom_id, title, start_time, end_time, status, show_homework, show_answer, show_score)
VALUES (99999, '测试作业', NOW(), DATE_ADD(NOW(), INTERVAL 7 DAY), 2, 1, DEFAULT, 1);
SELECT CONCAT('新插入记录的 show_answer 值: ', show_answer) AS test_result
FROM classroom_homework
WHERE title = '测试作业';
ROLLBACK; -- 回滚测试数据

-- 7. 完成提示
SELECT '迁移测试完成！请检查上述输出确认字段添加成功。' AS message;
