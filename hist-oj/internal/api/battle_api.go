package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/hoj/hist-oj/internal/client"
	"github.com/hoj/hist-oj/internal/model"
	"github.com/hoj/hist-oj/internal/service"
	"github.com/hoj/hist-oj/internal/utils"
)

// BattleAPI 对战API处理器
type BattleAPI struct {
	battleService *service.BattleService
}

// NewBattleAPI 创建对战API
func NewBattleAPI() *BattleAPI {
	return &BattleAPI{
		battleService: service.NewBattleService(),
	}
}

// CreateRoomRequest 创建房间请求
type CreateRoomRequest struct {
	UserID   string `json:"userId"`
	Username string `json:"username"`
}

// JoinRoomRequest 加入房间请求
type JoinRoomRequest struct {
	RoomID   string `json:"roomId" binding:"required"`
	UserID   string `json:"userId" binding:"required"`
	Username string `json:"username" binding:"required"`
}

// ReadyBattleRequest 准备对战请求
type ReadyBattleRequest struct {
	RoomID string `json:"roomId" binding:"required"`
	UserID string `json:"userId" binding:"required"`
	Ready  bool   `json:"ready"` // true-准备, false-取消准备
}

// StartBattleRequest 开始对战请求
type StartBattleRequest struct {
	RoomID string `json:"roomId" binding:"required"`
}

// GiveupBattleRequest 放弃对战请求
type GiveupBattleRequest struct {
	RoomID   string `json:"roomId" binding:"required"`
	UserID   string `json:"userId" binding:"required"`
	Username string `json:"username" binding:"required"`
}

// DissolveRoomRequest 解散房间请求
type DissolveRoomRequest struct {
	RoomID string `json:"roomId" binding:"required"`
	UserID string `json:"userId" binding:"required"`
}

// LeaveRoomRequest 退出房间请求
type LeaveRoomRequest struct {
	RoomID string `json:"roomId" binding:"required"`
	UserID string `json:"userId" binding:"required"`
}

// ResetRoomRequest 重置房间请求（再来一局）
type ResetRoomRequest struct {
	RoomID string `json:"roomId" binding:"required"`
	UserID string `json:"userId" binding:"required"`
}

// SubmitACRequest AC提交请求
type SubmitACRequest struct {
	RoomID   string `json:"roomId" binding:"required"`
	UserID   string `json:"userId" binding:"required"`
	ProblemID string `json:"problemId" binding:"required"`
}

