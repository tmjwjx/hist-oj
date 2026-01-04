package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/hoj/hist-oj/internal/utils"
)

const (
	BaseURL              = "http://bingoj.cn"
	APILogin             = BaseURL + "/api/login"
	APISubmit            = BaseURL + "/api/submit-problem-judge"
	APIResult            = BaseURL + "/api/get-submission-detail"
	APIProblemNormal     = BaseURL + "/api/get-problem-detail"
	APIProblemContest    = BaseURL + "/api/get-contest-problem-details"
)

// StatusMap 状态映射表
var StatusMap = map[int]string{
	0:  "答案正确",
	-1: "答案错误",
	-2: "编译错误",
	-3: "格式错误",
	1:  "时间超限",
	2:  "内存超限",
	3:  "运行错误",
	4:  "NO",
	5:  "系统错误",
	6:  "等待中",
	7:  "判题中",
	8:  "部分正确",
	9:  "提交中",
	10: "提交失败",
}

// BingoJClient BingoJ OJ 客户端
type BingoJClient struct {
	httpClient *http.Client
	token      string
	logger     *zap.Logger
}

// NewBingoJClient 创建新的 BingoJ 客户端
func NewBingoJClient() *BingoJClient {
	return &BingoJClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: utils.GetLogger(),
	}
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Login 登录到 BingoJ OJ
func (c *BingoJClient) Login(username, password string) error {
	c.logger.Info("开始登录 BingoJ OJ", zap.String("username", username))

	reqBody := LoginRequest{
		Username: username,
		Password: password,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("序列化登录请求失败: %w", err)
	}

	req, err := http.NewRequest("POST", APILogin, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建登录请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("登录请求失败", zap.Error(err))
		return fmt.Errorf("登录请求失败: %w", err)
	}
	defer resp.Body.Close()

	c.logger.Debug("登录响应状态码", zap.Int("status_code", resp.StatusCode))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("登录失败，状态码: %d", resp.StatusCode)
	}

	token := resp.Header.Get("Authorization")
	if token == "" {
		c.logger.Error("未获取到 Token")
		return fmt.Errorf("未获取到 Token")
	}

	c.token = token
	c.logger.Info("登录成功", zap.String("username", username), zap.String("token", token[:20]+"..."))
	return nil
}

// ProblemDetail 题目详情
type ProblemDetail struct {
	ProblemId   string `json:"problemId"` // 数据库中的真实题目ID
	Title       string `json:"title"`
	Description string `json:"description"`
	Input       string `json:"input"`
	Output      string `json:"output"`
	Examples    string `json:"examples"`
	Hint        string `json:"hint"`
}

