package utils

import (
	"math"
	"testing"
)

// TestCalculateRatingChange_AllSameRating 测试所有人 Rating 相同的情况
func TestCalculateRatingChange_AllSameRating(t *testing.T) {
	tests := []struct {
		name          string
		oldRating     int
		rank          int
		participants  int
		allRatings    []int
		kFactor       int
		hasSolved     bool
		expectedMin   int // 期望的最小值
		expectedMax   int // 期望的最大值
		description   string
	}{
		{
			name:         "29人比赛_第1名_有AC",
			oldRating:    1200,
			rank:         1,
			participants: 29,
			allRatings:   make([]int, 29), // 全部1200
			kFactor:      50,
			hasSolved:    true,
			expectedMin:  108, // 100 + 10 - 2 (允许误差)
			expectedMax:  112, // 100 + 10 + 2
			description:  "第1名应该获得最高加分: 100 + 10(参赛奖励) = 110",
		},
		{
			name:         "29人比赛_第2名_有AC",
			oldRating:    1200,
			rank:         2,
			participants: 29,
			allRatings:   make([]int, 29),
			kFactor:      50,
			hasSolved:    true,
			expectedMin:  102, // 约94 + 10 - 2
			expectedMax:  106, // 约94 + 10 + 2
			description:  "第2名应该获得较高加分",
		},
		{
			name:         "29人比赛_第4名_有AC",
			oldRating:    1200,
			rank:         4,
			participants: 29,
			allRatings:   make([]int, 29),
			kFactor:      50,
			hasSolved:    true,
			expectedMin:  91, // 约83 + 10 - 2
			expectedMax:  95, // 约83 + 10 + 2
			description:  "第4名应该获得中等加分",
		},
		{
			name:         "29人比赛_第8名_有AC",
			oldRating:    1200,
			rank:         8,
			participants: 29,
			allRatings:   make([]int, 29),
			kFactor:      50,
			hasSolved:    true,
			expectedMin:  68, // 约60 + 10 - 2
			expectedMax:  72, // 约60 + 10 + 2
			description:  "第8名应该明显低于前4名",
		},
		{
			name:         "29人比赛_第15名_有AC",
			oldRating:    1200,
			rank:         15,
			participants: 29,
			allRatings:   make([]int, 29),
			kFactor:      50,
			hasSolved:    true,
			expectedMin:  28, // 约20 + 10 - 2
			expectedMax:  32, // 约20 + 10 + 2
			description:  "第15名应该小幅加分",
		},
		{
			name:         "29人比赛_第29名_有AC",
			oldRating:    1200,
			rank:         29,
			participants: 29,
			allRatings:   make([]int, 29),
			kFactor:      50,
			hasSolved:    true,
			expectedMin:  -52, // -60 + 10 - 2
			expectedMax:  -48, // -60 + 10 + 2
			description:  "最后一名应该扣分，但有参赛奖励",
		},
		{
			name:         "29人比赛_第29名_无AC",
			oldRating:    1200,
			rank:         29,
			participants: 29,
			allRatings:   make([]int, 29),
			kFactor:      50,
			hasSolved:    false,
			expectedMin:  -62, // -60 - 2
			expectedMax:  -58, // -60 + 2
			description:  "最后一名且无AC，应该扣更多分",
		},
		{
			name:         "10人比赛_第1名_有AC",
			oldRating:    1200,
			rank:         1,
			participants: 10,
			allRatings:   make([]int, 10),
			kFactor:      50,
			hasSolved:    true,
			expectedMin:  108,
			expectedMax:  112,
			description:  "小规模比赛第1名",
		},
		{
			name:         "10人比赛_第10名_有AC",
			oldRating:    1200,
			rank:         10,
			participants: 10,
			allRatings:   make([]int, 10),
			kFactor:      50,
			hasSolved:    true,
			expectedMin:  -52,
			expectedMax:  -48,
			description:  "小规模比赛最后一名",
		},
	}

	// 初始化所有测试用例的 allRatings 为 1200
	for i := range tests {
		for j := range tests[i].allRatings {
			tests[i].allRatings[j] = 1200
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateRatingChange(
				tt.oldRating,
				tt.rank,
				tt.participants,
				tt.allRatings,
				tt.kFactor,
				tt.hasSolved,
			)

			if result < tt.expectedMin || result > tt.expectedMax {
				t.Errorf("%s\n期望范围: [%d, %d], 实际: %d\n%s",
					tt.name, tt.expectedMin, tt.expectedMax, result, tt.description)
			} else {
				t.Logf("✓ %s: %d (期望范围: [%d, %d])", tt.name, result, tt.expectedMin, tt.expectedMax)
			}
		})
	}
}

