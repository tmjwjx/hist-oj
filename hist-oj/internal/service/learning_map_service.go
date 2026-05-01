package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/hoj/hist-oj/internal/model"
	"github.com/hoj/hist-oj/internal/utils"
)

var (
	ErrLearningMapNotFound       = errors.New("learning map not found")
	ErrLearningMapNotPublished   = errors.New("learning map is not published")
	ErrLearningMapPermissionDeny = errors.New("learning map permission denied")
	ErrLearningNodeNotFound      = errors.New("learning map node not found")
	ErrLearningNodeLocked        = errors.New("learning node is locked")
	ErrLearningNodeTypeMismatch  = errors.New("learning node type mismatch")
)

// LearningMapService 航海图服务
// 负责管理员 CRUD、发布校验、用户进度计算与状态更新。
type LearningMapService struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewLearningMapService(db *gorm.DB) *LearningMapService {
	return &LearningMapService{db: db, logger: utils.GetLogger()}
}

// LearningProblemInfo 题目基础信息（用于 problem 类型节点展示/校验）
type LearningProblemInfo struct {
	ID               uint64   `json:"id"`
	ProblemDisplayID string   `json:"problemDisplayId"`
	Title            string   `json:"title"`
	Difficulty       int      `json:"difficulty"`
	Tags             []string `json:"tags"`
}

// LearningMapNodeView 前端友好的节点结构
type LearningMapNodeView struct {
	model.LearningMapNode
	TagsList    []string             `json:"tagsList"`
	ProblemInfo *LearningProblemInfo `json:"problemInfo,omitempty"`
}

// LearningNodeProgressSnapshot 用户在节点上的快照状态
type LearningNodeProgressSnapshot struct {
	NodeID                 uint64     `json:"nodeId"`
	Status                 string     `json:"status"`
	CompletedAt            *time.Time `json:"completedAt,omitempty"`
	MasteredAt             *time.Time `json:"masteredAt,omitempty"`
	MissingPrerequisiteIDs []uint64   `json:"missingPrerequisiteIds,omitempty"`
}

// LearningMapSummary 航海图摘要
type LearningMapSummary struct {
	Total          int `json:"total"`
	Locked         int `json:"locked"`
	Available      int `json:"available"`
	InProgress     int `json:"inProgress"`
	Completed      int `json:"completed"`
	Mastered       int `json:"mastered"`
	CompletionRate int `json:"completionRate"`
}

// LearningMapFullResponse 用户端一次性渲染数据
type LearningMapFullResponse struct {
	Map             model.LearningMap              `json:"map"`
	Nodes           []LearningMapNodeView          `json:"nodes"`
	Edges           []model.LearningMapEdge        `json:"edges"`
	Progress        []LearningNodeProgressSnapshot `json:"progress"`
	Summary         LearningMapSummary             `json:"summary"`
	NextRecommended *LearningMapNodeView           `json:"nextRecommended,omitempty"`
}

// UpsertLearningMapInput 创建/更新航海图
// status: draft | published
type UpsertLearningMapInput struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	AccessMode  string `json:"accessMode"`
}

