package api

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/hoj/hist-oj/internal/service"
	"github.com/hoj/hist-oj/internal/utils"
)

type LearningMapAPI struct {
	service *service.LearningMapService
	logger  *zap.Logger
}

func NewLearningMapAPI(svc *service.LearningMapService) *LearningMapAPI {
	return &LearningMapAPI{service: svc, logger: utils.GetLogger()}
}

func parseUintParam(c *gin.Context, key string) (uint64, error) {
	val := strings.TrimSpace(c.Param(key))
	if val == "" {
		return 0, errors.New(key + " is required")
	}
	parsed, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		return 0, errors.New(key + " is invalid")
	}
	return parsed, nil
}

func mapServiceErrorToResponse(err error) (int, int, string) {
	if err == nil {
		return http.StatusOK, 200, "success"
	}
	switch {
	case errors.Is(err, service.ErrLearningMapNotFound), errors.Is(err, service.ErrLearningNodeNotFound):
		return http.StatusOK, 404, err.Error()
	case errors.Is(err, service.ErrLearningMapNotPublished):
		return http.StatusOK, 403, "learning map is not published"
	case errors.Is(err, service.ErrLearningNodeLocked):
		return http.StatusOK, 403, "node is locked"
	case errors.Is(err, service.ErrLearningNodeTypeMismatch):
		return http.StatusOK, 400, "node type mismatch"
	default:
		return http.StatusOK, 500, err.Error()
	}
}

func (api *LearningMapAPI) getUID(c *gin.Context) (string, bool) {
	uidVal, ok := c.Get("uid")
	if !ok {
		c.JSON(http.StatusOK, errorResponse(401, "用户未登录"))
		return "", false
	}
	uid, ok := uidVal.(string)
	if !ok || strings.TrimSpace(uid) == "" {
		c.JSON(http.StatusOK, errorResponse(401, "用户未登录"))
		return "", false
	}
	return uid, true
}

// ========== 用户端 ==========

func (api *LearningMapAPI) ListPublishedMaps(c *gin.Context) {
	maps, err := api.service.ListPublishedMaps()
	if err != nil {
		api.logger.Error("ListPublishedMaps failed", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "获取航海图列表失败"))
		return
	}
	c.JSON(http.StatusOK, successResponse(maps))
}

func (api *LearningMapAPI) GetMapGraph(c *gin.Context) {
	mapID, err := parseUintParam(c, "mapId")
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, err.Error()))
		return
	}
	learningMap, nodes, edges, err := api.service.GetPublishedMapGraph(mapID)
	if err != nil {
		_, code, msg := mapServiceErrorToResponse(err)
		c.JSON(http.StatusOK, errorResponse(code, msg))
		return
	}
	c.JSON(http.StatusOK, successResponse(gin.H{
		"map":   learningMap,
		"nodes": nodes,
		"edges": edges,
	}))
}

func (api *LearningMapAPI) GetMapProgress(c *gin.Context) {
	uid, ok := api.getUID(c)
	if !ok {
		return
	}
	mapID, err := parseUintParam(c, "mapId")
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, err.Error()))
		return
	}
	progress, summary, err := api.service.GetMapProgressForUser(mapID, uid)
	if err != nil {
		_, code, msg := mapServiceErrorToResponse(err)
		c.JSON(http.StatusOK, errorResponse(code, msg))
		return
	}
	c.JSON(http.StatusOK, successResponse(gin.H{
		"progress": progress,
		"summary":  summary,
	}))
}

func (api *LearningMapAPI) GetMapFull(c *gin.Context) {
	uid, ok := api.getUID(c)
	if !ok {
		return
	}
	mapID, err := parseUintParam(c, "mapId")
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, err.Error()))
		return
	}
	full, err := api.service.GetMapFullForUser(mapID, uid)
	if err != nil {
		_, code, msg := mapServiceErrorToResponse(err)
		c.JSON(http.StatusOK, errorResponse(code, msg))
		return
	}
	c.JSON(http.StatusOK, successResponse(full))
}

