package service

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/hoj/hist-oj/internal/client"
	"github.com/hoj/hist-oj/internal/config"
	"github.com/hoj/hist-oj/internal/model"
	"github.com/hoj/hist-oj/internal/utils"
)

type RatingService struct {
	db     *gorm.DB
	config *config.RatingConfig
}

func NewRatingService(db *gorm.DB, cfg *config.RatingConfig) *RatingService {
	return &RatingService{
		db:     db,
		config: cfg,
	}
}

// CalculateContestRating 计算比赛的rating变化
func (s *RatingService) CalculateContestRating(contestID int64) ([]model.RatingHistory, error) {
	logger := utils.GetLogger()
	logger.Info("开始计算比赛rating", zap.Int64("contest_id", contestID))

	// 先检查比赛状态（在事务外进行，避免不必要的回滚）
	var status model.ContestRatingStatus
	if err := s.db.Where("contest_id = ?", contestID).First(&status).Error; err == nil {
		if status.RatingCalculated {
			logger.Warn("比赛rating已计算过", zap.Int64("contest_id", contestID))
			var histories []model.RatingHistory
			s.db.Where("contest_id = ?", contestID).Find(&histories)
			return histories, nil
		}
		if !status.IsRated {
			logger.Info("比赛不是计分比赛", zap.Int64("contest_id", contestID))
			return nil, nil
		}
	} else if err != gorm.ErrRecordNotFound {
		logger.Error("查询比赛状态失败", zap.Int64("contest_id", contestID), zap.Error(err))
		return nil, fmt.Errorf("查询比赛状态失败: %w", err)
	}

	// 使用事务确保数据一致性（在确认需要计算后再开启事务）
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			if tx != nil {
				tx.Rollback()
			}
			logger.Error("计算rating时发生panic", 
				zap.Int64("contest_id", contestID),
				zap.Any("panic", r))
		}
	}()

	// 获取比赛信息
	contestInfo, err := client.GetContestInfo(contestID)
	if err != nil {
		tx.Rollback()
		logger.Error("获取比赛信息失败", zap.Int64("contest_id", contestID), zap.Error(err))
		return nil, fmt.Errorf("获取比赛信息失败: %w", err)
	}
	logger.Info("获取比赛信息成功", 
		zap.Int64("contest_id", contestID),
		zap.String("title", contestInfo.Title),
		zap.Int("type", contestInfo.Type))

	// 获取比赛排名数据
	rankReq := client.ContestRankDTO{
		CID:         contestID,
		CurrentPage: 1,
		Limit:       10000, // 获取所有排名
		RemoveStar:  true,
		ContainsEnd: true,
	}

	rankResp, err := client.GetContestOutsideScoreboard(rankReq)
	if err != nil {
		logger.Warn("获取比赛外榜失败，尝试使用认证接口", 
			zap.Int64("contest_id", contestID),
			zap.Error(err))
		// 如果外榜失败，尝试使用需要认证的接口
		rankResp, err = client.GetContestRank(rankReq)
		if err != nil {
			tx.Rollback()
			logger.Error("获取比赛排名失败", zap.Int64("contest_id", contestID), zap.Error(err))
			return nil, fmt.Errorf("获取比赛排名失败: %w", err)
		}
		logger.Info("使用认证接口获取排名成功", zap.Int64("contest_id", contestID))
	} else {
		logger.Info("获取比赛外榜成功", 
			zap.Int64("contest_id", contestID),
			zap.Int("participants", len(rankResp.Records)))
	}

	if len(rankResp.Records) < 3 {
		logger.Warn("参赛人数不足", 
			zap.Int64("contest_id", contestID),
			zap.Int("participants", len(rankResp.Records)))
		tx.Rollback()
		return nil, nil
	}

	// 获取所有参赛者的当前rating
	uids := make([]string, 0, len(rankResp.Records))
	for _, record := range rankResp.Records {
		uids = append(uids, record.UID)
	}

	var userRecords []model.UserRecord
	if err := tx.Where("uid IN ?", uids).Find(&userRecords).Error; err != nil {
		tx.Rollback()
		logger.Error("查询用户rating失败", 
			zap.Int64("contest_id", contestID),
			zap.Int("user_count", len(uids)),
			zap.Error(err))
		return nil, fmt.Errorf("查询用户rating失败: %w", err)
	}
	logger.Info("查询用户rating成功", 
		zap.Int64("contest_id", contestID),
		zap.Int("found_users", len(userRecords)),
		zap.Int("total_users", len(uids)))

	// 构建rating映射
	ratingMap := make(map[string]int)
	for _, ur := range userRecords {
		if ur.HistRating != nil {
			ratingMap[ur.UID] = *ur.HistRating
		} else {
			ratingMap[ur.UID] = s.config.InitialRating
		}
	}

	// 构建rating列表（按排名顺序）
	ratings := make([]int, 0, len(rankResp.Records))
	for _, record := range rankResp.Records {
		rating, ok := ratingMap[record.UID]
		if !ok {
			rating = s.config.InitialRating
		}
		ratings = append(ratings, rating)
	}

	// 计算所有rating变化
	changes := utils.CalculateAllRatingChanges(ratings, s.config.KFactor)
	logger.Info("rating变化计算完成",
		zap.Int64("contest_id", contestID),
		zap.Int("participants", len(rankResp.Records)),
		zap.Int("changes_count", len(changes)))

	// 查询每个用户的历史参赛次数（用于新手保护）
	contestCountMap := make(map[string]int64)
	for _, uid := range uids {
		var count int64
		tx.Model(&model.RatingHistory{}).Where("uid = ?", uid).Count(&count)
		contestCountMap[uid] = count
	}

	// 保存rating历史记录并更新用户rating
	histories := make([]model.RatingHistory, 0, len(rankResp.Records))
	now := time.Now()

	// 批量更新用户rating
	updateCount := 0
	for i, record := range rankResp.Records {
		oldRating := ratings[i]
		change := changes[i]

		// 新手保护：前3场比赛掉分减半
		contestCount := contestCountMap[record.UID]
		if contestCount < 3 && change < 0 {
			change = change / 2
			logger.Debug("新手保护生效",
				zap.String("uid", record.UID),
				zap.Int64("contest_count", contestCount),
				zap.Int("original_change", changes[i]),
				zap.Int("protected_change", change))
		}

		// 参与激励：过题了就额外+5分
		if record.AC > 0 {
			change += 5
			logger.Debug("参与激励生效",
				zap.String("uid", record.UID),
				zap.Int("ac", record.AC),
				zap.Int("bonus", 5))
		}

		newRating := utils.CalculateNewRating(oldRating, change)
		rank := i + 1

		history := model.RatingHistory{
			UID:          record.UID,
			ContestID:    uint64(contestID),
			OldRating:    &oldRating,
			NewRating:    newRating,
			RatingChange: change,
			Rank:         rank,
			Participants: len(rankResp.Records),
			CreatedAt:    now,
		}
		histories = append(histories, history)

		// 更新用户rating（如果用户不存在则创建）
		result := tx.Model(&model.UserRecord{}).
			Where("uid = ?", record.UID).
			Updates(map[string]interface{}{
				"hist_rating": newRating,
			})
		if result.Error != nil {
			tx.Rollback()
			logger.Error("更新用户rating失败", 
				zap.Int64("contest_id", contestID),
				zap.String("uid", record.UID),
				zap.Error(result.Error))
			return nil, fmt.Errorf("更新用户rating失败: %w", result.Error)
		}
		// 如果用户不存在（RowsAffected=0），尝试创建
		if result.RowsAffected == 0 {
			userRecord := model.UserRecord{
				UID:        record.UID,
				HistRating: &newRating,
			}
			if err := tx.Create(&userRecord).Error; err != nil {
				// 创建失败可能是用户已存在（并发情况），记录警告但继续
				logger.Warn("创建用户rating记录失败，可能已存在", 
					zap.String("uid", record.UID),
					zap.Error(err))
			} else {
				updateCount++
			}
		} else {
			updateCount++
		}
	}

	logger.Info("用户rating更新完成", 
		zap.Int64("contest_id", contestID),
		zap.Int("updated_users", updateCount),
		zap.Int("total_users", len(histories)))

	// 批量保存历史记录
	if err := tx.CreateInBatches(histories, 100).Error; err != nil {
		tx.Rollback()
		logger.Error("保存rating历史失败", 
			zap.Int64("contest_id", contestID),
			zap.Int("history_count", len(histories)),
			zap.Error(err))
		return nil, fmt.Errorf("保存rating历史失败: %w", err)
	}

	// 更新比赛rating状态
	calculatedAt := now
	status = model.ContestRatingStatus{
		ContestID:       uint64(contestID),
		IsRated:         true,
		RatingCalculated: true,
		CalculatedAt:    &calculatedAt,
	}
	if err := tx.Save(&status).Error; err != nil {
		tx.Rollback()
		logger.Error("更新比赛rating状态失败", 
			zap.Int64("contest_id", contestID),
			zap.Error(err))
		return nil, fmt.Errorf("更新比赛rating状态失败: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		logger.Error("提交事务失败", 
			zap.Int64("contest_id", contestID),
			zap.Error(err))
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	logger.Info("比赛rating计算完成", 
		zap.Int64("contest_id", contestID),
		zap.Int("participants", len(histories)),
		zap.Int("updated_users", updateCount))

	return histories, nil
}

// CanCalculateRating 检查是否可以计算rating
func (s *RatingService) CanCalculateRating(contestID int64) bool {
	var status model.ContestRatingStatus
	if err := s.db.Where("contest_id = ?", contestID).First(&status).Error; err != nil {
		return false
	}
	if !status.IsRated {
		return false
	}
	if status.RatingCalculated {
		return false
	}
	return true
}

