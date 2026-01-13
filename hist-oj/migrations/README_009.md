# 数据库迁移说明 - 添加 show_answer 字段

## 迁移文件
- 文件路径: `migrations/009_add_show_answer.sql`
- 迁移版本: 009
- 创建时间: 2025-01-13

## 迁移目的
在 `classroom_homework` 表中添加 `show_answer` 字段，用于控制学生提交作业后是否可以查看答案。

## SQL 语句
```sql
ALTER TABLE `classroom_homework`
ADD COLUMN `show_answer` INT DEFAULT 0 COMMENT '学生提交后是否可以查看答案（0: 否, 1: 是）' AFTER `show_homework`;
```

## 字段说明
- **字段名**: `show_answer`
- **类型**: INT
- **默认值**: 0 (不显示答案)
- **位置**: 在 `show_homework` 字段之后
- **注释**: 学生提交后是否可以查看答案（0: 否, 1: 是）

## 功能说明
- 当 `show_answer = 0` 时：学生提交作业后不能查看答案
- 当 `show_answer = 1` 时：学生提交作业后可以查看每道题的正确答案
  - 单选题/多选题/判断题：显示正确答案
  - 主观题：显示参考答案（如果教师设置了），否则显示"教师未设置答案"
  - 编程题：显示"参考题库题解"链接，点击可跳转到 HOJ 题库

## 执行方式

### 方式1：使用 MySQL 客户端
```bash
mysql -h 43.143.133.62 -P 3306 -u root -phist2025 hoj < migrations/009_add_show_answer.sql
```

### 方式2：登录 MySQL 后执行
```bash
mysql -h 43.143.133.62 -P 3306 -u root -phist2025 hoj
```

然后在 MySQL 命令行中执行：
```sql
ALTER TABLE `classroom_homework`
ADD COLUMN `show_answer` INT DEFAULT 0 COMMENT '学生提交后是否可以查看答案（0: 否, 1: 是）' AFTER `show_homework`;
```

### 方式3：使用 phpMyAdmin 或其他数据库管理工具
1. 登录 phpMyAdmin
2. 选择数据库 `hoj`
3. 选择表 `classroom_homework`
4. 点击"结构"标签
5. 添加新字段：
   - 字段名: `show_answer`
   - 类型: INT
   - 长度/值: 留空
   - 默认值: 0
   - 注释: 学生提交后是否可以查看答案（0: 否, 1: 是）
   - 位置: 在 `show_homework` 之后

## 验证步骤

执行迁移后，运行以下 SQL 验证字段是否添加成功：

```sql
-- 查看表结构
SHOW COLUMNS FROM classroom_homework;

-- 应该能看到 show_answer 字段
-- 字段应该位于 show_homework 和 show_score 之间

-- 测试查询
SELECT id, title, show_homework, show_answer, show_score FROM classroom_homework LIMIT 5;
```

## 回滚方案

如果需要回滚此迁移，执行以下 SQL：

```sql
ALTER TABLE `classroom_homework` DROP COLUMN `show_answer`;
```

## 影响分析

### 对现有功能的影响
- **无破坏性影响**: 新增字段设置了默认值 0，不会影响现有的作业
- **向后兼容**: 现有的作业默认不显示答案（show_answer = 0）
- **可选功能**: 只有教师在创建/编辑作业时主动勾选"允许学生提交后查看答案"，才会启用此功能

### 需要更新的代码
- ✅ 后端模型: `internal/model/classroom.go` - 已添加 `ShowAnswer` 字段
- ✅ 前端创建作业: `hoj-vue/src/views/classroom/teacher/CreateHomework.vue` - 已添加复选框
- ✅ 前端学生查看: `hoj-vue/src/views/classroom/student/HomeworkDetail.vue` - 已添加答案显示逻辑
- ✅ 编程题组件: `hoj-vue/src/components/classroom/ProgrammingQuestion.vue` - 已添加题解链接

## 测试建议

1. **创建新作业**: 测试勾选"允许学生提交后查看答案"选项
2. **编辑现有作业**: 测试修改"允许查看答案"设置
3. **学生视角**: 测试提交前后答案的显示/隐藏
4. **各种题型**: 测试单选、多选、判断、主观题、编程题的答案显示

## 注意事项
- 此迁移只影响 `classroom_homework` 表，不影响其他表
- 迁移是幂等的，多次执行会报错（字段已存在），但不会造成数据损坏
- 建议在执行前备份数据库