// UpsertLearningMapNodeInput 创建/更新节点
type UpsertLearningMapNodeInput struct {
	Type             string                 `json:"type"`
	Title            string                 `json:"title"`
	Description      string                 `json:"description"`
	Difficulty       string                 `json:"difficulty"`
	Tags             []string               `json:"tags"`
	X                float64                `json:"x"`
	Y                float64                `json:"y"`
	Level            int                    `json:"level"`
	Region           string                 `json:"region"`
	Published        *bool                  `json:"published"`
	KnowledgeContent string                 `json:"knowledgeContent"`
	ProblemID        *uint64                `json:"problemId"`
	ProblemDisplayID string                 `json:"problemDisplayId"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// UpsertLearningMapEdgeInput 创建/更新连线
type UpsertLearningMapEdgeInput struct {
	SourceNodeID uint64 `json:"sourceNodeId"`
	TargetNodeID uint64 `json:"targetNodeId"`
	Type         string `json:"type"`
}

// PublishValidationError 发布校验详情
type PublishValidationError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// LearningMapPermissionView 航海图权限覆盖项
type LearningMapPermissionView struct {
	UserID    string    `json:"userId"`
	Username  string    `json:"username"`
	Nickname  string    `json:"nickname"`
	Realname  string    `json:"realname"`
	Enabled   bool      `json:"enabled"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// LearningMapAccessConfig 航海图访问配置
type LearningMapAccessConfig struct {
	AccessMode  string                      `json:"accessMode"`
	Permissions []LearningMapPermissionView `json:"permissions"`
}

// LearningMapPermissionUser 用户搜索结果（用于权限设置）
type LearningMapPermissionUser struct {
	UserID   string `json:"userId"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Realname string `json:"realname"`
}

func normalizeLearningMapStatus(status string) string {
	status = strings.TrimSpace(strings.ToLower(status))
	if status == "" {
		return model.LearningMapStatusDraft
	}
	if status != model.LearningMapStatusDraft && status != model.LearningMapStatusPublished {
		return model.LearningMapStatusDraft
	}
	return status
}

func normalizeLearningMapAccessMode(mode string) string {
	mode = strings.TrimSpace(strings.ToLower(mode))
	if mode == model.LearningMapAccessAllClosed {
		return model.LearningMapAccessAllClosed
	}
	return model.LearningMapAccessAllOpen
}

func normalizeLearningNodeType(nodeType string) string {
	nodeType = strings.TrimSpace(strings.ToLower(nodeType))
	if nodeType != model.LearningMapNodeTypeKnowledge && nodeType != model.LearningMapNodeTypeProblem {
		return ""
	}
	return nodeType
}

func normalizeEdgeType(edgeType string) string {
	edgeType = strings.TrimSpace(strings.ToLower(edgeType))
	if edgeType == "" {
		return model.LearningMapEdgeTypePrerequisite
	}
	if edgeType != model.LearningMapEdgeTypePrerequisite && edgeType != model.LearningMapEdgeTypeRelated {
		return ""
	}
	return edgeType
}

func normalizeDifficulty(diff string) string {
	diff = strings.TrimSpace(strings.ToLower(diff))
	switch diff {
	case "beginner", "easy", "medium", "hard", "expert":
		return diff
	default:
		return "beginner"
	}
}

func encodeTags(tags []string) string {
	if len(tags) == 0 {
		return "[]"
	}
	cleaned := make([]string, 0, len(tags))
	for _, t := range tags {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		cleaned = append(cleaned, t)
	}
	if len(cleaned) == 0 {
		return "[]"
	}
	data, err := json.Marshal(cleaned)
	if err != nil {
		return "[]"
	}
	return string(data)
}

func decodeTags(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []string{}
	}
	var tags []string
	if err := json.Unmarshal([]byte(raw), &tags); err == nil {
		return tags
	}
	// 兼容历史 CSV
	parts := strings.Split(raw, ",")
	res := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			res = append(res, p)
		}
	}
	return res
}

func encodeMetadata(metadata map[string]interface{}) string {
	if len(metadata) == 0 {
		return "{}"
	}
	data, err := json.Marshal(metadata)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func (s *LearningMapService) ListAdminMaps() ([]model.LearningMap, error) {
	var maps []model.LearningMap
	err := s.db.Order("id DESC").Find(&maps).Error
	return maps, err
}

func (s *LearningMapService) ListPublishedMaps() ([]model.LearningMap, error) {
	var maps []model.LearningMap
	err := s.db.Where("status = ?", model.LearningMapStatusPublished).Order("id DESC").Find(&maps).Error
	return maps, err
}

func (s *LearningMapService) ListPublishedMapsForUser(uid string) ([]model.LearningMap, error) {
	uid = strings.TrimSpace(uid)
	if uid == "" {
		return []model.LearningMap{}, nil
	}
	maps, err := s.ListPublishedMaps()
	if err != nil {
		return nil, err
	}
	if len(maps) == 0 {
		return maps, nil
	}
	mapIDs := make([]uint64, 0, len(maps))
	for _, m := range maps {
		mapIDs = append(mapIDs, m.ID)
	}
	overrideMap, err := s.getPermissionOverrides(mapIDs, uid)
	if err != nil {
		return nil, err
	}
	filtered := make([]model.LearningMap, 0, len(maps))
	for _, m := range maps {
		override, ok := overrideMap[m.ID]
		if isMapAccessAllowed(m.AccessMode, ok, override) {
			filtered = append(filtered, m)
		}
	}
	return filtered, nil
}

func (s *LearningMapService) GetMapByID(mapID uint64) (*model.LearningMap, error) {
	var learningMap model.LearningMap
	if err := s.db.Where("id = ?", mapID).First(&learningMap).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrLearningMapNotFound
		}
		return nil, err
	}
	return &learningMap, nil
}

func (s *LearningMapService) CreateMap(input UpsertLearningMapInput) (*model.LearningMap, error) {
	learningMap := model.LearningMap{
		Title:       strings.TrimSpace(input.Title),
		Description: strings.TrimSpace(input.Description),
		Status:      normalizeLearningMapStatus(input.Status),
		AccessMode:  normalizeLearningMapAccessMode(input.AccessMode),
	}
	if learningMap.Title == "" {
		return nil, errors.New("title is required")
	}
	if err := s.db.Create(&learningMap).Error; err != nil {
		return nil, err
	}
	return &learningMap, nil
}

func (s *LearningMapService) UpdateMap(mapID uint64, input UpsertLearningMapInput) (*model.LearningMap, error) {
	learningMap, err := s.GetMapByID(mapID)
	if err != nil {
		return nil, err
	}
	accessMode := learningMap.AccessMode
	if strings.TrimSpace(input.AccessMode) != "" {
		accessMode = normalizeLearningMapAccessMode(input.AccessMode)
	}
	updates := map[string]interface{}{
		"title":       strings.TrimSpace(input.Title),
		"description": strings.TrimSpace(input.Description),
		"status":      normalizeLearningMapStatus(input.Status),
		"access_mode": accessMode,
	}
	if updates["title"] == "" {
		return nil, errors.New("title is required")
	}
	if err := s.db.Model(&model.LearningMap{}).Where("id = ?", mapID).Updates(updates).Error; err != nil {
		return nil, err
	}
	learningMap.Title = updates["title"].(string)
	learningMap.Description = updates["description"].(string)
	learningMap.Status = updates["status"].(string)
	learningMap.AccessMode = updates["access_mode"].(string)
	return learningMap, nil
}

func (s *LearningMapService) DeleteMap(mapID uint64) error {
	if _, err := s.GetMapByID(mapID); err != nil {
		return err
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("map_id = ?", mapID).Delete(&model.UserLearningProgress{}).Error; err != nil {
			return err
		}
		if err := tx.Where("map_id = ?", mapID).Delete(&model.LearningMapPermission{}).Error; err != nil {
			return err
		}
		if err := tx.Where("map_id = ?", mapID).Delete(&model.LearningMapEdge{}).Error; err != nil {
			return err
		}
		if err := tx.Where("map_id = ?", mapID).Delete(&model.LearningMapNode{}).Error; err != nil {
			return err
		}
		if err := tx.Where("id = ?", mapID).Delete(&model.LearningMap{}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (s *LearningMapService) mapNodesAndEdges(mapID uint64, onlyPublished bool) ([]model.LearningMapNode, []model.LearningMapEdge, error) {
	var nodes []model.LearningMapNode
	nodeQuery := s.db.Where("map_id = ?", mapID)
	if onlyPublished {
		nodeQuery = nodeQuery.Where("published = 1")
	}
	if err := nodeQuery.Order("id ASC").Find(&nodes).Error; err != nil {
		return nil, nil, err
	}
	var edges []model.LearningMapEdge
	if err := s.db.Where("map_id = ?", mapID).Order("id ASC").Find(&edges).Error; err != nil {
		return nil, nil, err
	}
	if onlyPublished {
		nodeIDs := make(map[uint64]struct{}, len(nodes))
		for _, n := range nodes {
			nodeIDs[n.ID] = struct{}{}
		}
		filtered := make([]model.LearningMapEdge, 0, len(edges))
		for _, e := range edges {
			if _, ok := nodeIDs[e.SourceNodeID]; !ok {
				continue
			}
			if _, ok := nodeIDs[e.TargetNodeID]; !ok {
				continue
			}
			filtered = append(filtered, e)
		}
		edges = filtered
	}
	return nodes, edges, nil
}

func (s *LearningMapService) buildNodeViews(nodes []model.LearningMapNode) ([]LearningMapNodeView, error) {
	problemInfoMap, err := s.getProblemInfoForNodes(nodes)
	if err != nil {
		return nil, err
	}
	views := make([]LearningMapNodeView, 0, len(nodes))
	for _, node := range nodes {
		view := LearningMapNodeView{
			LearningMapNode: node,
			TagsList:        decodeTags(node.Tags),
		}
		if node.Type == model.LearningMapNodeTypeProblem {
			if node.ProblemID != nil {
				if info, ok := problemInfoMap[*node.ProblemID]; ok {
					copyInfo := info
					view.ProblemInfo = &copyInfo
				}
			}
		}
		views = append(views, view)
	}
	return views, nil
}

func (s *LearningMapService) GetMapGraphForAdmin(mapID uint64) (*model.LearningMap, []LearningMapNodeView, []model.LearningMapEdge, error) {
	learningMap, err := s.GetMapByID(mapID)
	if err != nil {
		return nil, nil, nil, err
	}
	nodes, edges, err := s.mapNodesAndEdges(mapID, false)
	if err != nil {
		return nil, nil, nil, err
	}
	nodeViews, err := s.buildNodeViews(nodes)
	if err != nil {
		return nil, nil, nil, err
	}
	return learningMap, nodeViews, edges, nil
}

func (s *LearningMapService) GetPublishedMapGraph(mapID uint64) (*model.LearningMap, []LearningMapNodeView, []model.LearningMapEdge, error) {
	learningMap, err := s.GetMapByID(mapID)
	if err != nil {
		return nil, nil, nil, err
	}
	if learningMap.Status != model.LearningMapStatusPublished {
		return nil, nil, nil, ErrLearningMapNotPublished
	}
	nodes, edges, err := s.mapNodesAndEdges(mapID, true)
	if err != nil {
		return nil, nil, nil, err
	}
	nodeViews, err := s.buildNodeViews(nodes)
	if err != nil {
		return nil, nil, nil, err
	}
	return learningMap, nodeViews, edges, nil
}

func isMapAccessAllowed(accessMode string, hasOverride bool, overrideEnabled bool) bool {
	if hasOverride {
		return overrideEnabled
	}
	return normalizeLearningMapAccessMode(accessMode) == model.LearningMapAccessAllOpen
}

func (s *LearningMapService) getPermissionOverrides(mapIDs []uint64, uid string) (map[uint64]bool, error) {
	res := make(map[uint64]bool)
	if len(mapIDs) == 0 || strings.TrimSpace(uid) == "" {
		return res, nil
	}
	var rows []model.LearningMapPermission
	if err := s.db.Where("map_id IN ? AND user_id = ?", mapIDs, uid).Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		res[row.MapID] = row.Enabled
	}
	return res, nil
}

func (s *LearningMapService) ensureMapAccess(mapID uint64, uid string) (*model.LearningMap, error) {
	learningMap, err := s.GetMapByID(mapID)
	if err != nil {
		return nil, err
	}
	if learningMap.Status != model.LearningMapStatusPublished {
		return nil, ErrLearningMapNotPublished
	}
	overrideMap, err := s.getPermissionOverrides([]uint64{mapID}, uid)
	if err != nil {
		return nil, err
	}
	override, hasOverride := overrideMap[mapID]
	if !isMapAccessAllowed(learningMap.AccessMode, hasOverride, override) {
		return nil, ErrLearningMapPermissionDeny
	}
	return learningMap, nil
}

func (s *LearningMapService) GetPublishedMapGraphForUser(mapID uint64, uid string) (*model.LearningMap, []LearningMapNodeView, []model.LearningMapEdge, error) {
	if _, err := s.ensureMapAccess(mapID, uid); err != nil {
		return nil, nil, nil, err
	}
	return s.GetPublishedMapGraph(mapID)
}

func (s *LearningMapService) GetMapAccessConfig(mapID uint64) (*LearningMapAccessConfig, error) {
	learningMap, err := s.GetMapByID(mapID)
	if err != nil {
		return nil, err
	}
	type row struct {
		UserID    string    `gorm:"column:user_id"`
		Username  string    `gorm:"column:username"`
		Nickname  string    `gorm:"column:nickname"`
		Realname  string    `gorm:"column:realname"`
		Enabled   bool      `gorm:"column:enabled"`
		UpdatedAt time.Time `gorm:"column:update_time"`
	}
	var rows []row
	if err := s.db.Table("learning_map_permission p").
		Select("p.user_id, p.enabled, p.update_time, ui.username, ui.nickname, ui.realname").
		Joins("LEFT JOIN user_info ui ON ui.uuid = p.user_id").
		Where("p.map_id = ?", mapID).
		Order("p.update_time DESC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	views := make([]LearningMapPermissionView, 0, len(rows))
	for _, r := range rows {
		views = append(views, LearningMapPermissionView{
			UserID:    r.UserID,
			Username:  r.Username,
			Nickname:  r.Nickname,
			Realname:  r.Realname,
			Enabled:   r.Enabled,
			UpdatedAt: r.UpdatedAt,
		})
	}
	return &LearningMapAccessConfig{
		AccessMode:  normalizeLearningMapAccessMode(learningMap.AccessMode),
		Permissions: views,
	}, nil
}

func (s *LearningMapService) SetMapAccessMode(mapID uint64, accessMode string) error {
	if _, err := s.GetMapByID(mapID); err != nil {
		return err
	}
	mode := normalizeLearningMapAccessMode(accessMode)
	return s.db.Model(&model.LearningMap{}).Where("id = ?", mapID).Update("access_mode", mode).Error
}

func (s *LearningMapService) SetMapUserPermission(mapID uint64, userID string, enabled bool) error {
	if _, err := s.GetMapByID(mapID); err != nil {
		return err
	}
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return errors.New("userId is required")
	}
	now := time.Now()
	row := model.LearningMapPermission{
		MapID:     mapID,
		UserID:    userID,
		Enabled:   enabled,
		UpdatedAt: now,
	}
	return s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "map_id"}, {Name: "user_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{"enabled": enabled, "update_time": now}),
	}).Create(&row).Error
}

func (s *LearningMapService) BatchSetMapUserPermissions(mapID uint64, userIDs []string, enabled bool) error {
	if _, err := s.GetMapByID(mapID); err != nil {
		return err
	}
	uniq := make(map[string]struct{}, len(userIDs))
	rows := make([]model.LearningMapPermission, 0, len(userIDs))
	now := time.Now()
	for _, uid := range userIDs {
		uid = strings.TrimSpace(uid)
		if uid == "" {
			continue
		}
		if _, ok := uniq[uid]; ok {
			continue
		}
		uniq[uid] = struct{}{}
		rows = append(rows, model.LearningMapPermission{
			MapID:     mapID,
			UserID:    uid,
			Enabled:   enabled,
			UpdatedAt: now,
		})
	}
	if len(rows) == 0 {
		return nil
	}
	return s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "map_id"}, {Name: "user_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{"enabled": enabled, "update_time": now}),
	}).Create(&rows).Error
}

func (s *LearningMapService) DeleteMapUserPermission(mapID uint64, userID string) error {
	if _, err := s.GetMapByID(mapID); err != nil {
		return err
	}
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return errors.New("userId is required")
	}
	return s.db.Where("map_id = ? AND user_id = ?", mapID, userID).Delete(&model.LearningMapPermission{}).Error
}

func (s *LearningMapService) SearchUsersForMapPermission(keyword string, limit int) ([]LearningMapPermissionUser, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return []LearningMapPermissionUser{}, nil
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	type row struct {
		UserID   string `gorm:"column:uuid"`
		Username string `gorm:"column:username"`
		Nickname string `gorm:"column:nickname"`
		Realname string `gorm:"column:realname"`
	}
	var rows []row
	like := "%" + keyword + "%"
	if err := s.db.Table("user_info").
		Select("uuid, username, nickname, realname").
		Where("username LIKE ? OR nickname LIKE ? OR realname LIKE ?", like, like, like).
		Order("uuid ASC").
		Limit(limit).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	res := make([]LearningMapPermissionUser, 0, len(rows))
	for _, r := range rows {
		res = append(res, LearningMapPermissionUser{
			UserID:   r.UserID,
			Username: r.Username,
			Nickname: r.Nickname,
			Realname: r.Realname,
		})
	}
	return res, nil
}

func (s *LearningMapService) CreateNode(mapID uint64, input UpsertLearningMapNodeInput) (*model.LearningMapNode, error) {
	if _, err := s.GetMapByID(mapID); err != nil {
		return nil, err
	}
	node, err := s.newNodeModel(mapID, input)
	if err != nil {
		return nil, err
	}
	if err := s.db.Create(node).Error; err != nil {
		return nil, err
	}
	return node, nil
}

func (s *LearningMapService) UpdateNode(mapID, nodeID uint64, input UpsertLearningMapNodeInput) (*model.LearningMapNode, error) {
	if _, err := s.GetMapByID(mapID); err != nil {
		return nil, err
	}
	var node model.LearningMapNode
	if err := s.db.Where("map_id = ? AND id = ?", mapID, nodeID).First(&node).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrLearningNodeNotFound
		}
		return nil, err
	}
	updated, err := s.newNodeModel(mapID, input)
	if err != nil {
		return nil, err
	}
	updates := map[string]interface{}{
		"type":               updated.Type,
		"title":              updated.Title,
		"description":        updated.Description,
		"difficulty":         updated.Difficulty,
		"tags":               updated.Tags,
		"x":                  updated.X,
		"y":                  updated.Y,
		"level":              updated.Level,
		"region":             updated.Region,
		"published":          updated.Published,
		"knowledge_content":  updated.KnowledgeContent,
		"problem_id":         updated.ProblemID,
		"problem_display_id": updated.ProblemDisplayID,
		"metadata":           updated.Metadata,
	}
	if err := s.db.Model(&model.LearningMapNode{}).
		Where("map_id = ? AND id = ?", mapID, nodeID).
		Updates(updates).Error; err != nil {
		return nil, err
	}
	if err := s.db.Where("map_id = ? AND id = ?", mapID, nodeID).First(&node).Error; err != nil {
		return nil, err
	}
	return &node, nil
}

func (s *LearningMapService) DeleteNode(mapID, nodeID uint64) error {
	if _, err := s.GetMapByID(mapID); err != nil {
		return err
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		var node model.LearningMapNode
		if err := tx.Where("map_id = ? AND id = ?", mapID, nodeID).First(&node).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrLearningNodeNotFound
			}
			return err
		}
		if err := tx.Where("map_id = ? AND (source_node_id = ? OR target_node_id = ?)", mapID, nodeID, nodeID).
			Delete(&model.LearningMapEdge{}).Error; err != nil {
			return err
		}
		if err := tx.Where("map_id = ? AND node_id = ?", mapID, nodeID).
			Delete(&model.UserLearningProgress{}).Error; err != nil {
			return err
		}
		if err := tx.Where("map_id = ? AND id = ?", mapID, nodeID).
			Delete(&model.LearningMapNode{}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (s *LearningMapService) CreateEdge(mapID uint64, input UpsertLearningMapEdgeInput) (*model.LearningMapEdge, error) {
	if _, err := s.GetMapByID(mapID); err != nil {
		return nil, err
	}
	edge, err := s.newEdgeModel(mapID, input)
	if err != nil {
		return nil, err
	}
	if err := s.validateEdgeNodesExist(mapID, edge.SourceNodeID, edge.TargetNodeID); err != nil {
		return nil, err
	}
	if err := s.ensurePrerequisiteEdgeNoCycle(mapID, 0, *edge); err != nil {
		return nil, err
	}
	if err := s.db.Create(edge).Error; err != nil {
		return nil, err
	}
	return edge, nil
}

func (s *LearningMapService) UpdateEdge(mapID, edgeID uint64, input UpsertLearningMapEdgeInput) (*model.LearningMapEdge, error) {
	if _, err := s.GetMapByID(mapID); err != nil {
		return nil, err
	}
	var edge model.LearningMapEdge
	if err := s.db.Where("map_id = ? AND id = ?", mapID, edgeID).First(&edge).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("learning map edge not found")
		}
		return nil, err
	}
	updated, err := s.newEdgeModel(mapID, input)
	if err != nil {
		return nil, err
	}
	if err := s.validateEdgeNodesExist(mapID, updated.SourceNodeID, updated.TargetNodeID); err != nil {
		return nil, err
	}
	if err := s.ensurePrerequisiteEdgeNoCycle(mapID, edgeID, *updated); err != nil {
		return nil, err
	}
	updates := map[string]interface{}{
		"source_node_id": updated.SourceNodeID,
		"target_node_id": updated.TargetNodeID,
		"type":           updated.Type,
	}
	if err := s.db.Model(&model.LearningMapEdge{}).
		Where("map_id = ? AND id = ?", mapID, edgeID).
		Updates(updates).Error; err != nil {
		return nil, err
	}
	if err := s.db.Where("map_id = ? AND id = ?", mapID, edgeID).First(&edge).Error; err != nil {
		return nil, err
	}
	return &edge, nil
}

func (s *LearningMapService) DeleteEdge(mapID, edgeID uint64) error {
	if _, err := s.GetMapByID(mapID); err != nil {
		return err
	}
	result := s.db.Where("map_id = ? AND id = ?", mapID, edgeID).Delete(&model.LearningMapEdge{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("learning map edge not found")
	}
	return nil
}

func (s *LearningMapService) PublishMap(mapID uint64) error {
	if _, err := s.GetMapByID(mapID); err != nil {
		return err
	}
	validationErrors, err := s.ValidateMapForPublish(mapID)
	if err != nil {
		return err
	}
	if len(validationErrors) > 0 {
		msgParts := make([]string, 0, len(validationErrors))
		for _, ve := range validationErrors {
			msgParts = append(msgParts, ve.Message)
		}
		return fmt.Errorf("publish validation failed: %s", strings.Join(msgParts, "; "))
	}
	return s.db.Model(&model.LearningMap{}).
		Where("id = ?", mapID).
		Update("status", model.LearningMapStatusPublished).Error
}

func (s *LearningMapService) ValidateMapForPublish(mapID uint64) ([]PublishValidationError, error) {
	nodes, edges, err := s.mapNodesAndEdges(mapID, false)
	if err != nil {
		return nil, err
	}
	errs := make([]PublishValidationError, 0)
	if len(nodes) == 0 {
		errs = append(errs, PublishValidationError{Code: "EMPTY_MAP", Message: "至少需要一个航海点"})
		return errs, nil
	}
	nodeMap := make(map[uint64]model.LearningMapNode, len(nodes))
	for _, n := range nodes {
		nodeMap[n.ID] = n
		nType := normalizeLearningNodeType(n.Type)
		if nType == "" {
			errs = append(errs, PublishValidationError{Code: "INVALID_NODE_TYPE", Message: fmt.Sprintf("节点 %d 类型无效", n.ID)})
			continue
		}
		if strings.TrimSpace(n.Title) == "" {
			errs = append(errs, PublishValidationError{Code: "EMPTY_TITLE", Message: fmt.Sprintf("节点 %d 标题不能为空", n.ID)})
		}
		if nType == model.LearningMapNodeTypeKnowledge {
			if strings.TrimSpace(n.KnowledgeContent) == "" {
				errs = append(errs, PublishValidationError{Code: "EMPTY_KNOWLEDGE_CONTENT", Message: fmt.Sprintf("知识点节点 %d 需要填写学习内容", n.ID)})
			}
		}
		if nType == model.LearningMapNodeTypeProblem {
			if n.ProblemID == nil && strings.TrimSpace(n.ProblemDisplayID) == "" {
				errs = append(errs, PublishValidationError{Code: "EMPTY_PROBLEM_BINDING", Message: fmt.Sprintf("题目节点 %d 未绑定主 OJ 题目", n.ID)})
				continue
			}
			if _, err := s.resolveProblemForNode(&n); err != nil {
				errs = append(errs, PublishValidationError{Code: "PROBLEM_NOT_FOUND", Message: fmt.Sprintf("题目节点 %d 绑定题目无效: %v", n.ID, err)})
			}
		}
	}

	for _, e := range edges {
		if _, ok := nodeMap[e.SourceNodeID]; !ok {
			errs = append(errs, PublishValidationError{Code: "INVALID_EDGE", Message: fmt.Sprintf("连线 %d 的 sourceNodeId 不存在", e.ID)})
		}
		if _, ok := nodeMap[e.TargetNodeID]; !ok {
			errs = append(errs, PublishValidationError{Code: "INVALID_EDGE", Message: fmt.Sprintf("连线 %d 的 targetNodeId 不存在", e.ID)})
		}
		typeNormalized := normalizeEdgeType(e.Type)
		if typeNormalized == "" {
			errs = append(errs, PublishValidationError{Code: "INVALID_EDGE_TYPE", Message: fmt.Sprintf("连线 %d 的类型无效", e.ID)})
		}
		if typeNormalized == model.LearningMapEdgeTypePrerequisite && e.SourceNodeID == e.TargetNodeID {
			errs = append(errs, PublishValidationError{Code: "INVALID_EDGE", Message: fmt.Sprintf("前置连线 %d 不能自环", e.ID)})
		}
	}

	if len(errs) > 0 {
		return errs, nil
	}

	if hasCycle, cycleNodes := DetectPrerequisiteCycle(nodes, edges); hasCycle {
		errs = append(errs, PublishValidationError{
			Code:    "CYCLE_DETECTED",
			Message: fmt.Sprintf("检测到前置依赖环，请修正：%v", cycleNodes),
		})
	}
	return errs, nil
}

// GetMapFullForUser 一次性返回图谱、进度、推荐
func (s *LearningMapService) GetMapFullForUser(mapID uint64, uid string) (*LearningMapFullResponse, error) {
	learningMap, nodeViews, edges, err := s.GetPublishedMapGraphForUser(mapID, uid)
	if err != nil {
		return nil, err
	}
	nodes := make([]model.LearningMapNode, 0, len(nodeViews))
	for _, nv := range nodeViews {
		nodes = append(nodes, nv.LearningMapNode)
	}
	progressMap, err := s.recomputeAndSyncUserProgress(mapID, uid, nodes, edges)
	if err != nil {
		return nil, err
	}
	progress := make([]LearningNodeProgressSnapshot, 0, len(progressMap))
	for _, p := range progressMap {
		progress = append(progress, p)
	}
	sort.Slice(progress, func(i, j int) bool { return progress[i].NodeID < progress[j].NodeID })
	summary := BuildLearningMapSummary(progress)
	next := pickNextRecommendedNode(nodeViews, progressMap)
	return &LearningMapFullResponse{
		Map:             *learningMap,
		Nodes:           nodeViews,
		Edges:           edges,
		Progress:        progress,
		Summary:         summary,
		NextRecommended: next,
	}, nil
}

func pickNextRecommendedNode(nodes []LearningMapNodeView, progressMap map[uint64]LearningNodeProgressSnapshot) *LearningMapNodeView {
	availableNodes := make([]LearningMapNodeView, 0)
	for _, n := range nodes {
		if p, ok := progressMap[n.ID]; ok && (p.Status == model.LearningMapProgressAvailable || p.Status == model.LearningMapProgressInProgress) {
			availableNodes = append(availableNodes, n)
		}
	}
	if len(availableNodes) == 0 {
		return nil
	}
	sort.Slice(availableNodes, func(i, j int) bool {
		if availableNodes[i].Level != availableNodes[j].Level {
			return availableNodes[i].Level < availableNodes[j].Level
		}
		if availableNodes[i].Type != availableNodes[j].Type {
			return availableNodes[i].Type == model.LearningMapNodeTypeKnowledge
		}
		return availableNodes[i].ID < availableNodes[j].ID
	})
	res := availableNodes[0]
	return &res
}

func BuildLearningMapSummary(progress []LearningNodeProgressSnapshot) LearningMapSummary {
	summary := LearningMapSummary{Total: len(progress)}
	for _, p := range progress {
		switch p.Status {
		case model.LearningMapProgressLocked:
			summary.Locked++
		case model.LearningMapProgressAvailable:
			summary.Available++
		case model.LearningMapProgressInProgress:
			summary.InProgress++
		case model.LearningMapProgressCompleted:
			summary.Completed++
		case model.LearningMapProgressMastered:
			summary.Mastered++
		}
	}
	if summary.Total > 0 {
		done := summary.Completed + summary.Mastered
		summary.CompletionRate = int(float64(done) * 100 / float64(summary.Total))
	}
	return summary
}

func (s *LearningMapService) GetMapProgressForUser(mapID uint64, uid string) ([]LearningNodeProgressSnapshot, LearningMapSummary, error) {
	_, nodeViews, edges, err := s.GetPublishedMapGraphForUser(mapID, uid)
	if err != nil {
		return nil, LearningMapSummary{}, err
	}
	nodes := make([]model.LearningMapNode, 0, len(nodeViews))
	for _, n := range nodeViews {
		nodes = append(nodes, n.LearningMapNode)
	}
	progressMap, err := s.recomputeAndSyncUserProgress(mapID, uid, nodes, edges)
	if err != nil {
		return nil, LearningMapSummary{}, err
	}
	progress := make([]LearningNodeProgressSnapshot, 0, len(progressMap))
	for _, p := range progressMap {
		progress = append(progress, p)
	}
	sort.Slice(progress, func(i, j int) bool { return progress[i].NodeID < progress[j].NodeID })
	return progress, BuildLearningMapSummary(progress), nil
}

func (s *LearningMapService) MarkNodeStart(mapID, nodeID uint64, uid string) error {
	_, nodeViews, edges, err := s.GetPublishedMapGraphForUser(mapID, uid)
	if err != nil {
		return err
	}
	nodeMap := make(map[uint64]model.LearningMapNode, len(nodeViews))
	nodes := make([]model.LearningMapNode, 0, len(nodeViews))
	for _, nv := range nodeViews {
		nodeMap[nv.ID] = nv.LearningMapNode
		nodes = append(nodes, nv.LearningMapNode)
	}
	node, ok := nodeMap[nodeID]
	if !ok {
		return ErrLearningNodeNotFound
	}
	if node.Type != model.LearningMapNodeTypeKnowledge {
		return ErrLearningNodeTypeMismatch
	}
	progressMap, err := s.recomputeAndSyncUserProgress(mapID, uid, nodes, edges)
	if err != nil {
		return err
	}
	if err := ValidateKnowledgeActionAllowed(node.Type, progressMap[nodeID].Status); err != nil {
		return err
	}
	now := time.Now()
	row := model.UserLearningProgress{
		UserID:    uid,
		MapID:     mapID,
		NodeID:    nodeID,
		Status:    model.LearningMapProgressInProgress,
		UpdatedAt: now,
	}
	return s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "map_id"}, {Name: "node_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{"status": model.LearningMapProgressInProgress, "update_time": now}),
	}).Create(&row).Error
}

func (s *LearningMapService) MarkKnowledgeNodeCompleted(mapID, nodeID uint64, uid string) error {
	_, nodeViews, edges, err := s.GetPublishedMapGraphForUser(mapID, uid)
	if err != nil {
		return err
	}
	nodeMap := make(map[uint64]model.LearningMapNode, len(nodeViews))
	nodes := make([]model.LearningMapNode, 0, len(nodeViews))
	for _, nv := range nodeViews {
		nodeMap[nv.ID] = nv.LearningMapNode
		nodes = append(nodes, nv.LearningMapNode)
	}
	node, ok := nodeMap[nodeID]
	if !ok {
		return ErrLearningNodeNotFound
	}
	if node.Type != model.LearningMapNodeTypeKnowledge {
		return ErrLearningNodeTypeMismatch
	}
	progressMap, err := s.recomputeAndSyncUserProgress(mapID, uid, nodes, edges)
	if err != nil {
		return err
	}
	if err := ValidateKnowledgeActionAllowed(node.Type, progressMap[nodeID].Status); err != nil {
		return err
	}
	now := time.Now()
	row := model.UserLearningProgress{
		UserID:      uid,
		MapID:       mapID,
		NodeID:      nodeID,
		Status:      model.LearningMapProgressCompleted,
		CompletedAt: &now,
		UpdatedAt:   now,
	}
	if err := s.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "map_id"}, {Name: "node_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"status":       model.LearningMapProgressCompleted,
			"completed_at": now,
			"update_time":  now,
		}),
	}).Create(&row).Error; err != nil {
		return err
	}
	// 完成后立即重算，确保后继节点可解锁
	_, err = s.recomputeAndSyncUserProgress(mapID, uid, nodes, edges)
	return err
}

// ValidateKnowledgeActionAllowed 校验知识点操作（开始学习/标记完成）是否允许。
func ValidateKnowledgeActionAllowed(nodeType, status string) error {
	if nodeType != model.LearningMapNodeTypeKnowledge {
		return ErrLearningNodeTypeMismatch
	}
	if status == model.LearningMapProgressLocked {
		return ErrLearningNodeLocked
	}
	return nil
}

func (s *LearningMapService) SearchNodesForUser(mapID uint64, uid, keyword string) ([]LearningMapNodeView, error) {
	keyword = strings.TrimSpace(keyword)
	_, nodeViews, edges, err := s.GetPublishedMapGraphForUser(mapID, uid)
	if err != nil {
		return nil, err
	}
	nodes := make([]model.LearningMapNode, 0, len(nodeViews))
	for _, n := range nodeViews {
		nodes = append(nodes, n.LearningMapNode)
	}
	progressMap, err := s.recomputeAndSyncUserProgress(mapID, uid, nodes, edges)
	if err != nil {
		return nil, err
	}
	filtered := make([]LearningMapNodeView, 0)
	for _, n := range nodeViews {
		if keyword != "" {
			if !strings.Contains(strings.ToLower(n.Title), strings.ToLower(keyword)) &&
				!strings.Contains(strings.ToLower(n.Description), strings.ToLower(keyword)) {
				if n.ProblemInfo == nil || !strings.Contains(strings.ToLower(n.ProblemInfo.ProblemDisplayID), strings.ToLower(keyword)) {
					continue
				}
			}
		}
		if p, ok := progressMap[n.ID]; ok {
			n.Metadata = mergeStatusIntoMetadata(n.Metadata, p)
		}
		filtered = append(filtered, n)
	}
	return filtered, nil
}

func mergeStatusIntoMetadata(raw string, progress LearningNodeProgressSnapshot) string {
	meta := map[string]interface{}{}
	if strings.TrimSpace(raw) != "" {
		_ = json.Unmarshal([]byte(raw), &meta)
	}
	meta["status"] = progress.Status
	if len(progress.MissingPrerequisiteIDs) > 0 {
		meta["missingPrerequisiteIds"] = progress.MissingPrerequisiteIDs
	}
	data, _ := json.Marshal(meta)
	return string(data)
}

func (s *LearningMapService) RecommendNextNode(mapID uint64, uid string) (*LearningMapNodeView, LearningMapSummary, error) {
	_, nodeViews, edges, err := s.GetPublishedMapGraphForUser(mapID, uid)
	if err != nil {
		return nil, LearningMapSummary{}, err
	}
	nodes := make([]model.LearningMapNode, 0, len(nodeViews))
	for _, n := range nodeViews {
		nodes = append(nodes, n.LearningMapNode)
	}
	progressMap, err := s.recomputeAndSyncUserProgress(mapID, uid, nodes, edges)
	if err != nil {
		return nil, LearningMapSummary{}, err
	}
	progressList := make([]LearningNodeProgressSnapshot, 0, len(progressMap))
	for _, p := range progressMap {
		progressList = append(progressList, p)
	}
	summary := BuildLearningMapSummary(progressList)
	return pickNextRecommendedNode(nodeViews, progressMap), summary, nil
}

func (s *LearningMapService) SearchProblemsForAdmin(keyword string, limit int) ([]LearningProblemInfo, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	keyword = strings.TrimSpace(keyword)
	query := s.db.Table("problem").Select("id, problem_id, title, difficulty").Where("auth = 1")
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("problem_id LIKE ? OR title LIKE ?", like, like)
	}
	type row struct {
		ID         uint64 `gorm:"column:id"`
		ProblemID  string `gorm:"column:problem_id"`
		Title      string `gorm:"column:title"`
		Difficulty int    `gorm:"column:difficulty"`
	}
	var rows []row
	if err := query.Order("id DESC").Limit(limit).Find(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return []LearningProblemInfo{}, nil
	}
	ids := make([]uint64, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.ID)
	}
	tagMap, err := s.batchProblemTags(ids)
	if err != nil {
		return nil, err
	}
	res := make([]LearningProblemInfo, 0, len(rows))
	for _, r := range rows {
		res = append(res, LearningProblemInfo{
			ID:               r.ID,
			ProblemDisplayID: r.ProblemID,
			Title:            r.Title,
			Difficulty:       r.Difficulty,
			Tags:             tagMap[r.ID],
		})
	}
	return res, nil
}

func (s *LearningMapService) GetProblemForAdmin(identifier string) (*LearningProblemInfo, error) {
	identifier = strings.TrimSpace(identifier)
	if identifier == "" {
		return nil, errors.New("problem identifier is required")
	}
	var row struct {
		ID         uint64 `gorm:"column:id"`
		ProblemID  string `gorm:"column:problem_id"`
		Title      string `gorm:"column:title"`
		Difficulty int    `gorm:"column:difficulty"`
		Auth       int    `gorm:"column:auth"`
	}
	query := s.db.Table("problem").Select("id, problem_id, title, difficulty, auth")
	if id, err := strconv.ParseUint(identifier, 10, 64); err == nil {
		err = query.Where("id = ?", id).First(&row).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if err == nil {
			if row.Auth != 1 {
				return nil, errors.New("problem exists but is not public")
			}
			tagMap, err := s.batchProblemTags([]uint64{row.ID})
			if err != nil {
				return nil, err
			}
			return &LearningProblemInfo{ID: row.ID, ProblemDisplayID: row.ProblemID, Title: row.Title, Difficulty: row.Difficulty, Tags: tagMap[row.ID]}, nil
		}
	}
	if err := s.db.Table("problem").Select("id, problem_id, title, difficulty, auth").Where("problem_id = ?", identifier).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("problem not found")
		}
		return nil, err
	}
	if row.Auth != 1 {
		return nil, errors.New("problem exists but is not public")
	}
	tagMap, err := s.batchProblemTags([]uint64{row.ID})
	if err != nil {
		return nil, err
	}
	return &LearningProblemInfo{ID: row.ID, ProblemDisplayID: row.ProblemID, Title: row.Title, Difficulty: row.Difficulty, Tags: tagMap[row.ID]}, nil
}

func (s *LearningMapService) newNodeModel(mapID uint64, input UpsertLearningMapNodeInput) (*model.LearningMapNode, error) {
	nType := normalizeLearningNodeType(input.Type)
	if nType == "" {
		return nil, errors.New("invalid node type")
	}
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return nil, errors.New("node title is required")
	}
	published := true
	if input.Published != nil {
		published = *input.Published
	}
	node := &model.LearningMapNode{
		MapID:            mapID,
		Type:             nType,
		Title:            title,
		Description:      strings.TrimSpace(input.Description),
		Difficulty:       normalizeDifficulty(input.Difficulty),
		Tags:             encodeTags(input.Tags),
		X:                input.X,
		Y:                input.Y,
		Level:            input.Level,
		Region:           strings.TrimSpace(input.Region),
		Published:        published,
		KnowledgeContent: strings.TrimSpace(input.KnowledgeContent),
		ProblemID:        input.ProblemID,
		ProblemDisplayID: strings.TrimSpace(input.ProblemDisplayID),
		Metadata:         encodeMetadata(input.Metadata),
	}
	if nType == model.LearningMapNodeTypeKnowledge {
		if node.KnowledgeContent == "" {
			return nil, errors.New("knowledgeContent is required for knowledge node")
		}
		node.ProblemID = nil
		node.ProblemDisplayID = ""
	}
	if nType == model.LearningMapNodeTypeProblem {
		if node.ProblemID == nil && node.ProblemDisplayID == "" {
			return nil, errors.New("problemId or problemDisplayId is required for problem node")
		}
		problemInfo, err := s.resolveProblemForNode(node)
		if err != nil {
			return nil, err
		}
		if node.Title == "" {
			node.Title = problemInfo.Title
		}
		node.KnowledgeContent = ""
	}
	return node, nil
}

func (s *LearningMapService) newEdgeModel(mapID uint64, input UpsertLearningMapEdgeInput) (*model.LearningMapEdge, error) {
	etype := normalizeEdgeType(input.Type)
	if etype == "" {
		return nil, errors.New("invalid edge type")
	}
	if input.SourceNodeID == 0 || input.TargetNodeID == 0 {
		return nil, errors.New("sourceNodeId and targetNodeId are required")
	}
	if etype == model.LearningMapEdgeTypePrerequisite && input.SourceNodeID == input.TargetNodeID {
		return nil, errors.New("prerequisite edge cannot reference itself")
	}
	return &model.LearningMapEdge{
		MapID:        mapID,
		SourceNodeID: input.SourceNodeID,
		TargetNodeID: input.TargetNodeID,
		Type:         etype,
	}, nil
}

func (s *LearningMapService) validateEdgeNodesExist(mapID, sourceNodeID, targetNodeID uint64) error {
	var cnt int64
	if err := s.db.Table("learning_map_node").Where("map_id = ? AND id IN ?", mapID, []uint64{sourceNodeID, targetNodeID}).Count(&cnt).Error; err != nil {
		return err
	}
	if cnt != 2 {
		return errors.New("sourceNodeId or targetNodeId does not exist")
	}
	return nil
}

func (s *LearningMapService) ensurePrerequisiteEdgeNoCycle(mapID, ignoreEdgeID uint64, edge model.LearningMapEdge) error {
	if normalizeEdgeType(edge.Type) != model.LearningMapEdgeTypePrerequisite {
		return nil
	}
	var edges []model.LearningMapEdge
	if err := s.db.Where("map_id = ? AND type = ?", mapID, model.LearningMapEdgeTypePrerequisite).Find(&edges).Error; err != nil {
		return err
	}
	hasCycle, cyclePath := CheckPrerequisiteEdgeCreatesCycle(edges, ignoreEdgeID, edge.SourceNodeID, edge.TargetNodeID)
	if !hasCycle {
		return nil
	}
	return fmt.Errorf("前置依赖会形成环：%s", formatNodePath(cyclePath))
}

func formatNodePath(path []uint64) string {
	if len(path) == 0 {
		return ""
	}
	parts := make([]string, 0, len(path))
	for _, id := range path {
		parts = append(parts, strconv.FormatUint(id, 10))
	}
	return strings.Join(parts, " -> ")
}

// CheckPrerequisiteEdgeCreatesCycle 判断新增/更新一条 prerequisite 边后是否会形成环。
// ignoreEdgeID 用于更新场景：先逻辑删除该边，再判断候选边是否导致成环。
func CheckPrerequisiteEdgeCreatesCycle(
	edges []model.LearningMapEdge,
	ignoreEdgeID uint64,
	sourceNodeID uint64,
	targetNodeID uint64,
) (bool, []uint64) {
	if sourceNodeID == 0 || targetNodeID == 0 {
		return false, nil
	}
	adj := make(map[uint64][]uint64)
	for _, e := range edges {
		if ignoreEdgeID > 0 && e.ID == ignoreEdgeID {
			continue
		}
		if normalizeEdgeType(e.Type) != model.LearningMapEdgeTypePrerequisite {
			continue
		}
		adj[e.SourceNodeID] = append(adj[e.SourceNodeID], e.TargetNodeID)
	}

	// 若已存在 target -> ... -> source，新增 source -> target 将闭合成环。
	queue := []uint64{targetNodeID}
	visited := map[uint64]bool{targetNodeID: true}
	parent := make(map[uint64]uint64)
	found := false

	for len(queue) > 0 && !found {
		cur := queue[0]
		queue = queue[1:]
		for _, nxt := range adj[cur] {
			if visited[nxt] {
				continue
			}
			visited[nxt] = true
			parent[nxt] = cur
			if nxt == sourceNodeID {
				found = true
				break
			}
			queue = append(queue, nxt)
		}
	}
	if !found {
		return false, nil
	}

	// 还原路径：target -> ... -> source；最终返回环路：source -> target -> ... -> source。
	reversed := []uint64{sourceNodeID}
	for cur := sourceNodeID; cur != targetNodeID; {
		p, ok := parent[cur]
		if !ok {
			break
		}
		reversed = append(reversed, p)
		cur = p
	}
	pathTargetToSource := make([]uint64, len(reversed))
	for i := range reversed {
		pathTargetToSource[i] = reversed[len(reversed)-1-i]
	}
	cyclePath := make([]uint64, 0, len(pathTargetToSource)+1)
	cyclePath = append(cyclePath, sourceNodeID)
	cyclePath = append(cyclePath, pathTargetToSource...)
	return true, cyclePath
}

func (s *LearningMapService) resolveProblemForNode(node *model.LearningMapNode) (*LearningProblemInfo, error) {
	if node == nil {
		return nil, errors.New("node is nil")
	}
	if node.ProblemID != nil {
		problem, err := s.GetProblemForAdmin(strconv.FormatUint(*node.ProblemID, 10))
		if err != nil {
			return nil, err
		}
		node.ProblemDisplayID = problem.ProblemDisplayID
		pid := problem.ID
		node.ProblemID = &pid
		if strings.TrimSpace(node.Title) == "" {
			node.Title = problem.Title
		}
		return problem, nil
	}
	if strings.TrimSpace(node.ProblemDisplayID) != "" {
		problem, err := s.GetProblemForAdmin(node.ProblemDisplayID)
		if err != nil {
			return nil, err
		}
		node.ProblemDisplayID = problem.ProblemDisplayID
		pid := problem.ID
		node.ProblemID = &pid
		if strings.TrimSpace(node.Title) == "" {
			node.Title = problem.Title
		}
		return problem, nil
	}
	return nil, errors.New("problem binding is empty")
}

func (s *LearningMapService) batchProblemTags(problemIDs []uint64) (map[uint64][]string, error) {
	if len(problemIDs) == 0 {
		return map[uint64][]string{}, nil
	}
	type row struct {
		PID  uint64 `gorm:"column:pid"`
		Name string `gorm:"column:name"`
	}
	var rows []row
	err := s.db.Table("problem_tag pt").
		Select("pt.pid, t.name").
		Joins("LEFT JOIN tag t ON t.id = pt.tid").
		Where("pt.pid IN ?", problemIDs).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	res := make(map[uint64][]string)
	for _, r := range rows {
		name := strings.TrimSpace(r.Name)
		if name == "" {
			continue
		}
		res[r.PID] = append(res[r.PID], name)
	}
	for k := range res {
		sort.Strings(res[k])
	}
	return res, nil
}

func (s *LearningMapService) getProblemInfoForNodes(nodes []model.LearningMapNode) (map[uint64]LearningProblemInfo, error) {
	idsSet := make(map[uint64]struct{})
	displaySet := make(map[string]struct{})
	for i := range nodes {
		n := &nodes[i]
		if n.Type != model.LearningMapNodeTypeProblem {
			continue
		}
		if n.ProblemID != nil {
			idsSet[*n.ProblemID] = struct{}{}
		} else if strings.TrimSpace(n.ProblemDisplayID) != "" {
			displaySet[n.ProblemDisplayID] = struct{}{}
		}
	}
	ids := make([]uint64, 0, len(idsSet))
	for id := range idsSet {
		ids = append(ids, id)
	}
	displays := make([]string, 0, len(displaySet))
	for d := range displaySet {
		displays = append(displays, d)
	}
	type row struct {
		ID         uint64 `gorm:"column:id"`
		ProblemID  string `gorm:"column:problem_id"`
		Title      string `gorm:"column:title"`
		Difficulty int    `gorm:"column:difficulty"`
	}
	query := s.db.Table("problem").Select("id, problem_id, title, difficulty").Where("auth = 1")
	if len(ids) > 0 && len(displays) > 0 {
		query = query.Where("id IN ? OR problem_id IN ?", ids, displays)
	} else if len(ids) > 0 {
		query = query.Where("id IN ?", ids)
	} else if len(displays) > 0 {
		query = query.Where("problem_id IN ?", displays)
	} else {
		return map[uint64]LearningProblemInfo{}, nil
	}
	var rows []row
	if err := query.Find(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return map[uint64]LearningProblemInfo{}, nil
	}
	problemIDs := make([]uint64, 0, len(rows))
	for _, r := range rows {
		problemIDs = append(problemIDs, r.ID)
	}
	tagMap, err := s.batchProblemTags(problemIDs)
	if err != nil {
		return nil, err
	}
	res := make(map[uint64]LearningProblemInfo, len(rows))
	displayToID := make(map[string]uint64, len(rows))
	for _, r := range rows {
		info := LearningProblemInfo{
			ID:               r.ID,
			ProblemDisplayID: r.ProblemID,
			Title:            r.Title,
			Difficulty:       r.Difficulty,
			Tags:             tagMap[r.ID],
		}
		res[r.ID] = info
		displayToID[r.ProblemID] = r.ID
	}
	// 回填节点上仅有 displayId 的 ProblemID，便于后续进度计算。
	for i := range nodes {
		n := &nodes[i]
		if n.Type != model.LearningMapNodeTypeProblem {
			continue
		}
		if n.ProblemID == nil && n.ProblemDisplayID != "" {
			if pid, ok := displayToID[n.ProblemDisplayID]; ok {
				n.ProblemID = &pid
			}
		}
	}
	return res, nil
}

// recomputeAndSyncUserProgress 根据前置关系、用户手动进度、题目 AC 记录统一计算状态并持久化。
func (s *LearningMapService) recomputeAndSyncUserProgress(mapID uint64, uid string, nodes []model.LearningMapNode, edges []model.LearningMapEdge) (map[uint64]LearningNodeProgressSnapshot, error) {
	if len(nodes) == 0 {
		return map[uint64]LearningNodeProgressSnapshot{}, nil
	}
	problemIDs := make([]uint64, 0)
	for _, n := range nodes {
		if n.Type == model.LearningMapNodeTypeProblem && n.ProblemID != nil {
			problemIDs = append(problemIDs, *n.ProblemID)
		}
	}
	solvedSet, err := s.getSolvedProblemSet(uid, problemIDs)
	if err != nil {
		return nil, err
	}
	var rows []model.UserLearningProgress
	if err := s.db.Where("user_id = ? AND map_id = ?", uid, mapID).Find(&rows).Error; err != nil {
		return nil, err
	}
	progressRows := make(map[uint64]model.UserLearningProgress, len(rows))
	for _, r := range rows {
		progressRows[r.NodeID] = r
	}
	snapshot := ComputeLearningNodeProgress(nodes, edges, progressRows, solvedSet)
	if err := s.upsertProgressSnapshot(mapID, uid, snapshot, progressRows); err != nil {
		return nil, err
	}
	return snapshot, nil
}

func (s *LearningMapService) getSolvedProblemSet(uid string, problemIDs []uint64) (map[uint64]bool, error) {
	res := make(map[uint64]bool)
	if len(problemIDs) == 0 {
		return res, nil
	}
	type row struct {
		PID uint64 `gorm:"column:pid"`
	}
	var rows []row
	if err := s.db.Table("user_acproblem").
		Select("DISTINCT pid").
		Where("uid = ? AND pid IN ?", uid, problemIDs).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, r := range rows {
		res[r.PID] = true
	}
	return res, nil
}

func (s *LearningMapService) upsertProgressSnapshot(mapID uint64, uid string, snapshot map[uint64]LearningNodeProgressSnapshot, existing map[uint64]model.UserLearningProgress) error {
	now := time.Now()
	rows := make([]model.UserLearningProgress, 0, len(snapshot))
	for nodeID, p := range snapshot {
		row := model.UserLearningProgress{
			UserID:    uid,
			MapID:     mapID,
			NodeID:    nodeID,
			Status:    p.Status,
			UpdatedAt: now,
		}
		if p.Status == model.LearningMapProgressCompleted || p.Status == model.LearningMapProgressMastered {
			if ex, ok := existing[nodeID]; ok && ex.CompletedAt != nil {
				row.CompletedAt = ex.CompletedAt
			} else {
				row.CompletedAt = &now
			}
		}
		if p.Status == model.LearningMapProgressMastered {
			if ex, ok := existing[nodeID]; ok && ex.MasteredAt != nil {
				row.MasteredAt = ex.MasteredAt
			} else {
				row.MasteredAt = &now
			}
		}
		rows = append(rows, row)
	}
	if len(rows) == 0 {
		return nil
	}
	return s.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "map_id"}, {Name: "node_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"status":       clause.Expr{SQL: "VALUES(status)"},
			"completed_at": clause.Expr{SQL: "VALUES(completed_at)"},
			"mastered_at":  clause.Expr{SQL: "VALUES(mastered_at)"},
			"update_time":  now,
		}),
	}).Create(&rows).Error
}

// ComputeLearningNodeProgress 纯函数：核心解锁逻辑
// 规则：
// 1. problem 节点是否 completed 只看 user_acproblem。
// 2. 只有所有 prerequisite 前置都 completed/mastered，节点才 available。
// 3. locked 节点无法进入 in_progress/completed。
func ComputeLearningNodeProgress(
	nodes []model.LearningMapNode,
	edges []model.LearningMapEdge,
	existing map[uint64]model.UserLearningProgress,
	solvedProblemSet map[uint64]bool,
) map[uint64]LearningNodeProgressSnapshot {
	incoming := buildIncomingPrerequisites(edges)

	baseCompleted := make(map[uint64]bool, len(nodes))
	baseMastered := make(map[uint64]bool, len(nodes))
	baseInProgress := make(map[uint64]bool, len(nodes))
	completedAt := make(map[uint64]*time.Time, len(nodes))
	masteredAt := make(map[uint64]*time.Time, len(nodes))

	for _, node := range nodes {
		row, hasRow := existing[node.ID]
		switch node.Type {
		case model.LearningMapNodeTypeProblem:
			// problem 节点必须以 AC 记录为准。
			isSolved := false
			if node.ProblemID != nil {
				isSolved = solvedProblemSet[*node.ProblemID]
			}
			if isSolved {
				baseCompleted[node.ID] = true
				if hasRow && row.CompletedAt != nil {
					completedAt[node.ID] = row.CompletedAt
				}
				if hasRow && row.Status == model.LearningMapProgressMastered {
					baseMastered[node.ID] = true
					if row.MasteredAt != nil {
						masteredAt[node.ID] = row.MasteredAt
					}
				}
			}
		default:
			if !hasRow {
				continue
			}
			switch row.Status {
			case model.LearningMapProgressMastered:
				baseMastered[node.ID] = true
				baseCompleted[node.ID] = true
				if row.MasteredAt != nil {
					masteredAt[node.ID] = row.MasteredAt
				}
				if row.CompletedAt != nil {
					completedAt[node.ID] = row.CompletedAt
				}
			case model.LearningMapProgressCompleted:
				baseCompleted[node.ID] = true
				if row.CompletedAt != nil {
					completedAt[node.ID] = row.CompletedAt
				}
			case model.LearningMapProgressInProgress:
				baseInProgress[node.ID] = true
			}
		}
	}

	// 只有在“自身基础完成 + 所有 prerequisite 也有效完成”时，才视为有效完成。
	effectiveCompletedMemo := make(map[uint64]bool, len(nodes))
	effectiveCompletedDone := make(map[uint64]bool, len(nodes))
	var resolveEffectiveCompleted func(nodeID uint64, visiting map[uint64]bool) bool
	resolveEffectiveCompleted = func(nodeID uint64, visiting map[uint64]bool) bool {
		if done, ok := effectiveCompletedDone[nodeID]; ok && done {
			return effectiveCompletedMemo[nodeID]
		}
		if !baseCompleted[nodeID] {
			effectiveCompletedDone[nodeID] = true
			effectiveCompletedMemo[nodeID] = false
			return false
		}
		if visiting[nodeID] {
			// 保护性分支：若历史脏数据存在环，认为不可有效完成，避免错误传递解锁。
			effectiveCompletedDone[nodeID] = true
			effectiveCompletedMemo[nodeID] = false
			return false
		}
		visiting[nodeID] = true
		for _, pre := range incoming[nodeID] {
			if !resolveEffectiveCompleted(pre, visiting) {
				delete(visiting, nodeID)
				effectiveCompletedDone[nodeID] = true
				effectiveCompletedMemo[nodeID] = false
				return false
			}
		}
		delete(visiting, nodeID)
		effectiveCompletedDone[nodeID] = true
		effectiveCompletedMemo[nodeID] = true
		return true
	}
	effectiveCompleted := make(map[uint64]bool, len(nodes))
	for _, node := range nodes {
		effectiveCompleted[node.ID] = resolveEffectiveCompleted(node.ID, map[uint64]bool{})
	}

	result := make(map[uint64]LearningNodeProgressSnapshot, len(nodes))
	for _, node := range nodes {
		missing := make([]uint64, 0)
		for _, pre := range incoming[node.ID] {
			if !effectiveCompleted[pre] {
				missing = append(missing, pre)
			}
		}
		snapshot := LearningNodeProgressSnapshot{NodeID: node.ID}
		if len(missing) > 0 {
			snapshot.Status = model.LearningMapProgressLocked
			snapshot.MissingPrerequisiteIDs = missing
		} else {
			if baseMastered[node.ID] && effectiveCompleted[node.ID] {
				snapshot.Status = model.LearningMapProgressMastered
				snapshot.MasteredAt = masteredAt[node.ID]
				snapshot.CompletedAt = completedAt[node.ID]
			} else if effectiveCompleted[node.ID] {
				snapshot.Status = model.LearningMapProgressCompleted
				snapshot.CompletedAt = completedAt[node.ID]
			} else if baseInProgress[node.ID] {
				snapshot.Status = model.LearningMapProgressInProgress
			} else {
				snapshot.Status = model.LearningMapProgressAvailable
			}
		}
		result[node.ID] = snapshot
	}
	return result
}

func buildIncomingPrerequisites(edges []model.LearningMapEdge) map[uint64][]uint64 {
	incoming := make(map[uint64][]uint64)
	for _, e := range edges {
		if normalizeEdgeType(e.Type) != model.LearningMapEdgeTypePrerequisite {
			continue
		}
		incoming[e.TargetNodeID] = append(incoming[e.TargetNodeID], e.SourceNodeID)
	}
	return incoming
}

// DetectPrerequisiteCycle 检测 prerequisite 子图中的环。
// 返回: 是否有环 + 仍未被拓扑删除的节点 ID。
func DetectPrerequisiteCycle(nodes []model.LearningMapNode, edges []model.LearningMapEdge) (bool, []uint64) {
	if len(nodes) == 0 {
		return false, nil
	}
	nodeIDs := make(map[uint64]struct{}, len(nodes))
	for _, n := range nodes {
		nodeIDs[n.ID] = struct{}{}
	}
	inDegree := make(map[uint64]int, len(nodes))
	adj := make(map[uint64][]uint64, len(nodes))
	for _, n := range nodes {
		inDegree[n.ID] = 0
	}
	for _, e := range edges {
		if normalizeEdgeType(e.Type) != model.LearningMapEdgeTypePrerequisite {
			continue
		}
		if _, ok := nodeIDs[e.SourceNodeID]; !ok {
			continue
		}
		if _, ok := nodeIDs[e.TargetNodeID]; !ok {
			continue
		}
		adj[e.SourceNodeID] = append(adj[e.SourceNodeID], e.TargetNodeID)
		inDegree[e.TargetNodeID]++
	}
	queue := make([]uint64, 0)
	for id, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, id)
		}
	}
	visited := 0
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		visited++
		for _, nxt := range adj[cur] {
			inDegree[nxt]--
			if inDegree[nxt] == 0 {
				queue = append(queue, nxt)
			}
		}
	}
	if visited == len(nodes) {
		return false, nil
	}
	cycleNodes := make([]uint64, 0)
	for id, deg := range inDegree {
		if deg > 0 {
			cycleNodes = append(cycleNodes, id)
		}
	}
	sort.Slice(cycleNodes, func(i, j int) bool { return cycleNodes[i] < cycleNodes[j] })
	return true, cycleNodes
}

// CleanupEdgesAfterNodeDelete 返回删除某个节点后剩余的边（纯函数，供测试与校验使用）。
func CleanupEdgesAfterNodeDelete(edges []model.LearningMapEdge, nodeID uint64) []model.LearningMapEdge {
	filtered := make([]model.LearningMapEdge, 0, len(edges))
	for _, e := range edges {
		if e.SourceNodeID == nodeID || e.TargetNodeID == nodeID {
			continue
		}
		filtered = append(filtered, e)
	}
	return filtered
}
