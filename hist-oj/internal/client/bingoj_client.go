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
	BaseURL             = "https://bingoj.cn"
	APILogin            = BaseURL + "/api/login"
	APISubmit           = BaseURL + "/api/submit-problem-judge"
	APIResult           = BaseURL + "/api/get-submission-detail"
	APISubmissionList   = BaseURL + "/api/get-submission-list"
	APIProblemNormal    = BaseURL + "/api/get-problem-detail"
	APIProblemContest   = BaseURL + "/api/get-contest-problem-details"
	APIProblemAdmin     = BaseURL + "/api/admin/problem"
	APIProblemAdminList = BaseURL + "/api/admin/problem/get-problem-list"
)

// StatusMap 状态映射表（与 HOJ 官方标准一致）
// 参考：hoj-vue/src/common/constants.js - JUDGE_STATUS
var StatusMap = map[int]string{
	// HOJ 标准状态码
	-10: "未提交",  // NS - Not Submitted
	-5:  "结果未知", // SNR - Submitted Unknown Result
	-4:  "已取消",  // CA - Cancelled
	-3:  "格式错误", // PE - Presentation Error
	-2:  "编译错误", // CE - Compile Error
	-1:  "答案错误", // WA - Wrong Answer
	0:   "答案正确", // AC - Accepted
	1:   "时间超限", // TLE - Time Limit Exceeded
	2:   "内存超限", // MLE - Memory Limit Exceeded
	3:   "运行错误", // RE - Runtime Error
	4:   "系统错误", // SE - System Error（修正：原来是 "NO"）
	8:   "部分正确", // PAC - Partial Accepted
	10:  "提交失败", // SF - Submitted Failed

	// BingOJ 特有状态码（判题中状态）
	6: "等待中", // Pending
	7: "判题中", // Judging
	9: "提交中", // Submitting
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

// SetToken 直接设置 token（用于已登录用户）
func (c *BingoJClient) SetToken(token string) {
	c.token = token
	c.logger.Info("设置 Token 成功", zap.String("token", token[:20]+"..."))
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
	ID            int64  `json:"id"`        // 数据库主键ID（HOJ样例测试需要这个）
	ProblemId     string `json:"problemId"` // 显示ID（如 "0001"）
	Title         string `json:"title"`
	Difficulty    int    `json:"difficulty"` // 难度（0-5）
	Description   string `json:"description"`
	Input         string `json:"input"`
	Output        string `json:"output"`
	Examples      string `json:"examples"`
	Hint          string `json:"hint"`
	JudgeMode     string `json:"judgeMode"`     // 判题模式
	JudgeCaseMode string `json:"judgeCaseMode"` // 测试点评测模式（default/ergodic_without_error/...）
	TimeLimit     int64  `json:"timeLimit"`     // 时间限制（ms）
	MemoryLimit   int64  `json:"memoryLimit"`   // 内存限制（MB）
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
	req.Header.Set("Origin", BaseURL)
	req.Header.Set("Referer", BaseURL+"/")
	req.Header.Set("Url-Type", "general")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("提交代码请求失败", zap.Error(err))
		return "", fmt.Errorf("提交请求失败: %w", err)
	}
	defer resp.Body.Close()

	c.logger.Debug("提交响应状态码", zap.Int("status_code", resp.StatusCode))

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	// 打印请求和响应详情用于调试
	c.logger.Debug("提交代码请求详情",
		zap.String("url", APISubmit),
		zap.String("request_body", string(jsonData)),
		zap.Int("response_status", resp.StatusCode),
		zap.String("response_body", string(body)))

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
	Status       int    `json:"status"`
	Time         int    `json:"time"`
	Memory       int    `json:"memory"`
	Score        int    `json:"score"`
	ErrorMessage string `json:"errorMessage"` // 新增：错误信息（RE的具体原因等）
}

