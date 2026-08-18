CREATE TABLE IF NOT EXISTS `problem_verification` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `pid` bigint(20) unsigned NOT NULL,
  `case_version` varchar(40) NOT NULL,
  `judge_mode` varchar(32) NOT NULL DEFAULT 'default',
  `sync_status` tinyint(4) NOT NULL DEFAULT '0' COMMENT '0待同步,1同步成功,2同步失败',
  `sync_message` varchar(1000) DEFAULT NULL,
  `verification_status` tinyint(4) NOT NULL DEFAULT '0' COMMENT '0待验题,1判题中,2通过,3失败',
  `submit_id` bigint(20) unsigned DEFAULT NULL,
  `verified_uid` varchar(32) DEFAULT NULL,
  `gmt_create` datetime DEFAULT CURRENT_TIMESTAMP,
  `gmt_modified` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_problem_verification_pid` (`pid`),
  KEY `idx_problem_verification_submit_id` (`submit_id`),
  CONSTRAINT `problem_verification_ibfk_1` FOREIGN KEY (`pid`) REFERENCES `problem` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `problem_verification_ibfk_2` FOREIGN KEY (`submit_id`) REFERENCES `judge` (`submit_id`) ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
