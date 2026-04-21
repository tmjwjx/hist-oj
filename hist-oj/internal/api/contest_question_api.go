package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/hoj/hist-oj/internal/client"
	"github.com/hoj/hist-oj/internal/middleware"
	"github.com/hoj/hist-oj/internal/model"
	"github.com/hoj/hist-oj/internal/utils"
)

// ==================== 比赛问题答疑 ====================

// CreateContestQuestion 创建比赛问题（选手）
func (h *Handler) CreateContestQuestion(c *gin.Context) {
	logger := utils.GetLogger()

	var req struct {
		ContestID uint64 `json:"contestId" binding:"required"`
		Title     string `json:"title" binding:"required"`
		Content   string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	// 获取当前用户ID
	uid, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "未登录"))
		return
	}

	db := client.GetDB()

	// 检查比赛是否存在
	var contest model.Contest
	if err := db.Where("id = ?", req.ContestID).First(&contest).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, errorResponse(404, "比赛不存在"))
		} else {
			logger.Error("查询比赛失败", zap.Error(err))
			c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		}
		return
	}

	// 创建问题
	question := &model.ContestQuestion{
		ContestID:    req.ContestID,
		QuestionerID: uid.(string),
		Title:        req.Title,
		Content:      req.Content,
		Status:       "pending",
		Priority:     0,
	}

	if err := db.Create(question).Error; err != nil {
		logger.Error("创建问题失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "创建失败"))
		return
	}

	// 加载提问者信息
	db.Preload("Questioner").First(question)

	// 填充 HistRating
	if question.Questioner != nil {
		fillUserHistRatings(db, []*model.UserInfo{question.Questioner})
	}

	logger.Info("创建比赛问题", zap.Uint64("contest_id", req.ContestID), zap.String("questioner_id", uid.(string)))

	c.JSON(http.StatusOK, successResponse(question))
}

