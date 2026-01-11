package manager

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"multiplayer-game/internal/analytics"
	"multiplayer-game/internal/bot"
	"multiplayer-game/internal/game"

	"github.com/gorilla/websocket"
)

type Client struct {
	ID      string
	Conn    *websocket.Conn
	GameID  string
	Send    chan []byte
	Manager *Manager
}

type Manager struct {
	Games      map[string]*game.Game
	Clients    map[string]*Client
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan []byte
	Waiting    chan *Client
	mu         sync.Mutex
	Producer   *analytics.Producer
}

func NewManager(producer *analytics.Producer) *Manager {
	return &Manager{
		Games:      make(map[string]*game.Game),
		Clients:    make(map[string]*Client),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan []byte),
		Waiting:    make(chan *Client, 100),
		Producer:   producer,
	}
}

func (m *Manager) Run() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case client := <-m.Register:
			m.mu.Lock()
			m.Clients[client.ID] = client
			m.mu.Unlock()

			// Matchmaking logic
			select {
			case opponent := <-m.Waiting:
				// Match found!
				m.mu.Lock()
				// Check if opponent is still connected (prevent panic on closed channel)
				if _, ok := m.Clients[opponent.ID]; ok {
					m.CreateGame(client, opponent)
					m.mu.Unlock()
				} else {
					m.mu.Unlock()
					// Opponent disconnected while waiting. Treat current client as waiting.
					m.Waiting <- client
					go m.WaitForOpponentOrBot(client)
				}
			default:
				// No opponent, push to waiting
				m.Waiting <- client
				// Start timeout for Bot
				go m.WaitForOpponentOrBot(client)
			}

		case client := <-m.Unregister:
			m.mu.Lock()
			if _, ok := m.Clients[client.ID]; ok {
				delete(m.Clients, client.ID)
				close(client.Send)
				// Handle disconnect
				if client.GameID != "" {
					if g, exists := m.Games[client.GameID]; exists {
						// Notify opponent
						// Simple forfeit for now
						// In real app, start 30s timer
						m.EndGame(g, "Opponent Disconnected")
					}
				}
			}
			m.mu.Unlock()
		}
	}
}

func (m *Manager) CreateGame(p1, p2 *Client) {
	gameID := fmt.Sprintf("game-%d", time.Now().UnixNano())
	newGame := game.NewGame(gameID, p1.ID)
	newGame.Join(p2.ID)

	m.Games[gameID] = newGame

	p1.GameID = gameID
	p2.GameID = gameID

	startMsgP1, _ := json.Marshal(map[string]string{
		"type": "start", "gameID": gameID, "opponent": p2.ID, "turn": "P1", "you": "P1",
	})
	p1.Send <- startMsgP1

	startMsgP2, _ := json.Marshal(map[string]string{
		"type": "start", "gameID": gameID, "opponent": p1.ID, "turn": "P1", "you": "P2",
	})
	p2.Send <- startMsgP2
}

func (m *Manager) WaitForOpponentOrBot(client *Client) {
	time.Sleep(10 * time.Second)
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if still waiting (waiting channel might have been drained, but check GameID)
	// Also check if client is still connected!
	if _, connected := m.Clients[client.ID]; !connected {
		return
	}

	if client.GameID == "" {
		// Remove from waiting channel??
		// Actually, we need to carefully remove from channel or just check if they are still in 'Clients' and have no GameID
		// For simplicity, we assume if they are still connected and have no game, we start bot.

		// Create Bot Game
		botClient := &Client{ID: "Bot", Send: make(chan []byte, 100)} // Fake client
		// Drain bot channel to prevent blocking
		go func() {
			for range botClient.Send {
			}
		}()
		m.CreateGame(client, botClient)

		// Start Bot Logic Loop
		go m.RunBotLoop(client.GameID)
	}
}

func (m *Manager) RunBotLoop(gameID string) {
	// Simple polling or event driven?
	// Note: In real app, this should be event driven.
	// For now, we rely on the player move triggering the bot move.
}

func (m *Manager) HandleMove(client *Client, col int) {
	m.mu.Lock()
	g, ok := m.Games[client.GameID]
	m.mu.Unlock()

	if !ok {
		return
	}

	playerCode := game.Player1
	if client.ID == g.Player2ID {
		playerCode = game.Player2
	} else if client.ID != g.Player1ID {
		return // Spectator?
	}

	r, c, err := g.DropDisc(col, playerCode)
	if err != nil {
		// error
		return
	}

	// Broadcast move
	updateMsg, _ := json.Marshal(map[string]interface{}{
		"type": "update", "row": r, "col": c, "player": playerCode, "turn": g.Turn, "winner": g.Winner,
	})

	m.BroadcastToGame(g, updateMsg)

	if g.State == "finished" {
		m.EndGame(g, string(g.Winner))
		return
	}

	// Bot Turn?
	if g.Player2ID == "Bot" && g.Turn == game.Player2 {
		go func() {
			time.Sleep(500 * time.Millisecond) // Thinking time
			botLogic := bot.NewBot()
			botCol := botLogic.GetBestMove(g)

			// Apply Bot Move (Simulate Client)
			m.mu.Lock()
			r, c, err := g.DropDisc(botCol, game.Player2)
			m.mu.Unlock()

			if err == nil {
				botMsg, _ := json.Marshal(map[string]interface{}{
					"type": "update", "row": r, "col": c, "player": "P2", "turn": g.Turn, "winner": g.Winner,
				})
				m.BroadcastToGame(g, botMsg)

				if g.State == "finished" {
					m.EndGame(g, string(g.Winner))
				}
			}
		}()
	}
}

func (m *Manager) BroadcastToGame(g *game.Game, msg []byte) {
	if p1, ok := m.Clients[g.Player1ID]; ok {
		p1.Send <- msg
	}
	if p2, ok := m.Clients[g.Player2ID]; ok {
		p2.Send <- msg
	}
}

func (m *Manager) EndGame(g *game.Game, winner string) {
	// Send to Kafka
	if m.Producer != nil {
		m.Producer.SendGameEnded(g.ID, winner, 0) // Duration 0 for now
	}

	// Determine username to update wins
	// Note: We need a DB to map ID to Username. For now assume ID is username.
	if winner == "P1" {
		// Update P1 wins
	} else if winner == "P2" {
		// Update P2 wins
	}

	// Clean up game after delay?
	// delete(m.Games, g.ID)
}
