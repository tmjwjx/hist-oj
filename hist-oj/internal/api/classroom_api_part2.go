package api

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/hoj/hist-oj/internal/client"
	"github.com/hoj/hist-oj/internal/config"
	middlewarepkg "github.com/hoj/hist-oj/internal/middleware"
	"github.com/hoj/hist-oj/internal/model"
	"github.com/hoj/hist-oj/internal/service"
	"github.com/hoj/hist-oj/internal/utils"
)

// compareArrays 比较两个字符串数组是否相同（不考虑顺序）
func compareArrays(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	sort.Strings(a)
	sort.Strings(b)
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// normalizeJudgeAnswer 将判断题答案规范化为统一格式进行比较
// 仅支持: "true", "false"
// 非法值返回空字符串
func normalizeJudgeAnswer(answer string) string {
	normalized := strings.ToLower(strings.TrimSpace(answer))
	if normalized == "true" {
		return "true"
	}
	if normalized == "false" {
		return "false"
	}
	return ""
}

// compareJudgeAnswers 比较判断题答案(严格 true/false)
func compareJudgeAnswers(studentAnswer, correctAnswer string) bool {
	normalizedStudent := normalizeJudgeAnswer(studentAnswer)
	normalizedCorrect := normalizeJudgeAnswer(correctAnswer)
	if normalizedStudent == "" || normalizedCorrect == "" {
		return false
	}
	return normalizedStudent == normalizedCorrect
}

// FlexibleUint64 支持 JSON 中 number 或 string 两种格式的无符号整数
type FlexibleUint64 uint64

func (u *FlexibleUint64) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || bytes.Equal(data, []byte("null")) {
		return nil
	}

	// 兼容前端把数字 ID 序列化为字符串的场景
	if data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		s = strings.TrimSpace(s)
		if s == "" {
			return fmt.Errorf("questionId不能为空字符串")
		}
		v, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return fmt.Errorf("questionId必须是无符号整数: %w", err)
		}
		*u = FlexibleUint64(v)
		return nil
	}

	var v uint64
	if err := json.Unmarshal(data, &v); err == nil {
		*u = FlexibleUint64(v)
		return nil
	}

	return fmt.Errorf("questionId必须是无符号整数")
}

func flexibleUint64ToPtr(v *FlexibleUint64) *uint64 {
	if v == nil {
		return nil
	}
	id := uint64(*v)
	return &id
}

// ==================== 作业/考试功能 ====================

// CreateHomework 创建作业(教师)
func (h *Handler) CreateHomework(c *gin.Context) {
	logger := utils.GetLogger()

	// 读取请求体用于调试
	var bodyBytes []byte
	if c.Request.Body != nil {
		bodyBytes, _ = c.GetRawData()
		logger.Info("收到创建作业请求", zap.String("requestBody", string(bodyBytes)))
		// 重新设置请求体,以便后续读取
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	type HomeworkQuestionItem struct {
		QuestionID   *FlexibleUint64 `json:"questionId"`   // 题库题目ID(可选,支持 number/string)
		ProblemID    *string         `json:"problemId"`    // HOJ题目ID(可选,编程题使用,字符串类型支持"0001"等格式)
		QuestionType string          `json:"questionType"` // 题目类型(可选,用于设置默认分数)
		Score        int             `json:"score"`        // 分值(可选,默认根据题型设置)
	}

	var req struct {
		ClassroomID  uint64 `json:"classroomId" binding:"required"`
		Title        string `json:"title" binding:"required"`
		Description  string `json:"description"`
		StartTime    string `json:"startTime" binding:"required"` // RFC3339 format
		EndTime      string `json:"endTime" binding:"required"`   // RFC3339 format
		ShowScore    int    `json:"showScore"`
		ShowHomework int    `json:"showHomework"`
		ShowAnswer   int    `json:"showAnswer"`
		// 考试模式字段
		IsExamMode              int                    `json:"isExamMode"`
		ExamDuration            int                    `json:"examDuration"`
		AllowSubmitAfterMinutes int                    `json:"allowSubmitAfterMinutes"`
		DisableCopyPaste        int                    `json:"disableCopyPaste"`
		RequireFullscreen       int                    `json:"requireFullscreen"`
		DisallowTabSwitch       int                    `json:"disallowTabSwitch"`
		Questions               []HomeworkQuestionItem `json:"questions" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("请求参数错误", zap.Error(err), zap.String("detail", err.Error()))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误: "+err.Error()))
		return
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "startTime格式错误"))
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "endTime格式错误"))
		return
	}

	if endTime.Before(startTime) {
		c.JSON(http.StatusOK, errorResponse(400, "结束时间不能早于开始时间"))
		return
	}

	db := client.GetDB()

	// 创建作业
	homework := &model.ClassroomHomework{
		ClassroomID:             req.ClassroomID,
		Title:                   req.Title,
		Description:             req.Description,
		StartTime:               startTime,
		EndTime:                 endTime,
		ShowScore:               req.ShowScore,
		ShowHomework:            req.ShowHomework,
		ShowAnswer:              req.ShowAnswer,
		Status:                  1,
		IsExamMode:              req.IsExamMode,
		ExamDuration:            req.ExamDuration,
		AllowSubmitAfterMinutes: req.AllowSubmitAfterMinutes,
		DisableCopyPaste:        req.DisableCopyPaste,
		RequireFullscreen:       req.RequireFullscreen,
		DisallowTabSwitch:       req.DisallowTabSwitch,
	}

	if err := db.Create(homework).Error; err != nil {
		logger.Error("创建作业失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "创建失败"))
		return
	}

	// 添加作业题目
	for i, q := range req.Questions {
		score := q.Score
		if score == 0 {
			// 根据题型设置默认分数
			switch q.QuestionType {
			case "single_choice":
				score = 2 // 单选题默认2分
			case "multiple_choice":
				score = 5 // 多选题默认5分
			case "judge":
				score = 1 // 判断题默认1分
			case "subjective":
				score = 5 // 主观题默认5分
			case "programming":
				score = 20 // 编程题默认20分
			default:
				score = 2 // 默认2分
			}
		}

		// 验证至少提供一个ID
		if q.QuestionID == nil && q.ProblemID == nil {
			logger.Error("题目必须提供questionId或problemId")
			c.JSON(http.StatusOK, errorResponse(400, "题目必须提供questionId或problemId"))
			return
		}

		// 如果是编程题(有problemId),不添加到题库中,只使用ProblemID
		var questionIDPtr *uint64
		if q.ProblemID != nil {
			// 编程题: 不创建题库记录,QuestionID为nil
			questionIDPtr = nil
		} else {
			// 普通题目
			questionIDPtr = flexibleUint64ToPtr(q.QuestionID)
		}

		homeworkQuestion := &model.HomeworkQuestion{
			HomeworkID:    homework.ID,
			QuestionID:    questionIDPtr, // 使用指针,编程题时为nil
			ProblemID:     q.ProblemID,   // HOJ题目ID
			QuestionOrder: i + 1,
			Score:         score,
		}

		if err := db.Create(homeworkQuestion).Error; err != nil {
			logger.Error("添加作业题目失败", zap.Error(err))
			c.JSON(http.StatusOK, errorResponse(500, "添加题目失败"))
			return
		}
	}

	logger.Info("创建作业", zap.Uint64("id", homework.ID), zap.String("title", homework.Title))
	c.JSON(http.StatusOK, successResponse(homework))
}

// UpdateHomework 更新作业(教师)
func (h *Handler) UpdateHomework(c *gin.Context) {
	logger := utils.GetLogger()

	var req struct {
		ID           uint64 `json:"id" binding:"required"`
		ClassroomID  uint64 `json:"classroomId"`
		Title        string `json:"title" binding:"required"`
		Description  string `json:"description"`
		StartTime    string `json:"startTime" binding:"required"`
		EndTime      string `json:"endTime" binding:"required"`
		ShowHomework int    `json:"showHomework"`
		ShowScore    int    `json:"showScore"`
		ShowAnswer   int    `json:"showAnswer"`
		// 考试模式字段
		IsExamMode              int `json:"isExamMode"`
		ExamDuration            int `json:"examDuration"`
		AllowSubmitAfterMinutes int `json:"allowSubmitAfterMinutes"`
		DisableCopyPaste        int `json:"disableCopyPaste"`
		RequireFullscreen       int `json:"requireFullscreen"`
		DisallowTabSwitch       int `json:"disallowTabSwitch"`
		Questions               []struct {
			QuestionID   *FlexibleUint64 `json:"questionId"`   // 题库题目ID(可选,支持 number/string)
			ProblemID    *string         `json:"problemId"`    // HOJ题目ID(可选,编程题使用,字符串类型)
			QuestionType string          `json:"questionType"` // 题目类型(可选,用于设置默认分数)
			Score        int             `json:"score"`        // 分值(可选,默认根据题型设置)
		} `json:"questions" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	db := client.GetDB()

	// 检查作业是否存在
	var homework model.ClassroomHomework
	if err := db.Where("id = ?", req.ID).First(&homework).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, errorResponse(404, "作业不存在"))
		} else {
			logger.Error("查询作业失败", zap.Error(err))
			c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		}
		return
	}

	// 解析时间
	startTime, err := time.Parse("2006-01-02T15:04:05Z07:00", req.StartTime)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "开始时间格式错误"))
		return
	}
	endTime, err := time.Parse("2006-01-02T15:04:05Z07:00", req.EndTime)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "结束时间格式错误"))
		return
	}

	// 开始事务
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新作业基本信息
	homework.Title = req.Title
	homework.Description = req.Description
	homework.StartTime = startTime
	homework.EndTime = endTime
	homework.ShowHomework = req.ShowHomework
	homework.ShowScore = req.ShowScore
	homework.ShowAnswer = req.ShowAnswer
	homework.IsExamMode = req.IsExamMode
	homework.ExamDuration = req.ExamDuration
	homework.AllowSubmitAfterMinutes = req.AllowSubmitAfterMinutes
	homework.DisableCopyPaste = req.DisableCopyPaste
	homework.RequireFullscreen = req.RequireFullscreen
	homework.DisallowTabSwitch = req.DisallowTabSwitch

	// 更新状态
	now := time.Now()
	if now.Before(startTime) {
		homework.Status = 1 // 未开始
	} else if now.After(endTime) {
		homework.Status = 3 // 已结束
	} else {
		homework.Status = 2 // 进行中
	}

	if err := tx.Save(&homework).Error; err != nil {
		logger.Error("更新作业失败", zap.Error(err))
		tx.Rollback()
		c.JSON(http.StatusOK, errorResponse(500, "更新作业失败"))
		return
	}

	// 删除旧的作业题目关联
	if err := tx.Where("homework_id = ?", req.ID).Delete(&model.HomeworkQuestion{}).Error; err != nil {
		logger.Error("删除旧作业题目失败", zap.Error(err))
		tx.Rollback()
		c.JSON(http.StatusOK, errorResponse(500, "更新题目失败"))
		return
	}

	// 添加新的作业题目关联
	for i, q := range req.Questions {
		score := q.Score
		if score <= 0 {
			// 根据题型设置默认分数
			switch q.QuestionType {
			case "single_choice":
				score = 2 // 单选题默认2分
			case "multiple_choice":
				score = 5 // 多选题默认5分
			case "judge":
				score = 1 // 判断题默认1分
			case "subjective":
				score = 5 // 主观题默认5分
			case "programming":
				score = 20 // 编程题默认20分
			default:
				score = 2 // 默认2分
			}
		}

		// 验证至少提供一个ID
		if q.QuestionID == nil && q.ProblemID == nil {
			logger.Error("题目必须提供questionId或problemId")
			tx.Rollback()
			c.JSON(http.StatusOK, errorResponse(400, "题目必须提供questionId或problemId"))
			return
		}

		// 如果是编程题(有problemId),不添加到题库中,只使用ProblemID
		var questionIDPtr *uint64
		if q.ProblemID != nil {
			// 编程题: 不创建题库记录,QuestionID为nil
			questionIDPtr = nil
		} else {
			// 普通题目
			questionIDPtr = flexibleUint64ToPtr(q.QuestionID)
		}

		homeworkQuestion := &model.HomeworkQuestion{
			HomeworkID:    req.ID,
			QuestionID:    questionIDPtr, // 使用指针,编程题时为nil
			ProblemID:     q.ProblemID,   // HOJ题目ID
			QuestionOrder: i + 1,
			Score:         score,
		}

		if err := tx.Create(homeworkQuestion).Error; err != nil {
			logger.Error("添加作业题目失败", zap.Error(err))
			tx.Rollback()
			c.JSON(http.StatusOK, errorResponse(500, "添加题目失败"))
			return
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		logger.Error("提交事务失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "更新失败"))
		return
	}

	logger.Info("更新作业", zap.Uint64("id", homework.ID), zap.String("title", homework.Title))
	c.JSON(http.StatusOK, successResponse(homework))
}

// GetHomeworkList 获取作业列表（教师/学生）
func (h *Handler) GetHomeworkList(c *gin.Context) {
	logger := utils.GetLogger()
	classroomIDStr := c.Param("classroomId")
	classroomID, err := strconv.ParseUint(classroomIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "classroomId参数格式错误"))
		return
	}

	db := client.GetDB()
	var homeworks []model.ClassroomHomework

	if err := db.Where("classroom_id = ?", classroomID).
		Order("create_time DESC").
		Find(&homeworks).Error; err != nil {
		logger.Error("查询作业列表失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	// 获取当前用户ID（如果有）
	uid, hasUser := c.Get("userId")

	// 更新作业状态，并添加学生提交状态
	now := time.Now()
	for i := range homeworks {
		if now.Before(homeworks[i].StartTime) {
			homeworks[i].Status = 1 // 未开始
		} else if now.After(homeworks[i].EndTime) {
			homeworks[i].Status = 3 // 已结束
		} else {
			homeworks[i].Status = 2 // 进行中
		}

		// 如果是学生请求，检查是否已提交
		if hasUser {
			// 检查是否有正式提交记录（is_officially_submitted = 1）
			// 只要有一条正式提交记录，就认为已完成
			var submitCount int64
			db.Model(&model.HomeworkSubmit{}).
				Where("homework_id = ? AND uid = ? AND is_officially_submitted = 1", homeworks[i].ID, uid.(string)).
				Count(&submitCount)

			// 判断是否已完成（有正式提交记录）
			homeworks[i].IsCompleted = submitCount > 0
		}
	}

	c.JSON(http.StatusOK, successResponse(homeworks))
}

// GetHomeworkDetail 获取作业详情
func (h *Handler) GetHomeworkDetail(c *gin.Context) {
	logger := utils.GetLogger()
	homeworkIDStr := c.Param("homeworkId")
	homeworkID, err := strconv.ParseUint(homeworkIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "homeworkId参数格式错误"))
		return
	}

	// 获取当前用户ID
	uid, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "未登录"))
		return
	}

	db := client.GetDB()
	var homework model.ClassroomHomework

	if err := db.Where("id = ?", homeworkID).
		Preload("Questions.Question").
		First(&homework).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, errorResponse(404, "作业不存在"))
		} else {
			logger.Error("查询作业详情失败", zap.Error(err))
			c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		}
		return
	}

	// 手动排序Questions
	if len(homework.Questions) > 0 {
		sort.Slice(homework.Questions, func(i, j int) bool {
			return homework.Questions[i].QuestionOrder < homework.Questions[j].QuestionOrder
		})
	}

	// 更新作业状态
	now := time.Now()
	if now.Before(homework.StartTime) {
		homework.Status = 1 // 未开始
	} else if now.After(homework.EndTime) {
		homework.Status = 3 // 已结束
	} else {
		homework.Status = 2 // 进行中
	}

	// 安全检查：判断用户角色和权限
	var classroom model.Classroom
	isTeacher := false
	isAdmin := false
	isStudentSubmitted := false

	if err := db.Where("id = ?", homework.ClassroomID).First(&classroom).Error; err == nil {
		// 检查是否为教师
		isTeacher = classroom.TeacherID == uid.(string)

		// 检查是否为班级管理员（从user_roles表查询）
		var userRole model.ClassroomUserRole
		err := db.Where("uid = ? AND role = ?", uid.(string), "admin").First(&userRole).Error
		if err == nil {
			isAdmin = true
		} else if err != gorm.ErrRecordNotFound {
			// 只有非"记录不存在"的错误才需要记录日志
			logger.Warn("查询班级管理员角色失败", zap.Error(err))
		}
	}

	// 检查学生是否已正式提交作业
	var submissions []model.HomeworkSubmit
	if err := db.Where("homework_id = ? AND uid = ? AND is_officially_submitted = ?",
		homeworkID, uid.(string), 1).Find(&submissions).Error; err == nil && len(submissions) > 0 {
		isStudentSubmitted = true
	}

	// 判断是否应该显示答案
	shouldShowAnswer := false
	if isTeacher || isAdmin {
		// 教师和管理员始终可以看到答案
		shouldShowAnswer = true
	} else if isStudentSubmitted && homework.ShowAnswer == 1 {
		// 学生已提交且教师允许查看答案
		shouldShowAnswer = true
	}

	// 判断是否应该显示难度
	// 规则：只有教师和管理员可以看到难度，学生看不到（即使已提交）
	shouldShowDifficulty := isTeacher || isAdmin

	// 根据权限过滤敏感字段
	for i := range homework.Questions {
		if homework.Questions[i].Question != nil {
			// 如果不应显示答案，清空答案字段
			if !shouldShowAnswer {
				homework.Questions[i].Question.Answer = ""
			}
			// 如果不应显示难度，清零难度字段
			if !shouldShowDifficulty {
				homework.Questions[i].Question.Difficulty = 0
			}
		}
	}

	c.JSON(http.StatusOK, successResponse(homework))
}

