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

// RunCombinedRequest 本地测试+远程提交请求
type RunCombinedRequest struct {
	PID      string `json:"pid" binding:"required"`
	CID      string `json:"cid"`
	Mode     string `json:"mode" binding:"required"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
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

	// 1. 登录
	sendSSE(c, "log", "正在登录 OJ...")
	bingoJClient := client.NewBingoJClient()
	if err := bingoJClient.Login(req.Username, req.Password); err != nil {
		sendSSE(c, "log", fmt.Sprintf("登录失败: %s", err.Error()))
		return
	}

	// 2. 获取题目样例
	sendSSE(c, "log", "正在获取样例...")
	var problem *client.ProblemDetail
	var err error
	var displayID string
	var dbCID string
	var submitCID int

	if req.Mode == "contest" {
		displayID = strings.ToUpper(req.PID)
		dbCID = req.CID
		submitCID, _ = strconv.Atoi(req.CID)
		problem, err = bingoJClient.GetContestProblemDetail(displayID, req.CID)
	} else {
		displayID = req.PID
		dbCID = "0"
		submitCID = 0
		problem, err = bingoJClient.GetProblemDetail(req.PID)
	}

	if err != nil {
		sendSSE(c, "log", fmt.Sprintf("获取题目失败: %s", err.Error()))
		return
	}

	localInfo := "跳过 (无样例)"

	// 3. 本地样例测试
	if problem.Examples != "" {
		samples, err := service.ExtractSamples(problem.Examples)
		if err == nil && len(samples) > 0 {
			sendSSE(c, "log", fmt.Sprintf("开始自测 %d 组样例...", len(samples)))

			// 将 PID 转换为 int64
			pidInt, _ := strconv.ParseInt(req.PID, 10, 64)
			results, err := h.judgeService.TestLocalSamples(pidInt, req.Language, req.Code, req.Username, req.Password, samples)
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

	sendSSE(c, "log", fmt.Sprintf("提交ID: %s，判题中...", submitID))

	// 轮询判题结果
	for i := 0; i < 20; i++ {
		result, err := bingoJClient.GetSubmissionResult(submitID)
		if err != nil {
			continue
		}

		statusText := client.GetStatusText(result.Status)

		// 判题中
		if result.Status == 6 || result.Status == 7 || result.Status == 9 {
			sendSSE(c, "remote_status", map[string]string{"status": statusText})
			continue
		}

		// 判题完成
		sendSSE(c, "remote_status", map[string]string{"status": statusText})
		sendSSE(c, "log", fmt.Sprintf("最终结果: %s", statusText))

		// 保存到数据库
		_ = h.judgeService.SaveSubmissionHistory(
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
		)

		return
	}

	sendSSE(c, "log", "查询超时")
	sendSSE(c, "remote_status", map[string]string{"status": "超时"})
}