func (api *LearningMapAPI) StartNode(c *gin.Context) {
	uid, ok := api.getUID(c)
	if !ok {
		return
	}
	mapID, err := parseUintParam(c, "mapId")
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, err.Error()))
		return
	}
	nodeID, err := parseUintParam(c, "nodeId")
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, err.Error()))
		return
	}
	if err := api.service.MarkNodeStart(mapID, nodeID, uid); err != nil {
		_, code, msg := mapServiceErrorToResponse(err)
		c.JSON(http.StatusOK, errorResponse(code, msg))
		return
	}
	c.JSON(http.StatusOK, successResponse(gin.H{"ok": true}))
}

func (api *LearningMapAPI) CompleteNode(c *gin.Context) {
	uid, ok := api.getUID(c)
	if !ok {
		return
	}
	mapID, err := parseUintParam(c, "mapId")
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, err.Error()))
		return
	}
	nodeID, err := parseUintParam(c, "nodeId")
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, err.Error()))
		return
	}
	if err := api.service.MarkKnowledgeNodeCompleted(mapID, nodeID, uid); err != nil {
		_, code, msg := mapServiceErrorToResponse(err)
		c.JSON(http.StatusOK, errorResponse(code, msg))
		return
	}
	c.JSON(http.StatusOK, successResponse(gin.H{"ok": true}))
}

func (api *LearningMapAPI) SearchNodes(c *gin.Context) {
	uid, ok := api.getUID(c)
	if !ok {
		return
	}
	mapID, err := parseUintParam(c, "mapId")
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, err.Error()))
		return
	}
	keyword := c.Query("q")
	nodes, err := api.service.SearchNodesForUser(mapID, uid, keyword)
	if err != nil {
		_, code, msg := mapServiceErrorToResponse(err)
		c.JSON(http.StatusOK, errorResponse(code, msg))
		return
	}
	c.JSON(http.StatusOK, successResponse(nodes))
}

func (api *LearningMapAPI) RecommendNext(c *gin.Context) {
	uid, ok := api.getUID(c)
	if !ok {
		return
	}
	mapID, err := parseUintParam(c, "mapId")
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, err.Error()))
		return
	}
	next, summary, err := api.service.RecommendNextNode(mapID, uid)
	if err != nil {
		_, code, msg := mapServiceErrorToResponse(err)
		c.JSON(http.StatusOK, errorResponse(code, msg))
		return
	}
	c.JSON(http.StatusOK, successResponse(gin.H{
		"next":    next,
		"summary": summary,
	}))
}

// ========== 管理员端 ==========

func (api *LearningMapAPI) AdminListMaps(c *gin.Context) {
	maps, err := api.service.ListAdminMaps()
	if err != nil {
		api.logger.Error("AdminListMaps failed", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "获取航海图列表失败"))
		return
	}
	c.JSON(http.StatusOK, successResponse(maps))
}

func (api *LearningMapAPI) AdminCreateMap(c *gin.Context) {
	var req service.UpsertLearningMapInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}
	learningMap, err := api.service.CreateMap(req)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, err.Error()))
		return
	}
	c.JSON(http.StatusOK, successResponse(learningMap))
}

func (api *LearningMapAPI) AdminUpdateMap(c *gin.Context) {
	mapID, err := parseUintParam(c, "mapId")
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, err.Error()))
		return
	}
	var req service.UpsertLearningMapInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}
	learningMap, err := api.service.UpdateMap(mapID, req)
	if err != nil {
		_, code, msg := mapServiceErrorToResponse(err)
		c.JSON(http.StatusOK, errorResponse(code, msg))
		return
	}
	c.JSON(http.StatusOK, successResponse(learningMap))
}

func (api *LearningMapAPI) AdminDeleteMap(c *gin.Context) {
	mapID, err := parseUintParam(c, "mapId")
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, err.Error()))
		return
	}
	if err := api.service.DeleteMap(mapID); err != nil {
		_, code, msg := mapServiceErrorToResponse(err)
		c.JSON(http.StatusOK, errorResponse(code, msg))
		return
	}
	c.JSON(http.StatusOK, successResponse(gin.H{"ok": true}))
}

