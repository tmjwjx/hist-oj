-- 比赛违规查询：提交设备ID与User-Agent采集。
-- device_id 为前端生成并持久化在 localStorage 的随机UUID（不含任何浏览器指纹成分），
-- 同一台电脑/浏览器恒定，换浏览器或清空站点数据会生成新ID。
-- 历史提交两列均为 NULL，违规查询侧按"未采集设备"单独计数，不参与聚合。
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
  'judge', 'device_id',
  'varchar(64) DEFAULT NULL AFTER `ip`'
);
CALL hist_add_column_if_missing(
  'judge', 'user_agent',
  'varchar(255) DEFAULT NULL AFTER `device_id`'
);
DROP PROCEDURE hist_add_column_if_missing;
