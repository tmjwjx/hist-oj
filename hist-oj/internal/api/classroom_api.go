package api

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/hoj/hist-oj/internal/client"
	"github.com/hoj/hist-oj/internal/model"
	"github.com/hoj/hist-oj/internal/utils"
)

// ==================== 权限管理 ====================

// GrantUserRole 赋予用户角色（管理员）
func (h *Handler) GrantUserRole(c *gin.Context) {
	logger := utils.GetLogger()

	var req struct {
		UID  string `json:"uid" binding:"required"`
		Role string `json:"role" binding:"required,oneof=teacher student admin"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	db := client.GetDB()

	// 检查是否已有该角色
	var count int64
	db.Model(&model.ClassroomUserRole{}).Where("uid = ? AND role = ?", req.UID, req.Role).Count(&count)
	if count > 0 {
		c.JSON(http.StatusOK, errorResponse(400, "该用户已拥有此角色"))
		return
	}

	userRole := &model.ClassroomUserRole{
		UID:  req.UID,
		Role: req.Role,
	}

	if err := db.Create(userRole).Error; err != nil {
		logger.Error("赋予角色失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "操作失败"))
		return
	}

	logger.Info("赋予用户角色", zap.String("uid", req.UID), zap.String("role", req.Role))
	c.JSON(http.StatusOK, successResponse(userRole))
}

// RevokeUserRole 移除用户角色（管理员）
func (h *Handler) RevokeUserRole(c *gin.Context) {
	logger := utils.GetLogger()

	var req struct {
		UID  string `json:"uid" binding:"required"`
		Role string `json:"role" binding:"required,oneof=teacher student admin"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	db := client.GetDB()

	result := db.Where("uid = ? AND role = ?", req.UID, req.Role).Delete(&model.ClassroomUserRole{})
	if result.Error != nil {
		logger.Error("移除角色失败", zap.Error(result.Error))
		c.JSON(http.StatusOK, errorResponse(500, "操作失败"))
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusOK, errorResponse(404, "角色不存在"))
		return
	}

	logger.Info("移除用户角色", zap.String("uid", req.UID), zap.String("role", req.Role))
	c.JSON(http.StatusOK, successResponse(nil))
}

