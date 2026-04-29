-- 算法知识点与题目航海图 mock 数据
-- 说明：
-- 1) 该示例地图默认草稿状态（draft），避免在主题目库未就绪时发布校验失败。
-- 2) problem 类型节点使用模拟主 OJ 题号（MOCK-P001 ~ MOCK-P030）。

INSERT INTO learning_map (title, description, status, create_time, update_time)
VALUES (
  '算法知识点与题目航海图（示例）',
  '示例数据：用于演示知识点 + 题目节点、前置依赖、进度解锁逻辑。',
  'draft',
  NOW(), NOW()
);

SET @map_id = LAST_INSERT_ID();

-- ========== 知识点节点（20） ==========
INSERT INTO learning_map_node (
  map_id, type, title, description, difficulty, tags, x, y, level, region, published,
  knowledge_content, metadata, create_time, update_time
)
VALUES
(@map_id, 'knowledge', '数据结构', '基础数据结构总览', 'beginner', '["基础"]', 120, 120, 1, '基础区', 1, '# 数据结构\n掌握数组、链表、栈、队列等核心结构。', '{}', NOW(), NOW()),
(@map_id, 'knowledge', '数组', '线性结构入门', 'beginner', '["线性结构"]', 320, 80, 1, '基础区', 1, '# 数组\n学习下标访问、前缀和、双指针。', '{}', NOW(), NOW()),
(@map_id, 'knowledge', '链表', '指针结构与操作', 'easy', '["线性结构"]', 320, 180, 1, '基础区', 1, '# 链表\n掌握增删改查与快慢指针。', '{}', NOW(), NOW()),
(@map_id, 'knowledge', '栈', '后进先出结构', 'easy', '["线性结构"]', 320, 280, 1, '基础区', 1, '# 栈\n典型应用：括号匹配、单调栈。', '{}', NOW(), NOW()),
(@map_id, 'knowledge', '队列', '先进先出结构', 'easy', '["线性结构"]', 320, 380, 1, '基础区', 1, '# 队列\n典型应用：BFS、单调队列。', '{}', NOW(), NOW()),
(@map_id, 'knowledge', '单调栈', '维护单调性的栈技巧', 'medium', '["栈","优化"]', 560, 280, 2, '结构进阶区', 1, '# 单调栈\n用于区间最值、下一个更大元素。', '{}', NOW(), NOW()),
(@map_id, 'knowledge', '单调队列', '维护窗口最值', 'medium', '["队列","优化"]', 560, 380, 2, '结构进阶区', 1, '# 单调队列\n用于滑动窗口最值。', '{}', NOW(), NOW()),
(@map_id, 'knowledge', '并查集', '集合合并与连通性', 'medium', '["图论","数据结构"]', 560, 120, 2, '结构进阶区', 1, '# 并查集\n支持合并集合与查询连通性。', '{}', NOW(), NOW()),
(@map_id, 'knowledge', '树状数组', '前缀统计结构', 'medium', '["树结构","前缀"]', 760, 80, 3, '树结构区', 1, '# 树状数组\n支持单点修改、前缀和查询。', '{}', NOW(), NOW()),
(@map_id, 'knowledge', '线段树', '区间查询与修改', 'hard', '["树结构","区间"]', 760, 180, 3, '树结构区', 1, '# 线段树\n支持区间查询与区间更新。', '{}', NOW(), NOW()),
(@map_id, 'knowledge', '字符串', '字符串算法总览', 'easy', '["字符串"]', 120, 620, 1, '字符串区', 1, '# 字符串\n掌握匹配、哈希、字典树。', '{}', NOW(), NOW()),
(@map_id, 'knowledge', '哈希', '字符串哈希与判重', 'medium', '["字符串","哈希"]', 320, 580, 2, '字符串区', 1, '# 哈希\n常见技巧：滚动哈希、集合去重。', '{}', NOW(), NOW()),
(@map_id, 'knowledge', 'KMP', '线性匹配算法', 'medium', '["字符串","匹配"]', 320, 680, 2, '字符串区', 1, '# KMP\n掌握 next 数组与失配跳转。', '{}', NOW(), NOW()),
(@map_id, 'knowledge', 'Trie', '字典树', 'medium', '["字符串","树"]', 320, 780, 2, '字符串区', 1, '# Trie\n用于前缀匹配与统计。', '{}', NOW(), NOW()),
(@map_id, 'knowledge', '搜索', '搜索策略总览', 'beginner', '["搜索"]', 120, 980, 1, '搜索区', 1, '# 搜索\n学习 DFS、BFS、状态搜索。', '{}', NOW(), NOW()),
(@map_id, 'knowledge', 'DFS', '深度优先搜索', 'easy', '["搜索","图论"]', 320, 940, 2, '搜索区', 1, '# DFS\n用于遍历、回溯、连通分量。', '{}', NOW(), NOW()),
(@map_id, 'knowledge', 'BFS', '广度优先搜索', 'easy', '["搜索","最短路"]', 320, 1040, 2, '搜索区', 1, '# BFS\n用于最短路层序遍历。', '{}', NOW(), NOW()),
(@map_id, 'knowledge', '二分搜索', '答案二分与区间二分', 'medium', '["搜索","有序"]', 560, 940, 3, '搜索区', 1, '# 二分搜索\n用于有序判定与答案单调性。', '{}', NOW(), NOW()),
(@map_id, 'knowledge', '动态规划', '状态转移思想', 'medium', '["DP"]', 120, 1260, 1, 'DP区', 1, '# 动态规划\n掌握状态设计与转移方程。', '{}', NOW(), NOW()),
(@map_id, 'knowledge', '背包 DP', '经典背包模型', 'hard', '["DP","背包"]', 320, 1260, 2, 'DP区', 1, '# 背包 DP\n01 背包、完全背包、多重背包。', '{}', NOW(), NOW());

