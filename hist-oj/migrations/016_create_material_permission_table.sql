-- 创建资料库权限表
CREATE TABLE IF NOT EXISTS `classroom_material_permission` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `material_id` BIGINT UNSIGNED NOT NULL COMMENT '资料ID',
  `student_uid` VARCHAR(32) NOT NULL COMMENT '学生UID',
  `can_preview` INT NOT NULL DEFAULT 0 COMMENT '是否可预览(0否1是)',
  `can_download` INT NOT NULL DEFAULT 0 COMMENT '是否可下载(0否1是)',
  `create_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_material_id` (`material_id`),
  KEY `idx_student_uid` (`student_uid`),
  UNIQUE KEY `idx_material_student` (`material_id`, `student_uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='资料库文件权限表';
