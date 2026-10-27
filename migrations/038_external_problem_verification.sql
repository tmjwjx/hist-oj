CREATE TABLE IF NOT EXISTS `problem_external_verification` (
  `pid` bigint(20) unsigned NOT NULL,
  `token` char(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
  `enabled` tinyint(1) NOT NULL DEFAULT '0',
  `owner_uid` varchar(32) NOT NULL,
  `owner_username` varchar(255) NOT NULL,
  `gmt_create` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `gmt_modified` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`pid`),
  UNIQUE KEY `uk_external_verification_token` (`token`),
  CONSTRAINT `external_verification_problem_fk` FOREIGN KEY (`pid`) REFERENCES `problem` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `external_verification_submission` (
  `submit_id` bigint(20) unsigned NOT NULL,
  `pid` bigint(20) unsigned NOT NULL,
  `visitor_token_hash` char(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
  `nickname` varchar(40) NOT NULL DEFAULT '',
  `gmt_create` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`submit_id`),
  KEY `idx_external_submission_problem` (`pid`, `submit_id`),
  KEY `idx_external_submission_visitor` (`pid`, `visitor_token_hash`, `submit_id`),
  CONSTRAINT `external_submission_judge_fk` FOREIGN KEY (`submit_id`) REFERENCES `judge` (`submit_id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `external_submission_problem_fk` FOREIGN KEY (`pid`) REFERENCES `problem` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
