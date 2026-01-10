package service

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/hoj/hist-oj/internal/client"
	"github.com/hoj/hist-oj/internal/model"
	"github.com/hoj/hist-oj/internal/utils"
)

// BattleService 对战服务
type BattleService struct {
	db *gorm.DB
}

// NewBattleService 创建对战服务
func NewBattleService() *BattleService {
	return &BattleService{
		db: client.DB,
	}
}

// CreateRoom 创建对战房间
func (s *BattleService) CreateRoom(hostID, hostUsername string) (*model.BattleRoom, error) {
	logger := utils.GetLogger()

	// 检查用户是否已在其他房间中（只检查等待中和进行中的房间，不包括已结束的）
	var existingRoom model.BattleRoom
	err := s.db.Where("(host_id = ? OR challenger_id = ?) AND status IN (?)",
		hostID, hostID, []int{int(model.BattleStatusWaiting), int(model.BattleStatusInProgress)}).First(&existingRoom).Error

	if err == nil {
		// 用户已在其他房间中
		logger.Warn("用户已在其他房间中",
			zap.String("user_id", hostID),
			zap.String("existing_room_id", existingRoom.RoomID),
			zap.Int("status", int(existingRoom.Status)))
		return nil, fmt.Errorf("你已在房间 %s 中，请先退出当前房间", existingRoom.RoomID)
	} else if err != gorm.ErrRecordNotFound {
		logger.Error("查询用户现有房间失败",
			zap.String("user_id", hostID),
			zap.Error(err))
		return nil, fmt.Errorf("查询房间状态失败: %w", err)
	}

	// 生成唯一房间号（2个大写字母+4位数字）
	roomID := s.generateRoomID()

	room := &model.BattleRoom{
		RoomID:       roomID,
		HostID:       hostID,
		HostUsername: hostUsername,
		Status:       model.BattleStatusWaiting,
	}

	if err := s.db.Create(room).Error; err != nil {
		logger.Error("创建对战房间失败",
			zap.String("host_id", hostID),
			zap.Error(err))
		return nil, fmt.Errorf("创建房间失败: %w", err)
	}

	logger.Info("创建对战房间成功",
		zap.String("room_id", roomID),
		zap.String("host_id", hostID))
	return room, nil
}

// JoinRoom 加入对战房间
func (s *BattleService) JoinRoom(roomID, challengerID, challengerUsername string) (*model.BattleRoom, error) {
	logger := utils.GetLogger()

	// 检查用户是否已在其他房间中（只检查等待中和进行中的房间，不包括已结束的）
	var existingRoom model.BattleRoom
	err := s.db.Where("(host_id = ? OR challenger_id = ?) AND status IN (?)",
		challengerID, challengerID, []int{int(model.BattleStatusWaiting), int(model.BattleStatusInProgress)}).First(&existingRoom).Error

	if err == nil {
		// 用户已在其他房间中
		logger.Warn("用户已在其他房间中",
			zap.String("user_id", challengerID),
			zap.String("existing_room_id", existingRoom.RoomID),
			zap.Int("status", int(existingRoom.Status)))
		return nil, fmt.Errorf("你已在房间 %s 中，请先退出当前房间", existingRoom.RoomID)
	} else if err != gorm.ErrRecordNotFound {
		logger.Error("查询用户现有房间失败",
			zap.String("user_id", challengerID),
			zap.Error(err))
		return nil, fmt.Errorf("查询房间状态失败: %w", err)
	}

	// 查询房间
	var room model.BattleRoom
	if err := s.db.Where("room_id = ?", roomID).First(&room).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("房间不存在")
		}
		logger.Error("查询房间失败",
			zap.String("room_id", roomID),
			zap.Error(err))
		return nil, fmt.Errorf("查询房间失败: %w", err)
	}

	// 检查房间状态
	if room.Status != model.BattleStatusWaiting {
		return nil, fmt.Errorf("房间已开始或已结束")
	}

	// 检查是否是房主
	if room.HostID == challengerID {
		return nil, fmt.Errorf("不能加入自己创建的房间")
	}

	// 检查是否有挑战者
	if room.ChallengerID != nil {
		return nil, fmt.Errorf("房间已满员")
	}

	// 更新房间信息
	room.ChallengerID = &challengerID
	room.ChallengerUsername = &challengerUsername
	room.ChallengerReady = false // 新加入的挑战者默认未准备

	if err := s.db.Save(&room).Error; err != nil {
		logger.Error("加入房间失败",
			zap.String("room_id", roomID),
			zap.String("challenger_id", challengerID),
			zap.Error(err))
		return nil, fmt.Errorf("加入房间失败: %w", err)
	}

	logger.Info("加入房间成功",
		zap.String("room_id", roomID),
		zap.String("challenger_id", challengerID))
	return &room, nil
}

