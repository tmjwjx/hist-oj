-- Java 后端代码查重表。保留旧数据，避免 Go/Java 切换时重复建表。
CREATE TABLE IF NOT EXISTS plagiarism_check_config (
  id bigint unsigned NOT NULL AUTO_INCREMENT,
  cid bigint unsigned NOT NULL,
  pid bigint unsigned NOT NULL,
  cpid bigint unsigned NOT NULL,
  threshold int NOT NULL DEFAULT 50,
  created_by varchar(32) DEFAULT NULL,
  gmt_create datetime DEFAULT CURRENT_TIMESTAMP,
  gmt_modified datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id), UNIQUE KEY uk_plagiarism_cid_pid (cid, pid),
  KEY idx_plagiarism_config_cid (cid), KEY idx_plagiarism_config_cpid (cpid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS plagiarism_check (
  id bigint unsigned NOT NULL AUTO_INCREMENT,
  cid bigint unsigned NOT NULL,
  status varchar(20) NOT NULL DEFAULT 'pending',
  total_pairs int NOT NULL DEFAULT 0,
  checked_pairs int NOT NULL DEFAULT 0,
  total_submissions int NOT NULL DEFAULT 0,
  progress decimal(5,2) NOT NULL DEFAULT 0.00,
  error_message varchar(500) DEFAULT NULL,
  started_by varchar(32) DEFAULT NULL,
  gmt_create datetime DEFAULT CURRENT_TIMESTAMP,
  gmt_modified datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  started_at datetime DEFAULT NULL,
  completed_at datetime DEFAULT NULL,
  PRIMARY KEY (id), KEY idx_plagiarism_check_cid (cid), KEY idx_plagiarism_check_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS plagiarism_result (
  id bigint unsigned NOT NULL AUTO_INCREMENT,
  check_id bigint unsigned NOT NULL,
  cid bigint unsigned NOT NULL,
  cpid bigint unsigned NOT NULL,
  pid bigint unsigned NOT NULL,
  display_id varchar(255) DEFAULT NULL,
  problem_title varchar(255) DEFAULT NULL,
  submit_id_1 bigint unsigned NOT NULL,
  submit_id_2 bigint unsigned NOT NULL,
  contest_record_id_1 bigint unsigned DEFAULT NULL,
  contest_record_id_2 bigint unsigned DEFAULT NULL,
  uid_1 varchar(32) NOT NULL,
  uid_2 varchar(32) NOT NULL,
  username_1 varchar(255) DEFAULT NULL,
  username_2 varchar(255) DEFAULT NULL,
  language varchar(50) DEFAULT NULL,
  similarity_1_to_2 int DEFAULT 0,
  similarity_2_to_1 int DEFAULT 0,
  max_similarity int DEFAULT 0,
  is_over_threshold tinyint(1) DEFAULT 0,
  gmt_create datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id), KEY idx_plagiarism_result_check (check_id),
  KEY idx_plagiarism_result_max (max_similarity), KEY idx_plagiarism_result_threshold (is_over_threshold)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- MySQL 5.7 不支持 ADD COLUMN IF NOT EXISTS，使用元数据检查保证重复执行安全。
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
  'plagiarism_result', 'contest_record_id_1',
  'bigint unsigned DEFAULT NULL AFTER `submit_id_2`'
);
CALL hist_add_column_if_missing(
  'plagiarism_result', 'contest_record_id_2',
  'bigint unsigned DEFAULT NULL AFTER `contest_record_id_1`'
);
DROP PROCEDURE hist_add_column_if_missing;
