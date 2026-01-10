-- 修复 battle_room 表的 problem_id 字段类型
-- 从 BIGINT 改为 VARCHAR 以支持各种格式的显示ID（如 0051, Z001 等）

USE hoj;

-- 修改 battle_room 表的 problem_id 字段
ALTER TABLE battle_room
  MODIFY COLUMN problem_id VARCHAR(255) DEFAULT NULL COMMENT '对战题目ID（显示ID，如 0051, Z001）';

-- 修改 battle_record 表的 problem_id 字段
ALTER TABLE battle_record
  MODIFY COLUMN problem_id VARCHAR(255) NOT NULL COMMENT '对战题目ID（显示ID）';
