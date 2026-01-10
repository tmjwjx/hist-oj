-- 为旧的对战记录回填 opponent_rating 数据
-- 通过房间号匹配同一房间的两条记录，互换双方的 rating

-- 更新房主记录的 opponent_rating（使用挑战者的 rating）
UPDATE battle_record br1
INNER JOIN battle_record br2
  ON br1.room_id = br2.room_id
  AND br1.user_id != br2.user_id
INNER JOIN user_rating ur
  ON br2.user_id = ur.uid
SET br1.opponent_rating = ur.hist_rating
WHERE br1.opponent_rating IS NULL
  AND br1.id < br2.id;  -- 只更新房主记录，避免重复

-- 更新挑战者记录的 opponent_rating（使用房主的 rating）
UPDATE battle_record br1
INNER JOIN battle_record br2
  ON br1.room_id = br2.room_id
  AND br1.user_id != br2.user_id
INNER JOIN user_rating ur
  ON br2.user_id = ur.uid
SET br1.opponent_rating = ur.hist_rating
WHERE br1.opponent_rating IS NULL
  AND br1.id > br2.id;  -- 只更新挑战者记录，避免重复

-- 如果找不到 user_rating 表，尝试从 user 表获取
-- 备用方案：如果对手还在系统中，从 user 表获取 rating
UPDATE battle_record br1
INNER JOIN battle_record br2
  ON br1.room_id = br2.room_id
  AND br1.user_id != br2.user_id
INNER JOIN `user` u
  ON br2.user_id = u.uuid
SET br1.opponent_rating = u.rating
WHERE br1.opponent_rating IS NULL
  AND u.rating IS NOT NULL;
