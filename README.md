# Collaborative Task Management System

A real-time collaborative task management system built with React and Go, designed for scalability without relying on managed real-time databases.

## Architecture

See [ARCHITECTURE.md](./ARCHITECTURE.md) for detailed architecture recommendations and design decisions.

## Tech Stack

### Backend
- **Go** - High-performance backend server
- **WebSockets** - Real-time bidirectional communication
- **PostgreSQL** - Primary data store
- **Redis** - Message broker and caching layer

### Frontend
- **React 18+** with TypeScript
- **TanStack Query** - Server state management
- **WebSocket API** - Real-time updates
- **Tailwind CSS** - Styling

## Project Structure

```
.
├── backend/          # Go backend application
│   ├── cmd/
│   ├── internal/
│   ├── pkg/
│   └── migrations/
├── frontend/         # React frontend application
│   ├── src/
│   ├── public/
│   └── package.json
├── docker-compose.yml # Local development setup
└── README.md
```

## Quick Start

### Prerequisites
- Go 1.21+
- Node.js 18+
- Docker and Docker Compose (for local development)

### 1. Start Infrastructure

```bash
# Start PostgreSQL and Redis
make docker-up
# or
docker-compose up -d
```

### 2. Install Dependencies

```bash
make setup
# or manually:
cd backend && go mod download
cd frontend && npm install
```

### 3. Configure Environment

```bash
# Backend
cd backend
cp .env.example .env
# Edit .env if needed (defaults work with docker-compose)
```

### 4. Run the Application

**Terminal 1 - Backend:**
```bash
make backend
# or
cd backend && go run cmd/server/main.go
```

**Terminal 2 - Frontend:**
```bash
make frontend
# or
cd frontend && npm run dev
```

### 5. Open in Browser

- Frontend: http://localhost:3000
- Backend API: http://localhost:8080
- Health Check: http://localhost:8080/health

For detailed setup instructions, see [SETUP.md](./SETUP.md).

## Features

### Core Requirements
- ✅ **Multiple Projects**: Users can create and manage multiple projects
- ✅ **Task Management**: Add, update, and delete tasks within projects
- ✅ **Task Dependencies**: Support for task dependencies with validation
- ✅ **Status Transitions**: Validated status transitions (todo → in_progress → done, blocked states)
- ✅ **Comment Threads**: Real-time comment threads on tasks with nested replies
- ✅ **Real-Time Updates**: Changes visible to all clients in near real-time via WebSockets
- ✅ **Data Consistency**: Optimistic locking and conflict resolution across clients

### Technical Features
- ✅ **Efficient Updates**: Delta updates (only changed fields) to handle large project payloads (2MB+)
- ✅ **Horizontal Scaling**: Redis pub/sub for multi-instance WebSocket support
- ✅ **No Managed Real-Time DB**: Built with PostgreSQL + Redis, no Firebase/Supabase
- ✅ **Optimistic UI Updates**: Immediate UI feedback with rollback on error

## Development

See [ARCHITECTURE.md](./ARCHITECTURE.md) for detailed implementation guidance.
