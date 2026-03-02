-- 修复班级教师数据一致性
-- 问题描述：某些班级在 classroom_teacher 表中缺少创建者的记录
-- 解决方案：为所有在 classroom 表中存在但在 classroom_teacher 表中不存在的创建者添加记录

-- 1. 查看当前数据不一致的情况
SELECT
    c.id AS classroom_id,
    c.class_name,
    c.teacher_id AS creator_id,
    u.username AS creator_username,
    COUNT(ct.id) AS teacher_record_count
FROM classroom c
LEFT JOIN user u ON c.teacher_id = u.uuid
LEFT JOIN classroom_teacher ct ON ct.classroom_id = c.id AND ct.teacher_id = c.teacher_id
WHERE c.status = 1
GROUP BY c.id, c.class_name, c.teacher_id, u.username
HAVING teacher_record_count = 0;

-- 2. 为缺失的班级创建者添加 classroom_teacher 记录
INSERT INTO classroom_teacher (classroom_id, teacher_id, status, create_time, update_time)
SELECT
    c.id AS classroom_id,
    c.teacher_id,
    1 AS status,
    c.create_time,
    NOW()
FROM classroom c
WHERE c.status = 1
  AND c.teacher_id IS NOT NULL
  AND NOT EXISTS (
      -- 检查是否已存在记录
      SELECT 1
      FROM classroom_teacher ct
      WHERE ct.classroom_id = c.id
        AND ct.teacher_id = c.teacher_id
  )
  AND c.teacher_id != ''; -- 排除空的teacher_id

-- 3. 验证修复结果
SELECT
    c.id AS classroom_id,
    c.class_name,
    c.teacher_id AS creator_id,
    u.username AS creator_username,
    COUNT(ct.id) AS teacher_record_count,
    GROUP_CONCAT(ct2.teacher_id) AS all_teacher_ids,
    GROUP_CONCAT(u2.username) AS all_teacher_usernames
FROM classroom c
LEFT JOIN user u ON c.teacher_id = u.uuid
LEFT JOIN classroom_teacher ct ON ct.classroom_id = c.id AND ct.teacher_id = c.teacher_id
LEFT JOIN classroom_teacher ct2 ON ct2.classroom_id = c.id
LEFT JOIN user u2 ON ct2.teacher_id = u2.uuid
WHERE c.status = 1
GROUP BY c.id, c.class_name, c.teacher_id, u.username
ORDER BY c.id;

-- 4. 显示受影响的行数
SELECT CONCAT('成功修复 ', ROW_COUNT(), ' 个班级的教师数据') AS message;
