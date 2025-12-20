package client

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"

	"github.com/hoj/hist-oj/internal/config"
	"github.com/hoj/hist-oj/internal/utils"
)

// HOJ API响应结构
type CommonResult struct {
	Code    int         `json:"code"`
	Message string      `json:"msg"`
	Data    interface{} `json:"data"`
}

// ContestRankDTO 比赛排名请求参数
type ContestRankDTO struct {
	CID              int64   `json:"cid"`
	CurrentPage      int     `json:"currentPage"`
	Limit            int     `json:"limit"`
	ForceRefresh     bool    `json:"forceRefresh"`
	RemoveStar       bool    `json:"removeStar"`
	ConcernedList    []string `json:"concernedList"`
	Keyword          *string  `json:"keyword"`
	ContainsEnd      bool    `json:"containsEnd"`
}

// ContestRankRecord 比赛排名记录
type ContestRankRecord struct {
	UID      string `json:"uid"`
	Username string `json:"username"`
	Rank     int    `json:"rank"`
	AC       int    `json:"ac"`
	TotalTime int64 `json:"totalTime"`
	TotalScore int `json:"totalScore"` // OI赛制
}

// ContestRankResponse 比赛排名响应
type ContestRankResponse struct {
	Total   int                 `json:"total"`
	Records []ContestRankRecord `json:"records"`
}

// ContestInfo 比赛信息
type ContestInfo struct {
	ID        int64     `json:"id"`
	Type      int       `json:"type"`
	IsRating  bool      `json:"isRating"`
	Title     string    `json:"title"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
	Status    int       `json:"status"`
}

var HojAPIClient *resty.Client

func InitHojAPIClient(cfg *config.HojAPIConfig) {
	client := resty.New()
	client.SetBaseURL(cfg.BaseURL)
	client.SetTimeout(60 * time.Second)      // 默认60秒超时
	client.SetRetryCount(3)                  // 默认重试3次
	client.SetRetryWaitTime(2 * time.Second) // 默认重试间隔2秒
	client.SetRetryMaxWaitTime(6 * time.Second)

	HojAPIClient = client
	utils.GetLogger().Info("HOJ API client initialized", zap.String("base_url", cfg.BaseURL))
}

// GetContestInfo 获取比赛信息（从数据库直接读取，避免认证问题）
func GetContestInfo(contestID int64) (*ContestInfo, error) {
	logger := utils.GetLogger()
	logger.Debug("从数据库获取比赛信息", zap.Int64("contest_id", contestID))

	// 从数据库读取比赛信息
	var contest struct {
		ID        int64     `gorm:"column:id"`
		Type      int       `gorm:"column:type"`
		IsRating  bool      `gorm:"column:is_rating"`
		Title     string    `gorm:"column:title"`
		StartTime time.Time `gorm:"column:start_time"`
		EndTime   time.Time `gorm:"column:end_time"`
		Status    int       `gorm:"column:status"`
	}

	if err := DB.Table("contest").Where("id = ?", contestID).First(&contest).Error; err != nil {
		logger.Error("从数据库获取比赛信息失败",
			zap.Int64("contest_id", contestID),
			zap.Error(err))
		return nil, fmt.Errorf("获取比赛信息失败: %w", err)
	}

	contestInfo := &ContestInfo{
		ID:        contest.ID,
		Type:      contest.Type,
		IsRating:  contest.IsRating,
		Title:     contest.Title,
		StartTime: contest.StartTime,
		EndTime:   contest.EndTime,
		Status:    contest.Status,
	}

	logger.Debug("获取比赛信息成功",
		zap.Int64("contest_id", contestID),
		zap.Bool("is_rating", contest.IsRating))
	return contestInfo, nil
}

// GetContestRank 获取比赛排名（使用外榜接口，无需认证）
func GetContestRank(req ContestRankDTO) (*ContestRankResponse, error) {
	logger := utils.GetLogger()
	logger.Debug("使用外榜接口获取比赛排名", zap.Int64("contest_id", req.CID))

	// 使用外榜接口，无需认证
	return GetContestOutsideScoreboard(req)
}

// GetContestOutsideScoreboard 获取比赛外榜（无需认证）
func GetContestOutsideScoreboard(req ContestRankDTO) (*ContestRankResponse, error) {
	logger := utils.GetLogger()
	var result CommonResult
	var rankResponse ContestRankResponse

	logger.Debug("调用HOJ API获取比赛外榜", zap.Int64("contest_id", req.CID))
	resp, err := HojAPIClient.R().
		SetBody(req).
		SetResult(&result).
		Post("/api/get-contest-outside-scoreboard")

	if err != nil {
		logger.Error("HOJ API调用失败", 
			zap.Int64("contest_id", req.CID),
			zap.String("endpoint", "/api/get-contest-outside-scoreboard"),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get contest outside scoreboard: %w", err)
	}

	if resp.StatusCode() != 200 || result.Code != 200 {
		logger.Warn("HOJ API返回错误", 
			zap.Int64("contest_id", req.CID),
			zap.Int("http_status", resp.StatusCode()),
			zap.Int("api_code", result.Code),
			zap.String("message", result.Message))
		return nil, fmt.Errorf("API error: code=%d, msg=%s", result.Code, result.Message)
	}

	// 将data转换为ContestRankResponse
	dataBytes, err := json.Marshal(result.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal scoreboard data: %w", err)
	}

	if err := json.Unmarshal(dataBytes, &rankResponse); err != nil {
		logger.Error("解析比赛外榜失败", 
			zap.Int64("contest_id", req.CID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to unmarshal scoreboard response: %w", err)
	}

	logger.Debug("获取比赛外榜成功", 
		zap.Int64("contest_id", req.CID),
		zap.Int("total", rankResponse.Total),
		zap.Int("records", len(rankResponse.Records)))
	return &rankResponse, nil
}