// SaveHomeworkDraft 保存作业草稿（学生）- 自动保存
// 与 SubmitHomework 的区别：不设置 IsOfficiallySubmitted 标志
// 请求格式：{ homeworkId, answers: { questionId: answer, ... } }
func (h *Handler) SaveHomeworkDraft(c *gin.Context) {
	logger := utils.GetLogger()

	var req struct {
		HomeworkID  uint64            `json:"homeworkId" binding:"required"`
		Answers     map[string]string `json:"answers" binding:"required"` // questionId -> answer
		Attachments map[string]string `json:"attachments"`                // questionId -> attachment URLs (comma separated)
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	// 调试日志：打印收到的附件数据
	logger.Info("保存草稿 - 收到附件数据",
		zap.Int64("homeworkId", int64(req.HomeworkID)),
		zap.Any("attachments", req.Attachments))

	// 获取当前用户ID
	uid, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "未登录"))
		return
	}

	db := client.GetDB()

	// 检查作业是否存在
	var homework model.ClassroomHomework
	if err := db.Where("id = ?", req.HomeworkID).First(&homework).Error; err != nil {
		c.JSON(http.StatusOK, errorResponse(404, "作业不存在"))
		return
	}

	// 开始事务
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 逐个处理每个题目的答案（草稿模式，不判分）
	for questionIDStr, answer := range req.Answers {
		questionID, err := strconv.ParseUint(questionIDStr, 10, 64)
		if err != nil {
			logger.Warn("题目ID格式错误", zap.String("question_id", questionIDStr))
			continue
		}

		// 查找题目
		var question model.QuestionBank
		if err := tx.Where("id = ?", questionID).First(&question).Error; err != nil {
			logger.Warn("题目不存在", zap.Uint64("question_id", questionID))
			continue
		}

		normalizedAnswer, normErr := normalizeStudentAnswerForStorage(&question, answer)
		if normErr != nil {
			logger.Warn("草稿答案格式错误",
				zap.Uint64("question_id", questionID),
				zap.String("question_type", question.Type),
				zap.String("raw_answer", answer),
				zap.Error(normErr))
			tx.Rollback()
			c.JSON(http.StatusOK, errorResponse(400, "答案格式错误"))
			return
		}

		// 获取附件URL（如果有）
		attachment := ""
		if req.Attachments != nil {
			if att, ok := req.Attachments[questionIDStr]; ok {
				attachment = att
			}
		}

		// 调试日志：打印每个题目的附件信息
		logger.Info("保存草稿 - 题目附件",
			zap.String("questionId", questionIDStr),
			zap.String("attachment", attachment))

		// 查找是否已有草稿记录
		var existingSubmit model.HomeworkSubmit
		checkErr := tx.Where("homework_id = ? AND question_id = ? AND uid = ?",
			req.HomeworkID, questionID, uid.(string)).
			First(&existingSubmit).Error

		if checkErr == nil {
			// 更新已有记录（只更新需要的字段，避免更新 ExamStartTime 等字段）
			updates := map[string]interface{}{
				"answer": normalizedAnswer,
			}
			// 只有当 req.Attachments 不为 nil 且当前题目有附件数据时，才更新 attachment
			// 避免自动保存草稿时清空已上传的图片
			if req.Attachments != nil {
				updates["attachment"] = attachment
			}
			// 草稿保存不修改分数、提交状态和考试相关字段

			if err := tx.Model(&existingSubmit).Updates(updates).Error; err != nil {
				logger.Error("更新作业草稿失败", zap.Error(err))
				tx.Rollback()
				c.JSON(http.StatusOK, errorResponse(500, "保存草稿失败"))
				return
			}
		} else {
			// 创建新记录（草稿状态）
			submit := &model.HomeworkSubmit{
				HomeworkID:            req.HomeworkID,
				QuestionID:            &questionID, // 使用指针
				UID:                   uid.(string),
				Answer:                normalizedAnswer,
				Attachment:            attachment,
				Score:                 0, // 草稿不判分
				IsScored:              0,
				IsOfficiallySubmitted: 0, // 草稿状态
			}

			if err := tx.Create(submit).Error; err != nil {
				logger.Error("保存作业草稿失败", zap.Error(err))
				tx.Rollback()
				c.JSON(http.StatusOK, errorResponse(500, "保存草稿失败"))
				return
			}
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		logger.Error("提交事务失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "保存草稿失败"))
		return
	}

	logger.Info("保存作业草稿", zap.Uint64("homework_id", req.HomeworkID), zap.String("uid", uid.(string)))
	c.JSON(http.StatusOK, successResponse(map[string]interface{}{
		"message": "草稿保存成功",
	}))
}

// SubmitHomework 提交作业（学生）- 批量提交模式
// 请求格式：{ homeworkId, answers: { questionId: answer, ... }, attachments: { questionId: urls, ... } }
func (h *Handler) SubmitHomework(c *gin.Context) {
	logger := utils.GetLogger()

	var req struct {
		HomeworkID  uint64            `json:"homeworkId" binding:"required"`
		Answers     map[string]string `json:"answers" binding:"required"` // questionId -> answer
		Attachments map[string]string `json:"attachments"`                // questionId -> attachment URLs (comma separated)
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	// 调试日志：打印收到的附件数据
	logger.Info("提交作业 - 收到附件数据",
		zap.Int64("homeworkId", int64(req.HomeworkID)),
		zap.Any("attachments", req.Attachments))

	// 获取当前用户ID
	uid, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "未登录"))
		return
	}

	db := client.GetDB()

	// 检查作业时间
	var homework model.ClassroomHomework
	if err := db.Where("id = ?", req.HomeworkID).First(&homework).Error; err != nil {
		c.JSON(http.StatusOK, errorResponse(404, "作业不存在"))
		return
	}

	now := time.Now()
	if now.Before(homework.StartTime) {
		c.JSON(http.StatusOK, errorResponse(400, "作业未开始"))
		return
	}
	if now.After(homework.EndTime) {
		c.JSON(http.StatusOK, errorResponse(400, "作业已结束，无法提交"))
		return
	}

	// 开始事务
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	results := make([]map[string]interface{}, 0)

	// 逐个处理每个题目的答案
	for questionIDStr, answer := range req.Answers {
		questionID, err := strconv.ParseUint(questionIDStr, 10, 64)
		if err != nil {
			logger.Warn("题目ID格式错误", zap.String("question_id", questionIDStr))
			continue
		}

		// 查找题目
		var question model.QuestionBank
		if err := tx.Where("id = ?", questionID).First(&question).Error; err != nil {
			logger.Warn("题目不存在", zap.Uint64("question_id", questionID))
			continue
		}

		normalizedAnswer, normErr := normalizeStudentAnswerForStorage(&question, answer)
		if normErr != nil {
			logger.Warn("提交答案格式错误",
				zap.Uint64("question_id", questionID),
				zap.String("question_type", question.Type),
				zap.String("raw_answer", answer),
				zap.Error(normErr))
			tx.Rollback()
			c.JSON(http.StatusOK, errorResponse(400, "答案格式错误"))
			return
		}
		answer = normalizedAnswer

		// 获取附件URL（如果有）
		attachment := ""
		if req.Attachments != nil {
			if att, ok := req.Attachments[questionIDStr]; ok {
				attachment = att
			}
		}

		// 调试日志：打印每个题目的附件信息
		logger.Info("提交作业 - 题目附件",
			zap.String("questionId", questionIDStr),
			zap.String("attachment", attachment))

		// 计算分数（单选、多选、判断题自动判分）
		score := 0.0
		isScored := 0
		if question.Type == "single_choice" || question.Type == "multiple_choice" || question.Type == "judge" {
			// 对于单选题，直接比较字符串
			if question.Type == "single_choice" {
				if answer == question.Answer {
					// 获取该题在作业中的分值
					var homeworkQuestion model.HomeworkQuestion
					if err := tx.Where("homework_id = ? AND question_id = ?",
						req.HomeworkID, questionID).First(&homeworkQuestion).Error; err == nil {
						score = float64(homeworkQuestion.Score)
					}
				}
			} else if question.Type == "judge" {
				// 对于判断题，使用严格规范化比较（仅 true/false）
				if compareJudgeAnswers(answer, question.Answer) {
					// 获取该题在作业中的分值
					var homeworkQuestion model.HomeworkQuestion
					if err := tx.Where("homework_id = ? AND question_id = ?",
						req.HomeworkID, questionID).First(&homeworkQuestion).Error; err == nil {
						score = float64(homeworkQuestion.Score)
					}
				}
			} else if question.Type == "multiple_choice" {
				// 对于多选题，需要比较JSON数组或逗号分隔字符串
				var studentAnswers, correctAnswers []string

				// 解析学生答案（新格式：JSON数组）
				if err := json.Unmarshal([]byte(answer), &studentAnswers); err != nil {
					// 兼容旧格式：逗号分隔字符串
					studentAnswers = strings.Split(answer, ",")
					// 去除空格
					for i := range studentAnswers {
						studentAnswers[i] = strings.TrimSpace(studentAnswers[i])
					}
				}

				// 解析正确答案（新格式：JSON数组）
				if err := json.Unmarshal([]byte(question.Answer), &correctAnswers); err != nil {
					// 兼容旧格式：逗号分隔字符串
					correctAnswers = strings.Split(question.Answer, ",")
					// 去除空格
					for i := range correctAnswers {
						correctAnswers[i] = strings.TrimSpace(correctAnswers[i])
					}
				}

				// 记录调试信息
				logger.Info("多选题评分", zap.String("question_id", strconv.FormatUint(questionID, 10)),
					zap.String("student_answer", answer), zap.String("correct_answer", question.Answer),
					zap.Any("parsed_student", studentAnswers), zap.Any("parsed_correct", correctAnswers),
					zap.Bool("match", compareArrays(studentAnswers, correctAnswers)))

				// 比较答案
				if compareArrays(studentAnswers, correctAnswers) {
					var homeworkQuestion model.HomeworkQuestion
					if err := tx.Where("homework_id = ? AND question_id = ?",
						req.HomeworkID, questionID).First(&homeworkQuestion).Error; err == nil {
						score = float64(homeworkQuestion.Score)
					}
				}
			}
			isScored = 1
		}

		// 查找是否已提交
		var existingSubmit model.HomeworkSubmit
		checkErr := tx.Where("homework_id = ? AND question_id = ? AND uid = ?",
			req.HomeworkID, questionID, uid.(string)).
			First(&existingSubmit).Error

		if checkErr == nil {
			// 更新已有提交
			existingSubmit.Answer = answer
			existingSubmit.Attachment = attachment
			existingSubmit.Score = score
			existingSubmit.IsScored = isScored
			existingSubmit.IsOfficiallySubmitted = 1 // 标记为正式提交

			// 如果是考试模式，记录考试结束时间
			if homework.IsExamMode == 1 && existingSubmit.ExamEndTime == nil {
				existingSubmit.ExamEndTime = &now
			}

			if err := tx.Save(&existingSubmit).Error; err != nil {
				logger.Error("更新作业提交失败", zap.Error(err))
				tx.Rollback()
				c.JSON(http.StatusOK, errorResponse(500, "提交失败"))
				return
			}
		} else {
			// 创建新提交
			submit := &model.HomeworkSubmit{
				HomeworkID:            req.HomeworkID,
				QuestionID:            &questionID, // 使用指针
				UID:                   uid.(string),
				Answer:                answer,
				Attachment:            attachment,
				Score:                 score,
				IsScored:              isScored,
				IsOfficiallySubmitted: 1, // 标记为正式提交
			}

			// 如果是考试模式，记录考试结束时间
			if homework.IsExamMode == 1 {
				submit.ExamEndTime = &now
			}

			if err := tx.Create(submit).Error; err != nil {
				logger.Error("提交作业失败", zap.Error(err))
				tx.Rollback()
				c.JSON(http.StatusOK, errorResponse(500, "提交失败"))
				return
			}
		}

		results = append(results, map[string]interface{}{
			"questionId": questionID,
			"score":      score,
			"isScored":   isScored,
		})
	}

	// 对于考试模式，需要确保所有客观题都被标记为已评分（即使未作答）
	// 因为学生可能只提交了部分题目，但所有客观题都应该能自动评分
	if homework.IsExamMode == 1 {
		// 查询作业的所有题目
		var homeworkQuestions []model.HomeworkQuestion
		if err := tx.Where("homework_id = ?", req.HomeworkID).Find(&homeworkQuestions).Error; err != nil {
			logger.Warn("查询作业题目失败，跳过自动评分", zap.Error(err))
		} else {
			// 为每个客观题检查并更新未评分的提交记录
			for _, hq := range homeworkQuestions {
				// 只处理普通题目（非编程题）
				if hq.QuestionID == nil {
					continue
				}

				// 查询该题目是否已提交且已评分
				var existingSubmit model.HomeworkSubmit
				checkErr := tx.Where("homework_id = ? AND question_id = ? AND uid = ?",
					req.HomeworkID, hq.QuestionID, uid.(string)).
					First(&existingSubmit).Error

				if checkErr == nil && existingSubmit.IsScored == 0 {
					// 找到了提交记录但未评分，需要检查题目类型
					var question model.QuestionBank
					if err := tx.Where("id = ?", *hq.QuestionID).First(&question).Error; err == nil {
						// 是客观题，自动评分（即使是空答案也算已评分）
						if question.Type == "single_choice" || question.Type == "multiple_choice" || question.Type == "judge" {
							// 计算分数
							score := 0.0
							if question.Type == "single_choice" && existingSubmit.Answer == question.Answer {
								score = float64(hq.Score)
							} else if question.Type == "judge" && compareJudgeAnswers(existingSubmit.Answer, question.Answer) {
								score = float64(hq.Score)
							} else if question.Type == "multiple_choice" {
								// 多选题比较
								var studentAnswers, correctAnswers []string
								if err := json.Unmarshal([]byte(existingSubmit.Answer), &studentAnswers); err != nil {
									studentAnswers = strings.Split(existingSubmit.Answer, ",")
									for i := range studentAnswers {
										studentAnswers[i] = strings.TrimSpace(studentAnswers[i])
									}
								}
								if err := json.Unmarshal([]byte(question.Answer), &correctAnswers); err != nil {
									correctAnswers = strings.Split(question.Answer, ",")
									for i := range correctAnswers {
										correctAnswers[i] = strings.TrimSpace(correctAnswers[i])
									}
								}
								if compareArrays(studentAnswers, correctAnswers) {
									score = float64(hq.Score)
								}
							}

							// 更新为已评分
							if err := tx.Model(&existingSubmit).
								Updates(map[string]interface{}{
									"score":     score,
									"is_scored": 1,
								}).Error; err != nil {
								logger.Error("更新客观题评分失败",
									zap.Error(err),
									zap.Uint64("questionId", *hq.QuestionID),
									zap.String("uid", uid.(string)))
							} else {
								logger.Info("自动评分未作答客观题",
									zap.Uint64("questionId", *hq.QuestionID),
									zap.String("uid", uid.(string)),
									zap.String("questionType", question.Type),
									zap.Float64("score", score))
							}
						}
					}
				}
			}
		}
	}

	// 将该用户的所有编程题记录也标记为正式提交
	// 因为编程题是通过 SaveProgrammingSubmission 单独保存的
	updates := map[string]interface{}{
		"is_officially_submitted": 1,
	}
	// 如果是考试模式，同时记录考试结束时间
	if homework.IsExamMode == 1 {
		updates["exam_end_time"] = now
	}
	if err := tx.Model(&model.HomeworkSubmit{}).
		Where("homework_id = ? AND uid = ? AND problem_id IS NOT NULL AND is_officially_submitted = 0",
			req.HomeworkID, uid.(string)).
		Updates(updates).Error; err != nil {
		logger.Error("更新编程题正式提交状态失败", zap.Error(err))
		tx.Rollback()
		c.JSON(http.StatusOK, errorResponse(500, "提交失败"))
		return
	}

	logger.Info("批量提交作业（包括编程题）",
		zap.Uint64("homework_id", req.HomeworkID),
		zap.String("uid", uid.(string)))

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		logger.Error("提交事务失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "提交失败"))
		return
	}

	logger.Info("批量提交作业", zap.Uint64("homework_id", req.HomeworkID), zap.String("uid", uid.(string)))
	c.JSON(http.StatusOK, successResponse(map[string]interface{}{
		"results": results,
		"count":   len(results),
	}))
}

