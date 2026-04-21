package service

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/hoj/hist-oj/internal/client"
	"github.com/hoj/hist-oj/internal/config"
	"github.com/hoj/hist-oj/internal/model"
	"github.com/hoj/hist-oj/internal/utils"
)

type RatingService struct {
	db     *gorm.DB
	config *config.RatingConfig
}

var cfNewbieDisplayPromotions = []int{500, 350, 250, 150, 100, 50}

func NewRatingService(db *gorm.DB, cfg *config.RatingConfig) *RatingService {
	return &RatingService{
		db:     db,
		config: cfg,
	}
}

func getCFNewbieDisplayBonus(completedContests int64) int {
	if completedContests < 0 {
		return 0
	}
	if completedContests >= int64(len(cfNewbieDisplayPromotions)) {
		return 0
	}
	return cfNewbieDisplayPromotions[completedContests]
}

func getCFNewbieRemainingBonus(completedContests int64) int {
	if completedContests < 0 {
		completedContests = 0
	}
	if completedContests >= int64(len(cfNewbieDisplayPromotions)) {
		return 0
	}

	total := 0
	for i := completedContests; i < int64(len(cfNewbieDisplayPromotions)); i++ {
		total += cfNewbieDisplayPromotions[i]
	}
	return total
}

func buildCFNewbieBonusReason(bonus int) string {
	if bonus <= 0 {
		return ""
	}
	return fmt.Sprintf("cf_newbie_bonus:+%d", bonus)
}

