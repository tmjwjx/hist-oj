package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/hoj/hist-oj/internal/service"
	"github.com/hoj/hist-oj/internal/utils"
)

type Handler struct {
	ratingService *service.RatingService
	queryService  *service.QueryService
	scheduler     Scheduler
}

// Scheduler 定时任务接口
type Scheduler interface {
	TriggerCheck()
}

func NewHandler(ratingService *service.RatingService, queryService *service.QueryService) *Handler {
	return &Handler{
		ratingService: ratingService,
		queryService:  queryService,
	}
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
		"contestId":   contestID,
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

	// 如果没有指定初始rating，使用默认值1200
	if req.InitialRating == 0 {
		req.InitialRating = 1200
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

