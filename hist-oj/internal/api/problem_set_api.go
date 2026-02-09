package api

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/hoj/hist-oj/internal/model"
	"github.com/hoj/hist-oj/internal/service"
	"github.com/hoj/hist-oj/internal/utils"
)

// ProblemSetHandler 题目集处理器
type ProblemSetHandler struct {
	db                       *gorm.DB
	problemtoolsPDFGenerator *service.ProblemToolsPDFGenerator
}

// NewProblemSetHandler 创建题目集处理器
func NewProblemSetHandler(db *gorm.DB, problemtoolsPDFGenerator *service.ProblemToolsPDFGenerator) *ProblemSetHandler {
	return &ProblemSetHandler{
		db:                       db,
		problemtoolsPDFGenerator: problemtoolsPDFGenerator,
	}
}

// CreateProblemSetRequest 创建题目集请求
type CreateProblemSetRequest struct {
	Title       string `json:"title" binding:"required"`
	Author      string `json:"author"`
	ContestDate string `json:"contest_date"` // YYYY-MM-DD 格式
}

// UpdateProblemSetRequest 更新题目集请求
type UpdateProblemSetRequest struct {
	Title       string `json:"title"`
	Author      string `json:"author"`
	ContestDate string `json:"contest_date"`
}

// CreateProblemRequest 创建题目请求
type CreateProblemRequest struct {
	ProblemLetter string `json:"problem_letter"`
	Title         string `json:"title" binding:"required"`
	TimeLimit     int    `json:"time_limit"`
	Description   string `json:"description"`
	InputFormat   string `json:"input_format"`
	OutputFormat  string `json:"output_format"`
	Note          string `json:"note"`
	SortOrder     int    `json:"sort_order"`
}

// UpdateProblemRequest 更新题目请求
type UpdateProblemRequest struct {
	ProblemLetter string `json:"problem_letter"`
	Title         string `json:"title"`
	TimeLimit     int    `json:"time_limit"`
	Description   string `json:"description"`
	InputFormat   string `json:"input_format"`
	OutputFormat  string `json:"output_format"`
	Note          string `json:"note"`
	SortOrder     int    `json:"sort_order"`
}

// CreateExampleRequest 创建样例请求
type CreateExampleRequest struct {
	ExampleNo   int    `json:"example_no" binding:"required"`
	Description string `json:"description"`
	Input       string `json:"input" binding:"required"`
	Output      string `json:"output" binding:"required"`
}

// UpdateExampleRequest 更新样例请求
type UpdateExampleRequest struct {
	Description string `json:"description"`
	Input       string `json:"input"`
	Output      string `json:"output"`
}

// GetProblemSets 获取当前用户的题目集列表
func (h *ProblemSetHandler) GetProblemSets(c *gin.Context) {
	logger := utils.GetLogger()

	// 从 context 获取用户ID
	uid, exists := c.Get("uid")
	if !exists {
		logger.Warn("未登录用户")
		c.JSON(http.StatusOK, errorResponse(401, "用户未登录"))
		return
	}

	uidStr, ok := uid.(string)
	if !ok {
		logger.Warn("用户ID格式错误")
		c.JSON(http.StatusOK, errorResponse(400, "用户ID格式错误"))
		return
	}

	// 查询题目集（包含题目数量）
	var problemSets []model.ProblemSet
	if err := h.db.Where("user_id = ?", uidStr).Order("created_at DESC").Find(&problemSets).Error; err != nil {
		logger.Error("查询题目集失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	// 为每个题目集加载题目数据（用于显示题目数量）
	for i := range problemSets {
		var problems []model.ProblemSetProblem
		if err := h.db.Where("set_id = ?", problemSets[i].ID).
			Order("sort_order ASC").
			Find(&problems).Error; err != nil {
			logger.Error("查询题目失败",
				zap.Uint64("set_id", problemSets[i].ID),
				zap.Error(err))
		} else {
			problemSets[i].Problems = problems
		}
	}

	logger.Info("获取题目集列表成功", zap.String("uid", uidStr), zap.Int("count", len(problemSets)))
	c.JSON(http.StatusOK, successResponse(problemSets))
}

// GetProblemSet 获取题目集详情
func (h *ProblemSetHandler) GetProblemSet(c *gin.Context) {
	logger := utils.GetLogger()

	// 获取题目集ID
	id := c.Param("id")
	if id == "" {
		logger.Warn("缺少题目集ID")
		c.JSON(http.StatusOK, errorResponse(400, "缺少题目集ID"))
		return
	}

	// 从 context 获取用户ID
	uid, exists := c.Get("uid")
	if !exists {
		logger.Warn("未登录用户")
		c.JSON(http.StatusOK, errorResponse(401, "用户未登录"))
		return
	}

	uidStr, ok := uid.(string)
	if !ok {
		logger.Warn("用户ID格式错误")
		c.JSON(http.StatusOK, errorResponse(400, "用户ID格式错误"))
		return
	}

	// 查询题目集
	var problemSet model.ProblemSet
	if err := h.db.Where("id = ? AND user_id = ?", id, uidStr).
		Preload("Problems", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Preload("Problems.Examples", func(db *gorm.DB) *gorm.DB {
			return db.Order("example_no ASC")
		}).
		First(&problemSet).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Warn("题目集不存在或无权访问", zap.String("id", id), zap.String("uid", uidStr))
			c.JSON(http.StatusOK, errorResponse(404, "题目集不存在或无权访问"))
			return
		}
		logger.Error("查询题目集失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	// 添加调试日志
	logger.Info("查询题目集成功",
		zap.String("id", id),
		zap.String("title", problemSet.Title),
		zap.Int("problems_count", len(problemSet.Problems)))

	logger.Info("获取题目集详情成功", zap.String("id", id))
	c.JSON(http.StatusOK, successResponse(problemSet))
}

// CreateProblemSet 创建题目集
func (h *ProblemSetHandler) CreateProblemSet(c *gin.Context) {
	logger := utils.GetLogger()

	// 从 context 获取用户ID
	uid, exists := c.Get("uid")
	if !exists {
		logger.Warn("未登录用户")
		c.JSON(http.StatusOK, errorResponse(401, "用户未登录"))
		return
	}

	uidStr, ok := uid.(string)
	if !ok {
		logger.Warn("用户ID格式错误")
		c.JSON(http.StatusOK, errorResponse(400, "用户ID格式错误"))
		return
	}

	// 解析请求
	var req CreateProblemSetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "请求参数错误: "+err.Error()))
		return
	}

	// 解析日期
	var contestDate *time.Time
	if req.ContestDate != "" {
		parsedDate, err := time.Parse("2006-01-02", req.ContestDate)
		if err != nil {
			logger.Warn("日期格式错误", zap.String("contest_date", req.ContestDate), zap.Error(err))
			c.JSON(http.StatusOK, errorResponse(400, "日期格式错误，请使用 YYYY-MM-DD 格式"))
			return
		}
		contestDate = &parsedDate
	}

	// 将用户ID字符串转换为uint64
	userID, err := strconv.ParseUint(uidStr, 10, 64)
	if err != nil {
		logger.Warn("用户ID转换失败", zap.String("uid", uidStr), zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "用户ID格式错误"))
		return
	}

	// 创建题目集
	problemSet := model.ProblemSet{
		UserID:      userID,
		Title:       req.Title,
		Author:      req.Author,
		ContestDate: contestDate,
	}

	if err := h.db.Create(&problemSet).Error; err != nil {
		logger.Error("创建题目集失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "创建失败"))
		return
	}

	logger.Info("创建题目集成功", zap.String("uid", uidStr), zap.Uint64("id", problemSet.ID))
	c.JSON(http.StatusOK, successResponse(problemSet))
}

