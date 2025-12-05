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

// GetContestInfo 获取比赛信息
func GetContestInfo(contestID int64) (*ContestInfo, error) {
	logger := utils.GetLogger()
	var result CommonResult
	var contestInfo ContestInfo

	logger.Debug("调用HOJ API获取比赛信息", zap.Int64("contest_id", contestID))
	resp, err := HojAPIClient.R().
		SetQueryParam("cid", fmt.Sprintf("%d", contestID)).
		SetResult(&result).
		Get("/api/get-contest-info")

	if err != nil {
		logger.Error("HOJ API调用失败", 
			zap.Int64("contest_id", contestID),
			zap.String("endpoint", "/api/get-contest-info"),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get contest info: %w", err)
	}

	if resp.StatusCode() != 200 || result.Code != 200 {
		logger.Warn("HOJ API返回错误", 
			zap.Int64("contest_id", contestID),
			zap.Int("http_status", resp.StatusCode()),
			zap.Int("api_code", result.Code),
			zap.String("message", result.Message))
		return nil, fmt.Errorf("API error: code=%d, msg=%s", result.Code, result.Message)
	}

	// 将data转换为ContestInfo
	dataBytes, err := json.Marshal(result.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal contest data: %w", err)
	}

	if err := json.Unmarshal(dataBytes, &contestInfo); err != nil {
		logger.Error("解析比赛信息失败", 
			zap.Int64("contest_id", contestID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to unmarshal contest info: %w", err)
	}

	logger.Debug("获取比赛信息成功", 
		zap.Int64("contest_id", contestID),
		zap.String("title", contestInfo.Title))
	return &contestInfo, nil
}

// GetContestRank 获取比赛排名
func GetContestRank(req ContestRankDTO) (*ContestRankResponse, error) {
	logger := utils.GetLogger()
	var result CommonResult
	var rankResponse ContestRankResponse

	logger.Debug("调用HOJ API获取比赛排名", zap.Int64("contest_id", req.CID))
	resp, err := HojAPIClient.R().
		SetBody(req).
		SetResult(&result).
		Post("/api/get-contest-rank")

	if err != nil {
		logger.Error("HOJ API调用失败", 
			zap.Int64("contest_id", req.CID),
			zap.String("endpoint", "/api/get-contest-rank"),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get contest rank: %w", err)
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
		return nil, fmt.Errorf("failed to marshal rank data: %w", err)
	}

	if err := json.Unmarshal(dataBytes, &rankResponse); err != nil {
		logger.Error("解析比赛排名失败", 
			zap.Int64("contest_id", req.CID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to unmarshal rank response: %w", err)
	}

	logger.Debug("获取比赛排名成功", 
		zap.Int64("contest_id", req.CID),
		zap.Int("total", rankResponse.Total),
		zap.Int("records", len(rankResponse.Records)))
	return &rankResponse, nil
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

