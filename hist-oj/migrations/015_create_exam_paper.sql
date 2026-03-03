-- 试卷表
CREATE TABLE IF NOT EXISTS `classroom_exam_paper` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `title` varchar(255) NOT NULL COMMENT '试卷标题',
  `creator_id` varchar(32) NOT NULL COMMENT '创建者UUID',
  `is_shared` tinyint(1) DEFAULT '0' COMMENT '是否共享：0=私有，1=共享',
  `total_score` int DEFAULT '0' COMMENT '总分',
  `question_count` int DEFAULT '0' COMMENT '题目数量',
  `description` text COMMENT '试卷描述',
  `status` tinyint(1) DEFAULT '1' COMMENT '状态：1=正常，0=删除',
  `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_creator_id` (`creator_id`),
  KEY `idx_is_shared` (`is_shared`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='试卷表';

-- 试卷题目关联表
CREATE TABLE IF NOT EXISTS `classroom_exam_paper_question` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `exam_paper_id` bigint unsigned NOT NULL COMMENT '试卷ID',
  `question_id` bigint unsigned DEFAULT NULL COMMENT '客观题ID（题库中的题目）',
  `problem_id` varchar(64) DEFAULT NULL COMMENT '编程题ID（BingOJ题目ID）',
  `question_order` int NOT NULL COMMENT '题目顺序',
  `question_type` varchar(50) NOT NULL COMMENT '题目类型：single_choice, multiple_choice, judge, subjective, programming',
  `score` int NOT NULL DEFAULT '0' COMMENT '分值',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_exam_paper_order` (`exam_paper_id`, `question_order`),
  KEY `idx_exam_paper_id` (`exam_paper_id`),
  KEY `idx_question_id` (`question_id`),
  KEY `idx_problem_id` (`problem_id`),
  CONSTRAINT `fk_cepap_exam_paper` FOREIGN KEY (`exam_paper_id`) REFERENCES `classroom_exam_paper` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='试卷题目关联表';
