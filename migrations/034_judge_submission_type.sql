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
  'judge', 'submission_type',
  "varchar(32) NOT NULL DEFAULT 'user_submission' COMMENT '评测来源'"
);
UPDATE judge SET submission_type = 'creator_validation'
 WHERE COALESCE(is_problem_verification, 0) = 1
   AND (submission_type IS NULL OR submission_type = 'user_submission');
DROP PROCEDURE hist_add_column_if_missing;
