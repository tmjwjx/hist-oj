package api

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/hoj/hist-oj/internal/client"
	"github.com/hoj/hist-oj/internal/middleware"
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

	// 自动更新相关申请的状态
	// 如果用户手动被授予了角色，将相关的待审批申请标记为已批准
	for _, role := range req.Roles {
		now := time.Now()
		result := db.Model(&model.ClassroomRoleRequest{}).
			Where("uid = ? AND role = ? AND status = 0", uid, role).
			Updates(map[string]interface{}{
				"status":       1,
				"reviewer_uid": "system",
				"review_time":  now,
				"review_note":  "管理员手动添加角色",
			})

		if result.Error != nil {
			logger.Warn("更新申请状态失败", zap.Error(result.Error))
		} else if result.RowsAffected > 0 {
			logger.Info("自动更新申请状态为已批准",
				zap.String("uid", uid),
				zap.String("role", role),
				zap.Int64("affected", result.RowsAffected))
		}
	}

	logger.Info("更新用户角色成功",
		zap.String("uid", uid),
		zap.Any("roles", req.Roles))

	c.JSON(http.StatusOK, successResponse(gin.H{
		"uid":   uid,
		"roles": req.Roles,
	}))
}
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

	// 创建班级教师关联记录，确保数据一致性
	teacherRelation := &model.ClassroomTeacher{
		ClassroomID: classroom.ID,
		TeacherID:   teacherID.(string),
		Status:      1,
	}
	if err := db.Create(teacherRelation).Error; err != nil {
		logger.Error("创建班级教师关联失败", zap.Error(err))
		// 不阻断请求，只记录错误
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

	// 1. 查找该班级的所有文件夹ID（包括根目录 folder_id=0 的情况）
	var folders []model.ClassroomFolder
	if err := tx.Where("classroom_id = ? AND status = 1", classroomID).Find(&folders).Error; err != nil {
		logger.Error("查询班级文件夹失败", zap.Error(err))
		tx.Rollback()
		c.JSON(http.StatusOK, errorResponse(500, "删除失败"))
		return
	}

	// 收集所有文件夹ID，包括 0（根目录）
	allFolderIDs := []uint64{0} // 包含根目录
	for _, folder := range folders {
		allFolderIDs = append(allFolderIDs, folder.ID)
	}

	// 2. 查询所有这些文件夹下的资料文件
	var materials []model.ClassroomMaterial
	if err := tx.Where("folder_id IN ? AND status = 1", allFolderIDs).Find(&materials).Error; err != nil {
		logger.Error("查询资料文件失败", zap.Error(err))
		tx.Rollback()
		c.JSON(http.StatusOK, errorResponse(500, "删除失败"))
		return
	}

	logger.Info("找到班级资料文件", zap.Int("count", len(materials)))

	// 删除资料文件
	for _, material := range materials {
		if material.FilePath != "" {
			// 转换URL路径为文件系统路径
			filePath := "." + material.FilePath
			if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
				logger.Warn("删除资料文件失败", zap.String("path", filePath), zap.Error(err))
			} else if err == nil {
				logger.Info("成功删除资料文件", zap.String("path", filePath))
			}
		}
	}

	// 软删除资料文件数据库记录
	if len(materials) > 0 {
		if err := tx.Model(&model.ClassroomMaterial{}).
			Where("folder_id IN ?", allFolderIDs).
			Update("status", 0).Error; err != nil {
			logger.Error("软删除资料文件记录失败", zap.Error(err))
			tx.Rollback()
			c.JSON(http.StatusOK, errorResponse(500, "删除失败"))
			return
		}
		logger.Info("软删除资料文件记录", zap.Int("count", len(materials)))
	}

	// 3. 删除所有消息图片
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
			} else if err == nil {
				logger.Info("成功删除消息图片", zap.String("path", imagePath))
			}
		}
	}

	// 删除消息图片数据库记录
	if len(messages) > 0 {
		if err := tx.Where("classroom_id = ? AND msg_type = ?", classroomID, "image").
			Delete(&model.ClassroomMessage{}).Error; err != nil {
			logger.Error("删除消息图片记录失败", zap.Error(err))
			tx.Rollback()
			c.JSON(http.StatusOK, errorResponse(500, "删除失败"))
			return
		}
		logger.Info("删除消息图片记录", zap.Int("count", len(messages)))
	}

	// 4. 删除作业中学生上传的图片附件
	// 首先查找该班级的所有作业ID
	var homeworkIDs []uint64
	if err := tx.Model(&model.ClassroomHomework{}).
		Where("classroom_id = ? AND status = 1", classroomID).
		Pluck("id", &homeworkIDs).Error; err != nil {
		logger.Error("查询班级作业失败", zap.Error(err))
		tx.Rollback()
		c.JSON(http.StatusOK, errorResponse(500, "删除失败"))
		return
	}

	if len(homeworkIDs) > 0 {
		// 查询这些作业的所有提交记录
		var submissions []model.HomeworkSubmit
		if err := tx.Where("homework_id IN ? AND attachment != ''", homeworkIDs).
			Find(&submissions).Error; err != nil {
			logger.Error("查询作业提交记录失败", zap.Error(err))
			tx.Rollback()
			c.JSON(http.StatusOK, errorResponse(500, "删除失败"))
			return
		}

		// 删除作业附件图片并清空attachment字段
		attachmentCount := 0
		for _, submission := range submissions {
			if submission.Attachment != "" {
				// Attachment 可能包含多个图片URL,用逗号分隔
				attachments := strings.Split(submission.Attachment, ",")
				for _, attachment := range attachments {
					attachment = strings.TrimSpace(attachment)
					if attachment != "" {
						// 转换URL路径为文件系统路径
						filePath := "." + attachment
						if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
							logger.Warn("删除作业附件失败", zap.String("path", filePath), zap.Error(err))
						} else if err == nil {
							logger.Info("成功删除作业附件", zap.String("path", filePath))
							attachmentCount++
						}
					}
				}

				// 清空attachment字段
				if err := tx.Model(&model.HomeworkSubmit{}).
					Where("id = ?", submission.ID).
					Update("attachment", "").Error; err != nil {
					logger.Error("清空作业附件字段失败", zap.Uint64("submission_id", submission.ID), zap.Error(err))
				}
			}
		}
		logger.Info("删除作业附件", zap.Int("count", attachmentCount))
	}

	// 5. 删除数据库记录（软删除班级，级联删除关联数据）
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

	logger.Info("删除班级及关联文件", zap.Uint64("id", classroomID), zap.Int("files_deleted", len(materials)))
	c.JSON(http.StatusOK, successResponse(nil))
}

