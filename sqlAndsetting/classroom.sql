-- =====================================================
-- 班级功能数据库表结构
-- 创建日期: 2025-01-11
-- 说明: 包含权限管理、班级、签到、题库、作业、资料库、随机选人、即时通讯等功能
-- =====================================================

USE hoj;

-- 1. 班级用户角色表
CREATE TABLE IF NOT EXISTS `classroom_user_role` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    `uid` VARCHAR(32) NOT NULL COMMENT '用户ID',
    `role` VARCHAR(20) NOT NULL COMMENT '角色：teacher/student/admin',
    `create_time` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY `uk_classroom_user_role` (`uid`, `role`),
    KEY `idx_uid` (`uid`),
    KEY `idx_role` (`role`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='班级用户角色表';

-- 2. 班级表
CREATE TABLE IF NOT EXISTS `classroom` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    `class_name` VARCHAR(100) NOT NULL COMMENT '班级名称，如：数据结构',
    `class_belong` VARCHAR(100) NOT NULL COMMENT '班级隶属，如：计科211',
    `class_code` VARCHAR(8) NOT NULL COMMENT '8位班级码',
    `teacher_id` VARCHAR(32) NOT NULL COMMENT '创建教师ID',
    `status` TINYINT DEFAULT 1 COMMENT '状态：1正常 0已删除',
    `create_time` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY `uk_class_code` (`class_code`),
    KEY `idx_teacher_id` (`teacher_id`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='班级表';

-- 3. 班级学生表
CREATE TABLE IF NOT EXISTS `classroom_student` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    `classroom_id` BIGINT NOT NULL COMMENT '班级ID',
    `uid` VARCHAR(32) NOT NULL COMMENT '学生用户ID',
    `real_name` VARCHAR(50) NOT NULL COMMENT '真实姓名',
    `gender` VARCHAR(10) DEFAULT NULL COMMENT '性别：男/女',
    `student_class` VARCHAR(100) DEFAULT NULL COMMENT '班级，如：计科211',
    `student_no` VARCHAR(50) DEFAULT NULL COMMENT '学号',
    `status` TINYINT DEFAULT 1 COMMENT '状态：1正常 0已移除',
    `create_time` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY `uk_classroom_user` (`classroom_id`, `uid`),
    KEY `idx_classroom_id` (`classroom_id`),
    KEY `idx_uid` (`uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='班级学生表';

-- 4. 签到表
CREATE TABLE IF NOT EXISTS `classroom_checkin` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    `classroom_id` BIGINT NOT NULL COMMENT '班级ID',
    `checkin_code` VARCHAR(20) NOT NULL COMMENT '签到码',
    `checkin_name` VARCHAR(100) DEFAULT NULL COMMENT '签到名称',
    `start_time` DATETIME NOT NULL COMMENT '开始时间',
    `end_time` DATETIME DEFAULT NULL COMMENT '结束时间',
    `status` TINYINT DEFAULT 1 COMMENT '状态：1进行中 2已结束',
    `create_time` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    KEY `idx_classroom_id` (`classroom_id`),
    KEY `idx_checkin_code` (`checkin_code`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='签到表';

-- 5. 签到记录表
CREATE TABLE IF NOT EXISTS `classroom_checkin_record` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    `checkin_id` BIGINT NOT NULL COMMENT '签到ID',
    `uid` VARCHAR(32) NOT NULL COMMENT '学生用户ID',
    `status` VARCHAR(20) DEFAULT 'present' COMMENT '状态：present(签到)/absent(缺勤)/sick_leave(病假)/personal_leave(事假)',
    `checkin_time` DATETIME DEFAULT NULL COMMENT '签到时间',
    `remark` VARCHAR(200) DEFAULT NULL COMMENT '备注',
    `create_time` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY `uk_checkin_student` (`checkin_id`, `uid`),
    KEY `idx_checkin_id` (`checkin_id`),
    KEY `idx_uid` (`uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='签到记录表';

-- 6. 题库表
CREATE TABLE IF NOT EXISTS `question_bank` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    `title` VARCHAR(200) NOT NULL COMMENT '题目标题',
    `type` VARCHAR(20) NOT NULL COMMENT '题型：choice(选择)/judge(判断)/subjective(主观)/programming(编程)',
    `content` TEXT NOT NULL COMMENT '题目内容（支持Markdown）',
    `options` JSON DEFAULT NULL COMMENT '选项（选择/判断题）',
    `answer` TEXT DEFAULT NULL COMMENT '标准答案',
    `difficulty` TINYINT DEFAULT 1 COMMENT '难度：1简单 2中等 3困难',
    `score` INT DEFAULT 2 COMMENT '默认分值',
    `creator_id` VARCHAR(32) NOT NULL COMMENT '创建者ID',
    `is_shared` TINYINT DEFAULT 0 COMMENT '是否共享：0个人 1共享',
    `problem_id` BIGINT DEFAULT NULL COMMENT '关联的OJ题目ID（编程题）',
    `status` TINYINT DEFAULT 1 COMMENT '状态：1正常 0已删除',
    `create_time` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    KEY `idx_creator_id` (`creator_id`),
    KEY `idx_type` (`type`),
    KEY `idx_is_shared` (`is_shared`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='题库表';

-- 7. 作业/考试表
CREATE TABLE IF NOT EXISTS `classroom_homework` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    `classroom_id` BIGINT NOT NULL COMMENT '班级ID',
    `title` VARCHAR(200) NOT NULL COMMENT '作业标题',
    `description` TEXT DEFAULT NULL COMMENT '作业说明（支持Markdown）',
    `start_time` DATETIME NOT NULL COMMENT '开始时间',
    `end_time` DATETIME NOT NULL COMMENT '结束时间',
    `show_score` TINYINT DEFAULT 1 COMMENT '完成后是否显示成绩',
    `show_homework` TINYINT DEFAULT 0 COMMENT '完成后是否查看作业题目',
    `status` TINYINT DEFAULT 1 COMMENT '状态：1未开始 2进行中 3已结束',
    `create_time` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    KEY `idx_classroom_id` (`classroom_id`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='作业/考试表';

-- 8. 作业题目关联表
CREATE TABLE IF NOT EXISTS `homework_question` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    `homework_id` BIGINT NOT NULL COMMENT '作业ID',
    `question_id` BIGINT NOT NULL COMMENT '题目ID',
    `question_order` INT NOT NULL COMMENT '题目序号',
    `score` INT DEFAULT 2 COMMENT '本题分值',
    `create_time` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    UNIQUE KEY `uk_homework_question` (`homework_id`, `question_id`),
    KEY `idx_homework_id` (`homework_id`),
    KEY `idx_question_id` (`question_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='作业题目关联表';

-- 9. 作业提交记录表
CREATE TABLE IF NOT EXISTS `homework_submit` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    `homework_id` BIGINT NOT NULL COMMENT '作业ID',
    `question_id` BIGINT NOT NULL COMMENT '题目ID',
    `uid` VARCHAR(32) NOT NULL COMMENT '学生ID',
    `answer` TEXT DEFAULT NULL COMMENT '学生答案',
    `submit_id` BIGINT DEFAULT NULL COMMENT '提交记录ID（编程题）',
    `score` DECIMAL(5,2) DEFAULT 0 COMMENT '得分',
    `is_scored` TINYINT DEFAULT 0 COMMENT '是否已批改',
    `create_time` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY `uk_homework_question_student` (`homework_id`, `question_id`, `uid`),
    KEY `idx_homework_id` (`homework_id`),
    KEY `idx_uid` (`uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='作业提交记录表';

-- 10. 资料库文件夹表
CREATE TABLE IF NOT EXISTS `classroom_folder` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    `classroom_id` BIGINT NOT NULL COMMENT '班级ID',
    `folder_name` VARCHAR(100) NOT NULL COMMENT '文件夹名称',
    `parent_id` BIGINT DEFAULT 0 COMMENT '父文件夹ID，0表示根目录',
    `creator_id` VARCHAR(32) NOT NULL COMMENT '创建者ID',
    `sort_order` INT DEFAULT 0 COMMENT '排序',
    `status` TINYINT DEFAULT 1 COMMENT '状态：1正常 0已删除',
    `create_time` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    KEY `idx_classroom_id` (`classroom_id`),
    KEY `idx_parent_id` (`parent_id`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='资料库文件夹表';

-- 11. 资料库文件表
CREATE TABLE IF NOT EXISTS `classroom_material` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    `folder_id` BIGINT NOT NULL COMMENT '文件夹ID',
    `file_name` VARCHAR(255) NOT NULL COMMENT '文件名',
    `file_type` VARCHAR(20) NOT NULL COMMENT '文件类型：pdf/word/ppt/txt/mp4',
    `file_path` VARCHAR(500) NOT NULL COMMENT '文件存储路径',
    `file_size` BIGINT DEFAULT NULL COMMENT '文件大小（字节）',
    `creator_id` VARCHAR(32) NOT NULL COMMENT '上传者ID',
    `is_shared` TINYINT DEFAULT 0 COMMENT '是否共享：0个人 1共享',
    `download_count` INT DEFAULT 0 COMMENT '下载次数',
    `status` TINYINT DEFAULT 1 COMMENT '状态：1正常 0已删除',
    `create_time` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    KEY `idx_folder_id` (`folder_id`),
    KEY `idx_creator_id` (`creator_id`),
    KEY `idx_is_shared` (`is_shared`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='资料库文件表';

-- 12. 随机选人记录表
CREATE TABLE IF NOT EXISTS `classroom_random_pick` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    `classroom_id` BIGINT NOT NULL COMMENT '班级ID',
    `picked_uid` VARCHAR(32) NOT NULL COMMENT '被选中学生ID',
    `pick_time` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '选人时间',
    KEY `idx_classroom_id` (`classroom_id`),
    KEY `idx_pick_time` (`pick_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='随机选人记录表';

-- 13. 班级消息表
CREATE TABLE IF NOT EXISTS `classroom_message` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    `classroom_id` BIGINT NOT NULL COMMENT '班级ID',
    `sender_id` VARCHAR(32) NOT NULL COMMENT '发送者ID',
    `content` TEXT DEFAULT NULL COMMENT '消息内容',
    `image_url` VARCHAR(500) DEFAULT NULL COMMENT '图片URL',
    `msg_type` VARCHAR(20) DEFAULT 'text' COMMENT '消息类型：text/image',
    `create_time` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    KEY `idx_classroom_id` (`classroom_id`),
    KEY `idx_create_time` (`create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='班级消息表';

-- =====================================================
-- 初始化数据
-- =====================================================

-- 为 root 用户添加管理员和教师角色
-- 注意：root 用户的 uuid 需要根据实际情况修改，请先查询 user_info 表获取 uuid
-- 示例: SELECT uuid FROM user_info WHERE username = 'root';

-- 取消下面的注释并替换成实际的 uuid
-- INSERT IGNORE INTO `classroom_user_role` (`uid`, `role`) VALUES ('root的uuid值', 'admin');
-- INSERT IGNORE INTO `classroom_user_role` (`uid`, `role`) VALUES ('root的uuid值', 'teacher');