// ReadyBattle 准备对战（挑战者准备/取消准备）
func (s *BattleService) ReadyBattle(roomID, userID string, ready bool) (*model.BattleRoom, error) {
	logger := utils.GetLogger()

	// 查询房间
	var room model.BattleRoom
	if err := s.db.Where("room_id = ?", roomID).First(&room).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("房间不存在")
		}
		logger.Error("查询房间失败",
			zap.String("room_id", roomID),
			zap.Error(err))
		return nil, fmt.Errorf("查询房间失败: %w", err)
	}

	// 检查房间状态
	if room.Status != model.BattleStatusWaiting {
		return nil, fmt.Errorf("房间已开始或已结束")
	}

	// 检查是否是挑战者
	if room.ChallengerID == nil {
		return nil, fmt.Errorf("等待挑战者加入")
	}

	if *room.ChallengerID != userID {
		return nil, fmt.Errorf("只有挑战者可以准备")
	}

	// 更新准备状态
	room.ChallengerReady = ready

	if err := s.db.Save(&room).Error; err != nil {
		logger.Error("更新准备状态失败",
			zap.String("room_id", roomID),
			zap.String("user_id", userID),
			zap.Bool("ready", ready),
			zap.Error(err))
		return nil, fmt.Errorf("更新准备状态失败: %w", err)
	}

	logger.Info("更新准备状态成功",
		zap.String("room_id", roomID),
		zap.String("user_id", userID),
		zap.Bool("ready", ready))

	return s.GetRoomInfo(roomID)
}

// GetRoomInfo 获取房间信息
func (s *BattleService) GetRoomInfo(roomID string) (*model.BattleRoom, error) {
	var room model.BattleRoom
	if err := s.db.Where("room_id = ?", roomID).First(&room).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("房间不存在")
		}
		return nil, fmt.Errorf("查询房间失败: %w", err)
	}

	// 填充房主的 Rating
	var hostRecord model.UserRecord
	if err := s.db.Where("uid = ?", room.HostID).First(&hostRecord).Error; err == nil {
		room.HostRating = hostRecord.HistRating
	}

	// 填充挑战者的 Rating
	if room.ChallengerID != nil {
		var challengerRecord model.UserRecord
		if err := s.db.Where("uid = ?", *room.ChallengerID).First(&challengerRecord).Error; err == nil {
			room.ChallengerRating = challengerRecord.HistRating
		}
	}

	return &room, nil
}

// StartBattle 开始对战
func (s *BattleService) StartBattle(roomID string) (*model.BattleRoom, *client.ProblemBasicInfo, error) {
	logger := utils.GetLogger()

	// 查询房间
	var room model.BattleRoom
	if err := s.db.Where("room_id = ?", roomID).First(&room).Error; err != nil {
		return nil, nil, fmt.Errorf("房间不存在")
	}

	// 检查房间状态
	if room.Status != model.BattleStatusWaiting {
		return nil, nil, fmt.Errorf("房间状态不正确")
	}

	// 检查是否有挑战者
	if room.ChallengerID == nil {
		return nil, nil, fmt.Errorf("等待挑战者加入")
	}

	// 检查挑战者是否已准备
	if !room.ChallengerReady {
		return nil, nil, fmt.Errorf("挑战者未准备")
	}

	// 智能选题：获取双方都未AC的题目（返回显示ID）
	displayID, err := s.selectBattleProblem(room.HostID, *room.ChallengerID)
	if err != nil {
		return nil, nil, fmt.Errorf("选择题目失败: %w", err)
	}

	// 通过显示ID获取题目信息
	problem, err := client.GetProblemInfoByDisplayID(displayID)
	if err != nil {
		return nil, nil, fmt.Errorf("获取题目信息失败: %w", err)
	}

	// 更新房间状态（使用显示ID）
	now := time.Now()
	room.ProblemID = &displayID           // 显示ID（如 0051, Z001）
	room.Status = model.BattleStatusInProgress
	room.StartTime = &now

	if err := s.db.Save(&room).Error; err != nil {
		logger.Error("开始对战失败",
			zap.String("room_id", roomID),
			zap.Error(err))
		return nil, nil, fmt.Errorf("开始对战失败: %w", err)
	}

	logger.Info("开始对战成功",
		zap.String("room_id", roomID),
		zap.String("display_id", displayID))
	return &room, problem, nil
}

