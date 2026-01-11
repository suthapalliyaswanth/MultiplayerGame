package bot

import (
	"multiplayer-game/internal/game"
)

type Bot struct {
	ID string
}

func NewBot() *Bot {
	return &Bot{ID: "Bot"}
}

// Simple heuristic:
// 1. Can I win? -> Do it.
// 2. Will opponent win? -> Block it.
// 3. Otherwise -> Random/Center.

func (b *Bot) GetBestMove(g *game.Game) int {
	// 1. Check for winning move
	for c := 0; c < game.Cols; c++ {
		if canDrop(g, c) {
			if simulateMove(g, c, game.Player2) { // Bot is always Player2 in this context
				return c
			}
		}
	}

	// 2. Check for blocking move
	for c := 0; c < game.Cols; c++ {
		if canDrop(g, c) {
			if simulateMove(g, c, game.Player1) { // Opponent is Player1
				return c
			}
		}
	}

	// 3. Minimax / Heuristic (simplified for now to basic positional play)
	// Prefer center columns
	centerOrder := []int{3, 2, 4, 1, 5, 0, 6}
	for _, c := range centerOrder {
		if canDrop(g, c) {
			return c
		}
	}

	return 0
}

func canDrop(g *game.Game, col int) bool {
	return g.Board[0][col] == game.Empty
}

func simulateMove(g *game.Game, col int, player game.Player) bool {
	// Deep copy board or undo move?
	// For simplicity, just check the condition without mutating if possible
	// Or mutate and revert.

	// Find row
	var row int = -1
	for r := game.Rows - 1; r >= 0; r-- {
		if g.Board[r][col] == game.Empty {
			row = r
			break
		}
	}

	if row == -1 {
		return false
	}

	// Make move
	g.Board[row][col] = player
	win := g.CheckWin(row, col, player)
	// Revert
	g.Board[row][col] = game.Empty

	return win
}

// Minimax implementation (Optional - if needed for stronger bot)
func Minimax(g *game.Game, depth int, maximizingPlayer bool) int {
	// Placeholder for more complex logic
	return 0
}
