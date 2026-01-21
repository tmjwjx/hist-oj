package service

import (
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/hoj/hist-oj/internal/client"
	"github.com/hoj/hist-oj/internal/model"
	"github.com/hoj/hist-oj/internal/utils"
)

// TrainingService 训练服务
type TrainingService struct {
	db *gorm.DB
}

// NewTrainingService 创建训练服务
func NewTrainingService() *TrainingService {
	return &TrainingService{
		db: client.GetDB(),
	}
}

// JoinTraining 用户参加训练
func (s *TrainingService) JoinTraining(trainingID uint64, uid string) (*model.TrainingParticipant, error) {
	logger := utils.GetLogger()
	db := client.GetDB()

	// 检查是否已参加
	var record model.TrainingParticipant
	err := db.Where("training_id = ? AND uid = ?", trainingID, uid).First(&record).Error
	if err == nil {
		// 已存在记录,直接返回
		logger.Info("用户已参加训练", zap.Uint64("training_id", trainingID), zap.String("uid", uid))
		return &record, nil
	}
	if err != gorm.ErrRecordNotFound {
		logger.Error("查询训练记录失败", zap.Error(err))
		return nil, err
	}

	// 创建新的参与记录
	now := time.Now()
	record = model.TrainingParticipant{
		TrainingID: trainingID,
		UID:        uid,
		Status:     "not_started",
		JoinTime:   &now,
	}

	if err := db.Create(&record).Error; err != nil {
		logger.Error("创建训练记录失败", zap.Error(err))
		return nil, err
	}

	logger.Info("用户参加训练成功", zap.Uint64("training_id", trainingID), zap.String("uid", uid))
	return &record, nil
}

