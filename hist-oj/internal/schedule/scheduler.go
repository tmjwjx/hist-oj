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
	// 每N分钟检查一次
	interval := s.config.CheckInterval
	if interval < 1 {
		interval = 5
	}

	_, err := s.cron.AddFunc(fmt.Sprintf("*/%d * * * *", interval), s.checkAndCalculateRating)
	if err != nil {
		return err
	}

	s.cron.Start()
	utils.GetLogger().Info("定时任务已启动", zap.Int("interval_minutes", interval))

	// 启动时立即执行一次
	go s.checkAndCalculateRating()

	return nil
}

// Stop 停止定时任务
func (s *Scheduler) Stop() {
	s.cron.Stop()
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

