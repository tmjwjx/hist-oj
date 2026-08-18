-- 训练加入记录迁入 HOJ Java 训练模块。
CREATE TABLE IF NOT EXISTS `training_participant` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `training_id` bigint unsigned NOT NULL,
  `uid` varchar(32) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,
  `status` varchar(20) NOT NULL DEFAULT 'not_started',
  `join_time` datetime DEFAULT NULL,
  `gmt_create` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `gmt_modified` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_training_participant` (`training_id`, `uid`),
  KEY `idx_training_participant_uid` (`uid`),
  CONSTRAINT `fk_training_participant_training` FOREIGN KEY (`training_id`) REFERENCES `training` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_training_participant_user` FOREIGN KEY (`uid`) REFERENCES `user_info` (`uuid`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
