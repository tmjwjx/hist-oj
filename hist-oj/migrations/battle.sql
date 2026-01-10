-- 代码对战功能数据库表创建脚本
-- 创建时间: 2025-01-09

-- 对战房间表
CREATE TABLE IF NOT EXISTS `battle_room` (
  `id` BIGINT(20) NOT NULL AUTO_INCREMENT COMMENT '房间ID',
  `room_id` VARCHAR(10) NOT NULL COMMENT '房间号（6位）',
  `host_id` VARCHAR(20) NOT NULL COMMENT '房主用户ID',
  `host_username` VARCHAR(50) NOT NULL COMMENT '房主用户名',
  `challenger_id` VARCHAR(20) DEFAULT NULL COMMENT '挑战者用户ID',
  `challenger_username` VARCHAR(50) DEFAULT NULL COMMENT '挑战者用户名',
  `problem_id` BIGINT(20) DEFAULT NULL COMMENT '对战题目ID',
  `status` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '房间状态：0-等待中，1-对战中，2-已结束',
  `winner_id` VARCHAR(20) DEFAULT NULL COMMENT '获胜者用户ID',
  `end_reason` VARCHAR(20) DEFAULT NULL COMMENT '结束原因：ac-一方AC，giveup-一方放弃，timeout-超时',
  `start_time` DATETIME DEFAULT NULL COMMENT '开始时间',
  `end_time` DATETIME DEFAULT NULL COMMENT '结束时间',
  `gmt_create` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `gmt_modified` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_room_id` (`room_id`),
  KEY `idx_host_id` (`host_id`),
  KEY `idx_status` (`status`),
  KEY `idx_create_time` (`gmt_create`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='对战房间表';

-- 对战记录表
CREATE TABLE IF NOT EXISTS `battle_record` (
  `id` BIGINT(20) NOT NULL AUTO_INCREMENT COMMENT '记录ID',
  `room_id` VARCHAR(10) NOT NULL COMMENT '房间号',
  `user_id` VARCHAR(20) NOT NULL COMMENT '用户ID',
  `username` VARCHAR(50) NOT NULL COMMENT '用户名',
  `opponent_id` VARCHAR(20) NOT NULL COMMENT '对手用户ID',
  `opponent_username` VARCHAR(50) NOT NULL COMMENT '对手用户名',
  `problem_id` BIGINT(20) NOT NULL COMMENT '对战题目ID',
  `problem_title` VARCHAR(200) NOT NULL COMMENT '题目标题',
  `is_winner` TINYINT(1) NOT NULL COMMENT '是否获胜：0-失败，1-获胜',
  `end_reason` VARCHAR(20) NOT NULL COMMENT '结束原因',
  `submit_count` INT(11) DEFAULT 0 COMMENT '提交次数',
  `battle_time` INT(11) DEFAULT NULL COMMENT '对战时长（秒）',
  `gmt_create` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_is_winner` (`is_winner`),
  KEY `idx_battle_time` (`battle_time`),
  KEY `idx_create_time` (`gmt_create`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='对战记录表';

-- 用户对战统计表
CREATE TABLE IF NOT EXISTS `user_battle_stats` (
  `id` BIGINT(20) NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `user_id` VARCHAR(20) NOT NULL COMMENT '用户ID',
  `username` VARCHAR(50) NOT NULL COMMENT '用户名',
  `total_battles` INT(11) NOT NULL DEFAULT 0 COMMENT '总对战场次',
  `win_count` INT(11) NOT NULL DEFAULT 0 COMMENT '获胜场次',
  `lose_count` INT(11) NOT NULL DEFAULT 0 COMMENT '失败场次',
  `win_rate` DECIMAL(5,2) DEFAULT 0.00 COMMENT '胜率（%）',
  `total_submit_count` INT(11) DEFAULT 0 COMMENT '总提交次数',
  `avg_battle_time` INT(11) DEFAULT NULL COMMENT '平均对战时长（秒）',
  `gmt_create` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `gmt_modified` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_id` (`user_id`),
  KEY `idx_win_count` (`win_count`),
  KEY `idx_win_rate` (`win_rate`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户对战统计表';
