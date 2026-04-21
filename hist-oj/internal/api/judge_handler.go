package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/hoj/hist-oj/internal/client"
	"github.com/hoj/hist-oj/internal/model"
	"github.com/hoj/hist-oj/internal/service"
	"github.com/hoj/hist-oj/internal/utils"
)

// JudgeHandler 判题终端 Handler
type JudgeHandler struct {
	judgeService *service.JudgeService
	logger       *zap.Logger
}

// NewJudgeHandler 创建判题终端 Handler
func NewJudgeHandler(judgeService *service.JudgeService) *JudgeHandler {
	return &JudgeHandler{
		judgeService: judgeService,
		logger:       utils.GetLogger(),
	}
}

// GetInfo 获取题目信息
func (h *JudgeHandler) GetInfo(c *gin.Context) {
	var req service.GetInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	// 设置默认值
	if req.CID == "" {
		req.CID = "0"
	}

	h.logger.Info("获取题目信息",
		zap.String("pid", req.PID),
		zap.String("cid", req.CID),
		zap.String("mode", req.Mode))

	result, err := h.judgeService.GetInfo(&req)
	if err != nil {
		h.logger.Error("获取题目信息失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, err.Error()))
		return
	}

	c.JSON(http.StatusOK, successResponse(result))
}

// SubmitRequest 提交代码请求
type SubmitRequest struct {
	PID      string `json:"pid" binding:"required"`
	CID      string `json:"cid"`
	Mode     string `json:"mode" binding:"required"`
	Username string `json:"username"`
	Password string `json:"password"`
	Token    string `json:"token"`
	Language string `json:"language" binding:"required"`
	Code     string `json:"code" binding:"required"`
}

// SubmitResponse 提交代码响应
type SubmitResponse struct {
	SubmitID string `json:"submit_id"`
}

// Submit 提交代码
func (h *JudgeHandler) Submit(c *gin.Context) {
	var req SubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	// 设置默认值
	if req.CID == "" {
		req.CID = "0"
	}

	h.logger.Info("提交代码",
		zap.String("pid", req.PID),
		zap.String("cid", req.CID),
		zap.String("language", req.Language))

	// 优先使用 token，其次使用密码登录
	bingoJClient := h.judgeService.GetBingoJClient()

	if req.Token != "" {
		bingoJClient.SetToken(req.Token)
		h.logger.Info("使用提供的 Token", zap.String("username", req.Username))
	} else if req.Password != "" {
		if err := bingoJClient.Login(req.Username, req.Password); err != nil {
			h.logger.Error("登录失败", zap.Error(err))
			c.JSON(http.StatusOK, errorResponse(500, "登录失败"))
			return
		}
	} else {
		h.logger.Error("未提供 Token 或密码")
		c.JSON(http.StatusOK, errorResponse(400, "未提供认证信息（Token 或密码）"))
		return
	}

	// 提交代码
	submitID, err := bingoJClient.SubmitCode(req.PID, 0, req.Language, req.Code)
	if err != nil {
		h.logger.Error("提交失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, err.Error()))
		return
	}

	h.logger.Info("提交成功", zap.String("submit_id", submitID))
	c.JSON(http.StatusOK, successResponse(&SubmitResponse{SubmitID: submitID}))
}

// GetHistoryRequest 获取历史记录请求
type GetHistoryRequest struct {
	PID      string `json:"pid" binding:"required"`
	CID      string `json:"cid"`
	Page     int    `json:"page" binding:"required"`
	PageSize int    `json:"pageSize" binding:"required"`
}

// GetHistoryResponse 获取历史记录响应
type GetHistoryResponse struct {
	List  []*model.SubmissionHistory `json:"list"`
	Total int64                      `json:"total"`
}

// GetCaseDetailsRequest 获取测试点详情请求
type GetCaseDetailsRequest struct {
	SubmitID string `json:"submit_id" binding:"required"`
	Username string `json:"username"`
	Password string `json:"password"` // 密码（可选，与 Token 二选一）
	Token    string `json:"token"`    // Token（可选，与 Password 二选一）
}

// GetCaseDetailsResponse 获取测试点详情响应
type GetCaseDetailsResponse struct {
	SubmitID string                     `json:"submit_id"`
	Cases    []*service.JudgeCaseDetail `json:"cases"`
}

// GetHistory 获取提交历史（分页）
func (h *JudgeHandler) GetHistory(c *gin.Context) {
	var req GetHistoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	// 设置默认值
	if req.CID == "" {
		req.CID = "0"
	}
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 10
	}

	h.logger.Info("获取历史记录",
		zap.String("pid", req.PID),
		zap.String("cid", req.CID),
		zap.Int("page", req.Page),
		zap.Int("pageSize", req.PageSize))

	list, total, err := h.judgeService.GetHistory(req.PID, req.CID, req.Page, req.PageSize)
	if err != nil {
		h.logger.Error("获取历史记录失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, err.Error()))
		return
	}

	result := &GetHistoryResponse{
		List:  list,
		Total: total,
	}

	c.JSON(http.StatusOK, successResponse(result))
}

// GetCaseDetails 获取提交测试点详情
func (h *JudgeHandler) GetCaseDetails(c *gin.Context) {
	var req GetCaseDetailsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	h.logger.Info("获取测试点详情", zap.String("submit_id", req.SubmitID))

	// 认证：优先使用 Token，其次使用用户名+密码
	bingoJClient := h.judgeService.GetBingoJClient()
	if req.Token != "" {
		bingoJClient.SetToken(req.Token)
	} else if req.Password != "" {
		if err := bingoJClient.Login(req.Username, req.Password); err != nil {
			h.logger.Error("获取测试点详情登录失败", zap.Error(err))
			c.JSON(http.StatusOK, errorResponse(401, "用户认证失败"))
			return
		}
	} else {
		c.JSON(http.StatusOK, errorResponse(400, "未提供认证信息（Token 或密码）"))
		return
	}

	cases, err := h.judgeService.GetJudgeCaseDetails(req.SubmitID)
	if err != nil {
		h.logger.Error("获取测试点详情失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, err.Error()))
		return
	}

	result := &GetCaseDetailsResponse{
		SubmitID: req.SubmitID,
		Cases:    cases,
	}

	c.JSON(http.StatusOK, successResponse(result))
}

// GetContestTerminalCheckStatus 获取比赛题目判题终端检测状态（管理员）
func (h *JudgeHandler) GetContestTerminalCheckStatus(c *gin.Context) {
	contestIDStr := c.Param("contestId")
	contestID, err := strconv.ParseUint(contestIDStr, 10, 64)
	if err != nil || contestID == 0 {
		c.JSON(http.StatusOK, errorResponse(400, "contestId参数格式错误"))
		return
	}

	result, err := h.judgeService.CheckContestTerminalCheckStatus(contestID)
	if err != nil {
		h.logger.Error("查询比赛题目判题终端检测状态失败",
			zap.Uint64("contest_id", contestID),
			zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, successResponse(result))
}

// RunCombinedRequest 本地测试+远程提交请求
type RunCombinedRequest struct {
	PID      string `json:"pid" binding:"required"`
	CID      string `json:"cid"`
	Mode     string `json:"mode" binding:"required"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password"` // 密码（可选，与 Token 二选一）
	Token    string `json:"token"`    // Token（可选，与 Password 二选一）
	Language string `json:"language" binding:"required"`
	Code     string `json:"code" binding:"required"`
}

// SSEMessage SSE 消息
type SSEMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data,omitempty"`
	Msg  string      `json:"msg,omitempty"`
}

// sendSSE 发送 SSE 消息
func sendSSE(c *gin.Context, msgType string, data interface{}) {
	msg := SSEMessage{Type: msgType}

	switch msgType {
	case "log":
		msg.Msg = data.(string)
	default:
		msg.Data = data
	}

	jsonData, _ := json.Marshal(msg)
	fmt.Fprintf(c.Writer, "data: %s\n\n", jsonData)
	c.Writer.Flush()
}

// RunCombined 本地测试+远程提交（SSE 流式返回）
func (h *JudgeHandler) RunCombined(c *gin.Context) {
	var req RunCombinedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	// 设置 SSE 响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	// 设置默认值
	if req.CID == "" {
		req.CID = "0"
	}

	h.logger.Info("开始判题流程",
		zap.String("pid", req.PID),
		zap.String("language", req.Language))

	// 1. 登录（使用共享的客户端实例）
	sendSSE(c, "log", "正在登录 OJ...")
	bingoJClient := h.judgeService.GetBingoJClient()

	// 优先使用 Token，如果没有 Token 则使用密码登录
	if req.Token != "" {
		h.logger.Info("使用 Token 认证")
		bingoJClient.SetToken(req.Token)
		sendSSE(c, "log", "使用 Token 认证成功")
	} else if req.Password != "" {
		h.logger.Info("使用密码登录", zap.String("username", req.Username))
		if err := bingoJClient.Login(req.Username, req.Password); err != nil {
			sendSSE(c, "log", fmt.Sprintf("登录失败: %s", err.Error()))
			return
		}
		sendSSE(c, "log", "密码登录成功")
	} else {
		sendSSE(c, "log", "未提供认证信息（Token 或密码）")
		return
	}

	// 2. 获取题目样例
	sendSSE(c, "log", "正在获取样例...")
	var problem *client.ProblemDetail
	var err error
	var displayID string
	var dbCID string

	// 只支持普通模式
	displayID = req.PID
	dbCID = "0"

	// 统一使用显示ID，先尝试普通API
	sendSSE(c, "log", fmt.Sprintf("尝试获取题目: %s", req.PID))
	problem, err = bingoJClient.GetProblemDetail(req.PID)

	if err != nil {
		// 普通API失败，可能是隐藏题目，尝试使用管理员API
		sendSSE(c, "log", "普通API获取失败，尝试管理员API...")

		// 通过显示ID搜索数据库ID
		var dbPID int64
		dbPID, err = bingoJClient.SearchProblemByDisplayID(req.PID)
		if err != nil {
			sendSSE(c, "log", fmt.Sprintf("搜索题目失败: %s", err.Error()))
			sendSSE(c, "log", "获取题目失败")
			return
		}

		// 使用数据库ID调用管理员API
		sendSSE(c, "log", fmt.Sprintf("使用管理员API获取题目 (dbID: %d)", dbPID))
		problem, err = bingoJClient.GetProblemDetailAdmin(dbPID)
		if err != nil {
			sendSSE(c, "log", fmt.Sprintf("管理员API获取失败: %s", err.Error()))
			sendSSE(c, "log", "获取题目失败")
			return
		}

		sendSSE(c, "log", "管理员API获取成功")
	} else {
		sendSSE(c, "log", "普通API获取成功")
	}

	// 调试：输出题目ID
	h.logger.Info("题目详情获取成功",
		zap.String("displayID", displayID),
		zap.String("bingojProblemId", problem.ProblemId),
		zap.Int64("dbID", problem.ID),
		zap.String("title", problem.Title))

	localInfo := "跳过 (无样例)"
	var samples []service.SampleResult

	// 3. 本地样例测试
	if problem.Examples != "" {
		samples, err = service.ExtractSamples(problem.Examples)
		if err == nil && len(samples) > 0 {
			sendSSE(c, "log", fmt.Sprintf("开始自测 %d 组样例...", len(samples)))
			results, err := h.judgeService.TestLocalSamplesNative(
				req.Language,
				req.Code,
				samples,
				problem.TimeLimit,
				problem.MemoryLimit,
				problem.ID,
				problem.JudgeMode,
			)
			if err != nil {
				sendSSE(c, "compile_error", map[string]string{"msg": err.Error()})
				sendSSE(c, "log", fmt.Sprintf("样例自测失败: %s", err.Error()))
				return
			}

			// 发送每个样例的结果
			passCount := 0
			for _, result := range results {
				if result.IsOK {
					passCount++
				}
				sendSSE(c, "sample_res", result)
			}

			if passCount == len(results) {
				localInfo = "全部通过"
			} else {
				localInfo = fmt.Sprintf("%d/%d 通过", passCount, len(results))
			}
			sendSSE(c, "log", fmt.Sprintf("自测完成: %s", localInfo))
		} else {
			sendSSE(c, "log", "无有效样例")
		}
	}

	// 4. 判题终端独立判题（Go + 沙箱，不依赖 Java 远程判题结果）
	sendSSE(c, "log", "启动判题终端独立判题引擎（Go Sandbox）...")
	sendSSE(c, "remote_status", map[string]string{"status": "编译中"})

	submitID := service.GenerateLocalSubmitID()
	sendSSE(c, "remote_submit", map[string]string{"submit_id": submitID})
	sendSSE(c, "remote_status", map[string]string{"status": "判题中", "submit_id": submitID})

	casesForJudge, caseSource, caseErr := h.judgeService.BuildTerminalJudgeCases(problem.ID, samples)
	if caseErr != nil {
		h.logger.Warn("构建判题测试点时发生回退", zap.Error(caseErr), zap.Int64("pid", problem.ID))
		sendSSE(c, "log", fmt.Sprintf("测试点准备提示: %s", caseErr.Error()))
	}
	sendSSE(c, "log", fmt.Sprintf("测试点来源: %s", caseSource))

	if len(casesForJudge) == 0 {
		msg := "未找到可用测试点，无法独立判题"
		sendSSE(c, "log", msg)
		sendSSE(c, "remote_status", map[string]string{
			"status":       "系统错误",
			"submit_id":    submitID,
			"errorMessage": msg,
		})
		return
	}

	judgeResult, err := h.judgeService.RunTerminalJudge(
		req.Language,
		req.Code,
		casesForJudge,
		problem.TimeLimit,
		problem.MemoryLimit,
		problem.JudgeMode,
		problem.JudgeCaseMode,
		func(done, total, seq, status int) {
			sendSSE(c, "case_progress", map[string]interface{}{
				"done":   done,
				"total":  total,
				"seq":    seq,
				"status": status,
			})
		},
	)
	if err != nil {
		if service.IsNativeJudgeUnavailable(err) {
			sendSSE(c, "log", fmt.Sprintf("独立判题环境缺失: %s", err.Error()))
			sendSSE(c, "remote_status", map[string]string{
				"status":       "系统错误",
				"submit_id":    submitID,
				"errorMessage": err.Error(),
			})
			return
		}
		sendSSE(c, "compile_error", map[string]string{"msg": err.Error()})
		sendSSE(c, "remote_status", map[string]string{
			"status":       "编译错误",
			"submit_id":    submitID,
			"errorMessage": err.Error(),
		})
		sendSSE(c, "log", fmt.Sprintf("独立判题失败: %s", err.Error()))
		return
	}

	if err := h.judgeService.SaveLocalJudgeCaseDetails(submitID, judgeResult.CaseRecords); err != nil {
		h.logger.Warn("保存本地判题测试点详情失败", zap.String("submit_id", submitID), zap.Error(err))
		sendSSE(c, "log", fmt.Sprintf("警告: 保存测试点详情失败: %s", err.Error()))
	}

	if len(judgeResult.CaseDetails) > 0 {
		sendSSE(c, "case_details", map[string]interface{}{
			"submit_id": submitID,
			"cases":     judgeResult.CaseDetails,
		})
		sendSSE(c, "log", fmt.Sprintf("测试点详情已返回，共 %d 个测试点", len(judgeResult.CaseDetails)))
	}

	statusText := client.GetStatusText(judgeResult.FinalStatus)
	statusPayload := map[string]interface{}{
		"status":    statusText,
		"submit_id": submitID,
	}
	if strings.TrimSpace(judgeResult.ErrorMessage) != "" {
		statusPayload["errorMessage"] = judgeResult.ErrorMessage
	}
	if judgeResult.FirstFailedSeq > 0 {
		statusPayload["failed_seq"] = judgeResult.FirstFailedSeq
	}
	if judgeResult.FirstFailedStatus != 0 {
		statusPayload["failed_status"] = judgeResult.FirstFailedStatus
	}
	sendSSE(c, "remote_status", statusPayload)
	sendSSE(c, "log", fmt.Sprintf("独立判题完成: %s", statusText))

	if err := h.judgeService.SaveSubmissionHistory(
		submitID,
		displayID,
		dbCID,
		req.Username,
		statusText,
		fmt.Sprintf("%dms", judgeResult.MaxTime),
		fmt.Sprintf("%dKB", judgeResult.MaxMemory),
		req.Language,
		req.Code,
		localInfo,
	); err != nil {
		h.logger.Error("保存提交历史失败", zap.Error(err))
		sendSSE(c, "log", fmt.Sprintf("警告: 保存历史记录失败: %s", err.Error()))
	} else {
		sendSSE(c, "log", "提交历史已保存")
	}
}
