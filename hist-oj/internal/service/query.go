package service

import (
	"strconv"

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

// GetUserRating 获取用户rating信息（支持 UID 或用户名）
func (s *QueryService) GetUserRating(uidOrUsername string) (map[string]interface{}, error) {
	logger := utils.GetLogger()
	var userRecord model.UserRecord

	// 默认 Rating 值
	defaultRating := 1200

	// 先尝试通过 UID 查询
	actualUID := uidOrUsername
	if err := s.db.Where("uid = ?", uidOrUsername).First(&userRecord).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 如果通过 UID 找不到，尝试通过用户名查询 user_info 获取真正的 UUID
			var userInfo model.UserInfo
			if err := s.db.Where("username = ? OR uuid = ?", uidOrUsername, uidOrUsername).First(&userInfo).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					logger.Debug("用户不存在于HOJ系统", zap.String("uidOrUsername", uidOrUsername))
					return nil, nil
				}
				logger.Error("查询用户信息失败", zap.String("uidOrUsername", uidOrUsername), zap.Error(err))
				return nil, err
			}

			// 获取到真正的 UUID
			actualUID = userInfo.UUID

			// 再次尝试查询 user_record
			if err := s.db.Where("uid = ?", actualUID).First(&userRecord).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					// 用户存在，创建初始 Rating 记录
					logger.Info("为用户创建初始rating记录",
						zap.String("uuid", actualUID),
						zap.String("username", userInfo.Username),
						zap.Int("initial_rating", defaultRating))
					userRecord = model.UserRecord{
						UID:        actualUID,
						HistRating: &defaultRating,
					}
					if err := s.db.Create(&userRecord).Error; err != nil {
						logger.Error("创建用户rating记录失败", zap.String("uuid", actualUID), zap.Error(err))
						return nil, err
					}
				} else {
					logger.Error("查询用户记录失败", zap.String("uuid", actualUID), zap.Error(err))
					return nil, err
				}
			}
		} else {
			logger.Error("查询用户记录失败", zap.String("uidOrUsername", uidOrUsername), zap.Error(err))
			return nil, err
		}
	}

	rating := defaultRating
	if userRecord.HistRating != nil {
		rating = *userRecord.HistRating
	}

	// 获取最高rating
	var maxRating *int
	if err := s.db.Model(&model.RatingHistory{}).
		Where("uid = ?", actualUID).
		Select("MAX(new_rating) as max_rating").
		Scan(&maxRating).Error; err != nil {
		logger.Warn("查询最高rating失败", zap.String("uid", actualUID), zap.Error(err))
		// 继续执行，使用当前rating作为最高rating
	}

	finalMaxRating := rating
	if maxRating != nil && *maxRating > rating {
		finalMaxRating = *maxRating
	}

	// 获取颜色信息
	colorInfo := utils.GetRatingColorInfo(rating)

	result := map[string]interface{}{
		"uid":      actualUID,
		"rating":   rating,
		"maxRating": finalMaxRating,
		"color":    colorInfo.Color,
		"level":    colorInfo.Name,
		"levelZh":  colorInfo.NameZh,
	}

	return result, nil
}

