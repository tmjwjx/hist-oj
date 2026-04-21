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
	db             *gorm.DB
	bingoJClient   *client.BingoJClient
	hojClient      *client.HOJClient
	historyService *SubmissionHistoryService
	logger         *zap.Logger
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
	Username string `json:"username"` // 用户名（可选，用于日志）
	Password string `json:"password"` // 密码（可选，用于手动登录）
	Token    string `json:"token"`    // 已登录用户的token（优先使用）
}

// GetInfoResponse 获取题目信息响应
type GetInfoResponse struct {
	DisplayID string                     `json:"displayId"`
	Problem   *client.ProblemDetail      `json:"problem"`
	History   []*model.SubmissionHistory `json:"history"`
}

// GetInfo 获取题目信息
func (s *JudgeService) GetInfo(req *GetInfoRequest) (*GetInfoResponse, error) {
	// 优先使用 token（已登录用户场景）
	if req.Token != "" {
		s.bingoJClient.SetToken(req.Token)
		s.logger.Info("使用提供的 Token", zap.String("username", req.Username))
	} else if req.Password != "" {
		// 如果没有 token 但提供了密码，则使用密码登录
		if err := s.bingoJClient.Login(req.Username, req.Password); err != nil {
			s.logger.Error("登录失败", zap.Error(err))
			return nil, fmt.Errorf("登录失败: %w", err)
		}
	} else {
		// 既没有 token 也没有密码，返回错误
		s.logger.Error("未提供 Token 或密码")
		return nil, fmt.Errorf("未提供认证信息（Token 或密码）")
	}

	var problem *client.ProblemDetail
	var err error
	var displayID string
	var dbCID string

	// 根据模式获取题目
	if req.Mode == "contest" {
		displayID = strings.ToUpper(req.PID)
		dbCID = req.CID
		// 比赛模式暂不支持管理员API
		problem, err = s.bingoJClient.GetContestProblemDetail(displayID, req.CID)
	} else {
		displayID = req.PID
		dbCID = "0"

		// 统一使用显示ID，先尝试普通API
		s.logger.Info("使用显示ID获取题目", zap.String("display_id", req.PID))
		problem, err = s.bingoJClient.GetProblemDetail(req.PID)

		if err != nil {
			// 普通API失败，可能是隐藏题目，尝试使用管理员API
			s.logger.Warn("普通API获取失败，尝试管理员API", zap.Error(err))

			// 通过显示ID搜索数据库ID
			dbPID, searchErr := s.bingoJClient.SearchProblemByDisplayID(req.PID)
			if searchErr != nil {
				s.logger.Error("搜索题目失败", zap.Error(searchErr))
				return nil, fmt.Errorf("获取题目失败: %w (搜索失败: %v)", err, searchErr)
			}

			// 使用数据库ID调用管理员API
			s.logger.Info("找到数据库ID，使用管理员API获取", zap.Int64("pid", dbPID))
			problem, err = s.bingoJClient.GetProblemDetailAdmin(dbPID)
			if err != nil {
				s.logger.Error("管理员API获取失败", zap.Error(err))
				return nil, fmt.Errorf("获取题目失败: %w", err)
			}

			s.logger.Info("管理员API获取成功", zap.String("problem_id", problem.ProblemId))
		} else {
			s.logger.Info("普通API获取成功", zap.String("problem_id", problem.ProblemId))
		}
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
	ID             int    `json:"id"`
	IsOK           bool   `json:"is_ok"`
	Input          string `json:"input"`
	Expected       string `json:"expected"`
	Output         string `json:"output"`
	Stderr         string `json:"stderr"`                    // 错误输出
	DetailedStderr string `json:"detailed_stderr,omitempty"` // 额外错误详情（如退出码、信号等）
	Status         int    `json:"status"`                    // 判题状态码
}

// ContestTerminalCheckProblemStatus 比赛题目判题终端检测状态
type ContestTerminalCheckProblemStatus struct {
	PID          uint64 `json:"pid"`
	DisplayID    string `json:"displayId"`
	DisplayTitle string `json:"displayTitle"`
	Checked      bool   `json:"checked"`
	Reason       string `json:"reason,omitempty"`
}

// CheckContestTerminalCheckStatus 查询比赛题目是否完成判题终端检测
// 判定标准：至少存在一次 submission_history_case 记录，且覆盖该题全部有效 problem_case。
func (s *JudgeService) CheckContestTerminalCheckStatus(contestID uint64) (map[string]interface{}, error) {
	var contestProblems []model.ContestProblem
	if err := s.db.Where("cid = ?", contestID).Order("display_id ASC").Find(&contestProblems).Error; err != nil {
		return nil, fmt.Errorf("查询比赛题目失败: %w", err)
	}

	if len(contestProblems) == 0 {
		return map[string]interface{}{
			"contestId":         contestID,
			"totalProblems":     0,
			"checkedProblems":   0,
			"allChecked":        false,
			"uncheckedProblems": []ContestTerminalCheckProblemStatus{},
		}, nil
	}

	pids := make([]uint64, 0, len(contestProblems))
	for _, item := range contestProblems {
		pids = append(pids, item.PID)
	}

	type caseTotalRow struct {
		PID   uint64 `gorm:"column:pid"`
		Total int64  `gorm:"column:total"`
	}
	var caseTotals []caseTotalRow
	if err := s.db.Table("problem_case").
		Select("pid, COUNT(*) as total").
		Where("status = 0 AND pid IN ?", pids).
		Group("pid").
		Scan(&caseTotals).Error; err != nil {
		return nil, fmt.Errorf("查询测试点统计失败: %w", err)
	}

	totalCaseMap := make(map[uint64]int64, len(caseTotals))
	for _, row := range caseTotals {
		totalCaseMap[row.PID] = row.Total
	}

	checkedProblems := 0
	unchecked := make([]ContestTerminalCheckProblemStatus, 0)

	for _, cp := range contestProblems {
		totalCases := totalCaseMap[cp.PID]
		if totalCases <= 0 {
			unchecked = append(unchecked, ContestTerminalCheckProblemStatus{
				PID:          cp.PID,
				DisplayID:    cp.DisplayID,
				DisplayTitle: cp.DisplayTitle,
				Checked:      false,
				Reason:       "题目缺少有效测试点(problem_case)",
			})
			continue
		}

		var matched []string
		if err := s.db.Table("submission_history_case shc").
			Select("shc.submit_id").
			Joins("JOIN problem_case pc ON pc.id = shc.case_id").
			Where("pc.pid = ? AND pc.status = 0", cp.PID).
			Group("shc.submit_id").
			Having("COUNT(DISTINCT shc.case_id) >= ?", totalCases).
			Order("MAX(shc.create_time) DESC").
			Limit(1).
			Pluck("shc.submit_id", &matched).Error; err != nil {
			return nil, fmt.Errorf("查询题目检测记录失败(pid=%d): %w", cp.PID, err)
		}

		if len(matched) == 0 {
			unchecked = append(unchecked, ContestTerminalCheckProblemStatus{
				PID:          cp.PID,
				DisplayID:    cp.DisplayID,
				DisplayTitle: cp.DisplayTitle,
				Checked:      false,
				Reason:       fmt.Sprintf("未发现覆盖全部 %d 个测试点的判题终端检测记录", totalCases),
			})
			continue
		}

		checkedProblems++
	}

	result := map[string]interface{}{
		"contestId":         contestID,
		"totalProblems":     len(contestProblems),
		"checkedProblems":   checkedProblems,
		"allChecked":        checkedProblems == len(contestProblems),
		"uncheckedProblems": unchecked,
	}
	return result, nil
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
func (s *JudgeService) TestLocalSamples(pid int64, language, code, username, password, token string, samples []SampleResult) ([]SampleResult, error) {
	// 先登录 HOJ（共享的客户端会保持登录状态）
	s.logger.Info("开始登录 HOJ 后端")

	// 优先使用 Token，如果没有 Token 则使用密码登录
	if token != "" {
		s.logger.Info("使用 Token 认证 HOJ")
		s.hojClient.SetToken(token)
		s.logger.Info("HOJ Token 认证成功")
	} else if password != "" {
		if err := s.hojClient.Login(username, password); err != nil {
			s.logger.Error("登录 HOJ 失败", zap.Error(err))
			return nil, fmt.Errorf("登录 HOJ 失败: %w", err)
		}
		s.logger.Info("HOJ 密码登录成功")
	} else {
		return nil, fmt.Errorf("未提供认证信息（Token 或密码）")
	}

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
		result.Output = testResult.Output      // HOJ 已经提供了清理后的 userOutput
		result.IsOK = (testResult.Status == 0) // 0 表示 Accepted
		result.Stderr = testResult.Stderr      // 保存错误信息
		result.Status = testResult.Status      // 保存判题状态码

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

// JudgeCaseDetail 判题测试点详情
type JudgeCaseDetail struct {
	CaseID   int64  `json:"case_id" gorm:"column:case_id"`
	Status   *int   `json:"status" gorm:"column:status"`
	Time     *int   `json:"time" gorm:"column:time"`
	Memory   *int   `json:"memory" gorm:"column:memory"`
	Score    *int   `json:"score" gorm:"column:score"`
	GroupNum *int   `json:"group_num" gorm:"column:group_num"`
	Seq      *int   `json:"seq" gorm:"column:seq"`
	Mode     string `json:"mode" gorm:"column:mode"`
	Stderr   string `json:"stderr,omitempty" gorm:"column:stderr"`
}

// GetJudgeCaseDetails 获取判题测试点详情
func (s *JudgeService) GetJudgeCaseDetails(submitID string) ([]*JudgeCaseDetail, error) {
	var details []*JudgeCaseDetail

	err := s.db.Table("judge_case").
		Select("case_id, status, time, memory, score, group_num, seq, mode").
		Where("submit_id = ?", submitID).
		Order("CASE WHEN seq IS NULL THEN 1 ELSE 0 END").
		Order("seq ASC").
		Order("case_id ASC").
		Scan(&details).Error
	if err != nil {
		s.logger.Error("查询判题测试点详情失败",
			zap.String("submit_id", submitID),
			zap.Error(err))
		return nil, fmt.Errorf("查询判题测试点详情失败: %w", err)
	}

	if len(details) > 0 {
		return details, nil
	}

	// 判题终端本地独立判题详情（submission_history_case）
	var localDetails []*JudgeCaseDetail
	err = s.db.Table("submission_history_case").
		Select("case_id, status, time, memory, score, group_num, seq, mode, stderr").
		Where("submit_id = ?", submitID).
		Order("CASE WHEN seq IS NULL THEN 1 ELSE 0 END").
		Order("seq ASC").
		Order("case_id ASC").
		Scan(&localDetails).Error
	if err != nil {
		s.logger.Error("查询本地判题测试点详情失败",
			zap.String("submit_id", submitID),
			zap.Error(err))
		return nil, fmt.Errorf("查询本地判题测试点详情失败: %w", err)
	}

	if len(localDetails) > 0 {
		return localDetails, nil
	}

	return details, nil
}

// GetBingoJClient 获取 BingoJ 客户端实例
func (s *JudgeService) GetBingoJClient() *client.BingoJClient {
	return s.bingoJClient
}

// GetHOJClient 获取 HOJ 客户端实例
func (s *JudgeService) GetHOJClient() *client.HOJClient {
	return s.hojClient
}
