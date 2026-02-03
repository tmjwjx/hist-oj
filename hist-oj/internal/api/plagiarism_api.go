package api

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

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

// PlagiarismHandler 查重处理器
type PlagiarismHandler struct {
	db                *gorm.DB
	plagiarismService *service.PlagiarismService
}

// NewPlagiarismHandler 创建查重处理器
func NewPlagiarismHandler(db *gorm.DB) *PlagiarismHandler {
	return &PlagiarismHandler{
		db:                db,
		plagiarismService: service.NewPlagiarismService(db),
	}
}

// RegisterPlagiarismRoutes 注册查重路由
func RegisterPlagiarismRoutes(api *gin.RouterGroup, cfg *config.Config, db *gorm.DB) {
	handler := NewPlagiarismHandler(db)

	plagiarism := api.Group("/plagiarism")
	{
		// 测试路由 - 无需认证，用于验证路由是否注册成功
		plagiarism.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"code":    200,
				"message": "查重路由已注册",
			})
		})

		// 查重配置 - 需要比赛创建者或超级管理员权限
		plagiarism.GET("/contest/:id/config", handler.contestOwnerAuth(cfg, db), handler.GetConfig)
		plagiarism.POST("/contest/:id/config", handler.contestOwnerAuth(cfg, db), handler.SaveConfig)

		// 查重任务 - 需要比赛创建者或超级管理员权限
		plagiarism.POST("/contest/:id/check/start", handler.contestOwnerAuth(cfg, db), handler.StartCheck)
		plagiarism.GET("/contest/:id/check/progress", handler.contestOwnerAuth(cfg, db), handler.GetProgress)
		plagiarism.GET("/contest/:id/check/latest", handler.contestOwnerAuth(cfg, db), handler.GetLatestCheck)

		// 查重结果 - 需要比赛创建者或超级管理员权限
		plagiarism.GET("/check/:id/results", handler.contestOwnerAuth(cfg, db), handler.GetResults)
		plagiarism.GET("/check/:id/results/export", handler.contestOwnerAuth(cfg, db), handler.ExportResults)

		// 获取提交详情 - 需要比赛创建者或超级管理员权限
		plagiarism.GET("/submission/:id", handler.contestOwnerAuth(cfg, db), handler.GetSubmission)
	}
}

// GetConfig 获取查重配置
// @Summary 获取查重配置
// @Description 获取指定比赛的查重配置
// @Tags plagiarism
// @Param id path int true "比赛ID"
// @Success 200 {object} Response
// @Router /api/admin/plagiarism/contest/{id}/config [get]
func (h *PlagiarismHandler) GetConfig(c *gin.Context) {
	cidStr := c.Param("id")
	cid, err := strconv.ParseUint(cidStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的比赛ID",
		})
		return
	}

	configs, err := h.plagiarismService.GetConfig(cid)
	if err != nil {
		utils.GetLogger().Error("获取查重配置失败", zap.Uint64("cid", cid), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取查重配置失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"data":    configs,
		"message": "获取成功",
	})
}

// SaveConfigRequest 保存配置请求
type SaveConfigRequest struct {
	Configs []struct {
		CPID      uint64 `json:"cpid"`
		Threshold int    `json:"threshold"`
	} `json:"configs"`
}

// SaveConfig 保存查重配置
// @Summary 保存查重配置
// @Description 保存指定比赛的查重配置
// @Tags plagiarism
// @Param id path int true "比赛ID"
// @Param request body SaveConfigRequest true "配置数据"
// @Success 200 {object} Response
// @Router /api/admin/plagiarism/contest/{id}/config [post]
func (h *PlagiarismHandler) SaveConfig(c *gin.Context) {
	cidStr := c.Param("id")
	cid, err := strconv.ParseUint(cidStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的比赛ID",
		})
		return
	}

	var req SaveConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的请求数据",
		})
		return
	}

	// 获取当前用户ID
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "用户未登录",
		})
		return
	}

	// 转换配置
	configs := make([]model.PlagiarismCheckConfig, len(req.Configs))
	for i, cfg := range req.Configs {
		configs[i] = model.PlagiarismCheckConfig{
			CPID:      cfg.CPID,
			Threshold: cfg.Threshold,
		}
	}

	// 检查权限（超级管理员或比赛创建者）
	if !h.checkPermission(c, cid, userId.(string)) {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "权限不足",
		})
		return
	}

	if err := h.plagiarismService.SaveConfig(cid, configs, userId.(string)); err != nil {
		utils.GetLogger().Error("保存查重配置失败", zap.Uint64("cid", cid), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "保存查重配置失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "保存成功",
	})
}

