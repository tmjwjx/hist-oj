CREATE TABLE IF NOT EXISTS `problem_pending_save` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `uid` varchar(32) NOT NULL,
  `target_pid` bigint unsigned DEFAULT NULL,
  `candidate_pid` bigint unsigned DEFAULT NULL,
  `payload` mediumtext NOT NULL,
  `request_fingerprint` varchar(64) NOT NULL,
  `base_fingerprint` varchar(64) DEFAULT NULL,
  `candidate_fingerprint` varchar(64) DEFAULT NULL,
  `published_pid` bigint unsigned DEFAULT NULL,
  `superseded` tinyint(1) NOT NULL DEFAULT 0,
  `gmt_create` datetime DEFAULT CURRENT_TIMESTAMP,
  `gmt_modified` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_pending_candidate` (`candidate_pid`),
  KEY `idx_pending_owner` (`uid`,`target_pid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
