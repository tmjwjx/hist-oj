-- 为比赛 1002 插入测试 Rating 数据
-- 这个脚本会为比赛的前 10 名参赛者插入 Rating 变化记录

-- 首先创建比赛 Rating 状态记录
INSERT INTO contest_rating_status (contest_id, is_rated, rating_calculated, calculated_at)
VALUES (1002, 1, 1, NOW())
ON DUPLICATE KEY UPDATE
    is_rated = 1,
    rating_calculated = 1,
    calculated_at = NOW();

-- 插入测试 Rating 历史记录
-- 假设比赛前 10 名的 Rating 变化

-- 第 1 名：肖 (uid 假设为对应的用户 ID)
INSERT INTO rating_history (uid, contest_id, rank, old_rating, new_rating, rating_change, contest_time)
SELECT uid, 1002, 1, 1500, 1550, 50, NOW()
FROM user_info WHERE username = '肖'
ON DUPLICATE KEY UPDATE
    old_rating = 1500,
    new_rating = 1550,
    rating_change = 50;

-- 第 2 名：龚晴晴
INSERT INTO rating_history (uid, contest_id, rank, old_rating, new_rating, rating_change, contest_time)
SELECT uid, 1002, 2, 1480, 1520, 40, NOW()
FROM user_info WHERE username = '龚晴晴'
ON DUPLICATE KEY UPDATE
    old_rating = 1480,
    new_rating = 1520,
    rating_change = 40;

-- 第 3 名：张家任
INSERT INTO rating_history (uid, contest_id, rank, old_rating, new_rating, rating_change, contest_time)
SELECT uid, 1002, 3, 1460, 1495, 35, NOW()
FROM user_info WHERE username = '张家任'
ON DUPLICATE KEY UPDATE
    old_rating = 1460,
    new_rating = 1495,
    rating_change = 35;

-- 第 4 名：孙景逸
INSERT INTO rating_history (uid, contest_id, rank, old_rating, new_rating, rating_change, contest_time)
SELECT uid, 1002, 4, 1450, 1480, 30, NOW()
FROM user_info WHERE username = '孙景逸'
ON DUPLICATE KEY UPDATE
    old_rating = 1450,
    new_rating = 1480,
    rating_change = 30;

-- 第 5 名：马华恩
INSERT INTO rating_history (uid, contest_id, rank, old_rating, new_rating, rating_change, contest_time)
SELECT uid, 1002, 5, 1440, 1465, 25, NOW()
FROM user_info WHERE username = '马华恩'
ON DUPLICATE KEY UPDATE
    old_rating = 1440,
    new_rating = 1465,
    rating_change = 25;

-- 第 6 名：周深轩
INSERT INTO rating_history (uid, contest_id, rank, old_rating, new_rating, rating_change, contest_time)
SELECT uid, 1002, 6, 1430, 1450, 20, NOW()
FROM user_info WHERE username = '周深轩'
ON DUPLICATE KEY UPDATE
    old_rating = 1430,
    new_rating = 1450,
    rating_change = 20;

-- 第 7 名：李恒洋
INSERT INTO rating_history (uid, contest_id, rank, old_rating, new_rating, rating_change, contest_time)
SELECT uid, 1002, 7, 1420, 1435, 15, NOW()
FROM user_info WHERE username = '李恒洋'
ON DUPLICATE KEY UPDATE
    old_rating = 1420,
    new_rating = 1435,
    rating_change = 15;

-- 第 8 名：刘志文
INSERT INTO rating_history (uid, contest_id, rank, old_rating, new_rating, rating_change, contest_time)
SELECT uid, 1002, 8, 1500, 1490, -10, NOW()
FROM user_info WHERE username = '刘志文'
ON DUPLICATE KEY UPDATE
    old_rating = 1500,
    new_rating = 1490,
    rating_change = -10;

-- 第 9 名：王科林
INSERT INTO rating_history (uid, contest_id, rank, old_rating, new_rating, rating_change, contest_time)
SELECT uid, 1002, 9, 1510, 1495, -15, NOW()
FROM user_info WHERE username = '王科林'
ON DUPLICATE KEY UPDATE
    old_rating = 1510,
    new_rating = 1495,
    rating_change = -15;

-- 第 10 名：赵问悦
INSERT INTO rating_history (uid, contest_id, rank, old_rating, new_rating, rating_change, contest_time)
SELECT uid, 1002, 10, 1520, 1500, -20, NOW()
FROM user_info WHERE username = '赵问悦'
ON DUPLICATE KEY UPDATE
    old_rating = 1520,
    new_rating = 1500,
    rating_change = -20;

-- 查询插入结果
SELECT
    rh.uid,
    ui.username,
    rh.rank,
    rh.old_rating,
    rh.new_rating,
    rh.rating_change
FROM rating_history rh
JOIN user_info ui ON rh.uid = ui.uid
WHERE rh.contest_id = 1002
ORDER BY rh.rank;