// GetHomeworkSubmissions 获取作业提交情况（教师）
func (h *Handler) GetHomeworkSubmissions(c *gin.Context) {
	logger := utils.GetLogger()
	homeworkIDStr := c.Param("homeworkId")
	homeworkID, err := strconv.ParseUint(homeworkIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "homeworkId参数格式错误"))
		return
	}

	db := client.GetDB()

	// 首先查询作业的所有题目（按 question_order 排序）
	var homeworkQuestions []model.HomeworkQuestion
	if err := db.Where("homework_id = ?", homeworkID).
		Order("question_order ASC").
		Find(&homeworkQuestions).Error; err != nil {
		logger.Error("查询作业题目失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	// 构建题目顺序映射：questionId -> order，problemId -> order
	questionOrderMap := make(map[uint64]int)
	problemOrderMap := make(map[string]int)
	for idx, hq := range homeworkQuestions {
		if hq.QuestionID != nil {
			questionOrderMap[*hq.QuestionID] = idx
		}
		if hq.ProblemID != nil {
			problemOrderMap[*hq.ProblemID] = idx
		}
	}

	var submissions []model.HomeworkSubmit

	// 查询所有提交记录（包括普通题目和编程题）
	// 只显示已正式提交的记录（is_officially_submitted = 1）
	if err := db.Where("homework_id = ? AND is_officially_submitted = 1", homeworkID).
		Preload("Student").
		Preload("Question").
		Find(&submissions).Error; err != nil {
		logger.Error("查询作业提交失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	// 按照 question_order 重新排序 submissions
	sort.Slice(submissions, func(i, j int) bool {
		// 获取两个提交记录的顺序
		var orderI, orderJ int

		// 对于普通题目
		if submissions[i].QuestionID != nil {
			orderI = questionOrderMap[*submissions[i].QuestionID]
		} else if submissions[i].ProblemID != nil {
			orderI = problemOrderMap[*submissions[i].ProblemID]
		}

		if submissions[j].QuestionID != nil {
			orderJ = questionOrderMap[*submissions[j].QuestionID]
		} else if submissions[j].ProblemID != nil {
			orderJ = problemOrderMap[*submissions[j].ProblemID]
		}

		// 先按学生ID分组
		if submissions[i].UID != submissions[j].UID {
			return submissions[i].UID < submissions[j].UID
		}
		// 同一个学生内，按题目顺序排序
		return orderI < orderJ
	})

	// 获取作业所属的班级ID
	var homework model.ClassroomHomework
	if err := db.Where("id = ?", homeworkID).First(&homework).Error; err != nil {
		logger.Error("查询作业失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	// 获取班级的所有学生（用于获取真实姓名）
	var classroomStudents []model.ClassroomStudent
	if err := db.Where("classroom_id = ? AND status = 1", homework.ClassroomID).
		Find(&classroomStudents).Error; err != nil {
		logger.Error("查询班级学生失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	// 构建 uid -> realName 的映射
	studentRealNameMap := make(map[string]string)
	for _, cs := range classroomStudents {
		studentRealNameMap[cs.UID] = cs.RealName
	}

	// 构建 problemId -> homeworkQuestionId 的映射
	problemToQuestionMap := make(map[string]uint64)
	for _, hq := range homeworkQuestions {
		if hq.ProblemID != nil && *hq.ProblemID != "" {
			problemToQuestionMap[*hq.ProblemID] = hq.ID
		}
	}

	// 构建返回结果，添加 homeworkQuestionId 和 realName 字段
	type SubmissionWithHomeworkQuestionID struct {
		model.HomeworkSubmit
		HomeworkQuestionID *uint64 `json:"homeworkQuestionId,omitempty"` // 作业题目关联表ID（编程题评分时使用）
		RealName           string  `json:"realName,omitempty"`           // 班级学生真实姓名
	}

	result := make([]SubmissionWithHomeworkQuestionID, 0, len(submissions))
	for _, submit := range submissions {
		// 调试日志：检查attachment字段
		if submit.Attachment != "" {
			logger.Info("提交记录包含附件",
				zap.Uint64("question_id", *submit.QuestionID),
				zap.String("attachment", submit.Attachment))
		}

		enhancedSubmit := SubmissionWithHomeworkQuestionID{
			HomeworkSubmit: submit,
			RealName:       studentRealNameMap[submit.UID], // 添加班级学生真实姓名
		}

		// 对于编程题，添加 homeworkQuestionId
		if submit.ProblemID != nil && *submit.ProblemID != "" {
			if hqID, exists := problemToQuestionMap[*submit.ProblemID]; exists {
				enhancedSubmit.HomeworkQuestionID = &hqID
			}
		} else {
			// 对于普通题目，homeworkQuestionId 就是 questionId
			enhancedSubmit.HomeworkQuestionID = submit.QuestionID
		}

		result = append(result, enhancedSubmit)
	}

	logger.Info("查询作业提交成功",
		zap.Uint64("homework_id", homeworkID),
		zap.Int("count", len(result)))

	c.JSON(http.StatusOK, successResponse(result))
}

// GetStudentHomeworkStatus 获取学生作业完成情况
func (h *Handler) GetStudentHomeworkStatus(c *gin.Context) {
	logger := utils.GetLogger()
	homeworkIDStr := c.Param("homeworkId")
	homeworkID, err := strconv.ParseUint(homeworkIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "homeworkId参数格式错误"))
		return
	}

	db := client.GetDB()

	// 获取作业的所有题目
	var questions []model.HomeworkQuestion
	if err := db.Where("homework_id = ?", homeworkID).Find(&questions).Error; err != nil {
		logger.Error("查询作业题目失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	// 获取班级的所有学生
	var homework model.ClassroomHomework
	if err := db.Where("id = ?", homeworkID).First(&homework).Error; err != nil {
		c.JSON(http.StatusOK, errorResponse(404, "作业不存在"))
		return
	}

	var students []model.ClassroomStudent
	if err := db.Where("classroom_id = ? AND status = 1", homework.ClassroomID).
		Preload("User").
		Find(&students).Error; err != nil {
		logger.Error("查询学生列表失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	// 构建结果
	type StudentStatus struct {
		Student   *model.UserInfo `json:"student"`
		Completed int             `json:"completed"`
		Total     int             `json:"total"`
		Score     float64         `json:"score"`
	}

	result := make([]StudentStatus, 0, len(students))
	for _, student := range students {
		var totalScore float64
		completed := 0

		for _, question := range questions {
			var submit model.HomeworkSubmit

			// 根据题目类型查询不同的字段
			var err error
			if question.ProblemID != nil && *question.ProblemID != "" {
				// 编程题：使用 problem_id 查询
				err = db.Where("homework_id = ? AND problem_id = ? AND uid = ?",
					homeworkID, question.ProblemID, student.UID).
					First(&submit).Error
			} else {
				// 普通题目：使用 question_id 查询
				err = db.Where("homework_id = ? AND question_id = ? AND uid = ?",
					homeworkID, question.QuestionID, student.UID).
					First(&submit).Error
			}

			if err == nil {
				completed++
				totalScore += submit.Score
			}
		}

		result = append(result, StudentStatus{
			Student:   student.User,
			Completed: completed,
			Total:     len(questions),
			Score:     totalScore,
		})
	}

	c.JSON(http.StatusOK, successResponse(result))
}

// GetStudentHomeworkDetail 获取学生作业详情（学生）
func (h *Handler) GetStudentHomeworkDetail(c *gin.Context) {
	logger := utils.GetLogger()
	homeworkIDStr := c.Param("homeworkId")
	homeworkID, err := strconv.ParseUint(homeworkIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "homeworkId参数格式错误"))
		return
	}

	// 获取当前用户ID
	uid, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "未登录"))
		return
	}

	db := client.GetDB()

	// 获取该学生在这个作业中的所有提交记录
	var submissions []model.HomeworkSubmit
	if err := db.Where("homework_id = ? AND uid = ?", homeworkID, uid.(string)).
		Find(&submissions).Error; err != nil {
		logger.Error("查询学生作业提交记录失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	// 如果没有提交记录，返回空数据
	if len(submissions) == 0 {
		c.JSON(http.StatusOK, successResponse(nil))
		return
	}

	// 构建答案数据 map[questionID]answer
	answersMap := make(map[string]string)
	// 构建每题得分数据 map[questionID]score
	scoresMap := make(map[string]float64)
	// 构建每题评分状态数据 map[questionID]isScored
	isScoredMap := make(map[string]bool)
	// 构建每题附件数据 map[questionID]attachment
	attachmentsMap := make(map[string]string)
	var submitTime *time.Time
	totalScore := 0.0
	gradedScore := 0.0 // 已评分题目的得分
	// 检查是否已正式提交（只要有一条记录标记为已正式提交，就认为整个作业已提交）
	isOfficiallySubmitted := false
	hasUngraded := false // 是否有未评分的题目

	for _, submission := range submissions {
		// 将题目ID转为字符串作为key
		var questionIDStr string
		if submission.QuestionID != nil {
			questionIDStr = strconv.FormatUint(*submission.QuestionID, 10)
		} else if submission.ProblemID != nil {
			// 编程题使用 ProblemID 作为 key
			questionIDStr = *submission.ProblemID
		} else {
			continue // 跳过无效记录
		}
		answersMap[questionIDStr] = submission.Answer
		scoresMap[questionIDStr] = submission.Score
		isScoredMap[questionIDStr] = submission.IsScored == 1
		// 添加附件数据
		attachmentsMap[questionIDStr] = submission.Attachment

		// 累加所有题目的得分（包括0分）
		totalScore += submission.Score

		// 如果已评分，累加到已评分分数
		if submission.IsScored == 1 {
			gradedScore += submission.Score
		} else {
			hasUngraded = true
		}

		// 记录最新的提交时间
		if submitTime == nil || submission.UpdatedAt.After(*submitTime) {
			submitTime = &submission.UpdatedAt
		}

		// 检查是否已正式提交
		if submission.IsOfficiallySubmitted == 1 {
			isOfficiallySubmitted = true
		}
	}

	// 将答案map转为JSON字符串
	answersJSON, err := json.Marshal(answersMap)
	if err != nil {
		logger.Error("序列化答案失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "序列化答案失败"))
		return
	}

	// 将每题得分map转为JSON字符串
	scoresJSON, err := json.Marshal(scoresMap)
	if err != nil {
		logger.Error("序列化分数失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "序列化分数失败"))
		return
	}

	// 将每题评分状态map转为JSON字符串
	isScoredJSON, err := json.Marshal(isScoredMap)
	if err != nil {
		logger.Error("序列化评分状态失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "序列化评分状态失败"))
		return
	}

	// 将每题附件map转为JSON字符串
	attachmentsJSON, err := json.Marshal(attachmentsMap)
	if err != nil {
		logger.Error("序列化附件失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "序列化附件失败"))
		return
	}

	result := map[string]interface{}{
		"answers":               string(answersJSON),
		"scores":                string(scoresJSON),      // 每题得分
		"isScoredMap":           string(isScoredJSON),    // 每题评分状态
		"attachments":           string(attachmentsJSON), // 每题附件
		"submitTime":            submitTime,
		"isOfficiallySubmitted": isOfficiallySubmitted,
		"hasUngraded":           hasUngraded, // 是否有未评分的题目
	}

	// 只有已正式提交才返回分数（客观题会自动评分，主观题需要教师手动评分）
	if isOfficiallySubmitted {
		result["score"] = totalScore
		result["gradedScore"] = gradedScore // 已评分题目的得分
	}

	c.JSON(http.StatusOK, successResponse(result))
}

// ==================== 资料库功能 ====================

// CreateFolder 创建文件夹（教师）
func (h *Handler) CreateFolder(c *gin.Context) {
	var req struct {
		ClassroomID uint64 `json:"classroomId" binding:"required"`
		FolderName  string `json:"folderName" binding:"required"`
		ParentID    uint64 `json:"parentId"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	// 获取当前用户ID
	creatorID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "未登录"))
		return
	}

	db := client.GetDB()

	folder := &model.ClassroomFolder{
		ClassroomID: req.ClassroomID,
		FolderName:  req.FolderName,
		ParentID:    req.ParentID,
		CreatorID:   creatorID.(string),
		Status:      1,
	}

	if err := db.Create(folder).Error; err != nil {
		c.JSON(http.StatusOK, errorResponse(500, "创建失败"))
		return
	}

	c.JSON(http.StatusOK, successResponse(folder))
}

// GetFolders 获取文件夹列表（教师/学生）
func (h *Handler) GetFolders(c *gin.Context) {
	classroomIDStr := c.Param("classroomId")
	classroomID, err := strconv.ParseUint(classroomIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "classroomId参数格式错误"))
		return
	}

	parentIDStr := c.Query("parentId")
	parentID, _ := strconv.ParseUint(parentIDStr, 10, 64)

	db := client.GetDB()
	var folders []model.ClassroomFolder

	if err := db.Where("classroom_id = ? AND parent_id = ? AND status = 1", classroomID, parentID).
		Order("sort_order ASC, create_time ASC").
		Find(&folders).Error; err != nil {
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	c.JSON(http.StatusOK, successResponse(folders))
}

// DeleteFolder 删除文件夹（教师）
func (h *Handler) DeleteFolder(c *gin.Context) {
	folderIDStr := c.Param("folderId")
	folderID, err := strconv.ParseUint(folderIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "folderId参数格式错误"))
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

	// 1. 递归删除所有子文件夹
	var allFolderIDs []uint64
	allFolderIDs = append(allFolderIDs, folderID)

	// 查找所有子文件夹ID（包括嵌套的子文件夹）
	var subfolders []model.ClassroomFolder
	if err := tx.Where("parent_id = ? AND status = 1", folderID).Find(&subfolders).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusOK, errorResponse(500, "删除失败"))
		return
	}

	// 递归收集所有子文件夹ID
	for _, subfolder := range subfolders {
		allFolderIDs = append(allFolderIDs, subfolder.ID)
		// TODO: 这里可以进一步递归查找多层嵌套的子文件夹
	}

	// 2. 删除所有文件夹下的资料文件
	var materials []model.ClassroomMaterial
	if err := tx.Where("folder_id IN ? AND status = 1", allFolderIDs).Find(&materials).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusOK, errorResponse(500, "删除失败"))
		return
	}

	// 删除磁盘上的文件
	for _, material := range materials {
		if material.FilePath != "" {
			// 转换URL路径为文件系统路径
			filePath := "." + material.FilePath
			os.Remove(filePath) // 忽略错误
		}
	}

	// 3. 软删除所有资料记录
	if err := tx.Model(&model.ClassroomMaterial{}).
		Where("folder_id IN ?", allFolderIDs).
		Update("status", 0).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusOK, errorResponse(500, "删除失败"))
		return
	}

	// 4. 软删除所有子文件夹
	if err := tx.Model(&model.ClassroomFolder{}).
		Where("id IN ?", allFolderIDs).
		Update("status", 0).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusOK, errorResponse(500, "删除失败"))
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusOK, errorResponse(500, "删除失败"))
		return
	}

	c.JSON(http.StatusOK, successResponse(nil))
}

// UpdateFolder 更新文件夹名称（教师）
func (h *Handler) UpdateFolder(c *gin.Context) {
	var req struct {
		ID         uint64 `json:"id" binding:"required"`
		FolderName string `json:"folderName" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	db := client.GetDB()

	// 检查文件夹是否存在
	var folder model.ClassroomFolder
	if err := db.Where("id = ? AND status = 1", req.ID).First(&folder).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, errorResponse(404, "文件夹不存在"))
		} else {
			c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		}
		return
	}

	// 更新文件夹名称
	if err := db.Model(&folder).Update("folder_name", req.FolderName).Error; err != nil {
		c.JSON(http.StatusOK, errorResponse(500, "更新失败"))
		return
	}

	c.JSON(http.StatusOK, successResponse(nil))
}

// UploadMaterial 上传资料（教师）
func (h *Handler) UploadMaterial(c *gin.Context) {
	classroomIDStr := c.PostForm("classroomId")
	folderIDStr := c.PostForm("folderId")

	classroomID, err := strconv.ParseUint(classroomIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "classroomId参数格式错误"))
		return
	}

	folderID, err := strconv.ParseUint(folderIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "folderId参数格式错误"))
		return
	}

	isSharedStr := c.PostForm("isShared")
	isShared, _ := strconv.Atoi(isSharedStr)

	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "请选择要上传的文件"))
		return
	}

	// 创建唯一文件名
	uniqueFileName := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)

	// 保存文件到本地
	uploadDir := "./uploads/classroom"
	if err := c.SaveUploadedFile(file, fmt.Sprintf("%s/%s", uploadDir, uniqueFileName)); err != nil {
		c.JSON(http.StatusOK, errorResponse(500, "文件保存失败"))
		return
	}

	// 文件访问路径
	filePath := "/uploads/classroom/" + uniqueFileName

	// 获取当前用户ID
	creatorID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "未登录"))
		return
	}

	db := client.GetDB()

	// 验证folder（如果folderID不为0）
	if folderID != 0 {
		var folder model.ClassroomFolder
		if err := db.Where("id = ? AND classroom_id = ? AND status = 1", folderID, classroomID).First(&folder).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusOK, errorResponse(404, "文件夹不存在"))
			} else {
				c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
			}
			return
		}
	}

	// 检查文件类型
	allowedTypes := map[string]bool{
		"pdf":  true,
		"word": true,
		"ppt":  true,
		"txt":  true,
		"mp4":  true,
	}

	// 简单的文件类型判断
	fileType := "txt"
	for ext := range allowedTypes {
		if len(file.Filename) > len(ext) &&
			file.Filename[len(file.Filename)-len(ext):] == ext {
			fileType = ext
			break
		}
	}

	material := &model.ClassroomMaterial{
		ClassroomID:   classroomID,
		FolderID:      folderID,
		FileName:      file.Filename,
		FileType:      fileType,
		FilePath:      filePath,
		FileSize:      uint64(file.Size),
		CreatorID:     creatorID.(string),
		IsShared:      isShared,
		DownloadCount: 0,
		Status:        1,
	}

	if err := db.Create(material).Error; err != nil {
		c.JSON(http.StatusOK, errorResponse(500, "上传失败"))
		return
	}

	c.JSON(http.StatusOK, successResponse(material))
}

// GetMaterials 获取资料列表（教师/学生）
func (h *Handler) GetMaterials(c *gin.Context) {
	classroomIDStr := c.Param("classroomId")
	folderIDStr := c.Param("folderId")

	// 调试日志
	fmt.Printf("[DEBUG] GetMaterials called with classroomId=%s, folderId=%s, path=%s\n",
		classroomIDStr, folderIDStr, c.Request.URL.Path)

	var classroomID, folderID uint64
	var err error

	// 获取classroom_id（必须参数，防止跨班级数据泄露）
	if classroomIDStr == "" {
		c.JSON(http.StatusOK, errorResponse(400, "classroomId参数缺失"))
		return
	}

	classroomID, err = strconv.ParseUint(classroomIDStr, 10, 64)
	if err != nil {
		fmt.Printf("[DEBUG] ParseUint error: %v, input=%s\n", err, classroomIDStr)
		c.JSON(http.StatusOK, errorResponse(400, "classroomId参数格式错误"))
		return
	}

	// 支持字符串 'root' 表示根目录(folderId=0)
	if folderIDStr == "root" {
		folderID = 0
	} else {
		folderID, err = strconv.ParseUint(folderIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusOK, errorResponse(400, "folderId参数格式错误"))
			return
		}
	}

	db := client.GetDB()

	// 安全检查：验证folder并确保它属于指定的classroom
	var folder model.ClassroomFolder
	if folderID != 0 {
		// 非根目录：验证folder存在且属于该classroom
		if err := db.Where("id = ? AND classroom_id = ? AND status = 1", folderID, classroomID).First(&folder).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusOK, errorResponse(404, "文件夹不存在"))
			} else {
				c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
			}
			return
		}
	} else {
		// 根目录：验证classroom存在
		var classroom model.Classroom
		if err := db.Where("id = ? AND status = 1", classroomID).First(&classroom).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusOK, errorResponse(404, "班级不存在"))
			} else {
				c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
			}
			return
		}
	}

	var materials []model.ClassroomMaterial

	// 直接使用classroom_id过滤，确保数据隔离
	query := db.Where("classroom_id = ? AND status = 1", classroomID)

	// 如果指定了folderId，进一步过滤
	if folderID != 0 {
		query = query.Where("folder_id = ?", folderID)
	}

	if err := query.
		Preload("Creator").
		Order("create_time DESC").
		Find(&materials).Error; err != nil {
		fmt.Printf("[DEBUG] 查询资料失败: %v, classroomID=%d, folderID=%d\n", err, classroomID, folderID)
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	fmt.Printf("[DEBUG] 查询到 %d 个资料, classroomID=%d, folderID=%d\n", len(materials), classroomID, folderID)

	// 获取当前用户ID，用于权限查询
	uid, exists := c.Get("userId")
	if exists {
		// 批量查询所有资料的权限
		materialIDs := make([]uint64, len(materials))
		for i, m := range materials {
			materialIDs[i] = m.ID
		}

		var permissions []model.ClassroomMaterialPermission
		if err := db.Where("material_id IN ? AND student_uid = ?", materialIDs, uid.(string)).
			Find(&permissions).Error; err == nil {
			// 构建权限映射
			permissionMap := make(map[uint64]*model.ClassroomMaterialPermission)
			for i := range permissions {
				permissionMap[permissions[i].MaterialID] = &permissions[i]
			}

			// 为每个资料添加权限信息
			for i := range materials {
				if perm, exists := permissionMap[materials[i].ID]; exists {
					materials[i].Permission = &model.ClassroomMaterialPermission{
						CanPreview:  perm.CanPreview,
						CanDownload: perm.CanDownload,
					}
				} else {
					// 没有权限记录，默认不可访问
					materials[i].Permission = &model.ClassroomMaterialPermission{
						CanPreview:  0,
						CanDownload: 0,
					}
				}
			}
		}
	}

	c.JSON(http.StatusOK, successResponse(materials))
}

// DeleteMaterial 删除资料（教师）
func (h *Handler) DeleteMaterial(c *gin.Context) {
	materialIDStr := c.Param("materialId")
	materialID, err := strconv.ParseUint(materialIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "materialId参数格式错误"))
		return
	}

	db := client.GetDB()
	logger := utils.GetLogger()

	// 查询资料信息以获取文件路径
	var material model.ClassroomMaterial
	if err := db.Where("id = ?", materialID).First(&material).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, errorResponse(404, "资料不存在"))
		} else {
			c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		}
		return
	}

	// 删除COS上的文件及其预览缓存
	cosService, err := service.NewCOSService()
	if err == nil {
		// 构建COS文件路径
		cosPath := fmt.Sprintf("classroom/materials/%d/%d_%s", material.FolderID, material.ID, material.FileName)

		// 删除COS文件和预览缓存（避免持续计费）
		if err := cosService.DeleteMaterialAndCache(cosPath); err != nil {
			logger.Warn("删除COS文件失败", zap.String("cosPath", cosPath), zap.Error(err))
			// 继续执行，不因为COS删除失败而阻止本地删除
		} else {
			logger.Info("COS文件及预览缓存已删除", zap.String("cosPath", cosPath), zap.Uint64("material_id", material.ID))
		}
	} else {
		logger.Warn("COS服务初始化失败，跳过COS文件删除", zap.Error(err))
	}

	// 删除磁盘上的文件
	if material.FilePath != "" {
		// 转换URL路径为文件系统路径
		filePath := "." + material.FilePath
		os.Remove(filePath) // 忽略错误
	}

	// 软删除数据库记录
	result := db.Model(&model.ClassroomMaterial{}).
		Where("id = ?", materialID).
		Update("status", 0)

	if result.Error != nil {
		c.JSON(http.StatusOK, errorResponse(500, "删除失败"))
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusOK, errorResponse(404, "资料不存在"))
		return
	}

	c.JSON(http.StatusOK, successResponse(nil))
}

