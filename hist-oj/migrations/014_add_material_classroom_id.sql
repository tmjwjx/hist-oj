-- 添加classroom_id字段到classroom_material表
-- 解决跨班级数据泄露问题

-- 1. 添加classroom_id字段
ALTER TABLE classroom_material ADD COLUMN classroom_id BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER id;

-- 2. 添加索引
ALTER TABLE classroom_material ADD INDEX idx_classroom_id (classroom_id);

-- 3. 更新现有数据：通过folder_id反查classroom_id
-- 如果folder_id=0，暂时保持classroom_id=0（需要手动处理）
UPDATE classroom_material cm
SET classroom_id = (
    SELECT cf.classroom_id
    FROM classroom_folder cf
    WHERE cf.id = cm.folder_id
    LIMIT 1
)
WHERE cm.folder_id > 0;

-- 4. 对于folder_id=0的记录，需要根据创建者查找班级
-- 这部分需要根据实际情况手动更新
