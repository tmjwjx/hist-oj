# 手动调整 Rating 功能使用说明

## 📋 功能概述

本功能允许管理员手动调整用户的 Rating 值，用于处理作弊、违规等情况。

**核心特性**:
- ✅ 直接调整用户当前 Rating（影响 `user_record.hist_rating`）
- ✅ 自动记录到 Rating 历史（`rating_history` 表）
- ✅ 记录操作原因和操作人（便于审计）
- ✅ 前端图表用特殊颜色标注（红色警告图标）
- ✅ 下场比赛自动使用新 Rating 计算

---

## 🚀 快速开始

### 1. 执行数据库迁移

```bash
cd hist-oj
mysql -u root -p hoj < migrations/005_add_manual_rating_fields.sql
```

### 2. 编译并运行 hist-oj 服务

```bash
cd hist-oj
make build
make run
```

### 3. 使用 API 调整用户 Rating

#### **请求示例**

```bash
curl -X POST "http://localhost:9527/api/rating/admin/adjust" \
  -H "Content-Type: application/json" \
  -H "X-Operator-UID: admin" \
  -d '{
    "username": "test_user",
    "ratingChange": -100,
    "reason": "AI作弊"
  }'
```

#### **请求参数说明**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `username` | string | ✅ | 用户名 |
| `ratingChange` | int | ✅ | Rating 变化值（正数=增加，负数=减少，范围：-500 ~ +500） |
| `reason` | string | ✅ | 操作原因（如"AI作弊"、"账号违规"等） |

#### **请求头说明**

| Header | 说明 |
|--------|------|
| `X-Operator-UID` | 操作人 UID（用于审计，当前为简化实现） |

#### **响应示例**

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "username": "test_user",
    "oldRating": 1660,
    "newRating": 1560,
    "ratingChange": -100,
    "reason": "AI作弊",
    "operatorUID": "admin"
  }
}
```

---

## 📊 前端显示效果

### **Rating 历史图表**

- **比赛记录**: 蓝色点 (#409EFF)
- **手动调整**: 红色点 + ⚠️ 图标 (#FF4D4F)

**Tooltip 显示**:
```
⚠️ AI作弊
Rating: 1660 → 1560
变化: -100
2024-01-05 15:30:00
```

---

## 🧪 测试脚本

### **使用测试脚本**

```bash
cd hist-oj

# 基本用法：扣分 100
./test_manual_adjust.sh test_user -100 "测试扣分"

# 增加分数
./test_manual_adjust.sh test_user 50 "奖励贡献"

# 查看帮助
./test_manual_adjust.sh
```

### **手动测试（使用 curl）**

```bash
# 1. 扣分测试
curl -X POST "http://localhost:9527/api/rating/admin/adjust" \
  -H "Content-Type: application/json" \
  -H "X-Operator-UID: admin" \
  -d '{"username": "test_user", "ratingChange": -200, "reason": "比赛违规"}'

# 2. 奖励测试
curl -X POST "http://localhost:9527/api/rating/admin/adjust" \
  -H "Content-Type: application/json" \
  -H "X-Operator-UID: admin" \
  -d '{"username": "test_user", "ratingChange": 100, "reason": "发现Bug奖励"}'

# 3. 查看 Rating 历史
curl "http://localhost:9527/api/rating/history/test_user"

# 4. 查看当前 Rating
curl "http://localhost:9527/api/rating/user/test_user"
```

---

## 📝 数据库变更

### **rating_history 表新增字段**

| 字段 | 类型 | 说明 |
|------|------|------|
| `reason` | VARCHAR(255) | 操作原因（手动调整时的备注） |
| `is_manual` | TINYINT(1) | 是否为手动调整（0=比赛计算，1=手动调整） |
| `operator_uid` | VARCHAR(32) | 操作人 UID（手动调整时记录） |
| `contest_id` | BIGINT UNSIGNED | 改为可为 NULL（手动调整时为 NULL） |

### **查询示例**

```sql
-- 查看所有手动调整记录
SELECT * FROM rating_history WHERE is_manual = 1;