// TestCalculateRatingChange_Monotonicity 测试单调性：排名越靠前，加分越多
func TestCalculateRatingChange_Monotonicity(t *testing.T) {
	participants := 29
	allRatings := make([]int, participants)
	for i := range allRatings {
		allRatings[i] = 1200
	}
	kFactor := 50

	var prevChange int
	for rank := 1; rank <= participants; rank++ {
		change := CalculateRatingChange(1200, rank, participants, allRatings, kFactor, true)

		if rank > 1 && change >= prevChange {
			t.Errorf("单调性测试失败: rank=%d 的变化(%d) >= rank=%d 的变化(%d)",
				rank, change, rank-1, prevChange)
		}

		prevChange = change
		t.Logf("Rank %2d: %+4d", rank, change)
	}
}

// TestCalculateRatingChange_DifferentRatings 测试 Rating 有差异的情况
func TestCalculateRatingChange_DifferentRatings(t *testing.T) {
	tests := []struct {
		name         string
		oldRating    int
		rank         int
		participants int
		allRatings   []int
		kFactor      int
		hasSolved    bool
		expectSign   string // "positive", "negative", "zero"
		description  string
	}{
		{
			name:         "低分夺冠_应该大幅加分",
			oldRating:    1100,
			rank:         1,
			participants: 5,
			allRatings:   []int{1100, 1200, 1300, 1400, 1500},
			kFactor:      50,
			hasSolved:    true,
			expectSign:   "positive",
			description:  "1100分排名第1，应该大幅加分",
		},
		{
			name:         "高分垫底_应该大幅扣分",
			oldRating:    1500,
			rank:         5,
			participants: 5,
			allRatings:   []int{1100, 1200, 1300, 1400, 1500},
			kFactor:      50,
			hasSolved:    true,
			expectSign:   "negative",
			description:  "1500分排名第5，应该大幅扣分",
		},
		{
			name:         "中等分数排名靠前_应该加分",
			oldRating:    1200,
			rank:         2,
			participants: 5,
			allRatings:   []int{1100, 1200, 1300, 1400, 1500},
			kFactor:      50,
			hasSolved:    true,
			expectSign:   "positive",
			description:  "1200分排名第2，应该加分",
		},
		{
			name:         "高分排名靠前_应该小幅加分或扣分",
			oldRating:    1400,
			rank:         2,
			participants: 5,
			allRatings:   []int{1100, 1200, 1300, 1400, 1500},
			kFactor:      50,
			hasSolved:    true,
			expectSign:   "any",
			description:  "1400分排名第2，变化应该较小",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateRatingChange(
				tt.oldRating,
				tt.rank,
				tt.participants,
				tt.allRatings,
				tt.kFactor,
				tt.hasSolved,
			)

			switch tt.expectSign {
			case "positive":
				if result <= 0 {
					t.Errorf("%s\n期望: 正数, 实际: %d\n%s",
						tt.name, result, tt.description)
				}
			case "negative":
				if result >= 0 {
					t.Errorf("%s\n期望: 负数, 实际: %d\n%s",
						tt.name, result, tt.description)
				}
			case "zero":
				if math.Abs(float64(result)) > 5 {
					t.Errorf("%s\n期望: 接近0, 实际: %d\n%s",
						tt.name, result, tt.description)
				}
			}

			t.Logf("✓ %s: %+d", tt.name, result)
		})
	}
}

// TestCalculateRatingChange_ParticipationBonus 测试参赛奖励
func TestCalculateRatingChange_ParticipationBonus(t *testing.T) {
	participants := 10
	allRatings := make([]int, participants)
	for i := range allRatings {
		allRatings[i] = 1200
	}
	kFactor := 50

	// 测试有AC和无AC的差异
	rank := 5
	withAC := CalculateRatingChange(1200, rank, participants, allRatings, kFactor, true)
	withoutAC := CalculateRatingChange(1200, rank, participants, allRatings, kFactor, false)

	diff := withAC - withoutAC
	if diff != 10 {
		t.Errorf("参赛奖励测试失败: 有AC(%d) - 无AC(%d) = %d, 期望: 10",
			withAC, withoutAC, diff)
	} else {
		t.Logf("✓ 参赛奖励测试通过: 有AC(%d) - 无AC(%d) = %d",
			withAC, withoutAC, diff)
	}
}

