-- =====================================================
-- 班级角色申请表
-- 创建日期: 2025-03-06
-- 说明: 用于用户申请教师或学生角色
-- =====================================================

USE hoj;

-- 创建班级角色申请表
CREATE TABLE IF NOT EXISTS `classroom_role_request` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    `uid` VARCHAR(32) NOT NULL COMMENT '申请用户ID',
    `role` VARCHAR(20) NOT NULL COMMENT '申请角色：teacher/student',
    `reason` TEXT COMMENT '申请理由',
    `status` INT DEFAULT 0 COMMENT '状态：0待审批 1已批准 2已拒绝',
    `reviewer_uid` VARCHAR(32) DEFAULT NULL COMMENT '审批人用户ID',
    `review_time` DATETIME DEFAULT NULL COMMENT '审批时间',
    `review_note` TEXT COMMENT '审批备注',
    `create_time` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    KEY `idx_uid` (`uid`),
    KEY `idx_role` (`role`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='班级角色申请表';
