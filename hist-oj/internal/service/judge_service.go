package service

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/hoj/hist-oj/internal/client"
	"github.com/hoj/hist-oj/internal/model"
	"github.com/hoj/hist-oj/internal/utils"
)

// JudgeService 判题服务
type JudgeService struct {
	db                *gorm.DB
	bingoJClient      *client.BingoJClient
	hojClient         *client.HOJClient
	historyService    *SubmissionHistoryService
	logger            *zap.Logger
}

// NewJudgeService 创建判题服务
func NewJudgeService(db *gorm.DB, hojBaseURL string) *JudgeService {
	// HOJ 后端地址，从参数传入
	if hojBaseURL == "" {
		hojBaseURL = "http://43.143.133.62:6688" // 默认值
	}

	return &JudgeService{
		db:             db,
		bingoJClient:   client.NewBingoJClient(),
		hojClient:      client.NewHOJClient(hojBaseURL, "", ""),
		historyService: NewSubmissionHistoryService(db),
		logger:         utils.GetLogger(),
	}
}

// GetInfoRequest 获取题目信息请求
type GetInfoRequest struct {
	PID      string `json:"pid" binding:"required"`
	CID      string `json:"cid"`
	Mode     string `json:"mode" binding:"required"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// GetInfoResponse 获取题目信息响应
type GetInfoResponse struct {
	DisplayID string                      `json:"displayId"`
	Problem   *client.ProblemDetail       `json:"problem"`
	History   []*model.SubmissionHistory  `json:"history"`
}

// GetInfo 获取题目信息
func (s *JudgeService) GetInfo(req *GetInfoRequest) (*GetInfoResponse, error) {
	// 登录
	if err := s.bingoJClient.Login(req.Username, req.Password); err != nil {
		s.logger.Error("登录失败", zap.Error(err))
		return nil, fmt.Errorf("登录失败: %w", err)
	}

	var problem *client.ProblemDetail
	var err error
	var displayID string
	var dbCID string

	// 根据模式获取题目
	if req.Mode == "contest" {
		displayID = strings.ToUpper(req.PID)
		dbCID = req.CID
		problem, err = s.bingoJClient.GetContestProblemDetail(displayID, req.CID)
	} else {
		displayID = req.PID
		dbCID = "0"
		problem, err = s.bingoJClient.GetProblemDetail(req.PID)
	}

	if err != nil {
		s.logger.Error("获取题目失败", zap.Error(err))
		return nil, fmt.Errorf("获取题目失败: %w", err)
	}

	// 查询历史记录
	history, err := s.historyService.GetByPIDAndCID(displayID, dbCID, 10)
	if err != nil {
		s.logger.Warn("查询历史记录失败", zap.Error(err))
		history = []*model.SubmissionHistory{}
	}

	return &GetInfoResponse{
		DisplayID: displayID,
		Problem:   problem,
		History:   history,
	}, nil
}

// SampleResult 样例测试结果
type SampleResult struct {
	ID       int    `json:"id"`
	IsOK     bool   `json:"is_ok"`
	Input    string `json:"input"`
	Expected string `json:"expected"`
	Output   string `json:"output"`
}

// ExtractSamples 从题目样例中提取输入输出
func ExtractSamples(examples string) ([]SampleResult, error) {
	re := regexp.MustCompile(`<input>([\s\S]*?)</input><output>([\s\S]*?)</output>`)
	matches := re.FindAllStringSubmatch(examples, -1)

	if len(matches) == 0 {
		return nil, fmt.Errorf("未找到有效样例")
	}

	samples := make([]SampleResult, 0, len(matches))
	for i, match := range matches {
		samples = append(samples, SampleResult{
			ID:       i + 1,
			Input:    strings.TrimSpace(match[1]),
			Expected: strings.TrimSpace(match[2]),
		})
	}

	return samples, nil
}

// TestLocalSamples 本地测试样例（使用 HOJ 后端）
func (s *JudgeService) TestLocalSamples(pid int64, language, code, username, password string, samples []SampleResult) ([]SampleResult, error) {
	// 先登录 HOJ（共享的客户端会保持登录状态）
	s.logger.Info("开始登录 HOJ 后端")
	if err := s.hojClient.Login(username, password); err != nil {
		s.logger.Error("登录 HOJ 失败", zap.Error(err))
		return nil, fmt.Errorf("登录 HOJ 失败: %w", err)
	}
	s.logger.Info("HOJ 登录成功")

	results := make([]SampleResult, 0, len(samples))

	// 对每个样例调用 HOJ 测试接口
	for i, sample := range samples {
		s.logger.Info("开始测试样例", zap.Int("样例编号", i+1), zap.Int("总数", len(samples)))
		result := sample

		// 提交测试请求
		testReq := &client.TestJudgeReq{
			Pid:            pid,
			Type:           "public",
			Code:           code,
			Language:       language,
			UserInput:      sample.Input,
			ExpectedOutput: sample.Expected,
			IsRemoteJudge:  false,
		}

		s.logger.Info("提交测试请求", zap.String("输入", truncateString(sample.Input, 50)))
		testJudgeKey, err := s.hojClient.SubmitTestJudge(testReq)
		if err != nil {
			s.logger.Error("提交测试失败", zap.Error(err))
			result.IsOK = false
			result.Output = fmt.Sprintf("提交测试失败: %v", err)
			results = append(results, result)
			continue
		}
		s.logger.Info("提交测试成功", zap.String("testJudgeKey", testJudgeKey))

		// 轮询查询结果（最多等待 30 秒）
		var testResult *client.TestJudgeRes
		for j := 0; j < 30; j++ {
			time.Sleep(1 * time.Second)

			testResult, err = s.hojClient.GetTestJudgeResult(testJudgeKey)
			if err != nil {
				// 如果是结果不存在的错误，继续等待
				if strings.Contains(err.Error(), "不存在") || strings.Contains(err.Error(), "未找到") {
					s.logger.Debug("测试结果未就绪，继续等待", zap.Int("次数", j+1))
					continue
				}

				// 其他错误直接返回
				s.logger.Error("查询测试结果失败", zap.Error(err))
				result.IsOK = false
				result.Output = fmt.Sprintf("查询结果失败: %v", err)
				results = append(results, result)
				break
			}

			// 没有错误，检查判题状态
			// status: 0=AC, -1=WA, -2=CE, -3=PE, 1=TLE, 2=MLE, 3=RE, 5=判题中
			if testResult.Status == 5 || testResult.Status == 6 {
				// 还在判题中，继续等待
				s.logger.Debug("判题中，继续等待", zap.Int("次数", j+1), zap.Int("status", testResult.Status))
				continue
			}

			// 判题完成（无论成功还是失败）
			s.logger.Info("查询测试结果成功", zap.Int("等待次数", j+1), zap.Int("最终状态", testResult.Status))
			break
		}

		if testResult == nil {
			result.IsOK = false
			result.Output = "测试超时"
			results = append(results, result)
			continue
		}

		// 处理测试结果
		result.Output = testResult.Output // HOJ 已经提供了清理后的 userOutput
		result.IsOK = (testResult.Status == 0) // 0 表示 Accepted

		// 调试：打印完整结果
		s.logger.Info("收到测试结果",
			zap.Int("样例编号", i+1),
			zap.Int("status", testResult.Status),
			zap.String("output", testResult.Output),
			zap.String("stderr", testResult.Stderr),
			zap.String("expected", testResult.ExpectedOutput))

		// 如果程序输出为空但有错误信息，则显示错误信息
		if result.Output == "" && testResult.Stderr != "" {
			result.Output = testResult.Stderr
		}

		// 如果程序输出和错误信息都为空，显示提示
		if result.Output == "" {
			result.Output = "(无输出)"
		}

		s.logger.Info("样例测试完成",
			zap.Int("样例编号", i+1),
			zap.Bool("是否通过", result.IsOK),
			zap.Int("状态码", testResult.Status))

		results = append(results, result)

		// 如果不是最后一个样例，等待 2 秒再提交下一个
		if i < len(samples)-1 {
			s.logger.Info("等待 2 秒后提交下一个样例...")
			time.Sleep(2 * time.Second)
		}
	}

	return results, nil
}

// truncateString 截断字符串用于日志输出
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// SubmitAndWaitResult 提交代码并等待结果
func (s *JudgeService) SubmitAndWaitResult(pid string, cid int, language, code string) (*client.SubmissionResult, error) {
	// 提交代码
	submitID, err := s.bingoJClient.SubmitCode(pid, cid, language, code)
	if err != nil {
		return nil, fmt.Errorf("提交失败: %w", err)
	}

	s.logger.Info("代码已提交，等待判题结果", zap.String("submit_id", submitID))

	// 轮询判题结果（最多20次，每次间隔1秒）
	for i := 0; i < 20; i++ {
		time.Sleep(1 * time.Second)

		result, err := s.bingoJClient.GetSubmissionResult(submitID)
		if err != nil {
			s.logger.Warn("查询判题结果失败", zap.Error(err))
			continue
		}

		// 检查是否还在判题中
		if result.Status == 6 || result.Status == 7 || result.Status == 9 {
			continue
		}

		// 判题完成
		return result, nil
	}

	return nil, fmt.Errorf("查询判题结果超时")
}

// SaveSubmissionHistory 保存提交历史
func (s *JudgeService) SaveSubmissionHistory(
	submitID, pid, cid, username, result, timeUsed, memoryUsed, language, code, localInfo string,
) error {
	history := &model.SubmissionHistory{
		SubmitID:   submitID,
		PID:        pid,
		CID:        cid,
		Username:   username,
		Result:     result,
		TimeUsed:   timeUsed,
		MemoryUsed: memoryUsed,
		Language:   language,
		Code:       code,
		LocalInfo:  localInfo,
		SubmitTime: time.Now(),
	}

	return s.historyService.Create(history)
}

// GetHistory 获取提交历史（分页）
func (s *JudgeService) GetHistory(pid, cid string, page, pageSize int) ([]*model.SubmissionHistory, int64, error) {
	return s.historyService.GetByPIDAndCIDWithPage(pid, cid, page, pageSize)
}

// GetBingoJClient 获取 BingoJ 客户端实例
func (s *JudgeService) GetBingoJClient() *client.BingoJClient {
	return s.bingoJClient
}

// GetHOJClient 获取 HOJ 客户端实例
func (s *JudgeService) GetHOJClient() *client.HOJClient {
	return s.hojClient
}
