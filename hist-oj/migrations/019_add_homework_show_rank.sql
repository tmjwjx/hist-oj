-- 添加 show_rank 字段到 classroom_homework 表
-- 用于控制作业结束后是否向学生开放排行榜

ALTER TABLE `classroom_homework`
ADD COLUMN `show_rank` INT DEFAULT 0 COMMENT '作业结束后是否显示排行榜（0: 否, 1: 是）' AFTER `show_answer`;
