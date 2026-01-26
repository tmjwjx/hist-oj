package schedule

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/hoj/hist-oj/internal/client"
	"github.com/hoj/hist-oj/internal/config"
	"github.com/hoj/hist-oj/internal/model"
	"github.com/hoj/hist-oj/internal/service"
	"github.com/hoj/hist-oj/internal/utils"
)

type Scheduler struct {
	ratingService *service.RatingService
	config        *config.RatingConfig
	cron          *cron.Cron
}

func NewScheduler(ratingService *service.RatingService, cfg *config.RatingConfig) *Scheduler {
	return &Scheduler{
		ratingService: ratingService,
		config:        cfg,
		cron:          cron.New(),
	}
}

// Start 启动定时任务
func (s *Scheduler) Start() error {
	// 1. 每N分钟检查一次比赛rating计算任务
	interval := s.config.CheckInterval
	if interval < 1 {
		interval = 5
	}

	_, err := s.cron.AddFunc(fmt.Sprintf("*/%d * * * *", interval), s.checkAndCalculateRating)
	if err != nil {
		return err
	}

	// 2. 每分钟检查一次考试超时自动收卷
	_, err = s.cron.AddFunc("*/1 * * * *", s.collectExpiredExams)
	if err != nil {
		return err
	}

	s.cron.Start()
	utils.GetLogger().Info("定时任务已启动",
		zap.Int("rating_interval_minutes", interval),
		zap.Int("exam_collector_interval_minutes", 1))

	// 不再在启动时立即执行，避免与定时任务冲突
	// 如果需要立即检查，可以手动调用 TriggerScheduler 或等待下一个定时周期

	return nil
}

// Stop 停止定时任务
func (s *Scheduler) Stop() {
	s.cron.Stop()
}

// TriggerCheck 手动触发检查（公开方法）
func (s *Scheduler) TriggerCheck() {
	s.checkAndCalculateRating()
}

// checkAndCalculateRating 检查并计算rating
func (s *Scheduler) checkAndCalculateRating() {
	// 添加panic恢复，防止定时任务崩溃影响服务
	defer func() {
		if r := recover(); r != nil {
			logger := utils.GetLogger()
			logger.Error("定时任务执行时发生panic", zap.Any("panic", r))
		}
	}()

	logger := utils.GetLogger()
	logger.Info("开始检查需要计算rating的比赛")

	// 查询已结束但未计算rating的比赛
	now := time.Now()
	
	db := client.GetDB()
	
	// 先查询所有已结束的比赛
	var contests []model.Contest
	if err := db.Where("status = ? AND end_time <= ?", 1, now).
		Order("end_time ASC").
		Find(&contests).Error; err != nil {
		logger.Error("查询比赛失败", 
			zap.Time("current_time", now),
			zap.Error(err))
		return
	}
	
	if len(contests) == 0 {
		logger.Debug("没有已结束的比赛")
		return
	}
	
	logger.Debug("查询到已结束的比赛", zap.Int("count", len(contests)))

	// 过滤出需要计算rating的比赛
	var needCalculate []model.Contest
	for _, contest := range contests {
		var status model.ContestRatingStatus
		if err := db.Where("contest_id = ?", contest.ID).First(&status).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// 如果没有状态记录，默认需要计算
				logger.Debug("比赛没有rating状态记录，需要计算", 
					zap.Uint64("contest_id", contest.ID))
				needCalculate = append(needCalculate, contest)
			} else {
				// 查询失败，记录错误但继续处理其他比赛
				logger.Warn("查询比赛rating状态失败，跳过该比赛", 
					zap.Uint64("contest_id", contest.ID),
					zap.Error(err))
			}
			continue
		}
		
		// 检查是否为计分比赛且未计算
		if status.IsRated && !status.RatingCalculated {
			needCalculate = append(needCalculate, contest)
		} else {
			logger.Debug("比赛不需要计算rating", 
				zap.Uint64("contest_id", contest.ID),
				zap.Bool("is_rated", status.IsRated),
				zap.Bool("rating_calculated", status.RatingCalculated))
		}
	}
	
	contests = needCalculate

	if len(contests) == 0 {
		logger.Debug("没有需要计算rating的比赛")
		return
	}

	logger.Info("发现需要计算rating的比赛", zap.Int("count", len(contests)))

	for _, contest := range contests {
		if s.ratingService.CanCalculateRating(int64(contest.ID)) {
			logger.Info("开始计算比赛rating", zap.Uint64("contest_id", contest.ID))
			if _, err := s.ratingService.CalculateContestRating(int64(contest.ID)); err != nil {
				logger.Error("计算比赛rating失败", 
					zap.Uint64("contest_id", contest.ID),
					zap.Error(err))
			} else {
				logger.Info("比赛rating计算完成", zap.Uint64("contest_id", contest.ID))
			}
		}
	}
}

