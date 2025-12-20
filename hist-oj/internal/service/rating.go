package service

import (
	"fmt"
	"sort"
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
	err := s.db.Where("contest_id = ?", contestID).First(&status).Error
	if err == gorm.ErrRecordNotFound {
		// 如果找不到记录，说明比赛还没有设置 Rating 类型，不计算
		logger.Info("比赛未设置Rating类型", zap.Int64("contest_id", contestID))
		return nil, nil
	} else if err != nil {
		logger.Error("查询比赛状态失败", zap.Int64("contest_id", contestID), zap.Error(err))
		return nil, fmt.Errorf("查询比赛状态失败: %w", err)
	}

	// 检查是否已计算过
	if status.RatingCalculated {
		logger.Warn("比赛rating已计算过", zap.Int64("contest_id", contestID))
		var histories []model.RatingHistory
		s.db.Where("contest_id = ?", contestID).Find(&histories)
		return histories, nil
	}

	// 检查是否为 Rating 比赛
	if !status.IsRated {
		logger.Info("比赛不是Rating比赛", zap.Int64("contest_id", contestID))
		return nil, nil
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

	// 直接从数据库获取比赛排名数据（避免API认证问题）
	// 只统计比赛时间范围内的提交记录
	logger.Info("从数据库获取比赛排名",
		zap.Int64("contest_id", contestID),
		zap.Time("start_time", contestInfo.StartTime),
		zap.Time("end_time", contestInfo.EndTime))

	var contestRecords []model.ContestRecord
	if err := tx.Where("cid = ? AND submit_time >= ? AND submit_time <= ?",
		contestID, contestInfo.StartTime, contestInfo.EndTime).
		Find(&contestRecords).Error; err != nil {
		tx.Rollback()
		logger.Error("从数据库获取比赛记录失败", zap.Int64("contest_id", contestID), zap.Error(err))
		return nil, fmt.Errorf("获取比赛记录失败: %w", err)
	}

	logger.Info("获取到比赛提交记录",
		zap.Int64("contest_id", contestID),
		zap.Int("total_records", len(contestRecords)))

	// 聚合每个用户的比赛成绩
	type UserScore struct {
		UID        string
		Username   string
		ACProblems map[string]bool      // 记录已AC的题目（按 DisplayID）
		ProblemInfo map[string]*struct { // 每道题的详细信息（按 DisplayID）
			ACTime     uint64 // AC时间
			ErrorCount int    // AC前的错误次数
		}
		TotalTime  int64 // 总用时（秒，包含罚时）
	}

	userScores := make(map[string]*UserScore)

	// 按时间排序处理记录（确保先处理早期提交）
	sort.Slice(contestRecords, func(i, j int) bool {
		return contestRecords[i].Time < contestRecords[j].Time
	})

	logger.Info("开始聚合用户比赛成绩",
		zap.Int64("contest_id", contestID),
		zap.Int("total_records", len(contestRecords)))

	for _, record := range contestRecords {
		if _, exists := userScores[record.UID]; !exists {
			userScores[record.UID] = &UserScore{
				UID:        record.UID,
				Username:   record.Username,
				ACProblems: make(map[string]bool),
				ProblemInfo: make(map[string]*struct {
					ACTime     uint64
					ErrorCount int
				}),
				TotalTime:  0,
			}
		}

		user := userScores[record.UID]

		// 如果这道题已经AC过，跳过后续提交
		if user.ACProblems[record.DisplayID] {
			continue
		}

		// 初始化题目信息
		if user.ProblemInfo[record.DisplayID] == nil {
			user.ProblemInfo[record.DisplayID] = &struct {
				ACTime     uint64
				ErrorCount int
			}{
				ACTime:     0,
				ErrorCount: 0,
			}
		}

		// status == 1 表示 AC
		// status == 0 表示未AC但不罚时（编译错误、格式错误等）
		// status == -1 表示未AC且算罚时（WA、TLE、RE等）
		if record.Status == 1 {
			user.ACProblems[record.DisplayID] = true
			user.ProblemInfo[record.DisplayID].ACTime = record.Time

			// 计算该题的总时间：AC时间 + 罚时（错误次数 × 20分钟）
			penaltyTime := int64(user.ProblemInfo[record.DisplayID].ErrorCount) * 20 * 60
			problemTotalTime := int64(record.Time) + penaltyTime
			user.TotalTime += problemTotalTime

			logger.Debug("用户AC题目",
				zap.String("uid", record.UID),
				zap.String("username", record.Username),
				zap.String("display_id", record.DisplayID),
				zap.Uint64("problem_id", record.PID),
				zap.Uint64("ac_time", record.Time),
				zap.Int("error_count", user.ProblemInfo[record.DisplayID].ErrorCount),
				zap.Int64("penalty_time", penaltyTime),
				zap.Int64("problem_total_time", problemTotalTime))
		} else if record.Status == -1 {
			// status == -1 才计入罚时（WA、TLE、RE等）
			user.ProblemInfo[record.DisplayID].ErrorCount++
			logger.Debug("用户错误提交（计入罚时）",
				zap.String("uid", record.UID),
				zap.String("username", record.Username),
				zap.String("display_id", record.DisplayID),
				zap.Uint64("problem_id", record.PID),
				zap.Int("status", record.Status),
				zap.Int("current_error_count", user.ProblemInfo[record.DisplayID].ErrorCount))
		}
		// status == 0 不计入罚时，直接跳过
	}

	logger.Info("用户比赛成绩聚合完成",
		zap.Int64("contest_id", contestID),
		zap.Int("user_count", len(userScores)))

	// 转换为 RankResponse 格式
	rankResp := &client.ContestRankResponse{
		Total:   len(userScores),
		Records: make([]client.ContestRankRecord, 0, len(userScores)),
	}

	for _, user := range userScores {
		rankResp.Records = append(rankResp.Records, client.ContestRankRecord{
			UID:       user.UID,
			Username:  user.Username,
			AC:        len(user.ACProblems),
			TotalTime: user.TotalTime,
		})
	}

	// 按 AC 数降序，AC 数相同按总用时升序排序
	sort.Slice(rankResp.Records, func(i, j int) bool {
		if rankResp.Records[i].AC != rankResp.Records[j].AC {
			return rankResp.Records[i].AC > rankResp.Records[j].AC
		}
		return rankResp.Records[i].TotalTime < rankResp.Records[j].TotalTime
	})

	// 设置排名
	for i := range rankResp.Records {
		rankResp.Records[i].Rank = i + 1
	}

	logger.Info("从数据库获取比赛排名成功",
		zap.Int64("contest_id", contestID),
		zap.Int("participants", len(rankResp.Records)))

	// 输出前10名的详细排名信息（用于调试）
	logger.Info("========== 比赛排名详情（前10名）==========")
	for i := 0; i < len(rankResp.Records) && i < 10; i++ {
		record := rankResp.Records[i]
		logger.Info("排名详情",
			zap.Int("rank", record.Rank),
			zap.String("username", record.Username),
			zap.String("uid", record.UID),
			zap.Int("ac_count", record.AC),
			zap.Int64("total_time", record.TotalTime),
			zap.String("total_time_formatted", fmt.Sprintf("%d分%d秒", record.TotalTime/60, record.TotalTime%60)))
	}
	logger.Info("==========================================")

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

	// 构建用户rating信息列表（使用UID进行强关联绑定）
	// 通过UID而不是数组索引来确保每个用户都关联到正确的rating变化
	userInfos := make([]utils.UserRatingInfo, 0, len(rankResp.Records))
	for _, record := range rankResp.Records {
		rating, ok := ratingMap[record.UID]
		if !ok {
			rating = s.config.InitialRating
		}
		// 判断用户是否至少AC了一道题（AC数量 > 0）
		hasSolved := record.AC > 0
		userInfos = append(userInfos, utils.UserRatingInfo{
			UID:       record.UID,
			Rank:      record.Rank, // 使用record.Rank而不是索引，更可靠
			Rating:    rating,
			HasSolved: hasSolved,
		})
	}

	// 输出前10名的 Rating 信息（用于调试）
	logger.Info("========== Rating 计算输入（前10名）==========")
	for i := 0; i < len(userInfos) && i < 10; i++ {
		info := userInfos[i]
		logger.Info("用户 Rating 信息",
			zap.Int("rank", info.Rank),
			zap.String("uid", info.UID),
			zap.Int("old_rating", info.Rating),
			zap.Bool("has_solved", info.HasSolved))
	}
	logger.Info("==========================================")

	// 计算所有rating变化（通过UID进行强关联）
	// 返回 map[UID]ratingChange，确保每个用户都通过UID关联到正确的rating变化
	changesMap := utils.CalculateAllRatingChangesByUID(userInfos, s.config.KFactor)
	logger.Info("rating变化计算完成",
		zap.Int64("contest_id", contestID),
		zap.Int("participants", len(rankResp.Records)),
		zap.Int("changes_count", len(changesMap)))

	// 输出前10名的 Rating 变化（用于调试）
	logger.Info("========== Rating 变化详情（前10名）==========")
	for i := 0; i < len(rankResp.Records) && i < 10; i++ {
		record := rankResp.Records[i]
		oldRating := s.config.InitialRating
		if r, ok := ratingMap[record.UID]; ok {
			oldRating = r
		}
		change := changesMap[record.UID]
		newRating := utils.CalculateNewRating(oldRating, change)
		logger.Info("Rating 变化",
			zap.Int("rank", record.Rank),
			zap.String("username", record.Username),
			zap.Int("old_rating", oldRating),
			zap.Int("rating_change", change),
			zap.Int("new_rating", newRating))
	}
	logger.Info("==========================================")

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
	// 通过UID进行强关联绑定，确保每个用户都得到正确的rating变化
	updateCount := 0
	for _, record := range rankResp.Records {
		// 通过UID获取oldRating和change，确保强关联
		oldRating, ok := ratingMap[record.UID]
		if !ok {
			oldRating = s.config.InitialRating
		}
		
		// 通过UID从changesMap中获取rating变化（强关联绑定）
		change, exists := changesMap[record.UID]
		if !exists {
			logger.Error("找不到用户的rating变化",
				zap.String("uid", record.UID),
				zap.Int("rank", record.Rank))
			tx.Rollback()
			return nil, fmt.Errorf("找不到用户 %s 的rating变化", record.UID)
		}

		// 新手保护：前3场比赛掉分减半
		contestCount := contestCountMap[record.UID]
		originalChange := change
		if contestCount < 3 && change < 0 {
			change = change / 2
			logger.Debug("新手保护生效（掉分减半）",
				zap.String("uid", record.UID),
				zap.Int64("contest_count", contestCount),
				zap.Int("original_change", originalChange),
				zap.Int("protected_change", change))
		}

		// 新人奖励：前4场比赛，如果AC了至少一道题，额外+30分
		if contestCount < 4 && record.AC > 0 {
			change += 30
			logger.Debug("新人奖励生效（AC题目+30分）",
				zap.String("uid", record.UID),
				zap.Int64("contest_count", contestCount),
				zap.Int("ac_count", record.AC),
				zap.Int("bonus", 30),
				zap.Int("final_change", change))
		}

		newRating := utils.CalculateNewRating(oldRating, change)
		rank := record.Rank // 使用record.Rank，确保使用正确的排名

		// 创建 oldRating 的副本，避免指针问题
		oldRatingCopy := oldRating
		history := model.RatingHistory{
			UID:          record.UID,
			ContestID:    uint64(contestID),
			OldRating:    &oldRatingCopy,
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
	status.RatingCalculated = true
	status.CalculatedAt = &calculatedAt
	status.UpdatedAt = now

	// 如果是新记录，设置 CreatedAt
	if status.CreatedAt.IsZero() {
		status.CreatedAt = now
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
		// 如果找不到记录，不允许计算
		return false
	}
	if !status.IsRated {
		return false
	}
	if status.RatingCalculated {
		// 已经计算过，不允许重复计算
		return false
	}
	return true
}

