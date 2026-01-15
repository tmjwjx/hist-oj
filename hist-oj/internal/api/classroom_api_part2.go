package api

import (
	"bytes"
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
		QuestionID   *uint64 `json:"questionId"`   // 题库题目ID(可选)
		ProblemID    *string `json:"problemId"`    // HOJ题目ID(可选,编程题使用,字符串类型支持"0001"等格式)
		QuestionType string  `json:"questionType"` // 题目类型(可选,用于设置默认分数)
		Score        int     `json:"score"`        // 分值(可选,默认根据题型设置)
	}

	var req struct {
		ClassroomID  uint64                 `json:"classroomId" binding:"required"`
		Title        string                 `json:"title" binding:"required"`
		Description  string                 `json:"description"`
		StartTime    string                 `json:"startTime" binding:"required"` // RFC3339 format
		EndTime      string                 `json:"endTime" binding:"required"`   // RFC3339 format
		ShowScore    int                    `json:"showScore"`
		ShowHomework int                    `json:"showHomework"`
		ShowAnswer   int                    `json:"showAnswer"`
		Questions    []HomeworkQuestionItem `json:"questions" binding:"required"`
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
		ClassroomID:  req.ClassroomID,
		Title:        req.Title,
		Description:  req.Description,
		StartTime:    startTime,
		EndTime:      endTime,
		ShowScore:    req.ShowScore,
		ShowHomework: req.ShowHomework,
		ShowAnswer:   req.ShowAnswer,
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
			questionIDPtr = q.QuestionID
		}

		homeworkQuestion := &model.HomeworkQuestion{
			HomeworkID:    homework.ID,
			QuestionID:    questionIDPtr,  // 使用指针,编程题时为nil
			ProblemID:     q.ProblemID,    // HOJ题目ID
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
		Questions    []struct {
			QuestionID   *uint64 `json:"questionId"`   // 题库题目ID(可选)
			ProblemID    *string `json:"problemId"`    // HOJ题目ID(可选,编程题使用,字符串类型)
			QuestionType string  `json:"questionType"` // 题目类型(可选,用于设置默认分数)
			Score        int     `json:"score"`        // 分值(可选,默认根据题型设置)
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
			questionIDPtr = q.QuestionID
		}

		homeworkQuestion := &model.HomeworkQuestion{
			HomeworkID:    req.ID,
			QuestionID:    questionIDPtr,  // 使用指针,编程题时为nil
			ProblemID:     q.ProblemID,    // HOJ题目ID
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
			// 统计已作答的题目数（使用 DISTINCT 去重，统计 question_id 或 problem_id）
			var submittedQuestionCount int64
			db.Raw(`
				SELECT COUNT(DISTINCT CASE
					WHEN question_id IS NOT NULL THEN question_id
					WHEN problem_id IS NOT NULL THEN problem_id
					ELSE NULL
				END) as count
				FROM homework_submit
				WHERE homework_id = ? AND uid = ? AND is_officially_submitted = 1
			`, homeworks[i].ID, uid.(string)).Scan(&submittedQuestionCount)

			// 获取该作业的总题目数
			var questionCount int64
			db.Model(&model.HomeworkQuestion{}).
				Where("homework_id = ?", homeworks[i].ID).
				Count(&questionCount)

			// 判断是否已完成（已作答的题目数等于总题目数）
			homeworks[i].IsCompleted = submittedQuestionCount > 0 && int(submittedQuestionCount) >= int(questionCount)
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
		Preload("Questions", func(db *gorm.DB) *gorm.DB {
			return db.Order("question_order ASC")
		}).
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
			// 更新已有记录（保持原有的 IsOfficiallySubmitted 状态）
			existingSubmit.Answer = answer
			// 只有当 req.Attachments 不为 nil 且当前题目有附件数据时，才更新 attachment
			// 避免自动保存草稿时清空已上传的图片
			if req.Attachments != nil {
				existingSubmit.Attachment = attachment
			}
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
				HomeworkID:           req.HomeworkID,
				QuestionID:           &questionID, // 使用指针
				UID:                  uid.(string),
				Answer:               answer,
				Attachment:           attachment,
				Score:                0, // 草稿不判分
				IsScored:             0,
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
		Answers     map[string]string `json:"answers" binding:"required"`    // questionId -> answer
		Attachments map[string]string `json:"attachments"`                   // questionId -> attachment URLs (comma separated)
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
			existingSubmit.Attachment = attachment
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
				HomeworkID:           req.HomeworkID,
				QuestionID:           &questionID, // 使用指针
				UID:                  uid.(string),
				Answer:               answer,
				Attachment:           attachment,
				Score:                score,
				IsScored:             isScored,
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

	// 将该用户的所有编程题记录也标记为正式提交
	// 因为编程题是通过 SaveProgrammingSubmission 单独保存的
	if err := tx.Model(&model.HomeworkSubmit{}).
		Where("homework_id = ? AND uid = ? AND problem_id IS NOT NULL AND is_officially_submitted = 0",
			req.HomeworkID, uid.(string)).
		Update("is_officially_submitted", 1).Error; err != nil {
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
		RealName           string `json:"realName,omitempty"`            // 班级学生真实姓名
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
		"scores":                string(scoresJSON),         // 每题得分
		"isScoredMap":           string(isScoredJSON),      // 每题评分状态
		"attachments":            string(attachmentsJSON),    // 每题附件
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
		logger.Error("查询子文件夹失败", zap.Error(err))
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
		logger.Error("查询资料文件失败", zap.Error(err))
		tx.Rollback()
		c.JSON(http.StatusOK, errorResponse(500, "删除失败"))
		return
	}

	// 删除磁盘上的文件
	for _, material := range materials {
		if material.FilePath != "" {
			// 转换URL路径为文件系统路径
			filePath := "." + material.FilePath
			if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
				logger.Warn("删除资料文件失败", zap.String("path", filePath), zap.Error(err))
			}
		}
	}

	// 3. 软删除所有资料记录
	if err := tx.Model(&model.ClassroomMaterial{}).
		Where("folder_id IN ?", allFolderIDs).
		Update("status", 0).Error; err != nil {
		logger.Error("删除资料记录失败", zap.Error(err))
		tx.Rollback()
		c.JSON(http.StatusOK, errorResponse(500, "删除失败"))
		return
	}

	// 4. 软删除所有子文件夹
	if err := tx.Model(&model.ClassroomFolder{}).
		Where("id IN ?", allFolderIDs).
		Update("status", 0).Error; err != nil {
		logger.Error("删除文件夹失败", zap.Error(err))
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

	logger.Info("删除文件夹及关联内容", zap.Uint64("folder_id", folderID), zap.Int("files_deleted", len(materials)))
	c.JSON(http.StatusOK, successResponse(nil))
}

// UpdateFolder 更新文件夹名称（教师）
func (h *Handler) UpdateFolder(c *gin.Context) {
	logger := utils.GetLogger()

	var req struct {
		ID         uint64 `json:"id" binding:"required"`
		FolderName string `json:"folderName" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("请求参数错误", zap.Error(err))
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
			logger.Error("查询文件夹失败", zap.Error(err))
			c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		}
		return
	}

	// 更新文件夹名称
	if err := db.Model(&folder).Update("folder_name", req.FolderName).Error; err != nil {
		logger.Error("更新文件夹失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "更新失败"))
		return
	}

	logger.Info("更新文件夹", zap.Uint64("folder_id", req.ID), zap.String("name", req.FolderName))
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

	var folderID uint64
	var err error

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

	// 查询资料信息以获取文件路径
	var material model.ClassroomMaterial
	if err := db.Where("id = ?", materialID).First(&material).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, errorResponse(404, "资料不存在"))
		} else {
			logger.Error("查询资料失败", zap.Error(err))
			c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		}
		return
	}

	// 删除磁盘上的文件
	if material.FilePath != "" {
		// 转换URL路径为文件系统路径
		filePath := "." + material.FilePath
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			logger.Warn("删除资料文件失败", zap.String("path", filePath), zap.Error(err))
			// 继续删除数据库记录，即使文件删除失败
		} else if err == nil {
			logger.Info("成功删除资料文件", zap.String("path", filePath))
		}
	}

	// 软删除数据库记录
	result := db.Model(&model.ClassroomMaterial{}).
		Where("id = ?", materialID).
		Update("status", 0)

	if result.Error != nil {
		logger.Error("删除资料记录失败", zap.Error(result.Error))
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

// GradeProgrammingHomework 教师批改编程题
func (h *Handler) GradeProgrammingHomework(c *gin.Context) {
	logger := utils.GetLogger()

	var req struct {
		HomeworkID uint64  `json:"homeworkId" binding:"required"`
		ProblemID  string  `json:"problemId" binding:"required"`
		UID        string `json:"uid" binding:"required"`
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
		HomeworkID uint64  `json:"homeworkId" binding:"required"`
		ProblemID  string  `json:"problemId" binding:"required"` // 编程题ID
		SubmitID   interface{} `json:"submitId" binding:"required"` // 支持字符串或数字
		Code       string  `json:"code"`
		Language   string  `json:"language"`
		Token      string  `json:"token"` // BingOJ token
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
				// BingOJ 状态码定义（来自 app.py）：
				// status: -2=编译错误, -1=保留, 0=答案正确(AC), 1=答案错误(WA),
				//         -3=格式错误(PE), 2=时间超限(TLE), 3=内存超限(MLE),
				//         4=NO, 5=系统错误(SE), 6=等待中, 7=判题中, 8=部分正确(PC),
				//         9=提交中, 10=提交失败
				status := targetSubmission.Status

				// 详细日志：打印从 BingOJ 获取的提交信息
				logger.Info("从 BingOJ 获取到提交记录",
					zap.Uint64("submit_id", submitID),
					zap.Int("result", targetSubmission.Result),
					zap.Int("status", status),
					zap.String("score", fmt.Sprintf("%d", targetSubmission.Score)),
					zap.String("language", targetSubmission.Language),
					zap.String("celInfo", targetSubmission.CELInfo))

				// 检查是否还在判题中或提交中（status 6=等待中, 7=判题中, 9=提交中）
				if status == 6 || status == 7 || status == 9 {
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

				// 根据 status 判断评测结果
				switch status {
				case 0:
					// AC - 满分
					score = calculateQuestionScoreForProblem(req.HomeworkID, req.ProblemID)
					judgeResult = "AC"
				case -2:
					// 编译错误
					judgeResult = "CE"
					score = 0
				case -3:
					// 格式错误
					judgeResult = "PE"
					score = 0
				case 1:
					// 答案错误
					judgeResult = "WA"
					score = 0
				case 2:
					// 时间超限
					judgeResult = "TLE"
					score = 0
				case 3:
					// 内存超限
					judgeResult = "MLE"
					score = 0
				case 4:
					// NO
					judgeResult = "NO"
					score = 0
				case 5:
					// 系统错误
					judgeResult = "SE"
					score = 0
				case 8:
					// 部分正确 - 按比例计算分数
					fullScore := calculateQuestionScoreForProblem(req.HomeworkID, req.ProblemID)
					score = float64(targetSubmission.Score) / 100.0 * fullScore
					judgeResult = "PC"
				default:
					judgeResult = fmt.Sprintf("Unknown(%d)", status)
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

