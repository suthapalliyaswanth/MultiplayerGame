package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type GameEvent struct {
	Type      string    `json:"type"`
	GameID    string    `json:"game_id"`
	Winner    string    `json:"winner"`
	Duration  float64   `json:"duration_seconds"`
	Timestamp time.Time `json:"timestamp"`
}

func main() {
	topic := "game_ended"
	partition := 0

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{"localhost:9092"},
		Topic:     topic,
		Partition: partition,
		MinBytes:  10e3, // 10KB
		MaxBytes:  10e6, // 10MB
	})

	fmt.Println("Analytics Service Started. Listening for game events...")

	var totalGames int
	var p1Wins int
	var p2Wins int

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			break
		}
		
		var event GameEvent
		if err := json.Unmarshal(m.Value, &event); err != nil {
			log.Printf("Error unmarshalling event: %v", err)
			continue
		}

		totalGames++
		if event.Winner == "P1" {
			p1Wins++
		} else if event.Winner == "P2" {
			p2Wins++
		}

		fmt.Printf("Event Received: Game %s ended. Winner: %s\n", event.GameID, event.Winner)
		fmt.Printf("--- Live Stats ---\n")
		fmt.Printf("Total Games: %d\n", totalGames)
		fmt.Printf("P1 Wins: %d (%.1f%%)\n", p1Wins, float64(p1Wins)/float64(totalGames)*100)
		fmt.Printf("P2 Wins: %d (%.1f%%)\n", p2Wins, float64(p2Wins)/float64(totalGames)*100)
		fmt.Printf("------------------\n")
	}

	if err := r.Close(); err != nil {
		log.Fatal("failed to close reader:", err)
	}
}