// GetContestQuestions 获取比赛问题列表
func (h *Handler) GetContestQuestions(c *gin.Context) {
	logger := utils.GetLogger()

	contestIDStr := c.Param("contestId")
	contestID, err := strconv.ParseUint(contestIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "contestId参数格式错误"))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")              // 可选筛选条件
	allQuestions := c.Query("all") == "true" // 是否查看所有问题（用于管理员答疑列表）

	// 获取当前用户ID
	uid, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "未登录"))
		return
	}

	db := client.GetDB()

	// 检查用户是否为比赛管理员
	isContestAdmin := checkContestAdmin(db, uid.(string), contestID)

	logger.Info("GetContestQuestions",
		zap.String("uid", uid.(string)),
		zap.Uint64("contestId", contestID),
		zap.Bool("isContestAdmin", isContestAdmin),
		zap.Bool("allQuestions", allQuestions))

	var questions []model.ContestQuestion
	var total int64

	query := db.Model(&model.ContestQuestion{}).Where("contest_id = ?", contestID)

	// 权限检查
	if allQuestions {
		// 答疑列表：只有管理员可以看到所有问题
		if !isContestAdmin {
			c.JSON(http.StatusOK, errorResponse(403, "无权限访问"))
			return
		}
		// 管理员可以看到所有问题，不需要过滤
	} else {
		// 问题答疑：所有人（包括管理员）都只能看到自己的问题
		query = query.Where("questioner_id = ?", uid.(string))
	}

	// 状态筛选
	if status != "" && status != "all" {
		query = query.Where("status = ?", status)
	}

	// 获取总数
	query.Count(&total)

	// 获取问题列表
	if err := query.Preload("Questioner").
		Order("created_at DESC").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&questions).Error; err != nil {
		logger.Error("查询问题列表失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	// 填充 HistRating
	questioners := make([]*model.UserInfo, 0, len(questions))
	for i := range questions {
		if questions[i].Questioner != nil {
			questioners = append(questioners, questions[i].Questioner)
		}
	}
	fillUserHistRatings(db, questioners)

	// 为每个问题加载回复数量
	for i := range questions {
		var replyCount int64
		db.Model(&model.ContestQuestionReply{}).Where("question_id = ?", questions[i].ID).Count(&replyCount)
		// 将回复数量存储到一个非数据库字段
		questions[i].Replies = make([]model.ContestQuestionReply, 0)
	}

	c.JSON(http.StatusOK, successResponse(map[string]interface{}{
		"list":  questions,
		"total": total,
	}))
}

// GetContestQuestionDetail 获取比赛问题详情及对话记录
func (h *Handler) GetContestQuestionDetail(c *gin.Context) {
	logger := utils.GetLogger()

	questionIDStr := c.Param("questionId")
	questionID, err := strconv.ParseUint(questionIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "questionId参数格式错误"))
		return
	}

	// 获取当前用户ID
	uid, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "未登录"))
		return
	}

	db := client.GetDB()

	var question model.ContestQuestion
	if err := db.Preload("Questioner").First(&question, questionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Error("问题不存在", zap.Uint64("questionId", questionID))
			c.JSON(http.StatusOK, errorResponse(404, "问题不存在"))
		} else {
			logger.Error("查询问题失败", zap.Error(err))
			c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		}
		return
	}

	logger.Info("GetContestQuestionDetail",
		zap.String("uid", uid.(string)),
		zap.Uint64("questionId", questionID),
		zap.String("questionerId", question.QuestionerID),
		zap.Uint64("contestId", question.ContestID))

	// 权限检查：问题创建者或比赛管理员可以访问
	isContestAdmin := checkContestAdmin(db, uid.(string), question.ContestID)
	logger.Info("GetContestQuestionDetail 权限检查",
		zap.Bool("isContestAdmin", isContestAdmin),
		zap.Bool("isQuestioner", question.QuestionerID == uid.(string)))

	if question.QuestionerID != uid.(string) && !isContestAdmin {
		logger.Warn("GetContestQuestionDetail 权限不足",
			zap.String("uid", uid.(string)),
			zap.String("questionerId", question.QuestionerID),
			zap.Bool("isContestAdmin", isContestAdmin))
		c.JSON(http.StatusOK, errorResponse(403, "无权限访问"))
		return
	}

	// 加载回复记录
	var replies []model.ContestQuestionReply
	if err := db.Where("question_id = ?", questionID).
		Preload("Sender").
		Order("created_at ASC").
		Find(&replies).Error; err != nil {
		logger.Error("查询回复记录失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	// 填充 HistRating（Questioner 和 Senders）
	allUsers := make([]*model.UserInfo, 0, len(replies)+1)
	if question.Questioner != nil {
		allUsers = append(allUsers, question.Questioner)
	}
	for i := range replies {
		if replies[i].Sender != nil {
			allUsers = append(allUsers, replies[i].Sender)
		}
	}
	fillUserHistRatings(db, allUsers)

	logger.Info("查询回复成功",
		zap.Int("replyCount", len(replies)),
		zap.Uint64("questionId", questionID),
		zap.String("uid", uid.(string)))

	// 调试日志：检查每个回复的Sender信息
	for i, reply := range replies {
		if reply.Sender != nil {
			logger.Debug("回复Sender信息",
				zap.Int("index", i),
				zap.String("senderUsername", reply.Sender.Username),
				zap.String("senderNickname", reply.Sender.Nickname))
		} else {
			logger.Warn("回复Sender为空",
				zap.Int("index", i),
				zap.Uint64("replyId", reply.ID),
				zap.String("senderId", reply.SenderID))
		}
	}

	question.Replies = replies

	logger.Info("返回问题详情",
		zap.Uint64("questionId", questionID),
		zap.String("questionTitle", question.Title),
		zap.Int("replyCount", len(question.Replies)),
		zap.Bool("hasQuestioner", question.Questioner != nil))

	c.JSON(http.StatusOK, successResponse(question))
}

// SendQuestionReply 发送回复
func (h *Handler) SendQuestionReply(c *gin.Context) {
	logger := utils.GetLogger()

	questionIDStr := c.Param("questionId")
	questionID, err := strconv.ParseUint(questionIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "questionId参数格式错误"))
		return
	}

	var req struct {
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	if req.Content == "" {
		c.JSON(http.StatusOK, errorResponse(400, "回复内容不能为空"))
		return
	}

	// 获取当前用户ID
	uid, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "未登录"))
		return
	}

	db := client.GetDB()

	// 查询问题
	var question model.ContestQuestion
	if err := db.First(&question, questionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, errorResponse(404, "问题不存在"))
		} else {
			logger.Error("查询问题失败", zap.Error(err))
			c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		}
		return
	}

	// 检查权限：问题创建者或比赛管理员
	isContestAdmin := checkContestAdmin(db, uid.(string), question.ContestID)
	if question.QuestionerID != uid.(string) && !isContestAdmin {
		c.JSON(http.StatusOK, errorResponse(403, "无权限访问"))
		return
	}

	// 创建回复
	reply := &model.ContestQuestionReply{
		QuestionID: questionID,
		SenderID:   uid.(string),
		Content:    req.Content,
	}

	if err := db.Create(reply).Error; err != nil {
		logger.Error("发送回复失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "发送失败"))
		return
	}

	logger.Info("创建回复成功",
		zap.Uint64("replyId", reply.ID),
		zap.Uint64("questionId", questionID),
		zap.String("senderId", uid.(string)))

	// 如果问题状态是pending，更新为answered
	if question.Status == "pending" {
		db.Model(&question).Update("status", "answered")
	}

	// 重新加载回复，包含Sender信息
	if err := db.Preload("Sender").First(reply, reply.ID).Error; err != nil {
		logger.Error("加载回复Sender信息失败", zap.Error(err))
	} else {
		// 填充 HistRating
		if reply.Sender != nil {
			fillUserHistRatings(db, []*model.UserInfo{reply.Sender})
		}
		logger.Info("加载回复Sender信息成功",
			zap.Uint64("replyId", reply.ID),
			zap.Bool("hasSender", reply.Sender != nil),
			zap.String("senderUsername", func() string {
				if reply.Sender != nil {
					return reply.Sender.Username
				}
				return ""
			}()))
	}

	// TODO: 通过WebSocket或轮询通知对方

	c.JSON(http.StatusOK, successResponse(reply))
}