-- ========== 题目节点（30，模拟主 OJ 题号） ==========
INSERT INTO learning_map_node (
  map_id, type, title, description, difficulty, tags, x, y, level, region, published,
  knowledge_content, problem_display_id, metadata, create_time, update_time
)
VALUES
(@map_id, 'problem', '单调栈基础题 1', '模拟题目节点', 'easy', '["单调栈"]', 840, 260, 3, '题目区A', 1, '', 'MOCK-P001', '{}', NOW(), NOW()),
(@map_id, 'problem', '单调栈基础题 2', '模拟题目节点', 'easy', '["单调栈"]', 980, 260, 3, '题目区A', 1, '', 'MOCK-P002', '{}', NOW(), NOW()),
(@map_id, 'problem', '单调栈进阶题 1', '模拟题目节点', 'medium', '["单调栈"]', 1120, 260, 4, '题目区A', 1, '', 'MOCK-P003', '{}', NOW(), NOW()),
(@map_id, 'problem', '单调队列基础题 1', '模拟题目节点', 'easy', '["单调队列"]', 840, 380, 3, '题目区A', 1, '', 'MOCK-P004', '{}', NOW(), NOW()),
(@map_id, 'problem', '单调队列进阶题 1', '模拟题目节点', 'medium', '["单调队列"]', 980, 380, 4, '题目区A', 1, '', 'MOCK-P005', '{}', NOW(), NOW()),
(@map_id, 'problem', '并查集基础题 1', '模拟题目节点', 'easy', '["并查集"]', 840, 120, 3, '题目区B', 1, '', 'MOCK-P006', '{}', NOW(), NOW()),
(@map_id, 'problem', '并查集进阶题 1', '模拟题目节点', 'medium', '["并查集"]', 980, 120, 4, '题目区B', 1, '', 'MOCK-P007', '{}', NOW(), NOW()),
(@map_id, 'problem', '树状数组基础题 1', '模拟题目节点', 'easy', '["树状数组"]', 980, 80, 4, '题目区B', 1, '', 'MOCK-P008', '{}', NOW(), NOW()),
(@map_id, 'problem', '树状数组进阶题 1', '模拟题目节点', 'medium', '["树状数组"]', 1120, 80, 5, '题目区B', 1, '', 'MOCK-P009', '{}', NOW(), NOW()),
(@map_id, 'problem', '线段树基础题 1', '模拟题目节点', 'medium', '["线段树"]', 980, 180, 4, '题目区B', 1, '', 'MOCK-P010', '{}', NOW(), NOW()),
(@map_id, 'problem', '线段树进阶题 1', '模拟题目节点', 'hard', '["线段树"]', 1120, 180, 5, '题目区B', 1, '', 'MOCK-P011', '{}', NOW(), NOW()),
(@map_id, 'problem', '哈希基础题 1', '模拟题目节点', 'easy', '["字符串","哈希"]', 560, 580, 3, '题目区C', 1, '', 'MOCK-P012', '{}', NOW(), NOW()),
(@map_id, 'problem', 'KMP基础题 1', '模拟题目节点', 'easy', '["KMP"]', 560, 680, 3, '题目区C', 1, '', 'MOCK-P013', '{}', NOW(), NOW()),
(@map_id, 'problem', 'Trie基础题 1', '模拟题目节点', 'easy', '["Trie"]', 560, 780, 3, '题目区C', 1, '', 'MOCK-P014', '{}', NOW(), NOW()),
(@map_id, 'problem', '字符串综合题 1', '模拟题目节点', 'medium', '["字符串"]', 700, 680, 4, '题目区C', 1, '', 'MOCK-P015', '{}', NOW(), NOW()),
(@map_id, 'problem', 'DFS基础题 1', '模拟题目节点', 'easy', '["DFS"]', 560, 940, 3, '题目区D', 1, '', 'MOCK-P016', '{}', NOW(), NOW()),
(@map_id, 'problem', 'BFS基础题 1', '模拟题目节点', 'easy', '["BFS"]', 560, 1040, 3, '题目区D', 1, '', 'MOCK-P017', '{}', NOW(), NOW()),
(@map_id, 'problem', 'BFS最短路入门题', '模拟题目节点', 'medium', '["BFS","最短路"]', 700, 1040, 4, '题目区D', 1, '', 'MOCK-P018', '{}', NOW(), NOW()),
(@map_id, 'problem', '二分搜索基础题 1', '模拟题目节点', 'easy', '["二分"]', 760, 940, 4, '题目区D', 1, '', 'MOCK-P019', '{}', NOW(), NOW()),
(@map_id, 'problem', '二分搜索进阶题 1', '模拟题目节点', 'medium', '["二分"]', 900, 940, 5, '题目区D', 1, '', 'MOCK-P020', '{}', NOW(), NOW()),
(@map_id, 'problem', 'DP基础题 1', '模拟题目节点', 'easy', '["DP"]', 560, 1220, 3, '题目区E', 1, '', 'MOCK-P021', '{}', NOW(), NOW()),
(@map_id, 'problem', '背包DP基础题 1', '模拟题目节点', 'medium', '["背包DP"]', 560, 1300, 3, '题目区E', 1, '', 'MOCK-P022', '{}', NOW(), NOW()),
(@map_id, 'problem', '背包DP进阶题 1', '模拟题目节点', 'hard', '["背包DP"]', 700, 1300, 4, '题目区E', 1, '', 'MOCK-P023', '{}', NOW(), NOW()),
(@map_id, 'problem', '数组基础题 1', '模拟题目节点', 'easy', '["数组"]', 560, 60, 2, '题目区A', 1, '', 'MOCK-P024', '{}', NOW(), NOW()),
(@map_id, 'problem', '链表基础题 1', '模拟题目节点', 'easy', '["链表"]', 560, 180, 2, '题目区A', 1, '', 'MOCK-P025', '{}', NOW(), NOW()),
(@map_id, 'problem', '栈基础题 1', '模拟题目节点', 'easy', '["栈"]', 560, 240, 2, '题目区A', 1, '', 'MOCK-P026', '{}', NOW(), NOW()),
(@map_id, 'problem', '队列基础题 1', '模拟题目节点', 'easy', '["队列"]', 560, 340, 2, '题目区A', 1, '', 'MOCK-P027', '{}', NOW(), NOW()),
(@map_id, 'problem', '搜索综合题 1', '模拟题目节点', 'medium', '["搜索"]', 840, 980, 4, '题目区D', 1, '', 'MOCK-P028', '{}', NOW(), NOW()),
(@map_id, 'problem', '字符串进阶题 2', '模拟题目节点', 'hard', '["字符串"]', 840, 700, 5, '题目区C', 1, '', 'MOCK-P029', '{}', NOW(), NOW()),
(@map_id, 'problem', '数据结构综合题 1', '模拟题目节点', 'hard', '["数据结构"]', 1240, 220, 6, '题目区A', 1, '', 'MOCK-P030', '{}', NOW(), NOW());