// selectBattleProblem 智能选择对战题目（返回显示ID）
func (s *BattleService) selectBattleProblem(user1ID, user2ID string) (string, error) {
	logger := utils.GetLogger()

	// 获取双方AC题目列表
	user1ACProblems, err := client.GetUserACProblems(user1ID)
	if err != nil {
		logger.Error("获取用户1的AC题目失败",
			zap.String("user_id", user1ID),
			zap.Error(err))
		return "", err
	}

	user2ACProblems, err := client.GetUserACProblems(user2ID)
	if err != nil {
		logger.Error("获取用户2的AC题目失败",
			zap.String("user_id", user2ID),
			zap.Error(err))
		return "", err
	}

	// 计算双方AC题目的并集
	acUnion := make(map[int64]bool)
	for _, pid := range user1ACProblems {
		acUnion[pid] = true
	}
	for _, pid := range user2ACProblems {
		acUnion[pid] = true
	}

	// 获取题库所有题目（包含显示ID）
	allProblems, err := client.GetProblemListWithDisplayID()
	if err != nil {
		logger.Error("获取题目列表失败", zap.Error(err))
		return "", err
	}

	// 过滤出双方都未AC的题目，并验证题目确实存在
	candidateProblems := make([]struct {
		Pid       int64
		DisplayID string
	}, 0)

	for _, problem := range allProblems {
		// 跳过空的显示ID
		if problem.DisplayID == "" {
			logger.Warn("跳过空显示ID的题目",
				zap.Int64("pid", problem.Pid))
			continue
		}

		if !acUnion[problem.Pid] {
			candidateProblems = append(candidateProblems, struct {
				Pid       int64
				DisplayID string
			}{
				Pid:       problem.Pid,
				DisplayID: problem.DisplayID,
			})
		}
	}

	if len(candidateProblems) == 0 {
		return "", fmt.Errorf("没有找到合适的对战题目")
	}

	// 随机选择一个题目
	rand.Seed(time.Now().UnixNano())
	selectedIndex := rand.Intn(len(candidateProblems))
	selectedProblem := candidateProblems[selectedIndex]

	// 验证题目确实存在（双重验证）
	_, err = client.GetProblemInfoByDisplayID(selectedProblem.DisplayID)
	if err != nil {
		logger.Error("选择的题目不存在，尝试重新选择",
			zap.String("display_id", selectedProblem.DisplayID),
			zap.Error(err))

		// 如果验证失败，尝试从候选列表中移除该题目并重新选择
		if len(candidateProblems) == 1 {
			return "", fmt.Errorf("没有找到可用的对战题目")
		}

		// 重新选择（移除当前索引）
		newCandidates := make([]struct {
			Pid       int64
			DisplayID string
		}, 0, len(candidateProblems)-1)

		for i, p := range candidateProblems {
			if i != selectedIndex {
				newCandidates = append(newCandidates, p)
			}
		}

		// 尝试最多3次，直到找到一个存在的题目
		maxAttempts := 3
		for attempt := 0; attempt < maxAttempts; attempt++ {
			// 再次随机选择
			newIndex := rand.Intn(len(newCandidates))
			selectedProblem = newCandidates[newIndex]

			// 验证新选择的题目
			_, err = client.GetProblemInfoByDisplayID(selectedProblem.DisplayID)
			if err == nil {
				// 找到了有效题目
				logger.Info("重新选择对战题目成功",
					zap.String("display_id", selectedProblem.DisplayID),
					zap.Int("attempt", attempt+1))
				break
			}

			logger.Warn("重新选择的题目也不存在",
				zap.String("display_id", selectedProblem.DisplayID),
				zap.Int("attempt", attempt+1),
				zap.Error(err))

			// 如果只剩一个候选且验证失败，返回错误
			if len(newCandidates) == 1 {
				return "", fmt.Errorf("没有找到可用的对战题目")
			}

			// 移除无效的题目
			tempCandidates := make([]struct {
				Pid       int64
				DisplayID string
			}, 0, len(newCandidates)-1)

			for i, p := range newCandidates {
				if i != newIndex {
					tempCandidates = append(tempCandidates, p)
				}
			}
			newCandidates = tempCandidates
		}
	}

	logger.Info("选择对战题目",
		zap.String("user1_id", user1ID),
		zap.String("user2_id", user2ID),
		zap.String("display_id", selectedProblem.DisplayID),
		zap.Int64("pid", selectedProblem.Pid),
		zap.Int("candidate_count", len(candidateProblems)))

	return selectedProblem.DisplayID, nil
}

