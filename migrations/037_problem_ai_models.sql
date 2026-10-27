-- 保留原单模型字段；旧配置读取时自动映射，管理员保存时才写入模型列表。
-- 发布使用 models_json 的后端前执行；兼容 MySQL 5.7，重复执行安全。
DROP PROCEDURE IF EXISTS hist_add_problem_ai_models;
DELIMITER $$
CREATE PROCEDURE hist_add_problem_ai_models()
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'problem_ai_config' AND COLUMN_NAME = 'models_json'
  ) THEN
    ALTER TABLE problem_ai_config ADD COLUMN models_json mediumtext DEFAULT NULL COMMENT 'AI模型配置列表';
  END IF;
END$$
DELIMITER ;
CALL hist_add_problem_ai_models();
DROP PROCEDURE hist_add_problem_ai_models;
