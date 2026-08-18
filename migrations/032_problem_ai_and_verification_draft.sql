CREATE TABLE IF NOT EXISTS `problem_verification_draft` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `pid` bigint unsigned NOT NULL,
  `uid` varchar(32) NOT NULL,
  `language` varchar(40) NOT NULL,
  `code` mediumtext NOT NULL,
  `case_version` varchar(40) DEFAULT NULL,
  `gmt_create` datetime DEFAULT CURRENT_TIMESTAMP,
  `gmt_modified` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`), UNIQUE KEY `uk_verification_draft_user_problem` (`pid`,`uid`),
  KEY `idx_verification_draft_pid` (`pid`),
  CONSTRAINT `fk_verification_draft_problem` FOREIGN KEY (`pid`) REFERENCES `problem` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

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
  'problem_verification', 'sample_verified',
  'tinyint(1) NOT NULL DEFAULT 0'
);
CALL hist_add_column_if_missing(
  'judge', 'is_problem_verification',
  'tinyint(1) NOT NULL DEFAULT 0 COMMENT ''是否为标准程序验题提交'''
);
DROP PROCEDURE hist_add_column_if_missing;

CREATE TABLE IF NOT EXISTS `problem_ai_config` (
  `id` bigint unsigned NOT NULL,
  `enabled` tinyint(1) NOT NULL DEFAULT 0,
  `api_url` varchar(500) NOT NULL,
  `api_key` varchar(500) DEFAULT NULL,
  `model` varchar(100) DEFAULT NULL,
  `timeout_seconds` int NOT NULL DEFAULT 120,
  `system_prompt` text,
  `validation_prompt` text,
  `gmt_create` datetime DEFAULT CURRENT_TIMESTAMP,
  `gmt_modified` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `problem_ai_record` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `pid` bigint unsigned NOT NULL,
  `session_id` varchar(64) NOT NULL,
  `uid` varchar(32) NOT NULL,
  `question` text NOT NULL,
  `response` mediumtext,
  `status` varchar(20) NOT NULL DEFAULT 'running',
  `duration_ms` int DEFAULT NULL,
  `error_message` varchar(1000) DEFAULT NULL,
  `gmt_create` datetime DEFAULT CURRENT_TIMESTAMP,
  `gmt_modified` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`), KEY `idx_problem_ai_record_pid` (`pid`,`gmt_create`),
  CONSTRAINT `fk_problem_ai_record_problem` FOREIGN KEY (`pid`) REFERENCES `problem` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