-- ========== 依赖连线（prerequisite） ==========

-- 基础知识依赖
INSERT INTO learning_map_edge (map_id, source_node_id, target_node_id, type, create_time, update_time)
SELECT @map_id, s.id, t.id, 'prerequisite', NOW(), NOW()
FROM learning_map_node s JOIN learning_map_node t
WHERE s.map_id=@map_id AND t.map_id=@map_id AND s.title='数据结构' AND t.title='数组';

INSERT INTO learning_map_edge (map_id, source_node_id, target_node_id, type, create_time, update_time)
SELECT @map_id, s.id, t.id, 'prerequisite', NOW(), NOW()
FROM learning_map_node s JOIN learning_map_node t
WHERE s.map_id=@map_id AND t.map_id=@map_id AND s.title='数据结构' AND t.title='链表';

INSERT INTO learning_map_edge (map_id, source_node_id, target_node_id, type, create_time, update_time)
SELECT @map_id, s.id, t.id, 'prerequisite', NOW(), NOW()
FROM learning_map_node s JOIN learning_map_node t
WHERE s.map_id=@map_id AND t.map_id=@map_id AND s.title='数据结构' AND t.title='栈';

INSERT INTO learning_map_edge (map_id, source_node_id, target_node_id, type, create_time, update_time)
SELECT @map_id, s.id, t.id, 'prerequisite', NOW(), NOW()
FROM learning_map_node s JOIN learning_map_node t
WHERE s.map_id=@map_id AND t.map_id=@map_id AND s.title='数据结构' AND t.title='队列';

