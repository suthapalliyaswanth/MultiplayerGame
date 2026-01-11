package game

import (
	"fmt"
	"sync"
)

const (
	Rows = 6
	Cols = 7
)

type Player string

const (
	Player1 Player = "P1"
	Player2 Player = "P2"
	Empty   Player = ""
)

type Game struct {
	ID        string
	Board     [Rows][Cols]Player
	Turn      Player
	State     string // "waiting", "playing", "finished"
	Winner    Player
	Player1ID string
	Player2ID string
	mu        sync.Mutex
}

func NewGame(id, p1ID string) *Game {
	return &Game{
		ID:        id,
		Turn:      Player1,
		State:     "waiting",
		Player1ID: p1ID,
	}
}

func (g *Game) Join(p2ID string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.Player2ID = p2ID
	g.State = "playing"
}

func (g *Game) DropDisc(col int, player Player) (int, int, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.State != "playing" {
		return -1, -1, fmt.Errorf("game not active")
	}
	if g.Turn != player {
		return -1, -1, fmt.Errorf("not your turn")
	}
	if col < 0 || col >= Cols {
		return -1, -1, fmt.Errorf("invalid column")
	}

	// Find the first empty row from bottom
	for r := Rows - 1; r >= 0; r-- {
		if g.Board[r][col] == Empty {
			g.Board[r][col] = player
			if g.CheckWin(r, col, player) {
				g.State = "finished"
				g.Winner = player
			} else if g.IsFull() {
				g.State = "finished"
				g.Winner = "Draw"
			} else {
				g.SwitchTurn()
			}
			return r, col, nil
		}
	}

	return -1, -1, fmt.Errorf("column full")
}

func (g *Game) SwitchTurn() {
	if g.Turn == Player1 {
		g.Turn = Player2
	} else {
		g.Turn = Player1
	}
}

func (g *Game) IsFull() bool {
	for c := 0; c < Cols; c++ {
		if g.Board[0][c] == Empty {
			return false
		}
	}
	return true
}

func (g *Game) CheckWin(row, col int, player Player) bool {
	// Directions: Horizontal, Vertical, Diagonal /, Diagonal \
	directions := [][2]int{{0, 1}, {1, 0}, {1, 1}, {1, -1}}

	for _, d := range directions {
		count := 1 // Count the current piece
		
		// Check forward
		for i := 1; i < 4; i++ {
			r, c := row+d[0]*i, col+d[1]*i
			if r >= 0 && r < Rows && c >= 0 && c < Cols && g.Board[r][c] == player {
				count++
			} else {
				break
			}
		}

		// Check backward
		for i := 1; i < 4; i++ {
			r, c := row-d[0]*i, col-d[1]*i
			if r >= 0 && r < Rows && c >= 0 && c < Cols && g.Board[r][c] == player {
				count++
			} else {
				break
			}
		}

		if count >= 4 {
			return true
		}
	}
	return false
}
