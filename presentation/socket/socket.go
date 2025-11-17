package socket

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/hiamthach108/simplerank/pkg/logger"
	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
	"golang.org/x/net/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

// Topic types for different broadcast channels
const (
	TopicLeaderboard  = "leaderboard:"  // leaderboard:id
	TopicUser         = "user:"         // user:id
	TopicGlobal       = "global"        // broadcast to all
	TopicNotification = "notification:" // notification:userId
	TopicChat         = "chat:"         // chat:roomId
)

// MessageType represents the type of WebSocket message
type MessageType string

const (
	MessageTypeLeaderboardUpdate MessageType = "leaderboard_update"
	MessageTypeEntryUpdate       MessageType = "entry_update"
	MessageTypeSubscribe         MessageType = "subscribe"
	MessageTypeUnsubscribe       MessageType = "unsubscribe"
	MessageTypePing              MessageType = "ping"
	MessageTypePong              MessageType = "pong"
	MessageTypeNotification      MessageType = "notification"
	MessageTypeUserUpdate        MessageType = "user_update"
	MessageTypeSystemMessage     MessageType = "system_message"
	MessageTypeChatMessage       MessageType = "chat_message"
	MessageTypeError             MessageType = "error"
)

// Message represents a WebSocket message
type Message struct {
	Type      MessageType `json:"type"`
	Topic     string      `json:"topic,omitempty"` // Generic topic (replaces leaderboardId)
	Data      any         `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// Client represents a WebSocket client connection
type Client struct {
	hub            *Hub
	conn           *websocket.Conn
	send           chan []byte
	subscriptions  map[string]bool // subscribed topics
	userID         string          // optional user identifier
	subscriptionMu sync.RWMutex
}

// Hub maintains active clients and broadcasts messages
type Hub struct {
	clients     map[string]map[*Client]bool
	allClients  map[*Client]bool
	userClients map[string]map[*Client]bool
	register    chan *clientRegistration
	unregister  chan *Client
	broadcast   chan *BroadcastMessage
	mu          sync.RWMutex

	logger logger.ILogger
}

type clientRegistration struct {
	client *Client
	topic  string
}

// BroadcastMessage represents a message to be broadcast to specific clients
type BroadcastMessage struct {
	Topic   string // Topic to broadcast to (empty means broadcast to all)
	Message []byte
	UserID  string // Optional: for user-specific broadcasts
}

// NewHub creates a new WebSocket hub
func NewHub(logger logger.ILogger) *Hub {
	return &Hub{
		clients:     make(map[string]map[*Client]bool),
		allClients:  make(map[*Client]bool),
		userClients: make(map[string]map[*Client]bool),
		register:    make(chan *clientRegistration),
		unregister:  make(chan *Client),
		broadcast:   make(chan *BroadcastMessage, 256),
		logger:      logger,
	}
}

// Run starts the hub's main loop
func (h *Hub) Run(ctx context.Context) {
	h.logger.Info("Starting WebSocket Hub...")

	for {
		select {
		case <-ctx.Done():
			h.logger.Info("WebSocket Hub context cancelled, shutting down...")
			return

		case registration := <-h.register:
			h.mu.Lock()

			// Register to topic-specific clients
			if _, ok := h.clients[registration.topic]; !ok {
				h.clients[registration.topic] = make(map[*Client]bool)
			}
			h.clients[registration.topic][registration.client] = true

			// Register to all clients
			h.allClients[registration.client] = true

			// Register to user-specific clients if userID is set
			if registration.client.userID != "" {
				if _, ok := h.userClients[registration.client.userID]; !ok {
					h.userClients[registration.client.userID] = make(map[*Client]bool)
				}
				h.userClients[registration.client.userID][registration.client] = true
			}

			h.mu.Unlock()
			h.logger.Info("Client registered",
				"topic", registration.topic,
				"userID", registration.client.userID,
				"totalClients", len(h.clients[registration.topic]))

		case client := <-h.unregister:
			h.mu.Lock()

			// Remove from all clients
			delete(h.allClients, client)

			// Remove from user-specific clients
			if client.userID != "" {
				if clients, ok := h.userClients[client.userID]; ok {
					delete(clients, client)
					if len(clients) == 0 {
						delete(h.userClients, client.userID)
					}
				}
			}

			// Remove from topic-specific clients
			client.subscriptionMu.RLock()
			for topic := range client.subscriptions {
				if clients, ok := h.clients[topic]; ok {
					if _, ok := clients[client]; ok {
						delete(clients, client)
						h.logger.Info("Client unregistered",
							"topic", topic,
							"userID", client.userID,
							"remainingClients", len(clients))
					}
					if len(clients) == 0 {
						delete(h.clients, topic)
					}
				}
			}
			client.subscriptionMu.RUnlock()
			close(client.send)
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()

			var targetClients map[*Client]bool

			if message.UserID != "" {
				// Broadcast to specific user
				targetClients = h.userClients[message.UserID]
			} else if message.Topic == "" || message.Topic == TopicGlobal {
				// Broadcast to all clients
				targetClients = h.allClients
			} else {
				// Broadcast to topic-specific clients
				targetClients = h.clients[message.Topic]
			}

			for client := range targetClients {
				select {
				case client.send <- message.Message:
				default:
					h.logger.Warn("Client send channel full, closing connection")
					close(client.send)
					delete(targetClients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast sends a message to all clients subscribed to a specific topic
func (h *Hub) Broadcast(topic string, messageType MessageType, data any) {
	message := Message{
		Type:      messageType,
		Topic:     topic,
		Data:      data,
		Timestamp: time.Now(),
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		h.logger.Error("Failed to marshal broadcast message", "error", err)
		return
	}

	h.broadcast <- &BroadcastMessage{
		Topic:   topic,
		Message: jsonData,
	}
}

// BroadcastToUser sends a message to a specific user across all their connections
func (h *Hub) BroadcastToUser(userID string, messageType MessageType, data any) {
	message := Message{
		Type:      messageType,
		Topic:     TopicUser + userID,
		Data:      data,
		Timestamp: time.Now(),
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		h.logger.Error("Failed to marshal user broadcast message", "error", err)
		return
	}

	h.broadcast <- &BroadcastMessage{
		UserID:  userID,
		Message: jsonData,
	}
}

// BroadcastToAll sends a message to all connected clients
func (h *Hub) BroadcastToAll(messageType MessageType, data any) {
	message := Message{
		Type:      messageType,
		Topic:     TopicGlobal,
		Data:      data,
		Timestamp: time.Now(),
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		h.logger.Error("Failed to marshal global broadcast message", "error", err)
		return
	}

	h.broadcast <- &BroadcastMessage{
		Topic:   TopicGlobal,
		Message: jsonData,
	}
}

// GetConnectedClients returns the count of clients subscribed to a topic
func (h *Hub) GetConnectedClients(topic string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients[topic])
}

// GetTotalConnections returns the total number of active WebSocket connections
func (h *Hub) GetTotalConnections() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.allClients)
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(pongWait))

	for {
		var msg Message
		err := websocket.JSON.Receive(c.conn, &msg)
		if err != nil {
			if err.Error() != "EOF" {
				c.hub.logger.Error("Error reading from websocket", "error", err)
			}
			break
		}

		c.handleMessage(&msg)
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel
				c.conn.Close()
				return
			}

			if err := websocket.Message.Send(c.conn, string(message)); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			ping := Message{
				Type:      MessageTypePing,
				Timestamp: time.Now(),
			}
			if err := websocket.JSON.Send(c.conn, ping); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming messages from clients
func (c *Client) handleMessage(msg *Message) {
	switch msg.Type {
	case MessageTypeSubscribe:
		topic := msg.Topic
		if topic != "" {
			c.subscriptionMu.Lock()
			c.subscriptions[topic] = true
			c.subscriptionMu.Unlock()

			c.hub.register <- &clientRegistration{
				client: c,
				topic:  topic,
			}

			// Send confirmation
			response := Message{
				Type:      MessageTypeSubscribe,
				Topic:     topic,
				Data:      map[string]string{"status": "subscribed"},
				Timestamp: time.Now(),
			}
			c.sendMessage(response)
			c.hub.logger.Info("Client subscribed to topic", "topic", topic, "userID", c.userID)
		}

	case MessageTypeUnsubscribe:
		topic := msg.Topic
		if topic != "" {
			c.subscriptionMu.Lock()
			delete(c.subscriptions, topic)
			c.subscriptionMu.Unlock()

			// Send confirmation
			response := Message{
				Type:      MessageTypeUnsubscribe,
				Topic:     topic,
				Data:      map[string]string{"status": "unsubscribed"},
				Timestamp: time.Now(),
			}
			c.sendMessage(response)
			c.hub.logger.Info("Client unsubscribed from topic", "topic", topic, "userID", c.userID)
		}

	case MessageTypePong:
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
	}
}

// sendMessage sends a message to the client
func (c *Client) sendMessage(msg Message) {
	jsonData, err := json.Marshal(msg)
	if err != nil {
		c.hub.logger.Error("Failed to marshal message", "error", err)
		return
	}

	select {
	case c.send <- jsonData:
	default:
		c.hub.logger.Warn("Client send buffer full")
	}
}

// HandleWebSocket handles WebSocket connections
func HandleWebSocket(hub *Hub) echo.HandlerFunc {
	return func(c echo.Context) error {
		websocket.Handler(func(ws *websocket.Conn) {
			// Extract user ID from query parameter or header (optional)
			userID := c.QueryParam("userId")
			if userID == "" {
				userID = c.Request().Header.Get("X-User-ID")
			}

			client := &Client{
				hub:           hub,
				conn:          ws,
				send:          make(chan []byte, 256),
				subscriptions: make(map[string]bool),
				userID:        userID,
			}

			// Start client goroutines
			go client.writePump()
			client.readPump()
		}).ServeHTTP(c.Response(), c.Request())

		return nil
	}
}

// RegisterHooks registers the WebSocket hub lifecycle hooks
func RegisterHooks(lc fx.Lifecycle, hub *Hub) {
	var cancel context.CancelFunc
	var ctx context.Context

	lc.Append(fx.Hook{
		OnStart: func(startCtx context.Context) error {
			ctx, cancel = context.WithCancel(context.Background())
			go hub.Run(ctx)
			hub.logger.Info("WebSocket Hub started successfully")
			return nil
		},
		OnStop: func(stopCtx context.Context) error {
			hub.logger.Info("Shutting down WebSocket Hub...")
			if cancel != nil {
				cancel()
			}

			// Close all client connections
			hub.mu.Lock()
			for _, clients := range hub.clients {
				for client := range clients {
					client.conn.Close()
				}
			}
			hub.mu.Unlock()

			hub.logger.Info("WebSocket Hub stopped successfully")
			return nil
		},
	})
}