// StartCheck 开始查重
// @Summary 开始查重
// @Description 开始执行查重任务
// @Tags plagiarism
// @Param id path int true "比赛ID"
// @Success 200 {object} Response
// @Router /api/admin/plagiarism/contest/{id}/check/start [post]
func (h *PlagiarismHandler) StartCheck(c *gin.Context) {
	cidStr := c.Param("id")
	cid, err := strconv.ParseUint(cidStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的比赛ID",
		})
		return
	}

	// 获取当前用户ID
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "用户未登录",
		})
		return
	}

	// 检查权限
	if !h.checkPermission(c, cid, userId.(string)) {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "权限不足",
		})
		return
	}

	// 创建查重任务
	check, err := h.plagiarismService.CreateCheck(cid, userId.(string))
	if err != nil {
		utils.GetLogger().Error("创建查重任务失败", zap.Uint64("cid", cid), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	// 启动查重任务
	go func() {
		if err := h.plagiarismService.RunCheck(check.ID); err != nil {
			utils.GetLogger().Error("启动查重任务失败", zap.Uint64("checkId", check.ID), zap.Error(err))
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"data":    check,
		"message": "查重任务已启动",
	})
}

// GetProgress 获取查重进度
// @Summary 获取查重进度
// @Description 获取查重任务的执行进度
// @Tags plagiarism
// @Param id path int true "比赛ID"
// @Success 200 {object} Response
// @Router /api/admin/plagiarism/contest/{id}/check/progress [get]
func (h *PlagiarismHandler) GetProgress(c *gin.Context) {
	cidStr := c.Param("id")
	cid, err := strconv.ParseUint(cidStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的比赛ID",
		})
		return
	}

	check, err := h.plagiarismService.GetLatestCheck(cid)
	if err != nil {
		utils.GetLogger().Error("获取查重进度失败", zap.Uint64("cid", cid), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取查重进度失败",
		})
		return
	}

	if check == nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"data":    nil,
			"message": "暂无查重任务",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"data":    check,
		"message": "获取成功",
	})
}

// GetLatestCheck 获取最新查重任务
// @Summary 获取最新查重任务
// @Description 获取指定比赛的最新查重任务
// @Tags plagiarism
// @Param id path int true "比赛ID"
// @Success 200 {object} Response
// @Router /api/admin/plagiarism/contest/{id}/check/latest [get]
func (h *PlagiarismHandler) GetLatestCheck(c *gin.Context) {
	cidStr := c.Param("id")
	cid, err := strconv.ParseUint(cidStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的比赛ID",
		})
		return
	}

	check, err := h.plagiarismService.GetLatestCheck(cid)
	if err != nil {
		utils.GetLogger().Error("获取查重任务失败", zap.Uint64("cid", cid), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取查重任务失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"data":    check,
		"message": "获取成功",
	})
}

