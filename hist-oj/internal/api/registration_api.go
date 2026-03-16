package api

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/hoj/hist-oj/internal/client"
	"github.com/hoj/hist-oj/internal/model"
	"github.com/hoj/hist-oj/internal/utils"
)

// ==================== 用户端接口 ====================

// GetCompetitions 获取比赛列表
func (h *Handler) GetCompetitions(c *gin.Context) {
	logger := utils.GetLogger()
	db := client.GetDB()

	showHidden := c.DefaultQuery("show_hidden", "false") == "true"

	var competitions []model.Competition
	var err error

	if showHidden {
		// 管理员查看所有比赛
		err = db.Order("created_at DESC").Find(&competitions).Error
	} else {
		// 普通用户只看可见的比赛
		err = db.Where("visible = ?", true).Order("created_at DESC").Find(&competitions).Error
	}

	if err != nil {
		logger.Error("获取比赛列表失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "获取比赛列表失败"))
		return
	}

	c.JSON(http.StatusOK, successResponse(competitions))
}

// GetCompetition 获取单个比赛详情
func (h *Handler) GetCompetition(c *gin.Context) {
	logger := utils.GetLogger()
	db := client.GetDB()

	competitionID := c.Param("competitionId")

	var competition model.Competition
	err := db.Where("id = ?", competitionID).First(&competition).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, errorResponse(404, "比赛不存在"))
			return
		}
		logger.Error("获取比赛详情失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "获取比赛详情失败"))
		return
	}

	// 检查可见性（如果是管理员则可以查看隐藏的比赛）
	if !competition.Visible {
		// 检查是否为管理员
		isAdmin := false
		if userInfo, exists := c.Get("user"); exists {
			if user, ok := userInfo.(*model.UserInfo); ok {
				// 这里需要检查用户是否有管理员权限
				// 暂时简化处理，后续可以结合权限系统
				_ = user
				isAdmin = true
			}
		}
		if !isAdmin {
			c.JSON(http.StatusOK, errorResponse(403, "无权访问该比赛"))
			return
		}
	}

	c.JSON(http.StatusOK, successResponse(competition))
}

