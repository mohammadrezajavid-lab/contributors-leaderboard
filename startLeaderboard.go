package main

import (
	"github.com/charmbracelet/bubbles/table"
	"golang.project/go-fundamentals/leaderboard-poc-code/adapter"
	"golang.project/go-fundamentals/leaderboard-poc-code/leaderboard"
	"time"
)

func main() {

	redisAdapter := adapter.NewRedisAdapter(adapter.Config{
		Network:  "tcp",
		Host:     "127.0.0.1",
		Port:     6379,
		Password: "password1999",
		DB:       0,
	})

	columns := []table.Column{
		{Title: "Rank", Width: 6},
		{Title: "User", Width: 10},
		{Title: "Score", Width: 8},
	}

	leaderboardTimeRefresh := 2 * time.Second

	leaderboard := leaderboard.NewLeaderboardModel(redisAdapter, columns, leaderboardTimeRefresh)
	leaderboard.RunLeaderboard()

}
