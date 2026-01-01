-- 判题终端提交历史表
CREATE TABLE IF NOT EXISTS `submission_history` (
  `id` INT AUTO_INCREMENT PRIMARY KEY,
  `submit_id` VARCHAR(50) COMMENT '提交ID',
  `pid` VARCHAR(50) COMMENT '题目ID',
  `cid` VARCHAR(50) DEFAULT '0' COMMENT '比赛ID，普通模式为0',
  `username` VARCHAR(100) COMMENT '用户名',
  `result` VARCHAR(50) COMMENT '判题结果',
  `time_used` VARCHAR(20) COMMENT '耗时',
  `memory_used` VARCHAR(20) COMMENT '内存使用',
  `language` VARCHAR(50) COMMENT '编程语言',
  `code` MEDIUMTEXT COMMENT '提交代码',
  `local_info` VARCHAR(100) DEFAULT '' COMMENT '本地测试信息',
  `submit_time` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '提交时间',
  INDEX `idx_pid_cid` (`pid`, `cid`),
  INDEX `idx_username` (`username`),
  INDEX `idx_submit_time` (`submit_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='判题终端提交历史';
