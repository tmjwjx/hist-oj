-- 将原 Go 独立报名系统并入 HOJ 的 contest / contest_register。
ALTER TABLE `contest`
    ADD COLUMN `open_registration` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否要求填写比赛报名信息' AFTER `allow_end_submit`,
    ADD COLUMN `registration_fields` text COMMENT '报名字段配置 JSON' AFTER `open_registration`,
    ADD COLUMN `use_registration_name` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否使用报名信息组合比赛内名称' AFTER `registration_fields`,
    ADD COLUMN `registration_name_fields` text COMMENT '比赛内名称组合字段 JSON' AFTER `use_registration_name`;

ALTER TABLE `contest_register`
    ADD COLUMN `name` varchar(100) DEFAULT NULL COMMENT '姓名' AFTER `status`,
    ADD COLUMN `class` varchar(100) DEFAULT NULL COMMENT '班级' AFTER `name`,
    ADD COLUMN `college` varchar(100) DEFAULT NULL COMMENT '学院' AFTER `class`,
    ADD COLUMN `student_id` varchar(50) DEFAULT NULL COMMENT '学号' AFTER `college`,
    ADD COLUMN `gender` varchar(10) DEFAULT NULL COMMENT '性别' AFTER `student_id`,
    ADD COLUMN `qq` varchar(20) DEFAULT NULL COMMENT 'QQ' AFTER `gender`,
    ADD COLUMN `phone` varchar(30) DEFAULT NULL COMMENT '电话号码' AFTER `qq`;