// CopyMaterialToClassroom 复制资料到班级（教师）
func (h *Handler) CopyMaterialToClassroom(c *gin.Context) {
	var req struct {
		MaterialID uint64 `json:"materialId" binding:"required"`
		FolderID   uint64 `json:"folderId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	// 获取当前用户ID
	creatorID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "未登录"))
		return
	}

	db := client.GetDB()

	// 查找原资料
	var originalMaterial model.ClassroomMaterial
	if err := db.Where("id = ?", req.MaterialID).First(&originalMaterial).Error; err != nil {
		c.JSON(http.StatusOK, errorResponse(404, "资料不存在"))
		return
	}

	// 创建新资料
	newMaterial := &model.ClassroomMaterial{
		FolderID:      req.FolderID,
		FileName:      originalMaterial.FileName,
		FileType:      originalMaterial.FileType,
		FilePath:      originalMaterial.FilePath,
		FileSize:      originalMaterial.FileSize,
		CreatorID:     creatorID.(string),
		IsShared:      0,
		DownloadCount: 0,
		Status:        1,
	}

	if err := db.Create(newMaterial).Error; err != nil {
		c.JSON(http.StatusOK, errorResponse(500, "复制失败"))
		return
	}

	c.JSON(http.StatusOK, successResponse(newMaterial))
}

// ==================== 随机选人 ====================

// RandomPick 随机选人（教师）
func (h *Handler) RandomPick(c *gin.Context) {
	logger := utils.GetLogger()
	classroomIDStr := c.Param("classroomId")
	classroomID, err := strconv.ParseUint(classroomIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "classroomId参数格式错误"))
		return
	}

	db := client.GetDB()

	// 获取班级所有学生
	var students []model.ClassroomStudent
	if err := db.Where("classroom_id = ? AND status = 1", classroomID).
		Find(&students).Error; err != nil {
		logger.Error("查询学生列表失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	if len(students) == 0 {
		c.JSON(http.StatusOK, errorResponse(400, "班级中没有学生"))
		return
	}

	// 随机选择一个学生
	rand.Seed(time.Now().UnixNano())
	pickedStudent := students[rand.Intn(len(students))]

	// 记录选人
	pickRecord := &model.ClassroomRandomPick{
		ClassroomID: classroomID,
		PickedUID:   pickedStudent.UID,
	}

	if err := db.Create(pickRecord).Error; err != nil {
		logger.Error("记录选人失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "操作失败"))
		return
	}

	// 加载用户信息
	db.Preload("PickedUser").First(pickRecord)

	// 查询该学生在班级中的详细信息
	var studentInfo model.ClassroomStudent
	if err := db.Where("classroom_id = ? AND uid = ?", classroomID, pickedStudent.UID).
		First(&studentInfo).Error; err == nil {
		pickRecord.StudentInfo = &studentInfo
	}

	logger.Info("随机选人", zap.Uint64("classroom_id", classroomID), zap.String("picked_uid", pickedStudent.UID))

	// TODO: 通过WebSocket通知所有学生

	c.JSON(http.StatusOK, successResponse(pickRecord))
}

// GetPickHistory 获取选人历史（教师/学生）
func (h *Handler) GetPickHistory(c *gin.Context) {
	logger := utils.GetLogger()
	classroomIDStr := c.Param("classroomId")
	classroomID, err := strconv.ParseUint(classroomIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "classroomId参数格式错误"))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	db := client.GetDB()
	var picks []model.ClassroomRandomPick

	if err := db.Where("classroom_id = ?", classroomID).
		Preload("PickedUser").
		Order("pick_time DESC").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&picks).Error; err != nil {
		logger.Error("查询选人历史失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	// 为每条记录填充学生信息
	for i := range picks {
		var studentInfo model.ClassroomStudent
		if err := db.Where("classroom_id = ? AND uid = ?", classroomID, picks[i].PickedUID).
			First(&studentInfo).Error; err == nil {
			picks[i].StudentInfo = &studentInfo
		}
	}

	c.JSON(http.StatusOK, successResponse(picks))
}

// ==================== 班级消息 ====================

// SendMessage 发送消息（教师/学生）
func (h *Handler) SendMessage(c *gin.Context) {
	logger := utils.GetLogger()

	var req struct {
		ClassroomID uint64 `json:"classroomId" binding:"required"`
		Content     string `json:"content"`
		ImageURL    string `json:"imageUrl"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	if req.Content == "" && req.ImageURL == "" {
		c.JSON(http.StatusOK, errorResponse(400, "消息内容不能为空"))
		return
	}

	// 获取当前用户ID
	senderID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "未登录"))
		return
	}

	db := client.GetDB()

	msgType := "text"
	if req.ImageURL != "" {
		msgType = "image"
	}

	message := &model.ClassroomMessage{
		ClassroomID: req.ClassroomID,
		SenderID:    senderID.(string),
		Content:     req.Content,
		ImageURL:    req.ImageURL,
		MsgType:     msgType,
	}

	if err := db.Create(message).Error; err != nil {
		logger.Error("发送消息失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "发送失败"))
		return
	}

	// 加载发送者信息
	db.Preload("Sender").First(message)

	// 如果是学生，加载学生在班级中的信息
	var studentInfo model.ClassroomStudent
	if err := db.Where("classroom_id = ? AND uid = ?", req.ClassroomID, senderID.(string)).
		First(&studentInfo).Error; err == nil {
		message.StudentInfo = &studentInfo
	}

	logger.Info("发送班级消息", zap.Uint64("classroom_id", req.ClassroomID), zap.String("sender_id", senderID.(string)))

	// TODO: 通过WebSocket广播消息给班级所有人

	c.JSON(http.StatusOK, successResponse(message))
}

