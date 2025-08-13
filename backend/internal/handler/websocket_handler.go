package handler

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

// WebSocketMessage represents a message sent through WebSocket
type WebSocketMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// WebSocketClient represents a connected client
type WebSocketClient struct {
	ID       string
	Conn     *websocket.Conn
	Send     chan WebSocketMessage
	Hub      *WebSocketHub
	LastPing time.Time
	mu       sync.RWMutex
}

// WebSocketHub manages all WebSocket connections
type WebSocketHub struct {
	Clients    map[string]*WebSocketClient
	Broadcast  chan WebSocketMessage
	Register   chan *WebSocketClient
	Unregister chan *WebSocketClient
	Mutex      sync.RWMutex
}

// WebSocketHandler handles WebSocket connections
type WebSocketHandler struct {
	hub *WebSocketHub
}

// Ensure WebSocketHandler implements WebSocketBroadcaster interface
var _ interface {
	BroadcastMessage(msgType string, data interface{})
	SendMessageToClient(clientID, msgType string, data interface{}) bool
	GetConnectedClients() int
	GetClientIDs() []string
} = (*WebSocketHandler)(nil)

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler() *WebSocketHandler {
	hub := &WebSocketHub{
		Clients:    make(map[string]*WebSocketClient),
		Broadcast:  make(chan WebSocketMessage),
		Register:   make(chan *WebSocketClient),
		Unregister: make(chan *WebSocketClient),
	}

	handler := &WebSocketHandler{
		hub: hub,
	}

	// Start the hub
	go handler.hub.Run()

	return handler
}

// Run starts the WebSocket hub
func (h *WebSocketHub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Mutex.Lock()
			h.Clients[client.ID] = client
			h.Mutex.Unlock()
			log.Printf("WebSocket client connected: %s", client.ID)

		case client := <-h.Unregister:
			h.Mutex.Lock()
			if _, ok := h.Clients[client.ID]; ok {
				delete(h.Clients, client.ID)
				close(client.Send)
				log.Printf("WebSocket client disconnected: %s", client.ID)
			}
			h.Mutex.Unlock()

		case message := <-h.Broadcast:
			h.Mutex.RLock()
			for _, client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					// Client's send channel is full, close it
					close(client.Send)
					delete(h.Clients, client.ID)
				}
			}
			h.Mutex.RUnlock()
		}
	}
}

// HandleWebSocket handles WebSocket connections
func (wh *WebSocketHandler) HandleWebSocket(c *fiber.Ctx) error {
	// Use websocket.IsWebSocketUpgrade to check if it's a websocket request
	if !websocket.IsWebSocketUpgrade(c) {
		return fiber.ErrUpgradeRequired
	}

	// Use Fiber's websocket middleware approach
	return websocket.New(func(c *websocket.Conn) {
		// Generate unique client ID
		clientID := generateClientID()

		// Create new client
		client := &WebSocketClient{
			ID:       clientID,
			Conn:     c,
			Send:     make(chan WebSocketMessage, 256),
			Hub:      wh.hub,
			LastPing: time.Now(),
		}

		// Register client
		wh.hub.Register <- client

		// Start goroutines for reading and writing
		go client.writePump()
		client.readPump()
	})(c)
}

// readPump handles reading messages from the WebSocket connection
func (c *WebSocketClient) readPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	for {
		var msg WebSocketMessage
		err := c.Conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}

		// Handle incoming messages here
		log.Printf("Received message from client %s: %+v", c.ID, msg)
		
		// Echo the message back to the client for testing
		response := WebSocketMessage{
			Type: "echo",
			Data: map[string]interface{}{
				"original": msg,
				"timestamp": time.Now(),
			},
		}
		
		select {
		case c.Send <- response:
		default:
			log.Printf("Client %s send channel is full", c.ID)
		}
	}
}

// writePump handles writing messages to the WebSocket connection
func (c *WebSocketClient) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				// The hub closed the channel
				return
			}

			if err := c.Conn.WriteJSON(message); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}

		case <-ticker.C:
			// Send ping message
			pingMsg := WebSocketMessage{
				Type: "ping",
				Data: map[string]interface{}{"timestamp": time.Now()},
			}
			if err := c.Conn.WriteJSON(pingMsg); err != nil {
				return
			}
		}
	}
}

// BroadcastMessage broadcasts a message to all connected clients
func (wh *WebSocketHandler) BroadcastMessage(msgType string, data interface{}) {
	message := WebSocketMessage{
		Type: msgType,
		Data: data,
	}

	select {
	case wh.hub.Broadcast <- message:
	default:
		log.Println("Broadcast channel is full, message dropped")
	}
}

// SendMessageToClient sends a message to a specific client
func (wh *WebSocketHandler) SendMessageToClient(clientID, msgType string, data interface{}) bool {
	wh.hub.Mutex.RLock()
	client, exists := wh.hub.Clients[clientID]
	wh.hub.Mutex.RUnlock()

	if !exists {
		return false
	}

	message := WebSocketMessage{
		Type: msgType,
		Data: data,
	}

	select {
	case client.Send <- message:
		return true
	default:
		return false
	}
}

// GetConnectedClients returns the number of connected clients
func (wh *WebSocketHandler) GetConnectedClients() int {
	wh.hub.Mutex.RLock()
	defer wh.hub.Mutex.RUnlock()
	return len(wh.hub.Clients)
}

// GetClientIDs returns a list of connected client IDs
func (wh *WebSocketHandler) GetClientIDs() []string {
	wh.hub.Mutex.RLock()
	defer wh.hub.Mutex.RUnlock()
	
	ids := make([]string, 0, len(wh.hub.Clients))
	for id := range wh.hub.Clients {
		ids = append(ids, id)
	}
	return ids
}

// RegisterRoutes registers WebSocket routes
func (wh *WebSocketHandler) RegisterRoutes(app fiber.Router) {
	// WebSocket endpoint
	app.Get("/ws", wh.HandleWebSocket)

	// WebSocket status endpoint
	app.Get("/ws/status", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"connected_clients": wh.GetConnectedClients(),
			"client_ids":        wh.GetClientIDs(),
			"status":            "active",
		})
	})

	// Broadcast endpoint for testing
	app.Post("/ws/broadcast", func(c *fiber.Ctx) error {
		var req struct {
			Type string      `json:"type"`
			Data interface{} `json:"data"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid JSON",
			})
		}

		wh.BroadcastMessage(req.Type, req.Data)

		return c.JSON(fiber.Map{
			"message": "Message broadcasted",
			"clients": wh.GetConnectedClients(),
		})
	})

	// Send message to specific client endpoint
	app.Post("/ws/send/:clientId", func(c *fiber.Ctx) error {
		clientID := c.Params("clientId")
		
		var req struct {
			Type string      `json:"type"`
			Data interface{} `json:"data"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid JSON",
			})
		}

		success := wh.SendMessageToClient(clientID, req.Type, req.Data)
		if !success {
			return c.Status(404).JSON(fiber.Map{
				"error": "Client not found or message could not be sent",
			})
		}

		return c.JSON(fiber.Map{
			"message": "Message sent to client",
			"client":  clientID,
		})
	})
}

// generateClientID generates a unique client ID
func generateClientID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}