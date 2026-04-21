package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/hoj/hist-oj/internal/service"
	"github.com/hoj/hist-oj/internal/utils"
	"github.com/hoj/hist-oj/internal/websocket"
)

type Handler struct {
	ratingService *service.RatingService
	queryService  *service.QueryService
	scheduler     Scheduler
	judgeService  *service.JudgeService
	wsHub         *websocket.Hub
}

// Scheduler 定时任务接口
type Scheduler interface {
	TriggerCheck()
}

func NewHandler(ratingService *service.RatingService, queryService *service.QueryService, wsHub *websocket.Hub) *Handler {
	return &Handler{
		ratingService: ratingService,
		queryService:  queryService,
		wsHub:         wsHub,
	}
}

// SetJudgeService 设置判题服务
func (h *Handler) SetJudgeService(judgeService *service.JudgeService) {
	h.judgeService = judgeService
}

// SetScheduler 设置定时任务调度器
func (h *Handler) SetScheduler(scheduler Scheduler) {
	h.scheduler = scheduler
}

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func successResponse(data interface{}) Response {
	return Response{
		Code:    200,
		Message: "success",
		Data:    data,
	}
}

func errorResponse(code int, message string) Response {
	return Response{
		Code:    code,
		Message: message,
	}
}

// GetUserRating 获取用户rating信息
func (h *Handler) GetUserRating(c *gin.Context) {
	logger := utils.GetLogger()
	uid := c.Param("uid")

	if uid == "" {
		logger.Warn("请求参数错误", zap.String("param", "uid"), zap.String("value", uid))
		c.JSON(http.StatusOK, errorResponse(400, "uid参数不能为空"))
		return
	}

	// 检查是否有认证信息（可选认证）
	currentUserId, _ := c.Get("userId")
	logger.Info("获取用户rating",
		zap.String("uid", uid),
		zap.Any("current_user_id", currentUserId))

	result, err := h.queryService.GetUserRating(uid)
	if err != nil {
		logger.Error("查询用户rating失败", zap.String("uid", uid), zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败: "+err.Error()))
		return
	}

	if result == nil {
		logger.Warn("用户不存在", zap.String("uid", uid))
		c.JSON(http.StatusOK, errorResponse(404, "用户不存在"))
		return
	}

	logger.Debug("获取用户rating成功", zap.String("uid", uid))
	c.JSON(http.StatusOK, successResponse(result))
}

// GetRatingHistory 获取用户rating历史记录
func (h *Handler) GetRatingHistory(c *gin.Context) {
	logger := utils.GetLogger()
	uid := c.Param("uid")

	if uid == "" {
		logger.Warn("请求参数错误", zap.String("param", "uid"))
		c.JSON(http.StatusOK, errorResponse(400, "uid参数不能为空"))
		return
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	logger.Info("获取用户rating历史",
		zap.String("uid", uid),
		zap.Int("page", page),
		zap.Int("limit", limit))

	result, err := h.queryService.GetRatingHistory(uid, page, limit)
	if err != nil {
		logger.Error("查询rating历史失败",
			zap.String("uid", uid),
			zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, successResponse(result))
}

// GetRatingColor 获取rating颜色信息
func (h *Handler) GetRatingColor(c *gin.Context) {
	logger := utils.GetLogger()
	ratingStr := c.Param("rating")
	rating, err := strconv.Atoi(ratingStr)
	if err != nil {
		logger.Warn("请求参数错误",
			zap.String("param", "rating"),
			zap.String("value", ratingStr),
			zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "rating参数格式错误"))
		return
	}

	result := h.queryService.GetRatingColor(rating)
	c.JSON(http.StatusOK, successResponse(result))
}

// CalculateRating 手动触发rating计算
func (h *Handler) CalculateRating(c *gin.Context) {
	logger := utils.GetLogger()
	contestIDStr := c.Param("contestId")
	contestID, err := strconv.ParseInt(contestIDStr, 10, 64)
	if err != nil || contestID <= 0 {
		logger.Warn("请求参数错误",
			zap.String("param", "contestId"),
			zap.String("value", contestIDStr),
			zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "contestId参数格式错误"))
		return
	}

	logger.Info("手动触发rating计算", zap.Int64("contest_id", contestID))

	if !h.ratingService.CanCalculateRating(contestID) {
		logger.Warn("比赛不能计算rating", zap.Int64("contest_id", contestID))
		c.JSON(http.StatusOK, errorResponse(400, "该比赛不能计算rating或已经计算过"))
		return
	}

	histories, err := h.ratingService.CalculateContestRating(contestID)
	if err != nil {
		logger.Error("计算rating失败",
			zap.Int64("contest_id", contestID),
			zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "计算失败: "+err.Error()))
		return
	}

	result := map[string]interface{}{
		"contestId":    contestID,
		"participants": len(histories),
	}

	logger.Info("手动触发rating计算完成",
		zap.Int64("contest_id", contestID),
		zap.Int("participants", len(histories)))
	c.JSON(http.StatusOK, successResponse(result))
}