// CalculateContestRating 计算比赛的rating变化
func (s *RatingService) CalculateContestRating(contestID int64) ([]model.RatingHistory, error) {
	logger := utils.GetLogger()
	logger.Info("开始计算比赛rating", zap.Int64("contest_id", contestID))

	now := time.Now()

	// 使用事务并在开始时加行锁，防止并发计算
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

	// 使用 SELECT FOR UPDATE 加行锁，防止并发计算
	// 这会锁定该比赛的记录，直到事务提交或回滚
	var status model.ContestRatingStatus
	err := tx.Set("gorm:query_option", "FOR UPDATE").
		Where("contest_id = ?", contestID).
		First(&status).Error

	if err == gorm.ErrRecordNotFound {
		tx.Rollback()
		// 如果找不到记录，说明比赛还没有设置 Rating 类型，不计算
		logger.Info("比赛未设置Rating类型", zap.Int64("contest_id", contestID))
		return nil, nil
	} else if err != nil {
		tx.Rollback()
		logger.Error("查询比赛状态失败", zap.Int64("contest_id", contestID), zap.Error(err))
		return nil, fmt.Errorf("查询比赛状态失败: %w", err)
	}

	// 检查是否已计算过（在锁内再次检查，防止并发）
	if status.RatingCalculated {
		tx.Rollback()
		logger.Warn("比赛rating已计算过（并发检测）", zap.Int64("contest_id", contestID))
		var histories []model.RatingHistory
		s.db.Where("contest_id = ?", contestID).Find(&histories)
		return histories, nil
	}

	// 检查是否为 Rating 比赛
	if !status.IsRated {
		tx.Rollback()
		logger.Info("比赛不是Rating比赛", zap.Int64("contest_id", contestID))
		return nil, nil
	}

	// 立即更新 RatingCalculated = true，防止并发
	// 即使后续计算失败，也需要手动重置状态才能重新计算
	status.RatingCalculated = true
	calculatedAt := now
	status.CalculatedAt = &calculatedAt
	status.UpdatedAt = now
	if err := tx.Save(&status).Error; err != nil {
		tx.Rollback()
		logger.Error("更新比赛rating状态失败",
			zap.Int64("contest_id", contestID),
			zap.Error(err))
		return nil, fmt.Errorf("更新比赛rating状态失败: %w", err)
	}

	logger.Info("获取到计算锁并设置计算标志，开始处理", zap.Int64("contest_id", contestID))

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

	// 【重要】先获取全部比赛提交记录，再基于 skipUIDs 在内存中分流
	// 这样可以保留 Skip 用户的比赛节点（rating_history），便于重算链路与前端展示
	var contestRecords []model.ContestRecord
	query := `
			SELECT DISTINCT cr.*
			FROM contest_record cr
			WHERE cr.cid = ?
				AND cr.submit_time >= ?
				AND cr.submit_time <= ?
		`
	if err := tx.Raw(query, contestID, contestInfo.StartTime, contestInfo.EndTime).
		Scan(&contestRecords).Error; err != nil {
		tx.Rollback()
		logger.Error("从数据库获取比赛记录失败", zap.Int64("contest_id", contestID), zap.Error(err))
		return nil, fmt.Errorf("获取比赛记录失败: %w", err)
	}

	logger.Info("获取到比赛提交记录",
		zap.Int64("contest_id", contestID),
		zap.Int("total_records", len(contestRecords)))

	// 聚合每个用户的比赛成绩
	type UserScore struct {
		UID         string
		Username    string
		ACProblems  map[string]bool      // 记录已AC的题目（按 DisplayID）
		ProblemInfo map[string]*struct { // 每道题的详细信息（按 DisplayID）
			ACTime     uint64 // AC时间
			ErrorCount int    // AC前的错误次数
		}
		TotalTime int64 // 总用时（秒，包含罚时）
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
				TotalTime: 0,
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

	// 设置排名（并列名次保持一致）
	currentRank := 1
	for i := range rankResp.Records {
		if i == 0 {
			rankResp.Records[i].Rank = currentRank
			continue
		}
		prev := rankResp.Records[i-1]
		curr := rankResp.Records[i]
		if curr.AC == prev.AC && curr.TotalTime == prev.TotalTime {
			rankResp.Records[i].Rank = currentRank
		} else {
			currentRank = i + 1
			rankResp.Records[i].Rank = currentRank
		}
	}

	// 【新增】获取skip用户列表（重算过程中 is_applied 可能尚未置位，需全部读取）
	var skipUsers []model.ContestSkipUser
	tx.Where("contest_id = ?", contestID).Find(&skipUsers)

	skipUIDs := make(map[string]bool)
	skipUserInfo := make(map[string]*model.ContestSkipUser) // 存储skip用户信息
	for i := range skipUsers {
		skipUIDs[skipUsers[i].UID] = true
		skipUserInfo[skipUsers[i].UID] = &skipUsers[i]
	}

	logger.Info("获取到skip用户",
		zap.Int64("contest_id", contestID),
		zap.Int("skip_count", len(skipUsers)))

	// 【修改】在计算rating时，过滤掉skip的用户
	// 但需要保留所有记录用于最终展示
	filteredRecords := make([]client.ContestRankRecord, 0)
	skippedRecords := make([]client.ContestRankRecord, 0) // 被skip的记录
	rankUIDSet := make(map[string]bool, len(rankResp.Records))

	for _, record := range rankResp.Records {
		rankUIDSet[record.UID] = true
		if !skipUIDs[record.UID] {
			filteredRecords = append(filteredRecords, record)
		} else {
			skippedRecords = append(skippedRecords, record)
		}
	}

	// 无提交但被 skip 的用户不在 rankResp 中，需要补一个 skip 节点用于个人主页展示
	for _, skipUser := range skipUsers {
		if rankUIDSet[skipUser.UID] {
			continue
		}
		skippedRecords = append(skippedRecords, client.ContestRankRecord{
			UID:      skipUser.UID,
			Username: skipUser.Username,
			Rank:     0, // 无提交，按未排名处理
		})
	}

	// 【重要】使用原始排名计算rating，不要重新排名
	// skip用户只是不计入rating变化，但不应该改变其他用户的原始排名

	logger.Info("从数据库获取比赛排名成功",
		zap.Int64("contest_id", contestID),
		zap.Int64("total_participants", int64(len(rankResp.Records))),
		zap.Int64("valid_participants", int64(len(filteredRecords))),
		zap.Int64("skipped_participants", int64(len(skippedRecords))))

	// 输出前10名的详细排名信息（用于调试）
	logger.Info("========== 比赛排名详情（前10名）==========")
	for i := 0; i < len(rankResp.Records) && i < 10; i++ {
		record := rankResp.Records[i]
		isSkip := skipUIDs[record.UID]
		logger.Info("排名详情",
			zap.Int("rank", record.Rank),
			zap.String("username", record.Username),
			zap.String("uid", record.UID),
			zap.Bool("is_skip", isSkip),
			zap.Int("ac_count", record.AC),
			zap.Int64("total_time", record.TotalTime),
			zap.String("total_time_formatted", fmt.Sprintf("%d分%d秒", record.TotalTime/60, record.TotalTime%60)))
	}
	logger.Info("==========================================")

	if len(filteredRecords) < 3 {
		logger.Warn("有效参赛人数不足",
			zap.Int64("contest_id", contestID),
			zap.Int("participants", len(filteredRecords)))
		tx.Rollback()
		return nil, nil
	}

	// 获取所有需要处理用户（有效参赛者 + skip用户）的当前rating
	uids := make([]string, 0, len(rankResp.Records)+len(skipUsers))
	uidSeen := make(map[string]bool, len(rankResp.Records)+len(skipUsers))
	for _, record := range rankResp.Records {
		if uidSeen[record.UID] {
			continue
		}
		uids = append(uids, record.UID)
		uidSeen[record.UID] = true
	}
	for _, skipUser := range skipUsers {
		if uidSeen[skipUser.UID] {
			continue
		}
		uids = append(uids, skipUser.UID)
		uidSeen[skipUser.UID] = true
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

	// 查询每个用户历史已计分比赛次数（不含手动调整，不含 skip）
	contestCountMap := make(map[string]int64, len(uids))
	for _, uid := range uids {
		contestCountMap[uid] = 0
	}

	type userContestCountRow struct {
		UID   string `gorm:"column:uid"`
		Count int64  `gorm:"column:count"`
	}
	var contestCountRows []userContestCountRow
	if err := tx.Model(&model.RatingHistory{}).
		Select("uid, COUNT(*) as count").
		Where("uid IN ? AND contest_id IS NOT NULL AND is_manual = 0 AND is_skip = 0", uids).
		Group("uid").
		Scan(&contestCountRows).Error; err != nil {
		tx.Rollback()
		logger.Error("查询用户历史比赛次数失败",
			zap.Int64("contest_id", contestID),
			zap.Error(err))
		return nil, fmt.Errorf("查询用户历史比赛次数失败: %w", err)
	}
	for _, row := range contestCountRows {
		contestCountMap[row.UID] = row.Count
	}

	// 为非 skip 用户创建 rating 计算输入
	// 过滤掉 skip 用户后，按有效参赛者重新编号排名（1..N）
	userInfos := make([]utils.UserRatingInfo, 0, len(filteredRecords))
	displayRatingMap := make(map[string]int, len(filteredRecords))
	for i, record := range filteredRecords {
		displayRating, ok := ratingMap[record.UID]
		if !ok {
			displayRating = s.config.InitialRating
		}

		completedContests := contestCountMap[record.UID]

		calculationRating := displayRating + getCFNewbieRemainingBonus(completedContests)
		displayRatingMap[record.UID] = displayRating

		// 判断用户是否至少AC了一道题（AC数量 > 0）
		hasSolved := record.AC > 0
		newRank := i + 1
		userInfos = append(userInfos, utils.UserRatingInfo{
			UID:       record.UID,
			Rank:      newRank, // 重新排名：1, 2, 3, ...
			Rating:    calculationRating,
			HasSolved: hasSolved,
		})
	}

	// 输出前10名的 Rating 信息（用于调试）
	logger.Info("========== Rating 计算输入（前10名，有效参赛者重排后）==========")
	for i := 0; i < len(userInfos) && i < 10; i++ {
		info := userInfos[i]
		logger.Info("用户 Rating 信息",
			zap.Int("rank", info.Rank),
			zap.String("uid", info.UID),
			zap.Int("old_rating", info.Rating),
			zap.Bool("has_solved", info.HasSolved))
	}
	logger.Info("==========================================")

	// 计算所有 rating 变化（CF 规则）
	changesMap := utils.CalculateAllRatingChangesByUID(userInfos, s.config.KFactor)
	logger.Info("rating变化计算完成",
		zap.Int64("contest_id", contestID),
		zap.Int("participants", len(filteredRecords)),
		zap.Int("changes_count", len(changesMap)))

	// 输出前10名的 Rating 变化（用于调试）
	logger.Info("========== Rating 变化详情（前10名）==========")
	for i := 0; i < len(filteredRecords) && i < 10; i++ {
		record := filteredRecords[i]
		oldDisplayRating := s.config.InitialRating
		if r, ok := displayRatingMap[record.UID]; ok {
			oldDisplayRating = r
		}
		baseChange := changesMap[record.UID]
		completedContests := contestCountMap[record.UID]
		newbieBonus := getCFNewbieDisplayBonus(completedContests)
		finalChange := baseChange + newbieBonus
		newDisplayRating := utils.CalculateNewRating(oldDisplayRating, finalChange)
		logger.Info("Rating 变化",
			zap.Int("rank", record.Rank),
			zap.String("username", record.Username),
			zap.Int("old_display_rating", oldDisplayRating),
			zap.Int("base_change", baseChange),
			zap.Int("newbie_bonus", newbieBonus),
			zap.Int("rating_change", finalChange),
			zap.Int("new_display_rating", newDisplayRating))
	}
	logger.Info("==========================================")

	// 保存rating历史记录并更新用户rating
	histories := make([]model.RatingHistory, 0, len(filteredRecords)+len(skippedRecords))

	// 批量更新用户rating
	// 通过UID进行强关联绑定，确保每个用户都得到正确的rating变化
	updateCount := 0
	for _, record := range filteredRecords {
		// 通过UID获取oldRating和change，确保强关联
		oldDisplayRating, ok := displayRatingMap[record.UID]
		if !ok {
			oldDisplayRating = s.config.InitialRating
		}

		// 通过UID从changesMap中获取rating变化（强关联绑定）
		baseChange, exists := changesMap[record.UID]
		if !exists {
			logger.Error("找不到用户的rating变化",
				zap.String("uid", record.UID),
				zap.Int("rank", record.Rank))
			tx.Rollback()
			return nil, fmt.Errorf("找不到用户 %s 的rating变化", record.UID)
		}

		contestCount := contestCountMap[record.UID]
		newbieBonus := getCFNewbieDisplayBonus(contestCount)

		finalDisplayChange := baseChange + newbieBonus
		if newbieBonus > 0 {
			logger.Debug("新手保护加分生效",
				zap.String("uid", record.UID),
				zap.Int64("completed_contests", contestCount),
				zap.Int("base_change", baseChange),
				zap.Int("bonus", newbieBonus),
				zap.Int("final_change", finalDisplayChange))
		}

		newRating := utils.CalculateNewRating(oldDisplayRating, finalDisplayChange)
		rank := record.Rank // 使用record.Rank，确保使用正确的排名

		// 创建 oldRating 的副本，避免指针问题
		oldRatingCopy := oldDisplayRating
		contestIDCopy := uint64(contestID) // 转换为 uint64
		history := model.RatingHistory{
			UID:          record.UID,
			ContestID:    &contestIDCopy, // 使用指针
			OldRating:    &oldRatingCopy,
			NewRating:    newRating,
			RatingChange: finalDisplayChange,
			Rank:         rank,
			Participants: len(filteredRecords), // 只计算实际参与 Rating 的人数（不包括 Skip 用户）
			Reason:       buildCFNewbieBonusReason(newbieBonus),
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

	// 【新增】为skip用户创建特殊的rating历史记录
	if len(skippedRecords) > 0 {
		logger.Info("为skip用户创建rating历史记录",
			zap.Int64("contest_id", contestID),
			zap.Int("skip_count", len(skippedRecords)))

		for _, skipRecord := range skippedRecords {
			// 获取用户当前rating
			oldRating := s.config.InitialRating
			if r, ok := ratingMap[skipRecord.UID]; ok {
				oldRating = r
			}

			skipInfo := skipUserInfo[skipRecord.UID]
			oldRatingCopy := oldRating
			contestIDCopy := uint64(contestID)

			// 创建skip标记记录（rating不变）
			skipHistory := model.RatingHistory{
				UID:          skipRecord.UID,
				ContestID:    &contestIDCopy,
				OldRating:    &oldRatingCopy,
				NewRating:    oldRating,            // rating不变
				RatingChange: 0,                    // 变化为0
				Rank:         skipRecord.Rank,      // 保留原排名
				Participants: len(filteredRecords), // 只计算实际参与 Rating 的人数（不包括 Skip 用户）
				Reason:       skipInfo.Reason,
				IsSkip:       true, // 标记为skip
				SkipReason:   skipInfo.Reason,
				CreatedAt:    now,
			}
			histories = append(histories, skipHistory)
		}

		logger.Info("skip用户rating历史记录创建完成",
			zap.Int64("contest_id", contestID),
			zap.Int("skip_histories", len(skippedRecords)))

		// 更新skip用户的hist_rating字段（修复bug：确保hist_rating与rating_history一致）
		for _, skipRecord := range skippedRecords {
			oldRating := s.config.InitialRating
			if r, ok := ratingMap[skipRecord.UID]; ok {
				oldRating = r
			}

			// 更新hist_rating为重置后的rating（比赛前的rating）
			if err := tx.Model(&model.UserRecord{}).
				Where("uid = ?", skipRecord.UID).
				Update("hist_rating", oldRating).Error; err != nil {
				logger.Error("更新skip用户hist_rating失败",
					zap.String("uid", skipRecord.UID),
					zap.Int("rating", oldRating),
					zap.Error(err))
			} else {
				logger.Debug("更新skip用户hist_rating成功",
					zap.String("uid", skipRecord.UID),
					zap.Int("rating", oldRating))
			}
		}
	}

	// 批量保存历史记录（使用 ON DUPLICATE KEY UPDATE 处理重复）
	// 由于有唯一约束 (uid, contest_id)，重复记录会被忽略而不是报错
	for i := 0; i < len(histories); i += 100 {
		end := i + 100
		if end > len(histories) {
			end = len(histories)
		}
		batch := histories[i:end]

		// 使用 Clauses 处理唯一键冲突
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "uid"}, {Name: "contest_id"}},
			DoNothing: true, // 重复时什么都不做
		}).Create(&batch).Error; err != nil {
			// 如果不是重复键错误，回滚并返回
			if !strings.Contains(err.Error(), "Duplicate entry") {
				tx.Rollback()
				logger.Error("保存rating历史失败",
					zap.Int64("contest_id", contestID),
					zap.Int("batch", i/100+1),
					zap.Error(err))
				return nil, fmt.Errorf("保存rating历史失败: %w", err)
			}
			// 重复键错误，记录警告但继续
			logger.Warn("检测到重复的rating历史记录（已忽略）",
				zap.Int64("contest_id", contestID),
				zap.Int("batch", i/100+1),
				zap.Error(err))
		}
	}

	// 更新计算完成时间
	completedAt := time.Now()
	tx.Model(&model.ContestRatingStatus{}).
		Where("contest_id = ?", contestID).
		Update("calculated_at", completedAt)

	// 回放关联到本场比赛的手动调整（必须在本场比赛rating计算之后）
	appliedManualCount, err := s.applyRelatedManualAdjustments(tx, uint64(contestID))
	if err != nil {
		tx.Rollback()
		logger.Error("回放关联手动调整失败",
			zap.Int64("contest_id", contestID),
			zap.Error(err))
		return nil, fmt.Errorf("回放关联手动调整失败: %w", err)
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
		zap.Int("updated_users", updateCount),
		zap.Int("applied_manual_adjustments", appliedManualCount))

	return histories, nil
}