// collectExpiredExams 收集超时的考试并强制收卷
func (s *Scheduler) collectExpiredExams() {
	// 添加panic恢复，防止定时任务崩溃影响服务
	defer func() {
		if r := recover(); r != nil {
			logger := utils.GetLogger()
			logger.Error("考试收卷定时任务执行时发生panic", zap.Any("panic", r))
		}
	}()

	logger := utils.GetLogger()
	logger.Debug("开始检查超时的考试")

	db := client.GetDB()
	now := time.Now()

	// 1. 查询所有进行中的考试模式作业（已开始且未结束）
	var homeworks []model.ClassroomHomework
	if err := db.Where("is_exam_mode = ? AND end_time > ?", 1, now).
		Find(&homeworks).Error; err != nil {
		logger.Error("查询考试作业失败", zap.Error(err))
		return
	}

	if len(homeworks) == 0 {
		logger.Debug("没有进行中的考试作业")
		return
	}

	logger.Debug("查询到进行中的考试作业", zap.Int("count", len(homeworks)))

	totalForcedCount := 0

	// 2. 对每个作业，检查是否有学生超时未交卷
	for _, homework := range homeworks {
		// 计算该作业需要强制收卷的学生：
		// 条件：exam_start_time + exam_duration < NOW
		//       且 is_officially_submitted = 0
		//
		// 注意：这里使用 SQL 计算，避免时区问题
		// TIMESTAMPADD(MINUTE, exam_duration, exam_start_time) < NOW()

		sql := `
			UPDATE homework_submit
			SET is_officially_submitted = 1,
				is_forced_submit = 1,
				exam_end_time = ?
			WHERE homework_id = ?
			  AND is_officially_submitted = 0
			  AND exam_start_time IS NOT NULL
			  AND TIMESTAMPADD(MINUTE, ?, exam_start_time) < ?
		`

		result := db.Exec(sql, now, homework.ID, homework.ExamDuration, now)

		if result.Error != nil {
			logger.Error("批量强制收卷失败",
				zap.Uint64("homeworkId", homework.ID),
				zap.Error(result.Error))
			continue
		}

		// 检查是否真的更新了记录
		if result.RowsAffected == 0 {
			continue
		}

		forcedCount := int(result.RowsAffected)
		totalForcedCount += forcedCount

		logger.Info("定时任务：批量强制收卷",
			zap.Uint64("homeworkId", homework.ID),
			zap.String("homeworkTitle", homework.Title),
			zap.Int("forcedCount", forcedCount))

		// 3. 记录违规日志（针对被收卷的学生）
		// 查询刚才被更新的学生（exam_end_time = now）
		var submits []model.HomeworkSubmit
		if err := db.Where("homework_id = ? AND is_forced_submit = 1 AND exam_end_time = ?",
			homework.ID, now).
			Find(&submits).Error; err != nil {
			logger.Error("查询被收卷学生失败",
				zap.Uint64("homeworkId", homework.ID),
				zap.Error(err))
			continue
		}

		// 批量创建违规日志
		var violations []model.ExamViolationLog
		for _, submit := range submits {
			violations = append(violations, model.ExamViolationLog{
				HomeworkID:    homework.ID,
				UID:           submit.UID,
				ViolationType: "forced_submit",
				Description:   "考试超时系统自动收卷（定时任务）",
			})
		}

		if len(violations) > 0 {
			if err := db.Create(&violations).Error; err != nil {
				logger.Warn("批量记录违规日志失败",
					zap.Uint64("homeworkId", homework.ID),
					zap.Int("count", len(violations)),
					zap.Error(err))
			} else {
				logger.Info("已记录自动收卷违规日志",
					zap.Uint64("homeworkId", homework.ID),
					zap.Int("count", len(violations)))
			}
		}
	}

	if totalForcedCount > 0 {
		logger.Info("考试收卷定时任务完成",
			zap.Int("totalForcedCount", totalForcedCount))
	}
}