// GetContestParticipantsRating 获取比赛所有参赛者的rating信息
func (h *Handler) GetContestParticipantsRating(c *gin.Context) {
	logger := utils.GetLogger()
	contestIDStr := c.Param("contestId")
	contestID, err := strconv.ParseInt(contestIDStr, 10, 64)
	if err != nil || contestID <= 0 {
		logger.Warn("请求参数错误",
			zap.String("param", "contestId"),
			zap.String("value", contestIDStr))
		c.JSON(http.StatusOK, errorResponse(400, "contestId参数格式错误"))
		return
	}

	results, err := h.queryService.GetContestParticipantsRating(contestID)
	if err != nil {
		logger.Error("查询比赛参赛者rating失败",
			zap.Int64("contest_id", contestID),
			zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	c.JSON(http.StatusOK, successResponse(map[string]interface{}{
		"contestId":    contestID,
		"participants": len(results),
		"records":      results,
	}))
}

// GetBatchUserRating 批量获取用户rating信息
func (h *Handler) GetBatchUserRating(c *gin.Context) {
	logger := utils.GetLogger()

	var req struct {
		UIDs []string `json:"uids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	if len(req.UIDs) == 0 {
		c.JSON(http.StatusOK, errorResponse(400, "uids不能为空"))
		return
	}

	if len(req.UIDs) > 100 {
		c.JSON(http.StatusOK, errorResponse(400, "一次最多查询100个用户"))
		return
	}

	results, err := h.queryService.GetBatchUserRating(req.UIDs)
	if err != nil {
		logger.Error("批量查询用户rating失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	c.JSON(http.StatusOK, successResponse(results))
}

// InitializeUserRating 初始化用户默认rating
func (h *Handler) InitializeUserRating(c *gin.Context) {
	logger := utils.GetLogger()

	var req struct {
		UID           string `json:"uid" binding:"required"`
		InitialRating int    `json:"initialRating"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	err := h.queryService.InitializeUserRating(req.UID, req.InitialRating)
	if err != nil {
		logger.Error("初始化用户rating失败",
			zap.String("uid", req.UID),
			zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "初始化失败: "+err.Error()))
		return
	}

	logger.Info("初始化用户rating成功",
		zap.String("uid", req.UID),
		zap.Int("rating", req.InitialRating))
	c.JSON(http.StatusOK, successResponse(map[string]interface{}{
		"uid":    req.UID,
		"rating": req.InitialRating,
	}))
}

// SetContestRatingType 设置比赛的Rating类型
func (h *Handler) SetContestRatingType(c *gin.Context) {
	logger := utils.GetLogger()

	var req struct {
		ContestID uint64 `json:"contestId" binding:"required"`
		IsRating  bool   `json:"isRating"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	err := h.queryService.SetContestRatingType(req.ContestID, req.IsRating)
	if err != nil {
		logger.Error("设置比赛Rating类型失败",
			zap.Uint64("contest_id", req.ContestID),
			zap.Bool("is_rating", req.IsRating),
			zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "设置失败: "+err.Error()))
		return
	}

	logger.Info("设置比赛Rating类型成功",
		zap.Uint64("contest_id", req.ContestID),
		zap.Bool("is_rating", req.IsRating))
	c.JSON(http.StatusOK, successResponse(map[string]interface{}{
		"contestId": req.ContestID,
		"isRating":  req.IsRating,
	}))
}

// GetContestInfo 获取比赛信息（包括是否为Rating比赛）
func (h *Handler) GetContestInfo(c *gin.Context) {
	logger := utils.GetLogger()
	contestIDStr := c.Param("contestId")
	contestID, err := strconv.ParseUint(contestIDStr, 10, 64)
	if err != nil || contestID <= 0 {
		logger.Warn("请求参数错误",
			zap.String("param", "contestId"),
			zap.String("value", contestIDStr))
		c.JSON(http.StatusOK, errorResponse(400, "contestId参数格式错误"))
		return
	}

	contestInfo, err := h.queryService.GetContestInfo(contestID)
	if err != nil {
		logger.Error("查询比赛信息失败",
			zap.Uint64("contest_id", contestID),
			zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	if contestInfo == nil {
		c.JSON(http.StatusOK, errorResponse(404, "比赛不存在"))
		return
	}

	c.JSON(http.StatusOK, successResponse(contestInfo))
}

// GetBatchContestInfo 批量获取比赛信息（包括是否为Rating比赛）
func (h *Handler) GetBatchContestInfo(c *gin.Context) {
	logger := utils.GetLogger()

	var req struct {
		ContestIDs []uint64 `json:"contestIds" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	if len(req.ContestIDs) == 0 {
		c.JSON(http.StatusOK, errorResponse(400, "contestIds不能为空"))
		return
	}

	if len(req.ContestIDs) > 50 {
		c.JSON(http.StatusOK, errorResponse(400, "一次最多查询50个比赛"))
		return
	}

	results, err := h.queryService.GetBatchContestInfo(req.ContestIDs)
	if err != nil {
		logger.Error("批量查询比赛信息失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	c.JSON(http.StatusOK, successResponse(results))
}

// GetRatingRank 获取Rating排名列表
func (h *Handler) GetRatingRank(c *gin.Context) {
	logger := utils.GetLogger()

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "30"))
	if err != nil || limit < 1 || limit > 100 {
		limit = 30
	}

	keyword := c.Query("keyword")

	logger.Info("获取Rating排名",
		zap.Int("page", page),
		zap.Int("limit", limit),
		zap.String("keyword", keyword))

	result, err := h.queryService.GetRatingRank(page, limit, keyword)
	if err != nil {
		logger.Error("查询Rating排名失败",
			zap.Int("page", page),
			zap.Int("limit", limit),
			zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, successResponse(result))
}

// TriggerScheduler 手动触发定时任务
func (h *Handler) TriggerScheduler(c *gin.Context) {
	logger := utils.GetLogger()

	if h.scheduler == nil {
		logger.Error("定时任务调度器未初始化")
		c.JSON(http.StatusOK, errorResponse(500, "定时任务调度器未初始化"))
		return
	}

	logger.Info("手动触发定时任务")

	// 在后台执行，避免阻塞请求
	go h.scheduler.TriggerCheck()

	c.JSON(http.StatusOK, successResponse(map[string]interface{}{
		"message": "定时任务已触发，正在后台执行",
	}))
}

// HealthCheck 健康检查
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

// AdjustUserRating 手动调整用户rating（管理员）
func (h *Handler) AdjustUserRating(c *gin.Context) {
	logger := utils.GetLogger()

	var req struct {
		Username         string  `json:"username" binding:"required"`
		RatingChange     int     `json:"ratingChange" binding:"required"`
		Reason           string  `json:"reason" binding:"required"`
		RelatedContestID *uint64 `json:"relatedContestId"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	// 获取操作人UID（从认证中间件中获取）
	operatorUID, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "未授权"))
		return
	}

	// 获取操作人用户名
	operatorUsername := ""
	if username, exists := c.Get("username"); exists {
		operatorUsername = username.(string)
	}

	logger.Info("手动调整用户rating请求",
		zap.String("username", req.Username),
		zap.Int("rating_change", req.RatingChange),
		zap.String("reason", req.Reason),
		zap.Any("related_contest_id", req.RelatedContestID),
		zap.String("operator_uid", operatorUID.(string)))

	oldRating, newRating, ratingChange, err := h.ratingService.AdjustUserRating(
		req.Username,
		req.RatingChange,
		req.Reason,
		req.RelatedContestID,
		operatorUID.(string),
		operatorUsername,
	)

	if err != nil {
		logger.Error("手动调整用户rating失败",
			zap.String("username", req.Username),
			zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "调整失败: "+err.Error()))
		return
	}

	result := map[string]interface{}{
		"username":         req.Username,
		"oldRating":        oldRating,
		"newRating":        newRating,
		"ratingChange":     ratingChange,
		"reason":           req.Reason,
		"relatedContestId": req.RelatedContestID,
		"operatorUID":      operatorUID.(string),
	}

	logger.Info("手动调整用户rating成功",
		zap.String("username", req.Username),
		zap.Int("old_rating", oldRating),
		zap.Int("new_rating", newRating),
		zap.Int("rating_change", ratingChange))

	c.JSON(http.StatusOK, successResponse(result))
}

// CancelManualAdjustment 撤销手动调整（管理员）
func (h *Handler) CancelManualAdjustment(c *gin.Context) {
	logger := utils.GetLogger()

	var req struct {
		AdjustmentID uint64 `json:"adjustmentId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	operatorUID, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "未授权"))
		return
	}
	operatorUsername := ""
	if username, exists := c.Get("username"); exists {
		operatorUsername = username.(string)
	}

	result, err := h.ratingService.CancelManualAdjustment(req.AdjustmentID, operatorUID.(string), operatorUsername)
	if err != nil {
		logger.Error("撤销手动调整失败",
			zap.Uint64("adjustment_id", req.AdjustmentID),
			zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "撤销失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, successResponse(result))
}

// GetManualAdjustmentHistory 获取手动调整历史（管理员）
func (h *Handler) GetManualAdjustmentHistory(c *gin.Context) {
	logger := utils.GetLogger()

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	logger.Info("查询手动调整历史",
		zap.Int("page", page),
		zap.Int("limit", limit))

	result, err := h.queryService.GetManualAdjustmentHistory(page, limit)
	if err != nil {
		logger.Error("查询手动调整历史失败",
			zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	c.JSON(http.StatusOK, successResponse(result))
}

// BatchSkipContestUsers 批量Skip用户
func (h *Handler) BatchSkipContestUsers(c *gin.Context) {
	logger := utils.GetLogger()

	var req struct {
		ContestID  uint64   `json:"contestId" binding:"required"`
		Usernames  []string `json:"usernames" binding:"required"`
		Reason     string   `json:"reason" binding:"required"`
		AutoRecalc bool     `json:"autoRecalc"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "请求参数错误: "+err.Error()))
		return
	}

	// 获取操作人信息（从认证中间件获取）
	operatorUID, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "未授权"))
		return
	}

	// 获取操作人用户名（优先从context获取，否则从数据库查询）
	operatorUsername := ""
	if username, exists := c.Get("username"); exists {
		operatorUsername = username.(string)
	}

	result, err := h.ratingService.BatchSkipContestUsers(
		int64(req.ContestID),
		req.Usernames,
		req.Reason,
		operatorUID.(string),
		operatorUsername,
		req.AutoRecalc,
	)

	if err != nil {
		logger.Error("批量skip用户失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, err.Error()))
		return
	}

	logger.Info("批量skip用户成功",
		zap.Uint64("contest_id", req.ContestID),
		zap.Int("success_count", len(result.SuccessUsers)),
		zap.Int("failed_count", len(result.FailedUsers)))

	c.JSON(http.StatusOK, successResponse(result))
}

// GetContestSkipUsers 获取比赛的Skip用户列表
func (h *Handler) GetContestSkipUsers(c *gin.Context) {
	logger := utils.GetLogger()

	contestIDStr := c.Param("contestId")
	contestID, err := strconv.ParseInt(contestIDStr, 10, 64)
	if err != nil {
		logger.Warn("比赛ID参数错误", zap.String("contestId", contestIDStr))
		c.JSON(http.StatusOK, errorResponse(400, "比赛ID参数错误"))
		return
	}

	skipUsers, err := h.ratingService.GetContestSkipUsers(contestID)
	if err != nil {
		logger.Error("获取skip用户列表失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	c.JSON(http.StatusOK, successResponse(skipUsers))
}

// CancelSkip 取消Skip
func (h *Handler) CancelSkip(c *gin.Context) {
	logger := utils.GetLogger()

	var req struct {
		ContestID uint64   `json:"contestId" binding:"required"`
		UIDs      []string `json:"uids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "请求参数错误: "+err.Error()))
		return
	}

	// 获取操作人UID
	operatorUID, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "未授权"))
		return
	}

	// 获取操作人用户名
	operatorUsername := ""
	if username, exists := c.Get("username"); exists {
		operatorUsername = username.(string)
	}

	err := h.ratingService.CancelSkip(
		int64(req.ContestID),
		req.UIDs,
		operatorUID.(string),
		operatorUsername,
	)

	if err != nil {
		errMsg := err.Error()
		// 检查是否是"已存在重算任务"的特殊情况
		if strings.Contains(errMsg, "已存在从比赛") && strings.Contains(errMsg, "的重算任务") {
			// 这种情况不算失败，只是提示用户已有重算任务
			logger.Info("取消Skip成功，但已存在覆盖范围的重算任务",
				zap.Uint64("contest_id", req.ContestID),
				zap.Int("uid_count", len(req.UIDs)),
				zap.String("existing_task_info", errMsg))
			c.JSON(http.StatusOK, gin.H{
				"code":    200,
				"message": errMsg,
				"data": gin.H{
					"hasExistingTask": true,
					"message":         "已存在覆盖范围的重算任务，系统会自动处理本次修改",
				},
			})
			return
		}

		// 其他错误情况
		logger.Error("取消skip失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, err.Error()))
		return
	}

	logger.Info("取消skip成功",
		zap.Uint64("contest_id", req.ContestID),
		zap.Int("uid_count", len(req.UIDs)))

	c.JSON(http.StatusOK, successResponse(gin.H{
		"message": "取消skip成功",
	}))
}

// GetRecalculateProgress 获取重算进度
func (h *Handler) GetRecalculateProgress(c *gin.Context) {
	logger := utils.GetLogger()

	taskIDStr := c.Param("taskId")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 64)
	if err != nil {
		logger.Warn("任务ID参数错误", zap.String("taskId", taskIDStr))
		c.JSON(http.StatusOK, errorResponse(400, "任务ID参数错误"))
		return
	}

	progress, err := h.ratingService.GetRecalculateProgress(taskID)
	if err != nil {
		logger.Error("获取重算进度失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	c.JSON(http.StatusOK, successResponse(progress))
}

// RecalculateFromContest 从指定比赛开始重算
func (h *Handler) RecalculateFromContest(c *gin.Context) {
	logger := utils.GetLogger()

	contestIDStr := c.Param("contestId")
	contestID, err := strconv.ParseInt(contestIDStr, 10, 64)
	if err != nil {
		logger.Warn("比赛ID参数错误", zap.String("contestId", contestIDStr))
		c.JSON(http.StatusOK, errorResponse(400, "比赛ID参数错误"))
		return
	}

	// 获取操作人UID
	operatorUID, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "未授权"))
		return
	}

	// 创建重算任务
	taskID, err := h.ratingService.CreateRecalculateTask(contestID, operatorUID.(string))
	if err != nil {
		logger.Error("创建重算任务失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, err.Error()))
		return
	}

	// 异步执行
	go h.ratingService.ExecuteRecalculateTask(taskID)

	logger.Info("创建重算任务成功", zap.Uint64("task_id", taskID))

	c.JSON(http.StatusOK, successResponse(gin.H{
		"taskId":  taskID,
		"message": "重算任务已创建",
	}))
}

// ResetRecalculateLock 重置比赛重算锁
func (h *Handler) ResetRecalculateLock(c *gin.Context) {
	logger := utils.GetLogger()

	contestIDStr := c.Param("contestId")
	contestID, err := strconv.ParseInt(contestIDStr, 10, 64)
	if err != nil {
		logger.Warn("比赛ID参数错误", zap.String("contestId", contestIDStr))
		c.JSON(http.StatusOK, errorResponse(400, "比赛ID参数错误"))
		return
	}

	// 获取操作人UID
	operatorUID, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "未授权"))
		return
	}

	// 检查是否有正在运行的重算任务
	hasRunningTask, err := h.ratingService.HasRunningRecalculateTask(contestID)
	if err != nil {
		logger.Error("检查重算任务失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "检查重算任务失败"))
		return
	}

	if hasRunningTask {
		logger.Warn("尝试重置锁，但有任务正在运行", zap.Int64("contest_id", contestID))
		c.JSON(http.StatusOK, errorResponse(400, "有重算任务正在运行，无法重置锁"))
		return
	}

	// 重置锁
	err = h.ratingService.ResetRecalculateLock(contestID, operatorUID.(string))
	if err != nil {
		logger.Error("重置重算锁失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, err.Error()))
		return
	}

	logger.Info("重置重算锁成功", zap.Int64("contest_id", contestID), zap.String("operator", operatorUID.(string)))

	c.JSON(http.StatusOK, successResponse(gin.H{
		"message": "重算锁已重置",
	}))
}

// SyncContestSkipFlag 同步比赛的 Skip 标记到 rating_history 表
func (h *Handler) SyncContestSkipFlag(c *gin.Context) {
	logger := utils.GetLogger()

	contestIDStr := c.Param("contestId")
	contestID, err := strconv.ParseInt(contestIDStr, 10, 64)
	if err != nil || contestID <= 0 {
		logger.Warn("请求参数错误",
			zap.String("param", "contestId"),
			zap.String("value", contestIDStr))
		c.JSON(http.StatusOK, errorResponse(400, "contestId参数格式错误"))
		return
	}

	// 调用 service 层同步 Skip 标记
	updatedCount, err := h.ratingService.SyncContestSkipFlag(contestID)
	if err != nil {
		logger.Error("同步Skip标记失败",
			zap.Int64("contest_id", contestID),
			zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, err.Error()))
		return
	}

	logger.Info("同步Skip标记成功",
		zap.Int64("contest_id", contestID),
		zap.Int("updated_count", updatedCount))

	c.JSON(http.StatusOK, successResponse(gin.H{
		"message":      "Skip标记同步成功",
		"contestId":    contestID,
		"updatedCount": updatedCount,
	}))
}

// GetOperationLogs 获取操作日志
func (h *Handler) GetOperationLogs(c *gin.Context) {
	logger := utils.GetLogger()

	// 获取查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	operationType := c.DefaultQuery("type", "")
	timeRange := c.DefaultQuery("timeRange", "7d")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// 获取操作人UID（用于权限验证）
	operatorUID, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "未授权"))
		return
	}
	_ = operatorUID // 后续可用于权限验证

	// 获取操作日志
	logs, total, err := h.ratingService.GetOperationLogs(page, limit, operationType, timeRange)
	if err != nil {
		logger.Error("获取操作日志失败",
			zap.Int("page", page),
			zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, err.Error()))
		return
	}

	logger.Info("获取操作日志成功",
		zap.Int("page", page),
		zap.Int("limit", limit),
		zap.String("operation_type", operationType),
		zap.String("time_range", timeRange),
		zap.Int64("total", total),
		zap.Int("count", len(logs)))

	c.JSON(http.StatusOK, successResponse(gin.H{
		"records": logs,
		"total":   total,
		"page":    page,
		"limit":   limit,
	}))
}