// GetTrainingParticipants 获取训练参与者列表（包含进度信息）
func (s *TrainingService) GetTrainingParticipants(trainingID uint64) ([]model.TrainingParticipant, error) {
	logger := utils.GetLogger()
	db := client.GetDB()

	// 添加性能监控
	startTime := time.Now()

	// 使用 JOIN 查询一次性获取参与者和用户信息，避免 Preload 的 N+1 问题
	type ParticipantWithUser struct {
		ID         uint64
		TrainingID uint64
		UID        string
		Status     string
		JoinTime   *time.Time
		GmtCreate  time.Time
		GmtModified time.Time
		// 用户信息字段
		Username   string `gorm:"column:username"`
		Nickname   string `gorm:"column:nickname"`
		Realname   string `gorm:"column:realname"`
		Rating     int    `gorm:"column:rating"`
	}

	var participantsWithUser []ParticipantWithUser
	err := db.Table("training_participant tp").
		Select(`tp.id, tp.training_id, tp.uid, tp.status, tp.join_time, tp.gmt_create, tp.gmt_modified,
				u.username, u.nickname, u.realname,
				COALESCE(ur.hist_rating, 0) as rating`).
		Joins("LEFT JOIN user_info u ON tp.uid = u.uuid").
		Joins("LEFT JOIN user_record ur ON tp.uid = ur.uid").
		Where("tp.training_id = ?", trainingID).
		Order("tp.gmt_create ASC").
		Find(&participantsWithUser).Error

	if err != nil {
		logger.Error("查询训练参与者失败", zap.Error(err))
		return nil, err
	}

	// 转换为 model.TrainingParticipant 格式
	records := make([]model.TrainingParticipant, len(participantsWithUser))
	for i, p := range participantsWithUser {
		records[i] = model.TrainingParticipant{
			ID:         p.ID,
			TrainingID: p.TrainingID,
			UID:        p.UID,
			Status:     p.Status,
			JoinTime:   p.JoinTime,
			GmtCreate:  p.GmtCreate,
			GmtModified: p.GmtModified,
			User: &model.UserInfo{
				UUID:     p.UID,
				Username: p.Username,
				Nickname: p.Nickname,
				Realname: p.Realname,
				Rating:   p.Rating,
			},
		}
	}

	logger.Info("查询训练参与者和用户信息耗时",
		zap.Duration("duration", time.Since(startTime)),
		zap.Int("count", len(records)))

	// 查询训练的题目总数（快速 COUNT 查询）
	var totalCount int64
	db.Table("training_problem").
		Where("tid = ?", trainingID).
		Count(&totalCount)

	if totalCount == 0 {
		// 没有题目，直接返回
		for i := range records {
			records[i].Progress = &model.TrainingProgress{
				TrainingID:  trainingID,
				SolvedCount: 0,
				TotalCount:  0,
				Status:      records[i].Status,
			}
		}
		logger.Info("查询训练参与者成功（无题目）", zap.Int("count", len(records)))
		return records, nil
	}

	// 使用 user_acproblem 表快速查询每个用户的进度（与用户界面相同的方法）
	// 这个表缓存了每个用户已通过的题目，查询速度非常快
	progressStart := time.Now()
	type UserSolvedProblem struct {
		UID string `gorm:"column:uid"`
		PID uint64 `gorm:"column:pid"`
	}
	var solvedProblems []UserSolvedProblem

	// 1. 先获取该训练的所有题目ID
	var problemIDs []uint64
	db.Table("training_problem").
		Where("tid = ?", trainingID).
		Pluck("pid", &problemIDs)

	logger.Info("获取训练题目ID耗时", zap.Duration("duration", time.Since(progressStart)), zap.Int("problem_count", len(problemIDs)))

	// 2. 使用 user_acproblem 表查询参与者在这些题目上的AC记录
	queryStart := time.Now()
	// 与用户界面完全相同的查询方式
	db.Raw(`
		SELECT DISTINCT uid, pid
		FROM user_acproblem
		WHERE uid IN (?)
		AND pid IN (?)
	`, getParticipantUIDs(records), problemIDs).Scan(&solvedProblems)

	logger.Info("查询user_acproblem耗时",
		zap.Duration("duration", time.Since(queryStart)),
		zap.Int("solved_records", len(solvedProblems)),
		zap.Int("participants", len(records)))

	// 构建 map: uid -> 已解决题目数
	solvedCountMap := make(map[string]int64)
	for _, sp := range solvedProblems {
		solvedCountMap[sp.UID]++
	}

	// 为每个参与者设置进度信息（O(1) 查找）
	for i := range records {
		solvedCount := solvedCountMap[records[i].UID] // 如果不存在，默认为 0

		// 设置进度信息
		records[i].Progress = &model.TrainingProgress{
			TrainingID:  trainingID,
			SolvedCount: solvedCount,
			TotalCount:  totalCount,
			Status:      records[i].Status,
		}
	}

	logger.Info("查询训练参与者成功",
		zap.Int("count", len(records)),
		zap.Duration("total_duration", time.Since(startTime)))
	return records, nil
}

// GetTrainingParticipantByUser 获取用户在某个训练的记录
func (s *TrainingService) GetTrainingParticipantByUser(trainingID uint64, uid string) (*model.TrainingParticipant, error) {
	db := client.GetDB()

	var record model.TrainingParticipant
	err := db.Where("training_id = ? AND uid = ?", trainingID, uid).
		Preload("User").
		First(&record).Error

	if err != nil {
		return nil, err
	}

	return &record, nil
}

// UpdateTrainingStatus 更新训练状态
func (s *TrainingService) UpdateTrainingStatus(recordID uint64, status string) error {
	db := client.GetDB()

	result := db.Model(&model.TrainingParticipant{}).
		Where("id = ?", recordID).
		Update("status", status)

	if result.Error != nil {
		return result.Error
	}

	return nil
}