// GetMessages 获取消息列表（教师/学生）
func (h *Handler) GetMessages(c *gin.Context) {
	logger := utils.GetLogger()
	classroomIDStr := c.Param("classroomId")
	classroomID, err := strconv.ParseUint(classroomIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "classroomId参数格式错误"))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	db := client.GetDB()
	var messages []model.ClassroomMessage

	if err := db.Where("classroom_id = ?", classroomID).
		Preload("Sender").
		Order("create_time DESC").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&messages).Error; err != nil {
		logger.Error("查询消息列表失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	// 为每条消息查询学生信息（如果是学生发送的）
	for i := range messages {
		if messages[i].SenderID != "system" && messages[i].SenderID != "" {
			var studentInfo model.ClassroomStudent
			if err := db.Where("classroom_id = ? AND uid = ?", classroomID, messages[i].SenderID).
				First(&studentInfo).Error; err == nil {
				messages[i].StudentInfo = &studentInfo
			}
		}
	}

	// 反转顺序，使最新消息在最后
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	c.JSON(http.StatusOK, successResponse(messages))
}

// UploadMessageImage 上传讨论消息图片（教师/学生）
func (h *Handler) UploadMessageImage(c *gin.Context) {
	logger := utils.GetLogger()

	// 记录请求信息
	logger.Info("收到图片上传请求")

	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		logger.Warn("获取上传文件失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "请选择要上传的图片"))
		return
	}

	logger.Info("获取文件成功", zap.String("filename", file.Filename), zap.Int64("size", file.Size))

	// 验证文件类型
	allowedExtensions := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedExtensions[ext] {
		logger.Warn("不支持的文件格式", zap.String("ext", ext))
		c.JSON(http.StatusOK, errorResponse(400, "只支持上传图片格式（jpg, jpeg, png, gif, webp）"))
		return
	}

	// 验证文件大小（最大10MB）
	const maxFileSize = 10 * 1024 * 1024
	if file.Size > maxFileSize {
		logger.Warn("文件过大", zap.Int64("size", file.Size))
		c.JSON(http.StatusOK, errorResponse(400, "图片大小不能超过10MB"))
		return
	}

	// 获取用户ID
	uid, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "未登录"))
		return
	}

	// 获取班级ID
	classroomIDStr := c.PostForm("classroomId")
	if classroomIDStr == "" {
		c.JSON(http.StatusOK, errorResponse(400, "缺少班级ID"))
		return
	}
	classroomID, err := strconv.ParseUint(classroomIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "班级ID格式错误"))
		return
	}

	db := client.GetDB()

	// 创建唯一文件名并保存到本地
	uniqueFileName := fmt.Sprintf("msg_%d_%s%s", time.Now().Unix(), generateRandomString(8), ext)

	// 创建上传目录
	uploadDir := "./uploads/classroom/images"
	if err := c.SaveUploadedFile(file, fmt.Sprintf("%s/%s", uploadDir, uniqueFileName)); err != nil {
		logger.Error("保存图片失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "图片保存失败"))
		return
	}

	imageURL := "/uploads/classroom/images/" + uniqueFileName

	// 创建消息记录
	message := &model.ClassroomMessage{
		ClassroomID: classroomID,
		SenderID:    uid.(string),
		MsgType:     "image",
		ImageURL:    imageURL,
	}

	if err := db.Create(message).Error; err != nil {
		logger.Error("创建消息记录失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "消息创建失败"))
		return
	}

	// 预加载发送者信息
	if err := db.Preload("Sender").First(message, message.ID).Error; err != nil {
		logger.Error("加载发送者信息失败", zap.Error(err))
	}

	logger.Info("上传消息图片成功", zap.String("filename", file.Filename), zap.String("url", imageURL), zap.Uint64("message_id", message.ID))

	c.JSON(http.StatusOK, successResponse(message))
}

// UploadQuestionImage 上传客观题题目内容图片（仅支持 png/jpg/jpeg）
func (h *Handler) UploadQuestionImage(c *gin.Context) {
	logger := utils.GetLogger()

	uid, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "未登录"))
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		logger.Warn("获取上传文件失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "请选择要上传的图片"))
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{
		".png":  true,
		".jpg":  true,
		".jpeg": true,
	}
	if !allowedExts[ext] {
		c.JSON(http.StatusOK, errorResponse(400, "仅支持 png/jpg/jpeg 格式图片"))
		return
	}

	const maxFileSize = 10 * 1024 * 1024 // 10MB
	if file.Size <= 0 {
		c.JSON(http.StatusOK, errorResponse(400, "图片文件不能为空"))
		return
	}
	if file.Size > maxFileSize {
		c.JSON(http.StatusOK, errorResponse(400, "图片大小不能超过10MB"))
		return
	}

	src, err := file.Open()
	if err != nil {
		logger.Error("打开上传文件失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "读取图片失败"))
		return
	}
	defer src.Close()

	buffer := make([]byte, 512)
	n, readErr := src.Read(buffer)
	if readErr != nil && readErr != io.EOF {
		logger.Error("读取上传文件失败", zap.Error(readErr))
		c.JSON(http.StatusOK, errorResponse(500, "读取图片失败"))
		return
	}

	contentType := http.DetectContentType(buffer[:n])
	allowedContentTypes := map[string]bool{
		"image/png":  true,
		"image/jpeg": true,
	}
	if !allowedContentTypes[contentType] {
		c.JSON(http.StatusOK, errorResponse(400, "仅支持 png/jpg/jpeg 格式图片"))
		return
	}

	uniqueFileName := fmt.Sprintf("question_%s_%d_%s%s", uid.(string), time.Now().Unix(), generateRandomString(8), ext)
	uploadDir := "./uploads/classroom/questions"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		logger.Error("创建题目图片目录失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "创建上传目录失败"))
		return
	}

	fullPath := fmt.Sprintf("%s/%s", uploadDir, uniqueFileName)
	if err := c.SaveUploadedFile(file, fullPath); err != nil {
		logger.Error("保存题目图片失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "图片保存失败"))
		return
	}

	imageURL := "/uploads/classroom/questions/" + uniqueFileName
	logger.Info("上传客观题内容图片成功",
		zap.String("uid", uid.(string)),
		zap.String("filename", file.Filename),
		zap.String("url", imageURL))

	c.JSON(http.StatusOK, successResponse(map[string]interface{}{
		"url":      imageURL,
		"filename": file.Filename,
		"size":     file.Size,
	}))
}

// DeleteQuestionImages 删除客观题题目内容中已移除的图片（仅允许删除 uploads/classroom/questions 下的 png/jpg/jpeg）
func (h *Handler) DeleteQuestionImages(c *gin.Context) {
	logger := utils.GetLogger()

	if _, exists := c.Get("userId"); !exists {
		c.JSON(http.StatusOK, errorResponse(401, "未登录"))
		return
	}

	var req struct {
		URLs []string `json:"urls" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("删除题目图片参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}
	if len(req.URLs) == 0 {
		c.JSON(http.StatusOK, successResponse(map[string]interface{}{
			"deleted": 0,
			"skipped": 0,
		}))
		return
	}

	deleted, skipped := deleteQuestionImagesByURLs(logger, req.URLs)

	c.JSON(http.StatusOK, successResponse(map[string]interface{}{
		"deleted": deleted,
		"skipped": skipped,
	}))
}

// generateRandomString 生成随机字符串
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// GradeHomework 教师批改作业（主观题）
func (h *Handler) GradeHomework(c *gin.Context) {
	logger := utils.GetLogger()

	var req struct {
		HomeworkID uint64  `json:"homeworkId" binding:"required"`
		QuestionID uint64  `json:"questionId" binding:"required"`
		UID        string  `json:"uid" binding:"required"`
		Score      float64 `json:"score"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("请求参数错误", zap.Error(err), zap.Any("body", c.Request.Body))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误: "+err.Error()))
		return
	}

	logger.Info("教师评分请求", zap.Uint64("homework_id", req.HomeworkID), zap.Uint64("question_id", req.QuestionID),
		zap.String("uid", req.UID), zap.Float64("score", req.Score))

	db := client.GetDB()

	// 查找提交记录
	var submit model.HomeworkSubmit
	if err := db.Where("homework_id = ? AND question_id = ? AND uid = ?",
		req.HomeworkID, req.QuestionID, req.UID).First(&submit).Error; err != nil {
		logger.Error("提交记录不存在", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(404, "提交记录不存在"))
		return
	}

	// 获取题目分值限制
	var homeworkQuestion model.HomeworkQuestion
	if err := db.Where("homework_id = ? AND question_id = ?",
		req.HomeworkID, req.QuestionID).First(&homeworkQuestion).Error; err != nil {
		logger.Error("题目不存在", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(404, "题目不存在"))
		return
	}

	// 验证分数
	if req.Score < 0 || req.Score > float64(homeworkQuestion.Score) {
		c.JSON(http.StatusOK, errorResponse(400, "分数必须在0到"+strconv.Itoa(homeworkQuestion.Score)+"之间"))
		return
	}

	// 更新分数
	submit.Score = req.Score
	submit.IsScored = 1

	if err := db.Save(&submit).Error; err != nil {
		logger.Error("批改失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "批改失败"))
		return
	}

	logger.Info("批改作业", zap.Uint64("homework_id", req.HomeworkID),
		zap.String("uid", req.UID), zap.Float64("score", req.Score))
	c.JSON(http.StatusOK, successResponse(submit))
}

// GradeProgrammingHomework 教师批改编程题
func (h *Handler) GradeProgrammingHomework(c *gin.Context) {
	logger := utils.GetLogger()

	var req struct {
		HomeworkID uint64  `json:"homeworkId" binding:"required"`
		ProblemID  string  `json:"problemId" binding:"required"`
		UID        string  `json:"uid" binding:"required"`
		Score      float64 `json:"score"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误: "+err.Error()))
		return
	}

	logger.Info("教师评分编程题请求", zap.Uint64("homework_id", req.HomeworkID),
		zap.String("problem_id", req.ProblemID), zap.String("uid", req.UID), zap.Float64("score", req.Score))

	db := client.GetDB()

	// 查找提交记录
	var submit model.HomeworkSubmit
	if err := db.Where("homework_id = ? AND problem_id = ? AND uid = ?",
		req.HomeworkID, req.ProblemID, req.UID).First(&submit).Error; err != nil {
		logger.Error("提交记录不存在", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(404, "提交记录不存在"))
		return
	}

	// 获取题目分值限制
	var homeworkQuestion model.HomeworkQuestion
	if err := db.Where("homework_id = ? AND problem_id = ?",
		req.HomeworkID, req.ProblemID).First(&homeworkQuestion).Error; err != nil {
		logger.Error("题目不存在", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(404, "题目不存在"))
		return
	}

	// 验证分数
	if req.Score < 0 || req.Score > float64(homeworkQuestion.Score) {
		c.JSON(http.StatusOK, errorResponse(400, "分数必须在0到"+strconv.Itoa(homeworkQuestion.Score)+"之间"))
		return
	}

	// 更新分数
	submit.Score = req.Score
	submit.IsScored = 1

	if err := db.Save(&submit).Error; err != nil {
		logger.Error("批改编程题失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "批改失败"))
		return
	}

	logger.Info("批改编程题", zap.Uint64("homework_id", req.HomeworkID),
		zap.String("uid", req.UID), zap.Float64("score", req.Score))
	c.JSON(http.StatusOK, successResponse(submit))
}

// RecalculateScore 重新计算客观题分数
func (h *Handler) RecalculateScore(c *gin.Context) {
	logger := utils.GetLogger()

	var req struct {
		HomeworkID uint64 `json:"homeworkId" binding:"required"`
		QuestionID uint64 `json:"questionId" binding:"required"`
		UID        string `json:"uid" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	db := client.GetDB()

	// 查找题目
	var question model.QuestionBank
	if err := db.Where("id = ?", req.QuestionID).First(&question).Error; err != nil {
		logger.Error("题目不存在", zap.Uint64("question_id", req.QuestionID))
		c.JSON(http.StatusOK, errorResponse(404, "题目不存在"))
		return
	}

	// 检查是否为客观题
	if question.Type != "single_choice" && question.Type != "multiple_choice" && question.Type != "judge" {
		c.JSON(http.StatusOK, errorResponse(400, "只能重新计算客观题分数"))
		return
	}

	// 查找提交记录
	var submit model.HomeworkSubmit
	if err := db.Where("homework_id = ? AND question_id = ? AND uid = ?",
		req.HomeworkID, req.QuestionID, req.UID).First(&submit).Error; err != nil {
		logger.Error("提交记录不存在", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(404, "提交记录不存在"))
		return
	}

	// 重新计算分数
	score := 0.0
	if question.Type == "single_choice" {
		if submit.Answer == question.Answer {
			var homeworkQuestion model.HomeworkQuestion
			if err := db.Where("homework_id = ? AND question_id = ?",
				req.HomeworkID, req.QuestionID).First(&homeworkQuestion).Error; err == nil {
				score = float64(homeworkQuestion.Score)
			}
		}
	} else if question.Type == "judge" {
		// 对于判断题，使用严格规范化比较（仅 true/false）
		if compareJudgeAnswers(submit.Answer, question.Answer) {
			var homeworkQuestion model.HomeworkQuestion
			if err := db.Where("homework_id = ? AND question_id = ?",
				req.HomeworkID, req.QuestionID).First(&homeworkQuestion).Error; err == nil {
				score = float64(homeworkQuestion.Score)
			}
		}
	} else if question.Type == "multiple_choice" {
		// 对于多选题，需要比较JSON数组或逗号分隔字符串
		var studentAnswers, correctAnswers []string

		// 解析学生答案（新格式：JSON数组）
		if err := json.Unmarshal([]byte(submit.Answer), &studentAnswers); err != nil {
			// 兼容旧格式：逗号分隔字符串
			studentAnswers = strings.Split(submit.Answer, ",")
			// 去除空格
			for i := range studentAnswers {
				studentAnswers[i] = strings.TrimSpace(studentAnswers[i])
			}
		}

		// 解析正确答案（新格式：JSON数组）
		if err := json.Unmarshal([]byte(question.Answer), &correctAnswers); err != nil {
			// 兼容旧格式：逗号分隔字符串
			correctAnswers = strings.Split(question.Answer, ",")
			// 去除空格
			for i := range correctAnswers {
				correctAnswers[i] = strings.TrimSpace(correctAnswers[i])
			}
		}

		// 比较答案
		if compareArrays(studentAnswers, correctAnswers) {
			var homeworkQuestion model.HomeworkQuestion
			if err := db.Where("homework_id = ? AND question_id = ?",
				req.HomeworkID, req.QuestionID).First(&homeworkQuestion).Error; err == nil {
				score = float64(homeworkQuestion.Score)
			}
		}
	}

	// 更新分数
	submit.Score = score
	submit.IsScored = 1

	if err := db.Save(&submit).Error; err != nil {
		logger.Error("重新计算分数失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "重新计算失败"))
		return
	}

	logger.Info("重新计算分数", zap.Uint64("homework_id", req.HomeworkID),
		zap.Uint64("question_id", req.QuestionID), zap.String("uid", req.UID), zap.Float64("score", score))
	c.JSON(http.StatusOK, successResponse(submit))
}

// DeleteHomework 删除作业（教师）
func (h *Handler) DeleteHomework(c *gin.Context) {
	logger := utils.GetLogger()
	homeworkIDStr := c.Param("homeworkId")
	homeworkID, err := strconv.ParseUint(homeworkIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "homeworkId参数格式错误"))
		return
	}

	db := client.GetDB()

	// 检查作业是否存在
	var homework model.ClassroomHomework
	if err := db.Where("id = ?", homeworkID).First(&homework).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, errorResponse(404, "作业不存在"))
		} else {
			logger.Error("查询作业失败", zap.Error(err))
			c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		}
		return
	}

	// 开始事务
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 删除作业题目关联
	if err := tx.Where("homework_id = ?", homeworkID).Delete(&model.HomeworkQuestion{}).Error; err != nil {
		tx.Rollback()
		logger.Error("删除作业题目失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "删除失败"))
		return
	}

	// 删除作业提交记录
	if err := tx.Where("homework_id = ?", homeworkID).Delete(&model.HomeworkSubmit{}).Error; err != nil {
		tx.Rollback()
		logger.Error("删除作业提交记录失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "删除失败"))
		return
	}

	// 删除作业
	if err := tx.Delete(&homework).Error; err != nil {
		tx.Rollback()
		logger.Error("删除作业失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "删除失败"))
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		logger.Error("提交事务失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "删除失败"))
		return
	}

	logger.Info("删除作业", zap.Uint64("homework_id", homeworkID))
	c.JSON(http.StatusOK, successResponse(nil))
}

// RecallMessage 撤回消息
func (h *Handler) RecallMessage(c *gin.Context) {
	logger := utils.GetLogger()

	// 获取当前用户ID
	uid, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "未登录"))
		return
	}

	messageIDStr := c.Param("messageId")
	messageID, err := strconv.ParseUint(messageIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "消息ID格式错误"))
		return
	}

	db := client.GetDB()

	// 查找消息
	var message model.ClassroomMessage
	if err := db.Where("id = ?", messageID).First(&message).Error; err != nil {
		logger.Error("查找消息失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(404, "消息不存在"))
		return
	}

	// 查找班级信息，检查当前用户是否是教师
	var classroom model.Classroom
	if err := db.Where("id = ?", message.ClassroomID).First(&classroom).Error; err != nil {
		logger.Error("查找班级失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(404, "班级不存在"))
		return
	}

	isTeacher := classroom.TeacherID == uid.(string)
	isSender := message.SenderID == uid.(string)

	// 验证权限：教师可以撤回任意消息，学生只能撤回自己的消息
	if !isTeacher && !isSender {
		c.JSON(http.StatusOK, errorResponse(403, "无权撤回此消息"))
		return
	}

	// 获取操作者（撤回消息的人）的姓名
	var operatorName string
	if isTeacher {
		// 操作者是教师，获取教师姓名
		var teacherUserInfo model.UserInfo
		if err := db.Where("uuid = ?", uid.(string)).First(&teacherUserInfo).Error; err == nil {
			operatorName = teacherUserInfo.Nickname
			if operatorName == "" {
				operatorName = teacherUserInfo.Username
			}
		} else {
			operatorName = "教师"
		}
	} else {
		// 操作者是学生，获取学生在该班级的真实姓名
		var studentClassroomInfo model.ClassroomStudent
		if err := db.Where("classroom_id = ? AND uid = ?", message.ClassroomID, uid.(string)).
			First(&studentClassroomInfo).Error; err == nil {
			// 找到班级学生记录，使用真实姓名
			operatorName = studentClassroomInfo.RealName
		} else {
			// 未找到班级学生记录，尝试从用户信息表获取昵称
			var studentUserInfo model.UserInfo
			if err := db.Where("uuid = ?", uid.(string)).First(&studentUserInfo).Error; err == nil {
				operatorName = studentUserInfo.Nickname
				if operatorName == "" {
					operatorName = studentUserInfo.Username
				}
			} else {
				operatorName = "某学生"
			}
		}
	}

	// 创建系统消息
	systemContent := ""
	if isTeacher && isSender {
		// 教师撤回自己的消息
		systemContent = fmt.Sprintf("%s 撤回了一条消息", operatorName)
	} else if isTeacher {
		// 教师撤回学生的消息，需要获取学生姓名
		var senderName string
		var classroomStudent model.ClassroomStudent
		if err := db.Where("classroom_id = ? AND uid = ?", message.ClassroomID, message.SenderID).
			First(&classroomStudent).Error; err == nil {
			// 找到班级学生记录，使用真实姓名
			senderName = classroomStudent.RealName
		} else {
			// 未找到班级学生记录，尝试从用户信息表获取昵称
			var senderUserInfo model.UserInfo
			if err := db.Where("uuid = ?", message.SenderID).First(&senderUserInfo).Error; err == nil {
				senderName = senderUserInfo.Nickname
				if senderName == "" {
					senderName = senderUserInfo.Username
				}
			} else {
				senderName = "某学生"
			}
		}
		systemContent = fmt.Sprintf("教师 %s 撤回了 %s 的消息", operatorName, senderName)
	} else {
		// 学生撤回自己的消息
		systemContent = fmt.Sprintf("%s 撤回了一条消息", operatorName)
	}

	systemMessage := &model.ClassroomMessage{
		ClassroomID: message.ClassroomID,
		SenderID:    "system", // 系统消息
		Content:     systemContent,
		MsgType:     "system",
	}

	if err := db.Create(systemMessage).Error; err != nil {
		logger.Error("创建系统消息失败", zap.Error(err))
	}

	// 删除原消息
	if err := db.Delete(&message).Error; err != nil {
		logger.Error("撤回消息失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "撤回失败"))
		return
	}

	logger.Info("撤回消息", zap.Uint64("message_id", messageID), zap.String("operator", uid.(string)))
	c.JSON(http.StatusOK, successResponse(nil))
}

// ==================== 编程题提交记录 ====================

// SaveProgrammingSubmission 保存编程题提交记录
func (h *Handler) SaveProgrammingSubmission(c *gin.Context) {
	logger := utils.GetLogger()
	logger.Info("=== SaveProgrammingSubmission 被调用 ===")

	var req struct {
		HomeworkID uint64      `json:"homeworkId" binding:"required"`
		ProblemID  string      `json:"problemId" binding:"required"` // 编程题ID
		SubmitID   interface{} `json:"submitId" binding:"required"`  // 支持字符串或数字
		Code       string      `json:"code"`
		Language   string      `json:"language"`
		Token      string      `json:"token"` // BingOJ token
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	// 将 SubmitID 从 interface{} 转换为 uint64
	var submitID uint64
	switch v := req.SubmitID.(type) {
	case float64:
		submitID = uint64(v)
	case int:
		submitID = uint64(v)
	case int64:
		submitID = uint64(v)
	case uint64:
		submitID = v
	case string:
		// 尝试解析字符串
		id, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			logger.Error("submitId 无法解析为数字", zap.String("value", v), zap.Error(err))
			c.JSON(http.StatusOK, errorResponse(400, "submitId 必须是数字"))
			return
		}
		submitID = id
	default:
		logger.Error("submitId 类型错误", zap.Any("type", fmt.Sprintf("%T", req.SubmitID)))
		c.JSON(http.StatusOK, errorResponse(400, "submitId 格式错误"))
		return
	}

	logger.Info("SaveProgrammingSubmission 参数解析成功",
		zap.Uint64("homeworkId", req.HomeworkID),
		zap.String("problemId", req.ProblemID),
		zap.Uint64("submitId", submitID))

	// 获取当前用户ID
	uid, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "未登录"))
		return
	}

	db := client.GetDB()

	// 构建答案 JSON（包含提交ID、代码、语言）
	answerData := map[string]interface{}{
		"submitId": submitID,
		"code":     req.Code,
		"language": req.Language,
	}
	answerJSON, _ := json.Marshal(answerData)

	// 调试日志：检查代码是否为空
	logger.Info("准备保存编程题提交记录",
		zap.Uint64("homework_id", req.HomeworkID),
		zap.String("problem_id", req.ProblemID),
		zap.Uint64("submit_id", submitID),
		zap.Int("code_length", len(req.Code)),
		zap.String("language", req.Language),
		zap.String("answer_json", string(answerJSON)))

	// 查找是否已有提交记录（通过 problem_id 查询）
	var existingSubmit model.HomeworkSubmit
	checkErr := db.Where("homework_id = ? AND problem_id = ? AND uid = ?",
		req.HomeworkID, req.ProblemID, uid.(string)).
		First(&existingSubmit).Error

	logger.Info("查询编程题提交记录",
		zap.Uint64("homework_id", req.HomeworkID),
		zap.String("problem_id", req.ProblemID),
		zap.String("uid", uid.(string)),
		zap.String("found", fmt.Sprintf("%v", checkErr == nil)))

	var submit *model.HomeworkSubmit
	if checkErr == nil {
		// 更新已有记录
		existingSubmit.SubmitID = &submitID
		existingSubmit.Answer = string(answerJSON)
		// 注意：这里不设置 IsOfficiallySubmitted = 1
		// 编程题提交代码只是草稿保存，只有点击"提交作业"才算正式提交
		// 保持原有的分数
		submit = &existingSubmit

		if err := db.Save(&existingSubmit).Error; err != nil {
			logger.Error("更新编程题提交记录失败", zap.Error(err))
			c.JSON(http.StatusOK, errorResponse(500, "保存失败"))
			return
		}
		logger.Info("更新编程题提交记录成功", zap.Uint64("submit_id", existingSubmit.ID))
	} else {
		// 创建新记录
		submit = &model.HomeworkSubmit{
			HomeworkID:            req.HomeworkID,
			ProblemID:             &req.ProblemID, // 使用 ProblemID
			UID:                   uid.(string),
			Answer:                string(answerJSON),
			SubmitID:              &submitID,
			Score:                 0,
			IsScored:              0,
			IsOfficiallySubmitted: 0, // 编程题提交代码不算正式提交整个作业
		}

		if err := db.Create(submit).Error; err != nil {
			logger.Error("保存编程题提交记录失败", zap.Error(err))
			c.JSON(http.StatusOK, errorResponse(500, "保存失败"))
			return
		}
		logger.Info("创建编程题提交记录成功", zap.Uint64("submit_id", submit.ID))
	}

	// 异步获取评测结果（不阻塞响应）
	go func() {
		bingoJClient := h.judgeService.GetBingoJClient()

		// 设置 token（从请求中获取）
		if req.Token != "" {
			bingoJClient.SetToken(req.Token)
			logger.Info("设置 BingOJ Token 成功")
		} else {
			logger.Warn("未提供 BingOJ Token，无法获取提交结果")
			return
		}

		// 等待2秒，让 BingOJ 开始评测
		time.Sleep(2 * time.Second)

		// 轮询获取提交列表，查找当前提交的结果（最多重试20次，每次间隔2秒）
		for i := 0; i < 20; i++ {
			// 获取提交列表（仅看自己的，第一页，获取最新15条）
			submissions, err := bingoJClient.GetSubmissionList(true, 1, 15)
			if err != nil {
				logger.Warn("获取提交列表失败", zap.Error(err), zap.Int("retry", i+1))
				time.Sleep(3 * time.Second)
				continue
			}

			// 查找当前提交的记录
			var targetSubmission *client.SubmissionListItem
			for j := range submissions {
				submitIDStr := fmt.Sprintf("%v", submissions[j].SubmitID)
				if submitIDStr == fmt.Sprintf("%d", submitID) {
					targetSubmission = &submissions[j]
					break
				}
			}

			// 如果找到提交记录且判题完成
			if targetSubmission != nil {
				// HOJ 官方状态码定义（与 HOJ 系统保持一致）：
				// status: -10=未提交(NS), -5=结果未知(SNR), -4=已取消(CA),
				//         -3=格式错误(PE), -2=编译错误(CE), -1=答案错误(WA),
				//         0=通过(AC), 1=时间超限(TLE), 2=内存超限(MLE),
				//         3=运行错误(RE), 4=系统错误(SE), 5=等待中(Pending),
				//         6=编译中(CP), 7=判题中(Judging), 8=部分通过(PAC),
				//         9=提交中(Submitting), 10=提交失败(SF)
				status := targetSubmission.Status

				// 详细日志：打印从 BingOJ 获取的提交信息
				logger.Info("从 BingOJ 获取到提交记录",
					zap.Uint64("submit_id", submitID),
					zap.Int("result", targetSubmission.Result),
					zap.Int("status", status),
					zap.String("score", fmt.Sprintf("%d", targetSubmission.Score)),
					zap.String("language", targetSubmission.Language),
					zap.String("celInfo", targetSubmission.CELInfo))

				// 检查是否还在判题中或提交中（status 5=Pending, 6=Compiling, 7=Judging, 9=Submitting）
				if status == 5 || status == 6 || status == 7 || status == 9 {
					logger.Debug("还在判题中，继续等待",
						zap.Int("status", status))
					time.Sleep(2 * time.Second)
					continue
				}

				// 检查是否提交失败
				if status == 10 {
					logger.Warn("提交失败", zap.Uint64("submit_id", submitID))
					return
				}

				// 判题完成，更新数据库
				score := 0.0
				judgeResult := ""

				// 根据 status 判断评测结果（使用 HOJ 官方状态码标准）
				switch status {
				case 0:
					// AC - Accepted（通过）
					score = calculateQuestionScoreForProblem(req.HomeworkID, req.ProblemID)
					judgeResult = "AC"
				case -2:
					// CE - Compile Error（编译错误）
					judgeResult = "CE"
					score = 0
				case -3:
					// PE - Presentation Error（格式错误）
					judgeResult = "PE"
					score = 0
				case -1:
					// WA - Wrong Answer（答案错误）
					judgeResult = "WA"
					score = 0
				case 1:
					// TLE - Time Limit Exceeded（时间超限）
					judgeResult = "TLE"
					score = 0
				case 2:
					// MLE - Memory Limit Exceeded（内存超限）
					judgeResult = "MLE"
					score = 0
				case 3:
					// RE - Runtime Error（运行错误）
					judgeResult = "RE"
					score = 0
				case 4:
					// SE - System Error（系统错误）
					judgeResult = "SE"
					score = 0
				case 8:
					// PAC - Partial Accepted（部分通过）- 按比例计算分数
					fullScore := calculateQuestionScoreForProblem(req.HomeworkID, req.ProblemID)
					score = float64(targetSubmission.Score) / 100.0 * fullScore
					judgeResult = "PAC"
				case -4:
					// CA - Cancelled（已取消）
					judgeResult = "CA"
					score = 0
				case -5:
					// SNR - Submitted Unknown Result（结果未知）
					judgeResult = "SNR"
					score = 0
				default:
					// 未知状态，保留原始状态码数字
					judgeResult = fmt.Sprintf("Unknown(%d)", status)
					logger.Warn("未知的评测状态码",
						zap.Int("status", status),
						zap.Uint64("submit_id", submitID))
				}

				// 更新提交记录 - 只更新评测相关字段，不更新 answer
				// 原因：answer 字段可能已经被新的提交覆盖了
				// 我们应该保留数据库中最新的代码，而不是用旧代码覆盖
				db.Model(submit).Updates(map[string]interface{}{
					"score":        score,
					"judge_result": judgeResult,
					"is_scored":    1,
				})

				logger.Info("编程题评测完成",
					zap.Uint64("homework_id", req.HomeworkID),
					zap.String("problem_id", req.ProblemID),
					zap.Uint64("submit_id", submitID),
					zap.Int("status", status),
					zap.String("judge_result", judgeResult),
					zap.Float64("score", score))
				return
			}

			logger.Debug("未找到提交记录，继续等待", zap.Uint64("submit_id", submitID))
			time.Sleep(2 * time.Second)
		}

		logger.Warn("获取评测结果超时", zap.Uint64("submit_id", submitID))
	}()

	logger.Info("保存编程题提交记录", zap.Uint64("homework_id", req.HomeworkID),
		zap.String("problem_id", req.ProblemID), zap.Uint64("submit_id", submitID))
	c.JSON(http.StatusOK, successResponse(map[string]interface{}{
		"message": "保存成功",
	}))
}

// calculateQuestionScore 计算题目分数（用于普通题目，通过 question_id 查询）
func calculateQuestionScore(homeworkID, questionID uint64) float64 {
	db := client.GetDB()
	var homeworkQuestion model.HomeworkQuestion
	if err := db.Where("homework_id = ? AND question_id = ?", homeworkID, questionID).
		First(&homeworkQuestion).Error; err != nil {
		return 0
	}
	return float64(homeworkQuestion.Score)
}

// calculateQuestionScoreForProblem 计算编程题分数（通过 problem_id 查询）
func calculateQuestionScoreForProblem(homeworkID uint64, problemID string) float64 {
	db := client.GetDB()
	var homeworkQuestion model.HomeworkQuestion
	if err := db.Where("homework_id = ? AND problem_id = ?", homeworkID, problemID).
		First(&homeworkQuestion).Error; err != nil {
		return 0
	}
	return float64(homeworkQuestion.Score)
}