// CommonResponse 通用响应
type CommonResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"msg"`
	Data    interface{} `json:"data,omitempty"`
}

// CreateRoom 创建对战房间
// @Summary 创建对战房间
// @Description 创建一个1v1代码对战房间
// @Tags Battle
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} CommonResponse
// @Router /api/battle/create-room [post]
func (api *BattleAPI) CreateRoom(c *gin.Context) {
	logger := utils.GetLogger()

	// 从请求中获取用户信息
	var req CreateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 如果请求体为空，尝试使用默认用户
		req.UserID = "guest_user"
		req.Username = "Guest"
	}

	// 创建房间
	room, err := api.battleService.CreateRoom(req.UserID, req.Username)
	if err != nil {
		logger.Warn("创建房间失败",
			zap.String("user_id", req.UserID),
			zap.Error(err))
		c.JSON(http.StatusOK, CommonResponse{
			Code:    1,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, CommonResponse{
		Code:    0,
		Message: "创建成功",
		Data:    room,
	})
}

// JoinRoom 加入对战房间
// @Summary 加入对战房间
// @Description 通过房间号加入对战房间
// @Tags Battle
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body JoinRoomRequest true "加入房间请求"
// @Success 200 {object} CommonResponse
// @Router /api/battle/join-room [post]
func (api *BattleAPI) JoinRoom(c *gin.Context) {
	logger := utils.GetLogger()

	var req JoinRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("请求参数绑定失败",
			zap.Error(err))
		c.JSON(http.StatusOK, CommonResponse{
			Code:    1,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	logger.Info("收到加入房间请求",
		zap.String("room_id", req.RoomID),
		zap.String("user_id", req.UserID),
		zap.String("username", req.Username))

	// 加入房间
	room, err := api.battleService.JoinRoom(req.RoomID, req.UserID, req.Username)
	if err != nil {
		logger.Warn("加入房间失败",
			zap.String("room_id", req.RoomID),
			zap.String("user_id", req.UserID),
			zap.Error(err))
		c.JSON(http.StatusOK, CommonResponse{
			Code:    1,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, CommonResponse{
		Code:    0,
		Message: "加入成功",
		Data:    room,
	})
}

// ReadyBattle 准备对战
// @Summary 准备对战
// @Description 挑战者准备/取消准备
// @Tags Battle
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ReadyBattleRequest true "准备对战请求"
// @Success 200 {object} CommonResponse
// @Router /api/battle/ready [post]
func (api *BattleAPI) ReadyBattle(c *gin.Context) {
	logger := utils.GetLogger()

	var req ReadyBattleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("请求参数绑定失败",
			zap.Error(err))
		c.JSON(http.StatusOK, CommonResponse{
			Code:    1,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 准备/取消准备
	room, err := api.battleService.ReadyBattle(req.RoomID, req.UserID, req.Ready)
	if err != nil {
		logger.Warn("准备对战失败",
			zap.String("room_id", req.RoomID),
			zap.String("user_id", req.UserID),
			zap.Error(err))
		c.JSON(http.StatusOK, CommonResponse{
			Code:    1,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, CommonResponse{
		Code:    0,
		Message: map[bool]string{true: "已准备", false: "已取消准备"}[req.Ready],
		Data:    room,
	})
}

// GetRoomInfo 获取房间信息
// @Summary 获取房间信息
// @Description 获取对战房间的详细信息
// @Tags Battle
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param roomId query string true "房间号"
// @Success 200 {object} CommonResponse
// @Router /api/battle/room-info [get]
func (api *BattleAPI) GetRoomInfo(c *gin.Context) {
	roomID := c.Query("roomId")
	if roomID == "" {
		c.JSON(http.StatusOK, CommonResponse{
			Code:    1,
			Message: "房间号不能为空",
		})
		return
	}

	room, err := api.battleService.GetRoomInfo(roomID)
	if err != nil {
		c.JSON(http.StatusOK, CommonResponse{
			Code:    1,
			Message: err.Error(),
		})
		return
	}

	// 如果房间已开始对战，还需要返回题目信息
	resp := map[string]interface{}{
		"room": room,
	}

	if room.Status == 1 && room.ProblemID != nil {
		// 通过显示ID获取题目信息
		problem, err := client.GetProblemInfoByDisplayID(*room.ProblemID)
		if err == nil {
			resp["problem"] = problem
		}
		// 如果获取题目信息失败，也不影响返回房间信息
	}

	c.JSON(http.StatusOK, CommonResponse{
		Code:    0,
		Message: "查询成功",
		Data:    resp,
	})
}

// StartBattle 开始对战
// @Summary 开始对战
// @Description 房主开始对战，系统将选择对战题目
// @Tags Battle
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body StartBattleRequest true "开始对战请求"
// @Success 200 {object} CommonResponse
// @Router /api/battle/start-battle [post]
func (api *BattleAPI) StartBattle(c *gin.Context) {
	logger := utils.GetLogger()

	var req StartBattleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, CommonResponse{
			Code:    1,
			Message: "请求参数错误",
		})
		return
	}

	// 开始对战
	room, problem, err := api.battleService.StartBattle(req.RoomID)
	if err != nil {
		logger.Warn("开始对战失败",
			zap.String("room_id", req.RoomID),
			zap.Error(err))
		c.JSON(http.StatusOK, CommonResponse{
			Code:    1,
			Message: err.Error(),
		})
		return
	}

	// 构造响应数据
	resp := map[string]interface{}{
		"room":     room,
		"problem":  problem,
	}

	c.JSON(http.StatusOK, CommonResponse{
		Code:    0,
		Message: "对战开始",
		Data:    resp,
	})
}

// GiveupBattle 放弃对战
// @Summary 放弃对战
// @Description 放弃当前对战，判定对方获胜
// @Tags Battle
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body GiveupBattleRequest true "放弃对战请求"
// @Success 200 {object} CommonResponse
// @Router /api/battle/giveup [post]
func (api *BattleAPI) GiveupBattle(c *gin.Context) {
	logger := utils.GetLogger()

	var req GiveupBattleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("请求参数绑定失败",
			zap.Error(err))
		c.JSON(http.StatusOK, CommonResponse{
			Code:    1,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	logger.Info("收到放弃对战请求",
		zap.String("room_id", req.RoomID),
		zap.String("user_id", req.UserID))

	// 放弃对战
	winner, endReason, err := api.battleService.GiveupBattle(req.RoomID, req.UserID)
	if err != nil {
		logger.Warn("放弃对战失败",
			zap.String("room_id", req.RoomID),
			zap.String("user_id", req.UserID),
			zap.Error(err))
		c.JSON(http.StatusOK, CommonResponse{
			Code:    1,
			Message: err.Error(),
		})
		return
	}

	// 获取胜者信息
	winnerUsername := ""
	if room, err := api.battleService.GetRoomInfo(req.RoomID); err == nil {
		if room.HostID == winner {
			winnerUsername = room.HostUsername
		} else if room.ChallengerID != nil && *room.ChallengerID == winner {
			winnerUsername = *room.ChallengerUsername
		}
	}

	resp := map[string]interface{}{
		"winner":     winnerUsername,
		"winnerId":   winner,
		"endReason":  endReason,
	}

	c.JSON(http.StatusOK, CommonResponse{
		Code:    0,
		Message: "已放弃对战",
		Data:    resp,
	})
}

// DissolveRoom 解散房间
// @Summary 解散房间
// @Description 房主解散等待中的房间
// @Tags Battle
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body DissolveRoomRequest true "解散房间请求"
// @Success 200 {object} CommonResponse
// @Router /api/battle/dissolve-room [post]
func (api *BattleAPI) DissolveRoom(c *gin.Context) {
	logger := utils.GetLogger()

	var req DissolveRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("请求参数绑定失败",
			zap.Error(err))
		c.JSON(http.StatusOK, CommonResponse{
			Code:    1,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	logger.Info("收到解散房间请求",
		zap.String("room_id", req.RoomID),
		zap.String("user_id", req.UserID))

	// 解散房间
	err := api.battleService.DissolveRoom(req.RoomID, req.UserID)
	if err != nil {
		logger.Warn("解散房间失败",
			zap.String("room_id", req.RoomID),
			zap.String("user_id", req.UserID),
			zap.Error(err))
		c.JSON(http.StatusOK, CommonResponse{
			Code:    1,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, CommonResponse{
		Code:    0,
		Message: "房间已解散",
	})
}

// LeaveRoom 退出房间
// @Summary 退出房间
// @Description 挑战者退出等待中的房间
// @Tags Battle
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body LeaveRoomRequest true "退出房间请求"
// @Success 200 {object} CommonResponse
// @Router /api/battle/leave-room [post]
func (api *BattleAPI) LeaveRoom(c *gin.Context) {
	logger := utils.GetLogger()

	var req LeaveRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("请求参数绑定失败",
			zap.Error(err))
		c.JSON(http.StatusOK, CommonResponse{
			Code:    1,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	logger.Info("收到退出房间请求",
		zap.String("room_id", req.RoomID),
		zap.String("user_id", req.UserID))

	// 退出房间
	err := api.battleService.LeaveRoom(req.RoomID, req.UserID)
	if err != nil {
		logger.Warn("退出房间失败",
			zap.String("room_id", req.RoomID),
			zap.String("user_id", req.UserID),
			zap.Error(err))
		c.JSON(http.StatusOK, CommonResponse{
			Code:    1,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, CommonResponse{
		Code:    0,
		Message: "已退出房间",
	})
}

// ResetRoom 重置房间（再来一局）
// @Summary 重置房间
// @Description 重置已结束的房间，开始新一局
// @Tags Battle
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ResetRoomRequest true "重置房间请求"
// @Success 200 {object} CommonResponse
// @Router /api/battle/reset-room [post]
func (api *BattleAPI) ResetRoom(c *gin.Context) {
	logger := utils.GetLogger()

	var req ResetRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("请求参数绑定失败",
			zap.Error(err))
		c.JSON(http.StatusOK, CommonResponse{
			Code:    1,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	logger.Info("收到重置房间请求",
		zap.String("room_id", req.RoomID),
		zap.String("user_id", req.UserID))

	// 重置房间
	err := api.battleService.ResetRoom(req.RoomID, req.UserID)
	if err != nil {
		logger.Warn("重置房间失败",
			zap.String("room_id", req.RoomID),
			zap.String("user_id", req.UserID),
			zap.Error(err))
		c.JSON(http.StatusOK, CommonResponse{
			Code:    1,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, CommonResponse{
		Code:    0,
		Message: "房间已重置，可以开始新一局",
	})
}

// SubmitAC AC提交
// @Summary AC提交
// @Description 用户在题目中AC后，通知对战系统结束对战
// @Tags Battle
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body SubmitACRequest true "AC提交请求"
// @Success 200 {object} CommonResponse
// @Router /api/battle/submit-ac [post]
func (api *BattleAPI) SubmitAC(c *gin.Context) {
	logger := utils.GetLogger()

	var req SubmitACRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("请求参数绑定失败",
			zap.Error(err))
		c.JSON(http.StatusOK, CommonResponse{
			Code:    1,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	logger.Info("收到AC提交",
		zap.String("room_id", req.RoomID),
		zap.String("user_id", req.UserID),
		zap.String("problem_id", req.ProblemID))

	// 处理AC提交
	err := api.battleService.SubmitAC(req.RoomID, req.UserID, req.ProblemID)
	if err != nil {
		logger.Warn("处理AC提交失败",
			zap.String("room_id", req.RoomID),
			zap.String("user_id", req.UserID),
			zap.Error(err))
		c.JSON(http.StatusOK, CommonResponse{
			Code:    1,
			Message: err.Error(),
		})
		return
	}

	// 获取房间信息以判断胜负
	room, err := api.battleService.GetRoomInfo(req.RoomID)
	if err != nil {
		c.JSON(http.StatusOK, CommonResponse{
			Code:    0,
			Message: "AC已记录",
		})
		return
	}

	// 判断当前用户是否获胜
	isWinner := room.WinnerID != nil && *room.WinnerID == req.UserID

	resp := map[string]interface{}{
		"isWinner":  isWinner,
		"winnerId":  room.WinnerID,
		"endReason": room.EndReason,
	}

	c.JSON(http.StatusOK, CommonResponse{
		Code:    0,
		Message: "AC已记录",
		Data:    resp,
	})
}

// GetMyRecords 获取我的对战记录
// @Summary 获取我的对战记录
// @Description 分页获取当前用户的对战记录
// @Tags Battle
// @Accept json
// @Produce json
// @Param userId query string true "用户ID"
// @Param limit query int false "每页数量" default(10)
// @Param currentPage query int false "当前页" default(1)
// @Success 200 {object} CommonResponse
// @Router /api/battle/my-records [get]
func (api *BattleAPI) GetMyRecords(c *gin.Context) {
	// 从查询参数获取用户ID
	userID := c.Query("userId")
	if userID == "" {
		c.JSON(http.StatusOK, CommonResponse{
			Code:    1,
			Message: "用户ID不能为空",
		})
		return
	}

	// 获取分页参数
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	currentPage, _ := strconv.Atoi(c.DefaultQuery("currentPage", "1"))

	if limit <= 0 {
		limit = 10
	}
	if currentPage <= 0 {
		currentPage = 1
	}

	// 查询对战记录
	records, total, err := api.battleService.GetMyBattleRecords(userID, limit, currentPage)
	if err != nil {
		c.JSON(http.StatusOK, CommonResponse{
			Code:    1,
			Message: "查询失败",
		})
		return
	}

	resp := map[string]interface{}{
		"total":   total,
		"records": records,
	}

	c.JSON(http.StatusOK, CommonResponse{
		Code:    0,
		Message: "查询成功",
		Data:    resp,
	})
}

// GetAllRecords 获取所有对战记录（管理员功能）
// @Summary 获取所有对战记录
// @Description 管理员查询所有对战记录，支持按用户名、房间号、胜负筛选
// @Tags Battle
// @Accept json
// @Produce json
// @Param limit query int false "每页数量" default(20)
// @Param currentPage query int false "当前页" default(1)
// @Param username query string false "用户名筛选"
// @Param roomId query string false "房间号筛选"
// @Param isWinner query string false "胜负筛选 true/false"
// @Success 200 {object} CommonResponse
// @Router /api/battle/all-records [get]
func (api *BattleAPI) GetAllRecords(c *gin.Context) {
	// 获取分页参数
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	currentPage, _ := strconv.Atoi(c.DefaultQuery("currentPage", "1"))

	if limit <= 0 {
		limit = 20
	}
	if currentPage <= 0 {
		currentPage = 1
	}

	// 获取筛选参数
	username := c.Query("username")
	roomId := c.Query("roomId")
	isWinnerStr := c.Query("isWinner")

	// 查询对战记录
	records, total, err := api.battleService.GetAllBattleRecords(limit, currentPage, username, roomId, isWinnerStr)
	if err != nil {
		c.JSON(http.StatusOK, CommonResponse{
			Code:    1,
			Message: "查询失败",
		})
		return
	}

	resp := map[string]interface{}{
		"total":   total,
		"records": records,
	}

	c.JSON(http.StatusOK, CommonResponse{
		Code:    0,
		Message: "查询成功",
		Data:    resp,
	})
}

// GetRank 获取对战排行榜
// @Summary 获取对战排行榜
// @Description 分页获取对战排行榜
// @Tags Battle
// @Accept json
// @Produce json
// @Param limit query int false "每页数量" default(50)
// @Param currentPage query int false "当前页" default(1)
// @Success 200 {object} CommonResponse
// @Router /api/battle/rank [get]
func (api *BattleAPI) GetRank(c *gin.Context) {
	// 获取分页参数
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	currentPage, _ := strconv.Atoi(c.DefaultQuery("currentPage", "1"))
	username := c.Query("username") // 获取用户名搜索参数

	if limit <= 0 {
		limit = 50
	}
	if currentPage <= 0 {
		currentPage = 1
	}

	// 查询排行榜
	rankList, total, err := api.battleService.GetBattleRank(limit, currentPage, username)
	if err != nil {
		c.JSON(http.StatusOK, CommonResponse{
			Code:    1,
			Message: "查询失败",
		})
		return
	}

	// 添加排名信息
	type RankItem struct {
		model.UserBattleStats
		Rank int `json:"rank"`
	}

	result := make([]RankItem, 0, len(rankList))
	for i, stats := range rankList {
		result = append(result, RankItem{
			UserBattleStats: stats,
			Rank:            (currentPage-1)*limit + i + 1,
		})
	}

	resp := map[string]interface{}{
		"total":    total,
		"rankList": result,
	}

	c.JSON(http.StatusOK, CommonResponse{
		Code:    0,
		Message: "查询成功",
		Data:    resp,
	})
}

// RegisterBattleRoutes 注册对战相关路由
func RegisterBattleRoutes(router *gin.RouterGroup) {
	battleAPI := NewBattleAPI()

	// 所有对战路由（暂时不需要认证，由前端控制）
	battle := router.Group("/battle")
	{
		battle.POST("/create-room", battleAPI.CreateRoom)
		battle.POST("/join-room", battleAPI.JoinRoom)
		battle.POST("/ready", battleAPI.ReadyBattle) // 准备/取消准备
		battle.GET("/room-info", battleAPI.GetRoomInfo)
		battle.POST("/start-battle", battleAPI.StartBattle)
		battle.POST("/giveup", battleAPI.GiveupBattle)
		battle.POST("/dissolve-room", battleAPI.DissolveRoom)
		battle.POST("/leave-room", battleAPI.LeaveRoom)
		battle.POST("/reset-room", battleAPI.ResetRoom)
		battle.POST("/submit-ac", battleAPI.SubmitAC)
		battle.GET("/my-records", battleAPI.GetMyRecords)
		battle.GET("/all-records", battleAPI.GetAllRecords)
		battle.GET("/rank", battleAPI.GetRank)
	}

	// 管理员对战路由（需要管理员权限）
	adminBattle := router.Group("/admin/battle")
	{
		adminBattle.PUT("/record/exclude", AuthMiddleware(), battleAPI.ExcludeRecord) // 标记不计本场对决
	}
}

// ExcludeRecord 标记不计本场对决（管理员功能）
func (api *BattleAPI) ExcludeRecord(c *gin.Context) {
	logger := utils.GetLogger()

	var req struct {
		RecordID   int64 `json:"recordId" binding:"required"`
		IsExcluded bool  `json:"isExcluded"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	// 调用服务层标记记录
	err := api.battleService.ExcludeRecord(req.RecordID, req.IsExcluded)
	if err != nil {
		logger.Error("标记对决记录失败", zap.Error(err), zap.Int64("record_id", req.RecordID))
		c.JSON(http.StatusOK, errorResponse(500, "操作失败"))
		return
	}

	logger.Info("管理员标记对决记录",
		zap.Int64("record_id", req.RecordID),
		zap.Bool("is_excluded", req.IsExcluded))

	c.JSON(http.StatusOK, successResponse(nil))
}