// UpdateTrainingProgressBySubmission 根据提交情况自动更新训练进度状态
func (s *TrainingService) UpdateTrainingProgressBySubmission(trainingID uint64, uid string) error {
	logger := utils.GetLogger()
	db := client.GetDB()

	// 查询训练的题目总数
	var trainingProblemCount int64
	db.Table("training_problem").
		Where("tid = ?", trainingID).
		Count(&trainingProblemCount)

	if trainingProblemCount == 0 {
		return nil // 没有题目，不需要更新
	}

	// 查询用户在该训练中已解决的题目数量
	// training_record表记录了提交，需要关联judge表来判断是否通过(status=0)
	var solvedCount int64
	db.Table("training_record tr").
		Select("DISTINCT tr.pid").
		Joins("INNER JOIN judge j ON tr.submit_id = j.submit_id").
		Where("tr.tid = ? AND tr.uid = ? AND j.status = 0", trainingID, uid).
		Count(&solvedCount)

	// 获取参与记录
	var record model.TrainingParticipant
	err := db.Where("training_id = ? AND uid = ?", trainingID, uid).First(&record).Error
	if err != nil {
		return err
	}

	// 根据完成度自动更新状态
	// 状态逻辑:
	// - not_started: 还没有提交任何题目
	// - in_progress: 已经提交过至少一次,但还没完成所有题目
	// - completed: 完成了所有题目(通过所有题目)

	// 首先查询用户是否有提交记录(不管是否通过)
	var submissionCount int64
	db.Table("training_record").
		Where("tid = ? AND uid = ?", trainingID, uid).
		Count(&submissionCount)

	var newStatus string
	if submissionCount == 0 {
		// 没有任何提交记录
		newStatus = "not_started"
	} else if solvedCount < trainingProblemCount {
		// 有提交但还没全部通过
		newStatus = "in_progress"
	} else {
		// 通过了所有题目
		newStatus = "completed"
	}

	// 如果状态需要更新
	if record.Status != newStatus {
		if err := db.Model(&record).Update("status", newStatus).Error; err != nil {
			logger.Error("更新训练状态失败",
				zap.Uint64("training_id", trainingID),
				zap.String("uid", uid),
				zap.String("old_status", record.Status),
				zap.String("new_status", newStatus),
				zap.Error(err))
			return err
		}
		logger.Info("训练状态已自动更新",
			zap.Uint64("training_id", trainingID),
			zap.String("uid", uid),
			zap.String("status", newStatus),
			zap.Int64("solved", solvedCount),
			zap.Int64("total", trainingProblemCount))
	}

	return nil
}

// GetUserTrainingProgress 获取用户在所有训练中的进度
func (s *TrainingService) GetUserTrainingProgress(uid string) ([]model.TrainingProgress, error) {
	logger := utils.GetLogger()
	db := client.GetDB()

	// 查询用户参加的所有训练
	var participantRecords []model.TrainingParticipant
	err := db.Where("uid = ?", uid).Find(&participantRecords).Error
	if err != nil {
		logger.Error("查询用户训练记录失败", zap.Error(err))
		return nil, err
	}

	// 构建结果
	var progressList []model.TrainingProgress
	for _, record := range participantRecords {
		// 查询该训练的题目总数
		var totalCount int64
		db.Table("training_problem").
			Where("tid = ?", record.TrainingID).
			Count(&totalCount)

		// 查询用户在该训练中已解决的题目数量
		var solvedCount int64
		db.Table("training_record tr").
			Select("DISTINCT tr.pid").
			Joins("INNER JOIN judge j ON tr.submit_id = j.submit_id").
			Where("tr.tid = ? AND tr.uid = ? AND j.status = 0", record.TrainingID, uid).
			Count(&solvedCount)

		progressList = append(progressList, model.TrainingProgress{
			TrainingID:  record.TrainingID,
			SolvedCount: solvedCount,
			TotalCount:  totalCount,
			Status:      record.Status,
		})
	}

	logger.Info("查询用户训练进度成功", zap.String("uid", uid), zap.Int("count", len(progressList)))
	return progressList, nil
}

// getParticipantUIDs 提取参与者UID列表
func getParticipantUIDs(records []model.TrainingParticipant) []string {
	uids := make([]string, len(records))
	for i, record := range records {
		uids[i] = record.UID
	}
	return uids
}