// applyRelatedManualAdjustments 回放关联到指定比赛的手动调整
// 调用时机：某场比赛rating计算完成后，进入下一场比赛前
func (s *RatingService) applyRelatedManualAdjustments(tx *gorm.DB, contestID uint64) (int, error) {
	logger := utils.GetLogger()

	var adjustments []model.RatingHistory
	if err := tx.Where("is_manual = ? AND related_contest_id = ?", true, contestID).
		Order("created_at ASC, id ASC").
		Find(&adjustments).Error; err != nil {
		return 0, fmt.Errorf("查询关联手动调整失败: %w", err)
	}

	if len(adjustments) == 0 {
		return 0, nil
	}

	appliedCount := 0
	for _, adj := range adjustments {
		oldRating := s.config.InitialRating
		var userRecord model.UserRecord
		err := tx.Where("uid = ?", adj.UID).First(&userRecord).Error
		if err != nil {
			if err != gorm.ErrRecordNotFound {
				return appliedCount, fmt.Errorf("查询用户当前rating失败(uid=%s): %w", adj.UID, err)
			}
		} else if userRecord.HistRating != nil {
			oldRating = *userRecord.HistRating
		}

		newRating := utils.CalculateNewRating(oldRating, adj.RatingChange)

		// 更新 user_record
		if err == gorm.ErrRecordNotFound {
			newRatingCopy := newRating
			createRecord := model.UserRecord{
				UID:        adj.UID,
				HistRating: &newRatingCopy,
			}
			if err := tx.Create(&createRecord).Error; err != nil {
				return appliedCount, fmt.Errorf("创建用户rating记录失败(uid=%s): %w", adj.UID, err)
			}
		} else {
			if err := tx.Model(&model.UserRecord{}).
				Where("uid = ?", adj.UID).
				Update("hist_rating", newRating).Error; err != nil {
				return appliedCount, fmt.Errorf("更新用户rating失败(uid=%s): %w", adj.UID, err)
			}
		}

		// 回写手动调整记录中的 old/new rating，使历史与重算结果一致
		oldRatingCopy := oldRating
		if err := tx.Model(&model.RatingHistory{}).
			Where("id = ?", adj.ID).
			Updates(map[string]interface{}{
				"old_rating": oldRatingCopy,
				"new_rating": newRating,
			}).Error; err != nil {
			return appliedCount, fmt.Errorf("更新手动调整历史失败(id=%d): %w", adj.ID, err)
		}

		appliedCount++
	}

	logger.Info("已回放关联手动调整",
		zap.Uint64("contest_id", contestID),
		zap.Int("count", appliedCount))
	return appliedCount, nil
}

// applyManualAdjustmentsBeforeContest 回放起始比赛之前的关联手动调整，用于重算基线修正
func (s *RatingService) applyManualAdjustmentsBeforeContest(tx *gorm.DB, startContestID uint64, ratingResetMap map[string]int) (int, error) {
	if len(ratingResetMap) == 0 {
		return 0, nil
	}

	var adjustments []model.RatingHistory
	if err := tx.Where("is_manual = ? AND related_contest_id IS NOT NULL AND related_contest_id < ?", true, startContestID).
		Order("related_contest_id ASC, created_at ASC, id ASC").
		Find(&adjustments).Error; err != nil {
		return 0, fmt.Errorf("查询起点前手动调整失败: %w", err)
	}

	applied := 0
	for _, adj := range adjustments {
		currentRating, exists := ratingResetMap[adj.UID]
		if !exists {
			// 非本次重算涉及用户，跳过
			continue
		}
		ratingResetMap[adj.UID] = utils.CalculateNewRating(currentRating, adj.RatingChange)
		applied++
	}

	return applied, nil
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

// AdjustUserRating 手动调整用户rating
// 参数:
//   - username: 用户名
//   - delta: rating变化值（正数=增加，负数=减少）
//   - reason: 操作原因（必填，如"AI作弊"、"账号违规"等）
//   - operatorUID: 操作人UID（管理员）
//
// 返回:
//   - oldRating: 调整前的rating
//   - newRating: 调整后的rating
//   - ratingChange: 实际rating变化
func (s *RatingService) AdjustUserRating(username string, delta int, reason string, relatedContestID *uint64, operatorUID string, operatorUsername string) (int, int, int, error) {
	logger := utils.GetLogger()

	// 参数验证
	if username == "" {
		logger.Error("用户名不能为空")
		return 0, 0, 0, fmt.Errorf("用户名不能为空")
	}
	if reason == "" {
		logger.Error("操作原因不能为空")
		return 0, 0, 0, fmt.Errorf("操作原因不能为空")
	}
	if delta == 0 {
		logger.Error("rating变化值不能为0")
		return 0, 0, 0, fmt.Errorf("rating变化值不能为0")
	}
	if relatedContestID != nil && *relatedContestID == 0 {
		return 0, 0, 0, fmt.Errorf("关联比赛ID必须大于0")
	}

	// 查询用户信息（获取真实UID）
	var userInfo model.UserInfo
	if err := s.db.Where("username = ?", username).First(&userInfo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Error("用户不存在", zap.String("username", username))
			return 0, 0, 0, fmt.Errorf("用户不存在: %s", username)
		}
		logger.Error("查询用户信息失败", zap.String("username", username), zap.Error(err))
		return 0, 0, 0, fmt.Errorf("查询用户信息失败: %w", err)
	}

	// 校验关联比赛（如果指定）
	if relatedContestID != nil {
		var status model.ContestRatingStatus
		if err := s.db.Where("contest_id = ?", *relatedContestID).First(&status).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return 0, 0, 0, fmt.Errorf("关联比赛不存在或未开启Rating")
			}
			return 0, 0, 0, fmt.Errorf("查询关联比赛状态失败: %w", err)
		}
		if !status.IsRated {
			return 0, 0, 0, fmt.Errorf("关联比赛不是Rating比赛")
		}
		if !status.RatingCalculated {
			return 0, 0, 0, fmt.Errorf("关联比赛尚未完成Rating计算，暂不支持绑定")
		}
	}

	// 开启事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			logger.Error("调整rating时发生panic",
				zap.String("username", username),
				zap.Any("panic", r))
		}
	}()

	// 获取用户当前rating
	var userRecord model.UserRecord
	if err := tx.Where("uid = ?", userInfo.UUID).First(&userRecord).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 用户没有rating记录，创建初始记录
			initialRating := s.config.InitialRating
			logger.Info("用户没有rating记录，创建初始记录",
				zap.String("username", username),
				zap.String("uid", userInfo.UUID),
				zap.Int("initial_rating", initialRating))
			userRecord = model.UserRecord{
				UID:        userInfo.UUID,
				HistRating: &initialRating,
			}
			if err := tx.Create(&userRecord).Error; err != nil {
				tx.Rollback()
				logger.Error("创建用户rating记录失败", zap.String("uid", userInfo.UUID), zap.Error(err))
				return 0, 0, 0, fmt.Errorf("创建用户rating记录失败: %w", err)
			}
		} else {
			tx.Rollback()
			logger.Error("查询用户rating失败", zap.String("uid", userInfo.UUID), zap.Error(err))
			return 0, 0, 0, fmt.Errorf("查询用户rating失败: %w", err)
		}
	}

	oldRating := s.config.InitialRating
	if userRecord.HistRating != nil {
		oldRating = *userRecord.HistRating
	}

	// 计算新rating
	ratingChange := delta
	newRating := utils.CalculateNewRating(oldRating, ratingChange)

	// 更新用户rating
	if err := tx.Model(&model.UserRecord{}).
		Where("uid = ?", userInfo.UUID).
		Update("hist_rating", newRating).Error; err != nil {
		tx.Rollback()
		logger.Error("更新用户rating失败",
			zap.String("uid", userInfo.UUID),
			zap.Int("new_rating", newRating),
			zap.Error(err))
		return 0, 0, 0, fmt.Errorf("更新用户rating失败: %w", err)
	}

	// 插入rating历史记录
	// 这里使用 map 明确写入 contest_id = NULL，避免数据库默认值或历史约束导致落成 0，
	// 从而触发 (uid, contest_id) 唯一键冲突（如 Duplicate entry 'uid-0'）。
	now := time.Now()
	oldRatingCopy := oldRating
	historyValues := map[string]interface{}{
		"uid":                userInfo.UUID,
		"contest_id":         nil, // 手动调整不占用比赛节点，必须显式NULL
		"related_contest_id": relatedContestID,
		"old_rating":         oldRatingCopy,
		"new_rating":         newRating,
		"rating_change":      ratingChange,
		"rank":               0, // 手动调整没有排名
		"participants":       0, // 手动调整没有参赛人数
		"reason":             reason,
		"is_manual":          true,
		"operator_uid":       operatorUID,
		"created_at":         now,
	}

	if err := tx.Table("rating_history").Create(historyValues).Error; err != nil {
		tx.Rollback()
		logger.Error("创建rating历史记录失败",
			zap.String("uid", userInfo.UUID),
			zap.Error(err))
		return 0, 0, 0, fmt.Errorf("创建rating历史记录失败: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		logger.Error("提交事务失败",
			zap.String("username", username),
			zap.Error(err))
		return 0, 0, 0, fmt.Errorf("提交事务失败: %w", err)
	}

	logger.Info("手动调整用户rating成功",
		zap.String("username", username),
		zap.String("uid", userInfo.UUID),
		zap.Int("old_rating", oldRating),
		zap.Int("new_rating", newRating),
		zap.Int("rating_change", ratingChange),
		zap.String("reason", reason),
		zap.String("operator_uid", operatorUID))

	// 记录操作日志
	detailJSON, _ := json.Marshal(map[string]interface{}{
		"username":         username,
		"oldRating":        oldRating,
		"newRating":        newRating,
		"ratingChange":     ratingChange,
		"reason":           reason,
		"relatedContestId": relatedContestID,
	})
	s.LogOperation(operatorUID, operatorUsername, "personal_adjust", "user", username, string(detailJSON), "")

	return oldRating, newRating, ratingChange, nil
}

