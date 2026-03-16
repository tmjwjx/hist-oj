package websocket

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"github.com/hoj/hist-oj/internal/client"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// 允许所有来源，生产环境应该限制
		return true
	},
}

// HandleWebSocket 处理WebSocket连接请求
func (h *Hub) HandleWebSocket(c *gin.Context) {
	logger := h.logger

	// 获取参数
	competitionID := c.Param("competitionId")
	if competitionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少比赛ID"})
		return
	}

	// 解析比赛ID
	compID, err := strconv.ParseUint(competitionID, 10, 64)
	if err != nil {
		logger.Error("解析比赛ID失败", zap.Error(err), zap.String("competition_id", competitionID))
		c.JSON(http.StatusBadRequest, gin.H{"error": "比赛ID格式错误"})
		return
	}

	// 验证token（从URL参数或Header中获取）
	token := c.Query("token")
	if token == "" {
		token = c.GetHeader("Authorization")
	}

	if token == "" {
		logger.Warn("WebSocket连接缺少token")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	// 验证token并检查管理员权限
	userAuth, err := client.ValidateToken(token)
	if err != nil {
		logger.Warn("WebSocket token验证失败", zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "认证失败"})
		return
	}

	// 检查是否为管理员（root或admin角色）
	isAdmin := false
	for _, role := range userAuth.Roles {
		if role == "root" || role == "admin" {
			isAdmin = true
			break
		}
	}

	// 允许普通用户连接（只读模式，只接收更新）
	logger.Info("WebSocket连接请求",
		zap.String("username", userAuth.Username),
		zap.Bool("is_admin", isAdmin),
		zap.Strings("roles", userAuth.Roles))

	// 升级HTTP连接为WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("WebSocket升级失败", zap.Error(err))
		return
	}

	// 创建新客户端
	wsClient := &Client{
		ID:           uuid.New().String(),
		CompetitionID: compID,
		Conn:         conn,
		Send:         make(chan Message, 256),
		Logger:       logger,
	}

	// 注册客户端
	h.register <- wsClient

	// 启动读写循环
	go wsClient.WritePump()
	go wsClient.ReadPump(h)

	logger.Info("WebSocket连接已建立",
		zap.String("client_id", wsClient.ID),
		zap.Uint64("competition_id", compID),
		zap.String("username", userAuth.Username))
}

// BroadcastRegistrationUpdate 广播报名更新消息
func (h *Hub) BroadcastRegistrationUpdate(competitionID uint64, registration interface{}) {
	message := Message{
		Type:         "registration_update",
		CompetitionID: competitionID,
		Data:         registration,
		Timestamp:    time.Now().Unix(),
	}

	select {
	case h.broadcast <- message:
	default:
		h.logger.Warn("广播通道已满，丢弃消息")
	}
}

// BroadcastNewMessage 广播新消息通知
func (h *Hub) BroadcastNewMessage(competitionID uint64, registrationID uint64, messageCount int) {
	data := map[string]interface{}{
		"registration_id": registrationID,
		"message_count":   messageCount,
	}

	message := Message{
		Type:         "new_message",
		CompetitionID: competitionID,
		Data:         data,
		Timestamp:    time.Now().Unix(),
	}

	select {
	case h.broadcast <- message:
	default:
		h.logger.Warn("广播通道已满，丢弃消息")
	}
}

// BroadcastCompetitionUpdate 广播比赛信息更新
func (h *Hub) BroadcastCompetitionUpdate(competitionID uint64, competition interface{}) {
	message := Message{
		Type:         "competition_update",
		CompetitionID: competitionID,
		Data:         competition,
		Timestamp:    time.Now().Unix(),
	}

	select {
	case h.broadcast <- message:
	default:
		h.logger.Warn("广播通道已满，丢弃消息")
	}
}

// GetStats 获取WebSocket统计信息
func (h *Hub) GetStats() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	stats := map[string]interface{}{
		"total_clients":   len(h.clients),
		"competitions":    len(h.competitionClients),
		"buffer_capacity": cap(h.broadcast),
		"buffer_length":   len(h.broadcast),
	}

	// 每个比赛的客户端数
	competitionStats := make(map[uint64]int)
	for compID, clients := range h.competitionClients {
		competitionStats[compID] = len(clients)
	}
	stats["competition_clients"] = competitionStats

	return stats
}