// GetProblemDetail 获取题目详情（普通模式）
func (c *BingoJClient) GetProblemDetail(problemID string) (*ProblemDetail, error) {
	if c.token == "" {
		c.logger.Error("未登录，无法获取题目详情")
		return nil, fmt.Errorf("未登录")
	}

	c.logger.Info("获取题目详情", zap.String("problem_id", problemID))

	url := fmt.Sprintf("%s?problemId=%s", APIProblemNormal, problemID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Authorization", c.token)
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("获取题目详情请求失败", zap.Error(err))
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var result struct {
		Data struct {
			Problem *ProblemDetail `json:"problem"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		c.logger.Error("解析题目详情失败", zap.Error(err))
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if result.Data.Problem == nil {
		return nil, fmt.Errorf("题目不存在")
	}

	// 调试：打印完整的 problem 对象
	c.logger.Debug("题目详情完整响应",
		zap.String("problemId", result.Data.Problem.ProblemId),
		zap.String("title", result.Data.Problem.Title),
		zap.String("response", string(body)))

	return result.Data.Problem, nil
}

// GetContestProblemDetail 获取题目详情（比赛模式）
func (c *BingoJClient) GetContestProblemDetail(displayID, contestID string) (*ProblemDetail, error) {
	if c.token == "" {
		return nil, fmt.Errorf("未登录")
	}

	url := fmt.Sprintf("%s?displayId=%s&cid=%s", APIProblemContest, displayID, contestID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Authorization", c.token)
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var result struct {
		Data struct {
			Problem *ProblemDetail `json:"problem"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if result.Data.Problem == nil {
		return nil, fmt.Errorf("题目不存在")
	}

	return result.Data.Problem, nil
}

// SubmitRequest 提交请求
type SubmitRequest struct {
	PID      string      `json:"pid"`
	CID      int         `json:"cid"`
	Language string      `json:"language"`
	Code     string      `json:"code"`
	TID      interface{} `json:"tid"`
	GID      interface{} `json:"gid"`
	IsRemote bool        `json:"isRemote"`
}

// SubmitCode 提交代码
func (c *BingoJClient) SubmitCode(pid string, cid int, language, code string) (string, error) {
	if c.token == "" {
		c.logger.Error("未登录，无法提交代码")
		return "", fmt.Errorf("未登录")
	}

	c.logger.Info("提交代码", zap.String("pid", pid), zap.Int("cid", cid), zap.String("language", language))

	reqBody := SubmitRequest{
		PID:      pid,
		CID:      cid,
		Language: language,
		Code:     code,
		TID:      nil,
		GID:      nil,
		IsRemote: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("序列化提交请求失败: %w", err)
	}

	req, err := http.NewRequest("POST", APISubmit, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建提交请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.token)
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("提交代码请求失败", zap.Error(err))
		return "", fmt.Errorf("提交请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	var result struct {
		Data struct {
			SubmitID interface{} `json:"submitId"`
		} `json:"data"`
		Msg string `json:"msg"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		c.logger.Error("解析提交响应失败", zap.Error(err), zap.String("response", string(body)))
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	// 处理 submitId 可能是 string 或 number 的情况
	var submitID string
	switch v := result.Data.SubmitID.(type) {
	case string:
		submitID = v
	case float64:
		submitID = fmt.Sprintf("%.0f", v)
	case int:
		submitID = fmt.Sprintf("%d", v)
	case int64:
		submitID = fmt.Sprintf("%d", v)
	}

	if submitID == "" {
		c.logger.Error("提交失败", zap.String("msg", result.Msg))
		return "", fmt.Errorf("提交失败: %s", result.Msg)
	}

	c.logger.Info("代码提交成功", zap.String("submit_id", submitID))
	return submitID, nil
}

// SubmissionResult 提交结果
type SubmissionResult struct {
	Status int    `json:"status"`
	Time   int    `json:"time"`
	Memory int    `json:"memory"`
	Score  int    `json:"score"`
}

// GetSubmissionResult 获取提交结果
func (c *BingoJClient) GetSubmissionResult(submitID string) (*SubmissionResult, error) {
	if c.token == "" {
		c.logger.Error("未登录，无法获取提交结果")
		return nil, fmt.Errorf("未登录")
	}

	c.logger.Debug("查询提交结果", zap.String("submit_id", submitID))

	url := fmt.Sprintf("%s?submitId=%s", APIResult, submitID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Authorization", c.token)
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("查询提交结果请求失败", zap.Error(err))
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var result struct {
		Data struct {
			Submission *SubmissionResult `json:"submission"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		c.logger.Error("解析提交结果失败", zap.Error(err), zap.String("response", string(body)))
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if result.Data.Submission == nil {
		c.logger.Debug("提交记录不存在，可能还在处理中")
		return nil, fmt.Errorf("提交记录不存在")
	}

	c.logger.Debug("查询提交结果成功", zap.Int("status", result.Data.Submission.Status))
	return result.Data.Submission, nil
}

// GetStatusText 获取状态文本
func GetStatusText(statusCode int) string {
	if text, ok := StatusMap[statusCode]; ok {
		return text
	}
	return "未知状态"
}
