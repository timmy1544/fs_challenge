package websocket

import (
	"encoding/json"
	"log"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Hub maintains the set of active clients and broadcasts messages to clients
type Hub struct {
	// Registered clients mapped by workspace ID
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

// Message represents a WebSocket message
type Message struct {
	Type        string      `json:"type"`
	WorkspaceID uuid.UUID   `json:"workspace_id"`
	TaskID      *uuid.UUID  `json:"task_id,omitempty"`
	Data        interface{} `json:"data"`
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
	workspaceID := client.workspaceID
	if h.clients[workspaceID] == nil {
		h.clients[workspaceID] = make(map[*Client]bool)
	}
	h.clients[workspaceID][client] = true
	log.Printf("Client registered for workspace %s. Total clients: %d", workspaceID, len(h.clients[workspaceID]))
}

func (h *Hub) unregisterClient(client *Client) {
	workspaceID := client.workspaceID
	if clients, ok := h.clients[workspaceID]; ok {
		if _, exists := clients[client]; exists {
			delete(clients, client)
			close(client.send)
			if len(clients) == 0 {
				delete(h.clients, workspaceID)
			}
			log.Printf("Client unregistered from workspace %s. Remaining clients: %d", workspaceID, len(h.clients[workspaceID]))
		}
	}
}

func (h *Hub) broadcastToClients(message *Message) {
	workspaceID := message.WorkspaceID
	if clients, ok := h.clients[workspaceID]; ok {
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

// Broadcast sends a message to all clients in a workspace
func (h *Hub) Broadcast(message *Message) {
	// Publish to Redis for cross-instance communication
	channel := "workspace:" + message.WorkspaceID.String()
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
	pubsub := h.redisClient.PSubscribe(h.redisClient.Context(), "workspace:*")
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
