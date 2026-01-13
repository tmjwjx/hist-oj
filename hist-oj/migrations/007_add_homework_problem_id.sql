-- 为作业题目关联表添加 HOJ 题目ID字段
-- 用于支持编程题功能

-- 修改 homework_question 表
ALTER TABLE homework_question
ADD COLUMN problem_id VARCHAR(50) DEFAULT NULL COMMENT 'HOJ题目ID（编程题使用，字符串类型）' AFTER question_id,
ADD INDEX idx_problem_id (problem_id);

-- 修改 question_id 字段为可空（因为编程题可能不需要关联题库）
ALTER TABLE homework_question
MODIFY COLUMN question_id BIGINT UNSIGNED NULL;

-- 修改 homework_submit 表
ALTER TABLE homework_submit
ADD COLUMN problem_id VARCHAR(50) DEFAULT NULL COMMENT 'HOJ题目ID（编程题提交）' AFTER question_id,
ADD INDEX idx_problem_id (problem_id);

-- 修改 question_id 字段为可空
ALTER TABLE homework_submit
MODIFY COLUMN question_id BIGINT UNSIGNED NULL;