// GiveupBattle 放弃对战
func (s *BattleService) GiveupBattle(roomID, userID string) (winner string, endReason string, err error) {
	logger := utils.GetLogger()

	// 查询房间
	var room model.BattleRoom
	if err := s.db.Where("room_id = ?", roomID).First(&room).Error; err != nil {
		return "", "", fmt.Errorf("房间不存在")
	}

	// 检查房间状态
	if room.Status != model.BattleStatusInProgress {
		return "", "", fmt.Errorf("房间未开始或已结束")
	}

	// 判断胜负
	if room.HostID == userID {
		// 房主放弃，挑战者获胜
		winner = *room.ChallengerID
	} else if room.ChallengerID != nil && *room.ChallengerID == userID {
		// 挑战者放弃，房主获胜
		winner = room.HostID
	} else {
		return "", "", fmt.Errorf("用户不在房间中")
	}

	// 结束对战
	if err := s.endBattle(&room, winner, model.EndReasonGiveUp); err != nil {
		return "", "", fmt.Errorf("结束对战失败: %w", err)
	}

	logger.Info("放弃对战",
		zap.String("room_id", roomID),
		zap.String("loser", userID),
		zap.String("winner", winner))
	return winner, model.EndReasonGiveUp, nil
}

// SubmitAC 用户提交AC
func (s *BattleService) SubmitAC(roomID, userID string, problemID string) error {
	logger := utils.GetLogger()

	// 查询房间
	var room model.BattleRoom
	if err := s.db.Where("room_id = ?", roomID).First(&room).Error; err != nil {
		return fmt.Errorf("房间不存在")
	}

	// 检查房间状态
	if room.Status != model.BattleStatusInProgress {
		return fmt.Errorf("房间未开始或已结束")
	}

	// 检查是否是对战的题目（统一转换为大写并去除空格后比较）
	roomProblemID := strings.ToUpper(strings.TrimSpace(*room.ProblemID))
	submitProblemID := strings.ToUpper(strings.TrimSpace(problemID))

	if room.ProblemID == nil || roomProblemID != submitProblemID {
		logger.Warn("题目ID不匹配",
			zap.String("room_problem_id", roomProblemID),
			zap.String("submit_problem_id", submitProblemID))
		return fmt.Errorf("不是对战题目")
	}

	// 判断是哪方提交AC
	var winner string
	if room.HostID == userID {
		winner = room.HostID
	} else if room.ChallengerID != nil && *room.ChallengerID == userID {
		winner = *room.ChallengerID
	} else {
		return fmt.Errorf("用户不在房间中")
	}

	// 结束对战
	if err := s.endBattle(&room, winner, model.EndReasonAC); err != nil {
		return fmt.Errorf("结束对战失败: %w", err)
	}

	logger.Info("提交AC，对战结束",
		zap.String("room_id", roomID),
		zap.String("winner", winner))
	return nil
}

