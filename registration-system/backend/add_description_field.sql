-- 为比赛表添加比赛说明字段
ALTER TABLE histcontest_register_competitions ADD COLUMN description TEXT;

-- 如果已经有数据，初始化为空字符串
UPDATE histcontest_register_competitions SET description = '' WHERE description IS NULL;