// GetResults 获取查重结果
// @Summary 获取查重结果
// @Description 获取查重任务的结果列表（只返回超过阈值的）
// @Tags plagiarism
// @Param id path int true "查重任务ID"
// @Param displayId query string false "题目编号（可选，用于筛选）"
// @Success 200 {object} Response
// @Router /api/admin/plagiarism/check/{id}/results [get]
func (h *PlagiarismHandler) GetResults(c *gin.Context) {
	checkIdStr := c.Param("id")
	checkId, err := strconv.ParseUint(checkIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的查重任务ID",
		})
		return
	}

	// 获取题号筛选参数（可选）
	displayId := c.Query("displayId")

	// 调用 service 获取结果和总数
	results, totalCount, err := h.plagiarismService.GetResults(checkId, displayId)
	if err != nil {
		utils.GetLogger().Error("获取查重结果失败", zap.Uint64("checkId", checkId), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取查重结果失败",
		})
		return
	}

	utils.GetLogger().Info("GetResults: 返回结果", zap.Uint64("checkId", checkId),
		zap.String("displayId", displayId), zap.Int("count", len(results)),
		zap.Int64("totalCount", totalCount))

	// 返回结果和总数
	c.JSON(http.StatusOK, gin.H{
		"code":       200,
		"data":       results,
		"totalCount": totalCount,
		"message":    "获取成功",
	})
}

// ExportResults 导出查重结果（使用流式响应，避免大数据量超时）
// @Summary 导出查重结果
// @Description 导出查重结果为Excel
// @Tags plagiarism
// @Param id path int true "查重任务ID"
// @Success 200 {file} file
// @Router /api/admin/plagiarism/check/{id}/results/export [get]
func (h *PlagiarismHandler) ExportResults(c *gin.Context) {
	checkIdStr := c.Param("id")
	checkId, err := strconv.ParseUint(checkIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的查重任务ID",
		})
		return
	}

	utils.GetLogger().Info("ExportResults API 被调用", zap.Uint64("checkId", checkId))

	// 先统计结果数量
	var totalCount int64
	if err := h.db.Raw("SELECT COUNT(*) FROM plagiarism_result WHERE check_id = ?", checkId).Scan(&totalCount).Error; err != nil {
		utils.GetLogger().Error("统计查重结果失败", zap.Uint64("checkId", checkId), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "统计查重结果失败",
		})
		return
	}

	utils.GetLogger().Info("ExportResults: 数据统计", zap.Uint64("checkId", checkId), zap.Int64("totalCount", totalCount))

	// 设置响应头（使用流式传输）
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=plagiarism_results_%d.csv", checkId))
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
	c.Header("X-Accel-Buffering", "no") // 禁用 nginx 缓冲

	// 使用流式响应
	c.Stream(func(w io.Writer) bool {
		// 写入 BOM 以支持 Excel 打开中文
		w.Write([]byte{0xEF, 0xBB, 0xBF})

		// 写入 CSV 头
		w.Write([]byte("题目编号,题目标题,用户1,用户2,语言,相似度1to2(%),相似度2to1(%),最大相似度(%),超过阈值\n"))

		// 分批查询并写入（使用游标）
		batchSize := 500
		lastID := uint64(0)
		totalExported := 0

		for {
			var results []model.PlagiarismResult
			err := h.db.Where("check_id = ? AND id > ?", checkId, lastID).
				Order("id ASC").
				Limit(batchSize).
				Find(&results).Error

			if err != nil {
				utils.GetLogger().Error("查询查重结果失败", zap.Uint64("checkId", checkId), zap.Error(err))
				return false
			}

			if len(results) == 0 {
				break
			}

			// 写入数据
			for _, r := range results {
				overThreshold := "否"
				if r.IsOverThreshold {
					overThreshold = "是"
				}

				line := fmt.Sprintf("%s,%s,%s,%s,%s,%d,%d,%d,%s\n",
					csvFieldEscapeAPI(r.DisplayID),
					csvFieldEscapeAPI(r.ProblemTitle),
					csvFieldEscapeAPI(r.Username1),
					csvFieldEscapeAPI(r.Username2),
					csvFieldEscapeAPI(r.Language),
					r.Similarity1to2,
					r.Similarity2to1,
					r.MaxSimilarity,
					csvFieldEscapeAPI(overThreshold),
				)
				w.Write([]byte(line))
				lastID = r.ID
				totalExported++
			}

			// 每 5000 条记录打印一次日志
			if totalExported%5000 == 0 {
				utils.GetLogger().Info("ExportResults: 导出进度", zap.Uint64("checkId", checkId), zap.Int("exported", totalExported))
			}

			// 如果返回的记录数少于批次大小，说明已经到最后一批
			if len(results) < batchSize {
				break
			}
		}

		utils.GetLogger().Info("ExportResults: 流式导出完成", zap.Uint64("checkId", checkId), zap.Int("totalExported", totalExported))
		return false // 结束流
	})
}