// endBattle 结束对战（内部方法）
func (s *BattleService) endBattle(room *model.BattleRoom, winnerID, endReason string) error {
	logger := utils.GetLogger()
	now := time.Now()

	// 更新房间状态
	room.Status = model.BattleStatusEnded
	room.WinnerID = &winnerID
	room.EndReason = &endReason
	room.EndTime = &now

	// 计算对战时长
	var battleTime int
	if room.StartTime != nil {
		battleTime = int(now.Sub(*room.StartTime).Seconds())
	}

	// 开始事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 保存房间状态
	if err := tx.Save(room).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("保存房间状态失败: %w", err)
	}

	// 获取题目信息
	var problemTitle string
	if room.ProblemID != nil {
		if problem, err := client.GetProblemInfoByDisplayID(*room.ProblemID); err == nil {
			problemTitle = problem.Title
		}
	}

	// 获取双方的rating
	var hostRating, challengerRating *int
	var hostRecord model.UserRecord
	if err := s.db.Where("uid = ?", room.HostID).First(&hostRecord).Error; err == nil {
		hostRating = hostRecord.HistRating
	}

	var challengerRecord model.UserRecord
	if room.ChallengerID != nil {
		if err := s.db.Where("uid = ?", *room.ChallengerID).First(&challengerRecord).Error; err == nil {
			challengerRating = challengerRecord.HistRating
		}
	}

	// 创建双方对战记录
	records := s.createBattleRecords(room, winnerID, endReason, problemTitle, battleTime, hostRating, challengerRating)
	for _, record := range records {
		if err := tx.Create(&record).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("创建对战记录失败: %w", err)
		}
	}

	// 更新双方对战统计
	for _, record := range records {
		if err := s.updateUserStats(tx, record.UserID, record.Username, record.IsWinner, record.SubmitCount, battleTime); err != nil {
			tx.Rollback()
			return fmt.Errorf("更新用户统计失败: %w", err)
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	logger.Info("结束对战成功",
		zap.String("room_id", room.RoomID),
		zap.String("winner", winnerID),
		zap.String("end_reason", endReason),
		zap.Int("battle_time", battleTime))
	return nil
}

// createBattleRecords 创建对战记录
func (s *BattleService) createBattleRecords(room *model.BattleRoom, winnerID, endReason, problemTitle string, battleTime int, hostRating, challengerRating *int) []model.BattleRecord {
	records := make([]model.BattleRecord, 0, 2)

	// 房主记录
	hostRecord := model.BattleRecord{
		RoomID:           room.RoomID,
		UserID:           room.HostID,
		Username:         room.HostUsername,
		OpponentID:       *room.ChallengerID,
		OpponentUsername: *room.ChallengerUsername,
		OpponentRating:   challengerRating,
		ProblemID:        *room.ProblemID,
		ProblemTitle:     problemTitle,
		IsWinner:         room.HostID == winnerID,
		EndReason:        endReason,
		SubmitCount:      0, // TODO: 需要从提交记录中获取
		BattleTime:       &battleTime,
	}
	records = append(records, hostRecord)

	// 挑战者记录
	challengerRecord := model.BattleRecord{
		RoomID:           room.RoomID,
		UserID:           *room.ChallengerID,
		Username:         *room.ChallengerUsername,
		OpponentID:       room.HostID,
		OpponentUsername: room.HostUsername,
		OpponentRating:   hostRating,
		ProblemID:        *room.ProblemID,
		ProblemTitle:     problemTitle,
		IsWinner:         *room.ChallengerID == winnerID,
		EndReason:        endReason,
		SubmitCount:      0, // TODO: 需要从提交记录中获取
		BattleTime:       &battleTime,
	}
	records = append(records, challengerRecord)

	return records
}

// updateUserStats 更新用户对战统计
func (s *BattleService) updateUserStats(db *gorm.DB, userID, username string, isWinner bool, submitCount, battleTime int) error {
	logger := utils.GetLogger()

	var stats model.UserBattleStats
	err := db.Where("user_id = ?", userID).First(&stats).Error

	if err == gorm.ErrRecordNotFound {
		// 创建新统计
		stats = model.UserBattleStats{
			UserID:       userID,
			Username:     username,
			TotalBattles: 1,
		}
		if isWinner {
			stats.WinCount = 1
		} else {
			stats.LoseCount = 1
		}
		stats.TotalSubmitCount = submitCount
		stats.AvgBattleTime = &battleTime

		if err := db.Create(&stats).Error; err != nil {
			logger.Error("创建用户统计失败",
				zap.String("user_id", userID),
				zap.Error(err))
			return err
		}
	} else if err != nil {
		logger.Error("查询用户统计失败",
			zap.String("user_id", userID),
			zap.Error(err))
		return err
	} else {
		// 更新现有统计
		stats.TotalBattles++
		if isWinner {
			stats.WinCount++
		} else {
			stats.LoseCount++
		}
		stats.TotalSubmitCount += submitCount
		stats.Username = username // 确保用户名是最新的

		// 更新平均对战时长
		if stats.AvgBattleTime != nil {
			newAvg := (*stats.AvgBattleTime*(stats.TotalBattles-1) + battleTime) / stats.TotalBattles
			stats.AvgBattleTime = &newAvg
		} else {
			stats.AvgBattleTime = &battleTime
		}

		// 计算胜率
		if stats.TotalBattles > 0 {
			stats.WinRate = float64(stats.WinCount) / float64(stats.TotalBattles) * 100
		}

		if err := db.Save(&stats).Error; err != nil {
			logger.Error("更新用户统计失败",
				zap.String("user_id", userID),
				zap.Error(err))
			return err
		}
	}

	return nil
}

// GetBattleRank 获取对战排行榜
func (s *BattleService) GetBattleRank(limit, page int, username string) ([]model.UserBattleStats, int64, error) {
	logger := utils.GetLogger()

	var stats []model.UserBattleStats
	var total int64

	// 计算偏移量
	offset := (page - 1) * limit

	// 构建查询
	query := s.db.Model(&model.UserBattleStats{})

	// 添加用户名筛选条件（支持子串搜索）
	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}

	// 查询总数
	if err := query.Count(&total).Error; err != nil {
		logger.Error("查询排行榜总数失败", zap.Error(err))
		return nil, 0, fmt.Errorf("查询失败: %w", err)
	}

	// 查询排行榜（按胜场数降序，胜率降序）
	if err := query.Order("win_count DESC, win_rate DESC").
		Limit(limit).
		Offset(offset).
		Find(&stats).Error; err != nil {
		logger.Error("查询排行榜失败", zap.Error(err))
		return nil, 0, fmt.Errorf("查询失败: %w", err)
	}

	// 为每个用户填充 Rating 信息
	for i := range stats {
		var userRecord model.UserRecord
		if err := s.db.Where("uid = ?", stats[i].UserID).First(&userRecord).Error; err == nil {
			stats[i].Rating = userRecord.HistRating
		}
		// 如果查询失败（用户没有 rating），Rating 保持为 nil
	}

	logger.Debug("查询排行榜成功",
		zap.Int("page", page),
		zap.Int("limit", limit),
		zap.Int64("total", total))
	return stats, total, nil
}