// GetRatingHistory 获取用户rating历史记录（支持 UID 或用户名）
func (s *QueryService) GetRatingHistory(uidOrUsername string, page, limit int) (map[string]interface{}, error) {
	logger := utils.GetLogger()

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	offset := (page - 1) * limit

	// 先尝试获取真实的 UID（支持用户名查询）
	actualUID := uidOrUsername
	var userRecord model.UserRecord
	if err := s.db.Where("uid = ?", uidOrUsername).First(&userRecord).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 如果通过 UID 找不到，尝试通过用户名查询 user_info 获取真正的 UUID
			var userInfo model.UserInfo
			if err := s.db.Where("username = ? OR uuid = ?", uidOrUsername, uidOrUsername).First(&userInfo).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					logger.Debug("用户不存在于HOJ系统", zap.String("uidOrUsername", uidOrUsername))
					// 返回空结果而不是错误
					return map[string]interface{}{
						"total":   0,
						"page":    page,
						"limit":   limit,
						"records": []model.RatingHistory{},
					}, nil
				}
				logger.Error("查询用户信息失败", zap.String("uidOrUsername", uidOrUsername), zap.Error(err))
				return nil, err
			}
			// 获取到真正的 UUID
			actualUID = userInfo.UUID
		}
	}

	var histories []model.RatingHistory
	var total int64

	// 查询总数
	if err := s.db.Model(&model.RatingHistory{}).Where("uid = ?", actualUID).Count(&total).Error; err != nil {
		logger.Error("查询rating历史总数失败",
			zap.String("uid", actualUID),
			zap.Error(err))
		return nil, err
	}

	// 查询记录
	if err := s.db.Where("uid = ?", actualUID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&histories).Error; err != nil {
		logger.Error("查询rating历史失败",
			zap.String("uid", actualUID),
			zap.Int("page", page),
			zap.Int("limit", limit),
			zap.Error(err))
		return nil, err
	}

	// 填充比赛标题
	for i := range histories {
		var contest model.Contest
		if err := s.db.Where("id = ?", histories[i].ContestID).First(&contest).Error; err == nil {
			histories[i].ContestTitle = contest.Title
		}
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
		Order("`rank` ASC").
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

// GetBatchUserRating 批量获取用户rating信息
func (s *QueryService) GetBatchUserRating(uids []string) (map[string]interface{}, error) {
	logger := utils.GetLogger()

	// 查询所有用户的 rating 记录
	var userRecords []model.UserRecord
	if err := s.db.Where("uid IN ?", uids).Find(&userRecords).Error; err != nil {
		logger.Error("批量查询用户rating失败", zap.Error(err))
		return nil, err
	}

	// 构建 uid -> rating 的映射
	ratingMap := make(map[string]int)
	for _, record := range userRecords {
		if record.HistRating != nil {
			ratingMap[record.UID] = *record.HistRating
		} else {
			ratingMap[record.UID] = 1200
		}
	}

	// 对于没有记录的用户，返回默认值 1200
	results := make(map[string]interface{})
	for _, uid := range uids {
		rating, exists := ratingMap[uid]
		if !exists {
			rating = 1200
		}

		colorInfo := utils.GetRatingColorInfo(rating)
		results[uid] = map[string]interface{}{
			"uid":     uid,
			"rating":  rating,
			"color":   colorInfo.Color,
			"level":   colorInfo.Name,
			"levelZh": colorInfo.NameZh,
		}
	}

	return results, nil
}

// InitializeUserRating 初始化用户默认rating
func (s *QueryService) InitializeUserRating(uid string, initialRating int) error {
	logger := utils.GetLogger()

	// 检查用户是否已有 rating 记录
	var existingRecord model.UserRecord
	err := s.db.Where("uid = ?", uid).First(&existingRecord).Error

	if err == nil {
		// 记录已存在，不需要初始化
		logger.Info("用户已有rating记录，无需初始化", zap.String("uid", uid))
		return nil
	}

	if err != gorm.ErrRecordNotFound {
		logger.Error("查询用户记录失败", zap.String("uid", uid), zap.Error(err))
		return err
	}

	// 创建新的 rating 记录
	userRecord := model.UserRecord{
		UID:        uid,
		HistRating: &initialRating,
	}

	if err := s.db.Create(&userRecord).Error; err != nil {
		logger.Error("创建用户rating记录失败", zap.String("uid", uid), zap.Error(err))
		return err
	}

	logger.Info("初始化用户rating成功",
		zap.String("uid", uid),
		zap.Int("rating", initialRating))
	return nil
}

// SetContestRatingType 设置比赛的Rating类型
func (s *QueryService) SetContestRatingType(contestID uint64, isRating bool) error {
	logger := utils.GetLogger()

	// 更新比赛的 is_rating 字段
	result := s.db.Model(&model.Contest{}).
		Where("id = ?", contestID).
		Update("is_rating", isRating)

	if result.Error != nil {
		logger.Error("更新比赛Rating类型失败",
			zap.Uint64("contest_id", contestID),
			zap.Bool("is_rating", isRating),
			zap.Error(result.Error))
		return result.Error
	}

	if result.RowsAffected == 0 {
		logger.Warn("比赛不存在", zap.Uint64("contest_id", contestID))
		return gorm.ErrRecordNotFound
	}

	// 如果设置为 Rating 赛，需要在 contest_rating_status 表中创建或更新记录
	if isRating {
		var status model.ContestRatingStatus
		err := s.db.Where("contest_id = ?", contestID).First(&status).Error

		if err == gorm.ErrRecordNotFound {
			// 记录不存在，创建新记录
			status = model.ContestRatingStatus{
				ContestID:        contestID,
				IsRated:          true,
				RatingCalculated: false,
			}
			if err := s.db.Create(&status).Error; err != nil {
				logger.Error("创建比赛Rating状态记录失败",
					zap.Uint64("contest_id", contestID),
					zap.Error(err))
				return err
			}
			logger.Info("创建比赛Rating状态记录成功", zap.Uint64("contest_id", contestID))
		} else if err != nil {
			logger.Error("查询比赛Rating状态失败",
				zap.Uint64("contest_id", contestID),
				zap.Error(err))
			return err
		} else {
			// 记录已存在，更新 is_rated 字段
			if err := s.db.Model(&status).Update("is_rated", true).Error; err != nil {
				logger.Error("更新比赛Rating状态失败",
					zap.Uint64("contest_id", contestID),
					zap.Error(err))
				return err
			}
			logger.Info("更新比赛Rating状态成功", zap.Uint64("contest_id", contestID))
		}
	}

	logger.Info("设置比赛Rating类型成功",
		zap.Uint64("contest_id", contestID),
		zap.Bool("is_rating", isRating))
	return nil
}

// GetContestInfo 获取比赛信息
func (s *QueryService) GetContestInfo(contestID uint64) (map[string]interface{}, error) {
	logger := utils.GetLogger()

	var contest model.Contest
	if err := s.db.Where("id = ?", contestID).First(&contest).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Debug("比赛不存在", zap.Uint64("contest_id", contestID))
			return nil, nil
		}
		logger.Error("查询比赛信息失败",
			zap.Uint64("contest_id", contestID),
			zap.Error(err))
		return nil, err
	}

	result := map[string]interface{}{
		"id":        contest.ID,
		"title":     contest.Title,
		"type":      contest.Type,
		"isRating":  contest.IsRating,
		"startTime": contest.StartTime,
		"endTime":   contest.EndTime,
		"status":    contest.Status,
	}

	return result, nil
}

