-- ===================================
-- 资料库权限表创建脚本 v2
-- ===================================

-- 第一步：检查 classroom_material 表的 id 列类型
-- 请先执行以下查询，确认 id 列的类型：
SELECT
    COLUMN_NAME,
    DATA_TYPE,
    COLUMN_TYPE,
    IS_NULLABLE,
    COLUMN_KEY
FROM INFORMATION_SCHEMA.COLUMNS
WHERE TABLE_SCHEMA = 'hoj'
  AND TABLE_NAME = 'classroom_material'
  AND COLUMN_NAME = 'id';

-- 第二步：根据查询结果决定是否需要修改 classroom_material.id 的类型
-- 如果查询结果显示 DATA_TYPE 不是 'bigint' 或者 COLUMN_TYPE 不包含 'unsigned'
-- 请执行以下 ALTER TABLE 语句修改类型（注意：这可能会锁表，请在低峰期执行）

-- 注意：如果你不想修改现有表，可以跳过这一步，直接使用不包含外键约束的表创建语句

-- 方案A：修改 classroom_material.id 为 bigint unsigned（推荐，但需要停机维护）
-- ALTER TABLE classroom_material MODIFY COLUMN id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT;

-- 方案B：先创建不带外键约束的权限表（快速方案，不影响现有服务）
-- 如果 classroom_material.id 已经是 int 或其他类型，使用这个方案

-- ===================================
-- 第三步：创建 classroom_material_permission 表
-- ===================================

-- 删除旧表（如果存在）
DROP TABLE IF EXISTS `classroom_material_permission`;

-- 创建权限表（方案B：先不添加外键约束，避免类型不匹配）
CREATE TABLE `classroom_material_permission` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `material_id` BIGINT UNSIGNED NOT NULL COMMENT '资料ID',
  `student_uid` VARCHAR(32) NOT NULL COMMENT '学生UID',
  `can_preview` INT DEFAULT 0 COMMENT '是否可预览(0否1是)',
  `can_download` INT DEFAULT 0 COMMENT '是否可下载(0否1是)',
  `create_time` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `update_time` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  INDEX `idx_material_id` (`material_id`),
  INDEX `idx_student_uid` (`student_uid`),
  UNIQUE KEY `uk_material_student` (`material_id`, `student_uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='资料库文件权限表';

-- ===================================
-- 第四步：（可选）在确认类型匹配后添加外键约束
-- ===================================

-- 如果上面的第一步查询显示 classroom_material.id 是 BIGINT UNSIGNED
-- 可以执行以下语句添加外键约束：

-- ALTER TABLE classroom_material_permission
-- ADD CONSTRAINT fk_material_permission_material
--   FOREIGN KEY (material_id)
--   REFERENCES classroom_material (id)
--   ON DELETE CASCADE
--   ON UPDATE CASCADE;

-- 如果 classroom_material.id 不是 BIGINT UNSIGNED，但你想添加外键
-- 需要先修改 classroom_material.id 的类型（参考第二步的方案A）

-- ===================================
-- 验证表创建成功
-- ===================================
SHOW CREATE TABLE classroom_material_permission;

-- 验证唯一约束
SHOW INDEX FROM classroom_material_permission WHERE Key_name = 'uk_material_student';
