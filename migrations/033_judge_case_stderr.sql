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
  'judge_case', 'stderr',
  'longtext COMMENT ''该测试点用户程序与判题程序的标准错误流'' AFTER `user_output`'
);
DROP PROCEDURE hist_add_column_if_missing;
