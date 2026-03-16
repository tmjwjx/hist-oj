package websocket

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// Message WebSocket消息格式
type Message struct {
	Type         string      `json:"type"`         // 消息类型: "registration_update", "new_message", etc.
	CompetitionID uint64     `json:"competition_id"` // 目标比赛ID
	Data         interface{} `json:"data"`         // 消息数据
	Timestamp    int64       `json:"timestamp"`    // 时间戳
}

// Client WebSocket客户端
type Client struct {
	ID           string
	CompetitionID uint64  // 客户端关注的比赛ID
	Conn         *websocket.Conn
	Send         chan Message
	Logger       *zap.Logger
}

// Hub WebSocket连接管理中心
type Hub struct {
	// 注册的客户端，key为客户端ID
	clients map[string]*Client

	// 按比赛ID分组的客户端，用于广播
	competitionClients map[uint64]map[string]*Client

	// 注册和注销请求
	register   chan *Client
	unregister chan *Client

	// 广播消息（发送给特定比赛的所有客户端）
	broadcast chan Message

	// 互斥锁
	mu sync.RWMutex

	logger *zap.Logger
}

// NewHub 创建新的Hub
func NewHub(logger *zap.Logger) *Hub {
	return &Hub{
		clients:            make(map[string]*Client),
		competitionClients: make(map[uint64]map[string]*Client),
		register:           make(chan *Client),
		unregister:         make(chan *Client),
		broadcast:          make(chan Message, 256),
		logger:             logger,
	}
}

// Run 启动Hub的主循环
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
}

// registerClient 注册新客户端
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[client.ID] = client

	// 按比赛ID分组
	if h.competitionClients[client.CompetitionID] == nil {
		h.competitionClients[client.CompetitionID] = make(map[string]*Client)
	}
	h.competitionClients[client.CompetitionID][client.ID] = client

	h.logger.Info("WebSocket客户端已注册",
		zap.String("client_id", client.ID),
		zap.Uint64("competition_id", client.CompetitionID),
		zap.Int("total_clients", len(h.clients)))
}

// unregisterClient 注销客户端
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client.ID]; ok {
		delete(h.clients, client.ID)

		// 从比赛分组中移除
		if clients, ok := h.competitionClients[client.CompetitionID]; ok {
			delete(clients, client.ID)
			if len(clients) == 0 {
				delete(h.competitionClients, client.CompetitionID)
			}
		}

		close(client.Send)

		h.logger.Info("WebSocket客户端已断开",
			zap.String("client_id", client.ID),
			zap.Uint64("competition_id", client.CompetitionID),
			zap.Int("total_clients", len(h.clients)))
	}
}

// broadcastMessage 广播消息给特定比赛的所有客户端
func (h *Hub) broadcastMessage(message Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// 获取目标比赛的客户端列表
	clients, ok := h.competitionClients[message.CompetitionID]
	if !ok {
		return
	}

	// 发送给该比赛的所有客户端
	for _, client := range clients {
		select {
		case client.Send <- message:
		default:
			// 如果发送缓冲区满，关闭客户端
			h.unregister <- client
		}
	}

	h.logger.Debug("广播消息",
		zap.String("type", message.Type),
		zap.Uint64("competition_id", message.CompetitionID),
		zap.Int("client_count", len(clients)))
}

// WritePump 写入循环，将消息发送给客户端
func (c *Client) WritePump() {
	ticker := time.NewTicker(54 * time.Second) // 心跳间隔
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// Hub关闭了通道
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// 序列化消息
			data, err := json.Marshal(message)
			if err != nil {
				c.Logger.Error("序列化消息失败", zap.Error(err))
				return
			}

			// 发送消息
			if err := c.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
				c.Logger.Error("发送消息失败", zap.Error(err))
				return
			}

		case <-ticker.C:
			// 发送心跳
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ReadPump 读取循环，从客户端读取消息
func (c *Client) ReadPump(hub *Hub) {
	defer func() {
		hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.Logger.Error("WebSocket读取错误", zap.Error(err))
			}
			break
		}

		// 处理客户端发来的消息（如果需要）
		c.Logger.Debug("收到客户端消息", zap.String("message", string(message)))
	}
}