// UpdateProblemSet 更新题目集
func (h *ProblemSetHandler) UpdateProblemSet(c *gin.Context) {
	logger := utils.GetLogger()

	// 获取题目集ID
	id := c.Param("id")
	if id == "" {
		logger.Warn("缺少题目集ID")
		c.JSON(http.StatusOK, errorResponse(400, "缺少题目集ID"))
		return
	}

	// 从 context 获取用户ID
	uid, exists := c.Get("uid")
	if !exists {
		logger.Warn("未登录用户")
		c.JSON(http.StatusOK, errorResponse(401, "用户未登录"))
		return
	}

	uidStr, ok := uid.(string)
	if !ok {
		logger.Warn("用户ID格式错误")
		c.JSON(http.StatusOK, errorResponse(400, "用户ID格式错误"))
		return
	}

	// 解析请求
	var req UpdateProblemSetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "请求参数错误: "+err.Error()))
		return
	}

	// 查询题目集
	var problemSet model.ProblemSet
	if err := h.db.Where("id = ? AND user_id = ?", id, uidStr).First(&problemSet).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Warn("题目集不存在或无权访问", zap.String("id", id), zap.String("uid", uidStr))
			c.JSON(http.StatusOK, errorResponse(404, "题目集不存在或无权访问"))
			return
		}
		logger.Error("查询题目集失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	// 解析日期
	var contestDate *time.Time
	if req.ContestDate != "" {
		parsedDate, err := time.Parse("2006-01-02", req.ContestDate)
		if err != nil {
			logger.Warn("日期格式错误", zap.String("contest_date", req.ContestDate), zap.Error(err))
			c.JSON(http.StatusOK, errorResponse(400, "日期格式错误，请使用 YYYY-MM-DD 格式"))
			return
		}
		contestDate = &parsedDate
	}

	// 更新字段
	updates := map[string]interface{}{}
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Author != "" {
		updates["author"] = req.Author
	}
	if req.ContestDate != "" {
		updates["contest_date"] = contestDate
	}

	if err := h.db.Model(&problemSet).Updates(updates).Error; err != nil {
		logger.Error("更新题目集失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "更新失败"))
		return
	}

	// 重新查询更新后的数据
	h.db.First(&problemSet, id)

	logger.Info("更新题目集成功", zap.String("id", id))
	c.JSON(http.StatusOK, successResponse(problemSet))
}

// DeleteProblemSet 删除题目集
func (h *ProblemSetHandler) DeleteProblemSet(c *gin.Context) {
	logger := utils.GetLogger()

	// 获取题目集ID
	id := c.Param("id")
	if id == "" {
		logger.Warn("缺少题目集ID")
		c.JSON(http.StatusOK, errorResponse(400, "缺少题目集ID"))
		return
	}

	// 从 context 获取用户ID
	uid, exists := c.Get("uid")
	if !exists {
		logger.Warn("未登录用户")
		c.JSON(http.StatusOK, errorResponse(401, "用户未登录"))
		return
	}

	uidStr, ok := uid.(string)
	if !ok {
		logger.Warn("用户ID格式错误")
		c.JSON(http.StatusOK, errorResponse(400, "用户ID格式错误"))
		return
	}

	// 删除题目集（会级联删除题目和样例）
	result := h.db.Where("id = ? AND user_id = ?", id, uidStr).Delete(&model.ProblemSet{})
	if result.Error != nil {
		logger.Error("删除题目集失败", zap.Error(result.Error))
		c.JSON(http.StatusOK, errorResponse(500, "删除失败"))
		return
	}

	if result.RowsAffected == 0 {
		logger.Warn("题目集不存在或无权访问", zap.String("id", id), zap.String("uid", uidStr))
		c.JSON(http.StatusOK, errorResponse(404, "题目集不存在或无权访问"))
		return
	}

	logger.Info("删除题目集成功", zap.String("id", id))
	c.JSON(http.StatusOK, successResponse(nil))
}

// CreateProblem 添加题目
func (h *ProblemSetHandler) CreateProblem(c *gin.Context) {
	logger := utils.GetLogger()

	// 获取题目集ID
	setID := c.Param("id")
	if setID == "" {
		logger.Warn("缺少题目集ID")
		c.JSON(http.StatusOK, errorResponse(400, "缺少题目集ID"))
		return
	}

	// 从 context 获取用户ID
	uid, exists := c.Get("uid")
	if !exists {
		logger.Warn("未登录用户")
		c.JSON(http.StatusOK, errorResponse(401, "用户未登录"))
		return
	}

	uidStr, ok := uid.(string)
	if !ok {
		logger.Warn("用户ID格式错误")
		c.JSON(http.StatusOK, errorResponse(400, "用户ID格式错误"))
		return
	}

	// 验证题目集所有权
	var problemSet model.ProblemSet
	if err := h.db.Where("id = ? AND user_id = ?", setID, uidStr).First(&problemSet).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Warn("题目集不存在或无权访问", zap.String("id", setID), zap.String("uid", uidStr))
			c.JSON(http.StatusOK, errorResponse(404, "题目集不存在或无权访问"))
			return
		}
		logger.Error("查询题目集失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	// 解析请求
	var req CreateProblemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "请求参数错误: "+err.Error()))
		return
	}

	// 创建题目
	problem := model.ProblemSetProblem{
		SetID:         problemSet.ID,
		ProblemLetter: req.ProblemLetter,
		Title:         req.Title,
		TimeLimit:     req.TimeLimit,
		Description:   req.Description,
		InputFormat:   req.InputFormat,
		OutputFormat:  req.OutputFormat,
		Note:          req.Note,
		SortOrder:     req.SortOrder,
	}

	if problem.TimeLimit == 0 {
		problem.TimeLimit = 1000 // 默认 1000ms
	}

	if err := h.db.Create(&problem).Error; err != nil {
		logger.Error("创建题目失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "创建失败"))
		return
	}

	logger.Info("创建题目成功", zap.String("set_id", setID), zap.Uint64("problem_id", problem.ID))
	c.JSON(http.StatusOK, successResponse(problem))
}