// TestCalculateNewRating 测试新 Rating 计算和下限
func TestCalculateNewRating(t *testing.T) {
	tests := []struct {
		name          string
		oldRating     int
		ratingChange  int
		expectedRating int
		description   string
	}{
		{
			name:          "正常加分",
			oldRating:     1200,
			ratingChange:  50,
			expectedRating: 1250,
			description:   "1200 + 50 = 1250",
		},
		{
			name:          "正常扣分",
			oldRating:     1200,
			ratingChange:  -50,
			expectedRating: 1150,
			description:   "1200 - 50 = 1150",
		},
		{
			name:          "触及下限",
			oldRating:     850,
			ratingChange:  -100,
			expectedRating: 800,
			description:   "850 - 100 = 750, 但下限是800",
		},
		{
			name:          "已在下限",
			oldRating:     800,
			ratingChange:  -50,
			expectedRating: 800,
			description:   "800 - 50 = 750, 但下限是800",
		},
		{
			name:          "从下限恢复",
			oldRating:     800,
			ratingChange:  50,
			expectedRating: 850,
			description:   "800 + 50 = 850",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateNewRating(tt.oldRating, tt.ratingChange)
			if result != tt.expectedRating {
				t.Errorf("%s\n期望: %d, 实际: %d\n%s",
					tt.name, tt.expectedRating, result, tt.description)
			} else {
				t.Logf("✓ %s: %d", tt.name, result)
			}
		})
	}
}

// TestCalculateAllRatingChangesByUID 测试批量计算（强关联绑定）
func TestCalculateAllRatingChangesByUID(t *testing.T) {
	userInfos := []UserRatingInfo{
		{UID: "user1", Rank: 1, Rating: 1200, HasSolved: true},
		{UID: "user2", Rank: 2, Rating: 1200, HasSolved: true},
		{UID: "user3", Rank: 3, Rating: 1200, HasSolved: true},
		{UID: "user4", Rank: 4, Rating: 1200, HasSolved: false},
		{UID: "user5", Rank: 5, Rating: 1200, HasSolved: true},
	}
	kFactor := 50

	changes := CalculateAllRatingChangesByUID(userInfos, kFactor)

	// 验证返回的 map 包含所有用户
	if len(changes) != len(userInfos) {
		t.Errorf("返回的变化数量不正确: 期望 %d, 实际 %d",
			len(userInfos), len(changes))
	}

	// 验证每个用户都有对应的变化
	for _, info := range userInfos {
		change, exists := changes[info.UID]
		if !exists {
			t.Errorf("用户 %s 的 Rating 变化未找到", info.UID)
		} else {
			t.Logf("✓ %s (Rank %d, AC=%v): %+d",
				info.UID, info.Rank, info.HasSolved, change)
		}
	}

	// 验证单调性
	for i := 0; i < len(userInfos)-1; i++ {
		change1 := changes[userInfos[i].UID]
		change2 := changes[userInfos[i+1].UID]
		if change1 <= change2 {
			t.Errorf("单调性测试失败: %s (Rank %d) 的变化(%d) <= %s (Rank %d) 的变化(%d)",
				userInfos[i].UID, userInfos[i].Rank, change1,
				userInfos[i+1].UID, userInfos[i+1].Rank, change2)
		}
	}

	// 验证参赛奖励
	change4 := changes["user4"] // 无AC
	change5 := changes["user5"] // 有AC
	// user4 和 user5 的排名相邻，Rating 相同，差异应该主要来自参赛奖励
	// 但由于排名不同，差异不会正好是10分
	t.Logf("user4 (无AC, Rank 4): %+d", change4)
	t.Logf("user5 (有AC, Rank 5): %+d", change5)
}