// GetClassroomList 获取班级列表（教师）
// 包括：作为主教师创建的班级 + 作为额外教师加入的班级
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

	// 1. 查询用户作为主教师的班级
	var primaryClassrooms []model.Classroom
	if err := db.Where("teacher_id = ? AND status = 1", teacherID.(string)).
		Preload("Teacher").
		Preload("Teachers", "status = ?", 1).
		Preload("Teachers.Teacher").
		Order("create_time DESC").
		Find(&primaryClassrooms).Error; err != nil {
		logger.Error("查询主教师班级列表失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	// 2. 查询用户作为额外教师的班级
	var classroomTeachers []model.ClassroomTeacher
	if err := db.Where("teacher_id = ? AND status = 1", teacherID.(string)).
		Find(&classroomTeachers).Error; err != nil {
		logger.Error("查询额外教师关联失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	// 3. 合并班级列表（使用 map 去重）
	classroomMap := make(map[uint64]model.Classroom)
	for _, c := range primaryClassrooms {
		classroomMap[c.ID] = c
	}

	// 对于额外教师的班级，需要加载完整的班级信息
	for _, ct := range classroomTeachers {
		// 如果该班级不在 map 中，需要查询
		if _, exists := classroomMap[ct.ClassroomID]; !exists {
			var classroom model.Classroom
			if err := db.Where("id = ? AND status = 1", ct.ClassroomID).
				Preload("Teacher").
				Preload("Teachers", "status = ?", 1).
				Preload("Teachers.Teacher").
				First(&classroom).Error; err == nil {
				classroomMap[classroom.ID] = classroom
			}
		}
	}

	// 4. 转换为数组并排序
	classrooms = make([]model.Classroom, 0, len(classroomMap))
	for _, classroom := range classroomMap {
		classrooms = append(classrooms, classroom)
	}

	// 按创建时间排序
	for i := 0; i < len(classrooms); i++ {
		for j := i + 1; j < len(classrooms); j++ {
			if classrooms[i].CreatedAt.Before(classrooms[j].CreatedAt) {
				classrooms[i], classrooms[j] = classrooms[j], classrooms[i]
			}
		}
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
		Preload("Teachers.Teacher").
		Order("create_time DESC").
		Find(&classrooms).Error; err != nil {
		logger.Error("查询所有班级列表失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	// 在应用层过滤掉 status=0 的教师记录
	for i := range classrooms {
		if len(classrooms[i].Teachers) > 0 {
			filteredTeachers := make([]model.ClassroomTeacher, 0, len(classrooms[i].Teachers))
			for _, teacher := range classrooms[i].Teachers {
				if teacher.Status == 1 {
					filteredTeachers = append(filteredTeachers, teacher)
				}
			}
			classrooms[i].Teachers = filteredTeachers
		}
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
		Preload("Teachers.Teacher").
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
		ClassCode    string `json:"classCode" binding:"required"`
		RealName     string `json:"realName" binding:"required"`
		Gender       string `json:"gender"`
		StudentClass string `json:"studentClass"`
		StudentNo    string `json:"studentNo"`
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
		Preload("Teachers", "status = ?", 1).
		Preload("Teachers.Teacher").
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

	// 尝试多种方式获取参数
	var req struct {
		ClassroomID    uint64 `json:"classroomId"`
		UID            string `json:"uid"`
		ClassroomIDStr string `form:"classroomId"`
		UIDStr         string `form:"uid"`
	}

	// 先尝试从查询参数获取
	if c.Query("classroomId") != "" || c.Query("uid") != "" {
		req.ClassroomIDStr = c.Query("classroomId")
		req.UIDStr = c.Query("uid")
		classroomID, err := strconv.ParseUint(req.ClassroomIDStr, 10, 64)
		if err != nil {
			logger.Warn("classroomId 参数格式错误", zap.Error(err))
			c.JSON(http.StatusOK, errorResponse(400, "classroomId 参数格式错误"))
			return
		}
		req.ClassroomID = classroomID
		req.UID = req.UIDStr
	} else {
		// 从 JSON 请求体获取
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Warn("请求参数错误", zap.Error(err))
			c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
			return
		}
	}

	if req.ClassroomID == 0 {
		c.JSON(http.StatusOK, errorResponse(400, "classroomId 不能为空"))
		return
	}

	if req.UID == "" {
		c.JSON(http.StatusOK, errorResponse(400, "uid 不能为空"))
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

// AddClassroomStudent 添加学生到班级（教师）
func (h *Handler) AddClassroomStudent(c *gin.Context) {
	logger := utils.GetLogger()

	// 获取当前用户ID
	currentUID, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "未登录"))
		return
	}

	classroomIDStr := c.Param("classroomId")
	classroomID, err := strconv.ParseUint(classroomIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "classroomId参数格式错误"))
		return
	}

	var req struct {
		UID          string `json:"uid" binding:"required"`
		RealName     string `json:"realName" binding:"required"`
		Gender       string `json:"gender"`
		StudentClass string `json:"studentClass"`
		StudentNo    string `json:"studentNo"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	db := client.GetDB()

	// 验证当前用户是否为班级教师
	var classroom model.Classroom
	if err := db.Where("id = ?", classroomID).First(&classroom).Error; err != nil {
		c.JSON(http.StatusOK, errorResponse(404, "班级不存在"))
		return
	}

	// 检查是否为教师（主教师或协教教师）
	isTeacher := false
	if classroom.TeacherID == currentUID.(string) {
		isTeacher = true
	} else {
		// 检查是否为协教教师
		var teacherCount int64
		db.Model(&model.ClassroomTeacher{}).
			Where("classroom_id = ? AND teacher_id = ? AND status = 1", classroomID, currentUID.(string)).
			Count(&teacherCount)
		if teacherCount > 0 {
			isTeacher = true
		}
	}

	if !isTeacher {
		c.JSON(http.StatusOK, errorResponse(403, "只有教师可以添加学生"))
		return
	}

	// 检查学生是否已经在班级中（包括已移除的）
	var existingStudent model.ClassroomStudent
	err = db.Where("classroom_id = ? AND uid = ?", classroomID, req.UID).First(&existingStudent).Error
	if err == nil {
		// 学生记录存在
		if existingStudent.Status == 1 {
			c.JSON(http.StatusOK, errorResponse(400, "学生已在班级中"))
			return
		} else {
			// 学生已被移除，重新激活
			updates := map[string]interface{}{
				"status":        1,
				"real_name":     req.RealName,
				"gender":        req.Gender,
				"student_class": req.StudentClass,
				"student_no":    req.StudentNo,
			}
			if err := db.Model(&existingStudent).Updates(updates).Error; err != nil {
				logger.Error("重新激活学生失败", zap.Error(err))
				c.JSON(http.StatusOK, errorResponse(500, "操作失败"))
				return
			}
			logger.Info("重新激活学生", zap.Uint64("classroom_id", classroomID), zap.String("uid", req.UID))
			c.JSON(http.StatusOK, successResponse(nil))
			return
		}
	} else if err != gorm.ErrRecordNotFound {
		logger.Error("查询学生记录失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	// 创建新的班级学生记录
	student := &model.ClassroomStudent{
		ClassroomID:  classroomID,
		UID:          req.UID,
		RealName:     req.RealName,
		Gender:       req.Gender,
		StudentClass: req.StudentClass,
		StudentNo:    req.StudentNo,
		Status:       1, // 正常状态
	}

	if err := db.Create(student).Error; err != nil {
		logger.Error("添加学生失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "操作失败"))
		return
	}

	logger.Info("添加学生到班级", zap.Uint64("classroom_id", classroomID), zap.String("uid", req.UID))
	c.JSON(http.StatusOK, successResponse(nil))
}

// SearchStudentsToAdd 搜索可添加到班级的学生（教师）
func (h *Handler) SearchStudentsToAdd(c *gin.Context) {
	logger := utils.GetLogger()

	// 获取当前用户ID
	currentUID, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "未登录"))
		return
	}

	classroomIDStr := c.Param("classroomId")
	classroomID, err := strconv.ParseUint(classroomIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "classroomId参数格式错误"))
		return
	}

	// 获取搜索关键词
	keyword := c.Query("keyword")
	if keyword == "" {
		c.JSON(http.StatusOK, errorResponse(400, "keyword参数不能为空"))
		return
	}

	db := client.GetDB()

	// 验证当前用户是否为班级教师
	var classroom model.Classroom
	if err := db.Where("id = ?", classroomID).First(&classroom).Error; err != nil {
		c.JSON(http.StatusOK, errorResponse(404, "班级不存在"))
		return
	}

	// 检查是否为教师
	isTeacher := false
	if classroom.TeacherID == currentUID.(string) {
		isTeacher = true
	} else {
		var teacherCount int64
		db.Model(&model.ClassroomTeacher{}).
			Where("classroom_id = ? AND teacher_id = ? AND status = 1", classroomID, currentUID.(string)).
			Count(&teacherCount)
		if teacherCount > 0 {
			isTeacher = true
		}
	}

	if !isTeacher {
		c.JSON(http.StatusOK, errorResponse(403, "只有教师可以搜索学生"))
		return
	}

	// 搜索用户（排除已在班级中的学生）
	type UserInfoDB struct {
		UUID     string `gorm:"column:uuid"`
		Username string `gorm:"column:username"`
		Nickname string `gorm:"column:nickname"`
		Realname string `gorm:"column:realname"`
		Email    string `gorm:"column:email"`
		Status   int    `gorm:"column:status"`
	}

	var usersDB []UserInfoDB

	// 构建搜索条件：优先精确匹配，然后模糊匹配
	// 使用原生 SQL 实现相关性排序
	searchOrder := fmt.Sprintf(
		"CASE "+
			"WHEN username = '%s' THEN 1 "+
			"WHEN realname = '%s' THEN 2 "+
			"WHEN username LIKE '%s%%' THEN 3 "+
			"WHEN realname LIKE '%s%%' THEN 4 "+
			"ELSE 5 END, username ASC",
		keyword, keyword, keyword, keyword,
	)

	query := db.Table("user_info").
		Select("uuid, username, nickname, realname, email, status").
		Where("(username LIKE ? OR realname LIKE ? OR nickname LIKE ?) AND status = 0",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%").
		Where("uuid NOT IN (SELECT uid FROM classroom_student WHERE classroom_id = ? AND status = 1)", classroomID).
		Order(searchOrder).
		Limit(20)
	if err := query.Find(&usersDB).Error; err != nil {
		logger.Error("查询用户列表失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	// 构建返回结果
	type UserRecord struct {
		UID      string `json:"uid"`
		Username string `json:"username"`
		Nickname string `json:"nickname"`
		Realname string `json:"realname"`
		Email    string `json:"email"`
	}

	records := make([]UserRecord, len(usersDB))
	for i, u := range usersDB {
		records[i] = UserRecord{
			UID:      u.UUID,
			Username: u.Username,
			Nickname: u.Nickname,
			Realname: u.Realname,
			Email:    u.Email,
		}
	}

	c.JSON(http.StatusOK, successResponse(records))
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
			Preload("Teachers", "status = ?", 1).
			Preload("Teachers.Teacher").
			Find(&classrooms).Error; err != nil {
			logger.Error("查询班级详情失败", zap.Error(err))
			c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
			return
		}
	}

	c.JSON(http.StatusOK, successResponse(classrooms))
}

// GetTeacherClassrooms 获取教师创建的班级列表
// 包括：作为主教师创建的班级 + 作为额外教师加入的班级
func (h *Handler) GetTeacherClassrooms(c *gin.Context) {
	logger := utils.GetLogger()

	// 获取当前用户ID
	uid, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "未登录"))
		return
	}

	db := client.GetDB()
	var classrooms []model.Classroom

	// 1. 查询用户作为主教师的班级
	var primaryClassrooms []model.Classroom
	if err := db.Where("teacher_id = ? AND status = 1", uid.(string)).
		Preload("Teacher").
		Preload("Teachers", "status = ?", 1).
		Preload("Teachers.Teacher").
		Order("create_time DESC").
		Find(&primaryClassrooms).Error; err != nil {
		logger.Error("查询主教师班级列表失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	// 2. 查询用户作为额外教师的班级
	var classroomTeachers []model.ClassroomTeacher
	if err := db.Where("teacher_id = ? AND status = 1", uid.(string)).
		Find(&classroomTeachers).Error; err != nil {
		logger.Error("查询额外教师关联失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	// 3. 合并班级列表（使用 map 去重）
	classroomMap := make(map[uint64]model.Classroom)
	for _, c := range primaryClassrooms {
		classroomMap[c.ID] = c
	}

	// 对于额外教师的班级，需要加载完整的班级信息
	for _, ct := range classroomTeachers {
		// 如果该班级不在 map 中，需要查询
		if _, exists := classroomMap[ct.ClassroomID]; !exists {
			var classroom model.Classroom
			if err := db.Where("id = ? AND status = 1", ct.ClassroomID).
				Preload("Teacher").
				Preload("Teachers", "status = ?", 1).
				Preload("Teachers.Teacher").
				First(&classroom).Error; err == nil {
				classroomMap[classroom.ID] = classroom
			}
		}
	}

	// 4. 转换为数组并排序
	classrooms = make([]model.Classroom, 0, len(classroomMap))
	for _, classroom := range classroomMap {
		classrooms = append(classrooms, classroom)
	}

	// 按创建时间排序
	for i := 0; i < len(classrooms); i++ {
		for j := i + 1; j < len(classrooms); j++ {
			if classrooms[i].CreatedAt.Before(classrooms[j].CreatedAt) {
				classrooms[i], classrooms[j] = classrooms[j], classrooms[i]
			}
		}
	}

	c.JSON(http.StatusOK, successResponse(classrooms))
}

// ==================== 签到功能 ====================

// CreateCheckin 创建签到（教师）
func (h *Handler) CreateCheckin(c *gin.Context) {
	logger := utils.GetLogger()

	var req struct {
		ClassroomID           uint64  `json:"classroomId" binding:"required"`
		CheckinName           string  `json:"checkinName"`
		CheckinType           string  `json:"checkinType"`                  // code 或 qrcode
		QrcodeRefreshInterval *int    `json:"qrcodeRefreshInterval"`        // 二维码刷新间隔(秒)
		StartTime             string  `json:"startTime" binding:"required"` // RFC3339 format
		EndTime               *string `json:"endTime"`                      // RFC3339 format (指针类型以支持null)
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

	// 设置默认值
	checkinType := req.CheckinType
	if checkinType == "" {
		checkinType = "code"
	}

	qrcodeRefreshInterval := 15 // 默认15秒
	if req.QrcodeRefreshInterval != nil {
		qrcodeRefreshInterval = *req.QrcodeRefreshInterval
	}

	checkin := &model.ClassroomCheckin{
		ClassroomID:           req.ClassroomID,
		CheckinCode:           checkinCode,
		CheckinName:           req.CheckinName,
		CheckinType:           checkinType,
		QrcodeRefreshInterval: qrcodeRefreshInterval,
		StartTime:             startTime,
		EndTime:               endTime,
		Status:                1,
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
		CheckinID   *uint64 `json:"checkinId"` // 可选：验证签到码是否属于指定的签到表
		CheckinCode string  `json:"checkinCode" binding:"required"`
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

// GetCheckinListForStudent 获取班级签到列表（学生）- 只返回学生需要看到的信息
func (h *Handler) GetCheckinListForStudent(c *gin.Context) {
	logger := utils.GetLogger()
	classroomIDStr := c.Param("classroomId")
	classroomID, err := strconv.ParseUint(classroomIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "classroomId参数格式错误"))
		return
	}

	// 获取当前用户ID
	uid, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "未登录"))
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

	// 创建学生视图的签到列表（移除敏感信息）+ 包含当前用户的签到状态
	type StudentCheckinView struct {
		ID                    uint64     `json:"id"`
		ClassroomID           uint64     `json:"classroomId"`
		CheckinName           string     `json:"checkinName"`
		CheckinType           string     `json:"checkinType"`
		QrcodeRefreshInterval int        `json:"qrcodeRefreshInterval"`
		StartTime             time.Time  `json:"startTime"`
		EndTime               *time.Time `json:"endTime"`
		Status                int        `json:"status"`
		CreatedAt             time.Time  `json:"createdAt"`
		UserCheckinStatus     *string    `json:"userCheckinStatus"` // 当前用户的签到状态
	}

	studentCheckins := make([]StudentCheckinView, 0, len(checkins))

	// 批量查询所有签到的记录ID列表
	checkinIDs := make([]uint64, len(checkins))
	for i, checkin := range checkins {
		checkinIDs[i] = checkin.ID
	}

	// 查询当前用户在所有签到中的记录
	var userRecords []model.ClassroomCheckinRecord
	db.Where("checkin_id IN ? AND uid = ?", checkinIDs, uid.(string)).Find(&userRecords)

	// 创建签到ID到用户状态的映射
	userStatusMap := make(map[uint64]string)
	for _, record := range userRecords {
		userStatusMap[record.CheckinID] = record.Status
	}

	for _, checkin := range checkins {
		view := StudentCheckinView{
			ID:                    checkin.ID,
			ClassroomID:           checkin.ClassroomID,
			CheckinName:           checkin.CheckinName,
			CheckinType:           checkin.CheckinType,
			QrcodeRefreshInterval: checkin.QrcodeRefreshInterval,
			StartTime:             checkin.StartTime,
			EndTime:               checkin.EndTime,
			Status:                checkin.Status,
			CreatedAt:             checkin.CreatedAt,
		}

		// 添加当前用户的签到状态
		if status, ok := userStatusMap[checkin.ID]; ok {
			view.UserCheckinStatus = &status
		}

		studentCheckins = append(studentCheckins, view)
	}

	c.JSON(http.StatusOK, successResponse(studentCheckins))
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

// CreateCheckinRecord 创建签到记录（教师）- 为未签到学生创建记录
func (h *Handler) CreateCheckinRecord(c *gin.Context) {
	logger := utils.GetLogger()
	db := client.GetDB()

	checkinIDStr := c.Param("checkinId")
	checkinID, err := strconv.ParseUint(checkinIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "checkinId参数格式错误"))
		return
	}

	var req struct {
		UID    string `json:"uid" binding:"required"`
		Status string `json:"status" binding:"required,oneof=present absent sick_leave personal_leave"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	// 查询签到信息
	var checkin model.ClassroomCheckin
	if err := db.Where("id = ?", checkinID).First(&checkin).Error; err != nil {
		logger.Error("查询签到信息失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(404, "签到不存在"))
		return
	}

	// 检查是否已经存在记录
	var existingRecord model.ClassroomCheckinRecord
	if err := db.Where("checkin_id = ? AND uid = ?", checkinID, req.UID).First(&existingRecord).Error; err == nil {
		c.JSON(http.StatusOK, errorResponse(400, "该用户已有签到记录"))
		return
	}

	// 创建签到记录
	record := &model.ClassroomCheckinRecord{
		CheckinID:   checkinID,
		UID:         req.UID,
		Status:      req.Status,
		CheckinTime: nil, // 教师手动创建的记录，签到时间设为空
	}

	if err := db.Create(record).Error; err != nil {
		logger.Error("创建签到记录失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "创建失败"))
		return
	}

	logger.Info("创建签到记录", zap.Uint64("checkin_id", checkinID), zap.String("uid", req.UID), zap.String("status", req.Status))
	c.JSON(http.StatusOK, successResponse(record))
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

// UpdateCheckin 更新签到信息（教师）
func (h *Handler) UpdateCheckin(c *gin.Context) {
	logger := utils.GetLogger()
	checkinIDStr := c.Param("checkinId")
	checkinID, err := strconv.ParseUint(checkinIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "checkinId参数格式错误"))
		return
	}

	var req struct {
		CheckinID   uint64  `json:"checkinId"`
		CheckinName string  `json:"checkinName"`
		StartTime   string  `json:"startTime"`
		EndTime     *string `json:"endTime"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	db := client.GetDB()

	// 查询签到是否存在
	var checkin model.ClassroomCheckin
	if err := db.Where("id = ?", checkinID).First(&checkin).Error; err != nil {
		logger.Error("查询签到失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(404, "签到不存在"))
		return
	}

	// 更新签到信息
	updates := map[string]interface{}{
		"checkin_name": req.CheckinName,
		"start_time":   req.StartTime,
	}

	if req.EndTime != nil {
		updates["end_time"] = *req.EndTime
	} else {
		updates["end_time"] = nil
	}

	// 根据当前时间和新的时间范围自动更新 status
	now := time.Now()
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err == nil {
		// 检查是否需要更新状态
		var newStatus int
		if req.EndTime != nil && *req.EndTime != "" {
			endTime, err := time.Parse(time.RFC3339, *req.EndTime)
			if err == nil && now.After(endTime) {
				newStatus = 0 // 已结束
			} else if (now.Equal(startTime) || now.After(startTime)) && now.Before(endTime) {
				newStatus = 1 // 进行中
			} else {
				newStatus = 0 // 未开始
			}
		} else if now.Equal(startTime) || now.After(startTime) {
			newStatus = 1 // 进行中（没有结束时间）
		} else {
			newStatus = 0 // 未开始
		}

		updates["status"] = newStatus
	}

	if err := db.Model(&checkin).Updates(updates).Error; err != nil {
		logger.Error("更新签到失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "更新失败"))
		return
	}

	logger.Info("更新签到", zap.Uint64("checkin_id", checkinID))
	c.JSON(http.StatusOK, successResponse(nil))
}

// DeleteCheckin 删除签到（教师）
func (h *Handler) DeleteCheckin(c *gin.Context) {
	logger := utils.GetLogger()
	checkinIDStr := c.Param("checkinId")
	checkinID, err := strconv.ParseUint(checkinIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "checkinId参数格式错误"))
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

	// 删除签到记录
	if err := tx.Where("checkin_id = ?", checkinID).Delete(&model.ClassroomCheckinRecord{}).Error; err != nil {
		logger.Error("删除签到记录失败", zap.Error(err))
		tx.Rollback()
		c.JSON(http.StatusOK, errorResponse(500, "删除失败"))
		return
	}

	// 删除签到
	if err := tx.Where("id = ?", checkinID).Delete(&model.ClassroomCheckin{}).Error; err != nil {
		logger.Error("删除签到失败", zap.Error(err))
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

	logger.Info("删除签到", zap.Uint64("checkin_id", checkinID))
	c.JSON(http.StatusOK, successResponse(nil))
}

// ==================== 题库功能 ====================

// CreateQuestion 创建题目（教师）
func (h *Handler) CreateQuestion(c *gin.Context) {
	logger := utils.GetLogger()

	var req struct {
		Title      string `json:"title" binding:"required"`
		Type       string `json:"type" binding:"required,oneof=single_choice multiple_choice judge subjective programming composite"`
		Content    string `json:"content" binding:"required"`
		Options    string `json:"options"` // JSON string
		Answer     string `json:"answer"`
		Analysis   string `json:"analysis"` // 题目解析
		Tags       string `json:"tags"`     // 题目标签（JSON数组）
		Course     string `json:"course"`   // 题目所属课程
		Difficulty int    `json:"difficulty"`
		Score      int    `json:"score"`
		IsShared   int    `json:"isShared"`
		ProblemID  string `json:"problemId"` // 编程题的OJ题目ID(字符串类型,支持"0001"等格式)
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

	var optionsInput *string
	if strings.TrimSpace(req.Options) != "" {
		optionsValue := req.Options
		optionsInput = &optionsValue
	}
	normalizedAnswer, normalizedOptions, normalizeErr := normalizeQuestionAnswerForStorage(req.Type, req.Answer, optionsInput)
	if normalizeErr != nil {
		c.JSON(http.StatusOK, errorResponse(400, normalizeErr.Error()))
		return
	}

	if req.Type == "composite" {
		subQuestions, _, err := parseCompositeSubQuestionsFromStored(normalizedOptions)
		if err != nil {
			c.JSON(http.StatusOK, errorResponse(400, err.Error()))
			return
		}
		totalScore := 0
		for _, subQuestion := range subQuestions {
			totalScore += subQuestion.Score
		}
		req.Score = totalScore
	}

	// 初始化题目对象
	question := &model.QuestionBank{
		Title:      req.Title,
		Type:       req.Type,
		Content:    req.Content,
		Options:    normalizedOptions,
		Answer:     normalizedAnswer,
		Analysis:   req.Analysis, // 题目解析
		Tags:       req.Tags,     // 题目标签
		Course:     req.Course,   // 题目所属课程
		Difficulty: req.Difficulty,
		Score:      req.Score,
		CreatorID:  creatorID.(string),
		IsShared:   req.IsShared,
		Status:     1,
	}

	if req.ProblemID != "" && req.Type == "programming" {
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
	course := c.Query("course")         // 课程筛选
	tag := c.Query("tag")               // 标签筛选
	difficulty := c.Query("difficulty") // 难度筛选
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

	// 课程筛选
	if course != "" {
		query = query.Where("course = ?", course)
	}

	// 标签筛选（使用JSON_CONTAINS）
	if tag != "" {
		// 转义标签中的特殊字符，确保 JSON 格式正确
		escapedTag := strings.ReplaceAll(tag, `\`, `\\`)
		escapedTag = strings.ReplaceAll(escapedTag, `"`, `\"`)
		query = query.Where("JSON_CONTAINS(tags, ?)", fmt.Sprintf(`"%s"`, escapedTag))
	}

	// 难度筛选
	if difficulty != "" {
		difficultyInt, err := strconv.Atoi(difficulty)
		if err == nil {
			query = query.Where("difficulty = ?", difficultyInt)
		}
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
		Analysis   *string `json:"analysis"` // 题目解析
		Tags       *string `json:"tags"`     // 题目标签
		Course     *string `json:"course"`   // 题目所属课程
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

	var question model.QuestionBank
	if err := db.Where("id = ? AND status = 1", questionID).First(&question).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, errorResponse(404, "题目不存在"))
		} else {
			logger.Error("查询题目失败", zap.Error(err))
			c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		}
		return
	}
	originalImageURLs := extractQuestionImageURLsFromFields(question.Title, question.Content, question.Analysis, question.Options)

	updates := make(map[string]interface{})
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Type != nil {
		validTypes := map[string]bool{
			"single_choice":   true,
			"multiple_choice": true,
			"judge":           true,
			"subjective":      true,
			"programming":     true,
			"composite":       true,
		}
		if !validTypes[*req.Type] {
			c.JSON(http.StatusOK, errorResponse(400, "题型不合法"))
			return
		}
		updates["type"] = *req.Type
	}
	if req.Content != nil {
		updates["content"] = *req.Content
	}
	if req.Analysis != nil {
		updates["analysis"] = *req.Analysis
	}
	if req.Tags != nil {
		updates["tags"] = *req.Tags
	}
	if req.Course != nil {
		updates["course"] = *req.Course
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

	needNormalizeAnswer := req.Type != nil || req.Options != nil || req.Answer != nil
	if needNormalizeAnswer {
		targetType := question.Type
		if req.Type != nil {
			targetType = *req.Type
		}

		targetAnswer := question.Answer
		if req.Answer != nil {
			targetAnswer = *req.Answer
		}

		var targetOptions *string
		if req.Options != nil {
			if strings.TrimSpace(*req.Options) != "" {
				optionsValue := *req.Options
				targetOptions = &optionsValue
			} else {
				targetOptions = nil
			}
		} else {
			targetOptions = question.Options
		}

		normalizedAnswer, normalizedOptions, err := normalizeQuestionAnswerForStorage(targetType, targetAnswer, targetOptions)
		if err != nil {
			c.JSON(http.StatusOK, errorResponse(400, err.Error()))
			return
		}

		updates["answer"] = normalizedAnswer
		if normalizedOptions == nil {
			updates["options"] = nil
		} else {
			updates["options"] = *normalizedOptions
		}

		if targetType == "composite" {
			subQuestions, _, err := parseCompositeSubQuestionsFromStored(normalizedOptions)
			if err != nil {
				c.JSON(http.StatusOK, errorResponse(400, err.Error()))
				return
			}
			totalScore := 0
			for _, subQuestion := range subQuestions {
				totalScore += subQuestion.Score
			}
			updates["score"] = totalScore
		}
	}

	// 组合题总分固定等于子题分值总和，不允许手动改分
	if question.Type == "composite" && req.Type == nil && req.Options == nil && req.Answer == nil && req.Score != nil {
		subQuestions, _, err := parseCompositeSubQuestionsFromStored(question.Options)
		if err != nil {
			c.JSON(http.StatusOK, errorResponse(400, err.Error()))
			return
		}
		totalScore := 0
		for _, subQuestion := range subQuestions {
			totalScore += subQuestion.Score
		}
		updates["score"] = totalScore
	}

	if len(updates) == 0 {
		c.JSON(http.StatusOK, errorResponse(400, "没有要更新的字段"))
		return
	}

	nextTitle := question.Title
	if value, ok := updates["title"].(string); ok {
		nextTitle = value
	}
	nextContent := question.Content
	if value, ok := updates["content"].(string); ok {
		nextContent = value
	}
	nextAnalysis := question.Analysis
	if value, ok := updates["analysis"].(string); ok {
		nextAnalysis = value
	}
	nextOptions := question.Options
	if value, exists := updates["options"]; exists {
		if value == nil {
			nextOptions = nil
		} else if optionText, ok := value.(string); ok {
			nextOptions = &optionText
		}
	}
	updatedImageURLs := extractQuestionImageURLsFromFields(nextTitle, nextContent, nextAnalysis, nextOptions)
	removedImageURLs := diffRemovedQuestionImageURLs(originalImageURLs, updatedImageURLs)

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

	if len(removedImageURLs) > 0 {
		deleted, skipped := deleteQuestionImagesByURLs(logger, removedImageURLs)
		logger.Info("更新题目后回收图片",
			zap.Uint64("question_id", questionID),
			zap.Int("removed", len(removedImageURLs)),
			zap.Int("deleted", deleted),
			zap.Int("skipped", skipped))
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

	var question model.QuestionBank
	if err := db.Where("id = ? AND status = 1", questionID).First(&question).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, errorResponse(404, "题目不存在"))
		} else {
			logger.Error("查询题目失败", zap.Error(err))
			c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		}
		return
	}
	imageURLs := extractQuestionImageURLsFromFields(question.Title, question.Content, question.Analysis, question.Options)

	if err := db.Model(&question).Update("status", 0).Error; err != nil {
		logger.Error("删除题目失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "删除失败"))
		return
	}

	if len(imageURLs) > 0 {
		deleted, skipped := deleteQuestionImagesByURLs(logger, imageURLs)
		logger.Info("删除题目后回收图片",
			zap.Uint64("question_id", questionID),
			zap.Int("images", len(imageURLs)),
			zap.Int("deleted", deleted),
			zap.Int("skipped", skipped))
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

// ==================== 管理员题库管理功能 ====================

// AdminGetQuestionBank 管理员获取所有题库题目（包括公开和私有的）
func (h *Handler) AdminGetQuestionBank(c *gin.Context) {
	logger := utils.GetLogger()

	// 获取查询参数
	questionType := c.Query("type")
	isSharedStr := c.Query("isShared")
	searchField := c.Query("searchField") // 搜索字段：title, id, creator
	keyword := c.Query("keyword")
	course := c.Query("course")         // 课程筛选
	tag := c.Query("tag")               // 标签筛选
	difficulty := c.Query("difficulty") // 难度筛选
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	db := client.GetDB()

	// 管理员可以看到所有题目（无论是否共享）
	query := db.Model(&model.QuestionBank{}).Where("status = 1")

	if questionType != "" {
		query = query.Where("type = ?", questionType)
	}

	if isSharedStr != "" {
		isShared, _ := strconv.Atoi(isSharedStr)
		query = query.Where("is_shared = ?", isShared)
	}

	// 课程筛选
	if course != "" {
		query = query.Where("course = ?", course)
	}

	// 标签筛选（使用JSON_CONTAINS）
	if tag != "" {
		// 转义标签中的特殊字符，确保 JSON 格式正确
		escapedTag := strings.ReplaceAll(tag, `\`, `\\`)
		escapedTag = strings.ReplaceAll(escapedTag, `"`, `\"`)
		query = query.Where("JSON_CONTAINS(tags, ?)", fmt.Sprintf(`"%s"`, escapedTag))
	}

	// 难度筛选
	if difficulty != "" {
		difficultyInt, err := strconv.Atoi(difficulty)
		if err == nil {
			query = query.Where("difficulty = ?", difficultyInt)
		}
	}

	// 根据选择的字段进行搜索
	if keyword != "" {
		switch searchField {
		case "id":
			// 搜索题目ID
			if questionID, err := strconv.ParseUint(keyword, 10, 64); err == nil {
				query = query.Where("id = ?", questionID)
			}
		case "creator":
			// 搜索创建者用户名
			query = query.Where("creator_id IN (SELECT uuid FROM user_info WHERE username LIKE ?)", "%"+keyword+"%")
		case "title":
			// 搜索题目标题（默认）
			fallthrough
		default:
			query = query.Where("title LIKE ?", "%"+keyword+"%")
		}
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

// AdminUpdateQuestion 管理员更新题目
func (h *Handler) AdminUpdateQuestion(c *gin.Context) {
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
		Analysis   *string `json:"analysis"` // 题目解析
		Tags       *string `json:"tags"`     // 题目标签
		Course     *string `json:"course"`   // 题目所属课程
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

	// 检查题目是否存在
	var question model.QuestionBank
	if err := db.Where("id = ? AND status = 1", questionID).First(&question).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, errorResponse(404, "题目不存在"))
		} else {
			logger.Error("查询题目失败", zap.Error(err))
			c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		}
		return
	}
	originalImageURLs := extractQuestionImageURLsFromFields(question.Title, question.Content, question.Analysis, question.Options)

	updates := make(map[string]interface{})
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Type != nil {
		validTypes := map[string]bool{
			"single_choice":   true,
			"multiple_choice": true,
			"judge":           true,
			"subjective":      true,
			"programming":     true,
			"composite":       true,
		}
		if !validTypes[*req.Type] {
			c.JSON(http.StatusOK, errorResponse(400, "题型不合法"))
			return
		}
		updates["type"] = *req.Type
	}
	if req.Content != nil {
		updates["content"] = *req.Content
	}
	if req.Analysis != nil {
		updates["analysis"] = *req.Analysis
	}
	if req.Tags != nil {
		updates["tags"] = *req.Tags
	}
	if req.Course != nil {
		updates["course"] = *req.Course
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

	needNormalizeAnswer := req.Type != nil || req.Options != nil || req.Answer != nil
	if needNormalizeAnswer {
		targetType := question.Type
		if req.Type != nil {
			targetType = *req.Type
		}

		targetAnswer := question.Answer
		if req.Answer != nil {
			targetAnswer = *req.Answer
		}

		var targetOptions *string
		if req.Options != nil {
			if strings.TrimSpace(*req.Options) != "" {
				optionsValue := *req.Options
				targetOptions = &optionsValue
			} else {
				targetOptions = nil
			}
		} else {
			targetOptions = question.Options
		}

		normalizedAnswer, normalizedOptions, err := normalizeQuestionAnswerForStorage(targetType, targetAnswer, targetOptions)
		if err != nil {
			c.JSON(http.StatusOK, errorResponse(400, err.Error()))
			return
		}

		updates["answer"] = normalizedAnswer
		if normalizedOptions == nil {
			updates["options"] = nil
		} else {
			updates["options"] = *normalizedOptions
		}

		if targetType == "composite" {
			subQuestions, _, err := parseCompositeSubQuestionsFromStored(normalizedOptions)
			if err != nil {
				c.JSON(http.StatusOK, errorResponse(400, err.Error()))
				return
			}
			totalScore := 0
			for _, subQuestion := range subQuestions {
				totalScore += subQuestion.Score
			}
			updates["score"] = totalScore
		}
	}

	// 组合题总分固定等于子题分值总和，不允许手动改分
	if question.Type == "composite" && req.Type == nil && req.Options == nil && req.Answer == nil && req.Score != nil {
		subQuestions, _, err := parseCompositeSubQuestionsFromStored(question.Options)
		if err != nil {
			c.JSON(http.StatusOK, errorResponse(400, err.Error()))
			return
		}
		totalScore := 0
		for _, subQuestion := range subQuestions {
			totalScore += subQuestion.Score
		}
		updates["score"] = totalScore
	}

	if len(updates) == 0 {
		c.JSON(http.StatusOK, errorResponse(400, "没有要更新的字段"))
		return
	}

	nextTitle := question.Title
	if value, ok := updates["title"].(string); ok {
		nextTitle = value
	}
	nextContent := question.Content
	if value, ok := updates["content"].(string); ok {
		nextContent = value
	}
	nextAnalysis := question.Analysis
	if value, ok := updates["analysis"].(string); ok {
		nextAnalysis = value
	}
	nextOptions := question.Options
	if value, exists := updates["options"]; exists {
		if value == nil {
			nextOptions = nil
		} else if optionText, ok := value.(string); ok {
			nextOptions = &optionText
		}
	}
	updatedImageURLs := extractQuestionImageURLsFromFields(nextTitle, nextContent, nextAnalysis, nextOptions)
	removedImageURLs := diffRemovedQuestionImageURLs(originalImageURLs, updatedImageURLs)

	if err := db.Model(&question).Updates(updates).Error; err != nil {
		logger.Error("更新题目失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "更新失败"))
		return
	}

	if len(removedImageURLs) > 0 {
		deleted, skipped := deleteQuestionImagesByURLs(logger, removedImageURLs)
		logger.Info("管理员更新题目后回收图片",
			zap.Uint64("id", questionID),
			zap.Int("removed", len(removedImageURLs)),
			zap.Int("deleted", deleted),
			zap.Int("skipped", skipped))
	}

	logger.Info("管理员更新题目", zap.Uint64("id", questionID))
	c.JSON(http.StatusOK, successResponse(question))
}

// AdminDeleteQuestion 管理员删除题目（软删除）
func (h *Handler) AdminDeleteQuestion(c *gin.Context) {
	logger := utils.GetLogger()
	questionIDStr := c.Param("questionId")
	questionID, err := strconv.ParseUint(questionIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "questionId参数格式错误"))
		return
	}

	db := client.GetDB()

	// 检查题目是否存在
	var question model.QuestionBank
	if err := db.Where("id = ? AND status = 1", questionID).First(&question).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, errorResponse(404, "题目不存在"))
		} else {
			logger.Error("查询题目失败", zap.Error(err))
			c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		}
		return
	}
	imageURLs := extractQuestionImageURLsFromFields(question.Title, question.Content, question.Analysis, question.Options)

	// 软删除：更新status为0
	if err := db.Model(&question).Update("status", 0).Error; err != nil {
		logger.Error("删除题目失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "删除失败"))
		return
	}

	if len(imageURLs) > 0 {
		deleted, skipped := deleteQuestionImagesByURLs(logger, imageURLs)
		logger.Info("管理员删除题目后回收图片",
			zap.Uint64("id", questionID),
			zap.Int("images", len(imageURLs)),
			zap.Int("deleted", deleted),
			zap.Int("skipped", skipped))
	}

	logger.Info("管理员删除题目", zap.Uint64("id", questionID))
	c.JSON(http.StatusOK, successResponse(nil))
}

// AdminCreateQuestion 管理员创建题目
func (h *Handler) AdminCreateQuestion(c *gin.Context) {
	logger := utils.GetLogger()

	// 获取当前用户ID
	creatorID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "未登录"))
		return
	}

	var req struct {
		Title      string `json:"title" binding:"required"`
		Type       string `json:"type" binding:"required,oneof=single_choice multiple_choice judge subjective composite"`
		Content    string `json:"content" binding:"required"`
		Options    string `json:"options"`
		Answer     string `json:"answer"`
		Analysis   string `json:"analysis"` // 题目解析
		Tags       string `json:"tags"`     // 题目标签
		Course     string `json:"course"`   // 题目所属课程
		Difficulty int    `json:"difficulty"`
		Score      int    `json:"score"`
		IsShared   int    `json:"isShared"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	// 设置默认值
	if req.Difficulty == 0 {
		req.Difficulty = 1
	}
	if req.Score == 0 {
		switch req.Type {
		case "single_choice":
			req.Score = 2
		case "multiple_choice":
			req.Score = 5
		case "judge":
			req.Score = 1
		case "subjective":
			req.Score = 5
		default:
			req.Score = 2
		}
	}
	if req.IsShared == 0 {
		req.IsShared = 0
	}

	db := client.GetDB()

	var optionsInput *string
	if strings.TrimSpace(req.Options) != "" {
		optionsValue := req.Options
		optionsInput = &optionsValue
	}
	normalizedAnswer, normalizedOptions, normalizeErr := normalizeQuestionAnswerForStorage(req.Type, req.Answer, optionsInput)
	if normalizeErr != nil {
		c.JSON(http.StatusOK, errorResponse(400, normalizeErr.Error()))
		return
	}

	if req.Type == "composite" {
		subQuestions, _, err := parseCompositeSubQuestionsFromStored(normalizedOptions)
		if err != nil {
			c.JSON(http.StatusOK, errorResponse(400, err.Error()))
			return
		}
		totalScore := 0
		for _, subQuestion := range subQuestions {
			totalScore += subQuestion.Score
		}
		req.Score = totalScore
	}

	question := &model.QuestionBank{
		Title:      req.Title,
		Type:       req.Type,
		Content:    req.Content,
		Options:    normalizedOptions,
		Answer:     normalizedAnswer,
		Analysis:   req.Analysis, // 题目解析
		Tags:       req.Tags,     // 题目标签
		Course:     req.Course,   // 题目所属课程
		Difficulty: req.Difficulty,
		Score:      req.Score,
		CreatorID:  creatorID.(string),
		IsShared:   req.IsShared,
		Status:     1,
	}

	if err := db.Create(question).Error; err != nil {
		logger.Error("创建题目失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "创建失败"))
		return
	}

	logger.Info("管理员创建题目", zap.Uint64("id", question.ID), zap.String("type", req.Type))
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

// ==================== 二维码签到功能 ====================

// GetQrcodeInfo 获取二维码信息（教师）
func (h *Handler) GetQrcodeInfo(c *gin.Context) {
	logger := utils.GetLogger()

	checkinIDStr := c.Param("checkinId")
	checkinID, err := strconv.ParseUint(checkinIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "checkinId参数格式错误"))
		return
	}

	db := client.GetDB()

	var checkin model.ClassroomCheckin
	if err := db.Where("id = ?", checkinID).First(&checkin).Error; err != nil {
		logger.Error("查询签到失败", zap.Error(err), zap.Uint64("checkin_id", checkinID))
		c.JSON(http.StatusOK, errorResponse(404, "签到不存在"))
		return
	}

	// 如果还没有二维码token，生成一个
	if checkin.QrcodeToken == "" || checkin.QrcodeExpiresAt == nil || time.Now().After(*checkin.QrcodeExpiresAt) {
		token := generateQrcodeToken(checkinID, checkin.QrcodeRefreshInterval)
		expiresAt := time.Now().Add(time.Duration(checkin.QrcodeRefreshInterval) * time.Second)

		checkin.QrcodeToken = token
		checkin.QrcodeExpiresAt = &expiresAt

		if err := db.Save(&checkin).Error; err != nil {
			logger.Error("更新二维码token失败", zap.Error(err))
			c.JSON(http.StatusOK, errorResponse(500, "生成二维码失败"))
			return
		}
	}

	// 计算剩余秒数
	var refreshIn int64
	if checkin.QrcodeExpiresAt != nil {
		refreshIn = int64(checkin.QrcodeExpiresAt.Sub(time.Now()).Seconds())
		if refreshIn < 0 {
			refreshIn = 0
		}
	}

	// 生成二维码URL
	qrcodeURL := fmt.Sprintf("/api/classroom/checkin/%d/qrcode/scan?token=%s", checkinID, checkin.QrcodeToken)

	response := map[string]interface{}{
		"qrcodeToken": checkin.QrcodeToken,
		"qrcodeUrl":   qrcodeURL,
		"expiresAt":   checkin.QrcodeExpiresAt,
		"refreshIn":   refreshIn,
	}

	c.JSON(http.StatusOK, successResponse(response))
}

// RefreshQrcode 刷新二维码（教师）
func (h *Handler) RefreshQrcode(c *gin.Context) {
	logger := utils.GetLogger()

	checkinIDStr := c.Param("checkinId")
	checkinID, err := strconv.ParseUint(checkinIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "checkinId参数格式错误"))
		return
	}

	db := client.GetDB()

	var checkin model.ClassroomCheckin
	if err := db.Where("id = ?", checkinID).First(&checkin).Error; err != nil {
		logger.Error("查询签到失败", zap.Error(err), zap.Uint64("checkin_id", checkinID))
		c.JSON(http.StatusOK, errorResponse(404, "签到不存在"))
		return
	}

	// 生成新的token
	token := generateQrcodeToken(checkinID, checkin.QrcodeRefreshInterval)
	expiresAt := time.Now().Add(time.Duration(checkin.QrcodeRefreshInterval) * time.Second)

	checkin.QrcodeToken = token
	checkin.QrcodeExpiresAt = &expiresAt

	if err := db.Save(&checkin).Error; err != nil {
		logger.Error("刷新二维码token失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "刷新失败"))
		return
	}

	// 计算剩余秒数
	refreshIn := int64(checkin.QrcodeRefreshInterval)

	// 生成二维码URL
	qrcodeURL := fmt.Sprintf("/api/classroom/checkin/%d/qrcode/scan?token=%s", checkinID, token)

	response := map[string]interface{}{
		"qrcodeToken": token,
		"qrcodeUrl":   qrcodeURL,
		"expiresAt":   expiresAt,
		"refreshIn":   refreshIn,
	}

	c.JSON(http.StatusOK, successResponse(response))
}

// SubmitQrcodeCheckin 学生通过二维码签到
func (h *Handler) SubmitQrcodeCheckin(c *gin.Context) {
	logger := utils.GetLogger()

	var req struct {
		Token     string `json:"token" binding:"required"`
		CheckinID uint64 `json:"checkinId" binding:"required"`
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

	// 查询签到
	var checkin model.ClassroomCheckin
	if err := db.Where("id = ?", req.CheckinID).First(&checkin).Error; err != nil {
		c.JSON(http.StatusOK, errorResponse(404, "签到不存在"))
		return
	}

	// 验证签到类型
	if checkin.CheckinType != "qrcode" {
		c.JSON(http.StatusOK, errorResponse(400, "不是二维码签到类型"))
		return
	}

	// 验证token
	if checkin.QrcodeToken != req.Token {
		c.JSON(http.StatusOK, errorResponse(400, "二维码token无效"))
		return
	}

	// 检查是否过期
	if checkin.QrcodeExpiresAt == nil || time.Now().After(*checkin.QrcodeExpiresAt) {
		c.JSON(http.StatusOK, errorResponse(400, "二维码已过期"))
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

	// 检查是否已经签到过
	var existingRecord model.ClassroomCheckinRecord
	if err := db.Where("checkin_id = ? AND uid = ?", req.CheckinID, uid).First(&existingRecord).Error; err == nil {
		c.JSON(http.StatusOK, errorResponse(400, "已经签到过了"))
		return
	}

	// 创建签到记录
	checkinTime := now
	record := &model.ClassroomCheckinRecord{
		CheckinID:   req.CheckinID,
		UID:         uid.(string),
		Status:      "present",
		CheckinTime: &checkinTime,
	}

	if err := db.Create(record).Error; err != nil {
		logger.Error("创建签到记录失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "签到失败"))
		return
	}

	logger.Info("二维码签到成功", zap.Uint64("checkin_id", req.CheckinID), zap.String("uid", uid.(string)))
	c.JSON(http.StatusOK, successResponse(nil))
}

// generateQrcodeToken 生成二维码token
func generateQrcodeToken(checkinID uint64, refreshInterval int) string {
	timestamp := time.Now().UnixNano()
	rand.Seed(timestamp)
	random := fmt.Sprintf("%08x", rand.Uint32())

	token := fmt.Sprintf("%d_%d_%s", checkinID, timestamp/1e6, random)
	return token
}

// ==================== 班级教师管理 ====================

// AddClassroomTeacher 为班级添加教师（仅超级管理员）
func (h *Handler) AddClassroomTeacher(c *gin.Context) {
	logger := utils.GetLogger()

	var req struct {
		ClassroomID uint64 `json:"classroomId" binding:"required"`
		TeacherID   string `json:"teacherId"`
		Username    string `json:"username"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	// 如果提供了 username，需要先查找用户的 UUID
	if req.Username != "" && req.TeacherID == "" {
		// 通过用户名查找用户
		var userInfo model.UserInfo
		if err := client.GetDB().Where("username = ?", req.Username).First(&userInfo).Error; err != nil {
			c.JSON(http.StatusOK, errorResponse(404, "用户不存在"))
			return
		}
		req.TeacherID = userInfo.UUID
	}

	// 如果既没有 teacherId 也没有 username，返回错误
	if req.TeacherID == "" && req.Username == "" {
		c.JSON(http.StatusOK, errorResponse(400, "请提供教师ID或用户名"))
		return
	}

	db := client.GetDB()

	// 检查班级是否存在
	var classroom model.Classroom
	if err := db.Where("id = ? AND status = 1", req.ClassroomID).First(&classroom).Error; err != nil {
		c.JSON(http.StatusOK, errorResponse(404, "班级不存在"))
		return
	}

	// 检查该用户是否已经是班级的学生
	var studentCount int64
	db.Model(&model.ClassroomStudent{}).
		Where("classroom_id = ? AND uid = ? AND status = 1", req.ClassroomID, req.TeacherID).
		Count(&studentCount)
	if studentCount > 0 {
		c.JSON(http.StatusOK, errorResponse(400, "该用户是班级学生，不能同时添加为教师"))
		return
	}

	// 检查该用户是否是班级的主教师
	if classroom.TeacherID == req.TeacherID {
		c.JSON(http.StatusOK, errorResponse(400, "该用户已经是班级的主教师"))
		return
	}

	// 检查教师是否已经是该班级的额外教师（只检查正常状态的记录）
	var existingTeacher model.ClassroomTeacher
	err := db.Model(&model.ClassroomTeacher{}).
		Where("classroom_id = ? AND teacher_id = ? AND status = 1", req.ClassroomID, req.TeacherID).
		First(&existingTeacher).Error

	if err == nil {
		// 记录已存在
		if existingTeacher.Status == 1 {
			// 已经是正常状态的教师
			c.JSON(http.StatusOK, errorResponse(400, "该教师已经是班级教师"))
			return
		} else {
			// 数据库错误或未找到记录
			c.JSON(http.StatusOK, errorResponse(400, "添加失败：该用户不在系统中或查询教师记录时发生错误"))
			return
		}
	}

	classroomTeacher := &model.ClassroomTeacher{
		ClassroomID: req.ClassroomID,
		TeacherID:   req.TeacherID,
		Status:      1,
	}

	if err := db.Create(classroomTeacher).Error; err != nil {
		logger.Error("添加班级教师失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "添加失败：保存教师记录到数据库时发生错误"))
		return
	}

	logger.Info("添加班级教师", zap.Uint64("classroom_id", req.ClassroomID), zap.String("teacher_id", req.TeacherID))
	c.JSON(http.StatusOK, successResponse(nil))
}

// RemoveClassroomTeacher 从班级移除教师（仅超级管理员）
func (h *Handler) RemoveClassroomTeacher(c *gin.Context) {
	logger := utils.GetLogger()

	var req struct {
		ClassroomID  uint64 `json:"classroomId" binding:"required"`
		TeacherID    string `json:"teacherId" binding:"required"`
		NewTeacherID string `json:"newTeacherId"` // 删除主教师时，必须指定新主教师
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	db := client.GetDB()

	// 获取班级信息（包含主教师ID）
	var classroom model.Classroom
	if err := db.Where("id = ? AND status = 1", req.ClassroomID).First(&classroom).Error; err != nil {
		c.JSON(http.StatusOK, errorResponse(404, "班级不存在"))
		return
	}

	// 检查要删除的是否是主教师
	isPrimaryTeacher := (req.TeacherID == classroom.TeacherID)

	if isPrimaryTeacher {
		// 删除主教师，必须指定新主教师
		if req.NewTeacherID == "" {
			c.JSON(http.StatusOK, errorResponse(400, "删除主教师必须指定新的主教师"))
			return
		}

		// 验证新主教师必须是该班级的教师
		var newTeacherExists int64
		db.Model(&model.ClassroomTeacher{}).
			Where("classroom_id = ? AND teacher_id = ? AND status = 1", req.ClassroomID, req.NewTeacherID).
			Count(&newTeacherExists)
		if newTeacherExists == 0 {
			c.JSON(http.StatusOK, errorResponse(400, "新主教师不是该班级的教师"))
			return
		}

		// 更新班级的主教师ID
		if err := db.Model(&classroom).Update("teacher_id", req.NewTeacherID).Error; err != nil {
			logger.Error("更新班级主教师失败", zap.Error(err))
			c.JSON(http.StatusOK, errorResponse(500, "更新主教师失败"))
			return
		}

		// 如果主教师在 classroom_teacher 表中，也需要软删除该记录
		db.Model(&model.ClassroomTeacher{}).
			Where("classroom_id = ? AND teacher_id = ?", req.ClassroomID, req.TeacherID).
			Update("status", 0)

		logger.Info("转移班级主教师",
			zap.Uint64("classroom_id", req.ClassroomID),
			zap.String("old_teacher_id", req.TeacherID),
			zap.String("new_teacher_id", req.NewTeacherID))

		c.JSON(http.StatusOK, successResponse(nil))
		return
	}

	// 删除非主教师，需要检查删除后是否还有教师

	// 统计该班级在 classroom_teacher 表中的额外教师数量
	var extraTeacherCount int64
	db.Model(&model.ClassroomTeacher{}).
		Where("classroom_id = ? AND status = 1", req.ClassroomID).
		Count(&extraTeacherCount)

	// 计算删除该教师后剩余的教师数量
	remainingTeachers := extraTeacherCount

	// 检查要删除的教师是否是额外教师
	var isExtraTeacherCount int64
	db.Model(&model.ClassroomTeacher{}).
		Where("classroom_id = ? AND teacher_id = ? AND status = 1", req.ClassroomID, req.TeacherID).
		Count(&isExtraTeacherCount)

	if isExtraTeacherCount > 0 {
		// 要删除的是额外教师，删除后数量减1
		remainingTeachers = extraTeacherCount - 1
	}

	// 检查主教师是否在额外教师列表中
	var primaryTeacherInExtra int64
	db.Model(&model.ClassroomTeacher{}).
		Where("classroom_id = ? AND teacher_id = ? AND status = 1", req.ClassroomID, classroom.TeacherID).
		Count(&primaryTeacherInExtra)

	if primaryTeacherInExtra == 0 && classroom.TeacherID != "" {
		// 主教师不在额外教师列表中，所以需要+1
		remainingTeachers++
	}

	// 如果删除后没有教师了，不允许删除
	if remainingTeachers <= 0 {
		c.JSON(http.StatusOK, errorResponse(400, "班级至少需要保留一个教师"))
		return
	}

	// 软删除教师关联
	result := db.Model(&model.ClassroomTeacher{}).
		Where("classroom_id = ? AND teacher_id = ?", req.ClassroomID, req.TeacherID).
		Update("status", 0)

	if result.Error != nil {
		logger.Error("移除班级教师失败", zap.Error(result.Error))
		c.JSON(http.StatusOK, errorResponse(500, "移除失败"))
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusOK, errorResponse(404, "教师关联不存在"))
		return
	}

	logger.Info("移除班级教师", zap.Uint64("classroom_id", req.ClassroomID), zap.String("teacher_id", req.TeacherID))
	c.JSON(http.StatusOK, successResponse(nil))
}

// GetClassroomTeachers 获取班级的所有教师列表
func (h *Handler) GetClassroomTeachers(c *gin.Context) {
	logger := utils.GetLogger()
	classroomIDStr := c.Param("classroomId")
	classroomID, err := strconv.ParseUint(classroomIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "classroomId参数格式错误"))
		return
	}

	db := client.GetDB()

	// 先查询班级信息，获取主教师ID
	var classroom model.Classroom
	if err := db.Where("id = ? AND status = 1", classroomID).
		First(&classroom).Error; err != nil {
		logger.Error("查询班级信息失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	// 查询额外的教师
	var classroomTeachers []model.ClassroomTeacher
	if err := db.Where("classroom_id = ? AND status = 1", classroomID).
		Order("create_time ASC").
		Find(&classroomTeachers).Error; err != nil {
		logger.Error("查询班级教师列表失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	// 为每个教师加载用户信息（先尝试从user_info，如果不存在则从user表）
	teacherUIDs := make([]string, len(classroomTeachers))
	for i, ct := range classroomTeachers {
		teacherUIDs[i] = ct.TeacherID
	}

	// 批量查询用户信息
	for i := range classroomTeachers {
		userInfo, err := client.GetUserInfoWithRealname(classroomTeachers[i].TeacherID)
		if err == nil {
			classroomTeachers[i].Teacher = &model.UserInfo{
				UUID:     userInfo.UUID,
				Username: userInfo.Username,
				Nickname: userInfo.Nickname,
				Realname: userInfo.Realname,
			}
		} else {
			logger.Warn("获取教师用户信息失败", zap.String("teacherId", classroomTeachers[i].TeacherID), zap.Error(err))
		}
	}

	// 如果主教师不在额外教师列表中，添加主教师
	primaryTeacherFound := false
	for _, ct := range classroomTeachers {
		if ct.TeacherID == classroom.TeacherID {
			primaryTeacherFound = true
			break
		}
	}

	if !primaryTeacherFound && classroom.TeacherID != "" {
		// 查询主教师信息（从 user_info 表获取）
		var teacher model.UserInfo
		err := db.Where("uuid = ?", classroom.TeacherID).First(&teacher).Error
		if err == nil {
			// 创建主教师的 ClassroomTeacher 记录用于返回
			// 使用 classroomID * 1000000 作为主教师的虚拟 ID，确保不为 0 且不与真实 ID 冲突
			primaryTeacherRecord := &model.ClassroomTeacher{
				ID:          classroomID * 1000000,
				ClassroomID: classroomID,
				TeacherID:   classroom.TeacherID,
				Status:      1,
				Teacher:     &teacher,
			}
			classroomTeachers = append([]model.ClassroomTeacher{*primaryTeacherRecord}, classroomTeachers...)
		} else {
			logger.Warn("获取主教师信息失败", zap.String("teacherId", classroom.TeacherID), zap.Error(err))
		}
	}

	c.JSON(http.StatusOK, successResponse(classroomTeachers))
}

// SearchTeachers 搜索教师（用于管理员添加班级教师时）
// 搜索所有用户，管理员可以将任何用户添加为班级教师
func (h *Handler) SearchTeachers(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		c.JSON(http.StatusOK, errorResponse(400, "关键词不能为空"))
		return
	}

	// 搜索所有用户（不限制角色）
	users, err := client.SearchAllUsers(keyword)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(500, "搜索失败"))
		return
	}

	// 转换为 model.UserInfo 格式
	result := make([]model.UserInfo, len(users))
	for i, u := range users {
		result[i] = model.UserInfo{
			UUID:     u.UUID,
			Username: u.Username,
			Nickname: u.Nickname,
			Realname: u.Realname,
		}
	}

	c.JSON(http.StatusOK, successResponse(result))
}

// ==================== 试卷库功能 ====================

// CreateExamPaper 创建试卷
func (h *Handler) CreateExamPaper(c *gin.Context) {
	logger := utils.GetLogger()

	var req struct {
		Title       string                     `json:"title" binding:"required"`
		Description string                     `json:"description"`
		IsShared    int                        `json:"isShared"`
		Questions   []ExamPaperQuestionRequest `json:"questions" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("创建试卷请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "用户未登录"))
		return
	}

	db := client.GetDB()

	// 计算总分和题目数量
	totalScore := 0
	for _, q := range req.Questions {
		totalScore += q.Score
	}

	// 创建试卷
	examPaper := &model.ExamPaper{
		Title:         req.Title,
		CreatorID:     uid.(string),
		IsShared:      req.IsShared,
		IsPublic:      0, // 主界面公开由管理员在后台统一控制
		TotalScore:    totalScore,
		QuestionCount: len(req.Questions),
		Description:   req.Description,
		Status:        1,
	}

	if err := db.Create(examPaper).Error; err != nil {
		logger.Error("创建试卷失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "创建失败"))
		return
	}

	// 创建试卷题目关联
	for i, q := range req.Questions {
		examPaperQuestion := &model.ExamPaperQuestion{
			ExamPaperID:   examPaper.ID,
			QuestionID:    q.QuestionID,
			ProblemID:     q.ProblemID,
			QuestionOrder: i + 1,
			QuestionType:  q.QuestionType,
			Score:         q.Score,
		}
		if err := db.Create(examPaperQuestion).Error; err != nil {
			logger.Error("创建试卷题目关联失败", zap.Error(err))
		}
	}

	logger.Info("创建试卷成功", zap.Uint64("paper_id", examPaper.ID), zap.String("title", req.Title))
	c.JSON(http.StatusOK, successResponse(examPaper))
}

// UpdateExamPaper 更新试卷 (已移除重复函数，使用下方的正确实现)

// ExamPaperQuestionRequest 试卷题目请求结构
type ExamPaperQuestionRequest struct {
	QuestionID   *uint64 `json:"questionId"`   // 客观题ID
	ProblemID    *string `json:"problemId"`    // 编程题ID
	QuestionType string  `json:"questionType"` // single_choice, multiple_choice, judge, subjective, programming
	Score        int     `json:"score"`
}

// GetExamPaperList 获取试卷列表
func (h *Handler) GetExamPaperList(c *gin.Context) {
	logger := utils.GetLogger()

	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "用户未登录"))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	keyword := c.Query("keyword")
	isSharedStr := c.Query("isShared")
	isPublicStr := c.Query("isPublic")

	db := client.GetDB()

	// 查询试卷：自己创建的 + 共享的
	query := db.Model(&model.ExamPaper{}).Where("status = 1 AND (creator_id = ? OR is_shared = 1)", uid)

	if keyword != "" {
		query = query.Where("title LIKE ?", "%"+keyword+"%")
	}

	if isSharedStr != "" {
		isShared, _ := strconv.Atoi(isSharedStr)
		query = query.Where("is_shared = ?", isShared)
	}
	if isPublicStr != "" {
		isPublic, _ := strconv.Atoi(isPublicStr)
		query = query.Where("is_public = ?", isPublic)
	}

	var total int64
	query.Count(&total)

	var papers []model.ExamPaper
	if err := query.Preload("Creator").
		Offset((page - 1) * limit).
		Limit(limit).
		Order("create_time DESC").
		Find(&papers).Error; err != nil {
		logger.Error("查询试卷列表失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	c.JSON(http.StatusOK, successResponse(map[string]interface{}{
		"total":  total,
		"page":   page,
		"limit":  limit,
		"papers": papers,
	}))
}

// GetExamPaperDetail 获取试卷详情
func (h *Handler) GetExamPaperDetail(c *gin.Context) {
	logger := utils.GetLogger()

	paperIDStr := c.Param("paperId")
	paperID, err := strconv.ParseUint(paperIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "试卷ID格式错误"))
		return
	}

	db := client.GetDB()

	var paper model.ExamPaper
	if err := db.Preload("Creator").
		Preload("Questions", func(db *gorm.DB) *gorm.DB {
			return db.Order("question_order ASC")
		}).
		Preload("Questions.Question").
		Where("id = ? AND status = 1", paperID).
		First(&paper).Error; err != nil {
		logger.Error("查询试卷详情失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(404, "试卷不存在"))
		return
	}

	c.JSON(http.StatusOK, successResponse(paper))
}

// GetPublicExamPaperList 获取主界面公开练习试卷列表（无需认证）
func (h *Handler) GetPublicExamPaperList(c *gin.Context) {
	logger := utils.GetLogger()

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	keyword := c.Query("keyword")
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	db := client.GetDB()
	query := db.Model(&model.ExamPaper{}).Where("status = 1 AND is_public = 1")
	if keyword != "" {
		query = query.Where("title LIKE ?", "%"+keyword+"%")
	}

	var total int64
	query.Count(&total)

	var papers []model.ExamPaper
	if err := query.Preload("Creator").
		Offset((page - 1) * limit).
		Limit(limit).
		Order("create_time DESC").
		Find(&papers).Error; err != nil {
		logger.Error("查询公开试卷列表失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	c.JSON(http.StatusOK, successResponse(map[string]interface{}{
		"total":  total,
		"page":   page,
		"limit":  limit,
		"papers": papers,
	}))
}

// GetPublicExamPaperDetail 获取主界面公开练习试卷详情（无需认证）
func (h *Handler) GetPublicExamPaperDetail(c *gin.Context) {
	logger := utils.GetLogger()

	paperIDStr := c.Param("paperId")
	paperID, err := strconv.ParseUint(paperIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "试卷ID格式错误"))
		return
	}

	db := client.GetDB()

	var paper model.ExamPaper
	if err := db.Preload("Creator").
		Preload("Questions", func(db *gorm.DB) *gorm.DB {
			return db.Order("question_order ASC")
		}).
		Preload("Questions.Question").
		Where("id = ? AND status = 1 AND is_public = 1", paperID).
		First(&paper).Error; err != nil {
		logger.Error("查询公开试卷详情失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(404, "试卷不存在或未公开"))
		return
	}

	// 公开详情默认不返回标准答案与解析，避免一次性暴露全部答案
	for i := range paper.Questions {
		if paper.Questions[i].Question != nil {
			paper.Questions[i].Question.Answer = ""
			paper.Questions[i].Question.Analysis = ""
		}
	}

	c.JSON(http.StatusOK, successResponse(paper))
}

// GetPublicExamPaperQuestionAnswer 获取公开练习某题标准答案（按题懒加载）
func (h *Handler) GetPublicExamPaperQuestionAnswer(c *gin.Context) {
	logger := utils.GetLogger()

	paperIDStr := c.Param("paperId")
	paperID, err := strconv.ParseUint(paperIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "试卷ID格式错误"))
		return
	}

	questionIDStr := c.Param("questionId")
	questionID, err := strconv.ParseUint(questionIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "题目ID格式错误"))
		return
	}

	db := client.GetDB()

	// 校验公开试卷存在
	var paper model.ExamPaper
	if err := db.Where("id = ? AND status = 1 AND is_public = 1", paperID).First(&paper).Error; err != nil {
		c.JSON(http.StatusOK, errorResponse(404, "试卷不存在或未公开"))
		return
	}

	// 校验题目属于该试卷
	var paperQuestion model.ExamPaperQuestion
	if err := db.Where("exam_paper_id = ? AND question_id = ?", paperID, questionID).First(&paperQuestion).Error; err != nil {
		c.JSON(http.StatusOK, errorResponse(404, "题目不在当前试卷中"))
		return
	}

	var question model.QuestionBank
	if err := db.Where("id = ? AND status = 1", questionID).First(&question).Error; err != nil {
		logger.Error("查询题目答案失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(404, "题目不存在"))
		return
	}

	c.JSON(http.StatusOK, successResponse(map[string]interface{}{
		"questionId": question.ID,
		"answer":     question.Answer,
		"analysis":   question.Analysis,
	}))
}

// UpdateExamPaper 更新试卷
func (h *Handler) UpdateExamPaper(c *gin.Context) {
	logger := utils.GetLogger()

	paperIDStr := c.Param("paperId")
	paperID, err := strconv.ParseUint(paperIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "试卷ID格式错误"))
		return
	}

	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "用户未登录"))
		return
	}

	var req struct {
		Title       *string                    `json:"title"`
		Description *string                    `json:"description"`
		IsShared    *int                       `json:"isShared"`
		Questions   []ExamPaperQuestionRequest `json:"questions"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("更新试卷请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	logger.Info("更新试卷请求", zap.Uint64("paper_id", paperID),
		zap.Bool("has_title", req.Title != nil),
		zap.Bool("has_description", req.Description != nil),
		zap.Bool("has_isShared", req.IsShared != nil),
		zap.Int("questions_count", len(req.Questions)))
	if req.IsShared != nil {
		logger.Info("isShared值", zap.Int("isShared", *req.IsShared))
	}

	db := client.GetDB()

	// 检查试卷是否存在以及权限
	var paper model.ExamPaper
	if err := db.Where("id = ? AND status = 1", paperID).First(&paper).Error; err != nil {
		c.JSON(http.StatusOK, errorResponse(404, "试卷不存在"))
		return
	}

	// 只有创建者可以修改
	if paper.CreatorID != uid.(string) {
		c.JSON(http.StatusOK, errorResponse(403, "无权修改此试卷"))
		return
	}

	// 更新基本信息
	updates := map[string]interface{}{}
	if req.Title != nil {
		updates["title"] = *req.Title
		paper.Title = *req.Title
	}
	if req.Description != nil {
		updates["description"] = *req.Description
		paper.Description = *req.Description
	}
	if req.IsShared != nil {
		updates["is_shared"] = *req.IsShared
		paper.IsShared = *req.IsShared
	}

	if len(updates) > 0 {
		if err := db.Model(&paper).Updates(updates).Error; err != nil {
			logger.Error("更新试卷失败", zap.Error(err))
			c.JSON(http.StatusOK, errorResponse(500, "更新失败"))
			return
		}
	}

	// 如果提供了题目列表，更新题目
	if req.Questions != nil && len(req.Questions) > 0 {
		// 删除旧的题目关联
		db.Where("exam_paper_id = ?", paperID).Delete(&model.ExamPaperQuestion{})

		// 计算新的总分和题目数量
		totalScore := 0
		for i, q := range req.Questions {
			totalScore += q.Score
			examPaperQuestion := &model.ExamPaperQuestion{
				ExamPaperID:   paperID,
				QuestionID:    q.QuestionID,
				ProblemID:     q.ProblemID,
				QuestionOrder: i + 1,
				QuestionType:  q.QuestionType,
				Score:         q.Score,
			}
			if err := db.Create(examPaperQuestion).Error; err != nil {
				logger.Error("创建试卷题目关联失败", zap.Error(err))
			}
		}

		// 更新总分和题目数量
		db.Model(&paper).Updates(map[string]interface{}{
			"total_score":    totalScore,
			"question_count": len(req.Questions),
		})
	}

	logger.Info("更新试卷成功", zap.Uint64("paper_id", paperID))
	c.JSON(http.StatusOK, successResponse(nil))
}

// DeleteExamPaper 删除试卷（软删除）
func (h *Handler) DeleteExamPaper(c *gin.Context) {
	logger := utils.GetLogger()

	paperIDStr := c.Param("paperId")
	paperID, err := strconv.ParseUint(paperIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "试卷ID格式错误"))
		return
	}

	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "用户未登录"))
		return
	}

	db := client.GetDB()

	// 检查试卷是否存在
	var paper model.ExamPaper
	if err := db.Where("id = ? AND status = 1", paperID).First(&paper).Error; err != nil {
		c.JSON(http.StatusOK, errorResponse(404, "试卷不存在"))
		return
	}

	// 只有创建者可以删除
	if paper.CreatorID != uid.(string) {
		c.JSON(http.StatusOK, errorResponse(403, "无权删除此试卷"))
		return
	}

	// 软删除
	if err := db.Model(&paper).Update("status", 0).Error; err != nil {
		logger.Error("删除试卷失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "删除失败"))
		return
	}

	logger.Info("删除试卷成功", zap.Uint64("paper_id", paperID))
	c.JSON(http.StatusOK, successResponse(nil))
}