// GetProgrammingSubmissions 获取编程题提交历史
func (h *Handler) GetProgrammingSubmissions(c *gin.Context) {
	logger := utils.GetLogger()

	homeworkIDStr := c.Query("homeworkId")
	questionIDStr := c.Query("questionId")

	homeworkID, err := strconv.ParseUint(homeworkIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "homeworkId参数格式错误"))
		return
	}

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

	// 先查询 homework_question 表，获取 problem_id
	var homeworkQuestion model.HomeworkQuestion
	if err := db.Where("id = ?", questionID).First(&homeworkQuestion).Error; err != nil {
		logger.Error("查询作业题目失败", zap.Error(err), zap.Uint64("question_id", questionID))
		c.JSON(http.StatusOK, errorResponse(404, "题目不存在"))
		return
	}

	// 判断是编程题还是普通题目
	var submissions []model.HomeworkSubmit
	if homeworkQuestion.ProblemID != nil && *homeworkQuestion.ProblemID != "" {
		// 编程题：使用 problem_id 查询
		if err := db.Where("homework_id = ? AND problem_id = ? AND uid = ?",
			homeworkID, homeworkQuestion.ProblemID, uid.(string)).
			Order("create_time DESC").
			Find(&submissions).Error; err != nil {
			logger.Error("查询编程题提交历史失败", zap.Error(err))
			c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
			return
		}
	} else {
		// 普通题目：使用 question_id 查询
		if err := db.Where("homework_id = ? AND question_id = ? AND uid = ?",
			homeworkID, homeworkQuestion.QuestionID, uid.(string)).
			Order("create_time DESC").
			Find(&submissions).Error; err != nil {
			logger.Error("查询题目提交历史失败", zap.Error(err))
			c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
			return
		}
	}

	// 解析答案中的代码和语言
	type ProgrammingSubmission struct {
		SubmitID   *uint64 `json:"submitId"`
		Code       string  `json:"code"`
		Language   string  `json:"language"`
		SubmitTime string  `json:"submitTime"`
		Result     string  `json:"result,omitempty"`
	}

	result := make([]ProgrammingSubmission, 0, len(submissions))
	for _, sub := range submissions {
		var answerData map[string]interface{}
		if err := json.Unmarshal([]byte(sub.Answer), &answerData); err == nil {
			// 安全地获取代码和语言
			code := ""
			if c, ok := answerData["code"].(string); ok {
				code = c
			}

			language := ""
			if l, ok := answerData["language"].(string); ok {
				language = l
			}

			// 获取评测结果
			resultStr := ""
			if sub.JudgeResult != "" {
				resultStr = sub.JudgeResult
			}

			// 使用 UpdatedAt 作为提交时间（因为可能会更新已有记录）
			submitTime := sub.UpdatedAt.Format("2006-01-02 15:04:05")

			logger.Debug("解析编程题提交记录",
				zap.Uint64("submit_id", *sub.SubmitID),
				zap.String("code_length", fmt.Sprintf("%d", len(code))),
				zap.String("language", language),
				zap.String("result", resultStr))

			result = append(result, ProgrammingSubmission{
				SubmitID:   sub.SubmitID,
				Code:       code,
				Language:   language,
				SubmitTime: submitTime,
				Result:     resultStr,
			})
		} else {
			logger.Warn("解析答案JSON失败", zap.Error(err), zap.String("answer", sub.Answer))
		}
	}

	logger.Info("返回编程题提交历史", zap.Int("count", len(result)))
	c.JSON(http.StatusOK, successResponse(result))
}

// ==================== 学生班级个人信息管理 ====================

// GetClassroomStudentInfo 获取学生在某个班级中的个人信息（学生）
func (h *Handler) GetClassroomStudentInfo(c *gin.Context) {
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

	// 查询学生在该班级的信息
	var student model.ClassroomStudent
	if err := db.Where("classroom_id = ? AND uid = ? AND status = 1", classroomID, uid.(string)).
		First(&student).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, errorResponse(404, "未找到班级信息或你不在该班级中"))
		} else {
			logger.Error("查询学生班级信息失败", zap.Error(err))
			c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		}
		return
	}

	c.JSON(http.StatusOK, successResponse(student))
}

// UpdateClassroomStudentInfo 更新学生在某个班级中的个人信息（学生）
func (h *Handler) UpdateClassroomStudentInfo(c *gin.Context) {
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

	var req struct {
		RealName     string `json:"realName" binding:"required"`
		Gender       string `json:"gender"`
		StudentClass string `json:"studentClass"`
		StudentNo    string `json:"studentNo"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误: "+err.Error()))
		return
	}

	db := client.GetDB()

	// 查询学生在该班级的信息
	var student model.ClassroomStudent
	if err := db.Where("classroom_id = ? AND uid = ? AND status = 1", classroomID, uid.(string)).
		First(&student).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, errorResponse(404, "未找到班级信息或你不在该班级中"))
		} else {
			logger.Error("查询学生班级信息失败", zap.Error(err))
			c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		}
		return
	}

	// 更新信息
	student.RealName = req.RealName
	student.Gender = req.Gender
	student.StudentClass = req.StudentClass
	student.StudentNo = req.StudentNo

	if err := db.Save(&student).Error; err != nil {
		logger.Error("更新学生班级信息失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "更新失败"))
		return
	}

	logger.Info("学生更新班级信息成功",
		zap.Uint64("classroomId", classroomID),
		zap.String("uid", uid.(string)),
		zap.String("realName", req.RealName))

	c.JSON(http.StatusOK, successResponse(student))
}

// UploadHomeworkAttachment 上传作业附件（学生）- 支持主观题上传图片
func (h *Handler) UploadHomeworkAttachment(c *gin.Context) {
	logger := utils.GetLogger()

	// 获取当前用户ID
	uid, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "未登录"))
		return
	}

	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		logger.Warn("获取上传文件失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "请选择要上传的文件"))
		return
	}

	// 验证文件类型（只允许图片）
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".bmp":  true,
		".webp": true,
	}

	if !allowedExts[ext] {
		c.JSON(http.StatusOK, errorResponse(400, "只支持上传图片文件（jpg, jpeg, png, gif, bmp, webp）"))
		return
	}

	// 验证文件大小（限制为10MB）
	const maxFileSize = 10 * 1024 * 1024 // 10MB
	if file.Size > maxFileSize {
		c.JSON(http.StatusOK, errorResponse(400, "图片文件大小不能超过10MB"))
		return
	}

	// 创建唯一文件名并保存到本地
	uniqueFileName := fmt.Sprintf("homework_%s_%d%s%s", uid.(string), time.Now().Unix(), generateRandomString(8), ext)

	// 创建上传目录
	uploadDir := "./uploads/classroom/homework"
	if err := c.SaveUploadedFile(file, fmt.Sprintf("%s/%s", uploadDir, uniqueFileName)); err != nil {
		logger.Error("保存图片失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "图片保存失败"))
		return
	}

	imageURL := "/uploads/classroom/homework/" + uniqueFileName

	logger.Info("上传作业附件成功", zap.String("filename", file.Filename), zap.String("url", imageURL), zap.String("uid", uid.(string)))

	c.JSON(http.StatusOK, successResponse(map[string]interface{}{
		"url":      imageURL,
		"filename": file.Filename,
		"size":     file.Size,
	}))
}

// GetHomeworkAnalysis 获取作业学情分析（教师）
func (h *Handler) GetHomeworkAnalysis(c *gin.Context) {
	logger := utils.GetLogger()
	homeworkIDStr := c.Param("homeworkId")
	homeworkID, err := strconv.ParseUint(homeworkIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "homeworkId参数格式错误"))
		return
	}

	db := client.GetDB()

	// 获取作业信息
	var homework model.ClassroomHomework
	if err := db.Where("id = ?", homeworkID).First(&homework).Error; err != nil {
		logger.Error("查询作业失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	// 获取作业的所有题目（按顺序）
	var homeworkQuestions []model.HomeworkQuestion
	if err := db.Where("homework_id = ?", homeworkID).
		Order("question_order ASC").
		Find(&homeworkQuestions).Error; err != nil {
		logger.Error("查询作业题目失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	// 获取班级的所有学生
	var classroomStudents []model.ClassroomStudent
	if err := db.Where("classroom_id = ? AND status = 1", homework.ClassroomID).
		Preload("User").
		Find(&classroomStudents).Error; err != nil {
		logger.Error("查询班级学生失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	// 构建学生信息映射
	studentMap := make(map[string]*model.ClassroomStudent)
	for i := range classroomStudents {
		studentMap[classroomStudents[i].UID] = &classroomStudents[i]
	}

	// 获取所有提交记录
	var submissions []model.HomeworkSubmit
	if err := db.Where("homework_id = ? AND is_officially_submitted = 1", homeworkID).
		Preload("Question").
		Find(&submissions).Error; err != nil {
		logger.Error("查询提交记录失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	// 统计提交和未提交学生
	submittedUIDs := make(map[string]bool)
	for _, submit := range submissions {
		submittedUIDs[submit.UID] = true
	}

	var submittedStudents []interface{}
	var unsubmittedStudents []interface{}

	for _, cs := range classroomStudents {
		username := ""
		if cs.User != nil {
			username = cs.User.Username
		}

		studentInfo := map[string]interface{}{
			"uid":      cs.UID,
			"realName": cs.RealName,
			"username": username,
		}
		if submittedUIDs[cs.UID] {
			submittedStudents = append(submittedStudents, studentInfo)
		} else {
			unsubmittedStudents = append(unsubmittedStudents, studentInfo)
		}
	}

	// 分析每道题目
	questionAnalysis := make([]map[string]interface{}, 0)

	for _, hq := range homeworkQuestions {
		analysis := map[string]interface{}{
			"homeworkQuestionId": hq.ID,
			"questionOrder":      hq.QuestionOrder,
			"score":              hq.Score,
		}

		// 编程题
		if hq.ProblemID != nil && *hq.ProblemID != "" {
			analysis["type"] = "programming"
			analysis["title"] = fmt.Sprintf("编程题 - %s", *hq.ProblemID)

			// 获取该题的所有提交
			var questionSubmissions []model.HomeworkSubmit
			for _, s := range submissions {
				if s.ProblemID != nil && *s.ProblemID == *hq.ProblemID {
					questionSubmissions = append(questionSubmissions, s)
				}
			}

			// 计算平均分
			totalScore := 0.0
			count := len(questionSubmissions)
			for _, s := range questionSubmissions {
				totalScore += s.Score
			}
			avgScore := 0.0
			if count > 0 {
				avgScore = totalScore / float64(count)
			}

			analysis["submittedCount"] = count
			analysis["avgScore"] = avgScore

			// 编程题：按分数段分布
			scoreSegments := []map[string]interface{}{
				{"label": "满分", "minScore": float64(hq.Score), "maxScore": float64(hq.Score), "count": 0, "selectedBy": []interface{}{}},
				{"label": "80%以上", "minScore": float64(hq.Score) * 0.8, "maxScore": float64(hq.Score), "count": 0, "selectedBy": []interface{}{}},
				{"label": "60%-80%", "minScore": float64(hq.Score) * 0.6, "maxScore": float64(hq.Score) * 0.8, "count": 0, "selectedBy": []interface{}{}},
				{"label": "60%以下", "minScore": 0.0, "maxScore": float64(hq.Score) * 0.6, "count": 0, "selectedBy": []interface{}{}},
			}

			for _, s := range questionSubmissions {
				score := s.Score

				for _, segment := range scoreSegments {
					minScore := segment["minScore"].(float64)
					maxScore := segment["maxScore"].(float64)

					if (score >= minScore && score < maxScore) || (score == maxScore && score == float64(hq.Score)) {
						segment["count"] = segment["count"].(int) + 1
						if cs, exists := studentMap[s.UID]; exists {
							username := ""
							if cs.User != nil {
								username = cs.User.Username
							}
							segment["selectedBy"] = append(segment["selectedBy"].([]interface{}), map[string]interface{}{
								"uid":      s.UID,
								"realName": cs.RealName,
								"username": username,
								"score":    s.Score,
							})
						}
						break
					}
				}
			}

			// 计算百分比并生成选项数据
			optionStats := make([]map[string]interface{}, 0)
			for _, segment := range scoreSegments {
				selectedCount := segment["count"].(int)
				percentage := 0.0
				if count > 0 {
					percentage = float64(selectedCount) / float64(count) * 100
				}

				optionStats = append(optionStats, map[string]interface{}{
					"label":         segment["label"],
					"selectedCount": selectedCount,
					"percentage":    percentage,
					"selectedBy":    segment["selectedBy"],
				})
			}

			analysis["options"] = optionStats
			analysis["scoreSegments"] = true // 标记这是分数段分布

			// 编程题的提交学生列表
			var submittedBy []interface{}
			submittedUIDs := make(map[string]bool)
			for _, s := range questionSubmissions {
				submittedUIDs[s.UID] = true
				if cs, exists := studentMap[s.UID]; exists {
					username := ""
					if cs.User != nil {
						username = cs.User.Username
					}
					submittedBy = append(submittedBy, map[string]interface{}{
						"uid":      s.UID,
						"realName": cs.RealName,
						"username": username,
						"score":    s.Score,
					})
				}
			}
			analysis["submittedBy"] = submittedBy

			// 编程题的未提交学生列表
			var unsubmittedBy []interface{}
			for _, cs := range classroomStudents {
				if !submittedUIDs[cs.UID] {
					username := ""
					if cs.User != nil {
						username = cs.User.Username
					}
					unsubmittedBy = append(unsubmittedBy, map[string]interface{}{
						"uid":      cs.UID,
						"realName": cs.RealName,
						"username": username,
					})
				}
			}
			analysis["unsubmittedBy"] = unsubmittedBy
			analysis["unsubmittedCount"] = len(unsubmittedBy)
		} else if hq.QuestionID != nil {
			// 普通题目（从题库）
			var question model.QuestionBank
			if err := db.Where("id = ?", *hq.QuestionID).First(&question).Error; err != nil {
				logger.Error("查询题目失败", zap.Error(err), zap.Uint64("question_id", *hq.QuestionID))
				continue
			}

			analysis["type"] = question.Type
			analysis["title"] = question.Title
			analysis["questionId"] = question.ID

			// 添加正确答案（仅用于教师端分析，学生端不应该调用此接口）
			// 注意：如果答案为空字符串，表示教师未设置参考答案
			if question.Answer != "" {
				analysis["answer"] = question.Answer
			} else {
				analysis["answer"] = ""
			}

			// 获取该题的所有提交
			var questionSubmissions []model.HomeworkSubmit
			for _, s := range submissions {
				if s.QuestionID != nil && *s.QuestionID == *hq.QuestionID {
					questionSubmissions = append(questionSubmissions, s)
				}
			}

			// 计算平均分
			totalScore := 0.0
			count := len(questionSubmissions)
			for _, s := range questionSubmissions {
				totalScore += s.Score
			}
			avgScore := 0.0
			if count > 0 {
				avgScore = totalScore / float64(count)
			}

			analysis["submittedCount"] = count
			analysis["avgScore"] = avgScore

			// 对所有题目类型生成扇形图数据
			if question.Type == "single_choice" || question.Type == "judge" || question.Type == "multiple_choice" {
				// 选择题：按选项分布
				var options []map[string]interface{}

				// 对于判断题，生成默认选项
				if question.Type == "judge" {
					options = []map[string]interface{}{
						{"label": "对", "content": "正确"},
						{"label": "错", "content": "错误"},
					}
				} else if question.Options != nil && *question.Options != "" {
					// 单选题和多选题从数据库读取选项
					if err := json.Unmarshal([]byte(*question.Options), &options); err != nil {
						logger.Error("解析题目选项失败", zap.Error(err), zap.String("options", *question.Options))
						// 如果解析失败，生成默认选项
						options = []map[string]interface{}{
							{"label": "A", "content": "选项A"},
							{"label": "B", "content": "选项B"},
							{"label": "C", "content": "选项C"},
							{"label": "D", "content": "选项D"},
						}
					}
				} else {
					// 如果选项为空，生成默认选项
					logger.Warn("题目选项为空，使用默认选项", zap.String("type", question.Type), zap.Uint64("question_id", question.ID))
					options = []map[string]interface{}{
						{"label": "A", "content": "选项A"},
						{"label": "B", "content": "选项B"},
						{"label": "C", "content": "选项C"},
						{"label": "D", "content": "选项D"},
					}
				}

				// 生成选项统计数据
				if len(options) > 0 {
					optionStats := make([]map[string]interface{}, 0)
					for _, opt := range options {
						optLabel := opt["label"].(string)

						// 统计选择该选项的学生
						var selectedBy []interface{}
						selectedCount := 0

						logger.Info("开始统计选项", zap.String("question_type", question.Type), zap.String("option_label", optLabel), zap.Int("submissions", len(questionSubmissions)))

						for _, s := range questionSubmissions {
							// 处理答案：支持多种格式
							// 1. JSON数组格式：["A", "B"] 或 ["对"] 或 ["true"] 或 [true]
							// 2. 简单字符串格式："A" 或 "对" 或 "true"
							// 3. 布尔值格式：true 或 false
							var studentAnswers []string

							// 首先尝试解析为JSON数组（可能包含字符串或布尔值）
							var rawAnswers []interface{}
							if err := json.Unmarshal([]byte(s.Answer), &rawAnswers); err == nil {
								// 成功解析为数组，转换为字符串数组
								for _, rawAns := range rawAnswers {
									switch v := rawAns.(type) {
									case string:
										studentAnswers = append(studentAnswers, v)
									case bool:
										studentAnswers = append(studentAnswers, fmt.Sprintf("%v", v))
									case float64:
										studentAnswers = append(studentAnswers, fmt.Sprintf("%.0f", v))
									}
								}
							} else {
								// 如果解析失败，尝试当作简单字符串处理
								if s.Answer != "" {
									// 检查是否是布尔值字符串
									if s.Answer == "true" || s.Answer == "false" {
										studentAnswers = []string{s.Answer}
									} else {
										studentAnswers = []string{s.Answer}
									}
								}
							}

							logger.Info("学生答案解析", zap.String("uid", s.UID), zap.String("raw_answer", s.Answer), zap.Any("parsed_answers", studentAnswers), zap.Int("parsed_count", len(studentAnswers)))

							// 检查学生的答案中是否包含该选项
							for _, ans := range studentAnswers {
								logger.Info("比较答案", zap.String("option_label", optLabel), zap.String("student_answer", ans), zap.String("question_type", question.Type))
								// 答案存储格式为 "A", "B", "C", "对", "错", "true", "false" 等
								// 对于判断题，支持多种格式：对/正确/true, 错/错误/false
								matched := false
								if question.Type == "judge" {
									// 判断题：支持多种格式匹配
									if optLabel == "对" && (ans == "对" || ans == "正确" || ans == "true") {
										matched = true
									} else if optLabel == "错" && (ans == "错" || ans == "错误" || ans == "false") {
										matched = true
									}
								} else {
									// 其他题型：直接匹配
									if ans == optLabel {
										matched = true
									}
								}

								if matched {
									selectedCount++
									logger.Info("找到匹配", zap.String("option_label", optLabel), zap.String("student_answer", ans))
									if cs, exists := studentMap[s.UID]; exists {
										username := ""
										if cs.User != nil {
											username = cs.User.Username
										}
										selectedBy = append(selectedBy, map[string]interface{}{
											"uid":      s.UID,
											"realName": cs.RealName,
											"username": username,
											"score":    s.Score,
										})
									}
									// 找到该学生选择了此选项后，跳出内层循环（避免同一个学生重复计数）
									break
								}
							}
						}

						logger.Info("选项统计结果", zap.String("option_label", optLabel), zap.Int("selected_count", selectedCount))

						percentage := 0.0
						if count > 0 {
							percentage = float64(selectedCount) / float64(count) * 100
						}

						optionStats = append(optionStats, map[string]interface{}{
							"label":         optLabel,
							"content":       opt["content"],
							"selectedCount": selectedCount,
							"percentage":    percentage,
							"selectedBy":    selectedBy,
						})
					}

					analysis["options"] = optionStats
				} else {
					logger.Warn("题目没有选项", zap.String("type", question.Type), zap.Uint64("question_id", question.ID))
				}
			} else {
				// 主观题和编程题：按分数段分布
				scoreSegments := []map[string]interface{}{
					{"label": "满分", "minScore": float64(hq.Score), "maxScore": float64(hq.Score), "count": 0, "selectedBy": []interface{}{}},
					{"label": "80%以上", "minScore": float64(hq.Score) * 0.8, "maxScore": float64(hq.Score), "count": 0, "selectedBy": []interface{}{}},
					{"label": "60%-80%", "minScore": float64(hq.Score) * 0.6, "maxScore": float64(hq.Score) * 0.8, "count": 0, "selectedBy": []interface{}{}},
					{"label": "60%以下", "minScore": 0.0, "maxScore": float64(hq.Score) * 0.6, "count": 0, "selectedBy": []interface{}{}},
				}

				for _, s := range questionSubmissions {
					score := s.Score

					for _, segment := range scoreSegments {
						minScore := segment["minScore"].(float64)
						maxScore := segment["maxScore"].(float64)

						if (score >= minScore && score < maxScore) || (score == maxScore && score == float64(hq.Score)) {
							segment["count"] = segment["count"].(int) + 1
							if cs, exists := studentMap[s.UID]; exists {
								username := ""
								if cs.User != nil {
									username = cs.User.Username
								}
								segment["selectedBy"] = append(segment["selectedBy"].([]interface{}), map[string]interface{}{
									"uid":      s.UID,
									"realName": cs.RealName,
									"username": username,
									"score":    s.Score,
								})
							}
							break
						}
					}
				}

				// 计算百分比并添加到analysis
				optionStats := make([]map[string]interface{}, 0)
				for _, segment := range scoreSegments {
					selectedCount := segment["count"].(int)
					percentage := 0.0
					if count > 0 {
						percentage = float64(selectedCount) / float64(count) * 100
					}

					optionStats = append(optionStats, map[string]interface{}{
						"label":         segment["label"],
						"selectedCount": selectedCount,
						"percentage":    percentage,
						"selectedBy":    segment["selectedBy"],
					})
				}

				analysis["options"] = optionStats
				analysis["scoreSegments"] = true // 标记这是分数段分布
			}

			// 提交学生列表（所有题目类型）
			var submittedBy []interface{}
			for _, s := range questionSubmissions {
				logger.Info("处理提交学生", zap.String("question_type", question.Type), zap.String("uid", s.UID), zap.String("raw_answer", s.Answer))

				if cs, exists := studentMap[s.UID]; exists {
					username := ""
					if cs.User != nil {
						username = cs.User.Username
					}
					studentData := map[string]interface{}{
						"uid":      s.UID,
						"realName": cs.RealName,
						"username": username,
						"score":    s.Score,
					}

					// 对于选择题，添加选择的选项
					if question.Type == "single_choice" || question.Type == "judge" || question.Type == "multiple_choice" {
						var studentAnswer []string
						// 首先尝试解析为JSON数组
						if err := json.Unmarshal([]byte(s.Answer), &studentAnswer); err == nil {
							studentData["answer"] = studentAnswer
							logger.Info("解析答案成功(JSON数组)", zap.String("uid", s.UID), zap.Any("answer", studentAnswer))
						} else {
							// 如果解析失败，当作简单字符串处理
							if s.Answer != "" {
								studentAnswer = []string{s.Answer}
								studentData["answer"] = studentAnswer
								logger.Info("解析答案成功(简单字符串)", zap.String("uid", s.UID), zap.Any("answer", studentAnswer))
							} else {
								// 答案为空（学生未作答），设置为空数组，以便前端显示"未作答"
								studentData["answer"] = []string{}
								logger.Info("答案为空（未作答）", zap.String("uid", s.UID))
							}
						}
					}

					submittedBy = append(submittedBy, studentData)
				}
			}
			analysis["submittedBy"] = submittedBy

			// 客观题的未提交学生列表
			submittedUIDs := make(map[string]bool)
			for _, s := range questionSubmissions {
				submittedUIDs[s.UID] = true
			}

			var unsubmittedBy []interface{}
			for _, cs := range classroomStudents {
				if !submittedUIDs[cs.UID] {
					username := ""
					if cs.User != nil {
						username = cs.User.Username
					}
					unsubmittedBy = append(unsubmittedBy, map[string]interface{}{
						"uid":      cs.UID,
						"realName": cs.RealName,
						"username": username,
					})
				}
			}
			analysis["unsubmittedBy"] = unsubmittedBy
			analysis["unsubmittedCount"] = len(unsubmittedBy)
		}

		questionAnalysis = append(questionAnalysis, analysis)
	}

	// 构建返回数据
	result := map[string]interface{}{
		"totalStudentCount":   len(classroomStudents),
		"submittedCount":      len(submittedStudents),
		"unsubmittedCount":    len(unsubmittedStudents),
		"submittedStudents":   submittedStudents,
		"unsubmittedStudents": unsubmittedStudents,
		"questionAnalysis":    questionAnalysis,
	}

	logger.Info("获取学情分析成功",
		zap.Uint64("homework_id", homeworkID),
		zap.Int("total_students", len(classroomStudents)),
		zap.Int("submitted_count", len(submittedStudents)))

	c.JSON(http.StatusOK, successResponse(result))
}

// GetMaterialPermissions 获取某个资料的所有学生权限设置（教师）
func (h *Handler) GetMaterialPermissions(c *gin.Context) {
	logger := utils.GetLogger()

	materialIDStr := c.Param("materialId")
	materialID, err := strconv.ParseUint(materialIDStr, 10, 64)
	if err != nil {
		logger.Error("materialId参数格式错误", zap.String("materialId", materialIDStr), zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "materialId参数格式错误"))
		return
	}

	db := client.GetDB()

	// 检查资料是否存在
	var material model.ClassroomMaterial
	if err := db.Where("id = ? AND status = 1", materialID).First(&material).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Error("资料不存在", zap.Uint64("material_id", materialID))
			c.JSON(http.StatusOK, errorResponse(404, "资料不存在"))
		} else {
			logger.Error("查询资料失败", zap.Uint64("material_id", materialID), zap.Error(err))
			c.JSON(http.StatusOK, errorResponse(500, "查询资料失败，请稍后重试"))
		}
		return
	}

	logger.Info("开始获取资料权限", zap.Uint64("material_id", materialID), zap.String("file_name", material.FileName))

	// 获取班级ID：支持根目录(folder_id=0)的情况
	var classroomID uint64
	if material.FolderID == 0 {
		// 根目录资料，直接使用资料的班级ID
		classroomID = material.ClassroomID
		logger.Info("资料位于根目录",
			zap.Uint64("material_id", materialID),
			zap.Uint64("classroom_id", classroomID))
	} else {
		// 查询文件夹获取班级ID
		var folder model.ClassroomFolder
		if err := db.Where("id = ?", material.FolderID).First(&folder).Error; err != nil {
			logger.Error("查询文件夹失败", zap.Uint64("folder_id", material.FolderID), zap.Error(err))
			c.JSON(http.StatusOK, errorResponse(500, "查询文件夹失败，请稍后重试"))
			return
		}
		classroomID = folder.ClassroomID
		logger.Info("通过文件夹获取班级ID",
			zap.Uint64("material_id", materialID),
			zap.Uint64("folder_id", material.FolderID),
			zap.Uint64("classroom_id", classroomID))
	}

	// 获取班级的所有学生
	var students []model.ClassroomStudent
	if err := db.Where("classroom_id = ? AND status = 1", classroomID).
		Preload("User").
		Find(&students).Error; err != nil {
		logger.Error("查询学生列表失败",
			zap.Uint64("classroom_id", classroomID),
			zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询学生列表失败，请稍后重试"))
		return
	}

	logger.Info("查询学生列表成功",
		zap.Uint64("classroom_id", classroomID),
		zap.Int("student_count", len(students)))

	// 如果没有学生，记录详细信息
	if len(students) == 0 {
		logger.Warn("班级中没有学生",
			zap.Uint64("classroom_id", classroomID),
			zap.Uint64("material_id", materialID))
	}

	// 获取该资料的所有权限设置
	var permissions []model.ClassroomMaterialPermission
	if err := db.Where("material_id = ?", materialID).Find(&permissions).Error; err != nil {
		logger.Error("查询权限设置失败", zap.Uint64("material_id", materialID), zap.Error(err))
		// 检查是否是表不存在错误
		if strings.Contains(err.Error(), "Table") && strings.Contains(err.Error(), "doesn't exist") {
			logger.Error("权限表不存在，请先执行数据库迁移", zap.Error(err))
			c.JSON(http.StatusOK, errorResponse(500, "权限表不存在，请联系管理员"))
		} else {
			logger.Error("查询权限记录失败", zap.Uint64("material_id", materialID), zap.Error(err))
			c.JSON(http.StatusOK, errorResponse(500, "查询权限设置失败，请稍后重试"))
		}
		return
	}

	// 构建权限映射
	permissionMap := make(map[string]*model.ClassroomMaterialPermission)
	for i := range permissions {
		permissionMap[permissions[i].StudentUID] = &permissions[i]
	}

	// 构建返回数据
	type StudentPermission struct {
		UID         string `json:"uid"`
		RealName    string `json:"realName"`
		Username    string `json:"username"`
		CanPreview  bool   `json:"canPreview"`
		CanDownload bool   `json:"canDownload"`
	}

	result := make([]StudentPermission, 0, len(students))
	for _, student := range students {
		username := ""
		if student.User != nil {
			username = student.User.Username
		}

		perm := StudentPermission{
			UID:         student.UID,
			RealName:    student.RealName,
			Username:    username,
			CanPreview:  false,
			CanDownload: false,
		}

		if p, exists := permissionMap[student.UID]; exists {
			perm.CanPreview = p.CanPreview == 1
			perm.CanDownload = p.CanDownload == 1
		}

		result = append(result, perm)
	}

	logger.Info("获取资料权限设置成功",
		zap.Uint64("material_id", materialID),
		zap.String("file_name", material.FileName),
		zap.Int("student_count", len(result)),
		zap.Int("permission_count", len(permissions)))
	c.JSON(http.StatusOK, successResponse(result))
}

// SetMaterialPermissions 批量设置资料权限（教师）
func (h *Handler) SetMaterialPermissions(c *gin.Context) {
	logger := utils.GetLogger()

	var req struct {
		MaterialID  uint64                   `json:"materialId" binding:"required"`
		Permissions []map[string]interface{} `json:"permissions" binding:"required"` // [{uid, canPreview, canDownload}]
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	db := client.GetDB()

	// 检查资料是否存在
	var material model.ClassroomMaterial
	if err := db.Where("id = ? AND status = 1", req.MaterialID).First(&material).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, errorResponse(404, "资料不存在"))
		} else {
			logger.Error("查询资料失败", zap.Error(err))
			c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		}
		return
	}

	// 开始事务
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 删除旧权限
	if err := tx.Where("material_id = ?", req.MaterialID).Delete(&model.ClassroomMaterialPermission{}).Error; err != nil {
		logger.Error("删除旧权限失败", zap.Error(err))
		tx.Rollback()
		c.JSON(http.StatusOK, errorResponse(500, "设置失败"))
		return
	}

	// 批量创建新权限
	for _, perm := range req.Permissions {
		uid, _ := perm["uid"].(string)
		canPreview, _ := perm["canPreview"].(bool)
		canDownload, _ := perm["canDownload"].(bool)

		if uid == "" {
			continue
		}

		permission := &model.ClassroomMaterialPermission{
			MaterialID:  req.MaterialID,
			StudentUID:  uid,
			CanPreview:  0,
			CanDownload: 0,
		}

		if canPreview {
			permission.CanPreview = 1
		}
		if canDownload {
			permission.CanDownload = 1
		}

		if err := tx.Create(permission).Error; err != nil {
			logger.Error("创建权限失败", zap.Error(err))
			tx.Rollback()
			c.JSON(http.StatusOK, errorResponse(500, "设置失败"))
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		logger.Error("提交事务失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "设置失败"))
		return
	}

	logger.Info("批量设置资料权限成功", zap.Uint64("material_id", req.MaterialID), zap.Int("count", len(req.Permissions)))
	c.JSON(http.StatusOK, successResponse(nil))
}

// BatchSetAllMaterialPermissions 批量设置所有学生权限（教师）
func (h *Handler) BatchSetAllMaterialPermissions(c *gin.Context) {
	logger := utils.GetLogger()

	materialIDStr := c.Param("materialId")
	materialID, err := strconv.ParseUint(materialIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "materialId参数格式错误"))
		return
	}

	var req struct {
		CanPreview  bool `json:"canPreview"`
		CanDownload bool `json:"canDownload"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	db := client.GetDB()

	// 检查资料是否存在
	var material model.ClassroomMaterial
	if err := db.Where("id = ? AND status = 1", materialID).First(&material).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, errorResponse(404, "资料不存在"))
		} else {
			logger.Error("查询资料失败", zap.Error(err))
			c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		}
		return
	}

	// 获取文件夹所在的班级
	var folder model.ClassroomFolder
	if err := db.Where("id = ?", material.FolderID).First(&folder).Error; err != nil {
		logger.Error("查询文件夹失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	// 获取班级的所有学生
	var students []model.ClassroomStudent
	if err := db.Where("classroom_id = ? AND status = 1", folder.ClassroomID).
		Select("uid").
		Find(&students).Error; err != nil {
		logger.Error("查询学生列表失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	// 开始事务
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 删除旧权限
	if err := tx.Where("material_id = ?", materialID).Delete(&model.ClassroomMaterialPermission{}).Error; err != nil {
		logger.Error("删除旧权限失败", zap.Error(err))
		tx.Rollback()
		c.JSON(http.StatusOK, errorResponse(500, "设置失败"))
		return
	}

	// 为所有学生创建权限
	canPreviewVal := 0
	canDownloadVal := 0
	if req.CanPreview {
		canPreviewVal = 1
	}
	if req.CanDownload {
		canDownloadVal = 1
	}

	for _, student := range students {
		permission := &model.ClassroomMaterialPermission{
			MaterialID:  materialID,
			StudentUID:  student.UID,
			CanPreview:  canPreviewVal,
			CanDownload: canDownloadVal,
		}

		if err := tx.Create(permission).Error; err != nil {
			logger.Error("创建权限失败", zap.Error(err))
			tx.Rollback()
			c.JSON(http.StatusOK, errorResponse(500, "设置失败"))
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		logger.Error("提交事务失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "设置失败"))
		return
	}

	action := "全部关闭"
	if req.CanPreview || req.CanDownload {
		action = "全部开启"
	}

	logger.Info("批量设置所有学生权限成功",
		zap.Uint64("material_id", materialID),
		zap.String("action", action),
		zap.Int("student_count", len(students)))

	c.JSON(http.StatusOK, successResponse(nil))
}

// checkUserMaterialPermission 检查用户对资料的权限
func (h *Handler) checkUserMaterialPermission(db *gorm.DB, materialID uint64, uid string) (canPreview, canDownload bool, err error) {
	// 查询权限
	var permission model.ClassroomMaterialPermission
	err = db.Where("material_id = ? AND student_uid = ?", materialID, uid).First(&permission).Error

	if err == gorm.ErrRecordNotFound {
		// 没有权限记录，默认不可访问
		return false, false, nil
	}

	if err != nil {
		return false, false, err
	}

	return permission.CanPreview == 1, permission.CanDownload == 1, nil
}

// DownloadMaterial 下载资料文件（带权限验证）
func (h *Handler) DownloadMaterial(c *gin.Context) {
	logger := utils.GetLogger()

	materialIDStr := c.Param("materialId")
	materialID, err := strconv.ParseUint(materialIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "materialId参数格式错误"))
		return
	}

	db := client.GetDB()

	// 获取当前用户ID
	uid, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "未登录"))
		return
	}

	// 查询资料信息
	var material model.ClassroomMaterial
	if err := db.Where("id = ? AND status = 1", materialID).First(&material).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, errorResponse(404, "资料不存在"))
		} else {
			logger.Error("查询资料失败", zap.Error(err))
			c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		}
		return
	}

	// 检查用户是否为上传者（上传者始终有权限）
	if material.CreatorID != uid.(string) {
		// 检查下载权限
		_, canDownload, err := h.checkUserMaterialPermission(db, materialID, uid.(string))
		if err != nil {
			logger.Error("检查权限失败", zap.Error(err))
			c.JSON(http.StatusOK, errorResponse(500, "权限检查失败"))
			return
		}

		if !canDownload {
			logger.Warn("用户无下载权限",
				zap.Uint64("material_id", materialID),
				zap.String("uid", uid.(string)))
			c.JSON(http.StatusOK, errorResponse(403, "您没有下载该资料的权限"))
			return
		}
	}

	// 构建文件路径
	filePath := "." + material.FilePath
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		logger.Error("文件不存在", zap.String("path", filePath))
		c.JSON(http.StatusOK, errorResponse(404, "文件不存在"))
		return
	}

	// 设置响应头
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename=\""+material.FileName+"\"")
	c.Header("Content-Type", "application/octet-stream")
	c.File(filePath)

	logger.Info("资料下载成功",
		zap.Uint64("material_id", materialID),
		zap.String("filename", material.FileName),
		zap.String("uid", uid.(string)))
}

