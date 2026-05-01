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

func paramDisplayName(key string) string {
	switch key {
	case "mapId":
		return "航海图ID"
	case "nodeId":
		return "航海点ID"
	case "edgeId":
		return "连线ID"
	default:
		return key
	}
}

func parseUintParam(c *gin.Context, key string) (uint64, error) {
	val := strings.TrimSpace(c.Param(key))
	if val == "" {
		return 0, errors.New(paramDisplayName(key) + "不能为空")
	}
	parsed, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		return 0, errors.New(paramDisplayName(key) + "格式错误")
	}
	return parsed, nil
}

func localizeLearningMapMessage(msg string) string {
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return "操作失败"
	}

	replacements := map[string]string{
		"learning map not found":                                     "航海图不存在",
		"learning map is not published":                              "航海图尚未发布",
		"learning map permission denied":                             "您没有权限访问该航海图",
		"learning map node not found":                                "航海点不存在",
		"learning node is locked":                                    "当前航海点尚未解锁",
		"learning node type mismatch":                                "航海点类型不匹配",
		"learning map edge not found":                                "连线不存在",
		"title is required":                                          "标题不能为空",
		"node title is required":                                     "节点标题不能为空",
		"invalid node type":                                          "节点类型不合法",
		"invalid edge type":                                          "连线类型不合法",
		"knowledgeContent is required for knowledge node":            "知识点节点必须填写学习内容",
		"problemId or problemDisplayId is required for problem node": "题目节点必须填写主站题目ID或展示题号",
		"problem identifier is required":                             "题目标识不能为空",
		"problem not found":                                          "题目不存在",
		"problem exists but is not public":                           "题目存在但未公开",
		"sourceNodeId and targetNodeId are required":                 "起点和终点不能为空",
		"sourceNodeId or targetNodeId does not exist":                "起点或终点不存在",
		"prerequisite edge cannot reference itself":                  "前置依赖连线不能自环",
		"userId is required":                                         "用户ID不能为空",
	}
	if v, ok := replacements[msg]; ok {
		return v
	}

	if strings.HasPrefix(msg, "publish validation failed:") {
		detail := strings.TrimSpace(strings.TrimPrefix(msg, "publish validation failed:"))
		if detail == "" {
			return "发布校验失败"
		}
		return "发布校验失败：" + detail
	}

	if strings.Contains(msg, "sourceNodeId") && strings.Contains(msg, "targetNodeId") {
		return "起点或终点配置不正确"
	}

	return msg
}

func mapServiceErrorToResponse(err error) (int, int, string) {
	if err == nil {
		return http.StatusOK, 200, "success"
	}
	switch {
	case errors.Is(err, service.ErrLearningMapNotFound):
		return http.StatusOK, 404, "航海图不存在"
	case errors.Is(err, service.ErrLearningNodeNotFound):
		return http.StatusOK, 404, "航海点不存在"
	case errors.Is(err, service.ErrLearningMapNotPublished):
		return http.StatusOK, 403, "航海图尚未发布"
	case errors.Is(err, service.ErrLearningMapPermissionDeny):
		return http.StatusOK, 403, "您没有权限访问该航海图"
	case errors.Is(err, service.ErrLearningNodeLocked):
		return http.StatusOK, 403, "当前航海点尚未解锁"
	case errors.Is(err, service.ErrLearningNodeTypeMismatch):
		return http.StatusOK, 400, "航海点类型不匹配"
	default:
		return http.StatusOK, 500, localizeLearningMapMessage(err.Error())
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
	uid, ok := api.getUID(c)
	if !ok {
		return
	}
	maps, err := api.service.ListPublishedMapsForUser(uid)
	if err != nil {
		api.logger.Error("ListPublishedMaps failed", zap.Error(err))
		c.JSON(http.StatusOK, errorResponse(500, "获取航海图列表失败"))
		return
	}
	c.JSON(http.StatusOK, successResponse(maps))
}

func (api *LearningMapAPI) GetMapGraph(c *gin.Context) {
	uid, ok := api.getUID(c)
	if !ok {
		return
	}
	mapID, err := parseUintParam(c, "mapId")
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, err.Error()))
		return
	}
	learningMap, nodes, edges, err := api.service.GetPublishedMapGraphForUser(mapID, uid)
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
		c.JSON(http.StatusOK, errorResponse(400, localizeLearningMapMessage(err.Error())))
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
			c.JSON(http.StatusOK, errorResponse(400, localizeLearningMapMessage(err.Error())))
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
		c.JSON(http.StatusOK, errorResponse(404, localizeLearningMapMessage(err.Error())))
		return
	}
	c.JSON(http.StatusOK, successResponse(problem))
}

type adminMapAccessModeReq struct {
	AccessMode string `json:"accessMode"`
}

type adminMapUserPermissionReq struct {
	Enabled bool `json:"enabled"`
}

type adminMapBatchPermissionReq struct {
	UserIDs []string `json:"userIds"`
	Enabled bool     `json:"enabled"`
}

func (api *LearningMapAPI) AdminGetMapPermissions(c *gin.Context) {
	mapID, err := parseUintParam(c, "mapId")
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, err.Error()))
		return
	}
	cfg, err := api.service.GetMapAccessConfig(mapID)
	if err != nil {
		_, code, msg := mapServiceErrorToResponse(err)
		c.JSON(http.StatusOK, errorResponse(code, msg))
		return
	}
	c.JSON(http.StatusOK, successResponse(cfg))
}