// CreateRegistration 提交报名
func (h *Handler) CreateRegistration(c *gin.Context) {
	logger := utils.GetLogger()
	db := client.GetDB()

	// 获取当前用户信息
	userInfo, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "请先登录"))
		return
	}
	user := userInfo.(*model.UserInfo)

	// 使用 snake_case 结构接收前端数据（兼容原有前端）
	var req struct {
		CompetitionID uint64 `json:"competition_id" binding:"required"`
		UserUUID     string `json:"user_uuid"`
		Name         string `json:"name"`
		Class        string `json:"class"`
		College      string `json:"college"`
		StudentID    string `json:"student_id"`
		Gender       string `json:"gender"`
		ShirtSize    string `json:"shirt_size"`
		TeamName     string `json:"team_name"`
		QQ           string `json:"qq"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("参数解析失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	// 覆盖 user_uuid
	req.UserUUID = user.UUID

	// 检查比赛是否存在以及时间是否允许报名
	var competition model.Competition
	err := db.Where("id = ?", req.CompetitionID).First(&competition).Error
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(404, "比赛不存在"))
		return
	}

	now := time.Now()
	if now.Before(competition.StartTime) {
		c.JSON(http.StatusOK, errorResponse(400, "比赛尚未开始报名"))
		return
	}
	if now.After(competition.EndTime) {
		c.JSON(http.StatusOK, errorResponse(400, "报名时间已截止"))
		return
	}

	// 检查是否已经报名过
	var existingReg model.Registration
	err = db.Where("competition_id = ? AND user_uuid = ?", req.CompetitionID, req.UserUUID).First(&existingReg).Error
	if err == nil {
		c.JSON(http.StatusOK, errorResponse(400, "您已经报名过该比赛"))
		return
	}

	// 创建报名记录（转换为 model）
	registration := model.Registration{
		CompetitionID: req.CompetitionID,
		UserUUID:     req.UserUUID,
		Name:         req.Name,
		Class:        req.Class,
		College:      req.College,
		StudentID:    req.StudentID,
		Gender:       req.Gender,
		ShirtSize:    req.ShirtSize,
		TeamName:     req.TeamName,
		QQ:           req.QQ,
		Status:       "pending",
	}
	if err := db.Create(&registration).Error; err != nil {
		logger.Error("创建报名记录失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "报名失败"))
		return
	}

	logger.Info("用户报名成功",
		zap.String("user", user.Username),
		zap.Uint64("competition", req.CompetitionID))

	c.JSON(http.StatusOK, successResponse(registration))
}

// GetMyRegistration 获取我在某个比赛的报名信息
func (h *Handler) GetMyRegistration(c *gin.Context) {
	logger := utils.GetLogger()
	db := client.GetDB()

	userInfo, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "请先登录"))
		return
	}
	user := userInfo.(*model.UserInfo)

	competitionID := c.Param("competitionId")

	var registration model.Registration
	err := db.Where("competition_id = ? AND user_uuid = ?", competitionID, user.UUID).
		First(&registration).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, successResponse(nil))
			return
		}
		logger.Error("获取报名信息失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "获取报名信息失败"))
		return
	}

	c.JSON(http.StatusOK, successResponse(registration))
}

// UpdateRegistration 更新报名信息（用户修改自己的报名）
func (h *Handler) UpdateRegistration(c *gin.Context) {
	logger := utils.GetLogger()
	db := client.GetDB()

	userInfo, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "请先登录"))
		return
	}
	user := userInfo.(*model.UserInfo)

	registrationID := c.Param("id")

	var req struct {
		Name             string     `json:"name"`
		Class            string     `json:"class"`
		College          string     `json:"college"`
		StudentID        string     `json:"student_id"`
		Gender           string     `json:"gender"`
		ShirtSize        string     `json:"shirt_size"`
		TeamName         string     `json:"team_name"`
		QQ               string     `json:"qq"`
		Status           string     `json:"status"`
		Remark           string     `json:"remark"`
		LastViewTime     *time.Time `json:"last_view_time"`
		AdminLastViewTime *time.Time `json:"admin_last_view_time"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	// 查询报名记录
	var registration model.Registration
	err := db.Where("id = ?", registrationID).First(&registration).Error
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(404, "报名记录不存在"))
		return
	}

	// 验证是否为本人
	if registration.UserUUID != user.UUID {
		c.JSON(http.StatusOK, errorResponse(403, "无权修改他人的报名信息"))
		return
	}

	// 只有被退回的报名才允许修改
	if registration.Status != "rejected" {
		// 检查比赛时间
		var competition model.Competition
		err = db.Where("id = ?", registration.CompetitionID).First(&competition).Error
		if err != nil {
			c.JSON(http.StatusOK, errorResponse(404, "比赛不存在"))
			return
		}

		now := time.Now()
		if now.Before(competition.StartTime) {
			c.JSON(http.StatusOK, errorResponse(400, "比赛尚未开始报名"))
			return
		}
		if now.After(competition.EndTime) {
			c.JSON(http.StatusOK, errorResponse(400, "报名时间已截止"))
			return
		}
	}

	// 更新字段
	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Class != "" {
		updates["class"] = req.Class
	}
	if req.College != "" {
		updates["college"] = req.College
	}
	if req.StudentID != "" {
		updates["student_id"] = req.StudentID
	}
	if req.Gender != "" {
		updates["gender"] = req.Gender
	}
	if req.ShirtSize != "" {
		updates["shirt_size"] = req.ShirtSize
	}
	if req.TeamName != "" {
		updates["team_name"] = req.TeamName
	}
	if req.QQ != "" {
		updates["qq"] = req.QQ
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.Remark != "" {
		updates["remark"] = req.Remark
	}
	if req.LastViewTime != nil {
		updates["last_view_time"] = *req.LastViewTime
	}
	if req.AdminLastViewTime != nil {
		updates["admin_last_view_time"] = *req.AdminLastViewTime
	}

	if err := db.Model(&registration).Updates(updates).Error; err != nil {
		logger.Error("更新报名信息失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "更新失败"))
		return
	}

	// WebSocket 广播更新（如果有新消息）
	if req.Remark != "" {
		logger.Info("用户更新消息，广播WebSocket",
			zap.Uint64("registration_id", registration.ID),
			zap.Uint64("competition_id", registration.CompetitionID))
		h.wsHub.BroadcastRegistrationUpdate(registration.CompetitionID, registration)
	}

	c.JSON(http.StatusOK, successResponse(nil))
}

// ==================== 管理端接口 ====================

// AdminGetCompetitions 管理员获取所有比赛（包括隐藏的）
func (h *Handler) AdminGetCompetitions(c *gin.Context) {
	logger := utils.GetLogger()
	db := client.GetDB()

	var competitions []model.Competition
	if err := db.Unscoped().Order("created_at DESC").Find(&competitions).Error; err != nil {
		logger.Error("获取比赛列表失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "获取比赛列表失败"))
		return
	}

	c.JSON(http.StatusOK, successResponse(competitions))
}

// CreateCompetition 创建比赛
func (h *Handler) CreateCompetition(c *gin.Context) {
	logger := utils.GetLogger()
	db := client.GetDB()

	var req struct {
		Name        string `json:"name" binding:"required"`
		StartTime   string `json:"startTime" binding:"required"`
		EndTime     string `json:"endTime" binding:"required"`
		Fields      string `json:"fields" binding:"required"`
		LogoURL     string `json:"logoUrl"`
		Description string `json:"description"`
		Visible     bool   `json:"visible"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("参数解析失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	// 解析时间字符串
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		logger.Warn("startTime格式错误", zap.Error(err), zap.String("startTime", req.StartTime))
		c.JSON(http.StatusOK, errorResponse(400, "startTime格式错误，应为ISO 8601格式"))
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		logger.Warn("endTime格式错误", zap.Error(err), zap.String("endTime", req.EndTime))
		c.JSON(http.StatusOK, errorResponse(400, "endTime格式错误，应为ISO 8601格式"))
		return
	}

	competition := model.Competition{
		Name:        req.Name,
		StartTime:   startTime,
		EndTime:     endTime,
		Fields:      req.Fields,
		LogoURL:     req.LogoURL,
		Description: req.Description,
		Visible:     req.Visible,
	}

	if err := db.Create(&competition).Error; err != nil {
		logger.Error("创建比赛失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "创建比赛失败"))
		return
	}

	logger.Info("管理员创建比赛", zap.String("name", competition.Name))
	c.JSON(http.StatusOK, successResponse(competition))
}

// UpdateCompetition 更新比赛
func (h *Handler) UpdateCompetition(c *gin.Context) {
	logger := utils.GetLogger()
	db := client.GetDB()

	competitionID := c.Param("competitionId")

	var req struct {
		Name        string `json:"name"`
		StartTime   string `json:"startTime"`
		EndTime     string `json:"endTime"`
		Fields      string `json:"fields"`
		LogoURL     string `json:"logoUrl"`
		Description string `json:"description"`
		Visible     *bool  `json:"visible"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("参数解析失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	// 查询比赛是否存在
	var competition model.Competition
	if err := db.Where("id = ?", competitionID).First(&competition).Error; err != nil {
		c.JSON(http.StatusOK, errorResponse(404, "比赛不存在"))
		return
	}

	// 构建更新字段
	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.StartTime != "" {
		startTime, err := time.Parse(time.RFC3339, req.StartTime)
		if err != nil {
			logger.Warn("startTime格式错误", zap.Error(err), zap.String("startTime", req.StartTime))
			c.JSON(http.StatusOK, errorResponse(400, "startTime格式错误"))
			return
		}
		updates["start_time"] = startTime
	}
	if req.EndTime != "" {
		endTime, err := time.Parse(time.RFC3339, req.EndTime)
		if err != nil {
			logger.Warn("endTime格式错误", zap.Error(err), zap.String("endTime", req.EndTime))
			c.JSON(http.StatusOK, errorResponse(400, "endTime格式错误"))
			return
		}
		updates["end_time"] = endTime
	}
	if req.Fields != "" {
		updates["fields"] = req.Fields
	}
	if req.LogoURL != "" {
		updates["logo_url"] = req.LogoURL
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Visible != nil {
		updates["visible"] = *req.Visible
	}

	if err := db.Model(&competition).Updates(updates).Error; err != nil {
		logger.Error("更新比赛失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "更新比赛失败"))
		return
	}

	// 重新查询返回更新后的数据
	db.Where("id = ?", competitionID).First(&competition)

	// WebSocket 广播比赛更新（通知所有在线用户刷新数据）
	logger.Info("管理员更新比赛信息，广播WebSocket",
		zap.String("competition_id", competitionID),
		zap.String("name", competition.Name))
	h.wsHub.BroadcastCompetitionUpdate(competition.ID, competition)

	c.JSON(http.StatusOK, successResponse(competition))
}

// DeleteCompetition 删除比赛
func (h *Handler) DeleteCompetition(c *gin.Context) {
	logger := utils.GetLogger()
	db := client.GetDB()

	competitionID := c.Param("competitionId")

	// 先删除该比赛的所有报名
	if err := db.Where("competition_id = ?", competitionID).Delete(&model.Registration{}).Error; err != nil {
		logger.Error("删除比赛报名记录失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "删除比赛失败"))
		return
	}

	// 删除比赛
	if err := db.Delete(&model.Competition{}, competitionID).Error; err != nil {
		logger.Error("删除比赛失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "删除比赛失败"))
		return
	}

	logger.Info("管理员删除比赛", zap.String("id", competitionID))
	c.JSON(http.StatusOK, successResponse(nil))
}

// UpdateVisibility 更新比赛可见性
func (h *Handler) UpdateVisibility(c *gin.Context) {
	logger := utils.GetLogger()
	db := client.GetDB()

	competitionID := c.Param("competitionId")

	var req struct {
		Visible bool `json:"visible"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	if err := db.Model(&model.Competition{}).
		Where("id = ?", competitionID).
		Update("visible", req.Visible).Error; err != nil {
		logger.Error("更新比赛可见性失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "更新可见性失败"))
		return
	}

	c.JSON(http.StatusOK, successResponse(nil))
}

// GetRegistrations 获取比赛的所有报名（管理员）
func (h *Handler) GetRegistrations(c *gin.Context) {
	logger := utils.GetLogger()
	db := client.GetDB()

	competitionID := c.Param("competitionId")

	var registrations []model.Registration
	if err := db.Where("competition_id = ?", competitionID).
		Order("created_at DESC").
		Find(&registrations).Error; err != nil {
		logger.Error("获取报名列表失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "获取报名列表失败"))
		return
	}

	// 添加日志：打印每个报名的 admin_last_view_time
	for _, reg := range registrations {
		adminViewTime := time.Time{}
		if reg.AdminLastViewTime != nil {
			adminViewTime = *reg.AdminLastViewTime
		}
		logger.Info("报名数据",
			zap.Uint64("id", reg.ID),
			zap.String("name", reg.Name),
			zap.Time("admin_last_view_time", adminViewTime),
			zap.Bool("has_remark", reg.Remark != ""))
	}

	c.JSON(http.StatusOK, successResponse(registrations))
}

// UpdateRegistrationStatus 更新报名状态（管理员审核）
func (h *Handler) UpdateRegistrationStatus(c *gin.Context) {
	logger := utils.GetLogger()
	db := client.GetDB()

	registrationID := c.Param("id")

	var req struct {
		Status              string `json:"status" binding:"omitempty,oneof=pending approved rejected"`
		Remark              string `json:"remark"`
		AdminLastViewTime   string `json:"admin_last_view_time"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	logger.Info("UpdateRegistrationStatus调用",
		zap.String("registrationID", registrationID),
		zap.String("status", req.Status),
		zap.Bool("hasRemark", req.Remark != ""),
		zap.Int("remarkLength", len(req.Remark)),
		zap.Bool("hasAdminLastViewTime", req.AdminLastViewTime != ""))

	// 更新报名状态
	updates := map[string]interface{}{}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.Remark != "" {
		updates["remark"] = req.Remark
	}
	if req.AdminLastViewTime != "" {
		// 解析 ISO 8601 时间格式（前端发送的是 UTC 时间）
		// 前端发送格式: "2026-03-13T01:31:36.000Z"
		// 我们需要保持 UTC 时间，不要转换成当地时间
		parsedTime, err := time.Parse(time.RFC3339, req.AdminLastViewTime)
		if err != nil {
			logger.Warn("解析admin_last_view_time失败，使用原始值", zap.Error(err), zap.String("value", req.AdminLastViewTime))
			updates["admin_last_view_time"] = req.AdminLastViewTime
		} else {
			// 直接存储 time.Time 对象，GORM 会自动处理为 UTC
			// 这确保了存储的是 UTC 时间，读取时也是 UTC 时间
			updates["admin_last_view_time"] = parsedTime
			logger.Info("存储admin_last_view_time", zap.Time("parsed_time", parsedTime), zap.String("formatted", parsedTime.Format(time.RFC3339)))
		}
	}

	logger.Info("准备更新数据", zap.Any("updates", updates))

	result := db.Model(&model.Registration{}).
		Where("id = ?", registrationID).
		Updates(updates)

	if result.Error != nil {
		logger.Error("更新报名状态失败", zap.Error(result.Error))
		c.JSON(http.StatusOK, errorResponse(500, "更新报名状态失败"))
		return
	}

	logger.Info("更新报名状态成功",
		zap.String("id", registrationID),
		zap.Int64("rowsAffected", result.RowsAffected),
		zap.String("status", req.Status),
		zap.Bool("hasRemark", req.Remark != ""),
		zap.Bool("hasAdminLastViewTime", req.AdminLastViewTime != ""))

	// WebSocket 广播更新（如果有新消息或状态变更）
	if req.Remark != "" || req.Status != "" {
		// 获取比赛ID用于广播
		var registration model.Registration
		if err := db.Where("id = ?", registrationID).First(&registration).Error; err == nil {
			// 广播更新给所有管理员
			h.wsHub.BroadcastRegistrationUpdate(registration.CompetitionID, registration)
		}
	}

	c.JSON(http.StatusOK, successResponse(nil))
}

// UploadLogo 上传比赛Logo
func (h *Handler) UploadLogo(c *gin.Context) {
	logger := utils.GetLogger()

	// 限制上传文件大小为 10MB
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 10<<20)

	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "获取文件失败"))
		return
	}

	// 检查文件类型
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}

	// 打开文件进行类型检查
	fileReader, err := file.Open()
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(500, "打开文件失败"))
		return
	}
	defer fileReader.Close()

	// 读取前512字节用于检测文件类型
	buffer := make([]byte, 512)
	_, err = fileReader.Read(buffer)
	if err != nil && err != io.EOF {
		c.JSON(http.StatusOK, errorResponse(500, "读取文件失败"))
		return
	}
	fileType := http.DetectContentType(buffer)

	if !allowedTypes[fileType] {
		c.JSON(http.StatusOK, errorResponse(400, "不支持的文件类型，仅支持 jpg、png、gif、webp"))
		return
	}

	// 创建上传目录
	uploadDir := "./uploads/logos"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		logger.Error("创建上传目录失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "创建上传目录失败"))
		return
	}

	// 生成唯一文件名
	ext := filepath.Ext(file.Filename)
	hash := md5.New()
	hash.Write([]byte(fmt.Sprintf("%d%s", time.Now().UnixNano(), file.Filename)))
	filename := hex.EncodeToString(hash.Sum(nil)) + ext

	// 保存文件
	if err := c.SaveUploadedFile(file, filepath.Join(uploadDir, filename)); err != nil {
		logger.Error("保存文件失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "保存文件失败"))
		return
	}

	// 返回文件URL
	fileURL := fmt.Sprintf("/uploads/logos/%s", filename)

	logger.Info("上传比赛Logo", zap.String("url", fileURL))
	c.JSON(http.StatusOK, successResponse(gin.H{
		"url":      fileURL,
		"filename": filename,
	}))
}

// HojAutoLogin HOJ用户自动登录
func (h *Handler) HojAutoLogin(c *gin.Context) {
	logger := utils.GetLogger()
	db := client.GetDB()

	token := c.Query("token")
	username := c.Query("username")

	if token == "" || username == "" {
		c.JSON(http.StatusOK, errorResponse(400, "缺少token或username参数"))
		return
	}

	// 查找用户，如果不存在则自动创建
	var user model.UserInfo
	err := db.Where("username = ?", username).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 用户不存在，自动创建
			newUUID := strings.ReplaceAll(uuid.New().String(), "-", "")
			user = model.UserInfo{
				UUID:     newUUID,
				Username: username,
				Password: "", // 自动登录不需要密码
			}
			if err := db.Create(&user).Error; err != nil {
				logger.Error("创建用户失败", zap.Error(err))
				c.JSON(http.StatusOK, errorResponse(500, "创建用户失败"))
				return
			}
			logger.Info("自动创建新用户", zap.String("username", username))
		} else {
			logger.Error("数据库查询错误", zap.Error(err))
			c.JSON(http.StatusOK, errorResponse(500, "数据库查询错误"))
			return
		}
	}

	c.JSON(http.StatusOK, successResponse(gin.H{
		"uuid":     user.UUID,
		"username": user.Username,
	}))
}