// csvFieldEscapeAPI CSV 字段转义（API 版本）
func csvFieldEscapeAPI(field string) string {
	if field == "" {
		return ""
	}
	// 如果字段包含逗号、引号或换行符，需要用引号包裹，并将内部引号转义
	if strings.ContainsAny(field, "\",\"\n\r") {
		// 将引号转义为两个引号
		escaped := strings.ReplaceAll(field, "\"", "\"\"")
		return fmt.Sprintf("\"%s\"", escaped)
	}
	return field
}

// GetSubmission 获取提交详情
// @Summary 获取提交详情
// @Description 获取指定提交的详情和代码
// @Tags plagiarism
// @Param id path int true "提交ID"
// @Success 200 {object} Response
// @Router /api/admin/plagiarism/submission/{id} [get]
func (h *PlagiarismHandler) GetSubmission(c *gin.Context) {
	submitIdStr := c.Param("id")
	submitId, err := strconv.ParseUint(submitIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的提交ID",
		})
		return
	}

	submission, err := h.plagiarismService.GetSubmission(submitId)
	if err != nil {
		utils.GetLogger().Error("获取提交详情失败", zap.Uint64("submitId", submitId), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取提交详情失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"data":    submission,
		"message": "获取成功",
	})
}

// checkPermission 检查权限（超级管理员或比赛创建者）
func (h *PlagiarismHandler) checkPermission(c *gin.Context, cid uint64, userId string) bool {
	// 获取用户角色
	roles, err := middlewarepkg.GetUserRoles(h.db, userId)
	if err != nil {
		return false
	}

	// 超级管理员
	if middlewarepkg.HasRole(roles, middlewarepkg.RoleRoot) {
		return true
	}

	// 检查是否为比赛创建者
	var contest model.Contest
	if err := h.db.Where("id = ?", cid).First(&contest).Error; err != nil {
		return false
	}

	return contest.UID == userId
}

