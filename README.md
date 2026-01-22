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

- ✅ Real-time task updates across all clients
- ✅ Multi-user collaboration
- ✅ Optimistic UI updates
- ✅ Horizontal scaling support
- ✅ Conflict resolution
- ✅ Offline support (planned)

## Development

See [ARCHITECTURE.md](./ARCHITECTURE.md) for detailed implementation guidance.