// GetMyBattleRecords 获取我的对战记录
func (s *BattleService) GetMyBattleRecords(userID string, limit, page int) ([]model.BattleRecord, int64, error) {
	logger := utils.GetLogger()

	var records []model.BattleRecord
	var total int64

	offset := (page - 1) * limit

	// 查询总数
	if err := s.db.Model(&model.BattleRecord{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		logger.Error("查询对战记录总数失败",
			zap.String("user_id", userID),
			zap.Error(err))
		return nil, 0, fmt.Errorf("查询失败: %w", err)
	}

	// 查询记录（按创建时间降序）
	if err := s.db.Where("user_id = ?", userID).
		Order("gmt_create DESC").
		Limit(limit).
		Offset(offset).
		Find(&records).Error; err != nil {
		logger.Error("查询对战记录失败",
			zap.String("user_id", userID),
			zap.Error(err))
		return nil, 0, fmt.Errorf("查询失败: %w", err)
	}

	// 为每条记录填充用户的 Rating
	for i := range records {
		var userRecord model.UserRecord
		if err := s.db.Where("uid = ?", records[i].UserID).First(&userRecord).Error; err == nil {
			records[i].UserRating = userRecord.HistRating
		}
		// 如果查询失败，UserRating 保持为 nil
	}

	logger.Debug("查询对战记录成功",
		zap.String("user_id", userID),
		zap.Int("page", page),
		zap.Int64("total", total))
	return records, total, nil
}

// GetAllBattleRecords 获取所有对战记录（管理员功能）
func (s *BattleService) GetAllBattleRecords(limit, page int, username, roomId, isWinnerStr string) ([]model.BattleRecord, int64, error) {
	logger := utils.GetLogger()

	var records []model.BattleRecord
	var total int64

	offset := (page - 1) * limit

	// 构建查询
	query := s.db.Model(&model.BattleRecord{})

	// 添加筛选条件
	if username != "" {
		query = query.Where("username LIKE ? OR opponent_username LIKE ?", "%"+username+"%", "%"+username+"%")
	}
	if roomId != "" {
		query = query.Where("room_id LIKE ?", "%"+roomId+"%")
	}
	if isWinnerStr != "" {
		isWinner := isWinnerStr == "true"
		query = query.Where("is_winner = ?", isWinner)
	}

	// 查询总数
	if err := query.Count(&total).Error; err != nil {
		logger.Error("查询所有对战记录总数失败",
			zap.Error(err))
		return nil, 0, fmt.Errorf("查询失败: %w", err)
	}

	// 查询记录（按创建时间降序）
	if err := query.
		Order("gmt_create DESC").
		Limit(limit).
		Offset(offset).
		Find(&records).Error; err != nil {
		logger.Error("查询所有对战记录失败",
			zap.Error(err))
		return nil, 0, fmt.Errorf("查询失败: %w", err)
	}

	// 为每条记录填充用户的 Rating
	for i := range records {
		var userRecord model.UserRecord
		if err := s.db.Where("uid = ?", records[i].UserID).First(&userRecord).Error; err == nil {
			records[i].UserRating = userRecord.HistRating
		}
		// 如果查询失败，UserRating 保持为 nil
	}

	logger.Debug("查询所有对战记录成功",
		zap.Int("page", page),
		zap.Int64("total", total))
	return records, total, nil
}

// DissolveRoom 解散房间（仅房主可以解散等待中的房间）
func (s *BattleService) DissolveRoom(roomID, userID string) error {
	logger := utils.GetLogger()

	// 查询房间
	var room model.BattleRoom
	if err := s.db.Where("room_id = ?", roomID).First(&room).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Warn("房间不存在",
				zap.String("room_id", roomID))
			return fmt.Errorf("房间不存在")
		}
		logger.Error("查询房间失败",
			zap.String("room_id", roomID),
			zap.Error(err))
		return fmt.Errorf("查询房间失败")
	}

	// 检查是否是房主
	if room.HostID != userID {
		logger.Warn("无权解散房间",
			zap.String("room_id", roomID),
			zap.String("user_id", userID),
			zap.String("host_id", room.HostID))
		return fmt.Errorf("只有房主可以解散房间")
	}

	// 检查房间状态（允许解散等待中、对战中、已结束的房间）
	if room.Status == model.BattleStatusEnded {
		// 已结束的房间，直接删除
		if err := s.db.Delete(&room).Error; err != nil {
			logger.Error("删除已结束房间失败",
				zap.String("room_id", roomID),
				zap.Error(err))
			return fmt.Errorf("解散房间失败")
		}

		logger.Info("解散已结束房间成功",
			zap.String("room_id", roomID),
			zap.String("user_id", userID))
		return nil
	}

	if room.Status != model.BattleStatusWaiting && room.Status != model.BattleStatusInProgress {
		logger.Warn("房间状态不允许解散",
			zap.String("room_id", roomID),
			zap.Int("status", int(room.Status)))
		return fmt.Errorf("只能解散等待中或对战中的房间")
	}

	// 如果是进行中的对战，需要先结束对战
	if room.Status == model.BattleStatusInProgress {
		// 房主解散房间，判定对方获胜
		winnerID := *room.ChallengerID
		endReason := "dissolve" // 房主解散
		if err := s.endBattle(&room, winnerID, endReason); err != nil {
			logger.Error("结束对战失败",
				zap.String("room_id", roomID),
				zap.Error(err))
			return fmt.Errorf("结束对战失败: %w", err)
		}
	} else {
		// 等待中的房间，直接删除
		// 删除房间
		if err := s.db.Delete(&room).Error; err != nil {
			logger.Error("删除房间失败",
				zap.String("room_id", roomID),
				zap.Error(err))
			return fmt.Errorf("解散房间失败")
		}
	}

	logger.Info("解散房间成功",
		zap.String("room_id", roomID),
		zap.String("user_id", userID))
	return nil
}