INSERT INTO learning_map_edge (map_id, source_node_id, target_node_id, type, create_time, update_time)
SELECT @map_id, s.id, t.id, 'prerequisite', NOW(), NOW()
FROM learning_map_node s JOIN learning_map_node t
WHERE s.map_id=@map_id AND t.map_id=@map_id AND s.title='数组' AND t.title='单调栈';

INSERT INTO learning_map_edge (map_id, source_node_id, target_node_id, type, create_time, update_time)
SELECT @map_id, s.id, t.id, 'prerequisite', NOW(), NOW()
FROM learning_map_node s JOIN learning_map_node t
WHERE s.map_id=@map_id AND t.map_id=@map_id AND s.title='栈' AND t.title='单调栈';

INSERT INTO learning_map_edge (map_id, source_node_id, target_node_id, type, create_time, update_time)
SELECT @map_id, s.id, t.id, 'prerequisite', NOW(), NOW()
FROM learning_map_node s JOIN learning_map_node t
WHERE s.map_id=@map_id AND t.map_id=@map_id AND s.title='队列' AND t.title='单调队列';

INSERT INTO learning_map_edge (map_id, source_node_id, target_node_id, type, create_time, update_time)
SELECT @map_id, s.id, t.id, 'prerequisite', NOW(), NOW()
FROM learning_map_node s JOIN learning_map_node t
WHERE s.map_id=@map_id AND t.map_id=@map_id AND s.title='数组' AND t.title='二分搜索';