// CancelManualAdjustment 撤销一条手动调整（管理员）
// 实现方式：删除原手动调整记录，并直接回滚用户当前 hist_rating（不新增反向历史记录）
func (s *RatingService) CancelManualAdjustment(adjustmentID uint64, operatorUID string, operatorUsername string) (map[string]interface{}, error) {
	logger := utils.GetLogger()
	if adjustmentID == 0 {
		return nil, fmt.Errorf("调整记录ID不能为空")
	}

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			logger.Error("撤销手动调整时发生panic",
				zap.Uint64("adjustment_id", adjustmentID),
				zap.Any("panic", r))
		}
	}()

	// 1. 查询并锁定原调整记录
	var adjustment model.RatingHistory
	if err := tx.Set("gorm:query_option", "FOR UPDATE").
		Where("id = ? AND is_manual = ?", adjustmentID, true).
		First(&adjustment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("手动调整记录不存在")
		}
		return nil, fmt.Errorf("查询手动调整记录失败: %w", err)
	}
	if adjustment.RatingChange == 0 {
		return nil, fmt.Errorf("该记录变化值为0，无需撤销")
	}

	// 2. 获取用户信息
	var userInfo model.UserInfo
	if err := tx.Where("uuid = ?", adjustment.UID).First(&userInfo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("被调整用户不存在")
		}
		return nil, fmt.Errorf("查询被调整用户失败: %w", err)
	}

	// 3. 锁定并回滚当前用户 rating（不新增 rating_history 记录）
	var userRecord model.UserRecord
	if err := tx.Set("gorm:query_option", "FOR UPDATE").
		Where("uid = ?", adjustment.UID).
		First(&userRecord).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("用户rating记录不存在")
		}
		return nil, fmt.Errorf("查询用户rating失败: %w", err)
	}

	oldRating := s.config.InitialRating
	if userRecord.HistRating != nil {
		oldRating = *userRecord.HistRating
	}

	reverseDelta := -adjustment.RatingChange
	newRating := utils.CalculateNewRating(oldRating, reverseDelta)
	if err := tx.Model(&model.UserRecord{}).
		Where("uid = ?", adjustment.UID).
		Update("hist_rating", newRating).Error; err != nil {
		return nil, fmt.Errorf("回滚用户rating失败: %w", err)
	}

	// 4. 删除原手动调整记录
	if err := tx.Delete(&model.RatingHistory{}, adjustmentID).Error; err != nil {
		return nil, fmt.Errorf("删除手动调整记录失败: %w", err)
	}

	// 5. 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("提交撤销事务失败: %w", err)
	}

	// 6. 记录“撤销操作”日志（用于审计）
	targetID := fmt.Sprintf("%d", adjustmentID)
	detailJSON, _ := json.Marshal(map[string]interface{}{
		"adjustmentId":     adjustmentID,
		"uid":              adjustment.UID,
		"username":         userInfo.Username,
		"originalChange":   adjustment.RatingChange,
		"reverseChange":    reverseDelta,
		"oldRating":        oldRating,
		"newRating":        newRating,
		"relatedContestId": adjustment.RelatedContestID,
		"deleted":          true,
	})
	s.LogOperation(operatorUID, operatorUsername, "cancel_personal_adjust", "manual_adjustment", targetID, string(detailJSON), "")

	logger.Info("撤销手动调整成功",
		zap.Uint64("adjustment_id", adjustmentID),
		zap.String("username", userInfo.Username),
		zap.Int("reverse_delta", reverseDelta),
		zap.Int("old_rating", oldRating),
		zap.Int("new_rating", newRating))

	return map[string]interface{}{
		"adjustmentId":     adjustmentID,
		"username":         userInfo.Username,
		"oldRating":        oldRating,
		"newRating":        newRating,
		"ratingChange":     reverseDelta,
		"relatedContestId": adjustment.RelatedContestID,
		"deleted":          true,
	}, nil
}

// SkipResult Skip操作结果
type SkipResult struct {
	SuccessUsers       []string `json:"successUsers"`
	FailedUsers        []string `json:"failedUsers"`
	DuplicatedUsers    []string `json:"duplicatedUsers"`
	PendingRecalculate bool     `json:"pendingRecalculate"`
	TaskID             uint64   `json:"taskId"`
	NextContestIDs     []uint64 `json:"nextContestIds"`
}