// LeaveRoom 退出房间（任何人都可退出等待中的房间）
func (s *BattleService) LeaveRoom(roomID, userID string) error {
	logger := utils.GetLogger()

	// 查询房间
	var room model.BattleRoom
	if err := s.db.Where("room_id = ?", roomID).First(&room).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Warn("房间不存在",
				zap.String("room_id", roomID))
			return fmt.Errorf("房间不存在")
		}
		logger.Error("查询房间失败",
			zap.String("room_id", roomID),
			zap.Error(err))
		return fmt.Errorf("查询房间失败")
	}

	// 检查房间状态（只能退出等待中或已结束的房间，不能退出对战中）
	if room.Status == model.BattleStatusInProgress {
		logger.Warn("房间状态不允许退出",
			zap.String("room_id", roomID),
			zap.Int("status", int(room.Status)))
		return fmt.Errorf("对战中不能退出房间")
	}

	// 检查是否是房主
	if room.HostID == userID {
		// 房主退出：如果有挑战者，转移房主身份；否则删除房间
		if room.ChallengerID != nil {
			// 有挑战者，转移房主身份
			newHostID := *room.ChallengerID
			newHostUsername := *room.ChallengerUsername

			room.HostID = newHostID
			room.HostUsername = newHostUsername
			room.ChallengerID = nil
			room.ChallengerUsername = nil
			room.ChallengerReady = false // 清除准备状态

			if err := s.db.Save(&room).Error; err != nil {
				logger.Error("转移房主身份失败",
					zap.String("room_id", roomID),
					zap.String("user_id", userID),
					zap.Error(err))
				return fmt.Errorf("退出房间失败")
			}

			logger.Info("房主退出，房主身份已转移",
				zap.String("room_id", roomID),
				zap.String("old_host", userID),
				zap.String("new_host", newHostID))
			return nil
		} else {
			// 没有挑战者，删除房间
			if err := s.db.Delete(&room).Error; err != nil {
				logger.Error("删除房间失败",
					zap.String("room_id", roomID),
					zap.Error(err))
				return fmt.Errorf("退出房间失败")
			}

			logger.Info("房主退出，房间已删除",
				zap.String("room_id", roomID),
				zap.String("user_id", userID))
			return nil
		}
	}

	// 检查是否是挑战者
	if room.ChallengerID == nil || *room.ChallengerID != userID {
		logger.Warn("用户不在房间中",
			zap.String("room_id", roomID),
			zap.String("user_id", userID))
		return fmt.Errorf("你不在该房间中")
	}

	// 挑战者退出：清空挑战者信息
	room.ChallengerID = nil
	room.ChallengerUsername = nil
	room.ChallengerReady = false // 清除准备状态

	if err := s.db.Save(&room).Error; err != nil {
		logger.Error("退出房间失败",
			zap.String("room_id", roomID),
			zap.String("user_id", userID),
			zap.Error(err))
		return fmt.Errorf("退出房间失败")
	}

	logger.Info("挑战者退出房间成功",
		zap.String("room_id", roomID),
		zap.String("user_id", userID))
	return nil
}

