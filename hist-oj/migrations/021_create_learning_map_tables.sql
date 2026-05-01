-- 创建算法知识点与题目航海图相关表
-- 执行前请先: USE hoj;

CREATE TABLE IF NOT EXISTS `learning_map` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `title` varchar(120) NOT NULL,
  `description` text,
  `status` varchar(20) NOT NULL DEFAULT 'draft',
  `access_mode` varchar(20) NOT NULL DEFAULT 'all_open',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_status` (`status`),
  KEY `idx_access_mode` (`access_mode`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='算法航海图';

CREATE TABLE IF NOT EXISTS `learning_map_node` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `map_id` bigint unsigned NOT NULL,
  `type` varchar(20) NOT NULL,
  `title` varchar(160) NOT NULL,
  `description` text,
  `difficulty` varchar(20) DEFAULT 'beginner',
  `tags` text,
  `x` double NOT NULL DEFAULT 0,
  `y` double NOT NULL DEFAULT 0,
  `level` int DEFAULT 0,
  `region` varchar(100) DEFAULT NULL,
  `published` tinyint(1) NOT NULL DEFAULT 1,
  `knowledge_content` longtext,
  `problem_id` bigint unsigned DEFAULT NULL,
  `problem_display_id` varchar(80) DEFAULT NULL,
  `metadata` longtext,
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_map_id` (`map_id`),
  KEY `idx_type` (`type`),
  KEY `idx_problem_id` (`problem_id`),
  KEY `idx_problem_display_id` (`problem_display_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='算法航海图节点';

CREATE TABLE IF NOT EXISTS `learning_map_edge` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `map_id` bigint unsigned NOT NULL,
  `source_node_id` bigint unsigned NOT NULL,
  `target_node_id` bigint unsigned NOT NULL,
  `type` varchar(20) NOT NULL DEFAULT 'prerequisite',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_map_id` (`map_id`),
  KEY `idx_source_node_id` (`source_node_id`),
  KEY `idx_target_node_id` (`target_node_id`),
  KEY `idx_edge_type` (`type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='算法航海图连线';

CREATE TABLE IF NOT EXISTS `user_learning_progress` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` varchar(32) NOT NULL,
  `map_id` bigint unsigned NOT NULL,
  `node_id` bigint unsigned NOT NULL,
  `status` varchar(20) NOT NULL DEFAULT 'locked',
  `completed_at` datetime DEFAULT NULL,
  `mastered_at` datetime DEFAULT NULL,
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_map_node` (`user_id`, `map_id`, `node_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_map_id` (`map_id`),
  KEY `idx_node_id` (`node_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户算法航海图进度';

CREATE TABLE IF NOT EXISTS `learning_map_permission` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `map_id` bigint unsigned NOT NULL,
  `user_id` varchar(32) NOT NULL,
  `enabled` tinyint(1) NOT NULL DEFAULT 1,
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_map_user` (`map_id`, `user_id`),
  KEY `idx_map_id` (`map_id`),
  KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='航海图用户权限覆盖';