INSERT INTO learning_map_edge (map_id, source_node_id, target_node_id, type, create_time, update_time)
SELECT @map_id, s.id, t.id, 'prerequisite', NOW(), NOW()
FROM learning_map_node s JOIN learning_map_node t
WHERE s.map_id=@map_id AND t.map_id=@map_id AND s.title='数组' AND t.title='树状数组';

INSERT INTO learning_map_edge (map_id, source_node_id, target_node_id, type, create_time, update_time)
SELECT @map_id, s.id, t.id, 'prerequisite', NOW(), NOW()
FROM learning_map_node s JOIN learning_map_node t
WHERE s.map_id=@map_id AND t.map_id=@map_id AND s.title='树状数组' AND t.title='线段树';

INSERT INTO learning_map_edge (map_id, source_node_id, target_node_id, type, create_time, update_time)
SELECT @map_id, s.id, t.id, 'prerequisite', NOW(), NOW()
FROM learning_map_node s JOIN learning_map_node t
WHERE s.map_id=@map_id AND t.map_id=@map_id AND s.title='字符串' AND t.title='哈希';

INSERT INTO learning_map_edge (map_id, source_node_id, target_node_id, type, create_time, update_time)
SELECT @map_id, s.id, t.id, 'prerequisite', NOW(), NOW()
FROM learning_map_node s JOIN learning_map_node t
WHERE s.map_id=@map_id AND t.map_id=@map_id AND s.title='字符串' AND t.title='KMP';

INSERT INTO learning_map_edge (map_id, source_node_id, target_node_id, type, create_time, update_time)
SELECT @map_id, s.id, t.id, 'prerequisite', NOW(), NOW()
FROM learning_map_node s JOIN learning_map_node t
WHERE s.map_id=@map_id AND t.map_id=@map_id AND s.title='字符串' AND t.title='Trie';

INSERT INTO learning_map_edge (map_id, source_node_id, target_node_id, type, create_time, update_time)
SELECT @map_id, s.id, t.id, 'prerequisite', NOW(), NOW()
FROM learning_map_node s JOIN learning_map_node t
WHERE s.map_id=@map_id AND t.map_id=@map_id AND s.title='搜索' AND t.title='DFS';

INSERT INTO learning_map_edge (map_id, source_node_id, target_node_id, type, create_time, update_time)
SELECT @map_id, s.id, t.id, 'prerequisite', NOW(), NOW()
FROM learning_map_node s JOIN learning_map_node t
WHERE s.map_id=@map_id AND t.map_id=@map_id AND s.title='搜索' AND t.title='BFS';

INSERT INTO learning_map_edge (map_id, source_node_id, target_node_id, type, create_time, update_time)
SELECT @map_id, s.id, t.id, 'prerequisite', NOW(), NOW()
FROM learning_map_node s JOIN learning_map_node t
WHERE s.map_id=@map_id AND t.map_id=@map_id AND s.title='动态规划' AND t.title='背包 DP';

-- 知识点 -> 题目
INSERT INTO learning_map_edge (map_id, source_node_id, target_node_id, type, create_time, update_time)
SELECT @map_id, s.id, t.id, 'prerequisite', NOW(), NOW()
FROM learning_map_node s JOIN learning_map_node t
WHERE s.map_id=@map_id AND t.map_id=@map_id AND s.title='单调栈' AND t.title='单调栈基础题 1';