// SubmissionListItem 提交列表项
type SubmissionListItem struct {
	SubmitID     interface{} `json:"submitId"`     // 可能是 string 或 number
	ProblemID    string      `json:"problemId"`    // 题目ID
	DisplayID    string      `json:"displayId"`    // 显示ID
	ProblemTitle string      `json:"problemTitle"` // 题目标题
	Result       int         `json:"result"`       // 评测结果: 0=AC, -1=WA, -2=CE等
	Status       int         `json:"status"`       // 状态: 0=失败, 1=成功
	Username     string      `json:"username"`     // 用户名
	UID          string      `json:"uid"`          // 用户ID
	SubmitTime   string      `json:"submitTime"`   // 提交时间
	Language     string      `json:"language"`     // 编程语言
	JudgeTime    string      `json:"judgeTime"`    // 判题时间
	Time         int         `json:"time"`         // 运行时间(ms)
	Memory       int         `json:"memory"`       // 内存占用(KB)
	CELInfo      string      `json:"celInfo"`      // 编译错误信息
	Score        int         `json:"score"`        // 得分
	Code         string      `json:"code"`         // 提交的代码
}

// GetSubmissionList 获取提交列表
func (c *BingoJClient) GetSubmissionList(onlyMine bool, currentPage, limit int) ([]SubmissionListItem, error) {
	if c.token == "" {
		c.logger.Error("未登录，无法获取提交列表")
		return nil, fmt.Errorf("未登录")
	}

	c.logger.Debug("获取提交列表",
		zap.Bool("only_mine", onlyMine),
		zap.Int("page", currentPage),
		zap.Int("limit", limit))

	url := fmt.Sprintf("%s?onlyMine=%t&currentPage=%d&limit=%d&completeProblemID=false",
		APISubmissionList, onlyMine, currentPage, limit)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Authorization", c.token)
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("获取提交列表请求失败", zap.Error(err))
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var result struct {
		Data struct {
			Records []SubmissionListItem `json:"records"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		c.logger.Error("解析提交列表失败", zap.Error(err), zap.String("response", string(body)))
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	c.logger.Debug("获取提交列表成功", zap.Int("count", len(result.Data.Records)))
	return result.Data.Records, nil
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

// GetProblemDetailAdmin 管理员获取题目详情（可以访问隐藏题目和比赛题目）
func (c *BingoJClient) GetProblemDetailAdmin(pid int64) (*ProblemDetail, error) {
	if c.token == "" {
		c.logger.Error("未登录，无法获取题目详情")
		return nil, fmt.Errorf("未登录")
	}

	c.logger.Info("管理员获取题目详情", zap.Int64("pid", pid))

	url := fmt.Sprintf("%s?pid=%d", APIProblemAdmin, pid)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Authorization", c.token)
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("管理员获取题目详情请求失败", zap.Error(err))
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 管理员API返回格式: { data: Problem }
	var result struct {
		Data *ProblemDetail `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		c.logger.Error("解析管理员题目详情失败", zap.Error(err), zap.String("response", string(body)))
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if result.Data == nil {
		c.logger.Warn("管理员API返回题目为空", zap.String("response", string(body)))
		return nil, fmt.Errorf("题目不存在或无权限")
	}

	c.logger.Info("管理员获取题目成功", zap.String("problem_id", result.Data.ProblemId))
	return result.Data, nil
}

// SearchProblemByDisplayID 通过显示ID搜索题目，返回数据库ID
func (c *BingoJClient) SearchProblemByDisplayID(displayID string) (int64, error) {
	if c.token == "" {
		return 0, fmt.Errorf("未登录")
	}

	c.logger.Info("搜索题目", zap.String("display_id", displayID))

	// 调用管理员题目列表API
	url := fmt.Sprintf("%s?currentPage=1&limit=100&keyword=%s", APIProblemAdminList, displayID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Authorization", c.token)
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("读取响应失败: %w", err)
	}

	var result struct {
		Data struct {
			Records []struct {
				ID        int64  `json:"id"`
				ProblemId string `json:"problemId"`
				Title     string `json:"title"`
				Auth      int    `json:"auth"`
			} `json:"records"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return 0, fmt.Errorf("解析响应失败: %w", err)
	}

	// 在返回的题目列表中查找匹配的显示ID
	for _, problem := range result.Data.Records {
		if problem.ProblemId == displayID {
			c.logger.Info("找到题目", zap.String("display_id", displayID), zap.Int64("db_id", problem.ID))
			return problem.ID, nil
		}
	}

	return 0, fmt.Errorf("未找到匹配的题目")
}
