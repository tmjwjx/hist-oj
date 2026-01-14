-- 为作业提交记录表添加附件字段
-- 用于支持主观题上传图片功能

USE hoj;

-- 添加 attachment 字段
ALTER TABLE `homework_submit`
ADD COLUMN `attachment` VARCHAR(1000) DEFAULT NULL COMMENT '图片附件URL（主观题使用，多个图片用逗号分隔）' AFTER `answer`;

-- 添加索引以提高查询性能
ALTER TABLE `homework_submit`
ADD INDEX `idx_attachment` (`attachment`(255));
