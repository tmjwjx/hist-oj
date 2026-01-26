-- ==========================================
-- 考试模式功能数据库迁移
-- 版本: 013
-- 日期: 2025-01-25
-- 说明: 添加考试模式相关字段和表
-- ==========================================

-- 1. classroom_homework 表新增考试模式字段
ALTER TABLE classroom_homework
ADD COLUMN is_exam_mode INT DEFAULT 0 COMMENT '是否考试模式(0否1是)',
ADD COLUMN exam_duration INT NOT NULL DEFAULT 60 COMMENT '考试时长(分钟)',
ADD COLUMN allow_submit_after_minutes INT DEFAULT 0 COMMENT '开考后多少分钟允许交卷(0表示立即允许)',
ADD COLUMN disable_copy_paste INT DEFAULT 1 COMMENT '是否禁止复制粘贴(0否1是)',
ADD COLUMN require_fullscreen INT DEFAULT 1 COMMENT '是否要求全屏(0否1是)',
ADD COLUMN disallow_tab_switch INT DEFAULT 1 COMMENT '是否禁止切换标签页(0否1是)';

-- 2. homework_submit 表新增考试相关字段
ALTER TABLE homework_submit
ADD COLUMN exam_start_time DATETIME COMMENT '考试开始时间',
ADD COLUMN exam_end_time DATETIME COMMENT '考试结束时间',
ADD COLUMN is_forced_submit INT DEFAULT 0 COMMENT '是否强制收卷(0否1是)',
ADD COLUMN tab_switch_count INT DEFAULT 0 COMMENT '切换标签页次数',
ADD COLUMN fullscreen_exit_count INT DEFAULT 0 COMMENT '退出全屏次数',
ADD COLUMN copy_paste_attempt_count INT DEFAULT 0 COMMENT '尝试复制粘贴次数',
ADD COLUMN device_info VARCHAR(500) COMMENT '设备信息',
ADD COLUMN browser_info VARCHAR(500) COMMENT '浏览器信息';

-- 3. 创建学生题目顺序映射表
CREATE TABLE IF NOT EXISTS student_question_order (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    homework_id BIGINT UNSIGNED NOT NULL,
    uid VARCHAR(32) NOT NULL,
    order_mapping TEXT NOT NULL COMMENT '题目顺序映射JSON',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_homework_uid (homework_id, uid),
    INDEX idx_homework_id (homework_id),
    INDEX idx_uid (uid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='学生题目顺序映射表';

-- 4. 创建考试违规日志表
CREATE TABLE IF NOT EXISTS exam_violation_log (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    homework_id BIGINT UNSIGNED NOT NULL,
    uid VARCHAR(32) NOT NULL,
    violation_type VARCHAR(50) NOT NULL COMMENT '违规类型(tab_switch, fullscreen_exit, copy_attempt, paste_attempt, context_menu, devtools_attempt)',
    description TEXT COMMENT '违规详情',
    ip VARCHAR(50) COMMENT 'IP地址',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_homework_id (homework_id),
    INDEX idx_uid (uid),
    INDEX idx_violation_type (violation_type),
    INDEX idx_create_time (create_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='考试违规日志表';

-- ==========================================
-- 迁移验证 SQL
-- ==========================================
-- 验证 classroom_homework 表字段是否添加成功
-- SELECT COLUMN_NAME, COLUMN_TYPE, COLUMN_COMMENT
-- FROM INFORMATION_SCHEMA.COLUMNS
-- WHERE TABLE_SCHEMA = DATABASE()
-- AND TABLE_NAME = 'classroom_homework'
-- AND COLUMN_NAME IN ('is_exam_mode', 'exam_duration', 'allow_submit_after_minutes', 'disable_copy_paste', 'require_fullscreen', 'disallow_tab_switch');

-- 验证 homework_submit 表字段是否添加成功
-- SELECT COLUMN_NAME, COLUMN_TYPE, COLUMN_COMMENT
-- FROM INFORMATION_SCHEMA.COLUMNS
-- WHERE TABLE_SCHEMA = DATABASE()
-- AND TABLE_NAME = 'homework_submit'
-- AND COLUMN_NAME IN ('exam_start_time', 'exam_end_time', 'is_forced_submit', 'tab_switch_count', 'fullscreen_exit_count', 'copy_paste_attempt_count', 'device_info', 'browser_info');

-- 验证新表是否创建成功
-- SHOW TABLES LIKE 'student_question_order';
-- SHOW TABLES LIKE 'exam_violation_log';
