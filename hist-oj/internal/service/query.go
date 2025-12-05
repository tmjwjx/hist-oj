package service

import (
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/hoj/hist-oj/internal/model"
	"github.com/hoj/hist-oj/internal/utils"
)

type QueryService struct {
	db *gorm.DB
}

func NewQueryService(db *gorm.DB) *QueryService {
	return &QueryService{db: db}
}

// GetUserRating 获取用户rating信息
func (s *QueryService) GetUserRating(uid string) (map[string]interface{}, error) {
	logger := utils.GetLogger()
	var userRecord model.UserRecord
	if err := s.db.Where("uid = ?", uid).First(&userRecord).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Debug("用户不存在", zap.String("uid", uid))
			return nil, nil
		}
		logger.Error("查询用户记录失败", zap.String("uid", uid), zap.Error(err))
		return nil, err
	}

	rating := 0
	if userRecord.HistRating != nil {
		rating = *userRecord.HistRating
	}

	// 获取最高rating
	var maxRating *int
	if err := s.db.Model(&model.RatingHistory{}).
		Where("uid = ?", uid).
		Select("MAX(new_rating) as max_rating").
		Scan(&maxRating).Error; err != nil {
		logger.Warn("查询最高rating失败", zap.String("uid", uid), zap.Error(err))
		// 继续执行，使用当前rating作为最高rating
	}
	
	finalMaxRating := rating
	if maxRating != nil && *maxRating > rating {
		finalMaxRating = *maxRating
	}

	// 获取颜色信息
	colorInfo := utils.GetRatingColorInfo(rating)

	result := map[string]interface{}{
		"uid":      uid,
		"rating":   rating,
		"maxRating": finalMaxRating,
		"color":    colorInfo.Color,
		"level":    colorInfo.Name,
		"levelZh":  colorInfo.NameZh,
	}

	return result, nil
}

// GetRatingHistory 获取用户rating历史记录
func (s *QueryService) GetRatingHistory(uid string, page, limit int) (map[string]interface{}, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	offset := (page - 1) * limit

	var histories []model.RatingHistory
	var total int64

	// 查询总数
	if err := s.db.Model(&model.RatingHistory{}).Where("uid = ?", uid).Count(&total).Error; err != nil {
		logger := utils.GetLogger()
		logger.Error("查询rating历史总数失败", 
			zap.String("uid", uid),
			zap.Error(err))
		return nil, err
	}

	// 查询记录
	if err := s.db.Where("uid = ?", uid).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&histories).Error; err != nil {
		logger := utils.GetLogger()
		logger.Error("查询rating历史失败", 
			zap.String("uid", uid),
			zap.Int("page", page),
			zap.Int("limit", limit),
			zap.Error(err))
		return nil, err
	}

	result := map[string]interface{}{
		"total":   total,
		"page":    page,
		"limit":   limit,
		"records": histories,
	}

	return result, nil
}

// GetRatingColor 获取rating颜色信息
func (s *QueryService) GetRatingColor(rating int) map[string]interface{} {
	colorInfo := utils.GetRatingColorInfo(rating)

	return map[string]interface{}{
		"rating": rating,
		"color":  colorInfo.Color,
		"name":   colorInfo.Name,
		"nameZh": colorInfo.NameZh,
		"level":  colorInfo.Name,
	}
}

// GetContestParticipantsRating 获取比赛所有参赛者的rating信息
func (s *QueryService) GetContestParticipantsRating(contestID int64) ([]map[string]interface{}, error) {
	logger := utils.GetLogger()

	// 查询该比赛的所有rating历史记录
	var histories []model.RatingHistory
	if err := s.db.Where("contest_id = ?", contestID).
		Order("rank ASC").
		Find(&histories).Error; err != nil {
		logger.Error("查询比赛参赛者rating失败",
			zap.Int64("contest_id", contestID),
			zap.Error(err))
		return nil, err
	}

	// 构建返回结果
	results := make([]map[string]interface{}, 0, len(histories))
	for _, h := range histories {
		colorInfo := utils.GetRatingColorInfo(h.NewRating)
		results = append(results, map[string]interface{}{
			"uid":          h.UID,
			"rank":         h.Rank,
			"oldRating":    h.OldRating,
			"newRating":    h.NewRating,
			"ratingChange": h.RatingChange,
			"color":        colorInfo.Color,
			"level":        colorInfo.Name,
		})
	}

	return results, nil
}

