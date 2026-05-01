-- 航海图权限控制升级（兼容 MySQL 5.7/8.0）
-- 执行前请先: USE hoj;

SET @db_name := DATABASE();

SET @col_exists := (
  SELECT COUNT(1)
  FROM information_schema.columns
  WHERE table_schema = @db_name
    AND table_name = 'learning_map'
    AND column_name = 'access_mode'
);
SET @sql_col := IF(
  @col_exists = 0,
  'ALTER TABLE `learning_map` ADD COLUMN `access_mode` varchar(20) NOT NULL DEFAULT ''all_open'' AFTER `status`',
  'SELECT 1'
);
PREPARE stmt_col FROM @sql_col;
EXECUTE stmt_col;
DEALLOCATE PREPARE stmt_col;

SET @idx_exists := (
  SELECT COUNT(1)
  FROM information_schema.statistics
  WHERE table_schema = @db_name
    AND table_name = 'learning_map'
    AND index_name = 'idx_access_mode'
);
SET @sql_idx := IF(
  @idx_exists = 0,
  'CREATE INDEX `idx_access_mode` ON `learning_map` (`access_mode`)',
  'SELECT 1'
);
PREPARE stmt_idx FROM @sql_idx;
EXECUTE stmt_idx;
DEALLOCATE PREPARE stmt_idx;

CREATE TABLE IF NOT EXISTS `learning_map_permission` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `map_id` bigint unsigned NOT NULL,
  `user_id` varchar(32) NOT NULL,
  `enabled` tinyint(1) NOT NULL DEFAULT 1,
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_map_user` (`map_id`, `user_id`),
  KEY `idx_map_id` (`map_id`),
  KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='航海图用户权限覆盖';
