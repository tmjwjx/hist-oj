-- 为试卷表新增“是否公开到主界面”字段
ALTER TABLE `classroom_exam_paper`
ADD COLUMN `is_public` tinyint(1) DEFAULT '0' COMMENT '是否公开到主界面：0=否，1=是' AFTER `is_shared`;

-- 新增索引，便于主界面公开试卷列表查询
ALTER TABLE `classroom_exam_paper`
ADD INDEX `idx_is_public` (`is_public`);
