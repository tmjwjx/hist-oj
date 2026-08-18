-- 将代码对战运行时迁入 Java HOJ。保留历史记录，补齐旧 Go 表结构。

CREATE TABLE IF NOT EXISTS `battle_room` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `room_id` varchar(10) NOT NULL,
  `host_id` varchar(64) NOT NULL,
  `host_username` varchar(255) NOT NULL,
  `challenger_id` varchar(64) DEFAULT NULL,
  `challenger_username` varchar(255) DEFAULT NULL,
  `challenger_ready` tinyint(1) NOT NULL DEFAULT 0,
  `problem_id` varchar(255) DEFAULT NULL,
  `status` tinyint NOT NULL DEFAULT 0,
  `winner_id` varchar(64) DEFAULT NULL,
  `end_reason` varchar(20) DEFAULT NULL,
  `start_time` datetime DEFAULT NULL,
  `end_time` datetime DEFAULT NULL,
  `gmt_create` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `gmt_modified` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_battle_room_id` (`room_id`),
  KEY `idx_battle_host` (`host_id`),
  KEY `idx_battle_challenger` (`challenger_id`),
  KEY `idx_battle_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `battle_record` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `battle_pair_id` varchar(64) NOT NULL,
  `room_id` varchar(10) NOT NULL,
  `user_id` varchar(64) NOT NULL,
  `username` varchar(255) NOT NULL,
  `opponent_id` varchar(64) NOT NULL,
  `opponent_username` varchar(255) NOT NULL,
  `opponent_rating` int DEFAULT NULL,
  `problem_id` varchar(255) NOT NULL,
  `problem_title` varchar(255) NOT NULL,
  `is_winner` tinyint(1) NOT NULL,
  `end_reason` varchar(20) NOT NULL,
  `submit_count` int NOT NULL DEFAULT 0,
  `battle_time` int DEFAULT NULL,
  `is_excluded` tinyint(1) NOT NULL DEFAULT 0,
  `gmt_create` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_battle_pair` (`battle_pair_id`),
  KEY `idx_battle_record_user` (`user_id`),
  KEY `idx_battle_record_excluded` (`is_excluded`),
  KEY `idx_battle_record_created` (`gmt_create`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

SET @has_ready := (SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'battle_room' AND COLUMN_NAME = 'challenger_ready');
SET @ddl := IF(@has_ready = 0,
  'ALTER TABLE `battle_room` ADD COLUMN `challenger_ready` tinyint(1) NOT NULL DEFAULT 0 AFTER `challenger_username`',
  'SELECT 1');
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @has_pair := (SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'battle_record' AND COLUMN_NAME = 'battle_pair_id');
SET @ddl := IF(@has_pair = 0,
  'ALTER TABLE `battle_record` ADD COLUMN `battle_pair_id` varchar(64) NULL AFTER `id`',
  'SELECT 1');
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @has_pair_index := (SELECT COUNT(*) FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'battle_record'
    AND INDEX_NAME IN ('idx_battle_pair', 'idx_battle_pair_id'));
SET @ddl := IF(@has_pair_index = 0,
  'ALTER TABLE `battle_record` ADD INDEX `idx_battle_pair` (`battle_pair_id`)',
  'SELECT 1');
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @has_opponent_rating := (SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'battle_record' AND COLUMN_NAME = 'opponent_rating');
SET @ddl := IF(@has_opponent_rating = 0,
  'ALTER TABLE `battle_record` ADD COLUMN `opponent_rating` int DEFAULT NULL AFTER `opponent_username`',
  'SELECT 1');
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @has_excluded := (SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'battle_record' AND COLUMN_NAME = 'is_excluded');
SET @ddl := IF(@has_excluded = 0,
  'ALTER TABLE `battle_record` ADD COLUMN `is_excluded` tinyint(1) NOT NULL DEFAULT 0 AFTER `battle_time`',
  'SELECT 1');
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @has_excluded_index := (SELECT COUNT(*) FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'battle_record'
    AND INDEX_NAME IN ('idx_battle_record_excluded', 'idx_is_excluded'));
SET @ddl := IF(@has_excluded_index = 0,
  'ALTER TABLE `battle_record` ADD INDEX `idx_battle_record_excluded` (`is_excluded`)',
  'SELECT 1');
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

ALTER TABLE `battle_room`
  MODIFY COLUMN `host_id` varchar(64) NOT NULL,
  MODIFY COLUMN `challenger_id` varchar(64) DEFAULT NULL,
  MODIFY COLUMN `winner_id` varchar(64) DEFAULT NULL,
  MODIFY COLUMN `problem_id` varchar(255) DEFAULT NULL;

ALTER TABLE `battle_record`
  MODIFY COLUMN `user_id` varchar(64) NOT NULL,
  MODIFY COLUMN `opponent_id` varchar(64) NOT NULL,
  MODIFY COLUMN `problem_id` varchar(255) NOT NULL;

UPDATE `battle_record`
SET `battle_pair_id` = REPLACE(UUID(), '-', '')
WHERE `battle_pair_id` IS NULL OR `battle_pair_id` = '';

ALTER TABLE `battle_record`
  MODIFY COLUMN `battle_pair_id` varchar(64) NOT NULL;