INSERT INTO learning_map_edge (map_id, source_node_id, target_node_id, type, create_time, update_time)
SELECT @map_id, s.id, t.id, 'prerequisite', NOW(), NOW()
FROM learning_map_node s JOIN learning_map_node t
WHERE s.map_id=@map_id AND t.map_id=@map_id AND s.title='单调栈基础题 1' AND t.title='单调栈基础题 2';

INSERT INTO learning_map_edge (map_id, source_node_id, target_node_id, type, create_time, update_time)
SELECT @map_id, s.id, t.id, 'prerequisite', NOW(), NOW()
FROM learning_map_node s JOIN learning_map_node t
WHERE s.map_id=@map_id AND t.map_id=@map_id AND s.title='单调栈基础题 2' AND t.title='单调栈进阶题 1';

INSERT INTO learning_map_edge (map_id, source_node_id, target_node_id, type, create_time, update_time)
SELECT @map_id, s.id, t.id, 'prerequisite', NOW(), NOW()
FROM learning_map_node s JOIN learning_map_node t
WHERE s.map_id=@map_id AND t.map_id=@map_id AND s.title='单调队列' AND t.title='单调队列基础题 1';

INSERT INTO learning_map_edge (map_id, source_node_id, target_node_id, type, create_time, update_time)
SELECT @map_id, s.id, t.id, 'prerequisite', NOW(), NOW()
FROM learning_map_node s JOIN learning_map_node t
WHERE s.map_id=@map_id AND t.map_id=@map_id AND s.title='单调队列基础题 1' AND t.title='单调队列进阶题 1';

INSERT INTO learning_map_edge (map_id, source_node_id, target_node_id, type, create_time, update_time)
SELECT @map_id, s.id, t.id, 'prerequisite', NOW(), NOW()
FROM learning_map_node s JOIN learning_map_node t
WHERE s.map_id=@map_id AND t.map_id=@map_id AND s.title='并查集' AND t.title='并查集基础题 1';

INSERT INTO learning_map_edge (map_id, source_node_id, target_node_id, type, create_time, update_time)
SELECT @map_id, s.id, t.id, 'prerequisite', NOW(), NOW()
FROM learning_map_node s JOIN learning_map_node t
WHERE s.map_id=@map_id AND t.map_id=@map_id AND s.title='并查集基础题 1' AND t.title='并查集进阶题 1';

INSERT INTO learning_map_edge (map_id, source_node_id, target_node_id, type, create_time, update_time)
SELECT @map_id, s.id, t.id, 'prerequisite', NOW(), NOW()
FROM learning_map_node s JOIN learning_map_node t
WHERE s.map_id=@map_id AND t.map_id=@map_id AND s.title='树状数组' AND t.title='树状数组基础题 1';

INSERT INTO learning_map_edge (map_id, source_node_id, target_node_id, type, create_time, update_time)
SELECT @map_id, s.id, t.id, 'prerequisite', NOW(), NOW()
FROM learning_map_node s JOIN learning_map_node t
WHERE s.map_id=@map_id AND t.map_id=@map_id AND s.title='树状数组基础题 1' AND t.title='树状数组进阶题 1';

INSERT INTO learning_map_edge (map_id, source_node_id, target_node_id, type, create_time, update_time)
SELECT @map_id, s.id, t.id, 'prerequisite', NOW(), NOW()
FROM learning_map_node s JOIN learning_map_node t
WHERE s.map_id=@map_id AND t.map_id=@map_id AND s.title='线段树' AND t.title='线段树基础题 1';

INSERT INTO learning_map_edge (map_id, source_node_id, target_node_id, type, create_time, update_time)
SELECT @map_id, s.id, t.id, 'prerequisite', NOW(), NOW()
FROM learning_map_node s JOIN learning_map_node t
WHERE s.map_id=@map_id AND t.map_id=@map_id AND s.title='线段树基础题 1' AND t.title='线段树进阶题 1';

