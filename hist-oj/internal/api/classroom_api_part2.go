package api

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/hoj/hist-oj/internal/client"
	"github.com/hoj/hist-oj/internal/model"
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

// ==================== 作业/考试功能 ====================

// CreateHomework 创建作业（教师）
func (h *Handler) CreateHomework(c *gin.Context) {
	logger := utils.GetLogger()

	type HomeworkQuestionItem struct {
		QuestionID uint64 `json:"questionId" binding:"required"`
		Score      int    `json:"score"`
	}

	var req struct {
		ClassroomID  uint64                 `json:"classroomId" binding:"required"`
		Title        string                 `json:"title" binding:"required"`
		Description  string                 `json:"description"`
		StartTime    string                 `json:"startTime" binding:"required"` // RFC3339 format
		EndTime      string                 `json:"endTime" binding:"required"`   // RFC3339 format
		ShowScore    int                    `json:"showScore"`
		ShowHomework int                    `json:"showHomework"`
		Questions    []HomeworkQuestionItem `json:"questions" binding:"required"`
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
		ClassroomID:  req.ClassroomID,
		Title:        req.Title,
		Description:  req.Description,
		StartTime:    startTime,
		EndTime:      endTime,
		ShowScore:    req.ShowScore,
		ShowHomework: req.ShowHomework,
		Status:       1,
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
			score = 2 // 默认2分
		}

		homeworkQuestion := &model.HomeworkQuestion{
			HomeworkID:    homework.ID,
			QuestionID:    q.QuestionID,
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

// UpdateHomework 更新作业（教师）
func (h *Handler) UpdateHomework(c *gin.Context) {
	logger := utils.GetLogger()

	var req struct {
		ID          uint64 `json:"id" binding:"required"`
		ClassroomID uint64 `json:"classroomId"`
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
		StartTime   string `json:"startTime" binding:"required"`
		EndTime     string `json:"endTime" binding:"required"`
		ShowHomework int    `json:"showHomework"`
		ShowScore    int    `json:"showScore"`
		Questions   []struct {
			QuestionID uint64 `json:"questionId" binding:"required"`
			Score       int    `json:"score" binding:"required"`
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
			score = 10 // 默认分值
		}

		homeworkQuestion := &model.HomeworkQuestion{
			HomeworkID:    req.ID,
			QuestionID:    q.QuestionID,
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
			var submitCount int64
			db.Model(&model.HomeworkSubmit{}).
				Where("homework_id = ? AND uid = ? AND is_officially_submitted = 1",
					homeworks[i].ID, uid.(string)).
				Count(&submitCount)

			// 获取该作业的总题目数
			var questionCount int64
			db.Model(&model.HomeworkQuestion{}).
				Where("homework_id = ?", homeworks[i].ID).
				Count(&questionCount)

			// 判断是否已完成（提交的题目数等于总题目数）
			homeworks[i].IsCompleted = submitCount > 0 && int(submitCount) == int(questionCount)
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

	db := client.GetDB()
	var homework model.ClassroomHomework

	if err := db.Where("id = ?", homeworkID).
		Preload("Questions").
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

	// 更新作业状态
	now := time.Now()
	if now.Before(homework.StartTime) {
		homework.Status = 1 // 未开始
	} else if now.After(homework.EndTime) {
		homework.Status = 3 // 已结束
	} else {
		homework.Status = 2 // 进行中
	}

	c.JSON(http.StatusOK, successResponse(homework))
}

// SaveHomeworkDraft 保存作业草稿（学生）- 自动保存
// 与 SubmitHomework 的区别：不设置 IsOfficiallySubmitted 标志
// 请求格式：{ homeworkId, answers: { questionId: answer, ... } }
func (h *Handler) SaveHomeworkDraft(c *gin.Context) {
	logger := utils.GetLogger()

	var req struct {
		HomeworkID uint64            `json:"homeworkId" binding:"required"`
		Answers    map[string]string `json:"answers" binding:"required"` // questionId -> answer
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

		// 查找是否已有草稿记录
		var existingSubmit model.HomeworkSubmit
		checkErr := tx.Where("homework_id = ? AND question_id = ? AND uid = ?",
			req.HomeworkID, questionID, uid.(string)).
			First(&existingSubmit).Error

		if checkErr == nil {
			// 更新已有记录（保持原有的 IsOfficiallySubmitted 状态）
			existingSubmit.Answer = answer
			// 草稿保存不修改分数和提交状态

			if err := tx.Save(&existingSubmit).Error; err != nil {
				logger.Error("更新作业草稿失败", zap.Error(err))
				tx.Rollback()
				c.JSON(http.StatusOK, errorResponse(500, "保存草稿失败"))
				return
			}
		} else {
			// 创建新记录（草稿状态）
			submit := &model.HomeworkSubmit{
				HomeworkID:          req.HomeworkID,
				QuestionID:          questionID,
				UID:                 uid.(string),
				Answer:              answer,
				Score:               0, // 草稿不判分
				IsScored:            0,
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
// 请求格式：{ homeworkId, answers: { questionId: answer, ... } }
func (h *Handler) SubmitHomework(c *gin.Context) {
	logger := utils.GetLogger()

	var req struct {
		HomeworkID uint64            `json:"homeworkId" binding:"required"`
		Answers    map[string]string `json:"answers" binding:"required"` // questionId -> answer
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

		// 计算分数（单选、多选、判断题自动判分）
		score := 0.0
		isScored := 0
		if question.Type == "single_choice" || question.Type == "multiple_choice" || question.Type == "judge" {
			// 对于单选和判断题，直接比较字符串
			if question.Type == "single_choice" || question.Type == "judge" {
				if answer == question.Answer {
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
			existingSubmit.Score = score
			existingSubmit.IsScored = isScored
			existingSubmit.IsOfficiallySubmitted = 1 // 标记为正式提交

			if err := tx.Save(&existingSubmit).Error; err != nil {
				logger.Error("更新作业提交失败", zap.Error(err))
				tx.Rollback()
				c.JSON(http.StatusOK, errorResponse(500, "提交失败"))
				return
			}
		} else {
			// 创建新提交
			submit := &model.HomeworkSubmit{
				HomeworkID:          req.HomeworkID,
				QuestionID:          questionID,
				UID:                 uid.(string),
				Answer:              answer,
				Score:               score,
				IsScored:            isScored,
				IsOfficiallySubmitted: 1, // 标记为正式提交
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
	var submissions []model.HomeworkSubmit

	if err := db.Where("homework_id = ?", homeworkID).
		Preload("Student").
		Preload("Question").
		Order("uid, question_id").
		Find(&submissions).Error; err != nil {
		logger.Error("查询作业提交失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	c.JSON(http.StatusOK, successResponse(submissions))
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
			if err := db.Where("homework_id = ? AND question_id = ? AND uid = ?",
				homeworkID, question.QuestionID, student.UID).
				First(&submit).Error; err == nil {
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
	var submitTime *time.Time
	totalScore := 0.0
	gradedScore := 0.0 // 已评分题目的得分
	// 检查是否已正式提交（只要有一条记录标记为已正式提交，就认为整个作业已提交）
	isOfficiallySubmitted := false
	hasUngraded := false // 是否有未评分的题目

	for _, submission := range submissions {
		// 将题目ID转为字符串作为key
		questionIDStr := strconv.FormatUint(submission.QuestionID, 10)
		answersMap[questionIDStr] = submission.Answer
		scoresMap[questionIDStr] = submission.Score
		isScoredMap[questionIDStr] = submission.IsScored == 1

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

	result := map[string]interface{}{
		"answers":               string(answersJSON),
		"scores":                string(scoresJSON),         // 每题得分
		"isScoredMap":           string(isScoredJSON),      // 每题评分状态
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
	logger := utils.GetLogger()

	var req struct {
		ClassroomID uint64 `json:"classroomId" binding:"required"`
		FolderName  string `json:"folderName" binding:"required"`
		ParentID    uint64 `json:"parentId"`
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

	db := client.GetDB()

	folder := &model.ClassroomFolder{
		ClassroomID: req.ClassroomID,
		FolderName:  req.FolderName,
		ParentID:    req.ParentID,
		CreatorID:   creatorID.(string),
		Status:      1,
	}

	if err := db.Create(folder).Error; err != nil {
		logger.Error("创建文件夹失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "创建失败"))
		return
	}

	logger.Info("创建文件夹", zap.Uint64("id", folder.ID), zap.String("name", folder.FolderName))
	c.JSON(http.StatusOK, successResponse(folder))
}

// GetFolders 获取文件夹列表（教师/学生）
func (h *Handler) GetFolders(c *gin.Context) {
	logger := utils.GetLogger()
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
		logger.Error("查询文件夹列表失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	c.JSON(http.StatusOK, successResponse(folders))
}

// DeleteFolder 删除文件夹（教师）
func (h *Handler) DeleteFolder(c *gin.Context) {
	logger := utils.GetLogger()
	folderIDStr := c.Param("folderId")
	folderID, err := strconv.ParseUint(folderIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "folderId参数格式错误"))
		return
	}

	db := client.GetDB()

	// 软删除
	result := db.Model(&model.ClassroomFolder{}).
		Where("id = ?", folderID).
		Update("status", 0)

	if result.Error != nil {
		logger.Error("删除文件夹失败", zap.Error(result.Error))
		c.JSON(http.StatusOK, errorResponse(500, "删除失败"))
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusOK, errorResponse(404, "文件夹不存在"))
		return
	}

	logger.Info("删除文件夹", zap.Uint64("folder_id", folderID))
	c.JSON(http.StatusOK, successResponse(nil))
}

// UploadMaterial 上传资料（教师）
func (h *Handler) UploadMaterial(c *gin.Context) {
	logger := utils.GetLogger()

	folderIDStr := c.PostForm("folderId")
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
		logger.Warn("获取上传文件失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "请选择要上传的文件"))
		return
	}

	// 创建唯一文件名
	uniqueFileName := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)

	// 保存文件到本地
	uploadDir := "./uploads/classroom"
	if err := c.SaveUploadedFile(file, fmt.Sprintf("%s/%s", uploadDir, uniqueFileName)); err != nil {
		logger.Error("保存文件失败", zap.Error(err))
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
		logger.Error("上传资料失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "上传失败"))
		return
	}

	logger.Info("上传资料", zap.Uint64("id", material.ID), zap.String("name", material.FileName))
	c.JSON(http.StatusOK, successResponse(material))
}

// GetMaterials 获取资料列表（教师/学生）
func (h *Handler) GetMaterials(c *gin.Context) {
	logger := utils.GetLogger()
	folderIDStr := c.Param("folderId")
	folderID, err := strconv.ParseUint(folderIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "folderId参数格式错误"))
		return
	}

	db := client.GetDB()
	var materials []model.ClassroomMaterial

	if err := db.Where("folder_id = ? AND status = 1", folderID).
		Preload("Creator").
		Order("create_time DESC").
		Find(&materials).Error; err != nil {
		logger.Error("查询资料列表失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	c.JSON(http.StatusOK, successResponse(materials))
}

// DeleteMaterial 删除资料（教师）
func (h *Handler) DeleteMaterial(c *gin.Context) {
	logger := utils.GetLogger()
	materialIDStr := c.Param("materialId")
	materialID, err := strconv.ParseUint(materialIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "materialId参数格式错误"))
		return
	}

	db := client.GetDB()

	// 软删除
	result := db.Model(&model.ClassroomMaterial{}).
		Where("id = ?", materialID).
		Update("status", 0)

	if result.Error != nil {
		logger.Error("删除资料失败", zap.Error(result.Error))
		c.JSON(http.StatusOK, errorResponse(500, "删除失败"))
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusOK, errorResponse(404, "资料不存在"))
		return
	}

	logger.Info("删除资料", zap.Uint64("material_id", materialID))
	c.JSON(http.StatusOK, successResponse(nil))
}

// CopyMaterialToClassroom 复制资料到班级（教师）
func (h *Handler) CopyMaterialToClassroom(c *gin.Context) {
	logger := utils.GetLogger()

	var req struct {
		MaterialID  uint64 `json:"materialId" binding:"required"`
		FolderID    uint64 `json:"folderId" binding:"required"`
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
		logger.Error("复制资料失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "复制失败"))
		return
	}

	logger.Info("复制资料到班级", zap.Uint64("material_id", req.MaterialID), zap.Uint64("folder_id", req.FolderID))
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
		UID        string `json:"uid" binding:"required"`
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
	if question.Type == "single_choice" || question.Type == "judge" {
		if submit.Answer == question.Answer {
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

// RecallMessage 撤回消息（教师）
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

	// 验证权限：只有发送者本人可以撤回自己的消息
	if message.SenderID != uid.(string) {
		c.JSON(http.StatusOK, errorResponse(403, "无权撤回此消息"))
		return
	}

	// 删除消息
	if err := db.Delete(&message).Error; err != nil {
		logger.Error("撤回消息失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "撤回失败"))
		return
	}

	logger.Info("撤回消息", zap.Uint64("message_id", messageID))
	c.JSON(http.StatusOK, successResponse(nil))
}

// ClearMessages 清屏（删除当前用户在该班级的所有消息）
func (h *Handler) ClearMessages(c *gin.Context) {
	logger := utils.GetLogger()

	classroomIDStr := c.Param("classroomId")
	classroomID, err := strconv.ParseUint(classroomIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "班级ID格式错误"))
		return
	}

	db := client.GetDB()

	// 获取当前用户ID
	uid, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "未登录"))
		return
	}

	// 验证用户是否是该班级的教师或学生
	var classroom model.Classroom
	if err := db.Where("id = ?", classroomID).First(&classroom).Error; err != nil {
		c.JSON(http.StatusOK, errorResponse(404, "班级不存在"))
		return
	}

	// 只删除当前用户在该班级发送的消息
	if err := db.Where("classroom_id = ? AND sender_id = ?", classroomID, uid.(string)).
		Delete(&model.ClassroomMessage{}).Error; err != nil {
		logger.Error("清屏失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "清屏失败"))
		return
	}

	logger.Info("清屏", zap.Uint64("classroom_id", classroomID), zap.String("user_id", uid.(string)))
	c.JSON(http.StatusOK, successResponse(nil))
}