// ImportExamPaperToHomework 导入试卷到作业
func (h *Handler) ImportExamPaperToHomework(c *gin.Context) {
	logger := utils.GetLogger()

	// 使用浮点数类型接收参数（兼容前端发送的数字类型）
	var req struct {
		PaperID     float64 `json:"paperId" binding:"required"`
		ClassroomID float64 `json:"classroomId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("导入试卷请求参数绑定失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	logger.Info("导入试卷请求参数",
		zap.Float64("paper_id", req.PaperID),
		zap.Float64("classroom_id", req.ClassroomID))

	// 验证参数必须是正整数（不能是小数或负数）
	if req.PaperID <= 0 || req.PaperID != float64(uint64(req.PaperID)) {
		logger.Warn("paperId 参数格式错误", zap.Float64("paperId", req.PaperID))
		c.JSON(http.StatusOK, errorResponse(400, "paperId 参数格式错误"))
		return
	}

	if req.ClassroomID <= 0 || req.ClassroomID != float64(uint64(req.ClassroomID)) {
		logger.Warn("classroomId 参数格式错误", zap.Float64("classroomId", req.ClassroomID))
		c.JSON(http.StatusOK, errorResponse(400, "classroomId 参数格式错误"))
		return
	}

	// 转换为 uint64
	paperID := uint64(req.PaperID)
	classroomID := uint64(req.ClassroomID)

	// 获取当前用户
	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "用户未登录"))
		return
	}

	logger.Info("导入试卷请求", zap.Uint64("paper_id", paperID), zap.Uint64("classroom_id", classroomID), zap.String("uid", uid.(string)))

	db := client.GetDB()

	// 查询试卷（只能导入共享试卷或自己创建的试卷）
	var paper model.ExamPaper
	if err := db.Preload("Questions", func(db *gorm.DB) *gorm.DB {
		return db.Order("question_order ASC")
	}).
		Preload("Questions.Question").
		Where("id = ? AND status = 1", paperID).
		First(&paper).Error; err != nil {
		logger.Error("查询试卷失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(404, "试卷不存在"))
		return
	}

	// 权限检查：
	// 1. 管理员（root或admin）可以导入所有试卷
	// 2. 普通教师只能导入共享试卷或自己创建的试卷
	isAdmin := false
	if roles, err := middleware.GetUserRoles(db, uid.(string)); err == nil {
		isAdmin = middleware.HasAnyRole(roles, []string{middleware.RoleRoot, middleware.RoleAdmin})
	}

	if !isAdmin && paper.IsShared != 1 && paper.CreatorID != uid.(string) {
		logger.Warn("用户尝试导入无权访问的试卷",
			zap.String("uid", uid.(string)),
			zap.Uint64("paper_id", paper.ID),
			zap.String("creator_id", paper.CreatorID))
		c.JSON(http.StatusOK, errorResponse(403, "无权导入此试卷"))
		return
	}

	// 构建返回数据（包含客观题详情和编程题ID）
	result := make([]map[string]interface{}, 0, len(paper.Questions))
	for _, q := range paper.Questions {
		item := map[string]interface{}{
			"questionOrder": q.QuestionOrder,
			"questionType":  q.QuestionType,
			"score":         q.Score,
		}

		if q.QuestionType == "programming" {
			// 编程题
			if q.ProblemID != nil {
				item["problemId"] = *q.ProblemID
				item["title"] = "BingOJ 编程题 - " + *q.ProblemID
				item["type"] = "programming"
				item["difficulty"] = 5
			}
		} else {
			// 客观题
			if q.Question != nil {
				item["questionId"] = q.Question.ID
				item["id"] = q.Question.ID
				item["type"] = q.Question.Type
				item["title"] = q.Question.Title
				item["difficulty"] = q.Question.Difficulty
			}
		}

		result = append(result, item)
	}

	c.JSON(http.StatusOK, successResponse(map[string]interface{}{
		"questions":     result,
		"totalScore":    paper.TotalScore,
		"questionCount": paper.QuestionCount,
	}))
}

// ==================== 管理员专用API - 试卷管理 ====================

// AdminGetExamPaperList 管理员获取所有试卷列表（包括私有和共享）
func (h *Handler) AdminGetExamPaperList(c *gin.Context) {
	logger := utils.GetLogger()

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	keyword := c.Query("keyword")
	isSharedStr := c.Query("isShared")
	isPublicStr := c.Query("isPublic")

	db := client.GetDB()

	// 管理员可以看到所有试卷
	query := db.Model(&model.ExamPaper{}).Where("status = 1")

	if keyword != "" {
		query = query.Where("title LIKE ?", "%"+keyword+"%")
	}

	if isSharedStr != "" {
		isShared, _ := strconv.Atoi(isSharedStr)
		query = query.Where("is_shared = ?", isShared)
	}
	if isPublicStr != "" {
		isPublic, _ := strconv.Atoi(isPublicStr)
		query = query.Where("is_public = ?", isPublic)
	}

	var total int64
	query.Count(&total)

	var papers []model.ExamPaper
	if err := query.Preload("Creator").
		Preload("Questions.Question").
		Preload("Questions.ExamPaper").
		Offset((page - 1) * limit).
		Limit(limit).
		Order("create_time DESC").
		Find(&papers).Error; err != nil {
		logger.Error("管理员查询试卷列表失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	c.JSON(http.StatusOK, successResponse(map[string]interface{}{
		"total":  total,
		"page":   page,
		"limit":  limit,
		"papers": papers,
	}))
}

// AdminUpdateExamPaper 管理员更新试卷
func (h *Handler) AdminUpdateExamPaper(c *gin.Context) {
	logger := utils.GetLogger()

	paperIDStr := c.Param("paperId")
	paperID, err := strconv.ParseUint(paperIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "试卷ID格式错误"))
		return
	}

	var req struct {
		Title       string                     `json:"title" binding:"required"`
		Description string                     `json:"description"`
		IsShared    int                        `json:"isShared"`
		IsPublic    int                        `json:"isPublic"`
		Questions   []ExamPaperQuestionRequest `json:"questions"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("管理员更新试卷请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	db := client.GetDB()

	// 检查试卷是否存在
	var paper model.ExamPaper
	if err := db.Where("id = ? AND status = 1", paperID).First(&paper).Error; err != nil {
		c.JSON(http.StatusOK, errorResponse(404, "试卷不存在"))
		return
	}

	// 计算总分和题目数量
	totalScore := 0
	for _, q := range req.Questions {
		totalScore += q.Score
	}

	// 开启事务
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新试卷基本信息
	if err := tx.Model(&paper).Updates(map[string]interface{}{
		"title":          req.Title,
		"description":    req.Description,
		"is_shared":      req.IsShared,
		"is_public":      req.IsPublic,
		"total_score":    totalScore,
		"question_count": len(req.Questions),
	}).Error; err != nil {
		tx.Rollback()
		logger.Error("管理员更新试卷失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "更新失败"))
		return
	}

	// 删除旧的题目关联
	if err := tx.Where("exam_paper_id = ?", paperID).Delete(&model.ExamPaperQuestion{}).Error; err != nil {
		tx.Rollback()
		logger.Error("管理员删除旧试卷题目关联失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "更新失败"))
		return
	}

	// 创建新的题目关联
	for order, q := range req.Questions {
		paperQuestion := model.ExamPaperQuestion{
			ExamPaperID:   paperID,
			QuestionOrder: order + 1,
			QuestionType:  q.QuestionType,
			Score:         q.Score,
		}

		if q.QuestionType == "programming" {
			paperQuestion.ProblemID = q.ProblemID
		} else {
			paperQuestion.QuestionID = q.QuestionID
		}

		if err := tx.Create(&paperQuestion).Error; err != nil {
			tx.Rollback()
			logger.Error("管理员创建试卷题目关联失败", zap.Error(err))
			c.JSON(http.StatusOK, errorResponse(500, "更新失败"))
			return
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		logger.Error("管理员更新试卷事务提交失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "更新失败"))
		return
	}

	logger.Info("管理员更新试卷成功", zap.Uint64("paper_id", paperID))
	c.JSON(http.StatusOK, successResponse(nil))
}

// AdminDeleteExamPaper 管理员删除试卷
func (h *Handler) AdminDeleteExamPaper(c *gin.Context) {
	logger := utils.GetLogger()

	paperIDStr := c.Param("paperId")
	paperID, err := strconv.ParseUint(paperIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "试卷ID格式错误"))
		return
	}

	db := client.GetDB()

	// 软删除
	if err := db.Model(&model.ExamPaper{}).Where("id = ?", paperID).Update("status", 0).Error; err != nil {
		logger.Error("管理员删除试卷失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "删除失败"))
		return
	}

	logger.Info("管理员删除试卷成功", zap.Uint64("paper_id", paperID))
	c.JSON(http.StatusOK, successResponse(nil))
}

// SearchUsersForRoleManagement 搜索用户（用于角色管理）
// 允许所有管理员调用，用于查找用户并设置角色
func (h *Handler) SearchUsersForRoleManagement(c *gin.Context) {
	logger := utils.GetLogger()

	// 获取搜索参数
	keyword := c.Query("keyword")
	if keyword == "" {
		c.JSON(http.StatusOK, errorResponse(400, "关键词不能为空"))
		return
	}

	currentPageStr := c.DefaultQuery("currentPage", "1")
	limitStr := c.DefaultQuery("limit", "20")

	currentPage, _ := strconv.Atoi(currentPageStr)
	limit, _ := strconv.Atoi(limitStr)

	if currentPage < 1 {
		currentPage = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (currentPage - 1) * limit

	logger.Info("管理员搜索用户",
		zap.String("keyword", keyword),
		zap.Int("page", currentPage),
		zap.Int("limit", limit))

	db := client.GetDB()

	// 查询用户总数
	var total int64
	countQuery := db.Table("user_info").
		Where("username LIKE ? OR realname LIKE ? OR nickname LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	if err := countQuery.Count(&total).Error; err != nil {
		logger.Error("查询用户总数失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	// 查询用户列表
	type UserInfoDB struct {
		UUID     string `gorm:"column:uuid"`
		Username string `gorm:"column:username"`
		Nickname string `gorm:"column:nickname"`
		Realname string `gorm:"column:realname"`
		Email    string `gorm:"column:email"`
		Status   int    `gorm:"column:status"`
	}

	var usersDB []UserInfoDB
	query := db.Table("user_info").
		Select("uuid, username, nickname, realname, email, status").
		Where("username LIKE ? OR realname LIKE ? OR nickname LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%").
		Order("uuid ASC").
		Limit(limit).
		Offset(offset)

	if err := query.Find(&usersDB).Error; err != nil {
		logger.Error("查询用户列表失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	// 构建返回结果
	type UserRecord struct {
		UUID     string `json:"uid"`
		Username string `json:"username"`
		Nickname string `json:"nickname"`
		Realname string `json:"realname"`
		Email    string `json:"email"`
		Status   int    `json:"status"`
	}

	records := make([]UserRecord, len(usersDB))
	for i, u := range usersDB {
		records[i] = UserRecord{
			UUID:     u.UUID,
			Username: u.Username,
			Nickname: u.Nickname,
			Realname: u.Realname,
			Email:    u.Email,
			Status:   u.Status,
		}
	}

	type PagedResult struct {
		Records []UserRecord `json:"records"`
		Total   int64        `json:"total"`
	}

	c.JSON(http.StatusOK, successResponse(PagedResult{
		Records: records,
		Total:   total,
	}))
}

// ==================== 班级角色申请相关接口 ====================

// RoleApplicationRequest 申请角色请求
type RoleApplicationRequest struct {
	Role   string `json:"role" binding:"required"` // teacher, student
	Reason string `json:"reason"`
}

// CreateRoleApplication 申请班级角色（用户）
func (h *Handler) CreateRoleApplication(c *gin.Context) {
	logger := utils.GetLogger()

	// 获取当前用户ID
	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "未登录"))
		return
	}

	var req RoleApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	// 验证角色类型
	if req.Role != "teacher" && req.Role != "student" {
		c.JSON(http.StatusOK, errorResponse(400, "角色类型必须是 teacher 或 student"))
		return
	}

	db := client.GetDB()

	// 检查用户是否已经拥有该角色
	var existingRole model.ClassroomUserRole
	err := db.Where("uid = ? AND role = ?", uid.(string), req.Role).First(&existingRole).Error
	if err == nil {
		c.JSON(http.StatusOK, errorResponse(400, "您已经拥有该角色，无需申请"))
		return
	} else if err != gorm.ErrRecordNotFound {
		logger.Error("查询用户角色失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	// 检查是否有待审批的申请
	var pendingApplication model.ClassroomRoleRequest
	err = db.Where("uid = ? AND role = ? AND status = 0", uid.(string), req.Role).First(&pendingApplication).Error
	if err == nil {
		c.JSON(http.StatusOK, errorResponse(400, "您已有待审批的"+getRoleName(req.Role)+"申请，请耐心等待"))
		return
	} else if err != gorm.ErrRecordNotFound {
		logger.Error("查询申请记录失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	// 创建申请记录
	application := &model.ClassroomRoleRequest{
		UID:    uid.(string),
		Role:   req.Role,
		Reason: req.Reason,
		Status: 0, // 待审批
	}

	if err := db.Create(application).Error; err != nil {
		logger.Error("创建申请失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "申请失败"))
		return
	}

	logger.Info("用户申请班级角色",
		zap.String("uid", uid.(string)),
		zap.String("role", req.Role),
		zap.Uint64("applicationId", application.ID))

	c.JSON(http.StatusOK, successResponse(gin.H{
		"applicationId": application.ID,
		"message":       "申请已提交，请等待管理员审批",
	}))
}

// GetMyRoleApplications 获取当前用户的角色申请列表
func (h *Handler) GetMyRoleApplications(c *gin.Context) {
	logger := utils.GetLogger()

	// 获取当前用户ID
	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "未登录"))
		return
	}

	// 获取查询参数
	status := c.Query("status") // 0: 待审批, 1: 已批准, 2: 已拒绝, all（默认只返回待审批）

	db := client.GetDB()

	// 构建查询：只查询当前用户的申请
	query := db.Model(&model.ClassroomRoleRequest{}).Where("uid = ?", uid.(string))

	// 状态过滤
	if status != "" && status != "all" {
		statusInt, _ := strconv.Atoi(status)
		query = query.Where("status = ?", statusInt)
	} else {
		// 默认只返回待审批的申请
		query = query.Where("status = 0")
	}

	// 按创建时间倒序查询
	var applications []model.ClassroomRoleRequest
	if err := query.Order("create_time DESC").Find(&applications).Error; err != nil {
		logger.Error("查询申请列表失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	// 转换为返回格式
	type ApplicationResponse struct {
		ID        uint64 `json:"id"`
		Role      string `json:"role"`
		Reason    string `json:"reason"`
		Status    int    `json:"status"`
		CreatedAt string `json:"createdAt"`
	}

	result := make([]ApplicationResponse, 0, len(applications))
	for _, app := range applications {
		result = append(result, ApplicationResponse{
			ID:        app.ID,
			Role:      app.Role,
			Reason:    app.Reason,
			Status:    app.Status,
			CreatedAt: app.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	c.JSON(http.StatusOK, successResponse(gin.H{
		"applications": result,
		"count":        len(result),
	}))
}

// GetRoleApplications 获取角色申请列表（管理员）
func (h *Handler) GetRoleApplications(c *gin.Context) {
	logger := utils.GetLogger()

	// 获取查询参数
	role := c.Query("role")     // teacher, student, all
	status := c.Query("status") // 0: 待审批, 1: 已批准, 2: 已拒绝, all
	currentPage, _ := strconv.Atoi(c.DefaultQuery("currentPage", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	db := client.GetDB()

	// 验证管理员权限（从token中获取角色）
	roles, exists := c.Get("roles")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(403, "无权限访问"))
		return
	}

	roleList, ok := roles.([]string)
	if !ok {
		c.JSON(http.StatusOK, errorResponse(403, "无权限访问"))
		return
	}

	// 判断用户权限等级
	isSuperAdmin := false
	isProblemAdmin := false

	for _, role := range roleList {
		switch role {
		case "root", "admin":
			isSuperAdmin = true
		case "problem_admin":
			isProblemAdmin = true
		}
	}

	// 超级管理员拥有所有权限
	if isSuperAdmin {
		isProblemAdmin = true
	}

	if !isSuperAdmin && !isProblemAdmin {
		c.JSON(http.StatusOK, errorResponse(403, "无权限访问"))
		return
	}

	// 构建查询
	query := db.Model(&model.ClassroomRoleRequest{}).
		Preload("Applicant")

	// 角色过滤
	if role != "" && role != "all" {
		query = query.Where("role = ?", role)
	}

	// 状态过滤
	if status != "" && status != "all" {
		statusInt, _ := strconv.Atoi(status)
		query = query.Where("status = ?", statusInt)
	}

	// 权限过滤：问题管理员和普通管理员只能看到学生申请
	if !isSuperAdmin && isProblemAdmin {
		query = query.Where("role = ?", "student")
	}

	// 获取总数
	var total int64
	query.Count(&total)

	// 分页查询
	var applications []model.ClassroomRoleRequest
	offset := (currentPage - 1) * limit
	if err := query.Preload("Applicant").Preload("Reviewer").Order("create_time DESC").Offset(offset).Limit(limit).Find(&applications).Error; err != nil {
		logger.Error("查询申请列表失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	type ApplicationRecord struct {
		ID           uint64     `json:"id"`
		UID          string     `json:"uid"`
		Username     string     `json:"username"`
		Realname     string     `json:"realname"`
		Email        string     `json:"email"`
		Role         string     `json:"role"`
		RoleName     string     `json:"roleName"`
		Reason       string     `json:"reason"`
		Status       int        `json:"status"`
		StatusName   string     `json:"statusName"`
		ReviewerUID  string     `json:"reviewerUid,omitempty"`
		ReviewerName string     `json:"reviewerName,omitempty"`
		ReviewTime   *time.Time `json:"reviewTime,omitempty"`
		ReviewNote   string     `json:"reviewNote,omitempty"`
		CreatedAt    time.Time  `json:"createdAt"`
	}

	records := make([]ApplicationRecord, len(applications))
	for i, app := range applications {
		statusName := getStatusName(app.Status)
		roleName := getRoleName(app.Role)

		record := ApplicationRecord{
			ID:         app.ID,
			UID:        app.UID,
			Role:       app.Role,
			RoleName:   roleName,
			Reason:     app.Reason,
			Status:     app.Status,
			StatusName: statusName,
			ReviewTime: app.ReviewTime,
			ReviewNote: app.ReviewNote,
			CreatedAt:  app.CreatedAt,
		}

		if app.Applicant != nil {
			record.Username = app.Applicant.Username
			record.Realname = app.Applicant.Realname
			record.Email = app.Applicant.Nickname
		}

		if app.Reviewer != nil {
			record.ReviewerUID = app.ReviewerUID
			record.ReviewerName = app.Reviewer.Username
		}

		records[i] = record
	}

	type PagedResult struct {
		Records []ApplicationRecord `json:"records"`
		Total   int64               `json:"total"`
	}

	c.JSON(http.StatusOK, successResponse(PagedResult{
		Records: records,
		Total:   total,
	}))
}

// ReviewRoleApplicationRequest 审批申请请求
type ReviewRoleApplicationRequest struct {
	ApplicationIDs []uint64 `json:"applicationIds" binding:"required"`
	Action         string   `json:"action" binding:"required"` // approve, reject
	ReviewNote     string   `json:"reviewNote"`
}

// ReviewRoleApplication 批量审批角色申请（管理员）
func (h *Handler) ReviewRoleApplication(c *gin.Context) {
	logger := utils.GetLogger()

	// 获取当前用户ID
	reviewerUID, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "未登录"))
		return
	}

	var req ReviewRoleApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	// 验证操作类型
	if req.Action != "approve" && req.Action != "reject" {
		c.JSON(http.StatusOK, errorResponse(400, "操作类型必须是 approve 或 reject"))
		return
	}

	db := client.GetDB()

	// 验证管理员权限（从token中获取角色）
	roles, exists := c.Get("roles")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(403, "无权限访问"))
		return
	}

	roleList, ok := roles.([]string)
	if !ok {
		c.JSON(http.StatusOK, errorResponse(403, "无权限访问"))
		return
	}

	// 判断用户权限等级
	isSuperAdmin := false
	isProblemAdmin := false

	for _, role := range roleList {
		switch role {
		case "root", "admin":
			isSuperAdmin = true
		case "problem_admin":
			isProblemAdmin = true
		}
	}

	// 超级管理员拥有所有权限
	if isSuperAdmin {
		isProblemAdmin = true
	}

	if !isSuperAdmin && !isProblemAdmin {
		c.JSON(http.StatusOK, errorResponse(403, "无权限访问"))
		return
	}

	// 查询申请记录
	var applications []model.ClassroomRoleRequest
	if err := db.Where("id IN ?", req.ApplicationIDs).Find(&applications).Error; err != nil {
		logger.Error("查询申请记录失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	if len(applications) == 0 {
		c.JSON(http.StatusOK, errorResponse(404, "申请不存在"))
		return
	}

	now := time.Now()
	successCount := 0
	failedCount := 0

	for _, app := range applications {
		// 只能审批待审批的申请
		if app.Status != 0 {
			failedCount++
			continue
		}

		// 权限检查
		if app.Role == "teacher" && !isSuperAdmin {
			logger.Warn("非超级管理员尝试审批教师申请",
				zap.String("reviewer", reviewerUID.(string)),
				zap.Uint64("applicationId", app.ID))
			failedCount++
			continue
		}

		// 如果是批准操作，添加角色
		if req.Action == "approve" {
			// 检查是否已有该角色
			var existingRole model.ClassroomUserRole
			err := db.Where("uid = ? AND role = ?", app.UID, app.Role).First(&existingRole).Error

			if err == nil {
				// 已有角色，直接标记申请为已批准
				app.Status = 1
				app.ReviewerUID = reviewerUID.(string)
				app.ReviewTime = &now
				app.ReviewNote = req.ReviewNote
				if err := db.Save(&app).Error; err != nil {
					logger.Error("更新申请状态失败", zap.Error(err))
					failedCount++
				} else {
					successCount++
				}
			} else if err == gorm.ErrRecordNotFound {
				// 没有角色，需要添加
				newRole := &model.ClassroomUserRole{
					UID:  app.UID,
					Role: app.Role,
				}

				// 使用事务
				tx := db.Begin()
				if err := tx.Create(newRole).Error; err != nil {
					tx.Rollback()
					logger.Error("创建用户角色失败", zap.Error(err))
					failedCount++
					continue
				}

				// 更新申请状态
				app.Status = 1
				app.ReviewerUID = reviewerUID.(string)
				app.ReviewTime = &now
				app.ReviewNote = req.ReviewNote
				if err := tx.Save(&app).Error; err != nil {
					tx.Rollback()
					logger.Error("更新申请状态失败", zap.Error(err))
					failedCount++
					continue
				}

				tx.Commit()
				successCount++
			} else {
				logger.Error("查询用户角色失败", zap.Error(err))
				failedCount++
			}
		} else {
			// 拒绝操作
			app.Status = 2
			app.ReviewerUID = reviewerUID.(string)
			app.ReviewTime = &now
			app.ReviewNote = req.ReviewNote
			if err := db.Save(&app).Error; err != nil {
				logger.Error("更新申请状态失败", zap.Error(err))
				failedCount++
			} else {
				successCount++
			}
		}

		logger.Info("审批角色申请",
			zap.String("reviewer", reviewerUID.(string)),
			zap.String("action", req.Action),
			zap.String("applicant", app.UID),
			zap.String("role", app.Role),
			zap.Uint64("applicationId", app.ID))
	}

	result := gin.H{
		"successCount": successCount,
		"failedCount":  failedCount,
		"total":        len(applications),
	}

	if successCount > 0 {
		result["message"] = fmt.Sprintf("成功审批 %d 个申请", successCount)
		c.JSON(http.StatusOK, successResponse(result))
	} else {
		c.JSON(http.StatusOK, errorResponse(400, "审批失败，没有申请被处理"))
	}
}

// getRoleName 获取角色名称
func getRoleName(role string) string {
	switch role {
	case "teacher":
		return "教师"
	case "student":
		return "学生"
	default:
		return role
	}
}

// getStatusName 获取状态名称
func getStatusName(status int) string {
	switch status {
	case 0:
		return "待审批"
	case 1:
		return "已批准"
	case 2:
		return "已拒绝"
	default:
		return "未知"
	}
}

// CancelRoleApplication 取消角色申请（用户）
func (h *Handler) CancelRoleApplication(c *gin.Context) {
	logger := utils.GetLogger()

	// 获取当前用户ID
	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "未登录"))
		return
	}

	// 获取申请ID
	applicationIDStr := c.Param("applicationId")
	applicationID, err := strconv.ParseUint(applicationIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "申请ID格式错误"))
		return
	}

	db := client.GetDB()

	// 查询申请记录
	var application model.ClassroomRoleRequest
	if err := db.Where("id = ? AND uid = ?", applicationID, uid.(string)).First(&application).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, errorResponse(404, "申请不存在"))
		} else {
			logger.Error("查询申请失败", zap.Error(err))
			c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		}
		return
	}

	// 只能取消待审批的申请
	if application.Status != 0 {
		statusName := getStatusName(application.Status)
		c.JSON(http.StatusOK, errorResponse(400, "只能取消待审批的申请，当前状态："+statusName))
		return
	}

	// 删除申请记录
	if err := db.Delete(&application).Error; err != nil {
		logger.Error("取消申请失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "取消失败"))
		return
	}

	logger.Info("用户取消角色申请",
		zap.String("uid", uid.(string)),
		zap.Uint64("applicationId", applicationID),
		zap.String("role", application.Role))

	c.JSON(http.StatusOK, successResponse(gin.H{
		"message": "申请已取消",
	}))
}