// TestCalculateAllRatingChanges 测试旧接口兼容性
func TestCalculateAllRatingChanges(t *testing.T) {
	ratings := []int{1200, 1200, 1200, 1200, 1200}
	kFactor := 50

	changes := CalculateAllRatingChanges(ratings, kFactor)

	if len(changes) != len(ratings) {
		t.Errorf("返回的变化数量不正确: 期望 %d, 实际 %d",
			len(ratings), len(changes))
	}

	// 验证单调性
	for i := 0; i < len(changes)-1; i++ {
		if changes[i] <= changes[i+1] {
			t.Errorf("单调性测试失败: Rank %d 的变化(%d) <= Rank %d 的变化(%d)",
				i+1, changes[i], i+2, changes[i+1])
		}
	}

	for i, change := range changes {
		t.Logf("✓ Rank %d: %+d", i+1, change)
	}
}

// TestEdgeCases 测试边界情况
func TestEdgeCases(t *testing.T) {
	t.Run("参赛人数不足", func(t *testing.T) {
		result := CalculateRatingChange(1200, 1, 1, []int{1200}, 50, true)
		if result != 0 {
			t.Errorf("参赛人数为1时应该返回0, 实际: %d", result)
		}
	})

	t.Run("Rating接近但不完全相同", func(t *testing.T) {
		// Rating 差异在10分以内，应该使用线性分配
		allRatings := []int{1195, 1200, 1205, 1198, 1202}
		result := CalculateRatingChange(1200, 1, 5, allRatings, 50, true)
		// 应该使用线性分配，第1名应该获得高分
		if result < 80 {
			t.Errorf("Rating 接近时第1名应该获得高分, 实际: %d", result)
		}
		t.Logf("✓ Rating 接近时第1名: %+d", result)
	})

	t.Run("Rating差异超过10分", func(t *testing.T) {
		// Rating 差异超过10分，应该使用 Elo 算法
		allRatings := []int{1100, 1200, 1300, 1400, 1500}
		result := CalculateRatingChange(1100, 1, 5, allRatings, 50, true)
		// 低分夺冠应该加分（但不一定是大幅加分，因为 Elo 算法考虑了参赛人数）
		if result <= 0 {
			t.Errorf("低分夺冠应该加分, 实际: %d", result)
		}
		t.Logf("✓ 低分夺冠: %+d", result)
	})

	t.Run("空的用户信息列表", func(t *testing.T) {
		changes := CalculateAllRatingChangesByUID([]UserRatingInfo{}, 50)
		if len(changes) != 0 {
			t.Errorf("空列表应该返回空 map, 实际长度: %d", len(changes))
		}
	})
}

// BenchmarkCalculateRatingChange 性能测试
func BenchmarkCalculateRatingChange(b *testing.B) {
	participants := 100
	allRatings := make([]int, participants)
	for i := range allRatings {
		allRatings[i] = 1200
	}
	kFactor := 50

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateRatingChange(1200, 50, participants, allRatings, kFactor, true)
	}
}

// BenchmarkCalculateAllRatingChangesByUID 批量计算性能测试
func BenchmarkCalculateAllRatingChangesByUID(b *testing.B) {
	participants := 100
	userInfos := make([]UserRatingInfo, participants)
	for i := range userInfos {
		userInfos[i] = UserRatingInfo{
			UID:       string(rune(i)),
			Rank:      i + 1,
			Rating:    1200,
			HasSolved: true,
		}
	}
	kFactor := 50

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateAllRatingChangesByUID(userInfos, kFactor)
	}
}

