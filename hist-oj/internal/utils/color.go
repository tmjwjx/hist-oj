package utils

// RatingLevel Rating等级
type RatingLevel struct {
	MinRating int
	MaxRating int
	Color     string
	Name      string
	NameZh    string
}

var RatingLevels = []RatingLevel{
	{0, 0, "#000000", "Unrated", "未评级"},
	{1, 1199, "#808080", "Newbie", "新手"},
	{1200, 1399, "#008000", "Pupil", "学徒"},
	{1400, 1599, "#03A89E", "Specialist", "专家"},
	{1600, 1899, "#0000FF", "Expert", "专家"},
	{1900, 2099, "#AA00AA", "Candidate Master", "候选大师"},
	{2100, 2299, "#FF8C00", "Master", "大师"},
	{2300, 2399, "#FF8C00", "International Master", "国际大师"},
	{2400, 2599, "#FF0000", "Grandmaster", "特级大师"},
	{2600, 2999, "#FF0000", "International Grandmaster", "国际特级大师"},
	{3000, 3999, "#FF0000", "Legendary Grandmaster", "传奇特级大师"},
	{4000, 999999, "#FF0000", "Tourist", "Tourist"},
}

// GetRatingColorInfo 根据rating获取颜色信息
func GetRatingColorInfo(rating int) RatingLevel {
	if rating == 0 {
		return RatingLevels[0]
	}

	for _, level := range RatingLevels {
		if rating >= level.MinRating && rating <= level.MaxRating {
			return level
		}
	}

	return RatingLevels[0]
}
