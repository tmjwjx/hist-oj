-- 比赛答疑迁入 HOJ Java 后端。
CREATE TABLE IF NOT EXISTS `contest_question` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `contest_id` bigint unsigned NOT NULL,
  `questioner_id` varchar(32) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,
  `title` varchar(200) NOT NULL,
  `content` text NOT NULL,
  `status` varchar(20) NOT NULL DEFAULT 'pending',
  `priority` int NOT NULL DEFAULT 0,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_contest_question_contest` (`contest_id`),
  KEY `idx_contest_question_user` (`questioner_id`),
  KEY `idx_contest_question_status` (`status`),
  CONSTRAINT `fk_contest_question_contest` FOREIGN KEY (`contest_id`) REFERENCES `contest` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_contest_question_user` FOREIGN KEY (`questioner_id`) REFERENCES `user_info` (`uuid`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `contest_question_reply` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `question_id` bigint unsigned NOT NULL,
  `sender_id` varchar(32) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,
  `content` text NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_contest_question_reply_question` (`question_id`),
  KEY `idx_contest_question_reply_sender` (`sender_id`),
  KEY `idx_contest_question_reply_time` (`created_at`),
  CONSTRAINT `fk_contest_question_reply_question` FOREIGN KEY (`question_id`) REFERENCES `contest_question` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_contest_question_reply_sender` FOREIGN KEY (`sender_id`) REFERENCES `user_info` (`uuid`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

SET @has_hist_rating = (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'user_record' AND COLUMN_NAME = 'hist_rating'
);
SET @hist_rating_sql = IF(
  @has_hist_rating = 0,
  'ALTER TABLE `user_record` ADD COLUMN `hist_rating` int DEFAULT NULL COMMENT ''站内比赛 Rating'' AFTER `rating`',
  'SELECT 1'
);
PREPARE hist_rating_stmt FROM @hist_rating_sql;
EXECUTE hist_rating_stmt;
DEALLOCATE PREPARE hist_rating_stmt;
