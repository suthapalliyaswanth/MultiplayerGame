package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type GameEvent struct {
	Type      string    `json:"type"`
	GameID    string    `json:"game_id"`
	Winner    string    `json:"winner"`
	Duration  float64   `json:"duration_seconds"`
	Timestamp time.Time `json:"timestamp"`
}

var (
	totalGames int
	p1Wins     int
	p2Wins     int
	mu         sync.Mutex
)

func eventHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var event GameEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	totalGames++
	if event.Winner == "P1" {
		p1Wins++
	} else if event.Winner == "P2" {
		p2Wins++
	}

	log.Printf(
		"Event received | Game=%s Winner=%s | Total=%d P1=%d P2=%d",
		event.GameID, event.Winner, totalGames, p1Wins, p2Wins,
	)

	w.WriteHeader(http.StatusOK)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Analytics service running"))
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/events", eventHandler)
	http.HandleFunc("/health", healthHandler)

	log.Println("Analytics Service Started on port", port)

	// 🔴 THIS LINE IS REQUIRED — without it Render will kill the app
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