// GetMaterialPDFBase64 获取PDF文件的base64编码（用于前端PDF.js渲染）
// @deprecated 使用 GetMaterialPDFBinary 替代，性能更好
func (h *Handler) GetMaterialPDFBase64(c *gin.Context) {
	logger := utils.GetLogger()

	materialIDStr := c.Param("materialId")
	materialID, err := strconv.ParseUint(materialIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "materialId参数格式错误"))
		return
	}

	// 获取当前用户ID
	uid, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "未登录"))
		return
	}

	db := client.GetDB()

	// 查询资料信息
	var material model.ClassroomMaterial
	if err := db.Where("id = ? AND status = 1", materialID).First(&material).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Warn("PDF API: 资料不存在", zap.Uint64("material_id", materialID))
			c.JSON(http.StatusOK, errorResponse(404, "资料不存在"))
		} else {
			logger.Error("PDF API: 查询资料失败", zap.Error(err), zap.Uint64("material_id", materialID))
			c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		}
		return
	}
	logger.Info("PDF API: 查询资料成功", zap.Uint64("material_id", materialID), zap.String("filename", material.FileName), zap.String("file_path", material.FilePath))

	// 检查文件类型是否为PDF
	if !strings.HasSuffix(strings.ToLower(material.FileName), ".pdf") {
		logger.Warn("PDF API: 不是PDF文件", zap.String("filename", material.FileName))
		c.JSON(http.StatusOK, errorResponse(400, "该文件不是PDF文件"))
		return
	}

	// 检查用户角色 - 班级教师和管理员自动拥有所有资料的预览权限
	var classroomRoles []model.ClassroomUserRole
	if err := db.Where("uid = ?", uid.(string)).Find(&classroomRoles).Error; err != nil {
		logger.Error("PDF API: 查询用户角色失败", zap.Error(err), zap.String("uid", uid.(string)))
		c.JSON(http.StatusOK, errorResponse(500, "查询用户角色失败"))
		return
	}

	// 提取角色列表
	roleList := make([]string, len(classroomRoles))
	for i, r := range classroomRoles {
		roleList[i] = r.Role
	}

	// 检查是否为教师或管理员
	isTeacherOrAdmin := false
	for _, role := range roleList {
		if role == "teacher" || role == "admin" {
			isTeacherOrAdmin = true
			break
		}
	}

	// 同时检查 HOJ 系统管理员角色
	hojRoles, err := middlewarepkg.GetUserRoles(db, uid.(string))
	if err == nil {
		if middlewarepkg.HasRole(hojRoles, middlewarepkg.RoleRoot) ||
			middlewarepkg.HasRole(hojRoles, middlewarepkg.RoleAdmin) {
			isTeacherOrAdmin = true
		}
	}

	if !isTeacherOrAdmin {
		// 学生需要检查权限记录
		// 优先检查全局权限（student_uid IS NULL），如果没有再检查用户特定权限
		var permission model.ClassroomMaterialPermission
		err = db.Where("material_id = ? AND student_uid IS NULL", materialID).
			First(&permission).Error

		// 如果没有全局权限，检查用户特定权限
		if err != nil {
			err = db.Where("material_id = ? AND student_uid = ?", materialID, uid.(string)).
				First(&permission).Error
		}

		if err != nil {
			logger.Warn("PDF API: 学生没有权限记录", zap.Uint64("material_id", materialID), zap.String("uid", uid.(string)), zap.Error(err))
			c.JSON(http.StatusOK, errorResponse(403, "您没有权限查看该资料"))
			return
		}

		if permission.CanPreview != 1 {
			logger.Warn("PDF API: 学生没有预览权限", zap.Uint64("material_id", materialID), zap.String("uid", uid.(string)), zap.Int("can_preview", permission.CanPreview))
			c.JSON(http.StatusOK, errorResponse(403, "您没有预览权限"))
			return
		}
		logger.Info("PDF API: 学生权限检查通过", zap.Uint64("material_id", materialID), zap.String("uid", uid.(string)))
	} else {
		logger.Info("PDF API: 教师或管理员访问，跳过权限检查",
			zap.Uint64("material_id", materialID),
			zap.String("filename", material.FileName),
			zap.String("uid", uid.(string)),
			zap.Strings("classroom_roles", roleList),
			zap.Strings("hoj_roles", hojRoles))
	}

	// 构建文件路径
	filePath := "." + material.FilePath
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		logger.Error("文件不存在", zap.String("path", filePath))
		c.JSON(http.StatusOK, errorResponse(404, "文件不存在"))
		return
	}

	// 读取文件内容
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		logger.Error("读取文件失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "读取文件失败"))
		return
	}

	// 转换为base64
	base64Data := make([]byte, base64.StdEncoding.EncodedLen(len(fileData)))
	base64.StdEncoding.Encode(base64Data, fileData)

	// 返回JSON响应
	c.JSON(http.StatusOK, successResponse(map[string]interface{}{
		"fileName": material.FileName,
		"data":     string(base64Data),
		"size":     len(fileData),
	}))

	logger.Info("PDF base64获取成功",
		zap.Uint64("material_id", materialID),
		zap.String("filename", material.FileName),
		zap.String("uid", uid.(string)),
		zap.Int("size", len(fileData)))
}