// BatchSkipContestUsers 批量Skip用户（支持延迟重算）
func (s *RatingService) BatchSkipContestUsers(
	contestID int64,
	usernames []string,
	reason string,
	operatorUID string,
	operatorUsername string,
	autoRecalc bool,
) (*SkipResult, error) {

	logger := utils.GetLogger()
	logger.Info("开始批量skip用户",
		zap.Int64("contest_id", contestID),
		zap.Int("user_count", len(usernames)),
		zap.Bool("auto_recalc", autoRecalc))

	// 使用事务处理
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			logger.Error("批量skip时发生panic", zap.Any("panic", r))
		}
	}()

	// 1. 加行锁防止并发
	var status model.ContestRatingStatus
	err := tx.Set("gorm:query_option", "FOR UPDATE").
		Where("contest_id = ?", contestID).
		First(&status).Error

	if err == gorm.ErrRecordNotFound {
		tx.Rollback()
		return nil, fmt.Errorf("比赛不存在")
	} else if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("查询比赛状态失败: %w", err)
	}

	// 2. 检查是否有正在进行的重算
	if status.RecalculateLock {
		tx.Rollback()
		return nil, fmt.Errorf("比赛正在进行rating重算，请稍后再试")
	}

	// 3. 批量查询用户信息
	var users []model.UserInfo
	tx.Where("username IN ?", usernames).Find(&users)

	usernameToUID := make(map[string]string)
	uidToUsername := make(map[string]string)
	for _, user := range users {
		usernameToUID[user.Username] = user.UUID
		uidToUsername[user.UUID] = user.Username
	}

	result := &SkipResult{
		SuccessUsers:    make([]string, 0),
		FailedUsers:     make([]string, 0),
		DuplicatedUsers: make([]string, 0),
	}

	// 5. 批量插入skip用户（使用GORM的OnConflict处理重复）
	for _, username := range usernames {
		uid, ok := usernameToUID[username]
		if !ok {
			result.FailedUsers = append(result.FailedUsers, username)
			continue
		}

		skipUser := model.ContestSkipUser{
			ContestID:        uint64(contestID),
			UID:              uid,
			Username:         username,
			Reason:           reason,
			OperatorUID:      operatorUID,
			OperatorUsername: operatorUsername,
			IsApplied:        false,
		}

		// 使用GORM的Create插入，OnConflict处理重复
		err := tx.Create(&skipUser).Error

		if err == nil {
			// 插入成功
			result.SuccessUsers = append(result.SuccessUsers, username)
		} else {
			// 检查是否是重复键错误（MySQL错误码1062）
			if strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "1062") {
				// 重复，检查是否已存在
				var existingUser model.ContestSkipUser
				errCheck := tx.Where("contest_id = ? AND uid = ?", uint64(contestID), uid).First(&existingUser).Error
				if errCheck == nil {
					result.DuplicatedUsers = append(result.DuplicatedUsers, username)
				} else {
					result.FailedUsers = append(result.FailedUsers, username)
				}
			} else {
				// 其他错误
				logger.Error("插入skip用户失败",
					zap.String("username", username),
					zap.Error(err))
				result.FailedUsers = append(result.FailedUsers, username)
			}
		}
	}

	// 6. 更新比赛状态
	if len(result.SuccessUsers) > 0 {
		now := time.Now()
		tx.Model(&model.ContestRatingStatus{}).
			Where("contest_id = ?", contestID).
			Updates(map[string]interface{}{
				"has_pending_skip":     true,
				"skip_count":           status.SkipCount + len(result.SuccessUsers),
				"skip_data_changed_at": &now,
			})
	}

	tx.Commit()

	logger.Info("批量skip完成",
		zap.Int("success", len(result.SuccessUsers)),
		zap.Int("failed", len(result.FailedUsers)),
		zap.Int("duplicated", len(result.DuplicatedUsers)))

	// 7. 如果选择自动重算，创建重算任务
	if autoRecalc && len(result.SuccessUsers) > 0 {
		taskID, err := s.CreateRecalculateTask(contestID, operatorUID)
		if err != nil {
			logger.Error("创建重算任务失败", zap.Error(err))
			result.PendingRecalculate = false
		} else {
			result.PendingRecalculate = true
			result.TaskID = taskID
			// 异步执行重算
			go s.ExecuteRecalculateTask(taskID)
		}
	} else {
		result.PendingRecalculate = len(result.SuccessUsers) > 0
	}

	// 8. 查询需要级联重算的比赛
	if result.PendingRecalculate {
		var contests []model.Contest
		s.db.Where("id > ? AND is_rating = 1 AND rating_calculated = 1", contestID).
			Order("id ASC").
			Find(&contests)

		result.NextContestIDs = make([]uint64, len(contests))
		for i, c := range contests {
			result.NextContestIDs[i] = c.ID
		}
	}

	// 9. 记录操作日志
	detailJSON, _ := json.Marshal(map[string]interface{}{
		"contestId":  contestID,
		"usernames":  result.SuccessUsers,
		"reason":     reason,
		"autoRecalc": autoRecalc,
	})
	s.LogOperation(operatorUID, operatorUsername, "skip_user", "contest", fmt.Sprintf("%d", contestID), string(detailJSON), "")

	return result, nil
}

// GetContestSkipUsers 获取比赛的Skip用户列表
func (s *RatingService) GetContestSkipUsers(contestID int64) ([]model.ContestSkipUser, error) {
	var skipUsers []model.ContestSkipUser

	// 使用 GORM 的模型查询，自动应用 JSON 标签（camelCase）
	err := s.db.Where("contest_id = ?", contestID).
		Order("created_at DESC").
		Find(&skipUsers).Error

	return skipUsers, err
}

// CancelSkip 取消用户的skip标记
func (s *RatingService) CancelSkip(contestID int64, uids []string, operatorUID string, operatorUsername string) error {
	logger := utils.GetLogger()
	logger.Info("取消skip",
		zap.Int64("contest_id", contestID),
		zap.Int("uid_count", len(uids)))

	tx := s.db.Begin()

	// 1. 加锁
	var status model.ContestRatingStatus
	err := tx.Set("gorm:query_option", "FOR UPDATE").
		Where("contest_id = ?", contestID).
		First(&status).Error

	if err != nil {
		tx.Rollback()
		return fmt.Errorf("查询比赛状态失败: %w", err)
	}

	// 2. 检查重算锁
	if status.RecalculateLock {
		tx.Rollback()
		return fmt.Errorf("比赛正在进行rating重算，无法取消skip")
	}

	// 3. 删除skip记录
	result := tx.Where("contest_id = ? AND uid IN ?", contestID, uids).
		Delete(&model.ContestSkipUser{})

	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}

	// 4. 更新skip计数和最后修改时间
	deletedCount := int(result.RowsAffected)
	now := time.Now()
	tx.Model(&model.ContestRatingStatus{}).
		Where("contest_id = ?", contestID).
		Updates(map[string]interface{}{
			"skip_count":           gorm.Expr("GREATEST(skip_count - ?, 0)", deletedCount),
			"skip_data_changed_at": &now,
		})

	// 5. 检查是否还有其他skip用户
	var count int64
	tx.Model(&model.ContestSkipUser{}).
		Where("contest_id = ?", contestID).
		Count(&count)

	if count == 0 {
		// 没有skip用户了，清除pending标志
		tx.Model(&model.ContestRatingStatus{}).
			Where("contest_id = ?", contestID).
			Update("has_pending_skip", false)
	}

	tx.Commit()

	logger.Info("取消skip成功", zap.Int("deleted_count", deletedCount))

	// 记录操作日志
	detailJSON, _ := json.Marshal(map[string]interface{}{
		"contestId": contestID,
		"uids":      uids,
		"count":     deletedCount,
	})
	s.LogOperation(operatorUID, operatorUsername, "cancel_skip", "contest", fmt.Sprintf("%d", contestID), string(detailJSON), "")

	return nil
}

// CreateRecalculateTask 创建重算任务
func (s *RatingService) CreateRecalculateTask(contestID int64, operatorUID string) (uint64, error) {
	logger := utils.GetLogger()

	// 1. 【重要】检查并处理已存在的重算任务，避免冲突
	// 查找所有 pending 或 running 状态的重算任务
	var existingTasks []model.RatingRecalculateQueue
	s.db.Where("status IN ?", []string{"pending", "running"}).
		Find(&existingTasks)

	var deletedTaskIDs []uint64
	var coveringTask *model.RatingRecalculateQueue

	for _, task := range existingTasks {
		if int64(task.ContestID) <= contestID {
			// 已存在任务的重算范围包含或早于新任务（会覆盖新任务范围）
			// 例如：已存在任务从比赛100开始重算，新任务要从比赛105开始
			// 此时已存在任务已经会重算比赛105及后续，新任务是多余的
			coveringTask = &task
			break
		} else if int64(task.ContestID) > contestID {
			// 已存在任务的重算范围在新任务之后（会被新任务覆盖）
			// 例如：已存在任务从比赛110开始重算，新任务要从比赛100开始
			// 此时新任务会重算比赛100及后续（包含110），已存在任务应该被删除
			deletedTaskIDs = append(deletedTaskIDs, task.ID)
		}
	}

	// 删除会被新任务覆盖的任务
	if len(deletedTaskIDs) > 0 {
		s.db.Delete(&model.RatingRecalculateQueue{}, deletedTaskIDs)
		logger.Info("删除被新任务覆盖的重算任务",
			zap.Int64s("deleted_task_ids", uint64SliceToInt64(deletedTaskIDs)),
			zap.Int64("new_contest_id", contestID))
	}

	// 如果已有任务会覆盖新任务范围，拒绝创建新任务
	if coveringTask != nil {
		logger.Info("拒绝创建新重算任务，已存在覆盖范围的任务",
			zap.Uint64("existing_task_id", coveringTask.ID),
			zap.Uint64("existing_contest_id", coveringTask.ContestID),
			zap.Int64("requested_contest_id", contestID))
		return 0, fmt.Errorf("已存在从比赛 %d 开始的重算任务（任务ID: %d），无需重复创建",
			coveringTask.ContestID, coveringTask.ID)
	}

	// 2. 查询需要重算的比赛
	var contests []model.Contest
	s.db.Where("id >= ? AND is_rating = 1 AND status = 1", contestID).
		Order("id ASC").
		Find(&contests)

	if len(contests) == 0 {
		return 0, fmt.Errorf("没有需要重算的比赛")
	}

	// 3. 创建重算任务
	task := model.RatingRecalculateQueue{
		ContestID:         uint64(contestID),
		Status:            "pending",
		TotalContests:     len(contests),
		ProcessedContests: 0,
		CreatedBy:         operatorUID,
	}

	if err := s.db.Create(&task).Error; err != nil {
		return 0, err
	}

	logger.Info("创建重算任务",
		zap.Uint64("task_id", task.ID),
		zap.Int("total_contests", len(contests)))

	return task.ID, nil
}

