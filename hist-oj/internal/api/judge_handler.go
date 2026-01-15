package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

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
	Total int64                       `json:"total"`
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
	var submitCID int

	// 只支持普通模式
	displayID = req.PID
	dbCID = "0"
	submitCID = 0
	problem, err = bingoJClient.GetProblemDetail(req.PID)

	if err != nil {
		sendSSE(c, "log", fmt.Sprintf("获取题目失败: %s", err.Error()))
		return
	}

	// 调试：输出题目ID
	h.logger.Info("题目详情获取成功",
		zap.String("displayID", displayID),
		zap.String("bingojProblemId", problem.ProblemId),
		zap.Int64("dbID", problem.ID),
		zap.String("title", problem.Title))

	// 样例测试需要使用 HOJ 数据库的主键ID（problem.ID）
	// 这个ID会被发送到HOJ后端的样例测试接口
	pidForTest := fmt.Sprintf("%d", problem.ID)

	h.logger.Info("样例测试将使用数据库ID", zap.String("dbID", pidForTest))

	localInfo := "跳过 (无样例)"

	// 3. 本地样例测试
	if problem.Examples != "" {
		samples, err := service.ExtractSamples(problem.Examples)
		if err == nil && len(samples) > 0 {
			sendSSE(c, "log", fmt.Sprintf("开始自测 %d 组样例...", len(samples)))

			// 使用 pidForTest（用户输入的 HOJ 数据库题目 ID）进行样例测试
			pidInt, err := strconv.ParseInt(pidForTest, 10, 64)
			if err != nil {
				h.logger.Error("解析题目ID失败", zap.Error(err), zap.String("pidForTest", pidForTest))
				sendSSE(c, "log", fmt.Sprintf("题目ID格式错误: %s", pidForTest))
				return
			}

			h.logger.Info("开始本地样例测试",
				zap.Int64("hojPid", pidInt),
				zap.String("显示ID", req.PID),
				zap.Int("样例数量", len(samples)))
			results, err := h.judgeService.TestLocalSamples(pidInt, req.Language, req.Code, req.Username, req.Password, req.Token, samples)
			if err != nil {
				sendSSE(c, "compile_error", map[string]string{"msg": err.Error()})
				sendSSE(c, "log", "本地编译失败")
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

	// 4. 提交远程 OJ
	sendSSE(c, "log", "正在提交远程 OJ...")
	sendSSE(c, "remote_status", map[string]string{"status": "提交中"})

	// 提交代码
	submitID, err := bingoJClient.SubmitCode(displayID, submitCID, req.Language, req.Code)
	if err != nil {
		sendSSE(c, "log", fmt.Sprintf("提交失败: %s", err.Error()))
		sendSSE(c, "remote_status", map[string]string{"status": "提交失败"})
		return
	}

	h.logger.Info("远程提交成功", zap.String("submit_id", submitID))
	sendSSE(c, "log", fmt.Sprintf("提交ID: %s，等待 2 秒后查询结果...", submitID))

	// 等待 2 秒后再查询结果，避免提交过快导致查询失败
	time.Sleep(2 * time.Second)

	// 轮询判题结果
	for i := 0; i < 20; i++ {
		h.logger.Debug("查询判题结果", zap.Int("次数", i+1), zap.String("submit_id", submitID))
		result, err := bingoJClient.GetSubmissionResult(submitID)
		if err != nil {
			h.logger.Warn("查询判题结果失败，继续重试", zap.Error(err), zap.Int("次数", i+1))
			time.Sleep(1 * time.Second)
			continue
		}

		statusText := client.GetStatusText(result.Status)
		h.logger.Info("判题结果", zap.Int("status", result.Status), zap.String("status_text", statusText))

		// 判题中
		if result.Status == 6 || result.Status == 7 || result.Status == 9 {
			sendSSE(c, "remote_status", map[string]string{"status": statusText})
			time.Sleep(1 * time.Second)
			continue
		}

		// 判题完成
		sendSSE(c, "remote_status", map[string]string{"status": statusText})
		sendSSE(c, "log", fmt.Sprintf("最终结果: %s", statusText))

		// 保存到数据库
		if err := h.judgeService.SaveSubmissionHistory(
			submitID,
			displayID,
			dbCID,
			req.Username,
			statusText,
			fmt.Sprintf("%dms", result.Time),
			fmt.Sprintf("%dKB", result.Memory),
			req.Language,
			req.Code,
			localInfo,
		); err != nil {
			h.logger.Error("保存提交历史失败", zap.Error(err))
			sendSSE(c, "log", fmt.Sprintf("警告: 保存历史记录失败: %s", err.Error()))
		} else {
			sendSSE(c, "log", "提交历史已保存")
		}

		return
	}

	sendSSE(c, "log", "查询超时")
	sendSSE(c, "remote_status", map[string]string{"status": "超时"})
}
