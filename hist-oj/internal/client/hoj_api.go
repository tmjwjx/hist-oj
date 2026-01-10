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

// GetUserACProblems 获取用户AC的题目列表
func GetUserACProblems(uid string) ([]int64, error) {
	logger := utils.GetLogger()

	// 直接从数据库查询
	type UserAcproblem struct {
		Pid      int64 `gorm:"column:pid"`
		SubmitID int64 `gorm:"column:submit_id"`
	}

	var acProblems []UserAcproblem
	if err := DB.Table("user_acproblem").
		Select("DISTINCT pid, submit_id").
		Where("uid = ?", uid).
		Order("submit_id ASC").
		Find(&acProblems).Error; err != nil {
		logger.Error("查询用户AC题目失败",
			zap.String("uid", uid),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get user AC problems: %w", err)
	}

	pids := make([]int64, 0, len(acProblems))
	for _, p := range acProblems {
		pids = append(pids, p.Pid)
	}

	logger.Debug("查询用户AC题目成功",
		zap.String("uid", uid),
		zap.Int("count", len(pids)))
	return pids, nil
}

// GetProblemList 获取题目列表（从数据库直接读取）
func GetProblemList() ([]int64, error) {
	logger := utils.GetLogger()

	type ProblemBasic struct {
		Pid int64 `gorm:"column:id"`
	}

	var problems []ProblemBasic
	if err := DB.Table("problem").
		Select("id").
		Where("auth = 1"). // 仅获取公开题目
		Find(&problems).Error; err != nil {
		logger.Error("查询题目列表失败", zap.Error(err))
		return nil, fmt.Errorf("failed to get problem list: %w", err)
	}

	pids := make([]int64, 0, len(problems))
	for _, p := range problems {
		pids = append(pids, p.Pid)
	}

	logger.Debug("查询题目列表成功", zap.Int("count", len(pids)))
	return pids, nil
}

// ProblemListItem 题目列表项（包含显示ID）
type ProblemListItem struct {
	Pid       int64  `json:"pid"`
	DisplayID string `json:"displayId"`
}

// GetProblemListWithDisplayID 获取题目列表（包含显示ID）
func GetProblemListWithDisplayID() ([]ProblemListItem, error) {
	logger := utils.GetLogger()

	type ProblemDB struct {
		Pid       int64  `gorm:"column:id"`
		ProblemID string `gorm:"column:problem_id"`
	}

	var problemsDB []ProblemDB
	if err := DB.Table("problem").
		Select("id, problem_id").
		Where("auth = 1"). // 仅获取公开题目
		Find(&problemsDB).Error; err != nil {
		logger.Error("查询题目列表失败", zap.Error(err))
		return nil, fmt.Errorf("failed to get problem list: %w", err)
	}

	problems := make([]ProblemListItem, 0, len(problemsDB))
	for _, p := range problemsDB {
		problems = append(problems, ProblemListItem{
			Pid:       p.Pid,
			DisplayID: p.ProblemID,
		})
	}

	logger.Debug("查询题目列表成功", zap.Int("count", len(problems)))
	return problems, nil
}

// GetProblemInfo 获取题目详情（从数据库直接读取）
func GetProblemInfo(pid int64) (*ProblemBasicInfo, error) {
	logger := utils.GetLogger()

	type ProblemDB struct {
		Pid       int64  `gorm:"column:id"`
		ProblemID string `gorm:"column:problem_id"`
		Title     string `gorm:"column:title"`
		Diff      int    `gorm:"column:difficulty"`
		Auth      int    `gorm:"column:auth"`
	}

	var problemDB ProblemDB
	if err := DB.Table("problem").
		Where("id = ?", pid).
		First(&problemDB).Error; err != nil {
		logger.Error("查询题目详情失败",
			zap.Int64("pid", pid),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get problem info: %w", err)
	}

	problem := &ProblemBasicInfo{
		Pid:        problemDB.Pid,
		DisplayID:  problemDB.ProblemID,
		ProblemID:  problemDB.ProblemID,  // 使用显示ID（如 0051）而不是数字ID（如 1191）
		Title:      problemDB.Title,
		Difficulty: problemDB.Diff,
		Auth:       problemDB.Auth,
	}

	logger.Debug("查询题目详情成功",
		zap.Int64("pid", pid),
		zap.String("title", problem.Title))
	return problem, nil
}

// GetProblemInfoByDisplayID 通过显示ID获取题目详情（如 0051, Z001）
func GetProblemInfoByDisplayID(displayID string) (*ProblemBasicInfo, error) {
	logger := utils.GetLogger()

	type ProblemDB struct {
		Pid       int64  `gorm:"column:id"`
		ProblemID string `gorm:"column:problem_id"`
		Title     string `gorm:"column:title"`
		Diff      int    `gorm:"column:difficulty"`
		Auth      int    `gorm:"column:auth"`
	}

	var problemDB ProblemDB
	if err := DB.Table("problem").
		Where("problem_id = ?", displayID).
		First(&problemDB).Error; err != nil {
		logger.Error("通过显示ID查询题目详情失败",
			zap.String("display_id", displayID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get problem info by display id: %w", err)
	}

	problem := &ProblemBasicInfo{
		Pid:        problemDB.Pid,
		DisplayID:  problemDB.ProblemID,
		ProblemID:  problemDB.ProblemID,
		Title:      problemDB.Title,
		Difficulty: problemDB.Diff,
		Auth:       problemDB.Auth,
	}

	logger.Debug("通过显示ID查询题目详情成功",
		zap.String("display_id", displayID),
		zap.String("title", problem.Title))
	return problem, nil
}

// ProblemBasicInfo 题目基本信息
type ProblemBasicInfo struct {
	Pid        int64  `json:"pid"`         // 数据库主键ID
	DisplayID  string `json:"displayId"`   // 显示ID，如 0019（保留兼容）
	ProblemID  string `json:"problemId"`   // 显示ID，如 0019（前端使用的字段）
	Title      string `json:"title"`
	Difficulty int    `json:"difficulty"`
	Auth       int    `json:"auth"`
}

// GetUserInfo 获取用户信息（从数据库直接读取）
func GetUserInfo(uid string) (*UserInfoBasic, error) {
	logger := utils.GetLogger()

	type UserDB struct {
		UUID     string `gorm:"column:uuid"`
		Username string `gorm:"column:username"`
		Nickname string `gorm:"column:nickname"`
		Avatar   string `gorm:"column:avatar"`
	}

	var userDB UserDB
	if err := DB.Table("user_info").
		Where("uuid = ?", uid).
		First(&userDB).Error; err != nil {
		logger.Error("查询用户信息失败",
			zap.String("uid", uid),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	user := &UserInfoBasic{
		UID:      userDB.UUID,
		Username: userDB.Username,
		Nickname: userDB.Nickname,
		Avatar:   userDB.Avatar,
	}

	logger.Debug("查询用户信息成功",
		zap.String("uid", uid),
		zap.String("username", user.Username))
	return user, nil
}

// UserInfoBasic 用户基本信息
type UserInfoBasic struct {
	UID      string `json:"uid"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

