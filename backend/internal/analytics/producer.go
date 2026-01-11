package analytics

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	Writer *kafka.Writer
}

type GameEvent struct {
	Type      string    `json:"type"`
	GameID    string    `json:"game_id"`
	Winner    string    `json:"winner"`
	Duration  float64   `json:"duration_seconds"`
	Timestamp time.Time `json:"timestamp"`
}

func NewProducer(brokers []string, topic string) *Producer {
	if len(brokers) == 0 || brokers[0] == "" {
		return nil
	}
	return &Producer{
		Writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (p *Producer) SendGameEnded(gameID, winner string, duration float64) {
	event := GameEvent{
		Type:      "GAME_ENDED",
		GameID:    gameID,
		Winner:    winner,
		Duration:  duration,
		Timestamp: time.Now(),
	}

	payload, err := json.Marshal(event)
	if err != nil {
		log.Println("Failed to marshal event:", err)
		return
	}

	err = p.Writer.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte(gameID),
			Value: payload,
		},
	)
	if err != nil {
		log.Println("Failed to write to Kafka:", err)
	}
}