// GetUserRoles 获取用户角色列表
func (h *Handler) GetUserRoles(c *gin.Context) {
	logger := utils.GetLogger()
	uid := c.Param("userId")

	logger.Info("GetUserRoles called", zap.String("userId", uid))

	if uid == "" {
		logger.Warn("userId 参数为空")
		c.JSON(http.StatusOK, errorResponse(400, "userId不能为空"))
		return
	}

	db := client.GetDB()
	var roles []model.ClassroomUserRole
	if err := db.Where("uid = ?", uid).Find(&roles).Error; err != nil {
		logger.Error("查询用户角色失败", zap.String("uid", uid), zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	logger.Info("查询用户角色成功", zap.String("uid", uid), zap.Int("count", len(roles)))
	c.JSON(http.StatusOK, successResponse(roles))
}

// GetCurrentUserRoles 获取当前登录用户的角色列表
func (h *Handler) GetCurrentUserRoles(c *gin.Context) {
	logger := utils.GetLogger()

	// 从认证上下文中获取当前用户ID
	currentUserId, exists := c.Get("userId")
	if !exists {
		logger.Warn("获取当前用户角色失败：未认证")
		c.JSON(http.StatusOK, errorResponse(401, "用户未登录"))
		return
	}

	uid := currentUserId.(string)
	logger.Info("GetCurrentUserRoles called", zap.String("uid", uid))

	db := client.GetDB()
	var roles []model.ClassroomUserRole
	if err := db.Where("uid = ?", uid).Find(&roles).Error; err != nil {
		logger.Error("查询当前用户角色失败", zap.String("uid", uid), zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	logger.Info("获取当前用户角色成功", zap.String("uid", uid), zap.Int("count", len(roles)))
	c.JSON(http.StatusOK, successResponse(roles))
}

// UpdateUserRoles 更新用户的班级角色（管理员）
// 操作 hist-oj 自己的 classroom_user_role 表
func (h *Handler) UpdateUserRoles(c *gin.Context) {
	logger := utils.GetLogger()

	// 获取用户ID
	uid := c.Param("uid")
	if uid == "" {
		c.JSON(http.StatusOK, errorResponse(400, "用户ID不能为空"))
		return
	}

	var req struct {
		Roles []string `json:"roles" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	// 验证角色只能是 teacher 或 student
	validRoles := map[string]bool{
		"teacher": true,
		"student": true,
	}
	for _, role := range req.Roles {
		if !validRoles[role] {
			c.JSON(http.StatusOK, errorResponse(400, "无效的角色: "+role))
			return
		}
	}

	db := client.GetDB()

	// 开启事务
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 先删除该用户的所有旧角色
	if err := tx.Where("uid = ?", uid).Delete(&model.ClassroomUserRole{}).Error; err != nil {
		logger.Error("删除旧角色失败", zap.Error(err))
		tx.Rollback()
		c.JSON(http.StatusOK, errorResponse(500, "更新失败"))
		return
	}

	// 插入新角色
	for _, role := range req.Roles {
		userRole := &model.ClassroomUserRole{
			UID:  uid,
			Role: role,
		}
		if err := tx.Create(userRole).Error; err != nil {
			logger.Error("创建角色失败", zap.Error(err))
			tx.Rollback()
			c.JSON(http.StatusOK, errorResponse(500, "更新失败"))
			return
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		logger.Error("提交事务失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "更新失败"))
		return
	}

	logger.Info("更新用户角色成功",
		zap.String("uid", uid),
		zap.Any("roles", req.Roles))

	c.JSON(http.StatusOK, successResponse(gin.H{
		"uid":   uid,
		"roles": req.Roles,
	}))
}


// ==================== 班级管理 ====================

// CreateClassroom 创建班级（教师）
func (h *Handler) CreateClassroom(c *gin.Context) {
	logger := utils.GetLogger()

	var req struct {
		ClassName   string `json:"className" binding:"required"`
		ClassBelong string `json:"classBelong" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	// 获取当前用户ID
	teacherID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "未登录"))
		return
	}

	db := client.GetDB()

	// 生成8位班级码
	classCode := generateClassCode()
	classroom := &model.Classroom{
		ClassName:   req.ClassName,
		ClassBelong: req.ClassBelong,
		ClassCode:   classCode,
		TeacherID:   teacherID.(string),
		Status:      1,
	}

	if err := db.Create(classroom).Error; err != nil {
		logger.Error("创建班级失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "创建失败"))
		return
	}

	logger.Info("创建班级", zap.Uint64("id", classroom.ID), zap.String("name", classroom.ClassName))
	c.JSON(http.StatusOK, successResponse(classroom))
}

// DeleteClassroom 删除班级（教师）
func (h *Handler) DeleteClassroom(c *gin.Context) {
	logger := utils.GetLogger()
	classroomIDStr := c.Param("classroomId")
	classroomID, err := strconv.ParseUint(classroomIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "classroomId参数格式错误"))
		return
	}

	db := client.GetDB()

	// 开始事务
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 删除所有资料文件
	var materials []model.ClassroomMaterial
	if err := tx.Where("folder_id IN (SELECT id FROM classroom_folder WHERE classroom_id = ?)", classroomID).
		Find(&materials).Error; err != nil {
		logger.Error("查询资料文件失败", zap.Error(err))
		tx.Rollback()
		c.JSON(http.StatusOK, errorResponse(500, "删除失败"))
		return
	}

	// 删除资料文件
	for _, material := range materials {
		if material.FilePath != "" {
			// 转换URL路径为文件系统路径
			filePath := "." + material.FilePath
			if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
				logger.Warn("删除资料文件失败", zap.String("path", filePath), zap.Error(err))
			}
		}
	}

	// 2. 删除所有消息图片
	var messages []model.ClassroomMessage
	if err := tx.Where("classroom_id = ? AND msg_type = ?", classroomID, "image").
		Find(&messages).Error; err != nil {
		logger.Error("查询消息图片失败", zap.Error(err))
		tx.Rollback()
		c.JSON(http.StatusOK, errorResponse(500, "删除失败"))
		return
	}

	// 删除消息图片
	for _, message := range messages {
		if message.ImageURL != "" {
			// 转换URL路径为文件系统路径
			imagePath := "." + message.ImageURL
			if err := os.Remove(imagePath); err != nil && !os.IsNotExist(err) {
				logger.Warn("删除消息图片失败", zap.String("path", imagePath), zap.Error(err))
			}
		}
	}

	// 3. 删除数据库记录（软删除班级，级联删除关联数据）
	if err := tx.Model(&model.Classroom{}).Where("id = ?", classroomID).Update("status", 0).Error; err != nil {
		logger.Error("删除班级失败", zap.Error(err))
		tx.Rollback()
		c.JSON(http.StatusOK, errorResponse(500, "删除失败"))
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		logger.Error("提交事务失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "删除失败"))
		return
	}

	logger.Info("删除班级及关联文件", zap.Uint64("id", classroomID))
	c.JSON(http.StatusOK, successResponse(nil))
}

// GetClassroomList 获取班级列表（教师）
func (h *Handler) GetClassroomList(c *gin.Context) {
	logger := utils.GetLogger()

	// 获取当前用户ID
	teacherID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "未登录"))
		return
	}

	db := client.GetDB()
	var classrooms []model.Classroom

	if err := db.Where("teacher_id = ? AND status = 1", teacherID.(string)).
		Preload("Teacher").
		Order("create_time DESC").
		Find(&classrooms).Error; err != nil {
		logger.Error("查询班级列表失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	c.JSON(http.StatusOK, successResponse(classrooms))
}

// GetAllClassrooms 获取所有班级列表（管理员）
func (h *Handler) GetAllClassrooms(c *gin.Context) {
	logger := utils.GetLogger()

	db := client.GetDB()
	var classrooms []model.Classroom

	if err := db.Where("status = 1").
		Preload("Teacher").
		Order("create_time DESC").
		Find(&classrooms).Error; err != nil {
		logger.Error("查询所有班级列表失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	c.JSON(http.StatusOK, successResponse(classrooms))
}

// GetClassroomDetail 获取班级详情
func (h *Handler) GetClassroomDetail(c *gin.Context) {
	logger := utils.GetLogger()
	classroomIDStr := c.Param("classroomId")
	classroomID, err := strconv.ParseUint(classroomIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "classroomId参数格式错误"))
		return
	}

	db := client.GetDB()
	var classroom model.Classroom

	if err := db.Where("id = ? AND status = 1", classroomID).
		Preload("Teacher").
		First(&classroom).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, errorResponse(404, "班级不存在"))
		} else {
			logger.Error("查询班级详情失败", zap.Error(err))
			c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		}
		return
	}

	c.JSON(http.StatusOK, successResponse(classroom))
}

