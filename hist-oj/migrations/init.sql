-- Rating服务数据库初始化脚本
-- 在HOJ数据库中执行此脚本

-- 1. Rating历史记录表
CREATE TABLE IF NOT EXISTS `rating_history` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `uid` varchar(32) NOT NULL COMMENT '用户id',
  `contest_id` bigint unsigned NOT NULL COMMENT '比赛id',
  `old_rating` int DEFAULT NULL COMMENT '变化前rating',
  `new_rating` int NOT NULL COMMENT '变化后rating',
  `rating_change` int NOT NULL COMMENT 'rating变化值',
  `rank` int NOT NULL COMMENT '比赛排名',
  `participants` int NOT NULL COMMENT '参赛人数',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_uid` (`uid`),
  KEY `idx_contest_id` (`contest_id`),
  KEY `idx_created_at` (`created_at`),
  CONSTRAINT `rating_history_ibfk_1` FOREIGN KEY (`uid`) REFERENCES `user_info` (`uuid`) ON DELETE CASCADE,
  CONSTRAINT `rating_history_ibfk_2` FOREIGN KEY (`contest_id`) REFERENCES `contest` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Rating历史记录表';

-- 2. 比赛Rating状态表
CREATE TABLE IF NOT EXISTS `contest_rating_status` (
  `contest_id` bigint unsigned NOT NULL COMMENT '比赛id',
  `is_rated` tinyint(1) DEFAULT '0' COMMENT '是否计分比赛',
  `rating_calculated` tinyint(1) DEFAULT '0' COMMENT '是否已计算rating',
  `calculated_at` datetime DEFAULT NULL COMMENT '计算时间',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`contest_id`),
  KEY `idx_is_rated` (`is_rated`),
  KEY `idx_rating_calculated` (`rating_calculated`),
  CONSTRAINT `contest_rating_status_ibfk_1` FOREIGN KEY (`contest_id`) REFERENCES `contest` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='比赛Rating状态表';

-- 3. 新增hist_rating字段（HOJ内部比赛rating，与CF rating分离）
ALTER TABLE `user_record` 
ADD COLUMN IF NOT EXISTS `hist_rating` int(11) DEFAULT NULL COMMENT 'HOJ得分' AFTER `rating`;

-- 4. （可选）初始化hist_rating字段为初始值1500
-- UPDATE user_record SET hist_rating = 1500 WHERE hist_rating IS NULL;

