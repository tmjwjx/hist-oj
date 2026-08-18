-- Rating 由 Go 扩展层迁移到 HOJ Java 主后端。

DROP PROCEDURE IF EXISTS hist_add_column_if_missing;
DELIMITER $$
CREATE PROCEDURE hist_add_column_if_missing(
  IN table_name_value varchar(64),
  IN column_name_value varchar(64),
  IN column_definition_value text
)
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE()
      AND TABLE_NAME = table_name_value
      AND COLUMN_NAME = column_name_value
  ) THEN
    SET @hist_column_ddl = CONCAT(
      'ALTER TABLE `', table_name_value, '` ADD COLUMN `',
      column_name_value, '` ', column_definition_value
    );
    PREPARE hist_column_stmt FROM @hist_column_ddl;
    EXECUTE hist_column_stmt;
    DEALLOCATE PREPARE hist_column_stmt;
  END IF;
END$$
DELIMITER ;

CALL hist_add_column_if_missing(
  'contest', 'is_rating',
  'tinyint(1) NOT NULL DEFAULT 0 COMMENT ''是否计入站内 Rating'' AFTER `type`'
);
CALL hist_add_column_if_missing(
  'user_record', 'hist_rating',
  'int DEFAULT NULL COMMENT ''站内比赛 Rating'' AFTER `rating`'
);

CREATE TABLE IF NOT EXISTS `rating_history` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `uid` varchar(32) NOT NULL,
  `contest_id` bigint unsigned DEFAULT NULL,
  `related_contest_id` bigint unsigned DEFAULT NULL,
  `old_rating` int DEFAULT NULL,
  `new_rating` int NOT NULL,
  `rating_change` int NOT NULL,
  `rank` int NOT NULL DEFAULT 0,
  `participants` int NOT NULL DEFAULT 0,
  `reason` varchar(255) NOT NULL DEFAULT '',
  `is_manual` tinyint(1) NOT NULL DEFAULT 0,
  `is_skip` tinyint(1) NOT NULL DEFAULT 0,
  `skip_reason` varchar(500) NOT NULL DEFAULT '',
  `operator_uid` varchar(32) NOT NULL DEFAULT '',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`), UNIQUE KEY `uk_rating_uid_contest` (`uid`,`contest_id`),
  KEY `idx_rating_uid` (`uid`), KEY `idx_rating_contest` (`contest_id`),
  KEY `idx_rating_related_contest` (`related_contest_id`), KEY `idx_rating_manual` (`is_manual`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `contest_rating_status` (
  `contest_id` bigint unsigned NOT NULL,
  `is_rated` tinyint(1) NOT NULL DEFAULT 0,
  `rating_calculated` tinyint(1) NOT NULL DEFAULT 0,
  `calculated_at` datetime DEFAULT NULL,
  `skip_count` int NOT NULL DEFAULT 0,
  `has_pending_skip` tinyint(1) NOT NULL DEFAULT 0,
  `recalculate_status` varchar(20) NOT NULL DEFAULT 'none',
  `last_recalculate_at` datetime DEFAULT NULL,
  `recalculate_lock` tinyint(1) NOT NULL DEFAULT 0,
  `skip_data_changed_at` datetime DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`contest_id`), KEY `idx_rating_calculated` (`rating_calculated`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `contest_skip_users` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `contest_id` bigint unsigned NOT NULL,
  `uid` varchar(32) NOT NULL,
  `username` varchar(100) NOT NULL,
  `reason` varchar(500) NOT NULL,
  `operator_uid` varchar(32) NOT NULL,
  `operator_username` varchar(100) DEFAULT NULL,
  `is_applied` tinyint(1) NOT NULL DEFAULT 0,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`), UNIQUE KEY `uk_contest_skip_uid` (`contest_id`,`uid`),
  KEY `idx_contest_skip_uid` (`uid`), KEY `idx_contest_skip_applied` (`contest_id`,`is_applied`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `rating_recalculate_queue` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `contest_id` bigint unsigned NOT NULL,
  `status` varchar(20) NOT NULL DEFAULT 'pending',
  `total_contests` int NOT NULL DEFAULT 0,
  `processed_contests` int NOT NULL DEFAULT 0,
  `error_message` text,
  `created_by` varchar(32) DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `started_at` datetime DEFAULT NULL,
  `completed_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`), KEY `idx_rating_queue_status` (`status`), KEY `idx_rating_queue_contest` (`contest_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `rating_operation_logs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `operator_uid` varchar(32) NOT NULL,
  `operator_username` varchar(100) DEFAULT NULL,
  `operation_type` varchar(50) NOT NULL,
  `target_type` varchar(50) DEFAULT NULL,
  `target_id` varchar(100) DEFAULT NULL,
  `operation_detail` text,
  `ip` varchar(50) DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`), KEY `idx_rating_log_operator` (`operator_uid`),
  KEY `idx_rating_log_type` (`operation_type`), KEY `idx_rating_log_created` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 兼容已由旧 Go 服务创建的表。
CALL hist_add_column_if_missing('rating_history', 'related_contest_id', 'bigint unsigned DEFAULT NULL AFTER `contest_id`');
ALTER TABLE `rating_history` MODIFY COLUMN `contest_id` bigint unsigned DEFAULT NULL;
CALL hist_add_column_if_missing('rating_history', 'reason', 'varchar(255) NOT NULL DEFAULT '''' AFTER `participants`');
CALL hist_add_column_if_missing('rating_history', 'is_manual', 'tinyint(1) NOT NULL DEFAULT 0 AFTER `reason`');
CALL hist_add_column_if_missing('rating_history', 'is_skip', 'tinyint(1) NOT NULL DEFAULT 0 AFTER `is_manual`');
CALL hist_add_column_if_missing('rating_history', 'skip_reason', 'varchar(500) NOT NULL DEFAULT '''' AFTER `is_skip`');
CALL hist_add_column_if_missing('rating_history', 'operator_uid', 'varchar(32) NOT NULL DEFAULT '''' AFTER `skip_reason`');
CALL hist_add_column_if_missing('contest_rating_status', 'skip_count', 'int NOT NULL DEFAULT 0');
CALL hist_add_column_if_missing('contest_rating_status', 'has_pending_skip', 'tinyint(1) NOT NULL DEFAULT 0');
CALL hist_add_column_if_missing('contest_rating_status', 'recalculate_status', 'varchar(20) NOT NULL DEFAULT ''none''');
CALL hist_add_column_if_missing('contest_rating_status', 'last_recalculate_at', 'datetime DEFAULT NULL');
CALL hist_add_column_if_missing('contest_rating_status', 'recalculate_lock', 'tinyint(1) NOT NULL DEFAULT 0');
CALL hist_add_column_if_missing('contest_rating_status', 'skip_data_changed_at', 'datetime DEFAULT NULL');
DROP PROCEDURE hist_add_column_if_missing;