func (api *LearningMapAPI) AdminGetMap(c *gin.Context) {
	mapID, err := parseUintParam(c, "mapId")
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, err.Error()))
		return
	}
	learningMap, nodes, edges, err := api.service.GetMapGraphForAdmin(mapID)
	if err != nil {
		_, code, msg := mapServiceErrorToResponse(err)
		c.JSON(http.StatusOK, errorResponse(code, msg))
		return
	}
	c.JSON(http.StatusOK, successResponse(gin.H{
		"map":   learningMap,
		"nodes": nodes,
		"edges": edges,
	}))
}

func (api *LearningMapAPI) AdminCreateNode(c *gin.Context) {
	mapID, err := parseUintParam(c, "mapId")
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, err.Error()))
		return
	}
	var req service.UpsertLearningMapNodeInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}
	node, err := api.service.CreateNode(mapID, req)
	if err != nil {
		_, code, msg := mapServiceErrorToResponse(err)
		if code == 500 {
			code = 400
		}
		c.JSON(http.StatusOK, errorResponse(code, msg))
		return
	}
	c.JSON(http.StatusOK, successResponse(node))
}

func (api *LearningMapAPI) AdminUpdateNode(c *gin.Context) {
	mapID, err := parseUintParam(c, "mapId")
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, err.Error()))
		return
	}
	nodeID, err := parseUintParam(c, "nodeId")
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, err.Error()))
		return
	}
	var req service.UpsertLearningMapNodeInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}
	node, err := api.service.UpdateNode(mapID, nodeID, req)
	if err != nil {
		_, code, msg := mapServiceErrorToResponse(err)
		if code == 500 {
			code = 400
		}
		c.JSON(http.StatusOK, errorResponse(code, msg))
		return
	}
	c.JSON(http.StatusOK, successResponse(node))
}

func (api *LearningMapAPI) AdminDeleteNode(c *gin.Context) {
	mapID, err := parseUintParam(c, "mapId")
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, err.Error()))
		return
	}
	nodeID, err := parseUintParam(c, "nodeId")
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, err.Error()))
		return
	}
	if err := api.service.DeleteNode(mapID, nodeID); err != nil {
		_, code, msg := mapServiceErrorToResponse(err)
		c.JSON(http.StatusOK, errorResponse(code, msg))
		return
	}
	c.JSON(http.StatusOK, successResponse(gin.H{"ok": true}))
}

func (api *LearningMapAPI) AdminCreateEdge(c *gin.Context) {
	mapID, err := parseUintParam(c, "mapId")
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, err.Error()))
		return
	}
	var req service.UpsertLearningMapEdgeInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}
	edge, err := api.service.CreateEdge(mapID, req)
	if err != nil {
		_, code, msg := mapServiceErrorToResponse(err)
		if code == 500 {
			code = 400
		}
		c.JSON(http.StatusOK, errorResponse(code, msg))
		return
	}
	c.JSON(http.StatusOK, successResponse(edge))
}

func (api *LearningMapAPI) AdminUpdateEdge(c *gin.Context) {
	mapID, err := parseUintParam(c, "mapId")
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, err.Error()))
		return
	}
	edgeID, err := parseUintParam(c, "edgeId")
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, err.Error()))
		return
	}
	var req service.UpsertLearningMapEdgeInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}
	edge, err := api.service.UpdateEdge(mapID, edgeID, req)
	if err != nil {
		_, code, msg := mapServiceErrorToResponse(err)
		if code == 500 {
			code = 400
		}
		c.JSON(http.StatusOK, errorResponse(code, msg))
		return
	}
	c.JSON(http.StatusOK, successResponse(edge))
}

func (api *LearningMapAPI) AdminDeleteEdge(c *gin.Context) {
	mapID, err := parseUintParam(c, "mapId")
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, err.Error()))
		return
	}
	edgeID, err := parseUintParam(c, "edgeId")
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, err.Error()))
		return
	}
	if err := api.service.DeleteEdge(mapID, edgeID); err != nil {
		_, code, msg := mapServiceErrorToResponse(err)
		if code == 500 {
			code = 400
		}
		c.JSON(http.StatusOK, errorResponse(code, msg))
		return
	}
	c.JSON(http.StatusOK, successResponse(gin.H{"ok": true}))
}