// JoinClassroom 加入班级（学生）
func (h *Handler) JoinClassroom(c *gin.Context) {
	logger := utils.GetLogger()

	var req struct {
		ClassCode   string `json:"classCode" binding:"required"`
		RealName    string `json:"realName" binding:"required"`
		Gender      string `json:"gender"`
		StudentClass string `json:"studentClass"`
		StudentNo   string `json:"studentNo"`
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

	// 查找班级
	var classroom model.Classroom
	if err := db.Where("class_code = ? AND status = 1", req.ClassCode).First(&classroom).Error; err != nil {
		c.JSON(http.StatusOK, errorResponse(404, "班级码不存在或班级已删除"))
		return
	}

	// 检查是否已加入
	var count int64
	db.Model(&model.ClassroomStudent{}).
		Where("classroom_id = ? AND uid = ? AND status = 1", classroom.ID, uid.(string)).
		Count(&count)
	if count > 0 {
		c.JSON(http.StatusOK, errorResponse(400, "已加入该班级"))
		return
	}

	student := &model.ClassroomStudent{
		ClassroomID:  classroom.ID,
		UID:          uid.(string),
		RealName:     req.RealName,
		Gender:       req.Gender,
		StudentClass: req.StudentClass,
		StudentNo:    req.StudentNo,
		Status:       1,
	}

	if err := db.Create(student).Error; err != nil {
		logger.Error("加入班级失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "加入失败"))
		return
	}

	// 重新加载班级信息（包含 Teacher 关联）
	if err := db.Where("id = ? AND status = 1", classroom.ID).
		Preload("Teacher").
		First(&classroom).Error; err != nil {
		logger.Error("加载班级信息失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "加入成功但加载班级信息失败"))
		return
	}

	logger.Info("加入班级", zap.Uint64("classroom_id", classroom.ID), zap.String("uid", uid.(string)))
	c.JSON(http.StatusOK, successResponse(classroom))
}

// GetClassroomStudents 获取班级学生列表（教师）
func (h *Handler) GetClassroomStudents(c *gin.Context) {
	logger := utils.GetLogger()
	classroomIDStr := c.Param("classroomId")
	classroomID, err := strconv.ParseUint(classroomIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "classroomId参数格式错误"))
		return
	}

	db := client.GetDB()
	var students []model.ClassroomStudent

	if err := db.Where("classroom_id = ? AND status = 1", classroomID).
		Preload("User").
		Order("create_time DESC").
		Find(&students).Error; err != nil {
		logger.Error("查询学生列表失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	c.JSON(http.StatusOK, successResponse(students))
}

// RemoveStudent 移除学生（教师）
func (h *Handler) RemoveStudent(c *gin.Context) {
	logger := utils.GetLogger()

	var req struct {
		ClassroomID uint64 `json:"classroomId" binding:"required"`
		UID         string `json:"uid" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	db := client.GetDB()

	result := db.Model(&model.ClassroomStudent{}).
		Where("classroom_id = ? AND uid = ?", req.ClassroomID, req.UID).
		Update("status", 0)

	if result.Error != nil {
		logger.Error("移除学生失败", zap.Error(result.Error))
		c.JSON(http.StatusOK, errorResponse(500, "操作失败"))
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusOK, errorResponse(404, "学生不存在"))
		return
	}

	logger.Info("移除学生", zap.Uint64("classroom_id", req.ClassroomID), zap.String("uid", req.UID))
	c.JSON(http.StatusOK, successResponse(nil))
}

// UpdateStudentInfo 修改学生信息（学生/教师）
func (h *Handler) UpdateStudentInfo(c *gin.Context) {
	logger := utils.GetLogger()

	// 记录收到的请求
	logger.Info("=== UpdateStudentInfo 收到请求 ===",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("content-type", c.Request.Header.Get("Content-Type")),
	)

	// 读取请求体用于调试
	var bodyBytes []byte
	if c.Request.Body != nil {
		bodyBytes, _ = io.ReadAll(c.Request.Body)
		logger.Info("请求体内容", zap.String("body", string(bodyBytes)))
		// 重新设置请求体以便后续读取
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	var req struct {
		ClassroomID  uint64  `json:"classroomId" binding:"required"`
		UID          string  `json:"uid" binding:"required"`
		RealName     *string `json:"realName"`
		Gender       *string `json:"gender"`
		StudentClass *string `json:"studentClass"`
		StudentNo    *string `json:"studentNo"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("JSON绑定失败",
			zap.Error(err),
			zap.String("raw_body", string(bodyBytes)),
		)
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	// 记录请求参数用于调试
	logger.Info("UpdateStudentInfo 请求",
		zap.Uint64("classroomId", req.ClassroomID),
		zap.String("uid", req.UID),
		zap.Any("realName", req.RealName),
		zap.Any("gender", req.Gender),
		zap.Any("studentClass", req.StudentClass),
		zap.Any("studentNo", req.StudentNo),
	)

	db := client.GetDB()

	updates := make(map[string]interface{})
	if req.RealName != nil {
		updates["real_name"] = *req.RealName
	}
	if req.Gender != nil {
		updates["gender"] = *req.Gender
	}
	if req.StudentClass != nil {
		updates["student_class"] = *req.StudentClass
	}
	if req.StudentNo != nil {
		updates["student_no"] = *req.StudentNo
	}

	if len(updates) == 0 {
		c.JSON(http.StatusOK, errorResponse(400, "没有要更新的字段"))
		return
	}

	result := db.Model(&model.ClassroomStudent{}).
		Where("classroom_id = ? AND uid = ?", req.ClassroomID, req.UID).
		Updates(updates)

	if result.Error != nil {
		logger.Error("更新学生信息失败", zap.Error(result.Error))
		c.JSON(http.StatusOK, errorResponse(500, "更新失败"))
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusOK, errorResponse(404, "学生不存在"))
		return
	}

	logger.Info("更新学生信息", zap.Uint64("classroom_id", req.ClassroomID), zap.String("uid", req.UID))
	c.JSON(http.StatusOK, successResponse(nil))
}

// GetStudentClassrooms 获取学生加入的班级列表
func (h *Handler) GetStudentClassrooms(c *gin.Context) {
	logger := utils.GetLogger()

	// 获取当前用户ID
	uid, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "未登录"))
		return
	}

	db := client.GetDB()
	var classroomStudents []model.ClassroomStudent

	if err := db.Where("uid = ? AND status = 1", uid.(string)).
		Order("create_time DESC").
		Find(&classroomStudents).Error; err != nil {
		logger.Error("查询班级列表失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	// 提取班级信息
	classroomIDs := make([]uint64, 0, len(classroomStudents))
	for _, cs := range classroomStudents {
		classroomIDs = append(classroomIDs, cs.ClassroomID)
	}

	var classrooms []model.Classroom
	if len(classroomIDs) > 0 {
		if err := db.Where("id IN ? AND status = 1", classroomIDs).
			Preload("Teacher").
			Find(&classrooms).Error; err != nil {
			logger.Error("查询班级详情失败", zap.Error(err))
			c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
			return
		}
	}

	c.JSON(http.StatusOK, successResponse(classrooms))
}

// ==================== 签到功能 ====================

// CreateCheckin 创建签到（教师）
func (h *Handler) CreateCheckin(c *gin.Context) {
	logger := utils.GetLogger()

	var req struct {
		ClassroomID  uint64  `json:"classroomId" binding:"required"`
		CheckinName  string  `json:"checkinName"`
		StartTime    string  `json:"startTime" binding:"required"` // RFC3339 format
		EndTime      *string `json:"endTime"`                       // RFC3339 format (指针类型以支持null)
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "startTime格式错误"))
		return
	}

	var endTime *time.Time
	if req.EndTime != nil && *req.EndTime != "" {
		t, err := time.Parse(time.RFC3339, *req.EndTime)
		if err != nil {
			c.JSON(http.StatusOK, errorResponse(400, "endTime格式错误"))
			return
		}
		endTime = &t
	}

	db := client.GetDB()

	// 生成签到码
	checkinCode := generateCheckinCode()

	checkin := &model.ClassroomCheckin{
		ClassroomID: req.ClassroomID,
		CheckinCode: checkinCode,
		CheckinName: req.CheckinName,
		StartTime:   startTime,
		EndTime:     endTime,
		Status:      1,
	}

	if err := db.Create(checkin).Error; err != nil {
		logger.Error("创建签到失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "创建失败"))
		return
	}

	logger.Info("创建签到", zap.Uint64("id", checkin.ID), zap.String("code", checkinCode))
	c.JSON(http.StatusOK, successResponse(checkin))
}

// StudentCheckin 学生签到
func (h *Handler) StudentCheckin(c *gin.Context) {
	logger := utils.GetLogger()

	var req struct {
		CheckinID  *uint64 `json:"checkinId"`  // 可选：验证签到码是否属于指定的签到表
		CheckinCode string `json:"checkinCode" binding:"required"`
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

	// 查找签到
	var checkin model.ClassroomCheckin
	if err := db.Where("checkin_code = ?", req.CheckinCode).First(&checkin).Error; err != nil {
		c.JSON(http.StatusOK, errorResponse(404, "签到码不存在"))
		return
	}

	// 如果指定了 checkinId，验证签到码是否属于该签到表
	if req.CheckinID != nil && checkin.ID != *req.CheckinID {
		logger.Warn("签到码与当前签到表不匹配",
			zap.Uint64("expected_checkin_id", *req.CheckinID),
			zap.Uint64("actual_checkin_id", checkin.ID))
		c.JSON(http.StatusOK, errorResponse(400, "签到码不属于当前签到表"))
		return
	}

	// 检查签到状态
	if checkin.Status != 1 {
		c.JSON(http.StatusOK, errorResponse(400, "签到已结束"))
		return
	}

	// 检查是否在时间范围内
	now := time.Now()
	if now.Before(checkin.StartTime) {
		c.JSON(http.StatusOK, errorResponse(400, "签到未开始"))
		return
	}
	if checkin.EndTime != nil && now.After(*checkin.EndTime) {
		c.JSON(http.StatusOK, errorResponse(400, "签到已结束"))
		return
	}

	// 检查是否已签到
	var count int64
	db.Model(&model.ClassroomCheckinRecord{}).
		Where("checkin_id = ? AND uid = ?", checkin.ID, uid.(string)).
		Count(&count)
	if count > 0 {
		c.JSON(http.StatusOK, errorResponse(400, "已签到"))
		return
	}

	now = time.Now()
	record := &model.ClassroomCheckinRecord{
		CheckinID:   checkin.ID,
		UID:         uid.(string),
		Status:      "present",
		CheckinTime: &now,
	}

	if err := db.Create(record).Error; err != nil {
		logger.Error("签到失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "签到失败"))
		return
	}

	logger.Info("学生签到", zap.Uint64("checkin_id", checkin.ID), zap.String("uid", uid.(string)))
	c.JSON(http.StatusOK, successResponse(record))
}

// GetCheckinList 获取班级签到列表（教师）
func (h *Handler) GetCheckinList(c *gin.Context) {
	logger := utils.GetLogger()
	classroomIDStr := c.Param("classroomId")
	classroomID, err := strconv.ParseUint(classroomIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "classroomId参数格式错误"))
		return
	}

	db := client.GetDB()
	var checkins []model.ClassroomCheckin

	if err := db.Where("classroom_id = ?", classroomID).
		Order("create_time DESC").
		Find(&checkins).Error; err != nil {
		logger.Error("查询签到列表失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	c.JSON(http.StatusOK, successResponse(checkins))
}

// GetCheckinRecords 获取签到记录（教师）
func (h *Handler) GetCheckinRecords(c *gin.Context) {
	logger := utils.GetLogger()
	checkinIDStr := c.Param("checkinId")
	checkinID, err := strconv.ParseUint(checkinIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "checkinId参数格式错误"))
		return
	}

	db := client.GetDB()

	// 先查询签到信息，获取 classroomID
	var checkin model.ClassroomCheckin
	if err := db.Where("id = ?", checkinID).First(&checkin).Error; err != nil {
		logger.Error("查询签到信息失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(404, "签到不存在"))
		return
	}

	// 查询签到记录
	var records []model.ClassroomCheckinRecord
	if err := db.Where("checkin_id = ?", checkinID).
		Preload("Student").
		Find(&records).Error; err != nil {
		logger.Error("查询签到记录失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	// 查询班级学生信息（包含真实姓名）
	var classStudents []model.ClassroomStudent
	if err := db.Where("classroom_id = ?", checkin.ClassroomID).Find(&classStudents).Error; err != nil {
		logger.Warn("查询班级学生信息失败", zap.Error(err))
	}

	// 构建学生UID到真实姓名的映射
	studentMap := make(map[string]string)
	for _, cs := range classStudents {
		studentMap[cs.UID] = cs.RealName
	}

	// 为每条签到记录添加班级学生信息
	for i := range records {
		if realName, ok := studentMap[records[i].UID]; ok {
			records[i].ClassStudent = &model.ClassroomStudent{
				RealName: realName,
			}
		}
	}

	c.JSON(http.StatusOK, successResponse(records))
}

// UpdateCheckinRecord 修改签到状态（教师）
func (h *Handler) UpdateCheckinRecord(c *gin.Context) {
	logger := utils.GetLogger()

	var req struct {
		RecordID uint64 `json:"recordId" binding:"required"`
		Status   string `json:"status" binding:"required,oneof=present absent sick_leave personal_leave"`
		Remark   string `json:"remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	db := client.GetDB()

	updates := map[string]interface{}{
		"status": req.Status,
		"remark": req.Remark,
	}

	result := db.Model(&model.ClassroomCheckinRecord{}).
		Where("id = ?", req.RecordID).
		Updates(updates)

	if result.Error != nil {
		logger.Error("更新签到记录失败", zap.Error(result.Error))
		c.JSON(http.StatusOK, errorResponse(500, "更新失败"))
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusOK, errorResponse(404, "记录不存在"))
		return
	}

	logger.Info("更新签到记录", zap.Uint64("record_id", req.RecordID), zap.String("status", req.Status))
	c.JSON(http.StatusOK, successResponse(nil))
}

// EndCheckin 结束签到（教师）
func (h *Handler) EndCheckin(c *gin.Context) {
	logger := utils.GetLogger()
	checkinIDStr := c.Param("checkinId")
	checkinID, err := strconv.ParseUint(checkinIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "checkinId参数格式错误"))
		return
	}

	db := client.GetDB()

	result := db.Model(&model.ClassroomCheckin{}).
		Where("id = ?", checkinID).
		Update("status", 2)

	if result.Error != nil {
		logger.Error("结束签到失败", zap.Error(result.Error))
		c.JSON(http.StatusOK, errorResponse(500, "操作失败"))
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusOK, errorResponse(404, "签到不存在"))
		return
	}

	logger.Info("结束签到", zap.Uint64("checkin_id", checkinID))
	c.JSON(http.StatusOK, successResponse(nil))
}

// ==================== 题库功能 ====================

// CreateQuestion 创建题目（教师）
func (h *Handler) CreateQuestion(c *gin.Context) {
	logger := utils.GetLogger()

	var req struct {
		Title      string `json:"title" binding:"required"`
		Type       string `json:"type" binding:"required,oneof=single_choice multiple_choice judge subjective programming"`
		Content    string `json:"content" binding:"required"`
		Options    string `json:"options"` // JSON string
		Answer     string `json:"answer"`
		Difficulty int    `json:"difficulty"`
		Score      int    `json:"score"`
		IsShared   int    `json:"isShared"`
		ProblemID  uint64 `json:"problemId"` // 编程题的OJ题目ID
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	// 获取当前用户ID
	creatorID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "未登录"))
		return
	}

	// 设置默认值
	if req.Difficulty == 0 {
		req.Difficulty = 1
	}
	if req.Score == 0 {
		if req.Type == "programming" {
			req.Score = 20
		} else {
			req.Score = 2
		}
	}
	if req.IsShared == 0 {
		req.IsShared = 0 // 默认个人题库
	}

	db := client.GetDB()

	// 初始化题目对象
	question := &model.QuestionBank{
		Title:      req.Title,
		Type:       req.Type,
		Content:    req.Content,
		Difficulty: req.Difficulty,
		Score:      req.Score,
		CreatorID:  creatorID.(string),
		IsShared:   req.IsShared,
		Status:     1,
	}

	// 根据题型设置 Options 和 Answer
	if req.Type == "single_choice" || req.Type == "multiple_choice" {
		// 单选和多选题需要选项
		if req.Options != "" {
			question.Options = &req.Options
		}
		question.Answer = req.Answer
	} else if req.Type == "judge" {
		// 判断题不需要选项，设置为 NULL
		question.Options = nil
		question.Answer = req.Answer
	} else if req.Type == "subjective" {
		// 主观题不需要选项
		question.Options = nil
		question.Answer = req.Answer
	} else if req.Type == "programming" {
		// 编程题不需要选项和答案
		question.Options = nil
		question.Answer = ""
	}

	if req.ProblemID > 0 && req.Type == "programming" {
		question.ProblemID = &req.ProblemID
	}

	if err := db.Create(question).Error; err != nil {
		logger.Error("创建题目失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "创建失败"))
		return
	}

	logger.Info("创建题目", zap.Uint64("id", question.ID), zap.String("type", req.Type))
	c.JSON(http.StatusOK, successResponse(question))
}

// GetQuestionBank 获取题库列表（教师）
func (h *Handler) GetQuestionBank(c *gin.Context) {
	logger := utils.GetLogger()

	// 获取当前用户ID
	creatorID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "未登录"))
		return
	}

	// 获取查询参数
	questionType := c.Query("type")
	isSharedStr := c.Query("isShared")
	keyword := c.Query("keyword")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	db := client.GetDB()

	query := db.Model(&model.QuestionBank{}).Where("status = 1")

	// 查询个人题库和共享题库
	query = query.Where("creator_id = ? OR is_shared = 1", creatorID.(string))

	if questionType != "" {
		query = query.Where("type = ?", questionType)
	}

	if isSharedStr != "" {
		isShared, _ := strconv.Atoi(isSharedStr)
		query = query.Where("is_shared = ?", isShared)
	}

	if keyword != "" {
		query = query.Where("title LIKE ?", "%"+keyword+"%")
	}

	var total int64
	query.Count(&total)

	var questions []model.QuestionBank
	if err := query.Preload("Creator").
		Offset((page - 1) * limit).
		Limit(limit).
		Order("create_time DESC").
		Find(&questions).Error; err != nil {
		logger.Error("查询题库失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	c.JSON(http.StatusOK, successResponse(map[string]interface{}{
		"total":     total,
		"page":      page,
		"limit":     limit,
		"questions": questions,
	}))
}

// UpdateQuestion 更新题目（教师）
func (h *Handler) UpdateQuestion(c *gin.Context) {
	logger := utils.GetLogger()
	questionIDStr := c.Param("questionId")
	questionID, err := strconv.ParseUint(questionIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "questionId参数格式错误"))
		return
	}

	var req struct {
		Title      *string `json:"title"`
		Type       *string `json:"type"`
		Content    *string `json:"content"`
		Options    *string `json:"options"`
		Answer     *string `json:"answer"`
		Difficulty *int    `json:"difficulty"`
		Score      *int    `json:"score"`
		IsShared   *int    `json:"isShared"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	db := client.GetDB()

	updates := make(map[string]interface{})
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Type != nil {
		updates["type"] = *req.Type
	}
	if req.Content != nil {
		updates["content"] = *req.Content
	}
	if req.Options != nil {
		updates["options"] = *req.Options
	}
	if req.Answer != nil {
		updates["answer"] = *req.Answer
	}
	if req.Difficulty != nil {
		updates["difficulty"] = *req.Difficulty
	}
	if req.Score != nil {
		updates["score"] = *req.Score
	}
	if req.IsShared != nil {
		updates["is_shared"] = *req.IsShared
	}

	if len(updates) == 0 {
		c.JSON(http.StatusOK, errorResponse(400, "没有要更新的字段"))
		return
	}

	result := db.Model(&model.QuestionBank{}).
		Where("id = ?", questionID).
		Updates(updates)

	if result.Error != nil {
		logger.Error("更新题目失败", zap.Error(result.Error))
		c.JSON(http.StatusOK, errorResponse(500, "更新失败"))
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusOK, errorResponse(404, "题目不存在"))
		return
	}

	logger.Info("更新题目", zap.Uint64("question_id", questionID))
	c.JSON(http.StatusOK, successResponse(nil))
}

// DeleteQuestion 删除题目（教师）
func (h *Handler) DeleteQuestion(c *gin.Context) {
	logger := utils.GetLogger()
	questionIDStr := c.Param("questionId")
	questionID, err := strconv.ParseUint(questionIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "questionId参数格式错误"))
		return
	}

	db := client.GetDB()

	// 软删除
	result := db.Model(&model.QuestionBank{}).
		Where("id = ?", questionID).
		Update("status", 0)

	if result.Error != nil {
		logger.Error("删除题目失败", zap.Error(result.Error))
		c.JSON(http.StatusOK, errorResponse(500, "删除失败"))
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusOK, errorResponse(404, "题目不存在"))
		return
	}

	logger.Info("删除题目", zap.Uint64("question_id", questionID))
	c.JSON(http.StatusOK, successResponse(nil))
}

// GetQuestionDetail 获取题目详情
func (h *Handler) GetQuestionDetail(c *gin.Context) {
	logger := utils.GetLogger()
	questionIDStr := c.Param("questionId")
	questionID, err := strconv.ParseUint(questionIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "questionId参数格式错误"))
		return
	}

	db := client.GetDB()
	var question model.QuestionBank

	if err := db.Where("id = ? AND status = 1", questionID).
		Preload("Creator").
		First(&question).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, errorResponse(404, "题目不存在"))
		} else {
			logger.Error("查询题目详情失败", zap.Error(err))
			c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		}
		return
	}

	c.JSON(http.StatusOK, successResponse(question))
}

// ==================== 辅助函数 ====================

// generateClassCode 生成8位班级码
func generateClassCode() string {
	uuidStr := uuid.New().String()
	uuidStr = fmt.Sprintf("%s%s", uuidStr, uuidStr)
	uuidStr = uuidStr[:8]
	return uuidStr
}

// generateCheckinCode 生成6位签到码
func generateCheckinCode() string {
	rand.Seed(time.Now().UnixNano())
	code := rand.Intn(900000) + 100000 // 100000-999999
	return fmt.Sprintf("%06d", code)
}
