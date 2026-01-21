package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/hoj/hist-oj/internal/service"
	"github.com/hoj/hist-oj/internal/utils"
)

// JoinTraining 用户参加训练
func (h *Handler) JoinTraining(c *gin.Context) {
	logger := utils.GetLogger()

	// 获取训练ID
	trainingIDStr := c.Param("trainingId")
	trainingID, err := strconv.ParseUint(trainingIDStr, 10, 64)
	if err != nil {
		logger.Warn("trainingId参数格式错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	// 获取当前用户ID
	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "用户未登录"))
		return
	}

	// 调用服务层参加训练
	trainingService := service.NewTrainingService()
	record, err := trainingService.JoinTraining(trainingID, uid.(string))
	if err != nil {
		logger.Error("参加训练失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "参加训练失败"))
		return
	}

	logger.Info("用户参加训练成功", zap.Uint64("training_id", trainingID), zap.String("uid", uid.(string)))
	c.JSON(http.StatusOK, successResponse(record))
}

// GetTrainingParticipants 获取训练参与者列表
func (h *Handler) GetTrainingParticipants(c *gin.Context) {
	logger := utils.GetLogger()

	// 获取训练ID
	trainingIDStr := c.Param("trainingId")
	trainingID, err := strconv.ParseUint(trainingIDStr, 10, 64)
	if err != nil {
		logger.Warn("trainingId参数格式错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	// 调用服务层获取参与者列表
	trainingService := service.NewTrainingService()
	records, err := trainingService.GetTrainingParticipants(trainingID)
	if err != nil {
		logger.Error("获取训练参与者列表失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "获取失败"))
		return
	}

	c.JSON(http.StatusOK, successResponse(records))
}

// GetMyTrainingParticipant 获取用户在某个训练的记录
func (h *Handler) GetMyTrainingParticipant(c *gin.Context) {
	logger := utils.GetLogger()

	// 获取训练ID
	trainingIDStr := c.Param("trainingId")
	trainingID, err := strconv.ParseUint(trainingIDStr, 10, 64)
	if err != nil {
		logger.Warn("trainingId参数格式错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	// 获取当前用户ID
	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "用户未登录"))
		return
	}

	// 调用服务层获取记录
	trainingService := service.NewTrainingService()
	record, err := trainingService.GetTrainingParticipantByUser(trainingID, uid.(string))
	if err != nil {
		logger.Error("获取训练记录失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(404, "未找到训练记录"))
		return
	}

	c.JSON(http.StatusOK, successResponse(record))
}

// UpdateTrainingParticipantStatus 更新训练参与者状态(手动触发)
func (h *Handler) UpdateTrainingParticipantStatus(c *gin.Context) {
	logger := utils.GetLogger()

	// 获取训练ID
	trainingIDStr := c.Param("trainingId")
	trainingID, err := strconv.ParseUint(trainingIDStr, 10, 64)
	if err != nil {
		logger.Warn("trainingId参数格式错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	// 获取当前用户ID
	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "用户未登录"))
		return
	}

	// 调用服务层更新状态
	trainingService := service.NewTrainingService()
	if err := trainingService.UpdateTrainingProgressBySubmission(trainingID, uid.(string)); err != nil {
		logger.Error("更新训练状态失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "更新失败"))
		return
	}

	c.JSON(http.StatusOK, successResponse(nil))
}

// GetMyTrainingProgress 获取用户在所有训练中的进度
func (h *Handler) GetMyTrainingProgress(c *gin.Context) {
	logger := utils.GetLogger()

	// 获取当前用户ID
	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "用户未登录"))
		return
	}

	// 调用服务层获取进度
	trainingService := service.NewTrainingService()
	progress, err := trainingService.GetUserTrainingProgress(uid.(string))
	if err != nil {
		logger.Error("获取训练进度失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "获取失败"))
		return
	}

	c.JSON(http.StatusOK, successResponse(progress))
}

// RegisterTrainingRoutes 注册训练相关路由
func RegisterTrainingRoutes(router *gin.RouterGroup) {
	h := &Handler{}

	training := router.Group("/training")
	{
		// 需要认证的训练接口
		training.POST("/:trainingId/join", AuthMiddleware(), h.JoinTraining)
		training.GET("/:trainingId/participants", AuthMiddleware(), h.GetTrainingParticipants)
		training.GET("/:trainingId/my-record", AuthMiddleware(), h.GetMyTrainingParticipant)
		training.PUT("/:trainingId/status", AuthMiddleware(), h.UpdateTrainingParticipantStatus)
		training.GET("/my-progress", AuthMiddleware(), h.GetMyTrainingProgress)
	}
}