func (api *LearningMapAPI) AdminPublishMap(c *gin.Context) {
	mapID, err := parseUintParam(c, "mapId")
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, err.Error()))
		return
	}
	if err := api.service.PublishMap(mapID); err != nil {
		if strings.Contains(err.Error(), "publish validation failed") {
			c.JSON(http.StatusOK, errorResponse(400, err.Error()))
			return
		}
		_, code, msg := mapServiceErrorToResponse(err)
		c.JSON(http.StatusOK, errorResponse(code, msg))
		return
	}
	c.JSON(http.StatusOK, successResponse(gin.H{"ok": true}))
}

func (api *LearningMapAPI) AdminValidateMap(c *gin.Context) {
	mapID, err := parseUintParam(c, "mapId")
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, err.Error()))
		return
	}
	errs, err := api.service.ValidateMapForPublish(mapID)
	if err != nil {
		_, code, msg := mapServiceErrorToResponse(err)
		c.JSON(http.StatusOK, errorResponse(code, msg))
		return
	}
	c.JSON(http.StatusOK, successResponse(gin.H{
		"valid":  len(errs) == 0,
		"errors": errs,
	}))
}

func (api *LearningMapAPI) AdminSearchProblems(c *gin.Context) {
	keyword := c.Query("q")
	problems, err := api.service.SearchProblemsForAdmin(keyword, 20)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(500, "搜索题目失败"))
		return
	}
	c.JSON(http.StatusOK, successResponse(problems))
}

func (api *LearningMapAPI) AdminGetProblem(c *gin.Context) {
	identifier := c.Param("problemId")
	problem, err := api.service.GetProblemForAdmin(identifier)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(404, err.Error()))
		return
	}
	c.JSON(http.StatusOK, successResponse(problem))
}

func RegisterLearningMapRoutes(router *gin.RouterGroup, svc *service.LearningMapService) {
	api := NewLearningMapAPI(svc)

	learningMaps := router.Group("/learning-maps")
	{
		learningMaps.GET("", api.ListPublishedMaps)
		learningMaps.GET("/:mapId", api.GetMapGraph)
		learningMaps.GET("/:mapId/progress", AuthMiddleware(), api.GetMapProgress)
		learningMaps.GET("/:mapId/full", AuthMiddleware(), api.GetMapFull)
		learningMaps.POST("/:mapId/nodes/:nodeId/start", AuthMiddleware(), api.StartNode)
		learningMaps.POST("/:mapId/nodes/:nodeId/complete", AuthMiddleware(), api.CompleteNode)
		learningMaps.GET("/:mapId/search", AuthMiddleware(), api.SearchNodes)
		learningMaps.GET("/:mapId/recommend-next", AuthMiddleware(), api.RecommendNext)
	}

	admin := router.Group("/admin")
	admin.Use(AdminRoleAuthMiddleware())
	{
		admin.GET("/learning-maps", api.AdminListMaps)
		admin.GET("/learning-maps/:mapId", api.AdminGetMap)
		admin.POST("/learning-maps", api.AdminCreateMap)
		admin.PUT("/learning-maps/:mapId", api.AdminUpdateMap)
		admin.DELETE("/learning-maps/:mapId", api.AdminDeleteMap)
		admin.POST("/learning-maps/:mapId/nodes", api.AdminCreateNode)
		admin.PUT("/learning-maps/:mapId/nodes/:nodeId", api.AdminUpdateNode)
		admin.DELETE("/learning-maps/:mapId/nodes/:nodeId", api.AdminDeleteNode)
		admin.POST("/learning-maps/:mapId/edges", api.AdminCreateEdge)
		admin.PUT("/learning-maps/:mapId/edges/:edgeId", api.AdminUpdateEdge)
		admin.DELETE("/learning-maps/:mapId/edges/:edgeId", api.AdminDeleteEdge)
		admin.POST("/learning-maps/:mapId/publish", api.AdminPublishMap)
		admin.GET("/learning-maps/:mapId/validate", api.AdminValidateMap)
		admin.GET("/problems/search", api.AdminSearchProblems)
		admin.GET("/problems/:problemId", api.AdminGetProblem)
	}
}