// GetMaterialPDFBinary 获取PDF文件的二进制数据（性能更好，推荐使用）
func (h *Handler) GetMaterialPDFBinary(c *gin.Context) {
	logger := utils.GetLogger()

	materialIDStr := c.Param("materialId")
	materialID, err := strconv.ParseUint(materialIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "materialId参数格式错误"))
		return
	}

	// 获取当前用户ID
	uid, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "未登录"))
		return
	}

	db := client.GetDB()

	// 查询资料信息
	var material model.ClassroomMaterial
	if err := db.Where("id = ? AND status = 1", materialID).First(&material).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Warn("PDF Binary API: 资料不存在", zap.Uint64("material_id", materialID))
			c.JSON(http.StatusOK, errorResponse(404, "资料不存在"))
		} else {
			logger.Error("PDF Binary API: 查询资料失败", zap.Error(err), zap.Uint64("material_id", materialID))
			c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		}
		return
	}

	// 检查文件类型是否为PDF
	if !strings.HasSuffix(strings.ToLower(material.FileName), ".pdf") {
		c.JSON(http.StatusOK, errorResponse(400, "该文件不是PDF文件"))
		return
	}

	// 检查用户角色 - 班级教师和管理员自动拥有所有资料的预览权限
	var classroomRoles []model.ClassroomUserRole
	if err := db.Where("uid = ?", uid.(string)).Find(&classroomRoles).Error; err != nil {
		logger.Error("PDF Binary API: 查询用户角色失败", zap.Error(err), zap.String("uid", uid.(string)))
		c.JSON(http.StatusOK, errorResponse(500, "查询用户角色失败"))
		return
	}

	// 提取角色列表
	roleList := make([]string, len(classroomRoles))
	for i, r := range classroomRoles {
		roleList[i] = r.Role
	}

	// 检查是否为教师或管理员
	isTeacherOrAdmin := false
	for _, role := range roleList {
		if role == "teacher" || role == "admin" {
			isTeacherOrAdmin = true
			break
		}
	}

	// 同时检查 HOJ 系统管理员角色
	hojRoles, err := middlewarepkg.GetUserRoles(db, uid.(string))
	if err == nil {
		if middlewarepkg.HasRole(hojRoles, middlewarepkg.RoleRoot) ||
			middlewarepkg.HasRole(hojRoles, middlewarepkg.RoleAdmin) {
			isTeacherOrAdmin = true
		}
	}

	if !isTeacherOrAdmin {
		// 学生需要检查权限记录
		var permission model.ClassroomMaterialPermission
		err = db.Where("material_id = ? AND student_uid IS NULL", materialID).
			First(&permission).Error

		if err != nil {
			err = db.Where("material_id = ? AND student_uid = ?", materialID, uid.(string)).
				First(&permission).Error
		}

		if err != nil {
			logger.Warn("PDF Binary API: 学生没有权限记录",
				zap.Uint64("material_id", materialID),
				zap.String("uid", uid.(string)))
			c.JSON(http.StatusOK, errorResponse(403, "您没有权限查看该资料"))
			return
		}

		if permission.CanPreview != 1 {
			logger.Warn("PDF Binary API: 学生没有预览权限",
				zap.Uint64("material_id", materialID),
				zap.String("uid", uid.(string)))
			c.JSON(http.StatusOK, errorResponse(403, "您没有预览权限"))
			return
		}
	} else {
		logger.Info("PDF Binary API: 教师或管理员访问，跳过权限检查",
			zap.Uint64("material_id", materialID),
			zap.String("filename", material.FileName),
			zap.String("uid", uid.(string)),
			zap.Strings("classroom_roles", roleList),
			zap.Strings("hoj_roles", hojRoles))
	}

	// 构建文件路径
	filePath := "." + material.FilePath
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		logger.Error("文件不存在", zap.String("path", filePath))
		c.JSON(http.StatusOK, errorResponse(404, "文件不存在"))
		return
	}

	// 直接返回 PDF 文件
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", material.FileName))
	c.Header("Cache-Control", "private, max-age=3600")
	c.Header("Accept-Ranges", "bytes")

	c.File(filePath)

	logger.Info("PDF Binary获取成功",
		zap.Uint64("material_id", materialID),
		zap.String("filename", material.FileName),
		zap.String("uid", uid.(string)))
}

// ==================== PPT预览功能 - Office Online支持 ====================

// 临时预览令牌结构
type PreviewToken struct {
	MaterialID uint64 `json:"materialId"`
	UID        string `json:"uid"`
	ExpireAt   int64  `json:"expireAt"`
}

// 生成预览令牌（使用HMAC-SHA256签名）
func generatePreviewToken(materialID uint64, uid string, secret string) string {
	expireAt := time.Now().Add(5 * time.Minute).Unix()

	// 构建令牌数据
	data := fmt.Sprintf("%d|%s|%d", materialID, uid, expireAt)

	// 使用HMAC-SHA256签名
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	signature := hex.EncodeToString(h.Sum(nil))

	// 组合数据和签名
	tokenData := fmt.Sprintf("%s|%s", data, signature)

	// Base64编码
	return base64.URLEncoding.EncodeToString([]byte(tokenData))
}

// 验证预览令牌
func verifyPreviewToken(token string, secret string) (*PreviewToken, error) {
	// Base64解码
	tokenBytes, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token format")
	}

	tokenStr := string(tokenBytes)
	parts := strings.Split(tokenStr, "|")
	if len(parts) != 4 {
		return nil, fmt.Errorf("invalid token structure")
	}

	// 解析数据
	materialID, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid material ID")
	}

	uid := parts[1]
	expireAt, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid expire time")
	}

	signature := parts[3]

	// 检查过期时间
	if time.Now().Unix() > expireAt {
		return nil, fmt.Errorf("token expired")
	}

	// 重新计算签名验证
	data := fmt.Sprintf("%d|%s|%d", materialID, uid, expireAt)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	if signature != expectedSignature {
		return nil, fmt.Errorf("invalid token signature")
	}

	return &PreviewToken{
		MaterialID: materialID,
		UID:        uid,
		ExpireAt:   expireAt,
	}, nil
}

// GenerateMaterialPreviewToken 生成资料预览临时令牌
func (h *Handler) GenerateMaterialPreviewToken(c *gin.Context) {
	logger := utils.GetLogger()

	materialIDStr := c.Param("materialId")
	materialID, err := strconv.ParseUint(materialIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "materialId参数格式错误"))
		return
	}

	// 获取当前用户ID
	uid, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "未登录"))
		return
	}

	db := client.GetDB()

	// 查询资料信息
	var material model.ClassroomMaterial
	if err := db.Where("id = ? AND status = 1", materialID).First(&material).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, errorResponse(404, "资料不存在"))
		} else {
			logger.Error("查询资料失败", zap.Error(err))
			c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		}
		return
	}

	// 检查是否为Office文件
	fileType := strings.ToLower(strings.TrimPrefix(filepath.Ext(material.FileName), "."))
	if fileType != "ppt" && fileType != "pptx" && fileType != "doc" &&
		fileType != "docx" && fileType != "xls" && fileType != "xlsx" {
		c.JSON(http.StatusOK, errorResponse(400, "该文件类型不支持此预览方式"))
		return
	}

	// 检查用户角色 - 班级教师和管理员自动拥有所有资料的预览权限
	var classroomRoles []model.ClassroomUserRole
	if err := db.Where("uid = ?", uid.(string)).Find(&classroomRoles).Error; err != nil {
		logger.Error("查询用户角色失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询用户角色失败"))
		return
	}

	// 提取角色列表
	roleList := make([]string, len(classroomRoles))
	for i, r := range classroomRoles {
		roleList[i] = r.Role
	}

	// 检查是否为教师或管理员
	isTeacherOrAdmin := false
	for _, role := range roleList {
		if role == "teacher" || role == "admin" {
			isTeacherOrAdmin = true
			break
		}
	}

	// 同时检查 HOJ 系统管理员角色
	hojRoles, err := middlewarepkg.GetUserRoles(db, uid.(string))
	if err == nil {
		if middlewarepkg.HasRole(hojRoles, middlewarepkg.RoleRoot) ||
			middlewarepkg.HasRole(hojRoles, middlewarepkg.RoleAdmin) {
			isTeacherOrAdmin = true
		}
	}

	if !isTeacherOrAdmin {
		// 学生需要检查权限记录
		var permission model.ClassroomMaterialPermission
		err = db.Where("material_id = ? AND student_uid IS NULL", materialID).
			First(&permission).Error

		// 如果没有全局权限，检查用户特定权限
		if err != nil {
			err = db.Where("material_id = ? AND student_uid = ?", materialID, uid.(string)).
				First(&permission).Error
		}

		if err != nil {
			logger.Warn("学生没有权限记录",
				zap.Uint64("material_id", materialID),
				zap.String("uid", uid.(string)))
			c.JSON(http.StatusOK, errorResponse(403, "您没有权限查看该资料"))
			return
		}

		if permission.CanPreview != 1 {
			logger.Warn("学生没有预览权限",
				zap.Uint64("material_id", materialID),
				zap.String("uid", uid.(string)))
			c.JSON(http.StatusOK, errorResponse(403, "您没有预览权限"))
			return
		}
	}

	// 生成临时预览令牌
	secret := "histoj-preview-token-secret" // TODO: 从配置文件读取
	token := generatePreviewToken(materialID, uid.(string), secret)

	// 构建预览URL（使用相对路径，Office Online会完整访问）
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	host := c.Request.Host
	previewURL := fmt.Sprintf("%s://%s/api/classroom/material/preview/%s", scheme, host, token)

	logger.Info("生成预览令牌成功",
		zap.Uint64("material_id", materialID),
		zap.String("uid", uid.(string)))

	c.JSON(http.StatusOK, successResponse(gin.H{
		"previewUrl": previewURL,
		"token":      token,
	}))
}

// PreviewMaterialWithToken 使用临时令牌预览资料（供Office Online等服务访问）
func (h *Handler) PreviewMaterialWithToken(c *gin.Context) {
	logger := utils.GetLogger()

	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing token"})
		return
	}

	// 验证令牌
	secret := "histoj-preview-token-secret" // TODO: 从配置文件读取
	previewToken, err := verifyPreviewToken(token, secret)
	if err != nil {
		logger.Warn("预览令牌验证失败", zap.String("token", token), zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		return
	}

	db := client.GetDB()

	// 查询资料信息
	var material model.ClassroomMaterial
	if err := db.Where("id = ? AND status = 1", previewToken.MaterialID).First(&material).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Material not found"})
		} else {
			logger.Error("查询资料失败", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Query failed"})
		}
		return
	}

	// 构建文件路径
	filePath := "." + material.FilePath
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		logger.Error("文件不存在", zap.String("path", filePath))
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// 获取文件扩展名并设置正确的Content-Type
	ext := strings.ToLower(filepath.Ext(material.FileName))
	var contentType string
	switch ext {
	case ".ppt", ".pptx":
		contentType = "application/vnd.openxmlformats-officedocument.presentationml.presentation"
		if ext == ".ppt" {
			contentType = "application/vnd.ms-powerpoint"
		}
	case ".doc", ".docx":
		contentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
		if ext == ".doc" {
			contentType = "application/msword"
		}
	case ".xls", ".xlsx":
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
		if ext == ".xls" {
			contentType = "application/vnd.ms-excel"
		}
	default:
		contentType = "application/octet-stream"
	}

	// 设置安全响应头
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", material.FileName))
	c.Header("Cache-Control", "public, max-age=3600") // 增加缓存时间到1小时
	c.Header("X-Content-Type-Options", "nosniff")
	// 允许 Office Online 跨域访问
	c.Header("X-Frame-Options", "ALLOW-FROM https://view.officeapps.live.com")
	c.Header("Content-Security-Policy", "frame-ancestors 'self' https://view.officeapps.live.com")
	// 支持断点续传，加快大文件传输
	c.Header("Accept-Ranges", "bytes")

	// 返回文件
	c.File(filePath)

	logger.Info("预览文件访问成功",
		zap.Uint64("material_id", material.ID),
		zap.String("filename", material.FileName),
		zap.String("uid", previewToken.UID))
}

// ==================== 腾讯云COS文档预览 ====================

// getFileType 根据文件名获取文件类型
func getFileType(fileName string) string {
	ext := strings.ToLower(fileName[strings.LastIndex(fileName, ".")+1:])

	videoTypes := []string{"mp4", "webm", "ogv", "mov", "avi", "mkv", "flv", "m4v"}
	audioTypes := []string{"mp3", "wav", "aac", "ogg", "m4a", "flac"}
	imageTypes := []string{"jpg", "jpeg", "png", "gif", "bmp", "svg", "webp"}
	documentTypes := []string{"pdf", "txt", "md", "doc", "docx", "xls", "xlsx", "ppt", "pptx"}

	if containsItem(videoTypes, ext) {
		return "video"
	} else if containsItem(audioTypes, ext) {
		return "audio"
	} else if containsItem(imageTypes, ext) {
		return "image"
	} else if containsItem(documentTypes, ext) {
		return "document"
	}

	return "unknown"
}

// isDocumentType 判断是否为文档类型（需要使用腾讯云文档预览）
func isDocumentType(fileType string) bool {
	return fileType == "document"
}

// containsItem 检查切片是否包含元素
func containsItem(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// GetCOSPreviewUrl 获取腾讯云COS文档预览URL
func (h *Handler) GetCOSPreviewUrl(c *gin.Context) {
	logger := utils.GetLogger()

	materialIDStr := c.Param("materialId")
	materialID, err := strconv.ParseUint(materialIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "materialId参数格式错误"))
		return
	}

	db := client.GetDB()

	// 查询资料信息
	var material model.ClassroomMaterial
	if err := db.Where("id = ? AND status = 1", materialID).First(&material).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, errorResponse(404, "文件不存在"))
		} else {
			logger.Error("查询资料失败", zap.Error(err))
			c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		}
		return
	}

	// 初始化COS服务
	cosService, err := service.NewCOSService()
	if err != nil {
		logger.Error("初始化COS服务失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "COS服务初始化失败"))
		return
	}

	logger.Info("COS配置信息",
		zap.String("bucket", config.GlobalConfig.COS.Bucket),
		zap.String("region", config.GlobalConfig.COS.Region))

	// 构建COS文件路径（按文件夹ID组织）
	cosPath := fmt.Sprintf("classroom/materials/%d/%d_%s", material.FolderID, material.ID, material.FileName)

	logger.Info("准备处理COS文件",
		zap.Uint64("material_id", material.ID),
		zap.String("file_name", material.FileName),
		zap.Uint64("folder_id", material.FolderID),
		zap.String("cos_path", cosPath))

	// 检查文件是否已在COS
	exists, err := cosService.IsFileExists(cosPath)
	var fileURL string

	if !exists {
		// 文件不在COS，需要上传
		localPath := "." + material.FilePath
		if _, err := os.Stat(localPath); os.IsNotExist(err) {
			logger.Error("本地文件不存在", zap.String("path", localPath))
			c.JSON(http.StatusOK, errorResponse(404, "文件不存在"))
			return
		}

		fileURL, err := cosService.UploadFile(localPath, cosPath)
		if err != nil {
			logger.Error("上传到COS失败", zap.Error(err), zap.String("localPath", localPath))
			c.JSON(http.StatusOK, errorResponse(500, "上传到COS失败"))
			return
		}

		logger.Info("文件已上传到COS",
			zap.Uint64("material_id", material.ID),
			zap.String("cos_path", cosPath),
			zap.String("cos_url", fileURL))
	} else {
		// 文件已在COS，直接使用公共URL
		fileURL = fmt.Sprintf("https://%s.cos.%s.myqcloud.com/%s",
			config.GlobalConfig.COS.Bucket,
			config.GlobalConfig.COS.Region,
			cosPath)
		logger.Info("文件已在COS，直接使用",
			zap.Uint64("material_id", material.ID),
			zap.String("cos_path", cosPath))
	}

	// 根据文件类型生成不同的预览URL
	fileType := getFileType(material.FileName)
	var previewURL string

	if isDocumentType(fileType) {
		// 文档类型：使用腾讯云文档预览
		previewURL = cosService.GetDocPreviewURL(fileURL)
		logger.Info("使用文档预览URL",
			zap.String("file_type", fileType),
			zap.String("preview_url", previewURL))
	} else if fileType == "video" {
		// 视频类型：不支持在线预览
		logger.Info("视频文件不支持在线预览",
			zap.Uint64("material_id", material.ID),
			zap.String("file_name", material.FileName))
		c.JSON(http.StatusOK, errorResponse(400, "视频文件不支持在线预览，请下载后观看"))
		return
	} else {
		// 音频、图片：直接使用原始COS URL
		previewURL = fileURL
		logger.Info("使用原始URL直接预览",
			zap.String("file_type", fileType),
			zap.String("direct_url", previewURL))
	}

	logger.Info("返回预览URL",
		zap.String("file_type", fileType),
		zap.String("preview_url", previewURL),
		zap.String("cos_url", fileURL),
		zap.Bool("uploaded", !exists))

	c.JSON(http.StatusOK, successResponse(gin.H{
		"previewUrl": previewURL,
		"cosUrl":     fileURL,
		"fileType":   fileType,
		"uploaded":   !exists,
	}))
}
