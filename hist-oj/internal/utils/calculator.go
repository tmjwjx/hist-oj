package utils

import (
	"math"
	"sort"
	"strconv"
)

const (
	cfSearchRatingLow  = -10000
	cfSearchRatingHigh = 10000
)

// CalculateRatingChange 计算单个用户的原始 CF rating 变化（不含全局修正）
// 兼容旧接口保留 kFactor/hasSolved 参数，但 CF 算法不使用它们。
func CalculateRatingChange(oldRating, rank, participants int, allRatings []int, kFactor int, hasSolved bool) int {
	_ = kFactor
	_ = hasSolved

	if participants < 2 {
		return 0
	}

	if len(allRatings) != participants {
		participants = len(allRatings)
		if participants < 2 {
			return 0
		}
	}

	if rank < 1 {
		rank = 1
	}
	if rank > participants {
		rank = participants
	}

	skipIndex := rank - 1
	if skipIndex < 0 || skipIndex >= len(allRatings) {
		skipIndex = -1
	}

	seed := calculateSeedAtRating(oldRating, allRatings, skipIndex)
	midRank := math.Sqrt(float64(rank) * seed)
	needRating := calculateRatingBySeed(midRank, allRatings, skipIndex)

	return (needRating - oldRating) / 2
}

func winProbabilityAgainst(rating, opponentRating int) float64 {
	return 1.0 / (1.0 + math.Pow(10.0, float64(rating-opponentRating)/400.0))
}

func calculateSeedAtRating(rating int, allRatings []int, skipIndex int) float64 {
	seed := 1.0
	for idx, otherRating := range allRatings {
		if idx == skipIndex {
			continue
		}
		seed += winProbabilityAgainst(rating, otherRating)
	}
	return seed
}

func calculateRatingBySeed(targetSeed float64, allRatings []int, skipIndex int) int {
	left := cfSearchRatingLow
	right := cfSearchRatingHigh
	for right-left > 1 {
		mid := (left + right) / 2
		if calculateSeedAtRating(mid, allRatings, skipIndex) < targetSeed {
			right = mid
		} else {
			left = mid
		}
	}
	return left
}

// Rating 下限（CF 无固定下限，这里仅防止出现负数）
const MinRating = 0

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
	HasSolved bool   // 保留字段，CF算法不使用
}

type cfContestant struct {
	UID    string
	Rank   int
	Rating int
	Delta  int
}

func applyGlobalCorrectionOne(contestants []cfContestant) {
	if len(contestants) == 0 {
		return
	}
	sum := 0
	for _, c := range contestants {
		sum += c.Delta
	}
	inc := int(math.Floor(float64(-sum)/float64(len(contestants)) - 1.0))
	for i := range contestants {
		contestants[i].Delta += inc
	}
}

func applyGlobalCorrectionTwo(contestants []cfContestant) {
	n := len(contestants)
	if n == 0 {
		return
	}
	ordered := make([]*cfContestant, n)
	for i := range contestants {
		ordered[i] = &contestants[i]
	}
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].Rating != ordered[j].Rating {
			return ordered[i].Rating > ordered[j].Rating
		}
		return ordered[i].Rank < ordered[j].Rank
	})

	topCount := int(math.Min(float64(n), 4.0*math.Sqrt(float64(n))))
	if topCount < 1 {
		topCount = 1
	}

	topDeltaSum := 0
	for i := 0; i < topCount; i++ {
		topDeltaSum += ordered[i].Delta
	}

	inc := int(math.Floor(float64(-topDeltaSum) / float64(topCount)))
	if inc < -10 {
		inc = -10
	}
	if inc > 0 {
		inc = 0
	}
	for i := range contestants {
		contestants[i].Delta += inc
	}
}

// CalculateAllRatingChangesByUID 通过UID批量计算rating变化（强关联绑定）
// 返回 map[UID]ratingChange，确保每个用户都通过UID关联到正确的rating变化
func CalculateAllRatingChangesByUID(userInfos []UserRatingInfo, kFactor int) map[string]int {
	_ = kFactor

	participants := len(userInfos)
	if participants == 0 {
		return make(map[string]int)
	}

	contestants := make([]cfContestant, participants)
	for i, info := range userInfos {
		contestants[i] = cfContestant{
			UID:    info.UID,
			Rank:   info.Rank,
			Rating: info.Rating,
		}
	}
	sort.SliceStable(contestants, func(i, j int) bool {
		if contestants[i].Rank != contestants[j].Rank {
			return contestants[i].Rank < contestants[j].Rank
		}
		return contestants[i].UID < contestants[j].UID
	})

	allRatings := make([]int, participants)
	for i, c := range contestants {
		allRatings[i] = c.Rating
	}

	for i := range contestants {
		seed := calculateSeedAtRating(contestants[i].Rating, allRatings, i)
		midRank := math.Sqrt(float64(contestants[i].Rank) * seed)
		needRating := calculateRatingBySeed(midRank, allRatings, i)
		contestants[i].Delta = (needRating - contestants[i].Rating) / 2
	}

	applyGlobalCorrectionOne(contestants)
	applyGlobalCorrectionTwo(contestants)

	changes := make(map[string]int, participants)
	for _, c := range contestants {
		changes[c.UID] = c.Delta
	}
	return changes
}

// CalculateAllRatingChanges 批量计算所有参赛者的rating变化（保留旧接口以兼容）
// 注意：建议使用 CalculateAllRatingChangesByUID 进行强关联绑定
func CalculateAllRatingChanges(ratings []int, kFactor int) []int {
	participants := len(ratings)
	changes := make([]int, participants)

	userInfos := make([]UserRatingInfo, participants)
	for i := 0; i < participants; i++ {
		userInfos[i] = UserRatingInfo{
			UID:       strconv.Itoa(i),
			Rank:      i + 1,
			Rating:    ratings[i],
			HasSolved: true,
		}
	}

	changesMap := CalculateAllRatingChangesByUID(userInfos, kFactor)
	for i := 0; i < participants; i++ {
		changes[i] = changesMap[strconv.Itoa(i)]
	}
	return changes
}
