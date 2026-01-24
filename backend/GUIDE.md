# Go Backend Walkthrough Guide

This guide walks you through the Go backend service, explaining the architecture, code structure, and how everything works together.

## Table of Contents

1. [Project Structure](#project-structure)
2. [Application Startup](#application-startup)
3. [Architecture Layers](#architecture-layers)
4. [Request Flow Example](#request-flow-example)
5. [Key Go Concepts](#key-go-concepts)
6. [Understanding Each Component](#understanding-each-component)

---

## Project Structure

```
backend/
├── cmd/
│   └── server/
│       └── main.go          # Application entry point
├── internal/                 # Private application code
│   ├── api/                  # HTTP handlers and routes
│   ├── config/               # Configuration management
│   ├── database/             # Database models and connection
│   ├── repository/           # Data access layer
│   └── websocket/            # WebSocket hub and clients
└── go.mod                    # Go module dependencies
```

**Key Point**: The `internal/` directory contains code that can't be imported by other Go modules. This keeps our API private.

---

## Application Startup

Let's trace what happens when you run `go run cmd/server/main.go`:

### Step 1: Entry Point (`cmd/server/main.go`)

```go
func main() {
    // 1. Load environment variables from .env file
    godotenv.Load()
    
    // 2. Load configuration
    cfg := config.Load()
    
    // 3. Connect to PostgreSQL database
    db, err := database.NewConnection(cfg.DatabaseURL)
    
    // 4. Run database migrations (create tables)
    database.RunMigrations(db)
    
    // 5. Connect to Redis (for WebSocket scaling)
    redisClient, err := database.NewRedisClient(cfg.RedisURL)
    
    // 6. Create WebSocket hub (manages real-time connections)
    hub := websocket.NewHub(redisClient)
    go hub.Run()  // Start hub in a goroutine (background process)
    
    // 7. Create HTTP router with all routes
    router := api.NewRouter(db, hub, redisClient)
    
    // 8. Start HTTP server on port 8080
    router.Run(":8080")
}
```

**What's happening:**
- The `main()` function is the entry point of every Go program
- We initialize dependencies in order: config → database → Redis → WebSocket → HTTP server
- `go hub.Run()` starts the WebSocket hub in a separate goroutine (concurrent execution)

---

## Architecture Layers

The backend follows a **layered architecture** pattern:

```
┌─────────────────────────────────────┐
│   HTTP Request                      │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│   API Layer (Handlers)              │  ← Handles HTTP requests/responses
│   - Parse request                   │
│   - Validate input                  │
│   - Call repository                 │
│   - Send WebSocket updates          │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│   Repository Layer                  │  ← Data access abstraction
│   - Database queries                │
│   - Business logic                  │
│   - Returns models                  │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│   Database Layer (GORM)            │  ← ORM (Object-Relational Mapping)
│   - SQL queries                     │
│   - Model definitions               │
└─────────────────────────────────────┘
```

### Why This Structure?

- **Separation of Concerns**: Each layer has a single responsibility
- **Testability**: Easy to test each layer independently
- **Maintainability**: Changes in one layer don't affect others
- **Reusability**: Repository can be used by different handlers

---

## Request Flow Example

Let's trace a complete request: **Creating a Task**

### 1. Client Sends Request

```http
POST /api/v1/tasks
Content-Type: application/json

{
  "title": "Build feature",
  "project_id": "123e4567-e89b-12d3-a456-426614174000"
}
```

### 2. Router Matches Route (`internal/api/router.go`)

```go
api.POST("/tasks", taskHandler.CreateTask)
```

The router calls `taskHandler.CreateTask` function.

### 3. Handler Processes Request (`internal/api/task_handler.go`)

```go
func (h *TaskHandler) CreateTask(c *gin.Context) {
    // Step 1: Parse JSON request body
    var req CreateTaskRequest
    c.ShouldBindJSON(&req)  // Gin automatically parses JSON
    
    // Step 2: Create task model
    task := &database.Task{
        Title:     req.Title,
        ProjectID: req.ProjectID,
        Status:    "todo",
    }
    
    // Step 3: Save to database via repository
    h.taskRepo.Create(task)
    
    // Step 4: Broadcast real-time update via WebSocket
    h.hub.Broadcast(&websocket.Message{
        Type:      "TASK_CREATED",
        ProjectID:  task.ProjectID,
        FullData:   task,
    })
    
    // Step 5: Return JSON response
    c.JSON(http.StatusCreated, task)
}
```

**Key Points:**
- `c *gin.Context` is the request/response context (like Express.js `req, res`)
- `c.ShouldBindJSON()` automatically validates and parses JSON
- `c.JSON()` sends JSON response with status code

### 4. Repository Saves to Database (`internal/repository/task_repository.go`)

```go
func (r *TaskRepository) Create(task *database.Task) error {
    return r.db.Create(task).Error
}
```

**What's happening:**
- `r.db` is a GORM database connection
- `Create()` generates SQL: `INSERT INTO tasks (...) VALUES (...)`
- GORM automatically handles UUID generation, timestamps, etc.

### 5. WebSocket Broadcast (`internal/websocket/hub.go`)

```go
func (h *Hub) Broadcast(message *Message) {
    // Publish to Redis (for multi-instance scaling)
    channel := "project:" + message.ProjectID.String()
    h.redisClient.Publish(channel, data)
    
    // Also broadcast to local WebSocket clients
    h.broadcast <- message
}
```

**What's happening:**
- Message is published to Redis channel
- All server instances subscribed to that channel receive it
- Each instance broadcasts to its connected WebSocket clients

### 6. Response Sent to Client

```json
HTTP/1.1 201 Created
Content-Type: application/json

{
  "id": "789e4567-e89b-12d3-a456-426614174000",
  "title": "Build feature",
  "status": "todo",
  ...
}
```

---

## Key Go Concepts

### 1. Packages

```go
package api  // Package name
```

- Every `.go` file belongs to a package
- Files in the same directory must have the same package name
- `package main` is special - it's the entry point

### 2. Imports

```go
import (
    "net/http"                    // Standard library
    "github.com/gin-gonic/gin"    // Third-party package
    "yourusername/task-manager/internal/database"  // Local package
)
```

- Standard library: `"net/http"`, `"fmt"`, `"time"`
- Third-party: `"github.com/gin-gonic/gin"`
- Local: `"yourusername/task-manager/..."`

### 3. Structs (Like Classes)

```go
type TaskHandler struct {
    taskRepo *repository.TaskRepository
    hub      *websocket.Hub
}
```

- Similar to classes in OOP languages
- Fields can be public (capitalized) or private (lowercase)
- Methods are defined separately

### 4. Methods (Functions on Types)

```go
func (h *TaskHandler) CreateTask(c *gin.Context) {
    // h is the receiver (like 'this' in other languages)
    // Can access h.taskRepo, h.hub, etc.
}
```

- `(h *TaskHandler)` is the receiver
- `*` means pointer (reference)
- Methods can be called: `handler.CreateTask(context)`

### 5. Pointers

```go
task := &database.Task{...}  // & creates a pointer
func Create(task *database.Task)  // * means pointer parameter
```

- `&` gets the address (pointer) of a value
- `*` dereferences a pointer (gets the value)
- Pointers allow functions to modify the original value

### 6. Error Handling

```go
db, err := database.NewConnection(url)
if err != nil {
    log.Fatalf("Failed: %v", err)  // Exit program
    return err                       // Return error
}
```

- Go functions often return `(result, error)`
- Always check errors explicitly
- No exceptions - errors are values

### 7. Goroutines (Concurrency)

```go
go hub.Run()  // Runs in background
```

- `go` keyword starts a goroutine (lightweight thread)
- Allows concurrent execution
- Used for WebSocket hub, background tasks

### 8. Channels

```go
broadcast chan *Message  // Channel that carries Message pointers
message := <-broadcast   // Receive from channel
broadcast <- message     // Send to channel
```

- Channels are Go's way of communicating between goroutines
- Thread-safe communication
- Used in WebSocket hub for message passing

---

## Understanding Each Component

### 1. Configuration (`internal/config/config.go`)

```go
type Config struct {
    DatabaseURL string
    RedisURL    string
    JWTSecret   string
}

func Load() *Config {
    return &Config{
        DatabaseURL: getEnv("DATABASE_URL", "default-value"),
        // ...
    }
}
```

**Purpose**: Centralized configuration management
- Reads from environment variables
- Provides defaults
- Single source of truth

### 2. Database Models (`internal/database/models.go`)

```go
type Task struct {
    ID          uuid.UUID `gorm:"type:uuid;primary_key"`
    Title       string    `gorm:"not null"`
    Status      string    `gorm:"default:'todo'"`
    ProjectID   uuid.UUID `gorm:"type:uuid;not null;index"`
    Version     int       `gorm:"default:1"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

**Purpose**: Define database schema
- Struct tags (`` `gorm:"..."` ``) tell GORM how to map to database
- `BeforeCreate` hooks run automatically before saving

### 3. Repository Pattern (`internal/repository/task_repository.go`)

```go
type TaskRepository struct {
    db *gorm.DB
}

func (r *TaskRepository) Create(task *database.Task) error {
    return r.db.Create(task).Error
}

func (r *TaskRepository) GetByID(id uuid.UUID) (*database.Task, error) {
    var task database.Task
    err := r.db.Where("id = ?", id).First(&task).Error
    return &task, err
}
```

**Purpose**: Data access abstraction
- All database queries go through repositories
- Handlers don't know about SQL
- Easy to swap database implementations
- Business logic lives here (e.g., optimistic locking)

### 4. API Handlers (`internal/api/task_handler.go`)

```go
type TaskHandler struct {
    taskRepo *repository.TaskRepository
    hub      *websocket.Hub
}

func (h *TaskHandler) CreateTask(c *gin.Context) {
    // 1. Parse request
    var req CreateTaskRequest
    c.ShouldBindJSON(&req)
    
    // 2. Business logic
    task := &database.Task{...}
    
    // 3. Save via repository
    h.taskRepo.Create(task)
    
    // 4. Real-time update
    h.hub.Broadcast(...)
    
    // 5. Response
    c.JSON(201, task)
}
```

**Purpose**: HTTP request/response handling
- Receives HTTP requests
- Validates input
- Orchestrates business logic
- Sends responses
- Triggers real-time updates

### 5. Router (`internal/api/router.go`)

```go
func NewRouter(db *gorm.DB, hub *websocket.Hub) *gin.Engine {
    router := gin.Default()
    
    api := router.Group("/api/v1")
    {
        taskHandler := NewTaskHandler(db, hub)
        api.POST("/tasks", taskHandler.CreateTask)
        api.GET("/tasks", taskHandler.GetTasks)
        // ...
    }
    
    return router
}
```

**Purpose**: Route HTTP requests to handlers
- Defines URL patterns
- Groups related routes
- Applies middleware (CORS, auth, etc.)

### 6. WebSocket Hub (`internal/websocket/hub.go`)

```go
type Hub struct {
    clients    map[uuid.UUID]map[*Client]bool  // Project -> Clients
    broadcast  chan *Message
    register   chan *Client
    unregister chan *Client
}

func (h *Hub) Run() {
    for {
        select {
        case client := <-h.register:
            h.registerClient(client)
        case message := <-h.broadcast:
            h.broadcastToClients(message)
        }
    }
}
```

**Purpose**: Manages WebSocket connections
- Tracks connected clients by project
- Broadcasts messages to all clients in a project
- Handles connection lifecycle
- Uses channels for thread-safe communication

### 7. WebSocket Client (`internal/websocket/client.go`)

```go
type Client struct {
    hub      *Hub
    conn     *websocket.Conn
    send     chan []byte
    projectID uuid.UUID
}

func (c *Client) readPump() {
    // Reads messages from WebSocket
}

func (c *Client) writePump() {
    // Writes messages to WebSocket
}
```

**Purpose**: Individual WebSocket connection
- Each client connection has its own goroutines
- `readPump`: Receives messages from browser
- `writePump`: Sends messages to browser
- Handles ping/pong for connection health

---

## Common Patterns

### 1. Dependency Injection

```go
func NewTaskHandler(db *gorm.DB, hub *websocket.Hub) *TaskHandler {
    return &TaskHandler{
        taskRepo: repository.NewTaskRepository(db),
        hub:      hub,
    }
}
```

**Why**: Makes testing easier, dependencies explicit

### 2. Error Handling Pattern

```go
result, err := someFunction()
if err != nil {
    return nil, err  // Propagate error
}
// Use result
```

### 3. JSON Tag Binding

```go
type CreateTaskRequest struct {
    Title string `json:"title" binding:"required"`
}
```

- `json:"title"` maps JSON field to struct field
- `binding:"required"` validates field is present

### 4. Context Pattern (Gin)

```go
func (h *Handler) GetTask(c *gin.Context) {
    id := c.Param("id")           // URL parameter: /tasks/:id
    query := c.Query("status")     // Query string: ?status=todo
    var body Request
    c.ShouldBindJSON(&body)        // Request body
    c.JSON(200, result)             // Response
}
```

---

## How to Extend

### Adding a New Endpoint

1. **Add route** in `router.go`:
```go
api.GET("/new-endpoint", handler.NewEndpoint)
```

2. **Create handler method**:
```go
func (h *Handler) NewEndpoint(c *gin.Context) {
    // Implementation
}
```

### Adding a New Model

1. **Define model** in `database/models.go`
2. **Add to migrations** in `database/database.go`:
```go
db.AutoMigrate(&NewModel{})
```
3. **Create repository** in `repository/`
4. **Create handler** in `api/`

---

## Testing the Backend

### Manual Testing with curl

```bash
# Create a task
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{"title":"Test","project_id":"..."}'

# Get tasks
curl http://localhost:8080/api/v1/tasks?project_id=...
```

### Check Health

```bash
curl http://localhost:8080/health
```

---

## Next Steps

1. **Read the code**: Start with `main.go` and follow the flow
2. **Add logging**: Use `log.Printf()` to trace execution
3. **Add validation**: Enhance input validation
4. **Add tests**: Write unit tests for repositories
5. **Add authentication**: Implement JWT authentication

---

## Common Questions

**Q: Why use GORM instead of raw SQL?**
A: GORM provides type safety, automatic migrations, and handles common operations. You can still use raw SQL when needed.

**Q: What's the difference between `internal/` and `pkg/`?**
A: `internal/` can't be imported by other modules. `pkg/` can be imported by external code.

**Q: Why goroutines for WebSocket hub?**
A: The hub needs to run continuously in the background while the HTTP server handles requests. Goroutines allow concurrent execution.

**Q: How does Redis help with scaling?**
A: When you have multiple server instances, Redis pub/sub ensures WebSocket messages reach all instances, which then broadcast to their local clients.

---

This should give you a solid foundation! Start by reading `main.go` and following the code flow for a simple request like creating a task.
