-- 添加 is_officially_submitted 字段到 homework_submit 表
-- 用于区分草稿自动保存和用户正式提交
-- 0 = 草稿自动保存（默认）
-- 1 = 用户点击了"提交作业"按钮，正式提交

ALTER TABLE `homework_submit`
ADD COLUMN `is_officially_submitted` INT(1) NOT NULL DEFAULT 0 COMMENT '是否已正式提交（0=草稿自动保存，1=用户点击提交）'
AFTER `is_scored`;

-- 添加索引以提高查询性能
ALTER TABLE `homework_submit`
ADD INDEX `idx_is_officially_submitted` (`is_officially_submitted`);
