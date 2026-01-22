# Setup Guide

## Prerequisites

- Go 1.21 or higher
- Node.js 18+ and npm
- Docker and Docker Compose (for local development)
- PostgreSQL 14+ (or use Docker)
- Redis 7+ (or use Docker)

## Quick Start

### 1. Start Infrastructure Services

```bash
# Start PostgreSQL and Redis using Docker Compose
docker-compose up -d

# Verify services are running
docker-compose ps
```

### 2. Backend Setup

```bash
cd backend

# Copy environment file
cp .env.example .env

# Edit .env if needed (defaults should work with docker-compose)

# Install Go dependencies
go mod download

# Run the server
go run cmd/server/main.go
```

The backend will start on `http://localhost:8080`

### 3. Frontend Setup

```bash
cd frontend

# Install dependencies
npm install

# Start development server
npm run dev
```

The frontend will start on `http://localhost:3000`

## Database Setup

The application uses GORM's AutoMigrate feature to create tables automatically. On first run, the following tables will be created:

- `tasks` - Main task table
- `task_updates` - Audit log of task changes

If you prefer manual migrations, you can use tools like `golang-migrate` or `goose`.

## Testing the Application

1. Open `http://localhost:3000` in your browser
2. Create a new task using the input field
3. Open another browser window/tab to the same URL
4. Make changes in one window - they should appear in real-time in the other window!

## Project Structure

```
.
├── backend/              # Go backend
│   ├── cmd/
│   │   └── server/      # Application entry point
│   ├── internal/
│   │   ├── api/         # HTTP handlers and routes
│   │   ├── config/      # Configuration management
│   │   ├── database/    # Database models and connection
│   │   ├── repository/  # Data access layer
│   │   └── websocket/   # WebSocket hub and clients
│   └── go.mod
├── frontend/            # React frontend
│   ├── src/
│   │   ├── api/         # API client functions
│   │   ├── hooks/       # React hooks (useTasks, useWebSocket)
│   │   ├── App.tsx      # Main application component
│   │   └── types.ts     # TypeScript type definitions
│   └── package.json
└── docker-compose.yml   # Local development infrastructure
```

## Environment Variables

### Backend (.env)

```env
DATABASE_URL=postgres://postgres:postgres@localhost:5432/taskmanager?sslmode=disable
REDIS_URL=redis://localhost:6379
JWT_SECRET=your-secret-key-change-in-production
PORT=8080
```

### Frontend

Create a `.env` file in the frontend directory:

```env
VITE_API_URL=http://localhost:8080/api/v1
```

## Development Tips

### Backend

- The server auto-reloads on code changes if you use tools like `air` or `fresh`
- Install `air`: `go install github.com/cosmtrek/air@latest`
- Run with auto-reload: `air` (in backend directory)

### Frontend

- Hot module replacement is enabled by default with Vite
- Check browser console for WebSocket connection status
- Network tab shows real-time WebSocket messages

## Troubleshooting

### Database Connection Issues

- Ensure PostgreSQL is running: `docker-compose ps`
- Check connection string in `.env`
- Verify database exists: `docker exec -it taskmanager-postgres psql -U postgres -c "\l"`

### Redis Connection Issues

- Ensure Redis is running: `docker-compose ps`
- Test Redis: `docker exec -it taskmanager-redis redis-cli ping`

### WebSocket Not Connecting

- Check browser console for errors
- Verify backend is running on port 8080
- Check CORS settings in `backend/internal/api/router.go`
- Ensure workspace_id and user_id are provided in WebSocket URL

### Port Conflicts

- Backend default: 8080 (change in `.env`)
- Frontend default: 3000 (change in `vite.config.ts`)
- PostgreSQL: 5432
- Redis: 6379

## Next Steps

1. **Add Authentication**: Implement JWT-based authentication
2. **Add Workspaces**: Support multiple workspaces per user
3. **Add User Management**: User registration and login
4. **Improve UI**: Add more styling and better UX
5. **Add Tests**: Unit and integration tests
6. **Add CI/CD**: GitHub Actions or similar

## Production Deployment

1. Set proper environment variables
2. Use a production database (managed PostgreSQL)
3. Use a managed Redis service
4. Set up proper CORS origins
5. Enable HTTPS
6. Use a reverse proxy (nginx) for load balancing
7. Set up monitoring and logging
