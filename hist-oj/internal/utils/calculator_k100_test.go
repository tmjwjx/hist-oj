package utils

import (
	"fmt"
	"testing"
)

// TestCalculateRatingChange_28Users_K100 测试28个1200分用户，K=100的情况
func TestCalculateRatingChange_28Users_K100(t *testing.T) {
	participants := 28
	kFactor := 100
	initialRating := 1200

	// 创建28个初始Rating都是1200的用户
	allRatings := make([]int, participants)
	for i := 0; i < participants; i++ {
		allRatings[i] = initialRating
	}

	fmt.Println("========================================")
	fmt.Println("28人比赛 Rating 计算结果 (K=100)")
	fmt.Println("所有人初始 Rating: 1200")
	fmt.Println("========================================")
	fmt.Printf("%-6s | %-12s | %-12s | %-12s\n", "排名", "初始Rating", "Rating变化", "最终Rating")
	fmt.Println("----------------------------------------")

	// 计算每个排名的Rating变化
	for rank := 1; rank <= participants; rank++ {
		change := CalculateRatingChange(initialRating, rank, participants, allRatings, kFactor, true)
		newRating := CalculateNewRating(initialRating, change)

		fmt.Printf("%-6d | %-12d | %+12d | %-12d\n", rank, initialRating, change, newRating)
	}

	fmt.Println("========================================")

	// 验证第1名和最后一名的变化
	firstPlaceChange := CalculateRatingChange(initialRating, 1, participants, allRatings, kFactor, true)
	lastPlaceChange := CalculateRatingChange(initialRating, participants, participants, allRatings, kFactor, true)

	fmt.Printf("\n关键数据验证:\n")
	fmt.Printf("第1名变化: %+d (理论最大值: K×3.0 = %d)\n", firstPlaceChange, kFactor*3)
	fmt.Printf("最后一名变化: %+d (理论最小值: -K×1.8 = %d)\n", lastPlaceChange, -int(float64(kFactor)*1.8))
	fmt.Printf("总变化和: %+d (应该接近0)\n", calculateTotalChange(allRatings, kFactor))

	// 基本断言
	if firstPlaceChange <= 0 {
		t.Errorf("第1名应该加分，但得到: %d", firstPlaceChange)
	}

	if lastPlaceChange >= 0 {
		t.Errorf("最后一名应该减分，但得到: %d", lastPlaceChange)
	}

	// 验证第1名的变化在合理范围内 (应该接近 K×3.0)
	expectedMax := kFactor * 3
	if firstPlaceChange > expectedMax {
		t.Errorf("第1名变化 %d 超过了理论最大值 %d", firstPlaceChange, expectedMax)
	}

	// 验证最后一名的变化在合理范围内 (应该接近 -K×1.8)
	expectedMin := -int(float64(kFactor) * 1.8)
	if lastPlaceChange < expectedMin {
		t.Errorf("最后一名变化 %d 超过了理论最小值 %d", lastPlaceChange, expectedMin)
	}
}

// calculateTotalChange 计算所有人的总变化（用于验证零和性质）
func calculateTotalChange(allRatings []int, kFactor int) int {
	participants := len(allRatings)
	totalChange := 0

	for rank := 1; rank <= participants; rank++ {
		change := CalculateRatingChange(allRatings[0], rank, participants, allRatings, kFactor, true)
		totalChange += change
	}

	return totalChange
}

// TestCalculateRatingChange_29Users_K100 测试29个1200分用户，K=100的情况（对比）
func TestCalculateRatingChange_29Users_K100(t *testing.T) {
	participants := 29
	kFactor := 100
	initialRating := 1200

	// 创建29个初始Rating都是1200的用户
	allRatings := make([]int, participants)
	for i := 0; i < participants; i++ {
		allRatings[i] = initialRating
	}

	fmt.Println("\n========================================")
	fmt.Println("29人比赛 Rating 计算结果 (K=100)")
	fmt.Println("所有人初始 Rating: 1200")
	fmt.Println("========================================")
	fmt.Printf("%-6s | %-12s | %-12s | %-12s\n", "排名", "初始Rating", "Rating变化", "最终Rating")
	fmt.Println("----------------------------------------")

	// 计算每个排名的Rating变化
	for rank := 1; rank <= participants; rank++ {
		change := CalculateRatingChange(initialRating, rank, participants, allRatings, kFactor, true)
		newRating := CalculateNewRating(initialRating, change)

		fmt.Printf("%-6d | %-12d | %+12d | %-12d\n", rank, initialRating, change, newRating)
	}

	fmt.Println("========================================")

	// 验证第1名和最后一名的变化
	firstPlaceChange := CalculateRatingChange(initialRating, 1, participants, allRatings, kFactor, true)
	lastPlaceChange := CalculateRatingChange(initialRating, participants, participants, allRatings, kFactor, true)

	fmt.Printf("\n关键数据验证:\n")
	fmt.Printf("第1名变化: %+d (理论最大值: K×3.0 = %d)\n", firstPlaceChange, kFactor*3)
	fmt.Printf("最后一名变化: %+d (理论最小值: -K×1.8 = %d)\n", lastPlaceChange, -int(float64(kFactor)*1.8))
}