// UpdateProblem 更新题目
func (h *ProblemSetHandler) UpdateProblem(c *gin.Context) {
	logger := utils.GetLogger()

	// 获取题目ID
	problemID := c.Param("problemId")
	if problemID == "" {
		logger.Warn("缺少题目ID")
		c.JSON(http.StatusOK, errorResponse(400, "缺少题目ID"))
		return
	}

	// 从 context 获取用户ID
	uid, exists := c.Get("uid")
	if !exists {
		logger.Warn("未登录用户")
		c.JSON(http.StatusOK, errorResponse(401, "用户未登录"))
		return
	}

	uidStr, ok := uid.(string)
	if !ok {
		logger.Warn("用户ID格式错误")
		c.JSON(http.StatusOK, errorResponse(400, "用户ID格式错误"))
		return
	}

	// 查询题目（需要验证用户所有权）
	var problem model.ProblemSetProblem
	if err := h.db.Joins("JOIN xcpc_problem_set ON xcpc_problem_set_problem.set_id = xcpc_problem_set.id").
		Where("xcpc_problem_set_problem.id = ? AND xcpc_problem_set.user_id = ?", problemID, uidStr).
		First(&problem).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Warn("题目不存在或无权访问", zap.String("problem_id", problemID), zap.String("uid", uidStr))
			c.JSON(http.StatusOK, errorResponse(404, "题目不存在或无权访问"))
			return
		}
		logger.Error("查询题目失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	// 解析请求
	var req UpdateProblemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "请求参数错误: "+err.Error()))
		return
	}

	// 记录接收到的数据（用于调试）
	logger.Info("UpdateProblem 接收到的数据",
		zap.String("problem_id", problemID),
		zap.String("title", req.Title),
		zap.String("description", req.Description),
		zap.Int("time_limit", req.TimeLimit))

	// 更新字段（直接更新所有字段，允许清空）
	updates := map[string]interface{}{}

	if req.ProblemLetter != "" {
		updates["problem_letter"] = req.ProblemLetter
	}
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.TimeLimit > 0 {
		updates["time_limit"] = req.TimeLimit
	}

	// 这些字段允许为空字符串，所以总是更新（允许清空）
	updates["description"] = req.Description
	updates["input_format"] = req.InputFormat
	updates["output_format"] = req.OutputFormat
	updates["note"] = req.Note
	updates["sort_order"] = req.SortOrder

	if err := h.db.Model(&problem).Updates(updates).Error; err != nil {
		logger.Error("更新题目失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "更新失败"))
		return
	}

	// 重新查询更新后的数据
	h.db.First(&problem, problemID)

	logger.Info("更新题目成功", zap.String("problem_id", problemID))
	c.JSON(http.StatusOK, successResponse(problem))
}

// DeleteProblem 删除题目
func (h *ProblemSetHandler) DeleteProblem(c *gin.Context) {
	logger := utils.GetLogger()

	// 获取题目ID
	problemID := c.Param("problemId")
	if problemID == "" {
		logger.Warn("缺少题目ID")
		c.JSON(http.StatusOK, errorResponse(400, "缺少题目ID"))
		return
	}

	// 从 context 获取用户ID
	uid, exists := c.Get("uid")
	if !exists {
		logger.Warn("未登录用户")
		c.JSON(http.StatusOK, errorResponse(401, "用户未登录"))
		return
	}

	uidStr, ok := uid.(string)
	if !ok {
		logger.Warn("用户ID格式错误")
		c.JSON(http.StatusOK, errorResponse(400, "用户ID格式错误"))
		return
	}

	// 删除题目（需要验证用户所有权，会级联删除样例）
	result := h.db.Joins("JOIN xcpc_problem_set ON xcpc_problem_set_problem.set_id = xcpc_problem_set.id").
		Where("xcpc_problem_set_problem.id = ? AND xcpc_problem_set.user_id = ?", problemID, uidStr).
		Delete(&model.ProblemSetProblem{})

	if result.Error != nil {
		logger.Error("删除题目失败", zap.Error(result.Error))
		c.JSON(http.StatusOK, errorResponse(500, "删除失败"))
		return
	}

	if result.RowsAffected == 0 {
		logger.Warn("题目不存在或无权访问", zap.String("problem_id", problemID), zap.String("uid", uidStr))
		c.JSON(http.StatusOK, errorResponse(404, "题目不存在或无权访问"))
		return
	}

	logger.Info("删除题目成功", zap.String("problem_id", problemID))
	c.JSON(http.StatusOK, successResponse(nil))
}