// contestOwnerAuth 比赛创建者权限中间件
// 只有比赛创建者或超级管理员可以访问
// 权限不足返回 404（而不是 403）
// 使用 HOJ API 验证 token，而不是自己验证 JWT（避免 secret 不匹配问题）
func (h *PlagiarismHandler) contestOwnerAuth(cfg *config.Config, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := utils.GetLogger()

		logger.Info("contestOwnerAuth: 开始验证", zap.String("path", c.Request.URL.Path), zap.String("method", c.Request.Method))

		// 从 Header 中获取 token
		token := c.GetHeader("Authorization")
		if token == "" {
			logger.Warn("未提供认证token", zap.String("path", c.Request.URL.Path))
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "Not Found",
			})
			c.Abort()
			return
		}

		logger.Info("contestOwnerAuth: Token已获取，开始验证")
		// 调用 HOJ API 验证 token（与 AuthMiddleware 相同的方式）
		userAuth, err := client.ValidateToken(token)
		if err != nil {
			logger.Warn("Token验证失败", zap.Error(err), zap.String("path", c.Request.URL.Path))
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "Not Found",
			})
			c.Abort()
			return
		}

		logger.Info("contestOwnerAuth: Token验证成功", zap.String("uid", userAuth.UID))

		// 将用户信息存入context，供后续handler使用
		userId := userAuth.UID
		c.Set("userId", userId)
		c.Set("uid", userId)
		c.Set("username", userAuth.Username)
		c.Set("roles", userAuth.Roles)

		logger.Info("用户认证成功",
			zap.String("uid", userId),
			zap.String("username", userAuth.Username),
			zap.String("path", c.Request.URL.Path))

		// 获取比赛ID（从路径参数中提取）
		cidStr := c.Param("id")
		if cidStr == "" {
			// 某些路由可能没有 id 参数
			c.Next()
			return
		}

		// 检查是否为 /check/:id/... 或 /submission/:id 路由（查重任务/提交相关路由）
		// 这类路由的 :id 是查重任务ID或提交ID，需要先查询获取比赛ID
		var cid uint64

		// 注意：/contest/:id/check/latest 路径包含 /check/ 但 :id 是比赛ID
		// 所以需要精确匹配：只有 /check/:id/... 或 /submission/:id 路由时，:id 才是查重任务/提交ID
		pathParts := strings.Split(c.Request.URL.Path, "/")
		// 实际路径格式: /plagiarism-api/api/plagiarism/contest/:id/... 或 /plagiarism-api/api/plagiarism/check/:id/...
		// pathParts[0]="", [1]="plagiarism-api", [2]="api", [3]="plagiarism", [4]="contest"或"check"或"submission", ...

		isCheckOrSubmissionRoute := false
		if len(pathParts) >= 5 {
			routeSegment := pathParts[4] // contest/check/submission
			if routeSegment == "check" || routeSegment == "submission" {
				isCheckOrSubmissionRoute = true
			}
		}

		if isCheckOrSubmissionRoute {
			// :id 是查重任务ID或提交ID，需要查询获取比赛ID
			id, parseErr := strconv.ParseUint(cidStr, 10, 64)
			if parseErr != nil {
				logger.Warn("无效的ID", zap.String("id", cidStr), zap.Error(parseErr))
				c.JSON(http.StatusNotFound, gin.H{
					"code":    404,
					"message": "Not Found",
				})
				c.Abort()
				return
			}

			// 根据路由类型查询不同的表
			if len(pathParts) >= 5 && pathParts[4] == "check" {
				// 查询查重任务获取比赛ID
				var check model.PlagiarismCheck
				if queryErr := h.db.Where("id = ?", id).First(&check).Error; queryErr != nil {
					logger.Warn("查重任务不存在", zap.Uint64("checkId", id), zap.Error(queryErr))
					c.JSON(http.StatusNotFound, gin.H{
						"code":    404,
						"message": "Not Found",
					})
					c.Abort()
					return
				}
				cid = check.CID
			} else if len(pathParts) >= 5 && pathParts[4] == "submission" {
				// 查询提交记录获取比赛ID（如果需要的话）
				// 目前提交相关路由可能不需要比赛ID验证
				c.Next()
				return
			}
		} else {
			// :id 是比赛ID，直接解析
			parsedCid, parseErr := strconv.ParseUint(cidStr, 10, 64)
			if parseErr != nil {
				logger.Warn("无效的比赛ID", zap.String("cid", cidStr), zap.Error(parseErr))
				c.JSON(http.StatusNotFound, gin.H{
					"code":    404,
					"message": "Not Found",
				})
				c.Abort()
				return
			}
			cid = parsedCid
		}

		// 检查权限（超级管理员或比赛创建者）
		roles, err := middlewarepkg.GetUserRoles(h.db, userId)
		if err != nil || !middlewarepkg.HasRole(roles, middlewarepkg.RoleRoot) {
			// 不是超级管理员，检查是否为比赛创建者
			var contest model.Contest
			if err := h.db.Where("id = ?", cid).First(&contest).Error; err != nil {
				logger.Warn("比赛不存在", zap.Uint64("cid", cid), zap.Error(err))
				c.JSON(http.StatusNotFound, gin.H{
					"code":    404,
					"message": "Not Found",
				})
				c.Abort()
				return
			}

			if contest.UID != userId {
				logger.Warn("权限不足：用户不是比赛创建者或超级管理员",
					zap.String("uid", userId),
					zap.Uint64("cid", cid),
					zap.String("path", c.Request.URL.Path))
				c.JSON(http.StatusNotFound, gin.H{
					"code":    404,
					"message": "Not Found",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
