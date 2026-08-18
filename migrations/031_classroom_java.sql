-- 课堂扩展层迁移到 Java 主后端。所有表保留原 Go 服务使用的字段与数据。
CREATE TABLE IF NOT EXISTS `classroom_user_role` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT, `uid` varchar(32) NOT NULL, `role` varchar(20) NOT NULL,
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP, `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`), UNIQUE KEY `uk_classroom_user_role` (`uid`,`role`), KEY `idx_classroom_role_uid` (`uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `classroom` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT, `class_name` varchar(100) NOT NULL, `class_belong` varchar(100) NOT NULL,
  `class_code` varchar(8) NOT NULL, `teacher_id` varchar(32) NOT NULL, `status` tinyint NOT NULL DEFAULT 1,
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP, `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`), UNIQUE KEY `uk_class_code` (`class_code`), KEY `idx_classroom_teacher` (`teacher_id`), KEY `idx_classroom_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `classroom_teacher` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT, `classroom_id` bigint unsigned NOT NULL, `teacher_id` varchar(32) NOT NULL,
  `status` tinyint NOT NULL DEFAULT 1, `create_time` datetime DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`), UNIQUE KEY `uk_classroom_teacher` (`classroom_id`,`teacher_id`), KEY `idx_teacher_classroom` (`teacher_id`,`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `classroom_student` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT, `classroom_id` bigint unsigned NOT NULL, `uid` varchar(32) NOT NULL,
  `real_name` varchar(50) NOT NULL, `gender` varchar(10) DEFAULT NULL, `student_class` varchar(100) DEFAULT NULL,
  `student_no` varchar(50) DEFAULT NULL, `status` tinyint NOT NULL DEFAULT 1, `create_time` datetime DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`), UNIQUE KEY `uk_classroom_user` (`classroom_id`,`uid`), KEY `idx_student_uid` (`uid`), KEY `idx_student_classroom` (`classroom_id`,`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `classroom_role_request` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT, `uid` varchar(32) NOT NULL, `role` varchar(20) NOT NULL,
  `reason` text, `status` tinyint NOT NULL DEFAULT 0, `reviewer_uid` varchar(32) DEFAULT NULL,
  `review_time` datetime DEFAULT NULL, `review_note` text, `create_time` datetime DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`), KEY `idx_role_request_uid` (`uid`), KEY `idx_role_request_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `classroom_checkin` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT, `classroom_id` bigint unsigned NOT NULL, `checkin_code` varchar(20) NOT NULL,
  `checkin_name` varchar(100) DEFAULT NULL, `checkin_type` varchar(20) NOT NULL DEFAULT 'code', `qrcode_token` varchar(255) DEFAULT NULL,
  `qrcode_expires_at` datetime DEFAULT NULL, `qrcode_refresh_interval` int NOT NULL DEFAULT 15, `start_time` datetime NOT NULL,
  `end_time` datetime DEFAULT NULL, `status` tinyint NOT NULL DEFAULT 1, `create_time` datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`), KEY `idx_checkin_classroom` (`classroom_id`), KEY `idx_checkin_code` (`checkin_code`), KEY `idx_checkin_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `classroom_checkin_record` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT, `checkin_id` bigint unsigned NOT NULL, `uid` varchar(32) NOT NULL,
  `status` varchar(20) NOT NULL DEFAULT 'present', `checkin_time` datetime DEFAULT NULL, `remark` varchar(200) DEFAULT NULL,
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP, `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`), UNIQUE KEY `uk_checkin_student` (`checkin_id`,`uid`), KEY `idx_checkin_record_uid` (`uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `question_bank` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT, `title` varchar(200) NOT NULL, `type` varchar(20) NOT NULL, `content` text NOT NULL,
  `options` json DEFAULT NULL, `answer` text, `analysis` text, `tags` json DEFAULT NULL, `course` varchar(100) DEFAULT NULL,
  `difficulty` int NOT NULL DEFAULT 1, `score` int NOT NULL DEFAULT 2, `creator_id` varchar(32) NOT NULL,
  `is_shared` tinyint NOT NULL DEFAULT 0, `problem_id` varchar(50) DEFAULT NULL, `status` tinyint NOT NULL DEFAULT 1,
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP, `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`), KEY `idx_question_type` (`type`), KEY `idx_question_creator` (`creator_id`), KEY `idx_question_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `classroom_homework` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT, `classroom_id` bigint unsigned NOT NULL, `title` varchar(200) NOT NULL,
  `description` text, `start_time` datetime NOT NULL, `end_time` datetime NOT NULL, `show_score` tinyint NOT NULL DEFAULT 1,
  `show_homework` tinyint NOT NULL DEFAULT 0, `show_answer` tinyint NOT NULL DEFAULT 0, `show_rank` tinyint NOT NULL DEFAULT 0,
  `status` tinyint NOT NULL DEFAULT 1, `is_exam_mode` tinyint NOT NULL DEFAULT 0, `exam_duration` int NOT NULL DEFAULT 60,
  `allow_submit_after_minutes` int NOT NULL DEFAULT 0, `disable_copy_paste` tinyint NOT NULL DEFAULT 1,
  `require_fullscreen` tinyint NOT NULL DEFAULT 1, `disallow_tab_switch` tinyint NOT NULL DEFAULT 1,
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP, `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`), KEY `idx_homework_classroom` (`classroom_id`), KEY `idx_homework_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `homework_question` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT, `homework_id` bigint unsigned NOT NULL, `question_id` bigint unsigned DEFAULT NULL,
  `problem_id` varchar(50) DEFAULT NULL, `question_order` int NOT NULL, `score` int NOT NULL DEFAULT 2,
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY (`id`), UNIQUE KEY `uk_homework_order` (`homework_id`,`question_order`),
  KEY `idx_homework_question` (`homework_id`), KEY `idx_homework_question_id` (`question_id`), KEY `idx_homework_problem` (`problem_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `homework_submit` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT, `homework_id` bigint unsigned NOT NULL, `question_id` bigint unsigned DEFAULT NULL,
  `problem_id` varchar(50) DEFAULT NULL, `uid` varchar(32) NOT NULL, `answer` text, `attachment` varchar(1000) DEFAULT NULL,
  `submit_id` bigint unsigned DEFAULT NULL, `score` decimal(5,2) NOT NULL DEFAULT 0, `is_scored` tinyint NOT NULL DEFAULT 0,
  `is_officially_submitted` tinyint NOT NULL DEFAULT 0, `judge_result` varchar(50) DEFAULT NULL,
  `exam_start_time` datetime DEFAULT NULL, `exam_end_time` datetime DEFAULT NULL, `is_forced_submit` tinyint NOT NULL DEFAULT 0,
  `tab_switch_count` int NOT NULL DEFAULT 0, `fullscreen_exit_count` int NOT NULL DEFAULT 0, `copy_paste_attempt_count` int NOT NULL DEFAULT 0,
  `device_info` varchar(500) DEFAULT NULL, `browser_info` varchar(500) DEFAULT NULL,
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP, `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`), UNIQUE KEY `uk_homework_submit_question` (`homework_id`,`question_id`,`uid`), KEY `idx_submit_homework` (`homework_id`), KEY `idx_submit_uid` (`uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `classroom_exam_paper` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT, `title` varchar(255) NOT NULL, `creator_id` varchar(32) NOT NULL,
  `is_shared` tinyint NOT NULL DEFAULT 0, `is_public` tinyint NOT NULL DEFAULT 0, `total_score` int NOT NULL DEFAULT 0,
  `question_count` int NOT NULL DEFAULT 0, `description` text, `status` tinyint NOT NULL DEFAULT 1,
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP, `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`), KEY `idx_paper_creator` (`creator_id`), KEY `idx_paper_public` (`is_public`), KEY `idx_paper_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `classroom_exam_paper_question` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT, `exam_paper_id` bigint unsigned NOT NULL, `question_id` bigint unsigned DEFAULT NULL,
  `problem_id` varchar(64) DEFAULT NULL, `question_order` int NOT NULL, `question_type` varchar(50) NOT NULL, `score` int NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`), UNIQUE KEY `uk_paper_question_order` (`exam_paper_id`,`question_order`), KEY `idx_paper_question` (`question_id`), KEY `idx_paper_problem` (`problem_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `classroom_folder` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT, `classroom_id` bigint unsigned NOT NULL, `folder_name` varchar(100) NOT NULL,
  `parent_id` bigint unsigned NOT NULL DEFAULT 0, `creator_id` varchar(32) NOT NULL, `sort_order` int NOT NULL DEFAULT 0,
  `status` tinyint NOT NULL DEFAULT 1, `create_time` datetime DEFAULT CURRENT_TIMESTAMP, `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`), KEY `idx_folder_classroom` (`classroom_id`), KEY `idx_folder_parent` (`parent_id`), KEY `idx_folder_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `classroom_material` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT, `classroom_id` bigint unsigned NOT NULL DEFAULT 0, `folder_id` bigint unsigned NOT NULL,
  `file_name` varchar(255) NOT NULL, `file_type` varchar(20) NOT NULL, `file_path` varchar(500) NOT NULL, `file_size` bigint unsigned DEFAULT NULL,
  `creator_id` varchar(32) NOT NULL, `is_shared` tinyint NOT NULL DEFAULT 0, `download_count` int NOT NULL DEFAULT 0, `status` tinyint NOT NULL DEFAULT 1,
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP, `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`), KEY `idx_material_classroom` (`classroom_id`), KEY `idx_material_folder` (`folder_id`), KEY `idx_material_creator` (`creator_id`), KEY `idx_material_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `classroom_material_permission` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT, `material_id` bigint unsigned NOT NULL, `student_uid` varchar(32) NOT NULL,
  `can_preview` tinyint NOT NULL DEFAULT 0, `can_download` tinyint NOT NULL DEFAULT 0, `create_time` datetime DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP, PRIMARY KEY (`id`), UNIQUE KEY `uk_material_student` (`material_id`,`student_uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `classroom_random_pick` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT, `classroom_id` bigint unsigned NOT NULL, `picked_uid` varchar(32) NOT NULL,
  `pick_time` datetime DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY (`id`), KEY `idx_pick_classroom` (`classroom_id`), KEY `idx_pick_time` (`pick_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `classroom_message` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT, `classroom_id` bigint unsigned NOT NULL, `sender_id` varchar(32) NOT NULL,
  `content` text, `image_url` varchar(500) DEFAULT NULL, `msg_type` varchar(20) NOT NULL DEFAULT 'text', `create_time` datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`), KEY `idx_message_classroom` (`classroom_id`), KEY `idx_message_time` (`create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `student_question_order` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT, `homework_id` bigint unsigned NOT NULL, `uid` varchar(32) NOT NULL,
  `order_mapping` text NOT NULL, `create_time` datetime DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY (`id`), UNIQUE KEY `uk_order_homework_uid` (`homework_id`,`uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `exam_violation_log` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT, `homework_id` bigint unsigned NOT NULL, `uid` varchar(32) NOT NULL,
  `violation_type` varchar(50) NOT NULL, `description` text, `ip` varchar(50) DEFAULT NULL, `create_time` datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`), KEY `idx_violation_homework` (`homework_id`), KEY `idx_violation_uid` (`uid`), KEY `idx_violation_time` (`create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 已安装过 Go 课堂版本的实例只会跳过 CREATE TABLE，因此补齐历史表缺失字段。
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

CALL hist_add_column_if_missing('classroom_checkin', 'checkin_type', 'varchar(20) NOT NULL DEFAULT ''code''');
CALL hist_add_column_if_missing('classroom_checkin', 'qrcode_token', 'varchar(255) DEFAULT NULL');
CALL hist_add_column_if_missing('classroom_checkin', 'qrcode_expires_at', 'datetime DEFAULT NULL');
CALL hist_add_column_if_missing('classroom_checkin', 'qrcode_refresh_interval', 'int NOT NULL DEFAULT 15');

CALL hist_add_column_if_missing('question_bank', 'analysis', 'text');
CALL hist_add_column_if_missing('question_bank', 'tags', 'json DEFAULT NULL');
CALL hist_add_column_if_missing('question_bank', 'course', 'varchar(100) DEFAULT NULL');
CALL hist_add_column_if_missing('question_bank', 'problem_id', 'varchar(50) DEFAULT NULL');

CALL hist_add_column_if_missing('classroom_homework', 'show_answer', 'tinyint NOT NULL DEFAULT 0');
CALL hist_add_column_if_missing('classroom_homework', 'show_rank', 'tinyint NOT NULL DEFAULT 0');
CALL hist_add_column_if_missing('classroom_homework', 'is_exam_mode', 'tinyint NOT NULL DEFAULT 0');
CALL hist_add_column_if_missing('classroom_homework', 'exam_duration', 'int NOT NULL DEFAULT 60');
CALL hist_add_column_if_missing('classroom_homework', 'allow_submit_after_minutes', 'int NOT NULL DEFAULT 0');
CALL hist_add_column_if_missing('classroom_homework', 'disable_copy_paste', 'tinyint NOT NULL DEFAULT 1');
CALL hist_add_column_if_missing('classroom_homework', 'require_fullscreen', 'tinyint NOT NULL DEFAULT 1');
CALL hist_add_column_if_missing('classroom_homework', 'disallow_tab_switch', 'tinyint NOT NULL DEFAULT 1');

CALL hist_add_column_if_missing('homework_question', 'problem_id', 'varchar(50) DEFAULT NULL');
ALTER TABLE `homework_question` MODIFY COLUMN `question_id` bigint unsigned DEFAULT NULL;

CALL hist_add_column_if_missing('homework_submit', 'problem_id', 'varchar(50) DEFAULT NULL');
CALL hist_add_column_if_missing('homework_submit', 'attachment', 'varchar(1000) DEFAULT NULL');
CALL hist_add_column_if_missing('homework_submit', 'is_officially_submitted', 'tinyint NOT NULL DEFAULT 0');
CALL hist_add_column_if_missing('homework_submit', 'judge_result', 'varchar(50) DEFAULT NULL');
CALL hist_add_column_if_missing('homework_submit', 'exam_start_time', 'datetime DEFAULT NULL');
CALL hist_add_column_if_missing('homework_submit', 'exam_end_time', 'datetime DEFAULT NULL');
CALL hist_add_column_if_missing('homework_submit', 'is_forced_submit', 'tinyint NOT NULL DEFAULT 0');
CALL hist_add_column_if_missing('homework_submit', 'tab_switch_count', 'int NOT NULL DEFAULT 0');
CALL hist_add_column_if_missing('homework_submit', 'fullscreen_exit_count', 'int NOT NULL DEFAULT 0');
CALL hist_add_column_if_missing('homework_submit', 'copy_paste_attempt_count', 'int NOT NULL DEFAULT 0');
CALL hist_add_column_if_missing('homework_submit', 'device_info', 'varchar(500) DEFAULT NULL');
CALL hist_add_column_if_missing('homework_submit', 'browser_info', 'varchar(500) DEFAULT NULL');
ALTER TABLE `homework_submit` MODIFY COLUMN `question_id` bigint unsigned DEFAULT NULL;

CALL hist_add_column_if_missing('classroom_material', 'classroom_id', 'bigint unsigned NOT NULL DEFAULT 0');
CALL hist_add_column_if_missing(
  'classroom_material', 'update_time',
  'datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP'
);
CALL hist_add_column_if_missing('classroom_exam_paper', 'is_public', 'tinyint NOT NULL DEFAULT 0');
DROP PROCEDURE hist_add_column_if_missing;