// CreateExample 添加样例
func (h *ProblemSetHandler) CreateExample(c *gin.Context) {
	logger := utils.GetLogger()

	// 获取题目ID
	problemID := c.Param("problemId")
	if problemID == "" {
		logger.Warn("缺少题目ID")
		c.JSON(http.StatusOK, errorResponse(400, "缺少题目ID"))
		return
	}

	// 从 context 获取用户ID
	uid, exists := c.Get("uid")
	if !exists {
		logger.Warn("未登录用户")
		c.JSON(http.StatusOK, errorResponse(401, "用户未登录"))
		return
	}

	uidStr, ok := uid.(string)
	if !ok {
		logger.Warn("用户ID格式错误")
		c.JSON(http.StatusOK, errorResponse(400, "用户ID格式错误"))
		return
	}

	// 验证题目所有权
	var problem model.ProblemSetProblem
	if err := h.db.Joins("JOIN xcpc_problem_set ON xcpc_problem_set_problem.set_id = xcpc_problem_set.id").
		Where("xcpc_problem_set_problem.id = ? AND xcpc_problem_set.user_id = ?", problemID, uidStr).
		First(&problem).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Warn("题目不存在或无权访问", zap.String("problem_id", problemID), zap.String("uid", uidStr))
			c.JSON(http.StatusOK, errorResponse(404, "题目不存在或无权访问"))
			return
		}
		logger.Error("查询题目失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	// 解析请求
	var req CreateExampleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "请求参数错误: "+err.Error()))
		return
	}

	// 创建样例
	example := model.ProblemExample{
		ProblemID:   problem.ID,
		ExampleNo:   req.ExampleNo,
		Description: req.Description,
		Input:       req.Input,
		Output:      req.Output,
	}

	if err := h.db.Create(&example).Error; err != nil {
		logger.Error("创建样例失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "创建失败"))
		return
	}

	logger.Info("创建样例成功", zap.String("problem_id", problemID))
	c.JSON(http.StatusOK, successResponse(example))
}

// UpdateExample 更新样例
func (h *ProblemSetHandler) UpdateExample(c *gin.Context) {
	logger := utils.GetLogger()

	// 获取样例ID
	exampleID := c.Param("exampleId")
	if exampleID == "" {
		logger.Warn("缺少样例ID")
		c.JSON(http.StatusOK, errorResponse(400, "缺少样例ID"))
		return
	}

	// 从 context 获取用户ID
	uid, exists := c.Get("uid")
	if !exists {
		logger.Warn("未登录用户")
		c.JSON(http.StatusOK, errorResponse(401, "用户未登录"))
		return
	}

	uidStr, ok := uid.(string)
	if !ok {
		logger.Warn("用户ID格式错误")
		c.JSON(http.StatusOK, errorResponse(400, "用户ID格式错误"))
		return
	}

	// 查询样例（需要验证用户所有权）
	var example model.ProblemExample
	if err := h.db.Joins("JOIN xcpc_problem_set_problem ON xcpc_problem_example.problem_id = xcpc_problem_set_problem.id").
		Joins("JOIN xcpc_problem_set ON xcpc_problem_set_problem.set_id = xcpc_problem_set.id").
		Where("xcpc_problem_example.id = ? AND xcpc_problem_set.user_id = ?", exampleID, uidStr).
		First(&example).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Warn("样例不存在或无权访问", zap.String("example_id", exampleID), zap.String("uid", uidStr))
			c.JSON(http.StatusOK, errorResponse(404, "样例不存在或无权访问"))
			return
		}
		logger.Error("查询样例失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	// 解析请求
	var req UpdateExampleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "请求参数错误: "+err.Error()))
		return
	}

	// 记录接收到的数据（用于调试）
	logger.Info("UpdateExample 接收到的数据",
		zap.String("example_id", exampleID),
		zap.String("description", req.Description),
		zap.String("input", req.Input),
		zap.String("output", req.Output))

	// 更新字段（直接更新所有字段，允许清空）
	updates := map[string]interface{}{
		"description": req.Description,
		"input":       req.Input,
		"output":      req.Output,
	}

	if err := h.db.Model(&example).Updates(updates).Error; err != nil {
		logger.Error("更新样例失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "更新失败"))
		return
	}

	// 重新查询更新后的数据
	h.db.First(&example, exampleID)

	logger.Info("更新样例成功", zap.String("example_id", exampleID))
	c.JSON(http.StatusOK, successResponse(example))
}

// DeleteExample 删除样例
func (h *ProblemSetHandler) DeleteExample(c *gin.Context) {
	logger := utils.GetLogger()

	// 获取样例ID
	exampleID := c.Param("exampleId")
	if exampleID == "" {
		logger.Warn("缺少样例ID")
		c.JSON(http.StatusOK, errorResponse(400, "缺少样例ID"))
		return
	}

	// 从 context 获取用户ID
	uid, exists := c.Get("uid")
	if !exists {
		logger.Warn("未登录用户")
		c.JSON(http.StatusOK, errorResponse(401, "用户未登录"))
		return
	}

	uidStr, ok := uid.(string)
	if !ok {
		logger.Warn("用户ID格式错误")
		c.JSON(http.StatusOK, errorResponse(400, "用户ID格式错误"))
		return
	}

	// 删除样例（需要验证用户所有权）
	result := h.db.Joins("JOIN xcpc_problem_set_problem ON xcpc_problem_example.problem_id = xcpc_problem_set_problem.id").
		Joins("JOIN xcpc_problem_set ON xcpc_problem_set_problem.set_id = xcpc_problem_set.id").
		Where("xcpc_problem_example.id = ? AND xcpc_problem_set.user_id = ?", exampleID, uidStr).
		Delete(&model.ProblemExample{})

	if result.Error != nil {
		logger.Error("删除样例失败", zap.Error(result.Error))
		c.JSON(http.StatusOK, errorResponse(500, "删除失败"))
		return
	}

	if result.RowsAffected == 0 {
		logger.Warn("样例不存在或无权访问", zap.String("example_id", exampleID), zap.String("uid", uidStr))
		c.JSON(http.StatusOK, errorResponse(404, "样例不存在或无权访问"))
		return
	}

	logger.Info("删除样例成功", zap.String("example_id", exampleID))
	c.JSON(http.StatusOK, successResponse(nil))
}

