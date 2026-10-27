-- 课堂作业提交在进入判题队列前永久绑定来源，避免历史提交和考试信息进入公共列表。
-- 必须先执行本迁移，再发布使用 homework_id 的后端及判题服务。
DROP PROCEDURE IF EXISTS hist_add_classroom_judge_source;
DELIMITER $$
CREATE PROCEDURE hist_add_classroom_judge_source()
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'judge' AND COLUMN_NAME = 'homework_id'
  ) THEN
    ALTER TABLE judge ADD COLUMN homework_id bigint unsigned DEFAULT NULL COMMENT '所属课堂作业id';
  END IF;
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.STATISTICS
    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'judge' AND INDEX_NAME = 'idx_judge_homework'
  ) THEN
    ALTER TABLE judge ADD INDEX idx_judge_homework (homework_id);
  END IF;
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.STATISTICS
    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'homework_submit' AND INDEX_NAME = 'idx_homework_submit_judge'
  ) THEN
    ALTER TABLE homework_submit ADD INDEX idx_homework_submit_judge (submit_id);
  END IF;
END$$
DELIMITER ;
CALL hist_add_classroom_judge_source();
DROP PROCEDURE hist_add_classroom_judge_source;

-- 仅回填仍保留、且提交者与题目一致的作业关联；旧版被后续提交覆盖的关联无法可靠推断。
UPDATE judge j
JOIN (
  SELECT submit_id, uid, problem_id, MIN(homework_id) AS homework_id
  FROM homework_submit
  WHERE submit_id IS NOT NULL
  GROUP BY submit_id, uid, problem_id
) hs ON hs.submit_id = j.submit_id AND hs.uid = j.uid AND hs.problem_id = j.display_pid
SET j.homework_id = hs.homework_id, j.share = 0, j.gmt_modified = j.gmt_modified
WHERE j.homework_id IS NULL;

UPDATE judge SET share = 0, gmt_modified = gmt_modified WHERE homework_id IS NOT NULL AND share <> 0;
