# 4 in a Row - Multiplayer Game

A real-time multiplayer 4-in-a-row game built with **GoLang**, **React**, **Postgres**, and **Kafka**.

## Features
- **Real-Time Gameplay**: Powered by WebSockets.
- **Matchmaking**: Auto-pairs players or falls back to a Bot after 10s.
- **Competitive Bot**: Basic AI blocking/winning logic.
- **Leaderboard**: Tracks top winners using Postgres.
- **Analytics**: Decoupled game event tracking using Kafka.

## Tech Stack
- **Backend**: Go (Gin, Gorilla WebSocket)
- **Frontend**: React (Vite)
- **Database**: PostgreSQL
- **Message Queue**: Apache Kafka

## Getting Started

### 1. Infrastructure (Docker)
Start the required services:
```bash
docker-compose up -d
```

### 2. Backend
```bash
cd backend
go mod tidy
go run main.go
```

### 3. Frontend
```bash
cd frontend
npm install
npm run dev
```

### 4. Analytics Consumer
```bash
cd analytics
go mod tidy
go run main.go
```

## How to Play
1. Enter your username.
2. Click "Join Game".
3. Wait for an opponent or play against the Bot.
4. Click columns to drop your disc.
5. Connect 4 to win!