// GeneratePDF 生成并下载 PDF
func (h *ProblemSetHandler) GeneratePDF(c *gin.Context) {
	logger := utils.GetLogger()

	// 获取题目集ID
	id := c.Param("id")
	if id == "" {
		logger.Warn("缺少题目集ID")
		c.JSON(http.StatusOK, errorResponse(400, "缺少题目集ID"))
		return
	}

	// 从 context 获取用户ID
	uid, exists := c.Get("uid")
	if !exists {
		logger.Warn("未登录用户")
		c.JSON(http.StatusOK, errorResponse(401, "用户未登录"))
		return
	}

	uidStr, ok := uid.(string)
	if !ok {
		logger.Warn("用户ID格式错误")
		c.JSON(http.StatusOK, errorResponse(400, "用户ID格式错误"))
		return
	}

	// 查询题目集（包含题目和样例）
	var problemSet model.ProblemSet
	if err := h.db.Where("id = ? AND user_id = ?", id, uidStr).
		Preload("Problems", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Preload("Problems.Examples", func(db *gorm.DB) *gorm.DB {
			return db.Order("example_no ASC")
		}).
		First(&problemSet).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Warn("题目集不存在或无权访问", zap.String("id", id), zap.String("uid", uidStr))
			c.JSON(http.StatusOK, errorResponse(404, "题目集不存在或无权访问"))
			return
		}
		logger.Error("查询题目集失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	// 检查是否有题目
	if len(problemSet.Problems) == 0 {
		logger.Warn("题目集没有题目", zap.String("id", id))
		c.JSON(http.StatusOK, errorResponse(400, "题目集没有题目，无法生成PDF"))
		return
	}

	// 查询题目集的所有图片
	var images []model.ProblemSetImage
	if err := h.db.Where("set_id = ?", id).Find(&images).Error; err != nil {
		logger.Error("查询图片失败", zap.Error(err))
		// 继续处理，没有图片也可以生成 PDF
		images = []model.ProblemSetImage{}
	}

	// 生成 PDF
	pdfBytes, err := h.problemtoolsPDFGenerator.GenerateProblemSetPDF(&problemSet, images)
	if err != nil {
		logger.Error("生成PDF失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "生成PDF失败: "+err.Error()))
		return
	}

	// 验证 PDF 内容
	logger.Info("PDF生成成功，准备发送", zap.Int("pdf_size", len(pdfBytes)))

	// 设置响应头
	filename := fmt.Sprintf("%s_%s.pdf", problemSet.Title, strconv.FormatUint(problemSet.ID, 10))
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
	c.Header("Content-Length", fmt.Sprintf("%d", len(pdfBytes)))
	c.Data(http.StatusOK, "application/pdf", pdfBytes)

	logger.Info("生成PDF成功", zap.String("id", id), zap.Int("size", len(pdfBytes)))
}

// ReorderProblemsRequest 重新排序题目请求
type ReorderProblemsRequest struct {
	ProblemIDs []uint64 `json:"problem_ids" binding:"required"`
}

// ReorderProblems 批量重新排序题目
func (h *ProblemSetHandler) ReorderProblems(c *gin.Context) {
	logger := utils.GetLogger()

	// 获取题目集ID
	setID := c.Param("id")
	if setID == "" {
		logger.Warn("缺少题目集ID")
		c.JSON(http.StatusOK, errorResponse(400, "缺少题目集ID"))
		return
	}

	// 从 context 获取用户ID
	uid, exists := c.Get("uid")
	if !exists {
		logger.Warn("未登录用户")
		c.JSON(http.StatusOK, errorResponse(401, "用户未登录"))
		return
	}

	uidStr, ok := uid.(string)
	if !ok {
		logger.Warn("用户ID格式错误")
		c.JSON(http.StatusOK, errorResponse(400, "用户ID格式错误"))
		return
	}

	// 验证题目集所有权
	var problemSet model.ProblemSet
	if err := h.db.Where("id = ? AND user_id = ?", setID, uidStr).First(&problemSet).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Warn("题目集不存在或无权访问", zap.String("id", setID), zap.String("uid", uidStr))
			c.JSON(http.StatusOK, errorResponse(404, "题目集不存在或无权访问"))
			return
		}
		logger.Error("查询题目集失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	// 解析请求
	var req ReorderProblemsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "请求参数错误: "+err.Error()))
		return
	}

	// 验证所有题目都属于该题目集
	var problems []model.ProblemSetProblem
	if err := h.db.Where("set_id = ? AND id IN ?", setID, req.ProblemIDs).Find(&problems).Error; err != nil {
		logger.Error("查询题目失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	if len(problems) != len(req.ProblemIDs) {
		logger.Warn("部分题目不存在或不属于该题目集", zap.Int("requested", len(req.ProblemIDs)), zap.Int("found", len(problems)))
		c.JSON(http.StatusOK, errorResponse(400, "部分题目不存在或不属于该题目集"))
		return
	}

	// 更新每个题目的排序
	for i, problemID := range req.ProblemIDs {
		if err := h.db.Model(&model.ProblemSetProblem{}).
			Where("id = ?", problemID).
			Update("sort_order", i).Error; err != nil {
			logger.Error("更新题目排序失败", zap.Error(err))
			c.JSON(http.StatusOK, errorResponse(500, "更新排序失败"))
			return
		}
	}

	logger.Info("重新排序题目成功", zap.String("set_id", setID), zap.Int("count", len(req.ProblemIDs)))
	c.JSON(http.StatusOK, successResponse(nil))
}

