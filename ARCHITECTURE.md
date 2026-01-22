# Architecture Recommendations

## Overview
Building a collaborative task management system with React frontend and Go backend, supporting real-time updates without managed real-time databases.

## Core Architecture Decisions

### 1. Real-Time Communication Strategy

**Recommended: WebSockets with Message Broker Pattern**

- **Go Backend**: Use `gorilla/websocket` or `nhooyr.io/websocket` for WebSocket handling
- **Why WebSockets**: Low latency, bidirectional communication, efficient for real-time updates
- **Scaling Challenge**: WebSocket connections are stateful - need a message broker for multi-instance deployments

**Alternative Approaches:**
- **Server-Sent Events (SSE)**: Simpler, unidirectional (server → client). Good for read-heavy workloads
- **Long Polling**: Fallback option, less efficient but works everywhere

**For Multi-Instance Scaling:**
- Use **Redis Pub/Sub** or **NATS** as message broker
- When instance A receives an update, publish to broker
- All instances (A, B, C) receive message and push to their connected clients

### 2. Database Strategy

**Recommended: PostgreSQL + Redis**

- **PostgreSQL**: Primary data store
  - Use proper indexing (especially on `task_id`, `user_id`, `updated_at`)
  - Consider partitioning for very large datasets (by date or tenant)
  - Use `NOTIFY/LISTEN` for database-level change detection (optional)
  
- **Redis**: 
  - Session/connection management
  - Message broker for WebSocket scaling
  - Caching layer for frequently accessed data
  - Rate limiting

**Schema Considerations:**
```sql
-- Tasks table with optimistic locking
CREATE TABLE tasks (
    id UUID PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    status VARCHAR(50),
    assignee_id UUID,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    version INTEGER DEFAULT 1  -- For optimistic locking
);

-- Task updates log (for conflict resolution)
CREATE TABLE task_updates (
    id UUID PRIMARY KEY,
    task_id UUID REFERENCES tasks(id),
    user_id UUID,
    change_type VARCHAR(50),
    old_value JSONB,
    new_value JSONB,
    timestamp TIMESTAMP DEFAULT NOW()
);
```

### 3. Backend Architecture (Go)

**Recommended Structure:**
```
backend/
├── cmd/
│   └── server/          # Application entry point
├── internal/
│   ├── api/             # HTTP handlers
│   ├── websocket/       # WebSocket hub and handlers
│   ├── models/          # Data models
│   ├── repository/      # Database access layer
│   ├── service/         # Business logic
│   └── broker/          # Message broker (Redis/NATS)
├── pkg/
│   └── middleware/      # Shared middleware
└── migrations/          # Database migrations
```

**Key Components:**

1. **WebSocket Hub**: Manages all active connections
   - Map of `room_id` → `[]*Client`
   - Broadcast messages to all clients in a room
   - Handle connection lifecycle

2. **Message Broker Integration**:
   - Subscribe to Redis channels per room/workspace
   - When message received from broker, broadcast to local clients
   - When local update occurs, publish to broker

3. **Optimistic Locking**:
   - Use version numbers to prevent lost updates
   - Return conflict errors when versions don't match
   - Client retries with latest version

### 4. Frontend Architecture (React)

**Recommended Stack:**
- **React 18+** with TypeScript
- **State Management**: 
  - React Query / TanStack Query for server state
  - Zustand or Context API for client state
- **WebSocket Client**: Native WebSocket API or `socket.io-client`
- **UI Framework**: Tailwind CSS or Material-UI

**Real-Time Strategy:**
1. **WebSocket Connection**: Single persistent connection per user
2. **Message Types**: 
   - `TASK_CREATED`, `TASK_UPDATED`, `TASK_DELETED`
   - `USER_JOINED`, `USER_LEFT`
3. **Optimistic Updates**: Update UI immediately, rollback on error
4. **Conflict Resolution**: Show conflict dialog, merge or overwrite

### 5. Scalability Patterns

**Horizontal Scaling:**
- Stateless API servers (except WebSocket connections)
- Load balancer with sticky sessions for WebSockets (or use message broker)
- Database connection pooling
- Read replicas for read-heavy operations

**Performance Optimizations:**
- **Pagination**: Limit initial data load
- **Incremental Updates**: Only send changed fields
- **Debouncing**: Batch rapid updates (e.g., typing indicators)
- **Virtual Scrolling**: For large task lists
- **IndexedDB**: Cache tasks locally for offline support

**Data Consistency:**
- **Eventual Consistency**: Acceptable for collaborative editing
- **Last-Write-Wins**: Simple but can lose data
- **Operational Transform (OT)**: Complex, good for text editing
- **CRDTs**: Best for conflict-free merging (overkill for simple tasks)

### 6. Security Considerations

- **Authentication**: JWT tokens or session-based
- **Authorization**: Row-level security (users can only see their tasks/workspaces)
- **Rate Limiting**: Prevent abuse (Redis-based)
- **Input Validation**: Sanitize all user inputs
- **CORS**: Properly configured for production

### 7. Deployment Architecture

```
┌─────────────┐
│   Load      │
│  Balancer   │
└──────┬──────┘
       │
   ┌───┴───┬─────────┬─────────┐
   │       │         │         │
┌──▼──┐ ┌──▼──┐  ┌──▼──┐  ┌──▼──┐
│ Go  │ │ Go  │  │ Go  │  │ Go  │
│App 1│ │App 2│  │App 3│  │App N│
└──┬──┘ └──┬──┘  └──┬──┘  └──┬──┘
   │       │         │         │
   └───┬───┴─────────┴─────────┘
       │
   ┌───▼───┐     ┌──────────┐
   │ Redis │     │PostgreSQL│
   │Pub/Sub│     │  (Main)  │
   └───────┘     └──────────┘
```

## Implementation Phases

### Phase 1: MVP
- Basic CRUD operations
- Single-instance WebSocket server
- PostgreSQL database
- Simple React frontend

### Phase 2: Real-Time
- WebSocket hub implementation
- Real-time update broadcasting
- Optimistic UI updates

### Phase 3: Scalability
- Redis message broker integration
- Multi-instance support
- Connection management

### Phase 4: Advanced Features
- Conflict resolution
- Offline support
- Performance optimizations

## Technology Stack Summary

**Backend:**
- Go 1.21+
- `gorilla/websocket` or `nhooyr.io/websocket`
- `gorm.io/gorm` or `database/sql` for database
- `github.com/redis/go-redis` for Redis
- `github.com/gin-gonic/gin` or `net/http` for HTTP

**Frontend:**
- React 18+ with TypeScript
- TanStack Query for data fetching
- WebSocket API or socket.io-client
- Tailwind CSS for styling

**Infrastructure:**
- PostgreSQL 14+
- Redis 7+
- Docker for containerization (optional)

## Next Steps

1. Set up project structure
2. Initialize Go backend with basic HTTP server
3. Set up React frontend with Vite or Create React App
4. Implement database schema and migrations
5. Build WebSocket hub
6. Create React components with real-time updates
7. Add Redis for scaling
8. Implement conflict resolution
