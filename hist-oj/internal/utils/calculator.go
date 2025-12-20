package utils

import (
	"fmt"
	"math"
)

// CalculateRatingChange 计算rating变化
// oldRating: 当前rating
// rank: 排名（从1开始）
// participants: 参赛人数
// allRatings: 所有参赛者的rating列表（按排名顺序）
// kFactor: K值（调整系数）
// hasSolved: 是否至少AC了一道题（暂时保留参数，但不再使用）
func CalculateRatingChange(oldRating, rank, participants int, allRatings []int, kFactor int, hasSolved bool) int {
	if participants < 2 {
		return 0
	}

	// 检查是否所有人的rating都相同（或非常接近）
	allSameRating := true
	firstRating := allRatings[0]
	for _, r := range allRatings {
		if math.Abs(float64(r-firstRating)) > 10 { // 允许10分的误差
			allSameRating = false
			break
		}
	}

	// 如果所有人rating相同，使用简化的线性分配
	if allSameRating {
		// 使用线性插值：第1名获得最高分，最后一名获得最低分
		// 更激进的幅度：第1名+150分，最后一名-90分
		maxChange := float64(kFactor) * 3.0  // 第1名最多加 K×3.0 分
		minChange := -float64(kFactor) * 1.8 // 最后一名最多减 K×1.8 分

		// 线性插值计算rating变化
		// rank=1 时得到 maxChange，rank=participants 时得到 minChange
		ratingChange := maxChange - (maxChange-minChange)*float64(rank-1)/float64(participants-1)

		// 调试日志
		if rank <= 3 || rank >= participants-2 {
			println(fmt.Sprintf("🔵 线性分配: rank=%d, maxChange=%.2f, minChange=%.2f, ratingChange=%.2f",
				rank, maxChange, minChange, ratingChange))
		}

		return int(math.Round(ratingChange))
	}

	// 如果rating有差异，使用标准Elo算法
	// 计算期望排名（基于当前rating在所有参赛者中的位置）
	expectedRank := calculateExpectedRank(oldRating, allRatings)

	// 实际排名得分（排名越靠前，得分越高）
	actualScore := float64(participants - rank + 1)

	// 期望得分
	expectedScore := float64(participants) - expectedRank + 1

	// Rating变化（标准Elo算法）
	// 使用更大的系数使变化更明显
	ratingChange := float64(kFactor) * 3.0 * (actualScore - expectedScore) / float64(participants)

	return int(math.Round(ratingChange))
}

// calculateExpectedRank 计算期望排名
// 基于Elo算法：
// 1. 计算期望战胜的人数 = sum(1 / (1 + 10^((对方rating - 我的rating) / 400)))
// 2. 期望排名 = 总人数 - 期望战胜的人数 + 0.5（排除自己）
// 注意：allRatings 包含用户自己的rating
func calculateExpectedRank(rating int, allRatings []int) float64 {
	// 计算期望战胜的人数
	expectedWins := 0.0

	for _, otherRating := range allRatings {
		// 计算战胜对方的概率
		probability := 1.0 / (1.0 + math.Pow(10.0, float64(otherRating-rating)/400.0))
		expectedWins += probability
	}

	// 期望排名 = 总人数 - 期望战胜的人数 + 0.5
	// 0.5 是因为 allRatings 包含自己，战胜自己的概率是 0.5
	return float64(len(allRatings)) - expectedWins + 0.5
}

// Rating下限
const MinRating = 800

// CalculateNewRating 计算新rating
func CalculateNewRating(oldRating, ratingChange int) int {
	newRating := oldRating + ratingChange
	if newRating < MinRating {
		return MinRating
	}
	return newRating
}

// UserRatingInfo 用户rating信息（用于强关联绑定）
type UserRatingInfo struct {
	UID       string // 用户唯一标识
	Rank      int    // 排名（从1开始）
	Rating    int    // 当前rating
	HasSolved bool   // 是否至少AC了一道题
}

// CalculateAllRatingChangesByUID 通过UID批量计算rating变化（强关联绑定）
// 返回 map[UID]ratingChange，确保每个用户都通过UID关联到正确的rating变化
func CalculateAllRatingChangesByUID(userInfos []UserRatingInfo, kFactor int) map[string]int {
	participants := len(userInfos)
	if participants == 0 {
		return make(map[string]int)
	}

	// 构建所有rating列表（用于计算期望排名）
	allRatings := make([]int, participants)
	for i, info := range userInfos {
		allRatings[i] = info.Rating
	}

	// 计算每个用户的rating变化，通过UID进行强关联
	changes := make(map[string]int, participants)
	for _, info := range userInfos {
		change := CalculateRatingChange(info.Rating, info.Rank, participants, allRatings, kFactor, info.HasSolved)
		changes[info.UID] = change
	}

	return changes
}

// CalculateAllRatingChanges 批量计算所有参赛者的rating变化（保留旧接口以兼容）
// 注意：建议使用 CalculateAllRatingChangesByUID 进行强关联绑定
func CalculateAllRatingChanges(ratings []int, kFactor int) []int {
	participants := len(ratings)
	changes := make([]int, participants)

	for i := 0; i < participants; i++ {
		rank := i + 1
		oldRating := ratings[i]
		// 旧接口默认认为所有人都AC了题目
		change := CalculateRatingChange(oldRating, rank, participants, ratings, kFactor, true)
		changes[i] = change
	}

	return changes
}