func (api *LearningMapAPI) AdminSetMapAccessMode(c *gin.Context) {
	mapID, err := parseUintParam(c, "mapId")
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, err.Error()))
		return
	}
	var req adminMapAccessModeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}
	if err := api.service.SetMapAccessMode(mapID, req.AccessMode); err != nil {
		_, code, msg := mapServiceErrorToResponse(err)
		if code == 500 {
			code = 400
		}
		c.JSON(http.StatusOK, errorResponse(code, msg))
		return
	}
	c.JSON(http.StatusOK, successResponse(gin.H{"ok": true}))
}

func (api *LearningMapAPI) AdminSetMapUserPermission(c *gin.Context) {
	mapID, err := parseUintParam(c, "mapId")
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, err.Error()))
		return
	}
	userID := strings.TrimSpace(c.Param("userId"))
	if userID == "" {
		c.JSON(http.StatusOK, errorResponse(400, "用户ID不能为空"))
		return
	}
	var req adminMapUserPermissionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}
	if err := api.service.SetMapUserPermission(mapID, userID, req.Enabled); err != nil {
		_, code, msg := mapServiceErrorToResponse(err)
		if code == 500 {
			code = 400
		}
		c.JSON(http.StatusOK, errorResponse(code, msg))
		return
	}
	c.JSON(http.StatusOK, successResponse(gin.H{"ok": true}))
}

func (api *LearningMapAPI) AdminBatchSetMapUserPermissions(c *gin.Context) {
	mapID, err := parseUintParam(c, "mapId")
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, err.Error()))
		return
	}
	var req adminMapBatchPermissionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, errorResponse(400, "参数格式错误"))
		return
	}
	if err := api.service.BatchSetMapUserPermissions(mapID, req.UserIDs, req.Enabled); err != nil {
		_, code, msg := mapServiceErrorToResponse(err)
		if code == 500 {
			code = 400
		}
		c.JSON(http.StatusOK, errorResponse(code, msg))
		return
	}
	c.JSON(http.StatusOK, successResponse(gin.H{"ok": true}))
}

func (api *LearningMapAPI) AdminDeleteMapUserPermission(c *gin.Context) {
	mapID, err := parseUintParam(c, "mapId")
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(400, err.Error()))
		return
	}
	userID := strings.TrimSpace(c.Param("userId"))
	if userID == "" {
		c.JSON(http.StatusOK, errorResponse(400, "用户ID不能为空"))
		return
	}
	if err := api.service.DeleteMapUserPermission(mapID, userID); err != nil {
		_, code, msg := mapServiceErrorToResponse(err)
		if code == 500 {
			code = 400
		}
		c.JSON(http.StatusOK, errorResponse(code, msg))
		return
	}
	c.JSON(http.StatusOK, successResponse(gin.H{"ok": true}))
}

func (api *LearningMapAPI) AdminSearchMapPermissionUsers(c *gin.Context) {
	keyword := c.Query("q")
	users, err := api.service.SearchUsersForMapPermission(keyword, 20)
	if err != nil {
		c.JSON(http.StatusOK, errorResponse(500, "搜索用户失败"))
		return
	}
	c.JSON(http.StatusOK, successResponse(users))
}

func RegisterLearningMapRoutes(router *gin.RouterGroup, svc *service.LearningMapService) {
	api := NewLearningMapAPI(svc)

	learningMaps := router.Group("/learning-maps")
	{
		learningMaps.GET("", AuthMiddleware(), api.ListPublishedMaps)
		learningMaps.GET("/:mapId", AuthMiddleware(), api.GetMapGraph)
		learningMaps.GET("/:mapId/progress", AuthMiddleware(), api.GetMapProgress)
		learningMaps.GET("/:mapId/full", AuthMiddleware(), api.GetMapFull)
		learningMaps.POST("/:mapId/nodes/:nodeId/start", AuthMiddleware(), api.StartNode)
		learningMaps.POST("/:mapId/nodes/:nodeId/complete", AuthMiddleware(), api.CompleteNode)
		learningMaps.GET("/:mapId/search", AuthMiddleware(), api.SearchNodes)
		learningMaps.GET("/:mapId/recommend-next", AuthMiddleware(), api.RecommendNext)
	}

	admin := router.Group("/admin")
	admin.Use(AdminOrProblemAdminAuthMiddleware())
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
		admin.GET("/learning-maps/:mapId/permissions", api.AdminGetMapPermissions)
		admin.PUT("/learning-maps/:mapId/permissions/mode", api.AdminSetMapAccessMode)
		admin.PUT("/learning-maps/:mapId/permissions/:userId", api.AdminSetMapUserPermission)
		admin.POST("/learning-maps/:mapId/permissions/batch", api.AdminBatchSetMapUserPermissions)
		admin.DELETE("/learning-maps/:mapId/permissions/:userId", api.AdminDeleteMapUserPermission)
		admin.GET("/learning-maps/users/search", api.AdminSearchMapPermissionUsers)
		admin.GET("/problems/search", api.AdminSearchProblems)
		admin.GET("/problems/:problemId", api.AdminGetProblem)
	}
}
