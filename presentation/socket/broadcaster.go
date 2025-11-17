package socket

// IBroadcaster defines the interface for broadcasting messages via WebSocket
type IBroadcaster interface {
	Broadcast(topic string, messageType MessageType, data any)
	BroadcastToUser(userID string, messageType MessageType, data any)
	BroadcastToAll(messageType MessageType, data any)
	GetConnectedClients(topic string) int
	GetTotalConnections() int
}
