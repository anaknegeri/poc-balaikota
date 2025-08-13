package service

import (
	"encoding/json"
	"log"
	"time"
)

// WebSocketService handles WebSocket business logic
type WebSocketService struct {
	broadcaster WebSocketBroadcaster
}

// WebSocketBroadcaster interface for broadcasting messages
type WebSocketBroadcaster interface {
	BroadcastMessage(msgType string, data interface{})
	SendMessageToClient(clientID, msgType string, data interface{}) bool
	GetConnectedClients() int
	GetClientIDs() []string
}

// NewWebSocketService creates a new WebSocket service
func NewWebSocketService(broadcaster WebSocketBroadcaster) *WebSocketService {
	return &WebSocketService{
		broadcaster: broadcaster,
	}
}

// NotifyAlert broadcasts alert notifications to all connected clients
func (ws *WebSocketService) NotifyAlert(alertType, cameraName, message string, data interface{}) {
	notification := map[string]interface{}{
		"alert_type":   alertType,
		"camera_name":  cameraName,
		"message":      message,
		"timestamp":    time.Now(),
		"data":         data,
	}

	ws.broadcaster.BroadcastMessage("alert", notification)
	log.Printf("Alert notification sent: %s from %s", alertType, cameraName)
}



// SendPersonalizedMessage sends a message to a specific client
func (ws *WebSocketService) SendPersonalizedMessage(clientID, messageType string, data interface{}) bool {
	message := map[string]interface{}{
		"data":      data,
		"timestamp": time.Now(),
	}

	return ws.broadcaster.SendMessageToClient(clientID, messageType, message)
}

// GetConnectionStats returns connection statistics
func (ws *WebSocketService) GetConnectionStats() map[string]interface{} {
	return map[string]interface{}{
		"connected_clients": ws.broadcaster.GetConnectedClients(),
		"client_ids":        ws.broadcaster.GetClientIDs(),
		"timestamp":         time.Now(),
	}
}



// HandleClientMessage processes incoming messages from clients
func (ws *WebSocketService) HandleClientMessage(clientID string, messageType string, data json.RawMessage) error {
	log.Printf("Processing message from client %s: type=%s", clientID, messageType)

	switch messageType {
	case "ping":
		// Respond with pong
		ws.SendPersonalizedMessage(clientID, "pong", map[string]interface{}{
			"message": "pong",
		})

	case "subscribe":
		// Handle subscription requests
		var subData struct {
			Channels []string `json:"channels"`
		}
		if err := json.Unmarshal(data, &subData); err != nil {
			return err
		}
		
		// Send confirmation
		ws.SendPersonalizedMessage(clientID, "subscription_confirmed", map[string]interface{}{
			"channels": subData.Channels,
			"message":  "Subscribed successfully",
		})

	case "get_stats":
		// Send current statistics
		stats := ws.GetConnectionStats()
		ws.SendPersonalizedMessage(clientID, "stats", stats)

	default:
		log.Printf("Unknown message type: %s from client %s", messageType, clientID)
	}

	return nil
}