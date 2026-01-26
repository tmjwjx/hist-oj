package api

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/hoj/hist-oj/internal/client"
	"github.com/hoj/hist-oj/internal/model"
	"github.com/hoj/hist-oj/internal/utils"
)

// ==================== 考试模式相关接口 ====================

// StartExamRequest 开始考试请求
type StartExamRequest struct {
	DeviceInfo  string `json:"deviceInfo"`
	BrowserInfo string `json:"browserInfo"`
}

// OrderMappingItem 题目顺序映射项
type OrderMappingItem struct {
	OriginalOrder int `json:"originalOrder"` // 教师设置的原始顺序
	DisplayOrder  int `json:"displayOrder"`  // 学生看到的显示顺序
}

// StartExam 开始考试（学生点击"开始答题"）
func (h *Handler) StartExam(c *gin.Context) {
	logger := utils.GetLogger()
	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "未登录"))
		return
	}

	homeworkID := c.Param("homeworkId")
	if homeworkID == "" {
		c.JSON(http.StatusOK, errorResponse(400, "作业ID不能为空"))
		return
	}

	// 转换 homeworkID 为 uint64
	homeworkIDUint64, err := strconv.ParseUint(homeworkID, 10, 64)
	if err != nil {
		logger.Error("作业ID转换失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "作业ID格式错误"))
		return
	}

	var req StartExamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	db := client.GetDB()

	// 1. 获取作业信息
	var homework model.ClassroomHomework
	if err := db.Where("id = ?", homeworkID).First(&homework).Error; err != nil {
		logger.Error("获取作业信息失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(404, "作业不存在"))
		return
	}

	// 2. 检查是否是考试模式
	if homework.IsExamMode != 1 {
		c.JSON(http.StatusOK, errorResponse(400, "该作业不是考试模式"))
		return
	}

	// 3. 检查考试时间
	now := time.Now()
	if now.Before(homework.StartTime) {
		c.JSON(http.StatusOK, errorResponse(400, "考试尚未开始"))
		return
	}
	if now.After(homework.EndTime) {
		c.JSON(http.StatusOK, errorResponse(400, "考试已结束"))
		return
	}

	// 4. 检查是否已经开始了
	var existingSubmit model.HomeworkSubmit
	err = db.Where("homework_id = ? AND uid = ?", homeworkID, uid).
		Where("exam_start_time IS NOT NULL").
		First(&existingSubmit).Error
	if err == nil {
		// 已经开始过考试，返回现有信息
		// 检查 ExamStartTime 是否为 nil
		if existingSubmit.ExamStartTime == nil {
			logger.Error("考试开始时间为空", zap.Uint64("homeworkId", homeworkIDUint64), zap.String("uid", uid.(string)))
			c.JSON(http.StatusOK, errorResponse(500, "考试数据异常"))
			return
		}

		// 计算考试结束时间：取（开始时间+考试时长）和（作业结束时间）的较小值
		examEndTimeByDuration := existingSubmit.ExamStartTime.Add(time.Duration(homework.ExamDuration) * time.Minute)
		examEndTime := examEndTimeByDuration
		if examEndTimeByDuration.After(homework.EndTime) {
			examEndTime = homework.EndTime
		}

		// 计算是否允许交卷
		elapsedMinutes := int(time.Since(*existingSubmit.ExamStartTime).Minutes())
		canSubmit := elapsedMinutes >= homework.AllowSubmitAfterMinutes && existingSubmit.IsOfficiallySubmitted == 0

		// 获取学生保存的题目顺序
		var studentOrder model.StudentQuestionOrder
		questions := []gin.H{}
		if err := db.Where("homework_id = ? AND uid = ?", homeworkIDUint64, uid.(string)).First(&studentOrder).Error; err == nil {
			// 找到了题目顺序映射，需要返回乱序的题目
			var homeworkQuestions []model.HomeworkQuestion
			if err := db.Where("homework_id = ?", homeworkIDUint64).
				Preload("Question").
				Order("question_order ASC").
				Find(&homeworkQuestions).Error; err == nil {
				// 解析顺序映射
				var orderMapping []OrderMappingItem
				json.Unmarshal([]byte(studentOrder.OrderMapping), &orderMapping)

				// 根据映射关系重新排序题目
				questions = make([]gin.H, len(homeworkQuestions))
				for _, mapping := range orderMapping {
					if mapping.DisplayOrder > 0 && mapping.DisplayOrder <= len(homeworkQuestions) {
						// 找到对应原始顺序的题目
						for _, q := range homeworkQuestions {
							if q.QuestionOrder == mapping.OriginalOrder {
								questionData := gin.H{
									"id":            q.ID,
									"homeworkId":    q.HomeworkID,
									"questionId":    q.QuestionID,
									"problemId":     q.ProblemID,
									"displayOrder":  mapping.DisplayOrder,
									"questionOrder": q.QuestionOrder,
									"score":         q.Score,
								}
								if q.Question != nil {
									questionData["question"] = q.Question
								}
								questions[mapping.DisplayOrder-1] = questionData
								break
							}
						}
					}
				}
			}
		}

		c.JSON(http.StatusOK, successResponse(gin.H{
			"examStartTime":    existingSubmit.ExamStartTime,
			"examEndTime":      examEndTime,
			"remainingSeconds": getRemainingSeconds(homework.ExamDuration, existingSubmit.ExamStartTime, homework.EndTime),
			"canSubmit":        canSubmit,
			"hasStarted":       true,
			"isSubmitted":      existingSubmit.IsOfficiallySubmitted == 1,
			"isForcedSubmit":   existingSubmit.IsForcedSubmit == 1,
			"questions":        questions,
			"examConfig": gin.H{
				"examDuration":            homework.ExamDuration,
				"allowSubmitAfterMinutes": homework.AllowSubmitAfterMinutes,
				"disableCopyPaste":        homework.DisableCopyPaste,
				"requireFullscreen":       homework.RequireFullscreen,
				"disallowTabSwitch":       homework.DisallowTabSwitch,
			},
		}))
		return
	} else if err != gorm.ErrRecordNotFound {
		logger.Error("查询考试记录失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	// 5. 获取作业题目（预加载Question关联）
	var questions []model.HomeworkQuestion
	if err := db.Where("homework_id = ?", homeworkID).
		Preload("Question"). // 预加载题库信息
		Order("question_order ASC").
		Find(&questions).Error; err != nil {
		logger.Error("获取题目失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "获取题目失败"))
		return
	}

	// 6. 生成题目乱序并保存映射
	shuffledQuestions, orderMapping := shuffleQuestions(questions)

	// 保存题目顺序映射（使用 FirstOrCreate 避免重复）
	orderMappingJSON, err := json.Marshal(orderMapping)
	if err != nil {
		logger.Error("JSON序列化失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "数据处理失败"))
		return
	}

	// 使用 FirstOrCreate 处理学生多次点击的情况
	studentOrder := &model.StudentQuestionOrder{
		HomeworkID:   homeworkIDUint64,
		UID:          uid.(string),
		OrderMapping: string(orderMappingJSON),
	}
	if err := db.Where("homework_id = ? AND uid = ?", homeworkIDUint64, uid.(string)).
		Assign(studentOrder).
		FirstOrCreate(studentOrder).Error; err != nil {
		logger.Error("保存题目顺序失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "保存题目顺序失败"))
		return
	}

	// 7. 记录考试开始时间
	// 首先检查是否有草稿记录
	var submitCount int64
	db.Model(&model.HomeworkSubmit{}).
		Where("homework_id = ? AND uid = ?", homeworkID, uid).
		Count(&submitCount)

	examStartTime := now

	if submitCount == 0 {
		// 没有草稿记录，为每个题目创建一个空记录，标记考试开始时间
		var homeworkQuestions []model.HomeworkQuestion
		if err := db.Where("homework_id = ?", homeworkIDUint64).
			Find(&homeworkQuestions).Error; err != nil {
			logger.Error("获取题目列表失败", zap.Error(err))
			c.JSON(http.StatusOK, errorResponse(500, "获取题目失败"))
			return
		}

		// 为每个题目创建一个草稿记录，标记考试开始时间
		for _, hq := range homeworkQuestions {
			submit := &model.HomeworkSubmit{
				HomeworkID:     homeworkIDUint64,
				QuestionID:     hq.QuestionID,
				ProblemID:      hq.ProblemID,
				UID:            uid.(string),
				ExamStartTime:  &examStartTime,
				Score:          0,
				IsScored:       0,
				IsOfficiallySubmitted: 0,
			}
			if err := db.Create(submit).Error; err != nil {
				logger.Error("创建考试记录失败", zap.Error(err))
				c.JSON(http.StatusOK, errorResponse(500, "创建考试记录失败"))
				return
			}
		}
	} else {
		// 有草稿记录，更新 exam_start_time
		if err := db.Model(&model.HomeworkSubmit{}).
			Where("homework_id = ? AND uid = ?", homeworkID, uid).
			Updates(map[string]interface{}{"exam_start_time": examStartTime}).Error; err != nil {
			logger.Error("记录考试开始时间失败", zap.Error(err))
			c.JSON(http.StatusOK, errorResponse(500, "记录考试开始时间失败"))
			return
		}
	}

	// 8. 更新设备信息和浏览器信息
	if err := db.Model(&model.HomeworkSubmit{}).
		Where("homework_id = ? AND uid = ?", homeworkID, uid).
		Updates(map[string]interface{}{
			"device_info":  req.DeviceInfo,
			"browser_info": req.BrowserInfo,
		}).Error; err != nil {
		logger.Warn("更新设备信息失败", zap.Error(err))
	}

	logger.Info("学生开始考试",
		zap.String("uid", uid.(string)),
		zap.String("homeworkId", homeworkID),
		zap.Time("startTime", examStartTime))

	// 9. 计算考试结束时间：取（开始时间+考试时长）和（作业结束时间）的较小值
	examEndTimeByDuration := examStartTime.Add(time.Duration(homework.ExamDuration) * time.Minute)
	examEndTime := examEndTimeByDuration
	if examEndTimeByDuration.After(homework.EndTime) {
		examEndTime = homework.EndTime
	}

	// 10. 计算是否允许交卷
	canSubmit := homework.AllowSubmitAfterMinutes == 0

	// 11. 返回考试信息
	c.JSON(http.StatusOK, successResponse(gin.H{
		"examStartTime":    examStartTime,
		"examEndTime":      examEndTime,
		"remainingSeconds": getRemainingSeconds(homework.ExamDuration, &examStartTime, homework.EndTime),
		"canSubmit":        canSubmit,
		"questions":        shuffledQuestions,
		"examConfig": gin.H{
			"examDuration":            homework.ExamDuration,
			"allowSubmitAfterMinutes": homework.AllowSubmitAfterMinutes,
			"disableCopyPaste":        homework.DisableCopyPaste,
			"requireFullscreen":       homework.RequireFullscreen,
			"disallowTabSwitch":       homework.DisallowTabSwitch,
		},
	}))
}

// shuffleQuestions 生成题目乱序
func shuffleQuestions(questions []model.HomeworkQuestion) ([]gin.H, []OrderMappingItem) {
	// Fisher-Yates 洗牌算法
	n := len(questions)
	indices := make([]int, n)
	for i := 0; i < n; i++ {
		indices[i] = i
	}

	rand.Seed(time.Now().UnixNano())
	for i := n - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		indices[i], indices[j] = indices[j], indices[i]
	}

	// 构建乱序后的题目列表和顺序映射
	result := make([]gin.H, n)
	orderMapping := make([]OrderMappingItem, n)

	for displayOrder, originalIndex := range indices {
		question := questions[originalIndex]

		// 构建返回的题目对象（包含完整的question信息）
		questionData := gin.H{
			"id":            question.ID,
			"homeworkId":    question.HomeworkID,
			"questionId":    question.QuestionID,
			"problemId":     question.ProblemID,
			"displayOrder":  displayOrder + 1, // 学生看到的顺序
			"questionOrder": question.QuestionOrder, // 原始顺序
			"score":         question.Score,
		}

		// 如果有题库关联，包含question的完整信息
		if question.Question != nil {
			questionData["question"] = question.Question
		}

		result[displayOrder] = questionData

		// 记录映射关系
		orderMapping[displayOrder] = OrderMappingItem{
			OriginalOrder: question.QuestionOrder,
			DisplayOrder:  displayOrder + 1,
		}
	}

	return result, orderMapping
}

// getRemainingSeconds 计算剩余秒数
func getRemainingSeconds(examDuration int, examStartTime *time.Time, homeworkEndTime time.Time) int {
	// 检查 examStartTime 是否为 nil
	if examStartTime == nil {
		return 0
	}

	// 计算考试结束时间：取（开始时间+考试时长）和（作业结束时间）的较小值
	examEndTimeByDuration := examStartTime.Add(time.Duration(examDuration) * time.Minute)
	examEndTime := examEndTimeByDuration
	if examEndTimeByDuration.After(homeworkEndTime) {
		examEndTime = homeworkEndTime
	}

	remaining := examEndTime.Sub(time.Now())
	if remaining <= 0 {
		return 0
	}
	return int(remaining.Seconds())
}

// GetExamStatus 获取考试状态（学生端轮询）
func (h *Handler) GetExamStatus(c *gin.Context) {
	logger := utils.GetLogger()
	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "未登录"))
		return
	}

	homeworkID := c.Param("homeworkId")
	if homeworkID == "" {
		c.JSON(http.StatusOK, errorResponse(400, "作业ID不能为空"))
		return
	}

	// 转换 homeworkID 为 uint64
	homeworkIDUint64, err := strconv.ParseUint(homeworkID, 10, 64)
	if err != nil {
		logger.Error("作业ID转换失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "作业ID格式错误"))
		return
	}

	db := client.GetDB()

	// 获取作业信息
	var homework model.ClassroomHomework
	if err := db.Where("id = ?", homeworkIDUint64).First(&homework).Error; err != nil {
		logger.Error("获取作业信息失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(404, "作业不存在"))
		return
	}

	// 获取考试开始时间
	var submit model.HomeworkSubmit
	err = db.Where("homework_id = ? AND uid = ?", homeworkIDUint64, uid.(string)).
		Where("exam_start_time IS NOT NULL").
		First(&submit).Error

	if err == gorm.ErrRecordNotFound {
		// 还没开始考试
		c.JSON(http.StatusOK, successResponse(gin.H{
			"isOvertime":      false,
			"remainingSeconds": 0,
			"canSubmit":        false,
			"hasStarted":       false,
		}))
		return
	} else if err != nil {
		logger.Error("查询考试记录失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	// 计算剩余时间和是否超时
	remainingSeconds := getRemainingSeconds(homework.ExamDuration, submit.ExamStartTime, homework.EndTime)
	isOvertime := remainingSeconds <= 0

	// 如果超时且未正式提交，执行自动强制收卷
	if isOvertime && submit.IsOfficiallySubmitted == 0 {
		logger.Info("考试超时，自动强制收卷",
			zap.String("uid", uid.(string)),
			zap.Uint64("homeworkId", homeworkIDUint64))

		// 标记为正式提交和强制收卷
		now := time.Now()
		if err := db.Model(&model.HomeworkSubmit{}).
			Where("homework_id = ? AND uid = ? AND is_officially_submitted = ?", homeworkIDUint64, uid.(string), 0).
			Updates(map[string]interface{}{
				"is_officially_submitted": 1,
				"is_forced_submit":        1,
				"exam_end_time":           now,
			}).Error; err != nil {
			logger.Error("自动强制收卷失败", zap.Error(err))
		} else {
			logger.Info("自动强制收卷成功",
				zap.String("uid", uid.(string)),
				zap.Uint64("homeworkId", homeworkIDUint64))

			// 记录违规日志
			violation := &model.ExamViolationLog{
				HomeworkID:    homeworkIDUint64,
				UID:           uid.(string),
				ViolationType: "forced_submit",
				Description:   "考试超时系统自动收卷",
			}
			if err := db.Create(violation).Error; err != nil {
				logger.Warn("记录自动收卷违规日志失败", zap.Error(err))
			} else {
				logger.Info("已记录自动收卷违规日志",
					zap.String("uid", uid.(string)),
					zap.Uint64("homeworkId", homeworkIDUint64))
			}

			// 重新查询提交记录
			db.Where("homework_id = ? AND uid = ?", homeworkIDUint64, uid.(string)).
				Where("exam_start_time IS NOT NULL").
				First(&submit)
		}
	}

	// 检查是否允许交卷
	if submit.ExamStartTime == nil {
		logger.Error("考试开始时间为空", zap.Uint64("homeworkId", homeworkIDUint64), zap.String("uid", uid.(string)))
		c.JSON(http.StatusOK, errorResponse(500, "考试数据异常"))
		return
	}
	elapsedMinutes := int(time.Since(*submit.ExamStartTime).Minutes())
	canSubmit := elapsedMinutes >= homework.AllowSubmitAfterMinutes && submit.IsOfficiallySubmitted == 0

	// 获取学生保存的题目顺序
	var studentOrder model.StudentQuestionOrder
	questions := []gin.H{}
	if err := db.Where("homework_id = ? AND uid = ?", homeworkIDUint64, uid.(string)).First(&studentOrder).Error; err == nil {
		// 找到了题目顺序映射，需要返回乱序的题目
		var homeworkQuestions []model.HomeworkQuestion
		if err := db.Where("homework_id = ?", homeworkIDUint64).
			Preload("Question").
			Order("question_order ASC").
			Find(&homeworkQuestions).Error; err == nil {
			// 解析顺序映射
			var orderMapping []OrderMappingItem
			json.Unmarshal([]byte(studentOrder.OrderMapping), &orderMapping)

			// 根据映射关系重新排序题目
			questions = make([]gin.H, len(homeworkQuestions))
			for _, mapping := range orderMapping {
				if mapping.DisplayOrder > 0 && mapping.DisplayOrder <= len(homeworkQuestions) {
					// 找到对应原始顺序的题目
					for _, q := range homeworkQuestions {
						if q.QuestionOrder == mapping.OriginalOrder {
							questionData := gin.H{
								"id":            q.ID,
								"homeworkId":    q.HomeworkID,
								"questionId":    q.QuestionID,
								"problemId":     q.ProblemID,
								"displayOrder":  mapping.DisplayOrder,
								"questionOrder": q.QuestionOrder,
								"score":         q.Score,
							}
							if q.Question != nil {
								questionData["question"] = q.Question
							}
							questions[mapping.DisplayOrder-1] = questionData
							break
						}
					}
				}
			}
		}
	}

	c.JSON(http.StatusOK, successResponse(gin.H{
		"isOvertime":       isOvertime,
		"remainingSeconds": remainingSeconds,
		"canSubmit":        canSubmit,
		"hasStarted":       true,
		"isSubmitted":      submit.IsOfficiallySubmitted == 1,
		"isForcedSubmit":   submit.IsForcedSubmit == 1,
		"questions":        questions,
		"examConfig": gin.H{
			"examDuration":            homework.ExamDuration,
			"allowSubmitAfterMinutes": homework.AllowSubmitAfterMinutes,
		},
	}))
}

// LogViolationRequest 记录违规请求
type LogViolationRequest struct {
	HomeworkID    uint64 `json:"homeworkId" binding:"required"`
	ViolationType string `json:"violationType" binding:"required"`
	Description   string `json:"description"`
}

// LogViolation 记录违规行为
func (h *Handler) LogViolation(c *gin.Context) {
	logger := utils.GetLogger()
	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "未登录"))
		return
	}

	var req LogViolationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	db := client.GetDB()

	// 记录违规日志
	violation := &model.ExamViolationLog{
		HomeworkID:    req.HomeworkID,
		UID:           uid.(string),
		ViolationType: req.ViolationType,
		Description:   req.Description,
		IP:           c.ClientIP(),
	}

	if err := db.Create(violation).Error; err != nil {
		logger.Error("记录违规失败",
			zap.Error(err),
			zap.Uint64("homeworkId", req.HomeworkID),
			zap.String("uid", uid.(string)),
			zap.String("violationType", req.ViolationType))
		c.JSON(http.StatusOK, errorResponse(500, "记录失败"))
		return
	}

	// 更新提交记录中的违规计数
	updates := map[string]interface{}{}
	switch req.ViolationType {
	case "tab_switch":
		updates["tab_switch_count"] = gorm.Expr("tab_switch_count + 1")
	case "fullscreen_exit":
		updates["fullscreen_exit_count"] = gorm.Expr("fullscreen_exit_count + 1")
	case "copy_attempt", "paste_attempt":
		updates["copy_paste_attempt_count"] = gorm.Expr("copy_paste_attempt_count + 1")
	}

	if len(updates) > 0 {
		result := db.Model(&model.HomeworkSubmit{}).
			Where("homework_id = ? AND uid = ?", req.HomeworkID, uid).
			Updates(updates)
		if result.Error != nil {
			logger.Warn("更新违规计数失败", zap.Error(result.Error))
		} else {
			logger.Info("更新违规计数成功",
				zap.Uint64("homeworkId", req.HomeworkID),
				zap.String("uid", uid.(string)),
				zap.Int64("rowsAffected", result.RowsAffected))
		}
	}

	logger.Info("记录考试违规成功",
		zap.String("uid", uid.(string)),
		zap.Uint64("homeworkId", req.HomeworkID),
		zap.String("type", req.ViolationType),
		zap.String("description", req.Description))

	c.JSON(http.StatusOK, successResponse(nil))
}

// GetExamMonitoring 获取考试监控数据（教师端）
func (h *Handler) GetExamMonitoring(c *gin.Context) {
	logger := utils.GetLogger()
	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "未登录"))
		return
	}

	homeworkID := c.Param("homeworkId")
	if homeworkID == "" {
		c.JSON(http.StatusOK, errorResponse(400, "作业ID不能为空"))
		return
	}

	// 转换 homeworkID 为 uint64，确保类型匹配
	homeworkIDUint64, err := strconv.ParseUint(homeworkID, 10, 64)
	if err != nil {
		logger.Error("作业ID转换失败", zap.Error(err), zap.String("homeworkId", homeworkID))
		c.JSON(http.StatusOK, errorResponse(400, "作业ID格式错误"))
		return
	}

	db := client.GetDB()

	// 1. 获取作业信息
	var homework model.ClassroomHomework
	if err := db.Where("id = ?", homeworkIDUint64).First(&homework).Error; err != nil {
		logger.Error("获取作业信息失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(404, "作业不存在"))
		return
	}

	// 2. 检查权限（只有教师可以查看）
	if homework.IsExamMode != 1 {
		c.JSON(http.StatusOK, errorResponse(400, "该作业不是考试模式"))
		return
	}

	// 3. 获取班级所有学生
	var students []model.ClassroomStudent
	if err := db.Where("classroom_id = ?", homework.ClassroomID).Find(&students).Error; err != nil {
		logger.Error("获取学生列表失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "获取学生列表失败"))
		return
	}

	totalStudents := len(students)

	// 4. 统计各状态学生数量
	startedCount := 0
	submittedCount := 0
	inProgressCount := 0
	notStartedCount := 0

	type StudentStatus struct {
		UID               string `json:"uid"`
		Name              string `json:"name"`
		Status            string `json:"status"` // not_started, in_progress, submitted, forced_submit
		ExamStartTime     *time.Time `json:"examStartTime,omitempty"`
		ElapsedMinutes    int    `json:"elapsedMinutes"`
		RemainingSeconds  int    `json:"remainingSeconds"`
		ViolationCount    int    `json:"violationCount"`
		Violations        []gin.H `json:"violations,omitempty"`
	}

	studentStatuses := make([]StudentStatus, 0, totalStudents)

	for _, student := range students {
		// 查询该学生的提交记录
		var submits []model.HomeworkSubmit
		db.Where("homework_id = ? AND uid = ?", homeworkIDUint64, student.UID).Find(&submits)

		status := StudentStatus{
			UID:  student.UID,
			Name: student.RealName,
		}

		if len(submits) == 0 || submits[0].ExamStartTime == nil {
			// 未开始
			status.Status = "not_started"
			notStartedCount++
		} else {
			// 已开始
			startedCount++

			// 检查是否已提交
			hasSubmitted := false
			for _, submit := range submits {
				if submit.IsOfficiallySubmitted == 1 {
					hasSubmitted = true
					status.ExamStartTime = submit.ExamStartTime

					// 计算已用时长
					if submit.ExamEndTime != nil {
						// 有记录的结束时间，使用实际时间
						realElapsed := int(submit.ExamEndTime.Sub(*submit.ExamStartTime).Minutes())
						status.ElapsedMinutes = realElapsed
						logger.Debug("使用实际考试结束时间",
							zap.String("uid", student.UID),
							zap.Time("examStartTime", *submit.ExamStartTime),
							zap.Time("examEndTime", *submit.ExamEndTime),
							zap.Int("elapsedMinutes", realElapsed))
					} else {
						// 没有记录的结束时间（可能是旧数据）
						// 根据提交类型和考试时长计算
						logger.Warn("考试结束时间为空",
							zap.String("uid", student.UID),
							zap.Uint64("homeworkId", homeworkIDUint64),
							zap.Bool("isForcedSubmit", submit.IsForcedSubmit == 1))
						// 1. 计算理论考试结束时间（取考试时长和截止时间的较小值）
						examEndTimeByDuration := submit.ExamStartTime.Add(time.Duration(homework.ExamDuration) * time.Minute)
						examEndTime := examEndTimeByDuration
						if examEndTimeByDuration.After(homework.EndTime) {
							examEndTime = homework.EndTime
						}

						// 2. 计算已用时长
						status.ElapsedMinutes = int(examEndTime.Sub(*submit.ExamStartTime).Minutes())
					}

					if submit.IsForcedSubmit == 1 {
						status.Status = "forced_submit"
					} else {
						status.Status = "submitted"
					}
					submittedCount++
					break
				}
			}

			if !hasSubmitted {
				// 答题中
				status.Status = "in_progress"
				status.ExamStartTime = submits[0].ExamStartTime

				// 计算实际考试结束时间（取考试时长和截止时间的较小值）
				examEndTimeByDuration := submits[0].ExamStartTime.Add(time.Duration(homework.ExamDuration) * time.Minute)
				examEndTime := examEndTimeByDuration
				if examEndTimeByDuration.After(homework.EndTime) {
					examEndTime = homework.EndTime
				}

				// 已用时长：从开始到当前时间，但不超过实际考试结束时间
				now := time.Now()
				var effectiveEndTime time.Time
				if now.After(examEndTime) {
					effectiveEndTime = examEndTime
				} else {
					effectiveEndTime = now
				}
				status.ElapsedMinutes = int(effectiveEndTime.Sub(*submits[0].ExamStartTime).Minutes())

				status.RemainingSeconds = getRemainingSeconds(homework.ExamDuration, submits[0].ExamStartTime, homework.EndTime)
				inProgressCount++
			}

			// 获取违规记录
			var violations []model.ExamViolationLog
			if err := db.Where("homework_id = ? AND uid = ?", homeworkIDUint64, student.UID).Find(&violations).Error; err != nil {
				logger.Warn("查询违规记录失败",
					zap.Error(err),
					zap.Uint64("homeworkId", homeworkIDUint64),
					zap.String("uid", student.UID))
			}

			if len(violations) > 0 {
				status.ViolationCount = len(violations)
				status.Violations = make([]gin.H, 0, len(violations))

				// 按类型汇总违规次数
				violationMap := make(map[string]int)
				for _, v := range violations {
					violationMap[v.ViolationType]++
				}

				for vType, count := range violationMap {
					status.Violations = append(status.Violations, gin.H{
						"type":  vType,
						"count": count,
					})
				}

				logger.Debug("学生违规记录",
					zap.String("uid", student.UID),
					zap.Int("count", len(violations)),
					zap.Any("violations", status.Violations))
			}
		}

		studentStatuses = append(studentStatuses, status)
	}

	logger.Info("获取考试监控数据",
		zap.String("uid", uid.(string)),
		zap.Uint64("homeworkId", homeworkIDUint64))

	c.JSON(http.StatusOK, successResponse(gin.H{
		"totalStudents":   totalStudents,
		"startedCount":    startedCount,
		"submittedCount":  submittedCount,
		"inProgressCount": inProgressCount,
		"notStartedCount": notStartedCount,
		"studentList":     studentStatuses,
	}))
}

// ForceSubmitRequest 强制交卷请求
type ForceSubmitRequest struct {
	HomeworkID uint64 `json:"homeworkId" binding:"required"`
	UID        string `json:"uid" binding:"required"`
	Reason     string `json:"reason"`
}

// ForceSubmit 强制单个学生交卷（教师端）
func (h *Handler) ForceSubmit(c *gin.Context) {
	logger := utils.GetLogger()

	var req ForceSubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}

	// 添加调试日志
	logger.Debug("强制交卷请求",
		zap.Uint64("homeworkId", req.HomeworkID),
		zap.String("uid", req.UID),
		zap.String("reason", req.Reason))

	db := client.GetDB()

	// 1. 检查作业是否是考试模式
	var homework model.ClassroomHomework
	if err := db.Where("id = ?", req.HomeworkID).First(&homework).Error; err != nil {
		logger.Error("获取作业信息失败", zap.Error(err), zap.Uint64("homeworkId", req.HomeworkID))
		c.JSON(http.StatusOK, errorResponse(404, "作业不存在"))
		return
	}

	if homework.IsExamMode != 1 {
		logger.Warn("作业不是考试模式", zap.Uint64("homeworkId", req.HomeworkID), zap.Int("isExamMode", homework.IsExamMode))
		c.JSON(http.StatusOK, errorResponse(400, "该作业不是考试模式"))
		return
	}

	// 2. 获取学生的草稿答案
	var submits []model.HomeworkSubmit
	if err := db.Where("homework_id = ? AND uid = ?", req.HomeworkID, req.UID).
		Find(&submits).Error; err != nil {
		logger.Error("查询提交记录失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	if len(submits) == 0 {
		logger.Warn("学生未开始考试", zap.String("uid", req.UID), zap.Uint64("homeworkId", req.HomeworkID))
		c.JSON(http.StatusOK, errorResponse(404, "该学生未开始考试"))
		return
	}

	logger.Debug("学生提交记录",
		zap.String("uid", req.UID),
		zap.Int("count", len(submits)),
		zap.Int("draftCount", countDraftSubmits(submits)))

	// 3. 标记所有草稿为正式提交，并标记为强制收卷
	now := time.Now()
	result := db.Model(&model.HomeworkSubmit{}).
		Where("homework_id = ? AND uid = ? AND is_officially_submitted = 0", req.HomeworkID, req.UID).
		Updates(map[string]interface{}{
			"is_officially_submitted": 1,
			"is_forced_submit":        1,
			"exam_end_time":           now,
		})

	if result.Error != nil {
		logger.Error("强制交卷失败", zap.Error(result.Error))
		c.JSON(http.StatusOK, errorResponse(500, "强制交卷失败"))
		return
	}

	// 检查是否真的更新了记录
	if result.RowsAffected == 0 {
		logger.Warn("强制交卷未更新任何记录",
			zap.String("uid", req.UID),
			zap.Uint64("homeworkId", req.HomeworkID),
			zap.Int("draftCount", countDraftSubmits(submits)))
		c.JSON(http.StatusOK, errorResponse(400, "该学生没有未提交的草稿"))
		return
	}

	// 4. 记录违规日志
	violation := &model.ExamViolationLog{
		HomeworkID:    req.HomeworkID,
		UID:           req.UID,
		ViolationType: "forced_submit",
		Description:   req.Reason,
	}
	db.Create(violation)

	logger.Info("强制学生交卷成功",
		zap.String("uid", req.UID),
		zap.Uint64("homeworkId", req.HomeworkID),
		zap.String("reason", req.Reason),
		zap.Int64("rowsAffected", result.RowsAffected))

	c.JSON(http.StatusOK, successResponse(nil))
}

// countDraftSubmits 统计草稿提交数量
func countDraftSubmits(submits []model.HomeworkSubmit) int {
	count := 0
	for _, s := range submits {
		if s.IsOfficiallySubmitted == 0 {
			count++
		}
	}
	return count
}

// ForceSubmitAll 强制所有未交卷学生收卷（考试结束）
func (h *Handler) ForceSubmitAll(c *gin.Context) {
	logger := utils.GetLogger()
	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusOK, errorResponse(401, "未登录"))
		return
	}

	homeworkIDStr := c.Param("homeworkId")
	if homeworkIDStr == "" {
		c.JSON(http.StatusOK, errorResponse(400, "作业ID不能为空"))
		return
	}

	// 转换 homeworkID 为 uint64
	homeworkID, err := strconv.ParseUint(homeworkIDStr, 10, 64)
	if err != nil {
		logger.Error("作业ID转换失败", zap.Error(err), zap.String("homeworkIdStr", homeworkIDStr))
		c.JSON(http.StatusOK, errorResponse(400, "作业ID格式错误"))
		return
	}

	db := client.GetDB()

	// 1. 检查作业是否是考试模式
	var homework model.ClassroomHomework
	if err := db.Where("id = ?", homeworkID).First(&homework).Error; err != nil {
		logger.Error("获取作业信息失败", zap.Error(err), zap.Uint64("homeworkId", homeworkID))
		c.JSON(http.StatusOK, errorResponse(404, "作业不存在"))
		return
	}

	if homework.IsExamMode != 1 {
		logger.Warn("作业不是考试模式", zap.Uint64("homeworkId", homeworkID), zap.Int("isExamMode", homework.IsExamMode))
		c.JSON(http.StatusOK, errorResponse(400, "该作业不是考试模式"))
		return
	}

	// 2. 获取所有已开始但未提交的学生
	var inProgressStudents []struct {
		UID string
	}

	subQuery := db.Model(&model.HomeworkSubmit{}).
		Select("uid").
		Where("homework_id = ? AND exam_start_time IS NOT NULL", homeworkID).
		Group("uid").
		Having("SUM(is_officially_submitted) = 0")

	if err := subQuery.Scan(&inProgressStudents).Error; err != nil {
		logger.Error("查询未交卷学生失败", zap.Error(err), zap.Uint64("homeworkId", homeworkID))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	logger.Debug("批量强制收卷",
		zap.Uint64("homeworkId", homeworkID),
		zap.Int("inProgressCount", len(inProgressStudents)))

	// 3. 批量强制收卷
	now := time.Now()
	forcedCount := 0

	for _, student := range inProgressStudents {
		result := db.Model(&model.HomeworkSubmit{}).
			Where("homework_id = ? AND uid = ? AND is_officially_submitted = 0", homeworkID, student.UID).
			Updates(map[string]interface{}{
				"is_officially_submitted": 1,
				"is_forced_submit":        1,
				"exam_end_time":           now,
			})

		if result.Error != nil {
			logger.Error("强制收卷失败",
				zap.String("uid", student.UID),
				zap.Uint64("homeworkId", homeworkID),
				zap.Error(result.Error))
			continue
		}

		if result.RowsAffected > 0 {
			forcedCount++
		}
	}

	logger.Info("批量强制收卷完成",
		zap.String("uid", uid.(string)),
		zap.Uint64("homeworkId", homeworkID),
		zap.Int("forcedCount", forcedCount))

	c.JSON(http.StatusOK, successResponse(gin.H{
		"forcedCount": forcedCount,
	}))
}
