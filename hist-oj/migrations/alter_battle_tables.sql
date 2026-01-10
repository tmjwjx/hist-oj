-- 修复对战表用户ID字段长度问题
-- 将 VARCHAR(20) 扩展到 VARCHAR(64) 以支持 32 字符的 UID

USE hoj;

-- 修改 battle_room 表
ALTER TABLE battle_room
  MODIFY COLUMN host_id VARCHAR(64) NOT NULL COMMENT '房主用户ID',
  MODIFY COLUMN challenger_id VARCHAR(64) DEFAULT NULL COMMENT '挑战者用户ID',
  MODIFY COLUMN winner_id VARCHAR(64) DEFAULT NULL COMMENT '获胜者用户ID';

-- 修改 battle_record 表
ALTER TABLE battle_record
  MODIFY COLUMN user_id VARCHAR(64) NOT NULL COMMENT '用户ID',
  MODIFY COLUMN opponent_id VARCHAR(64) NOT NULL COMMENT '对手用户ID';

-- 修改 user_battle_stats 表
ALTER TABLE user_battle_stats
  MODIFY COLUMN user_id VARCHAR(64) NOT NULL COMMENT '用户ID';