// UpdateQuestionStatus 更新问题状态（管理员）
func (h *Handler) UpdateQuestionStatus(c *gin.Context) {
	logger := utils.GetLogger()

	questionIDStr := c.Param("questionId")
	questionID, err := strconv.ParseUint(questionIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "questionId参数格式错误"))
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	// 验证状态值
	if req.Status != "pending" && req.Status != "answered" && req.Status != "closed" {
		c.JSON(http.StatusOK, errorResponse(400, "无效的状态值"))
		return
	}

	// 获取当前用户ID
	uid, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "未登录"))
		return
	}

	db := client.GetDB()

	// 查询问题
	var question model.ContestQuestion
	if err := db.First(&question, questionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, errorResponse(404, "问题不存在"))
		} else {
			logger.Error("查询问题失败", zap.Error(err))
			c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		}
		return
	}

	// 检查权限：仅比赛管理员可以更新状态
	if !checkContestAdmin(db, uid.(string), question.ContestID) {
		c.JSON(http.StatusOK, errorResponse(403, "无权限操作"))
		return
	}

	// 更新状态
	if err := db.Model(&question).Update("status", req.Status).Error; err != nil {
		logger.Error("更新问题状态失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "更新失败"))
		return
	}

	logger.Info("更新比赛问题状态", zap.Uint64("question_id", questionID), zap.String("status", req.Status))

	c.JSON(http.StatusOK, successResponse(nil))
}