// MoveProblemUp 上移题目
func (h *ProblemSetHandler) MoveProblemUp(c *gin.Context) {
	logger := utils.GetLogger()

	// 获取题目ID
	problemID := c.Param("problemId")
	if problemID == "" {
		logger.Warn("缺少题目ID")
		c.JSON(http.StatusOK, errorResponse(400, "缺少题目ID"))
		return
	}

	// 从 context 获取用户ID
	uid, exists := c.Get("uid")
	if !exists {
		logger.Warn("未登录用户")
		c.JSON(http.StatusOK, errorResponse(401, "用户未登录"))
		return
	}

	uidStr, ok := uid.(string)
	if !ok {
		logger.Warn("用户ID格式错误")
		c.JSON(http.StatusOK, errorResponse(400, "用户ID格式错误"))
		return
	}

	// 查询当前题目
	var currentProblem model.ProblemSetProblem
	if err := h.db.Joins("JOIN xcpc_problem_set ON xcpc_problem_set_problem.set_id = xcpc_problem_set.id").
		Where("xcpc_problem_set_problem.id = ? AND xcpc_problem_set.user_id = ?", problemID, uidStr).
		First(&currentProblem).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Warn("题目不存在或无权访问", zap.String("problem_id", problemID))
			c.JSON(http.StatusOK, errorResponse(404, "题目不存在或无权访问"))
			return
		}
		logger.Error("查询题目失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	// 查找排序在当前题目之前的题目
	var prevProblem model.ProblemSetProblem
	if err := h.db.Where("set_id = ? AND sort_order < ?", currentProblem.SetID, currentProblem.SortOrder).
		Order("sort_order DESC").
		First(&prevProblem).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Info("题目已在顶部", zap.String("problem_id", problemID))
			c.JSON(http.StatusOK, successResponse(nil)) // 已经在顶部，不需要移动
			return
		}
		logger.Error("查询前一个题目失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	// 交换排序值
	currentSortOrder := currentProblem.SortOrder
	prevSortOrder := prevProblem.SortOrder
	currentProblemID := currentProblem.ID
	prevProblemID := prevProblem.ID

	// 使用 where 条件直接更新，避免模型实例问题
	if err := h.db.Model(&model.ProblemSetProblem{}).
		Where("id = ?", currentProblemID).
		Update("sort_order", prevSortOrder).Error; err != nil {
		logger.Error("更新当前题目排序失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "更新排序失败"))
		return
	}

	if err := h.db.Model(&model.ProblemSetProblem{}).
		Where("id = ?", prevProblemID).
		Update("sort_order", currentSortOrder).Error; err != nil {
		logger.Error("更新前一个题目排序失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "更新排序失败"))
		return
	}

	logger.Info("上移题目成功", zap.String("problem_id", problemID), zap.Int("current_sort", currentSortOrder), zap.Int("new_sort", prevSortOrder))
	c.JSON(http.StatusOK, successResponse(nil))
}

// MoveProblemDown 下移题目
func (h *ProblemSetHandler) MoveProblemDown(c *gin.Context) {
	logger := utils.GetLogger()

	// 获取题目ID
	problemID := c.Param("problemId")
	if problemID == "" {
		logger.Warn("缺少题目ID")
		c.JSON(http.StatusOK, errorResponse(400, "缺少题目ID"))
		return
	}

	// 从 context 获取用户ID
	uid, exists := c.Get("uid")
	if !exists {
		logger.Warn("未登录用户")
		c.JSON(http.StatusOK, errorResponse(401, "用户未登录"))
		return
	}

	uidStr, ok := uid.(string)
	if !ok {
		logger.Warn("用户ID格式错误")
		c.JSON(http.StatusOK, errorResponse(400, "用户ID格式错误"))
		return
	}

	// 查询当前题目
	var currentProblem model.ProblemSetProblem
	if err := h.db.Joins("JOIN xcpc_problem_set ON xcpc_problem_set_problem.set_id = xcpc_problem_set.id").
		Where("xcpc_problem_set_problem.id = ? AND xcpc_problem_set.user_id = ?", problemID, uidStr).
		First(&currentProblem).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Warn("题目不存在或无权访问", zap.String("problem_id", problemID))
			c.JSON(http.StatusOK, errorResponse(404, "题目不存在或无权访问"))
			return
		}
		logger.Error("查询题目失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	// 查找排序在当前题目之后的题目
	var nextProblem model.ProblemSetProblem
	if err := h.db.Where("set_id = ? AND sort_order > ?", currentProblem.SetID, currentProblem.SortOrder).
		Order("sort_order ASC").
		First(&nextProblem).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Info("题目已在底部", zap.String("problem_id", problemID))
			c.JSON(http.StatusOK, successResponse(nil)) // 已经在底部，不需要移动
			return
		}
		logger.Error("查询后一个题目失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	// 交换排序值
	currentSortOrder := currentProblem.SortOrder
	nextSortOrder := nextProblem.SortOrder
	currentProblemID := currentProblem.ID
	nextProblemID := nextProblem.ID

	// 使用 where 条件直接更新，避免模型实例问题
	if err := h.db.Model(&model.ProblemSetProblem{}).
		Where("id = ?", currentProblemID).
		Update("sort_order", nextSortOrder).Error; err != nil {
		logger.Error("更新当前题目排序失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "更新排序失败"))
		return
	}

	if err := h.db.Model(&model.ProblemSetProblem{}).
		Where("id = ?", nextProblemID).
		Update("sort_order", currentSortOrder).Error; err != nil {
		logger.Error("更新后一个题目排序失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "更新排序失败"))
		return
	}

	logger.Info("下移题目成功", zap.String("problem_id", problemID), zap.Int("current_sort", currentSortOrder), zap.Int("new_sort", nextSortOrder))
	c.JSON(http.StatusOK, successResponse(nil))
}

// GeneratePDFFromData 从前端数据直接生成PDF（所见即所得）
type GeneratePDFFromDataRequest struct {
	Title       string                `json:"title"`
	Author      string                `json:"author"`
	ContestDate string                `json:"contest_date"`
	Problems    []ProblemDataForPDF   `json:"problems"`
}

type ProblemDataForPDF struct {
	ProblemLetter string              `json:"problem_letter"`
	Title         string              `json:"title"`
	TimeLimit     int                 `json:"time_limit"`
	Description   string              `json:"description"`
	InputFormat   string              `json:"input_format"`
	OutputFormat  string              `json:"output_format"`
	Note          string              `json:"note"`
	Examples      []ExampleDataForPDF `json:"examples"`
	SortOrder     int                 `json:"sort_order"`
}

type ExampleDataForPDF struct {
	ExampleNo   int    `json:"example_no"`
	Description string `json:"description"`
	Input       string `json:"input"`
	Output      string `json:"output"`
}

func (h *ProblemSetHandler) GeneratePDFFromData(c *gin.Context) {
	logger := utils.GetLogger()

	// 从 context 获取用户ID
	uid, exists := c.Get("uid")
	if !exists {
		logger.Warn("未登录用户")
		c.JSON(http.StatusOK, errorResponse(401, "用户未登录"))
		return
	}

	uidStr, ok := uid.(string)
	if !ok {
		logger.Warn("用户ID格式错误")
		c.JSON(http.StatusOK, errorResponse(400, "用户ID格式错误"))
		return
	}

	// 解析请求
	var req GeneratePDFFromDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("请求参数错误", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "请求参数错误: "+err.Error()))
		return
	}

	// 构建临时的 ProblemSet 对象（不保存到数据库）
	problemSet := &model.ProblemSet{
		Title:    req.Title,
		Author:   req.Author,
		Problems: make([]model.ProblemSetProblem, 0, len(req.Problems)),
	}

	// 解析日期
	if req.ContestDate != "" {
		if t, err := time.Parse("2006-01-02", req.ContestDate); err == nil {
			problemSet.ContestDate = &t
		}
	}

	// 转换题目数据
	for i, p := range req.Problems {
		problemSetProblem := model.ProblemSetProblem{
			ID:            uint64(i + 1), // 临时ID，从1开始
			ProblemLetter: p.ProblemLetter,
			Title:         p.Title,
			TimeLimit:     p.TimeLimit,
			Description:   p.Description,
			InputFormat:   p.InputFormat,
			OutputFormat:  p.OutputFormat,
			Note:          p.Note,
			SortOrder:     p.SortOrder,
			Examples:      make([]model.ProblemExample, 0, len(p.Examples)),
		}

		// 转换样例数据
		for _, ex := range p.Examples {
			problemSetProblem.Examples = append(problemSetProblem.Examples, model.ProblemExample{
				ExampleNo:   ex.ExampleNo,
				Description: ex.Description,
				Input:       ex.Input,
				Output:      ex.Output,
			})
		}

		problemSet.Problems = append(problemSet.Problems, problemSetProblem)
	}

	logger.Info("从前端数据生成PDF", zap.String("uid", uidStr), zap.String("title", req.Title), zap.Int("problems_count", len(req.Problems)))

	// 生成 PDF（从前端数据生成，没有图片支持）
	pdfBytes, err := h.problemtoolsPDFGenerator.GenerateProblemSetPDF(problemSet, []model.ProblemSetImage{})
	if err != nil {
		logger.Error("生成PDF失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "生成PDF失败: "+err.Error()))
		return
	}

	// 设置响应头
	filename := fmt.Sprintf("%s.pdf", req.Title)
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
	c.Header("Content-Length", fmt.Sprintf("%d", len(pdfBytes)))
	c.Data(http.StatusOK, "application/pdf", pdfBytes)

	logger.Info("从前端数据生成PDF成功", zap.String("uid", uidStr), zap.Int("size", len(pdfBytes)))
}

// DiagnosticPDF 诊断PDF生成 - 用于线上排查问题
func (h *ProblemSetHandler) DiagnosticPDF(c *gin.Context) {
	logger := utils.GetLogger()

	// 获取题目集ID
	id := c.Param("id")
	if id == "" {
		logger.Warn("缺少题目集ID")
		c.JSON(http.StatusOK, errorResponse(400, "缺少题目集ID"))
		return
	}

	// 从 context 获取用户ID
	uid, exists := c.Get("uid")
	if !exists {
		logger.Warn("未登录用户")
		c.JSON(http.StatusOK, errorResponse(401, "用户未登录"))
		return
	}

	uidStr, ok := uid.(string)
	if !ok {
		logger.Warn("用户ID格式错误")
		c.JSON(http.StatusOK, errorResponse(400, "用户ID格式错误"))
		return
	}

	// 查询题目集（包含题目和样例）
	var problemSet model.ProblemSet
	if err := h.db.Where("id = ? AND user_id = ?", id, uidStr).
		Preload("Problems", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Preload("Problems.Examples", func(db *gorm.DB) *gorm.DB {
			return db.Order("example_no ASC")
		}).
		First(&problemSet).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Warn("题目集不存在或无权访问", zap.String("id", id), zap.String("uid", uidStr))
			c.JSON(http.StatusOK, errorResponse(404, "题目集不存在或无权访问"))
			return
		}
		logger.Error("查询题目集失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	// 构建诊断信息
	diagnostic := map[string]interface{}{
		"problem_set_id":   problemSet.ID,
		"title":            problemSet.Title,
		"author":           problemSet.Author,
		"problems_count":   len(problemSet.Problems),
		"template_dir":     "./templates/pdf", // 从配置中获取
		"problems":         []map[string]interface{}{},
	}

	for _, p := range problemSet.Problems {
		problemInfo := map[string]interface{}{
			"id":            p.ID,
			"problem_letter": p.ProblemLetter,
			"title":          p.Title,
			"time_limit":     p.TimeLimit,
			"examples_count": len(p.Examples),
		}
		diagnostic["problems"] = append(diagnostic["problems"].([]map[string]interface{}), problemInfo)
	}

	logger.Info("PDF诊断", zap.Any("diagnostic", diagnostic))
	c.JSON(http.StatusOK, successResponse(diagnostic))
}

// UploadImage 上传题目集图片
func (h *ProblemSetHandler) UploadImage(c *gin.Context) {
	logger := utils.GetLogger()

	// 获取题目集ID
	setID := c.Param("id")
	if setID == "" {
		logger.Warn("缺少题目集ID")
		c.JSON(http.StatusOK, errorResponse(400, "缺少题目集ID"))
		return
	}

	// 从 context 获取用户ID
	uid, exists := c.Get("uid")
	if !exists {
		logger.Warn("未登录用户")
		c.JSON(http.StatusOK, errorResponse(401, "用户未登录"))
		return
	}

	uidStr, ok := uid.(string)
	if !ok {
		logger.Warn("用户ID格式错误")
		c.JSON(http.StatusOK, errorResponse(400, "用户ID格式错误"))
		return
	}

	// 验证题目集所有权
	var problemSet model.ProblemSet
	if err := h.db.Where("id = ? AND user_id = ?", setID, uidStr).First(&problemSet).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Warn("题目集不存在或无权访问", zap.String("id", setID), zap.String("uid", uidStr))
			c.JSON(http.StatusOK, errorResponse(404, "题目集不存在或无权访问"))
			return
		}
		logger.Error("查询题目集失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	// 获取上传的文件
	fileHeader, err := c.FormFile("file")
	if err != nil {
		logger.Warn("获取上传文件失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(400, "获取上传文件失败: "+err.Error()))
		return
	}

	// 验证文件类型（只允许图片）
	allowedExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".bmp":  true,
		".pdf":  true,
	}

	filename := fileHeader.Filename
	ext := filepath.Ext(filename)
	if !allowedExtensions[strings.ToLower(ext)] {
		logger.Warn("不支持的文件类型", zap.String("filename", filename), zap.String("ext", ext))
		c.JSON(http.StatusOK, errorResponse(400, "不支持的文件类型，仅支持: jpg, jpeg, png, gif, bmp, pdf"))
		return
	}

	// 检查文件名是否已存在
	var existingImage model.ProblemSetImage
	if err := h.db.Where("set_id = ? AND filename = ?", setID, filename).First(&existingImage).Error; err == nil {
		logger.Warn("文件名已存在", zap.String("set_id", setID), zap.String("filename", filename))
		c.JSON(http.StatusOK, errorResponse(400, "文件名 '"+filename+"' 已存在，请使用不同的文件名或删除现有文件后重新上传"))
		return
	}

	// 创建存储目录
	storageDir := filepath.Join("/app/problemset_images", setID)
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		logger.Error("创建存储目录失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "创建存储目录失败"))
		return
	}

	// 生成唯一文件名（使用原始文件名，但已在数据库中检查重复）
	storagePath := filepath.Join(storageDir, filename)

	// 保存文件
	if err := c.SaveUploadedFile(fileHeader, storagePath); err != nil {
		logger.Error("保存文件失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "保存文件失败"))
		return
	}

	// 获取文件信息
	fileInfo, err := os.Stat(storagePath)
	if err != nil {
		logger.Error("获取文件信息失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "获取文件信息失败"))
		return
	}

	// 保存到数据库
	setIDUint, _ := strconv.ParseUint(setID, 10, 64)
	image := model.ProblemSetImage{
		SetID:    setIDUint,
		Filename: filename,
		FilePath: storagePath,
		FileSize: fileInfo.Size(),
		MimeType: fileHeader.Header.Get("Content-Type"),
	}

	if err := h.db.Create(&image).Error; err != nil {
		logger.Error("保存图片记录失败", zap.Error(err))
		// 删除已保存的文件
		os.Remove(storagePath)
		c.JSON(http.StatusOK, errorResponse(500, "保存图片记录失败"))
		return
	}

	logger.Info("上传图片成功", zap.String("set_id", setID), zap.String("filename", filename), zap.Int64("size", fileInfo.Size()))
	c.JSON(http.StatusOK, successResponse(image))
}

// ListImages 列出题目集的所有图片
func (h *ProblemSetHandler) ListImages(c *gin.Context) {
	logger := utils.GetLogger()

	// 获取题目集ID
	setID := c.Param("id")
	if setID == "" {
		logger.Warn("缺少题目集ID")
		c.JSON(http.StatusOK, errorResponse(400, "缺少题目集ID"))
		return
	}

	// 从 context 获取用户ID
	uid, exists := c.Get("uid")
	if !exists {
		logger.Warn("未登录用户")
		c.JSON(http.StatusOK, errorResponse(401, "用户未登录"))
		return
	}

	uidStr, ok := uid.(string)
	if !ok {
		logger.Warn("用户ID格式错误")
		c.JSON(http.StatusOK, errorResponse(400, "用户ID格式错误"))
		return
	}

	// 验证题目集所有权
	var problemSet model.ProblemSet
	if err := h.db.Where("id = ? AND user_id = ?", setID, uidStr).First(&problemSet).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Warn("题目集不存在或无权访问", zap.String("id", setID), zap.String("uid", uidStr))
			c.JSON(http.StatusOK, errorResponse(404, "题目集不存在或无权访问"))
			return
		}
		logger.Error("查询题目集失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	// 查询所有图片
	var images []model.ProblemSetImage
	if err := h.db.Where("set_id = ?", setID).Order("created_at DESC").Find(&images).Error; err != nil {
		logger.Error("查询图片列表失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	logger.Info("查询图片列表成功", zap.String("set_id", setID), zap.Int("count", len(images)))
	c.JSON(http.StatusOK, successResponse(images))
}

// DeleteImage 删除题目集图片
func (h *ProblemSetHandler) DeleteImage(c *gin.Context) {
	logger := utils.GetLogger()

	// 获取图片ID
	imageID := c.Param("imageId")
	if imageID == "" {
		logger.Warn("缺少图片ID")
		c.JSON(http.StatusOK, errorResponse(400, "缺少图片ID"))
		return
	}

	// 从 context 获取用户ID
	uid, exists := c.Get("uid")
	if !exists {
		logger.Warn("未登录用户")
		c.JSON(http.StatusOK, errorResponse(401, "用户未登录"))
		return
	}

	uidStr, ok := uid.(string)
	if !ok {
		logger.Warn("用户ID格式错误")
		c.JSON(http.StatusOK, errorResponse(400, "用户ID格式错误"))
		return
	}

	// 查询图片（需要验证用户所有权）
	var image model.ProblemSetImage
	if err := h.db.Joins("JOIN xcpc_problem_set ON xcpc_problem_set_image.set_id = xcpc_problem_set.id").
		Where("xcpc_problem_set_image.id = ? AND xcpc_problem_set.user_id = ?", imageID, uidStr).
		First(&image).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Warn("图片不存在或无权访问", zap.String("image_id", imageID), zap.String("uid", uidStr))
			c.JSON(http.StatusOK, errorResponse(404, "图片不存在或无权访问"))
			return
		}
		logger.Error("查询图片失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "查询失败"))
		return
	}

	// 真删除：从数据库删除记录
	if err := h.db.Delete(&image).Error; err != nil {
		logger.Error("删除图片记录失败", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "删除失败"))
		return
	}

	// 真删除：从磁盘删除物理文件
	if err := os.Remove(image.FilePath); err != nil {
		// 文件不存在或删除失败，记录警告但不影响操作
		if !os.IsNotExist(err) {
			logger.Warn("删除物理文件失败", zap.String("path", image.FilePath), zap.Error(err))
		}
	}

	logger.Info("删除图片成功", zap.String("image_id", imageID), zap.String("filename", image.Filename))
	c.JSON(http.StatusOK, successResponse(nil))
}
