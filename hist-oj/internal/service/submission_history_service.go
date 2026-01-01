package service

import (
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/hoj/hist-oj/internal/model"
	"github.com/hoj/hist-oj/internal/utils"
)

// SubmissionHistoryService 提交历史服务
type SubmissionHistoryService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewSubmissionHistoryService 创建提交历史服务
func NewSubmissionHistoryService(db *gorm.DB) *SubmissionHistoryService {
	return &SubmissionHistoryService{
		db:     db,
		logger: utils.GetLogger(),
	}
}

// Create 创建提交历史记录
func (s *SubmissionHistoryService) Create(history *model.SubmissionHistory) error {
	if err := s.db.Create(history).Error; err != nil {
		s.logger.Error("创建提交历史失败", zap.Error(err))
		return fmt.Errorf("创建提交历史失败: %w", err)
	}
	return nil
}

// GetByPIDAndCID 根据题目ID和比赛ID查询历史记录
func (s *SubmissionHistoryService) GetByPIDAndCID(pid, cid string, limit int) ([]*model.SubmissionHistory, error) {
	var histories []*model.SubmissionHistory

	query := s.db.Where("pid = ? AND cid = ?", pid, cid).
		Order("id DESC").
		Limit(limit)

	if err := query.Find(&histories).Error; err != nil {
		s.logger.Error("查询提交历史失败",
			zap.String("pid", pid),
			zap.String("cid", cid),
			zap.Error(err))
		return nil, fmt.Errorf("查询提交历史失败: %w", err)
	}

	return histories, nil
}

// GetByUsername 根据用户名查询历史记录
func (s *SubmissionHistoryService) GetByUsername(username string, limit int) ([]*model.SubmissionHistory, error) {
	var histories []*model.SubmissionHistory

	query := s.db.Where("username = ?", username).
		Order("id DESC").
		Limit(limit)

	if err := query.Find(&histories).Error; err != nil {
		s.logger.Error("查询用户提交历史失败",
			zap.String("username", username),
			zap.Error(err))
		return nil, fmt.Errorf("查询用户提交历史失败: %w", err)
	}

	return histories, nil
}
