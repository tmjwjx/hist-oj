package utils

import "testing"

func TestCalculateAllRatingChangesByUID_EqualRatingsMonotonic(t *testing.T) {
	userInfos := []UserRatingInfo{
		{UID: "u1", Rank: 1, Rating: 1200},
		{UID: "u2", Rank: 2, Rating: 1200},
		{UID: "u3", Rank: 3, Rating: 1200},
		{UID: "u4", Rank: 4, Rating: 1200},
		{UID: "u5", Rank: 5, Rating: 1200},
		{UID: "u6", Rank: 6, Rating: 1200},
		{UID: "u7", Rank: 7, Rating: 1200},
	}

	changes := CalculateAllRatingChangesByUID(userInfos, 32)

	ordered := make([]int, 0, len(userInfos))
	for _, info := range userInfos {
		delta, ok := changes[info.UID]
		if !ok {
			t.Fatalf("missing delta for uid=%s", info.UID)
		}
		ordered = append(ordered, delta)
	}

	if ordered[0] <= 0 {
		t.Fatalf("rank1 should gain rating, got %d", ordered[0])
	}
	if ordered[len(ordered)-1] >= 0 {
		t.Fatalf("last rank should lose rating, got %d", ordered[len(ordered)-1])
	}

	for i := 1; i < len(ordered); i++ {
		if ordered[i] > ordered[i-1] {
			t.Fatalf("delta should be non-increasing by rank, rank=%d delta=%d > prev=%d",
				i+1, ordered[i], ordered[i-1])
		}
	}

	sum := 0
	for _, d := range ordered {
		sum += d
	}
	if sum > 0 {
		t.Fatalf("total delta should not be positive after CF corrections, got %d", sum)
	}
}

func TestCalculateAllRatingChangesByUID_UIDBinding(t *testing.T) {
	userInfos := []UserRatingInfo{
		{UID: "alice", Rank: 3, Rating: 1450},
		{UID: "bob", Rank: 1, Rating: 1300},
		{UID: "carol", Rank: 2, Rating: 1500},
	}

	changes := CalculateAllRatingChangesByUID(userInfos, 85)
	if len(changes) != len(userInfos) {
		t.Fatalf("unexpected map size: got=%d want=%d", len(changes), len(userInfos))
	}
	for _, info := range userInfos {
		if _, ok := changes[info.UID]; !ok {
			t.Fatalf("missing uid in changes map: %s", info.UID)
		}
	}
}

func TestCalculateRatingChange_BaseDirection(t *testing.T) {
	ratings := []int{1200, 1200, 1200, 1200, 1200}
	first := CalculateRatingChange(1200, 1, 5, ratings, 32, true)
	last := CalculateRatingChange(1200, 5, 5, ratings, 32, true)

	if first <= 0 {
		t.Fatalf("rank1 raw delta should be positive, got %d", first)
	}
	if last >= 0 {
		t.Fatalf("last rank raw delta should be negative, got %d", last)
	}
}

func TestCalculateNewRating_MinFloor(t *testing.T) {
	if got := CalculateNewRating(1, -100); got != MinRating {
		t.Fatalf("new rating floor mismatch: got=%d want=%d", got, MinRating)
	}
	if got := CalculateNewRating(0, -200); got != MinRating {
		t.Fatalf("zero floor mismatch: got=%d want=%d", got, MinRating)
	}
	if got := CalculateNewRating(1200, 50); got != 1250 {
		t.Fatalf("new rating add mismatch: got=%d want=1250", got)
	}
}
