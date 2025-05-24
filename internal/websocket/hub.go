package websocket

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/tehsis/logmeup-api/internal/models"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin in development
		// In production, you should restrict this to your domain
		return true
	},
}

// Message types for WebSocket communication
type MessageType string

const (
	ActionCreated MessageType = "action_created"
	ActionUpdated MessageType = "action_updated"
	ActionDeleted MessageType = "action_deleted"
)

// WebSocket message structure
type Message struct {
	Type MessageType `json:"type"`
	Data interface{} `json:"data"`
}

// ActionMessage for action-related events
type ActionMessage struct {
	Type   MessageType    `json:"type"`
	Action *models.Action `json:"action,omitempty"`
	ID     int64          `json:"id,omitempty"` // For delete events
}

// Client represents a WebSocket connection
type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
}

// Hub maintains the set of active clients and broadcasts messages to them
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Inbound messages from the clients
	broadcast chan []byte

	// Register requests from the clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

// Run starts the hub and handles client registration/unregistration
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			log.Printf("Client connected. Total clients: %d", len(h.clients))

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("Client disconnected. Total clients: %d", len(h.clients))
			}

		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

// BroadcastActionCreated broadcasts when an action is created
func (h *Hub) BroadcastActionCreated(action *models.Action) {
	message := ActionMessage{
		Type:   ActionCreated,
		Action: action,
	}
	h.broadcastMessage(message)
}

// BroadcastActionUpdated broadcasts when an action is updated
func (h *Hub) BroadcastActionUpdated(action *models.Action) {
	message := ActionMessage{
		Type:   ActionUpdated,
		Action: action,
	}
	h.broadcastMessage(message)
}

// BroadcastActionDeleted broadcasts when an action is deleted
func (h *Hub) BroadcastActionDeleted(actionID int64) {
	message := ActionMessage{
		Type: ActionDeleted,
		ID:   actionID,
	}
	h.broadcastMessage(message)
}

// broadcastMessage sends a message to all connected clients
func (h *Hub) broadcastMessage(message interface{}) {
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	log.Printf("Broadcasting message: %s", string(data))
	h.broadcast <- data
}

// HandleWebSocket handles WebSocket connection requests
func (h *Hub) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &Client{
		hub:  h,
		conn: conn,
		send: make(chan []byte, 256),
	}

	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}

// readPump pumps messages from the websocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
	}
}

// writePump pumps messages from the hub to the websocket connection
func (c *Client) writePump() {
	defer c.conn.Close()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}
		}
	}
}