INSERT INTO learning_map_edge (map_id, source_node_id, target_node_id, type, create_time, update_time)
SELECT @map_id, s.id, t.id, 'prerequisite', NOW(), NOW()
FROM learning_map_node s JOIN learning_map_node t
WHERE s.map_id=@map_id AND t.map_id=@map_id AND s.title='哈希' AND t.title='哈希基础题 1';

INSERT INTO learning_map_edge (map_id, source_node_id, target_node_id, type, create_time, update_time)
SELECT @map_id, s.id, t.id, 'prerequisite', NOW(), NOW()
FROM learning_map_node s JOIN learning_map_node t
WHERE s.map_id=@map_id AND t.map_id=@map_id AND s.title='KMP' AND t.title='KMP基础题 1';

INSERT INTO learning_map_edge (map_id, source_node_id, target_node_id, type, create_time, update_time)
SELECT @map_id, s.id, t.id, 'prerequisite', NOW(), NOW()
FROM learning_map_node s JOIN learning_map_node t
WHERE s.map_id=@map_id AND t.map_id=@map_id AND s.title='Trie' AND t.title='Trie基础题 1';

INSERT INTO learning_map_edge (map_id, source_node_id, target_node_id, type, create_time, update_time)
SELECT @map_id, s.id, t.id, 'prerequisite', NOW(), NOW()
FROM learning_map_node s JOIN learning_map_node t
WHERE s.map_id=@map_id AND t.map_id=@map_id AND s.title='字符串综合题 1' AND t.title='字符串进阶题 2';

INSERT INTO learning_map_edge (map_id, source_node_id, target_node_id, type, create_time, update_time)
SELECT @map_id, s.id, t.id, 'prerequisite', NOW(), NOW()
FROM learning_map_node s JOIN learning_map_node t
WHERE s.map_id=@map_id AND t.map_id=@map_id AND s.title='DFS' AND t.title='DFS基础题 1';

INSERT INTO learning_map_edge (map_id, source_node_id, target_node_id, type, create_time, update_time)
SELECT @map_id, s.id, t.id, 'prerequisite', NOW(), NOW()
FROM learning_map_node s JOIN learning_map_node t
WHERE s.map_id=@map_id AND t.map_id=@map_id AND s.title='BFS' AND t.title='BFS基础题 1';

INSERT INTO learning_map_edge (map_id, source_node_id, target_node_id, type, create_time, update_time)
SELECT @map_id, s.id, t.id, 'prerequisite', NOW(), NOW()
FROM learning_map_node s JOIN learning_map_node t
WHERE s.map_id=@map_id AND t.map_id=@map_id AND s.title='BFS基础题 1' AND t.title='BFS最短路入门题';

INSERT INTO learning_map_edge (map_id, source_node_id, target_node_id, type, create_time, update_time)
SELECT @map_id, s.id, t.id, 'prerequisite', NOW(), NOW()
FROM learning_map_node s JOIN learning_map_node t
WHERE s.map_id=@map_id AND t.map_id=@map_id AND s.title='二分搜索' AND t.title='二分搜索基础题 1';

INSERT INTO learning_map_edge (map_id, source_node_id, target_node_id, type, create_time, update_time)
SELECT @map_id, s.id, t.id, 'prerequisite', NOW(), NOW()
FROM learning_map_node s JOIN learning_map_node t
WHERE s.map_id=@map_id AND t.map_id=@map_id AND s.title='二分搜索基础题 1' AND t.title='二分搜索进阶题 1';

INSERT INTO learning_map_edge (map_id, source_node_id, target_node_id, type, create_time, update_time)
SELECT @map_id, s.id, t.id, 'prerequisite', NOW(), NOW()
FROM learning_map_node s JOIN learning_map_node t
WHERE s.map_id=@map_id AND t.map_id=@map_id AND s.title='动态规划' AND t.title='DP基础题 1';