// 辅助函数：uint64切片转int64切片
func uint64SliceToInt64(slice []uint64) []int64 {
	result := make([]int64, len(slice))
	for i, v := range slice {
		result[i] = int64(v)
	}
	return result
}

// ExecuteRecalculateTask 执行重算任务（异步）
func (s *RatingService) ExecuteRecalculateTask(taskID uint64) error {
	logger := utils.GetLogger()

	// 1. 加载任务
	var task model.RatingRecalculateQueue
	if err := s.db.First(&task, taskID).Error; err != nil {
		// 任务可能已被删除（被新的重算任务覆盖）
		logger.Warn("重算任务不存在，可能已被删除", zap.Uint64("task_id", taskID))
		return nil // 返回nil而不是error，因为这是正常情况
	}

	// 2. 检查任务状态，防止重复执行
	if task.Status != "pending" {
		logger.Warn("重算任务状态已改变，跳过执行",
			zap.Uint64("task_id", taskID),
			zap.String("status", task.Status))
		return nil
	}

	// 3. 更新状态为运行中
	now := time.Now()
	result := s.db.Model(&task).Updates(map[string]interface{}{
		"status":     "running",
		"started_at": &now,
	})

	// 检查更新是否成功（防止并发更新）
	if result.RowsAffected == 0 {
		logger.Warn("更新任务状态失败，可能有其他进程正在处理", zap.Uint64("task_id", taskID))
		return nil
	}

	logger.Info("开始执行重算任务",
		zap.Uint64("task_id", taskID),
		zap.Uint64("contest_id", task.ContestID))

	// 3. 加锁比赛状态
	tx := s.db.Begin()

	// 设置重算锁
	tx.Model(&model.ContestRatingStatus{}).
		Where("contest_id >= ?", task.ContestID).
		Update("recalculate_lock", true)

	// 4. 获取需要重算的比赛
	var contests []model.Contest
	tx.Where("id >= ? AND is_rating = 1", task.ContestID).
		Order("id ASC").
		Find(&contests)

	logger.Info("需要重算的比赛",
		zap.Uint64("task_id", taskID),
		zap.Int("contest_count", len(contests)))

	if len(contests) == 0 {
		tx.Rollback()
		logger.Warn("没有需要重算的比赛", zap.Uint64("contest_id", task.ContestID))
		return fmt.Errorf("没有需要重算的比赛")
	}

	// 5. 【关键修复】清空旧rating记录并重置用户rating

	// 5.0 准备比赛ID列表
	contestIDs := make([]uint64, len(contests))
	for i, contest := range contests {
		contestIDs[i] = contest.ID
	}

	// 5.1 【重要】在删除rating_history之前，先找出所有在待重算比赛中有rating的用户
	// 这些用户是需要重置rating的
	type UserInRecalculatingContests struct {
		UID string
	}
	var usersInContests []UserInRecalculatingContests
	tx.Raw(`
			SELECT DISTINCT uid
			FROM (
				-- 1) 比赛rating历史中的用户
				SELECT uid
				FROM rating_history
				WHERE contest_id IN ?
					AND is_manual = 0
				UNION
				-- 2) 被 skip 的用户（某些历史版本中可能没有生成比赛节点）
				SELECT uid
				FROM contest_skip_users
				WHERE contest_id IN ?
				UNION
				-- 3) 关联到这些比赛的手动调整用户（防止手调被重复应用）
				SELECT uid
				FROM rating_history
				WHERE is_manual = 1
					AND related_contest_id IN ?
			) t
		`, contestIDs, contestIDs, contestIDs).Scan(&usersInContests)

	logger.Info("找到待重算比赛中的有rating的用户",
		zap.Uint64("task_id", taskID),
		zap.Int("user_count", len(usersInContests)))

	// 5.2 删除所有待重算比赛的 rating_history 记录

	logger.Info("准备删除rating历史记录",
		zap.Uint64("task_id", taskID),
		zap.Int("contest_count", len(contestIDs)),
		zap.Any("contest_ids", contestIDs))

	deleteResult := tx.Where("contest_id IN ?", contestIDs).Delete(&model.RatingHistory{})
	if deleteResult.Error != nil {
		tx.Rollback()
		logger.Error("删除旧rating历史记录失败", zap.Error(deleteResult.Error))
		return fmt.Errorf("删除旧rating历史记录失败: %w", deleteResult.Error)
	}

	logger.Info("删除旧rating历史记录完成",
		zap.Uint64("task_id", taskID),
		zap.Int64("deleted_count", deleteResult.RowsAffected),
		zap.Int("contest_count", len(contestIDs)))

	// 5.3 查找起始比赛之前的最后一个rating，用于重置用户rating
	// 关键修复：取每个用户在起始比赛之前的**最后一条记录**（按contest_id降序，取第一条）
	// 而不是取最高rating，因为rating应该是按照时间顺序递增的
	type LastRatingBeforeStart struct {
		UID       string
		NewRating int
	}

	// 使用子查询找到每个用户在起始比赛之前的最后一条记录
	var lastRatings []LastRatingBeforeStart
	if len(usersInContests) > 0 {
		tx.Raw(`
				SELECT r1.uid, r1.new_rating
				FROM rating_history r1
				WHERE r1.contest_id = (
					SELECT MAX(r2.contest_id)
					FROM rating_history r2
					WHERE r2.uid = r1.uid
						AND r2.is_manual = 0
						AND r2.contest_id > 0
						AND r2.contest_id < ?
				)
				AND r1.is_manual = 0
				AND r1.contest_id > 0
		`, task.ContestID).Scan(&lastRatings)
	}

	// 5.4 构建用户rating映射（包含初始rating）
	ratingResetMap := make(map[string]int)

	// 先填充有历史rating的用户
	for _, lr := range lastRatings {
		ratingResetMap[lr.UID] = lr.NewRating
	}

	// 对于没有历史rating的用户（首次参赛），使用初始rating
	for _, user := range usersInContests {
		if _, exists := ratingResetMap[user.UID]; !exists {
			ratingResetMap[user.UID] = s.config.InitialRating
		}
	}

	// 5.4.1 回放起始比赛之前的关联手动调整，确保重算基线正确
	appliedPreStartManual, err := s.applyManualAdjustmentsBeforeContest(tx, task.ContestID, ratingResetMap)
	if err != nil {
		tx.Rollback()
		logger.Error("回放起点前手动调整失败", zap.Error(err))
		return fmt.Errorf("回放起点前手动调整失败: %w", err)
	}

	logger.Info("构建用户rating映射完成",
		zap.Uint64("task_id", taskID),
		zap.Int("total_users", len(ratingResetMap)),
		zap.Int("new_users", len(usersInContests)-len(lastRatings)),
		zap.Int("manual_adjustments_applied", appliedPreStartManual),
		zap.Int("initial_rating", s.config.InitialRating))

	// 5.5 重置所有用户的 hist_rating
	resetCount := 0
	for uid, rating := range ratingResetMap {
		result := tx.Model(&model.UserRecord{}).
			Where("uid = ?", uid).
			Update("hist_rating", rating)
		if result.Error == nil && result.RowsAffected > 0 {
			resetCount++
		}
	}

	logger.Info("重置用户rating完成",
		zap.Uint64("task_id", taskID),
		zap.Int("reset_count", resetCount))

	// 5.6 重置所有比赛的rating_calculated状态
	for _, contest := range contests {
		tx.Model(&model.ContestRatingStatus{}).
			Where("contest_id = ?", contest.ID).
			Updates(map[string]interface{}{
				"rating_calculated":  false,
				"has_pending_skip":   false,
				"recalculate_status": "calculating",
			})
	}

	tx.Commit()

	// 6. 逐个重算比赛
	for i, contest := range contests {
		logger.Info("重算比赛rating",
			zap.Uint64("task_id", taskID),
			zap.Int("progress", i+1),
			zap.Int("total", len(contests)),
			zap.Uint64("contest_id", contest.ID),
			zap.String("title", contest.Title))

		if _, err := s.CalculateContestRating(int64(contest.ID)); err != nil {
			logger.Error("重算失败",
				zap.Uint64("contest_id", contest.ID),
				zap.Error(err))

			// 更新任务状态为失败
			s.db.Model(&task).Updates(map[string]interface{}{
				"status":        "failed",
				"error_message": err.Error(),
			})

			// 释放锁
			s.db.Model(&model.ContestRatingStatus{}).
				Where("contest_id >= ?", task.ContestID).
				Updates(map[string]interface{}{
					"recalculate_lock":   false,
					"recalculate_status": "failed",
				})

			return err
		}

		// 更新比赛状态
		s.db.Model(&model.ContestRatingStatus{}).
			Where("contest_id = ?", contest.ID).
			Update("recalculate_status", "completed")

		// 更新进度
		s.db.Model(&task).Update("processed_contests", i+1)
	}

	// 7. 重算完成，释放锁
	completedAt := time.Now()
	s.db.Model(&model.ContestRatingStatus{}).
		Where("contest_id >= ?", task.ContestID).
		Updates(map[string]interface{}{
			"recalculate_lock":    false,
			"recalculate_status":  "completed",
			"last_recalculate_at": &completedAt,
		})

	// 8. 标记skip记录为已应用
	logger.Info("标记skip记录为已应用",
		zap.Uint64("task_id", taskID),
		zap.Uint64("contest_id", task.ContestID))

	updateResult := s.db.Model(&model.ContestSkipUser{}).
		Where("contest_id >= ?", task.ContestID).
		Update("is_applied", true)

	if updateResult.Error != nil {
		logger.Error("标记skip记录为已应用失败", zap.Error(updateResult.Error))
	} else {
		logger.Info("标记skip记录为已应用成功",
			zap.Uint64("task_id", taskID),
			zap.Int64("affected_rows", updateResult.RowsAffected))
	}

	// 9. 更新任务状态
	s.db.Model(&task).Updates(map[string]interface{}{
		"status":       "completed",
		"completed_at": &completedAt,
	})

	logger.Info("重算任务完成",
		zap.Uint64("task_id", taskID),
		zap.Int("total_contests", len(contests)))

	// 获取操作人用户名
	var operator model.UserInfo
	operatorUsername := task.CreatedBy
	s.db.Where("uuid = ?", task.CreatedBy).First(&operator)
	if operator.Username != "" {
		operatorUsername = operator.Username
	}

	// 记录操作日志
	detailJSON, _ := json.Marshal(map[string]interface{}{
		"taskId":       taskID,
		"contestId":    task.ContestID,
		"contestCount": len(contests),
		"contestIds":   contestIDs,
	})
	s.LogOperation(task.CreatedBy, operatorUsername, "recalculate", "contest", fmt.Sprintf("%d", task.ContestID), string(detailJSON), "")

	return nil
}

