package utils

import (
	"math"
)

// CalculateRatingChange 计算rating变化
// oldRating: 当前rating
// rank: 排名（从1开始）
// participants: 参赛人数
// allRatings: 所有参赛者的rating列表（按排名顺序）
// kFactor: K值（调整系数）
func CalculateRatingChange(oldRating, rank, participants int, allRatings []int, kFactor int) int {
	if participants < 2 {
		return 0
	}

	// 计算期望排名（基于当前rating在所有参赛者中的位置）
	expectedRank := calculateExpectedRank(oldRating, allRatings)
	
	// 实际排名得分（排名越靠前，得分越高）
	actualScore := float64(participants - rank + 1)
	
	// 期望得分
	expectedScore := float64(participants) - expectedRank + 1
	
	// 计算rating变化
	ratingChange := int(math.Round(float64(kFactor) * (actualScore - expectedScore) / float64(participants)))
	
	return ratingChange
}

// calculateExpectedRank 计算期望排名
// 基于Elo算法：期望排名 = 1 + sum(1 / (1 + 10^((rating - otherRating) / 400)))
func calculateExpectedRank(rating int, allRatings []int) float64 {
	expectedRank := 1.0
	
	for _, otherRating := range allRatings {
		if otherRating != rating {
			probability := 1.0 / (1.0 + math.Pow(10.0, float64(otherRating-rating)/400.0))
			expectedRank += probability
		}
	}
	
	return expectedRank
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

// CalculateAllRatingChanges 批量计算所有参赛者的rating变化
func CalculateAllRatingChanges(ratings []int, kFactor int) []int {
	participants := len(ratings)
	changes := make([]int, participants)
	
	for i := 0; i < participants; i++ {
		rank := i + 1
		oldRating := ratings[i]
		change := CalculateRatingChange(oldRating, rank, participants, ratings, kFactor)
		changes[i] = change
	}
	
	return changes
}