INSERT INTO learning_map_edge (map_id, source_node_id, target_node_id, type, create_time, update_time)
SELECT @map_id, s.id, t.id, 'prerequisite', NOW(), NOW()
FROM learning_map_node s JOIN learning_map_node t
WHERE s.map_id=@map_id AND t.map_id=@map_id AND s.title='背包 DP' AND t.title='背包DP基础题 1';

INSERT INTO learning_map_edge (map_id, source_node_id, target_node_id, type, create_time, update_time)
SELECT @map_id, s.id, t.id, 'prerequisite', NOW(), NOW()
FROM learning_map_node s JOIN learning_map_node t
WHERE s.map_id=@map_id AND t.map_id=@map_id AND s.title='背包DP基础题 1' AND t.title='背包DP进阶题 1';

INSERT INTO learning_map_edge (map_id, source_node_id, target_node_id, type, create_time, update_time)
SELECT @map_id, s.id, t.id, 'prerequisite', NOW(), NOW()
FROM learning_map_node s JOIN learning_map_node t
WHERE s.map_id=@map_id AND t.map_id=@map_id AND s.title='数组' AND t.title='数组基础题 1';

INSERT INTO learning_map_edge (map_id, source_node_id, target_node_id, type, create_time, update_time)
SELECT @map_id, s.id, t.id, 'prerequisite', NOW(), NOW()
FROM learning_map_node s JOIN learning_map_node t
WHERE s.map_id=@map_id AND t.map_id=@map_id AND s.title='链表' AND t.title='链表基础题 1';

INSERT INTO learning_map_edge (map_id, source_node_id, target_node_id, type, create_time, update_time)
SELECT @map_id, s.id, t.id, 'prerequisite', NOW(), NOW()
FROM learning_map_node s JOIN learning_map_node t
WHERE s.map_id=@map_id AND t.map_id=@map_id AND s.title='栈' AND t.title='栈基础题 1';

INSERT INTO learning_map_edge (map_id, source_node_id, target_node_id, type, create_time, update_time)
SELECT @map_id, s.id, t.id, 'prerequisite', NOW(), NOW()
FROM learning_map_node s JOIN learning_map_node t
WHERE s.map_id=@map_id AND t.map_id=@map_id AND s.title='队列' AND t.title='队列基础题 1';

INSERT INTO learning_map_edge (map_id, source_node_id, target_node_id, type, create_time, update_time)
SELECT @map_id, s.id, t.id, 'prerequisite', NOW(), NOW()
FROM learning_map_node s JOIN learning_map_node t
WHERE s.map_id=@map_id AND t.map_id=@map_id AND s.title='搜索' AND t.title='搜索综合题 1';

INSERT INTO learning_map_edge (map_id, source_node_id, target_node_id, type, create_time, update_time)
SELECT @map_id, s.id, t.id, 'prerequisite', NOW(), NOW()
FROM learning_map_node s JOIN learning_map_node t
WHERE s.map_id=@map_id AND t.map_id=@map_id AND s.title='线段树进阶题 1' AND t.title='数据结构综合题 1';

-- ========== 关联连线（related） ==========
INSERT INTO learning_map_edge (map_id, source_node_id, target_node_id, type, create_time, update_time)
SELECT @map_id, s.id, t.id, 'related', NOW(), NOW()
FROM learning_map_node s JOIN learning_map_node t
WHERE s.map_id=@map_id AND t.map_id=@map_id AND s.title='BFS最短路入门题' AND t.title='二分搜索进阶题 1';

INSERT INTO learning_map_edge (map_id, source_node_id, target_node_id, type, create_time, update_time)
SELECT @map_id, s.id, t.id, 'related', NOW(), NOW()
FROM learning_map_node s JOIN learning_map_node t
WHERE s.map_id=@map_id AND t.map_id=@map_id AND s.title='字符串进阶题 2' AND t.title='搜索综合题 1';
