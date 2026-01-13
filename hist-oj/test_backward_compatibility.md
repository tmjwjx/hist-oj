# 作业功能向后兼容性测试说明

## 测试目的
确保新增的 `problem_id` 字段不会影响现有作业功能的正常运行。

## 数据库变更
1. `homework_question` 表新增 `problem_id` 字段（VARCHAR(50)，可为 NULL）
2. `homework_question` 表的 `question_id` 字段改为可为 NULL
3. `homework_submit` 表新增 `problem_id` 字段（VARCHAR(50)，可为 NULL）
4. `homework_submit` 表的 `question_id` 字段改为可为 NULL

## 向后兼容性保证

### 1. 数据模型层面（Go 后端）
- `QuestionID` 和 `ProblemID` 都定义为指针类型（`*uint64` 和 `*string`）
- 指针类型可以为 nil，因此旧数据（没有 problem_id）不会出错
- 新数据（有 problem_id）也能正常工作

### 2. API 验证层面
后端在 `CreateHomework` 和 `UpdateHomework` 中都有验证逻辑：
```go
// 验证至少提供一个ID
if q.QuestionID == nil && q.ProblemID == nil {
    logger.Error("题目必须提供questionId或problemId")
    c.JSON(http.StatusOK, errorResponse(400, "题目必须提供questionId或problemId"))
    return
}
```

这确保了：
- 旧前端只传 `questionId`，不传 `problemId`（或者传 null）→ 正常工作
- 新前端传 `problemId`，不传 `questionId`（或者传 null）→ 正常工作
- 前端两个都不传 → 返回错误（符合预期）

### 3. 前端数据映射层面
前端在提交数据时会明确设置字段：
```javascript
questions: this.selectedQuestions.map((q, index) => {
  const result = {
    score: Number(q.score) || 10,
    questionOrder: index + 1
  }
  // 编程题使用 problemId,其他题型使用 questionId
  if (q.type === 'programming' && q.problemId) {
    result.problemId = q.problemId
    result.questionId = null  // 明确设置为 null
  } else {
    result.questionId = q.id
    result.problemId = null   // 明确设置为 null
  }
  return result
})
```

这确保了：
- 普通题目：`questionId` 有值，`problemId` 为 null
- 编程题：`problemId` 有值，`questionId` 为 null
- 不会出现两个字段都有值或都为空的情况

## 测试场景

### 场景 1：创建只包含普通题目的作业（旧功能）
**输入**：
```json
{
  "title": "测试作业",
  "questions": [
    {
      "questionId": 123,
      "problemId": null,
      "score": 10
    }
  ]
}
```

**预期结果**：
- ✅ 作业创建成功
- ✅ `homework_question` 表中 `question_id` = 123，`problem_id` = NULL

### 场景 2：创建包含编程题的作业（新功能）
**输入**：
```json
{
  "title": "测试作业2",
  "questions": [
    {
      "questionId": null,
      "problemId": "0001",
      "score": 20
    }
  ]
}
```

**预期结果**：
- ✅ 作业创建成功
- ✅ `homework_question` 表中 `question_id` = NULL，`problem_id` = "0001"

### 场景 3：创建混合类型作业
**输入**：
```json
{
  "title": "混合作业",
  "questions": [
    {
      "questionId": 123,
      "problemId": null,
      "score": 10
    },
    {
      "questionId": null,
      "problemId": "0001",
      "score": 20
    }
  ]
}
```

**预期结果**：
- ✅ 作业创建成功
- ✅ 第一题：`question_id` = 123，`problem_id` = NULL
- ✅ 第二题：`question_id` = NULL，`problem_id` = "0001"

### 场景 4：旧版本前端（不发送 problemId 字段）
**输入**：
```json
{
  "title": "旧版前端作业",
  "questions": [
    {
      "questionId": 123,
      "score": 10
      // 注意：没有 problemId 字段
    }
  ]
}
```

**预期结果**：
- ✅ 后端将 `problemId` 解析为 nil（因为字段是可选的）
- ✅ 验证通过（至少有 questionId）
- ✅ 作业创建成功

### 场景 5：错误情况（两个ID都不提供）
**输入**：
```json
{
  "title": "错误作业",
  "questions": [
    {
      "questionId": null,
      "problemId": null,
      "score": 10
    }
  ]
}
```

**预期结果**：
- ❌ 返回 400 错误："题目必须提供questionId或problemId"
- ✅ 作业不会被创建

## 数据库查询兼容性

### 查询普通题目
```sql
SELECT * FROM homework_question WHERE question_id IS NOT NULL;
```
- ✅ 可以正常查询到所有普通题目

### 查询编程题
```sql
SELECT * FROM homework_question WHERE problem_id IS NOT NULL;
```
- ✅ 可以正常查询到所有编程题

### 关联查询
```sql
SELECT
    hq.*,
    qb.title,
    qb.type
FROM homework_question hq
LEFT JOIN question_bank qb ON hq.question_id = qb.id
WHERE hq.homework_id = 1;
```
- ✅ 普通题目会关联到 `question_bank` 表
- ✅ 编程题的 `qb.*` 字段为 NULL（因为 LEFT JOIN）

## 结论

✅ **向后兼容性完全保证**：
1. 旧数据（只有 question_id）继续正常工作
2. 新数据（有 problem_id）也能正常工作
3. 混合场景也能正常处理
4. 后端验证确保数据完整性
5. 数据库查询不会出现问题

⚠️ **注意事项**：
- 所有使用 `homework_question` 表的查询代码需要考虑 `question_id` 和 `problem_id` 都可能为 NULL
- 前端展示作业详情时需要区分普通题目和编程题