// GetRecalculateProgress 获取重算进度
func (s *RatingService) GetRecalculateProgress(taskID uint64) (map[string]interface{}, error) {
	var task model.RatingRecalculateQueue
	if err := s.db.First(&task, taskID).Error; err != nil {
		return nil, err
	}

	// 获取比赛列表
	var contests []model.Contest
	s.db.Where("id >= ? AND is_rating = 1", task.ContestID).
		Order("id ASC").
		Find(&contests)

	// 构建比赛状态列表
	contestStatusList := make([]map[string]interface{}, 0)
	for i, contest := range contests {
		status := "pending"
		if i < task.ProcessedContests {
			status = "completed"
		} else if i == task.ProcessedContests && task.Status == "running" {
			status = "calculating"
		}

		contestStatusList = append(contestStatusList, map[string]interface{}{
			"id":     contest.ID,
			"title":  contest.Title,
			"status": status,
		})
	}

	return map[string]interface{}{
		"taskId":            task.ID,
		"status":            task.Status,
		"totalContests":     task.TotalContests,
		"processedContests": task.ProcessedContests,
		"contests":          contestStatusList,
		"errorMessage":      task.ErrorMessage,
		"createdAt":         task.CreatedAt,
		"startedAt":         task.StartedAt,
		"completedAt":       task.CompletedAt,
	}, nil
}

// HasRunningRecalculateTask 检查是否有正在运行的重算任务
func (s *RatingService) HasRunningRecalculateTask(contestID int64) (bool, error) {
	var count int64
	err := s.db.Model(&model.RatingRecalculateQueue{}).
		Where("status = ?", "running").
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// ResetRecalculateLock 重置比赛重算锁
func (s *RatingService) ResetRecalculateLock(contestID int64, operatorUID string) error {
	logger := utils.GetLogger()

	// 使用事务确保数据一致性
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 获取当前状态
		var status model.ContestRatingStatus
		if err := tx.Where("contest_id = ?", contestID).First(&status).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// 如果记录不存在，创建新记录
				status = model.ContestRatingStatus{
					ContestID:         uint64(contestID),
					RecalculateLock:   false,
					RecalculateStatus: "none",
				}
				if err := tx.Create(&status).Error; err != nil {
					logger.Error("创建contest_rating_status记录失败",
						zap.Int64("contest_id", contestID),
						zap.Error(err))
					return err
				}
			} else {
				logger.Error("查询contest_rating_status失败",
					zap.Int64("contest_id", contestID),
					zap.Error(err))
				return err
			}
		}

		// 重置锁和状态
		if err := tx.Model(&status).
			Updates(map[string]interface{}{
				"recalculate_lock":   false,
				"recalculate_status": "none",
			}).Error; err != nil {
			logger.Error("更新recalculate_lock失败",
				zap.Int64("contest_id", contestID),
				zap.Error(err))
			return err
		}

		logger.Info("重置重算锁成功",
			zap.Int64("contest_id", contestID),
			zap.String("operator_uid", operatorUID))

		return nil
	})
}

// SyncContestSkipFlag 同步比赛的 Skip 标记到 rating_history 表
// 将 contest_skip_users 表中 is_applied=1 的用户同步到 rating_history 表的 is_skip 字段
func (s *RatingService) SyncContestSkipFlag(contestID int64) (int, error) {
	logger := utils.GetLogger()

	// 1. 获取比赛的 Skip 用户列表
	var skipUsers []model.ContestSkipUser
	if err := s.db.Where("contest_id = ? AND is_applied = ?", contestID, true).
		Find(&skipUsers).Error; err != nil {
		logger.Error("查询Skip用户失败",
			zap.Int64("contest_id", contestID),
			zap.Error(err))
		return 0, fmt.Errorf("查询Skip用户失败: %w", err)
	}

	if len(skipUsers) == 0 {
		logger.Info("没有Skip用户，无需同步", zap.Int64("contest_id", contestID))
		return 0, nil
	}

	// 2. 构建 UID 映射
	skipUIDMap := make(map[string]string) // uid -> reason
	for _, skipUser := range skipUsers {
		skipUIDMap[skipUser.UID] = skipUser.Reason
	}

	// 3. 批量更新 rating_history 表
	// 将所有 Skip 用户的 is_skip 设置为 1，并设置 skip_reason
	updatedCount := 0
	for uid, reason := range skipUIDMap {
		result := s.db.Model(&model.RatingHistory{}).
			Where("contest_id = ? AND uid = ?", contestID, uid).
			Updates(map[string]interface{}{
				"is_skip":     true,
				"skip_reason": reason,
			})

		if result.Error != nil {
			logger.Warn("更新Skip标记失败",
				zap.Int64("contest_id", contestID),
				zap.String("uid", uid),
				zap.Error(result.Error))
			continue
		}

		if result.RowsAffected > 0 {
			updatedCount++
		}
	}

	// 4. 将非 Skip 用户的 is_skip 设置为 0（清理旧数据）
	result := s.db.Model(&model.RatingHistory{}).
		Where("contest_id = ? AND is_skip = ?", contestID, true).
		Not("uid", func() []string {
			uids := make([]string, 0, len(skipUIDMap))
			for uid := range skipUIDMap {
				uids = append(uids, uid)
			}
			return uids
		}()).
		Update("is_skip", false)

	if result.Error != nil {
		logger.Warn("清理非Skip用户标记失败",
			zap.Int64("contest_id", contestID),
			zap.Error(result.Error))
	}

	logger.Info("同步Skip标记完成",
		zap.Int64("contest_id", contestID),
		zap.Int("total_skip_users", len(skipUsers)),
		zap.Int("updated_count", updatedCount),
		zap.Int64("cleaned_count", result.RowsAffected))

	return updatedCount, nil
}