// GetBatchContestInfo 批量获取比赛信息
func (s *QueryService) GetBatchContestInfo(contestIDs []uint64) (map[string]interface{}, error) {
	logger := utils.GetLogger()

	// 查询所有比赛的信息
	var contests []model.Contest
	if err := s.db.Where("id IN ?", contestIDs).Find(&contests).Error; err != nil {
		logger.Error("批量查询比赛信息失败", zap.Error(err))
		return nil, err
	}

	// 构建 contestID -> isRating 的映射
	results := make(map[string]interface{})
	for _, contest := range contests {
		results[strconv.FormatUint(contest.ID, 10)] = map[string]interface{}{
			"id":       contest.ID,
			"isRating": contest.IsRating,
		}
	}

	// 对于没有找到的比赛，返回默认值 false
	for _, id := range contestIDs {
		key := strconv.FormatUint(id, 10)
		if _, exists := results[key]; !exists {
			results[key] = map[string]interface{}{
				"id":       id,
				"isRating": false,
			}
		}
	}

	return results, nil
}

// GetRatingRank 获取 Rating 排名列表
func (s *QueryService) GetRatingRank(page, limit int, keyword string) (map[string]interface{}, error) {
	logger := utils.GetLogger()

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 30
	}

	offset := (page - 1) * limit

	// 构建子查询：获取每个用户最新的 rating
	// 从 rating_history 表中获取每个用户最新的 new_rating
	subQuery := s.db.Table("rating_history as rh1").
		Select("rh1.uid, rh1.new_rating").
		Joins("INNER JOIN (SELECT uid, MAX(created_at) as max_time FROM rating_history GROUP BY uid) as rh2 ON rh1.uid = rh2.uid AND rh1.created_at = rh2.max_time")

	// 构建主查询 - 使用 user_info 作为主表
	// 这样可以显示所有用户，即使他们没有 rating 记录
	query := s.db.Table("user_info").
		Select("user_info.uuid as uid, COALESCE(latest_rating.new_rating, 1200) as hist_rating, user_info.username, user_info.avatar, user_info.nickname, user_info.school, user_info.title_name, user_info.title_color").
		Joins("LEFT JOIN (?) as latest_rating ON user_info.uuid = latest_rating.uid", subQuery)

	// 如果有关键词搜索
	if keyword != "" {
		query = query.Where("user_info.username LIKE ? OR user_info.nickname LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 查询总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		logger.Error("查询 Rating 排名总数失败", zap.Error(err))
		return nil, err
	}

	// 查询排名数据
	type RankResult struct {
		UID        string  `json:"uid"`
		HistRating int     `json:"histRating"`
		Username   string  `json:"username"`
		Avatar     *string `json:"avatar"`
		Nickname   *string `json:"nickname"`
		School     *string `json:"school"`
		TitleName  *string `json:"titleName"`
		TitleColor *string `json:"titleColor"`
	}

	var results []RankResult
	if err := query.
		Order("hist_rating DESC, user_info.username ASC").
		Offset(offset).
		Limit(limit).
		Scan(&results).Error; err != nil {
		logger.Error("查询 Rating 排名失败", zap.Error(err))
		return nil, err
	}

	logger.Info("查询到 Rating 排名数据",
		zap.Int("total", int(total)),
		zap.Int("records", len(results)),
		zap.Int("page", page),
		zap.Int("limit", limit))

	// 构建返回结果
	records := make([]map[string]interface{}, 0, len(results))
	for _, r := range results {
		colorInfo := utils.GetRatingColorInfo(r.HistRating)

		record := map[string]interface{}{
			"uid":      r.UID,
			"username": r.Username,
			"rating":   r.HistRating,
			"color":    colorInfo.Color,
			"level":    colorInfo.Name,
			"levelZh":  colorInfo.NameZh,
		}

		if r.Avatar != nil {
			record["avatar"] = *r.Avatar
		}
		if r.Nickname != nil {
			record["nickname"] = *r.Nickname
		}
		if r.School != nil {
			record["school"] = *r.School
		}
		if r.TitleName != nil {
			record["titleName"] = *r.TitleName
		}
		if r.TitleColor != nil {
			record["titleColor"] = *r.TitleColor
		}

		records = append(records, record)
	}

	result := map[string]interface{}{
		"total":   total,
		"page":    page,
		"limit":   limit,
		"records": records,
	}

	return result, nil
}

