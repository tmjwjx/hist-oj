-- 创建 training_register 表（HOJ 私有训练注册表）
CREATE TABLE IF NOT EXISTS `training_register` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `tid` bigint unsigned NOT NULL COMMENT '训练id',
  `uid` varchar(32) NOT NULL COMMENT '用户id',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否可用',
  `gmt_create` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `gmt_modified` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  PRIMARY KEY (`id`),
  KEY `idx_tid` (`tid`),
  KEY `idx_uid` (`uid`),
  UNIQUE KEY `uk_tid_uid` (`tid`, `uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='私有训练注册记录表';
