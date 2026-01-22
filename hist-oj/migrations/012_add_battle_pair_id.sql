-- 添加 battle_pair_id 字段到 battle_record 表
-- 用于唯一标识一场对局，解决同一房间多次对局的记录匹配问题

-- 步骤1: 添加 battle_pair_id 字段（允许为 NULL）
ALTER TABLE `battle_record`
ADD COLUMN `battle_pair_id` VARCHAR(64) NULL COMMENT '对局唯一标识（同一场对局的两条记录共享此ID）'
AFTER `id`;

-- 步骤2: 添加索引
ALTER TABLE `battle_record`
ADD INDEX `idx_battle_pair_id` (`battle_pair_id`);

-- 步骤3: 为现有记录生成 battle_pair_id
-- 使用 room_id + problem_id + 创建时间（精确到分钟）作为标识
UPDATE `battle_record`
SET `battle_pair_id` = CONCAT(
    `room_id`,
    '_',
    COALESCE(`problem_id`, 'UNKNOWN'),
    '_',
    DATE_FORMAT(`gmt_create`, '%Y%m%d%H%i')
)
WHERE `battle_pair_id` IS NULL;

-- 步骤4: 将字段改为 NOT NULL
ALTER TABLE `battle_record`
MODIFY COLUMN `battle_pair_id` VARCHAR(64) NOT NULL COMMENT '对局唯一标识（同一场对局的两条记录共享此ID）';