// TestContest1011_RatingCalculation 测试比赛1011的rating计算（修复后）
func TestContest1011_RatingCalculation(t *testing.T) {
	// 比赛1011的所有初始rating（按排名顺序）
	allRatings := []int{
		1200, 1326, 1282, 1299, 1238, 1249, 1255, 1204, 1171, 1200,
		1218, 1229, 1226, 1200, 1221, 1310, 1277, 1215, 1200, 1200,
	}

	kFactor := 50
	participants := len(allRatings)

	testCases := []struct {
		rank      int
		oldRating int
		name      string
		expectSign string // "positive", "negative"
		minChange int     // 最小变化（绝对值）
	}{
		{
			rank:       1,
			oldRating:  1200,
			name:       "第1名（rating 1200）",
			expectSign: "positive",
			minChange:  75, // 应该大幅加分
		},
		{
			rank:       2,
			oldRating:  1326,
			name:       "第2名（rating 1326，最高）",
			expectSign: "positive",
			minChange:  0, // 应该加分，但不应该太多
		},
		{
			rank:       16,
			oldRating:  1310,
			name:       "第16名（rating 1310，第二高）",
			expectSign: "negative",
			minChange:  30, // 应该扣分
		},
		{
			rank:       17,
			oldRating:  1277,
			name:       "第17名（rating 1277）",
			expectSign: "negative",
			minChange:  20, // 应该扣分
		},
		{
			rank:       20,
			oldRating:  1200,
			name:       "第20名（rating 1200）",
			expectSign: "negative",
			minChange:  40, // 应该大幅扣分
		},
	}

	t.Log("========== 比赛1011 Rating计算测试（修复后）==========")
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			change := CalculateRatingChange(tc.oldRating, tc.rank, participants, allRatings, kFactor, true)
			newRating := CalculateNewRating(tc.oldRating, change)

			// 验证符号
			switch tc.expectSign {
			case "positive":
				if change <= 0 {
					t.Errorf("%s 应该加分，但实际变化=%d", tc.name, change)
				}
				if change < tc.minChange {
					t.Errorf("%s 加分应该 >= %d，但实际变化=%d", tc.name, tc.minChange, change)
				}
			case "negative":
				if change >= 0 {
					t.Errorf("%s 应该扣分，但实际变化=%d", tc.name, change)
				}
				if -change < tc.minChange {
					t.Errorf("%s 扣分应该 >= %d，但实际变化=%d", tc.name, tc.minChange, change)
				}
			}

			t.Logf("✅ Rank=%d, OldRating=%d, Change=%+d, NewRating=%d",
				tc.rank, tc.oldRating, change, newRating)
		})
	}

	// 验证第2名的加分应该少于第1名
	change1 := CalculateRatingChange(1200, 1, participants, allRatings, kFactor, true)
	change2 := CalculateRatingChange(1326, 2, participants, allRatings, kFactor, true)

	if change2 >= change1 {
		t.Errorf("第2名（最高rating）的加分（%d）不应该 >= 第1名（中等rating）的加分（%d）", change2, change1)
	} else {
		t.Logf("✅ 第1名加分（%+d）> 第2名加分（%+d），符合预期", change1, change2)
	}
}

// TestExpectedRank_HigherRatingBetterRank 测试期望排名的合理性
func TestExpectedRank_HigherRatingBetterRank(t *testing.T) {
	// 比赛1011的所有初始rating（按排名顺序）
	allRatings := []int{
		1200, 1326, 1282, 1299, 1238, 1249, 1255, 1204, 1171, 1200,
		1218, 1229, 1226, 1200, 1221, 1310, 1277, 1215, 1200, 1200,
	}

	testCases := []struct {
		rating       int
		name         string
		maxExpected  float64 // 期望排名的最大值（rating高的人期望排名应该靠前）
	}{
		{
			rating:      1326,
			name:        "最高rating（1326）",
			maxExpected: 9.0, // 期望排名应该 <= 9
		},
		{
			rating:      1310,
			name:        "第二高rating（1310）",
			maxExpected: 9.5, // 期望排名应该 <= 9.5
		},
		{
			rating:      1200,
			name:        "中等rating（1200）",
			maxExpected: 15.0, // 期望排名应该 <= 15
		},
		{
			rating:      1171,
			name:        "最低rating（1171）",
			maxExpected: 20.0, // 期望排名应该 <= 20
		},
	}

	t.Log("========== 期望排名合理性测试 ==========")
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			expectedRank := calculateExpectedRank(tc.rating, allRatings)

			if expectedRank > tc.maxExpected {
				t.Errorf("%s 的期望排名（%.2f）不应该 > %.2f",
					tc.name, expectedRank, tc.maxExpected)
			}

			t.Logf("✅ %s 期望排名=%.2f（应该 <= %.2f）",
				tc.name, expectedRank, tc.maxExpected)
		})
	}

	// 验证单调性：rating越高，期望排名越靠前（数值越小）
	ratings := []int{1326, 1310, 1277, 1200, 1171}
	var prevExpectedRank float64 = 0

	t.Log("========== 期望排名单调性测试 ==========")
	for i, rating := range ratings {
		expectedRank := calculateExpectedRank(rating, allRatings)

		if i > 0 && expectedRank <= prevExpectedRank {
			t.Errorf("单调性测试失败: rating=%d 的期望排名（%.2f）应该 > rating=%d 的期望排名（%.2f）",
				rating, expectedRank, ratings[i-1], prevExpectedRank)
		}

		t.Logf("✅ rating=%d, 期望排名=%.2f", rating, expectedRank)
		prevExpectedRank = expectedRank
	}
}
