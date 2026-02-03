-- ========================================
-- 代码查重功能数据库表
-- ========================================

-- 1. 查重配置表（存储每个比赛每个题目的查重阈值）
DROP TABLE IF EXISTS `plagiarism_check_config`;

CREATE TABLE `plagiarism_check_config` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `cid` bigint(20) unsigned NOT NULL COMMENT '比赛ID',
  `pid` bigint(20) unsigned NOT NULL COMMENT '题目ID',
  `cpid` bigint(20) unsigned NOT NULL COMMENT '比赛题目ID',
  `threshold` int(11) NOT NULL DEFAULT '50' COMMENT '查重率阈值(0-100)，超过此值将被记录',
  `created_by` varchar(32) DEFAULT NULL COMMENT '创建者用户ID',
  `gmt_create` datetime DEFAULT CURRENT_TIMESTAMP,
  `gmt_modified` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_cid_pid` (`cid`, `pid`),
  KEY `idx_cid` (`cid`),
  KEY `idx_pid` (`pid`),
  KEY `idx_cpid` (`cpid`),
  CONSTRAINT `plagiarism_config_ibfk_1` FOREIGN KEY (`cid`) REFERENCES `contest` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `plagiarism_config_ibfk_2` FOREIGN KEY (`cpid`) REFERENCES `contest_problem` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COMMENT='查重配置表';

-- 2. 查重任务表
DROP TABLE IF EXISTS `plagiarism_check`;

CREATE TABLE `plagiarism_check` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `cid` bigint(20) unsigned NOT NULL COMMENT '比赛ID',
  `status` varchar(20) NOT NULL DEFAULT 'pending' COMMENT '状态: pending-待开始, running-进行中, completed-已完成, failed-失败',
  `total_pairs` int(11) NOT NULL DEFAULT '0' COMMENT '总对比对数',
  `checked_pairs` int(11) NOT NULL DEFAULT '0' COMMENT '已检查对数',
  `total_submissions` int(11) NOT NULL DEFAULT '0' COMMENT '总提交数',
  `progress` decimal(5,2) DEFAULT '0.00' COMMENT '进度百分比(0-100)',
  `error_message` varchar(500) DEFAULT NULL COMMENT '错误信息',
  `started_by` varchar(32) DEFAULT NULL COMMENT '启动者用户ID',
  `gmt_create` datetime DEFAULT CURRENT_TIMESTAMP,
  `gmt_modified` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `started_at` datetime DEFAULT NULL COMMENT '开始时间',
  `completed_at` datetime DEFAULT NULL COMMENT '完成时间',
  PRIMARY KEY (`id`),
  KEY `idx_cid` (`cid`),
  KEY `idx_status` (`status`),
  KEY `idx_started_at` (`started_at`),
  CONSTRAINT `plagiarism_check_ibfk_1` FOREIGN KEY (`cid`) REFERENCES `contest` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COMMENT='查重任务表';

-- 3. 查重结果表
DROP TABLE IF EXISTS `plagiarism_result`;

CREATE TABLE `plagiarism_result` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `check_id` bigint(20) unsigned NOT NULL COMMENT '查重任务ID',
  `cid` bigint(20) unsigned NOT NULL COMMENT '比赛ID',
  `cpid` bigint(20) unsigned NOT NULL COMMENT '比赛题目ID',
  `pid` bigint(20) unsigned NOT NULL COMMENT '题目ID',
  `display_id` varchar(255) DEFAULT NULL COMMENT '题目显示ID(如A、B、C)',
  `problem_title` varchar(255) DEFAULT NULL COMMENT '题目标题',
  `submit_id_1` bigint(20) unsigned NOT NULL COMMENT '提交ID1',
  `submit_id_2` bigint(20) unsigned NOT NULL COMMENT '提交ID2',
  `uid_1` varchar(32) NOT NULL COMMENT '用户1 ID',
  `uid_2` varchar(32) NOT NULL COMMENT '用户2 ID',
  `username_1` varchar(255) DEFAULT NULL COMMENT '用户1 用户名',
  `username_2` varchar(255) DEFAULT NULL COMMENT '用户2 用户名',
  `language` varchar(50) DEFAULT NULL COMMENT '编程语言',
  `similarity_1_to_2` int(11) DEFAULT NULL COMMENT '1对2的相似度(0-100)',
  `similarity_2_to_1` int(11) DEFAULT NULL COMMENT '2对1的相似度(0-100)',
  `max_similarity` int(11) DEFAULT NULL COMMENT '最大相似度',
  `is_over_threshold` tinyint(1) DEFAULT '0' COMMENT '是否超过阈值(0-否,1-是)',
  `gmt_create` datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_check_id` (`check_id`),
  KEY `idx_cid` (`cid`),
  KEY `idx_cpid` (`cpid`),
  KEY `idx_max_similarity` (`max_similarity`),
  KEY `idx_is_over_threshold` (`is_over_threshold`),
  KEY `idx_uid_1` (`uid_1`),
  KEY `idx_uid_2` (`uid_2`),
  CONSTRAINT `plagiarism_result_ibfk_1` FOREIGN KEY (`check_id`) REFERENCES `plagiarism_check` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `plagiarism_result_ibfk_2` FOREIGN KEY (`cid`) REFERENCES `contest` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COMMENT='查重结果表';