-- 查看某个用户的所有手动调整记录
SELECT * FROM rating_history
WHERE uid = 'user_uuid' AND is_manual = 1
ORDER BY created_at DESC;

-- 查看某个操作人的所有操作记录
SELECT * FROM rating_history
WHERE operator_uid = 'admin_uuid' AND is_manual = 1
ORDER BY created_at DESC;
```

---

## 🔐 权限控制

### **当前实现（简化版）**

- ⚠️ **未实现真正的管理员验证**
- 操作人 UID 从请求头 `X-Operator-UID` 获取
- 默认值为 `admin`

### **后续改进（TODO）**

需要集成 HOJ 后端的认证系统：
1. 从 JWT Token 中解析用户信息
2. 调用 HOJ API 验证用户是否为管理员
3. 记录真实的操作人信息

---

## ⚙️ 配置说明

### **Rating 调整限制**

在 `internal/service/rating.go` 中定义：

```go
// 限制单次调整范围（-500 ~ +500）
if delta < -500 || delta > 500 {
    return 0, 0, 0, fmt.Errorf("rating变化值必须在-500到500之间")
}
```

### **Rating 下限**

在 `internal/utils/calculator.go` 中定义：

```go
// Rating 下限
const MinRating = 800
```

---

## 📂 文件清单

### **后端文件**

| 文件 | 说明 |
|------|------|
| `migrations/005_add_manual_rating_fields.sql` | 数据库迁移脚本 |
| `internal/model/rating.go` | RatingHistory 模型扩展 |
| `internal/service/rating.go` | AdjustUserRating 方法实现 |
| `internal/api/handler.go` | AdjustUserRating API 处理器 |
| `internal/api/routes.go` | 路由配置 |
| `internal/api/middleware.go` | 管理员权限验证中间件 |

### **前端文件**

| 文件 | 说明 |
|------|------|
| `src/components/oj/user/RatingChart.vue` | Rating 图表组件（支持显示手动调整） |

### **测试文件**

| 文件 | 说明 |
|------|------|
| `test_manual_adjust.sh` | 手动调整功能测试脚本 |

---

## 🎯 使用场景

### **场景 1: 处理作弊**

```json
{
  "username": "cheater_user",
  "ratingChange": -200,
  "reason": "比赛中使用AI作弊"
}
```

**结果**:
- 用户 Rating 扣除 200 分
- 历史图表显示红色警告标注
- 原因显示"比赛中使用AI作弊"

### **场景 2: 奖励贡献**

```json
{
  "username": "contributor",
  "ratingChange": 50,
  "reason": "发现系统漏洞并报告"
}
```

**结果**:
- 用户 Rating 增加 50 分
- 历史图表显示绿色正向标注

---

## ❓ 常见问题

### **Q1: 手动调整会影响历史比赛记录吗？**

**A**: 不会。手动调整只修改 `user_record.hist_rating`（当前 Rating），不会修改历史比赛的 rating 记录。

### **Q2: 下场比赛会使用新 Rating 吗？**

**A**: 会。比赛计算时从 `user_record.hist_rating` 读取当前 Rating，手动调整后立即生效。

### **Q3: 可以撤销手动调整吗？**

**A**: 可以。通过反向调整实现，例如之前扣了 100 分，现在可以通过 `ratingChange: 100` 加回来。

### **Q4: Rating 会低于 0 吗？**

**A**: 不会。Rating 有下限保护（当前设置为 800 分），在 `utils.CalculateNewRating` 中实现。

---

## 🔄 工作流程

```
管理员请求
    ↓
API: POST /api/rating/admin/adjust
    ↓
权限验证（中间件）
    ↓
参数验证（用户名、变化值、原因）
    ↓
数据库事务开始
    ↓
查询用户当前 Rating
    ↓
计算新 Rating
    ↓
更新 user_record.hist_rating
    ↓
插入 rating_history 记录（is_manual=true）
    ↓
事务提交
    ↓
返回结果
    ↓
前端显示（红色标注）
```

---

## 📞 联系方式

如有问题，请联系开发团队或提交 Issue。
