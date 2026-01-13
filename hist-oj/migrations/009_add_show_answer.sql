-- 添加 show_answer 字段到 classroom_homework 表
-- 用于控制学生提交作业后是否可以查看答案

ALTER TABLE `classroom_homework`
ADD COLUMN `show_answer` INT DEFAULT 0 COMMENT '学生提交后是否可以查看答案（0: 否, 1: 是）' AFTER `show_homework`;