// DeleteContestQuestion 删除问题
func (h *Handler) DeleteContestQuestion(c *gin.Context) {
	logger := utils.GetLogger()

	questionIDStr := c.Param("questionId")
	questionID, err := strconv.ParseUint(questionIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "questionId参数格式错误"))
		return
	}

	// 获取当前用户ID
	uid, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "未登录"))
		return
	}

	db := client.GetDB()

	// 查询问题
	var question model.ContestQuestion
	if err := db.First(&question, questionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, errorResponse(404, "问题不存在"))
		} else {
			logger.Error("查询问题失败", zap.Error(err))
			c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		}
		return
	}

	// 检查权限：问题创建者或比赛管理员
	isContestAdmin := checkContestAdmin(db, uid.(string), question.ContestID)
	if question.QuestionerID != uid.(string) && !isContestAdmin {
		c.JSON(http.StatusOK, errorResponse(403, "无权限操作"))
		return
	}

	// 删除问题（会级联删除回复）
	if err := db.Delete(&question).Error; err != nil {
		logger.Error("删除问题失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "删除失败"))
		return
	}

	logger.Info("删除比赛问题", zap.Uint64("question_id", questionID))

	c.JSON(http.StatusOK, successResponse(nil))
}

// checkContestAdmin 检查用户是否为比赛管理员
func checkContestAdmin(db *gorm.DB, uid string, contestID uint64) bool {
	logger := utils.GetLogger()

	// 查询比赛信息
	var contest model.Contest
	if err := db.Where("id = ?", contestID).First(&contest).Error; err != nil {
		logger.Error("查询比赛失败", zap.Error(err), zap.Uint64("contestId", contestID))
		return false
	}

	// 查询用户信息
	var user model.UserInfo
	if err := db.Where("uuid = ?", uid).First(&user).Error; err != nil {
		logger.Error("查询用户失败", zap.Error(err), zap.String("uid", uid))
		return false
	}

	// 检查是否为比赛创建者
	if contest.Author == user.Username {
		logger.Info("checkContestAdmin: 是比赛创建者",
			zap.String("contestAuthor", contest.Author),
			zap.String("userUsername", user.Username))
		return true
	}

	// 特殊处理：username为root的用户自动拥有超级管理员权限
	if user.Username == "root" || user.Username == "admin" {
		logger.Info("checkContestAdmin: 是系统管理员",
			zap.String("userUsername", user.Username))
		return true
	}

	// 检查是否为超级管理员（仅root，不包括admin和problem_admin）
	roles, err := middleware.GetUserRoles(db, uid)
	if err != nil {
		logger.Error("checkContestAdmin: 查询用户角色失败", zap.Error(err), zap.String("uid", uid))
		return false
	}

	if middleware.HasRole(roles, middleware.RoleRoot) {
		logger.Info("checkContestAdmin: 是超级管理员",
			zap.String("userUsername", user.Username),
			zap.Strings("roles", roles))
		return true
	}

	logger.Info("checkContestAdmin: 无权限",
		zap.String("contestAuthor", contest.Author),
		zap.String("userUsername", user.Username),
		zap.Strings("roles", roles))

	return false
}

// fillUserHistRatings 批量填充用户的 HistRating（从 user_record 表）
func fillUserHistRatings(db *gorm.DB, users []*model.UserInfo) {
	if len(users) == 0 {
		return
	}

	// 收集所有 UUID
	uuids := make([]string, 0, len(users))
	for _, user := range users {
		if user != nil {
			uuids = append(uuids, user.UUID)
		}
	}

	if len(uuids) == 0 {
		return
	}

	// 批量查询 user_record
	var records []model.UserRecord
	if err := db.Where("uid IN ?", uuids).Find(&records).Error; err != nil {
		return
	}

	// 创建 UUID -> HistRating 映射
	ratingMap := make(map[string]int)
	for _, record := range records {
		if record.HistRating != nil {
			ratingMap[record.UID] = *record.HistRating
		}
	}

	// 填充 HistRating
	for _, user := range users {
		if user != nil {
			if histRating, ok := ratingMap[user.UUID]; ok {
				user.HistRating = histRating
			} else {
				user.HistRating = 0 // 默认值
			}
		}
	}
}
