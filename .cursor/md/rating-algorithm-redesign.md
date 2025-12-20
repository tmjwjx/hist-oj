# Rating 算法重构设计方案

## 问题分析

### 当前问题
1. **初始 Rating 相同时计算不合理**：所有新人都是 1200 分，但排名第 8 的人加分和排名第 4 的一样多
2. **算法过于复杂**：包含 Elo 期望排名、排名加成、参赛人数因子等多个因素，容易出错
3. **不符合直觉**：用户无法理解为什么排名靠后反而加分更多

### 核心需求
1. **排名越靠前，加分越多**（单调递减）
2. **简单可靠**：算法逻辑清晰，不易出错
3. **公平合理**：考虑参赛人数和 Rating 差异

## 设计方案

### 方案：简化的 Elo Rating 系统

#### 核心思想
- 使用标准 Elo 算法计算期望排名
- 根据实际排名和期望排名的差异计算分数变化
- **去掉所有复杂的加成和惩罚**，只保留核心逻辑

#### 算法步骤

**1. 计算期望排名**
```
对于每个参赛者 A：
  期望排名 = 1 + Σ(战胜其他人的概率)

  战胜 B 的概率 = 1 / (1 + 10^((RatingB - RatingA) / 400))
```

**2. 计算 Rating 变化**
```
Rating 变化 = K × (实际排名得分 - 期望排名得分) / 参赛人数

其中：
- 实际排名得分 = 参赛人数 - 实际排名 + 1
- 期望排名得分 = 参赛人数 - 期望排名 + 1
- K = 50（固定值）
```

#### 特殊情况处理

**情况 1：所有人 Rating 相同（误差 ≤ 10 分）**

使用线性分配，确保单调性：

```
maxChange = K × 0.8 = 40 分（第 1 名）
minChange = -K × 0.5 = -25 分（最后 1 名）

Rating 变化 = maxChange - (maxChange - minChange) × (排名 - 1) / (参赛人数 - 1)
```

**示例**（29 人比赛）：
- 第 1 名：+40 分
- 第 2 名：+38 分
- 第 3 名：+35 分
- 第 4 名：+33 分
- 第 5 名：+31 分
- 第 8 名：+26 分
- 第 15 名：+7 分
- 第 29 名：-25 分

**情况 2：Rating 有差异**

使用标准 Elo 算法，但**不添加任何额外加成**。

#### 算法优点

1. **简单可靠**：
   - 只有两种情况：Rating 相同用线性分配，Rating 不同用标准 Elo
   - 没有复杂的分段加成逻辑

2. **单调性保证**：
   - Rating 相同时：线性分配天然保证单调性
   - Rating 不同时：Elo 算法在合理范围内也基本保证单调性

3. **符合直觉**：
   - 排名越靠前，加分越多
   - 打败高分选手加分更多，输给低分选手扣分更多

4. **易于调整**：
   - 只需要调整 K 值和线性分配的系数（0.8 和 0.5）

## 实现要点

### 代码结构

```go
func CalculateRatingChange(oldRating, rank, participants int, allRatings []int, kFactor int) int {
    // 1. 检查是否所有人 Rating 相同
    if allRatingsSame(allRatings) {
        return linearAllocation(rank, participants, kFactor)
    }

    // 2. 使用标准 Elo 算法
    return standardElo(oldRating, rank, participants, allRatings, kFactor)
}
```

### 关键函数

**1. allRatingsSame**
```go
func allRatingsSame(ratings []int) bool {
    if len(ratings) == 0 {
        return true
    }
    first := ratings[0]
    for _, r := range ratings {
        if math.Abs(float64(r - first)) > 10 {
            return false
        }
    }
    return true
}
```

**2. linearAllocation**
```go
func linearAllocation(rank, participants, kFactor int) int {
    maxChange := float64(kFactor) * 0.8
    minChange := -float64(kFactor) * 0.5

    change := maxChange - (maxChange - minChange) * float64(rank - 1) / float64(participants - 1)
    return int(math.Round(change))
}
```

**3. standardElo**
```go
func standardElo(oldRating, rank, participants int, allRatings []int, kFactor int) int {
    // 计算期望排名
    expectedRank := calculateExpectedRank(oldRating, allRatings)

    // 实际得分和期望得分
    actualScore := float64(participants - rank + 1)
    expectedScore := float64(participants) - expectedRank + 1

    // Rating 变化
    change := float64(kFactor) * (actualScore - expectedScore) / float64(participants)

    return int(math.Round(change))
}
```

## 测试验证

### 测试用例 1：所有人 1200 分（29 人）

| 排名 | 期望变化 | 说明 |
|------|---------|------|
| 1 | +40 | 第一名最高加分 |
| 2 | +38 | 递减 |
| 4 | +33 | 递减 |
| 8 | +26 | 明显低于前 4 名 |
| 15 | +7 | 中间位置小幅加分 |
| 29 | -25 | 最后一名扣分 |

### 测试用例 2：Rating 有差异

假设：
- A: 1300 分，排名第 5
- B: 1200 分，排名第 3
- C: 1100 分，排名第 1

期望：
- C（低分夺冠）：大幅加分
- B（中等分数排名靠前）：中等加分
- A（高分但排名靠后）：小幅加分或扣分

## 风险和注意事项

1. **Rating 下限**：保持 800 分的下限
2. **新手保护**：前 3 场比赛掉分减半（保持现有逻辑）
3. **K 值调整**：可以根据实际情况调整 K=50 这个值
4. **线性分配系数**：0.8 和 0.5 可以根据需要调整

## 总结

这个方案的核心优势是**简单可靠**：
- 只有两种情况，逻辑清晰
- 没有复杂的分段加成
- 保证单调性，符合直觉
- 易于测试和验证
