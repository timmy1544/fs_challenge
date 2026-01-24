package websocket

import (
	"encoding/json"
	"log"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Hub maintains the set of active clients and broadcasts messages to clients
type Hub struct {
	// Registered clients mapped by project ID
	clients map[uuid.UUID]map[*Client]bool

	// Inbound messages from clients
	broadcast chan *Message

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Redis client for pub/sub
	redisClient *redis.Client
}

// Message represents a WebSocket message with delta updates
type Message struct {
	Type      string                 `json:"type"`      // TASK_CREATED, TASK_UPDATED, TASK_DELETED, COMMENT_CREATED, COMMENT_UPDATED, COMMENT_DELETED, PROJECT_UPDATED
	ProjectID uuid.UUID             `json:"project_id"`
	TaskID    *uuid.UUID            `json:"task_id,omitempty"`
	CommentID *uuid.UUID            `json:"comment_id,omitempty"`
	// Delta contains only the changed fields to avoid sending full objects
	Delta     map[string]interface{} `json:"delta,omitempty"`
	// FullData is only sent for creates or when explicitly needed
	FullData  interface{}            `json:"full_data,omitempty"`
}

// NewHub creates a new Hub instance
func NewHub(redisClient *redis.Client) *Hub {
	hub := &Hub{
		clients:     make(map[uuid.UUID]map[*Client]bool),
		broadcast:   make(chan *Message, 256),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		redisClient: redisClient,
	}

	// Subscribe to Redis pub/sub for cross-instance communication
	go hub.subscribeToRedis()

	return hub
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastToClients(message)
		}
	}
}

func (h *Hub) registerClient(client *Client) {
	projectID := client.projectID
	if h.clients[projectID] == nil {
		h.clients[projectID] = make(map[*Client]bool)
	}
	h.clients[projectID][client] = true
	log.Printf("Client registered for project %s. Total clients: %d", projectID, len(h.clients[projectID]))
}

func (h *Hub) unregisterClient(client *Client) {
	projectID := client.projectID
	if clients, ok := h.clients[projectID]; ok {
		if _, exists := clients[client]; exists {
			delete(clients, client)
			close(client.send)
			if len(clients) == 0 {
				delete(h.clients, projectID)
			}
			log.Printf("Client unregistered from project %s. Remaining clients: %d", projectID, len(h.clients[projectID]))
		}
	}
}

func (h *Hub) broadcastToClients(message *Message) {
	projectID := message.ProjectID
	if clients, ok := h.clients[projectID]; ok {
		data, err := json.Marshal(message)
		if err != nil {
			log.Printf("Error marshaling message: %v", err)
			return
		}

		for client := range clients {
			select {
			case client.send <- data:
			default:
				close(client.send)
				delete(clients, client)
			}
		}
	}
}

// Broadcast sends a message to all clients in a project
func (h *Hub) Broadcast(message *Message) {
	// Publish to Redis for cross-instance communication
	channel := "project:" + message.ProjectID.String()
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling message for Redis: %v", err)
		return
	}

	if err := h.redisClient.Publish(h.redisClient.Context(), channel, data).Err(); err != nil {
		log.Printf("Error publishing to Redis: %v", err)
	}

	// Also broadcast locally
	h.broadcast <- message
}

// subscribeToRedis subscribes to Redis pub/sub channels
func (h *Hub) subscribeToRedis() {
	pubsub := h.redisClient.PSubscribe(h.redisClient.Context(), "project:*")
	defer pubsub.Close()

	ch := pubsub.Channel()
	for msg := range ch {
		var message Message
		if err := json.Unmarshal([]byte(msg.Payload), &message); err != nil {
			log.Printf("Error unmarshaling Redis message: %v", err)
			continue
		}

		// Broadcast to local clients (avoid re-publishing to Redis)
		h.broadcast <- &message
	}
}
