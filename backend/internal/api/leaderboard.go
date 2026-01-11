package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type LeaderboardHandler struct {
	DB *sql.DB
}

func NewLeaderboardHandler(db *sql.DB) *LeaderboardHandler {
	return &LeaderboardHandler{DB: db}
}

type UserStats struct {
	Username string `json:"username"`
	Wins     int    `json:"wins"`
}

func (h *LeaderboardHandler) GetLeaderboard(c *gin.Context) {
	rows, err := h.DB.Query("SELECT username, wins FROM users ORDER BY wins DESC LIMIT 10")
	if err != nil {
		// Table might not exist yet, return empty
		c.JSON(http.StatusOK, []UserStats{})
		return
	}
	defer rows.Close()

	var leaderboard []UserStats
	for rows.Next() {
		var u UserStats
		if err := rows.Scan(&u.Username, &u.Wins); err != nil {
			continue
		}
		leaderboard = append(leaderboard, u)
	}

	c.JSON(http.StatusOK, leaderboard)
}