// generateRoomID 生成房间号（2个大写字母+4位数字）
func (s *BattleService) generateRoomID() string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const digits = "0123456789"

	rand.Seed(time.Now().UnixNano())

	// 生成2个大写字母
	roomID := make([]byte, 6)
	for i := 0; i < 2; i++ {
		roomID[i] = letters[rand.Intn(len(letters))]
	}

	// 生成4位数字
	for i := 2; i < 6; i++ {
		roomID[i] = digits[rand.Intn(len(digits))]
	}

	return string(roomID)
}

// ResetRoom 重置房间（再来一局）
func (s *BattleService) ResetRoom(roomID, userID string) error {
	logger := utils.GetLogger()

	logger.Info("尝试重置房间",
		zap.String("room_id", roomID),
		zap.String("user_id", userID))

	// 查询房间
	var room model.BattleRoom
	err := s.db.Where("room_id = ?", roomID).First(&room).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Warn("房间不存在",
				zap.String("room_id", roomID))
			return fmt.Errorf("房间不存在")
		}
		logger.Error("查询房间失败",
			zap.String("room_id", roomID),
			zap.Error(err))
		return fmt.Errorf("查询房间失败")
	}

	// 检查用户是否在房间中
	if room.HostID != userID && (room.ChallengerID == nil || *room.ChallengerID != userID) {
		logger.Warn("用户不在房间中",
			zap.String("room_id", roomID),
			zap.String("user_id", userID))
		return fmt.Errorf("你不在该房间中")
	}

	// 检查房间状态
	// 如果房间状态是"等待中"，说明已经有人点过"再来一局"了，直接返回成功
	if room.Status == model.BattleStatusWaiting {
		logger.Info("房间已重置为等待状态",
			zap.String("room_id", roomID),
			zap.String("user_id", userID))
		return nil
	}

	// 如果房间状态不是"已结束"，则不允许重置
	if room.Status != model.BattleStatusEnded {
		logger.Warn("房间状态不允许重置",
			zap.String("room_id", roomID),
			zap.Int("current_status", int(room.Status)))
		return fmt.Errorf("只能重置已结束的对局")
	}

	// 重置房间状态为等待中（允许房主在挑战者退出后重置房间，等待新挑战者加入）
	room.Status = model.BattleStatusWaiting
	room.ProblemID = nil
	room.WinnerID = nil
	room.EndReason = nil
	room.StartTime = nil
	room.EndTime = nil
	room.ChallengerReady = false // 清除准备状态

	if err := s.db.Save(&room).Error; err != nil {
		logger.Error("重置房间失败",
			zap.String("room_id", roomID),
			zap.Error(err))
		return fmt.Errorf("重置房间失败")
	}

	logger.Info("房间重置成功",
		zap.String("room_id", roomID),
		zap.String("user_id", userID))
	return nil
}