// LogOperation 记录操作日志
func (s *RatingService) LogOperation(operatorUID, operatorUsername, operationType, targetType, targetID, operationDetail, ip string) error {
	logger := utils.GetLogger()

	log := model.RatingOperationLog{
		OperatorUID:      operatorUID,
		OperatorUsername: operatorUsername,
		OperationType:    operationType,
		TargetType:       targetType,
		TargetID:         targetID,
		OperationDetail:  operationDetail,
		IP:               ip,
	}

	if err := s.db.Create(&log).Error; err != nil {
		logger.Error("记录操作日志失败",
			zap.String("operation_type", operationType),
			zap.String("operator_uid", operatorUID),
			zap.Error(err))
		return err
	}

	return nil
}

// GetOperationLogs 获取操作日志
func (s *RatingService) GetOperationLogs(page, limit int, operationType, timeRange string) ([]model.RatingOperationLog, int64, error) {
	logger := utils.GetLogger()

	query := s.db.Model(&model.RatingOperationLog{})

	// 按操作类型筛选
	if operationType != "" {
		query = query.Where("operation_type = ?", operationType)
	}

	// 按时间范围筛选
	if timeRange != "" && timeRange != "all" {
		var startTime time.Time
		now := time.Now()

		switch timeRange {
		case "7d":
			startTime = now.AddDate(0, 0, -7)
		case "30d":
			startTime = now.AddDate(0, 0, -30)
		default:
			startTime = now.AddDate(0, 0, -7) // 默认7天
		}

		query = query.Where("created_at >= ?", startTime)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		logger.Error("查询操作日志总数失败", zap.Error(err))
		return nil, 0, err
	}

	// 分页查询
	var logs []model.RatingOperationLog
	if err := query.Order("created_at DESC").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&logs).Error; err != nil {
		logger.Error("查询操作日志失败", zap.Error(err))
		return nil, 0, err
	}

	logger.Info("查询操作日志成功",
		zap.Int("page", page),
		zap.Int("limit", limit),
		zap.String("operation_type", operationType),
		zap.String("time_range", timeRange),
		zap.Int64("total", total),
		zap.Int("count", len(logs)))

	return logs, total, nil
}

// GetDB 获取数据库连接
func (s *RatingService) GetDB() *gorm.DB {
	return s.db
}

// MigrateOperationLogs 从历史数据迁移操作日志
func (s *RatingService) MigrateOperationLogs() (int, []string, error) {
	logger := utils.GetLogger()

	migratedCount := 0
	var errors []string

	// 1. 从 contest_skip_users 表迁移 Skip 用户操作
	logger.Info("开始迁移 Skip 用户操作日志...")

	type ContestSkipUser struct {
		ID               uint64    `gorm:"primaryKey"`
		ContestID        uint64    `gorm:"type:bigint unsigned"`
		UID              string    `gorm:"type:varchar(32)"`
		Username         string    `gorm:"type:varchar(100)"`
		Reason           string    `gorm:"type:varchar(500)"`
		OperatorUID      string    `gorm:"type:varchar(32)"`
		OperatorUsername string    `gorm:"type:varchar(100)"`
		IsApplied        bool      `gorm:"type:tinyint(1)"`
		CreatedAt        time.Time `gorm:"autoCreateTime"`
	}

	var skipUsers []ContestSkipUser
	if err := s.db.Find(&skipUsers).Error; err == nil {
		for _, skip := range skipUsers {
			// 检查是否已存在（避免重复迁移）
			var existingCount int64
			s.db.Raw("SELECT COUNT(*) FROM rating_operation_logs WHERE operation_type = 'skip_user' AND target_id = ? AND created_at = ?",
				skip.ContestID, skip.CreatedAt).Scan(&existingCount)

			if existingCount == 0 {
				detail := map[string]interface{}{
					"contestId": skip.ContestID,
					"usernames": []string{skip.Username},
					"reason":    skip.Reason,
				}
				detailJSON, _ := json.Marshal(detail)

				if err := s.LogOperation(
					skip.OperatorUID,
					skip.OperatorUsername,
					"skip_user",
					"contest",
					fmt.Sprintf("%d", skip.ContestID),
					string(detailJSON),
					"",
				); err == nil {
					migratedCount++
				} else {
					errors = append(errors, "Skip用户迁移失败: "+err.Error())
				}
			}
		}
		logger.Info("Skip用户迁移完成", zap.Int("count", len(skipUsers)))
	} else {
		logger.Error("查询Skip用户失败", zap.Error(err))
	}

	// 2. 从 rating_recalculate_queue 表迁移重算操作
	logger.Info("开始迁移 Rating 重算操作日志...")

	type RecalculateQueue struct {
		ID                uint64     `gorm:"primaryKey"`
		ContestID         uint64     `gorm:"type:bigint unsigned"`
		Status            string     `gorm:"type:varchar(20)"`
		TotalContests     int        `gorm:"type:int"`
		ProcessedContests int        `gorm:"type:int"`
		ErrorMessage      string     `gorm:"type:text"`
		CreatedBy         string     `gorm:"type:varchar(32)"`
		CreatedAt         time.Time  `gorm:"autoCreateTime"`
		StartedAt         *time.Time `json:"startedAt"`
		CompletedAt       *time.Time `json:"completedAt"`
	}

	var recalcTasks []RecalculateQueue
	if err := s.db.Where("status = ?", "completed").Find(&recalcTasks).Error; err == nil {
		for _, task := range recalcTasks {
			// 检查是否已存在
			var existingCount int64
			s.db.Raw("SELECT COUNT(*) FROM rating_operation_logs WHERE operation_type = 'recalculate' AND target_id = ? AND created_at = ?",
				task.ContestID, task.CreatedAt).Scan(&existingCount)

			if existingCount == 0 {
				detail := map[string]interface{}{
					"taskId":       task.ID,
					"contestId":    task.ContestID,
					"contestCount": task.TotalContests,
				}
				detailJSON, _ := json.Marshal(detail)

				if err := s.LogOperation(
					task.CreatedBy,
					task.CreatedBy,
					"recalculate",
					"contest",
					fmt.Sprintf("%d", task.ContestID),
					string(detailJSON),
					"",
				); err == nil {
					migratedCount++
				} else {
					errors = append(errors, "重算任务迁移失败: "+err.Error())
				}
			}
		}
		logger.Info("重算任务迁移完成", zap.Int("count", len(recalcTasks)))
	} else {
		logger.Error("查询重算任务失败", zap.Error(err))
	}

	// 3. 从 rating_history 表迁移手动调整操作（is_manual = 1）
	logger.Info("开始迁移手动调整操作日志...")

	type UserInfo struct {
		UUID     string `gorm:"type:varchar(32)"`
		Username string `gorm:"type:varchar(100)"`
	}

	type RatingHistory struct {
		ID           uint64    `gorm:"primaryKey"`
		UID          string    `gorm:"type:varchar(32)"`
		ContestID    *uint64   `gorm:"type:bigint unsigned"`
		OldRating    *int      `gorm:"type:int"`
		NewRating    int       `gorm:"type:int"`
		RatingChange int       `gorm:"type:int"`
		Reason       string    `gorm:"type:varchar(500)"`
		IsManual     bool      `gorm:"type:tinyint(1)"`
		OperatorUID  string    `gorm:"type:varchar(32)"`
		CreatedAt    time.Time `gorm:"autoCreateTime"`
	}

	var manualAdjustments []RatingHistory
	if err := s.db.Where("is_manual = ?", true).Find(&manualAdjustments).Error; err == nil {
		for _, adj := range manualAdjustments {
			// 检查是否已存在
			var existingCount int64
			s.db.Raw("SELECT COUNT(*) FROM rating_operation_logs WHERE operation_type = 'personal_adjust' AND target_id = ? AND created_at = ?",
				adj.UID, adj.CreatedAt).Scan(&existingCount)

			if existingCount == 0 {
				// 获取用户名
				var userInfo UserInfo
				username := adj.UID
				s.db.Where("uuid = ?", adj.UID).First(&userInfo)
				if userInfo.Username != "" {
					username = userInfo.Username
				}

				detail := map[string]interface{}{
					"username":     username,
					"oldRating":    adj.OldRating,
					"newRating":    adj.NewRating,
					"ratingChange": adj.RatingChange,
					"reason":       adj.Reason,
				}
				detailJSON, _ := json.Marshal(detail)

				if err := s.LogOperation(
					adj.OperatorUID,
					username,
					"personal_adjust",
					"user",
					username,
					string(detailJSON),
					"",
				); err == nil {
					migratedCount++
				} else {
					errors = append(errors, "手动调整迁移失败: "+err.Error())
				}
			}
		}
		logger.Info("手动调整迁移完成", zap.Int("count", len(manualAdjustments)))
	} else {
		logger.Error("查询手动调整失败", zap.Error(err))
	}

	logger.Info("操作日志迁移完成", zap.Int("migrated", migratedCount), zap.Int("errors", len(errors)))

	return migratedCount, errors, nil
}
