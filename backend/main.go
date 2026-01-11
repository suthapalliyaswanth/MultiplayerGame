package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"multiplayer-game/internal/analytics"
	"multiplayer-game/internal/api"
	"multiplayer-game/internal/manager"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	// DB Connection
	// DB Connection
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://user:password@localhost/gamedb?sslmode=disable"
	}
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	// Create table if not exists
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (username TEXT PRIMARY KEY, wins INT DEFAULT 0)`)
	if err != nil {
		log.Println("Error creating table (might wait for Postgres):", err)
	}

	// Analytics (Optional)
	kafkaBroker := os.Getenv("KAFKA_BROKER")
	var producer *analytics.Producer
	if kafkaBroker != "" {
		producer = analytics.NewProducer([]string{kafkaBroker}, "game_ended")
		log.Println("Analytics enabled with broker:", kafkaBroker)
	} else {
		log.Println("Analytics disabled (KAFKA_BROKER not set)")
	}

	// Game Manager
	manager := manager.NewManager(producer)
	go manager.Run()

	r := gin.Default()

	// Leaderboard API
	lbHandler := api.NewLeaderboardHandler(db)
	r.GET("/api/leaderboard", lbHandler.GetLeaderboard)

	// WebSocket Endpoint
	r.GET("/ws", func(c *gin.Context) {
		serveWs(manager, c.Writer, c.Request)
	})

	log.Println("Server starting on :8081")
	if err := r.Run(":8081"); err != nil {
		log.Fatal("Server failed to start: ", err)
	}
}

func serveWs(m *manager.Manager, w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		username = fmt.Sprintf("User-%d", time.Now().UnixNano())
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &manager.Client{
		ID:      username,
		Conn:    conn,
		Send:    make(chan []byte, 256),
		Manager: m,
	}

	m.Register <- client

	// Start Write Pump
	go func() {
		defer conn.Close()
		for msg := range client.Send {
			conn.WriteMessage(websocket.TextMessage, msg)
		}
	}()

	// Read Pump
	defer func() {
		m.Unregister <- client
		conn.Close()
	}()

	for {

		// Parse message
		var msgMap map[string]interface{}
		if err := conn.ReadJSON(&msgMap); err != nil {
			log.Println("Error reading json:", err)
			break
		}

		if typeStr, ok := msgMap["type"].(string); ok && typeStr == "move" {
			if colFloat, ok := msgMap["col"].(float64); ok {
				m.HandleMove(client, int(colFloat))
			}
		}
	}
}
