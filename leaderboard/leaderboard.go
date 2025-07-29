package leaderboard

import (
	"context"
	"fmt"
	"golang.project/go-fundamentals/leaderboard-poc-code/adapter"
	"golang.project/go-fundamentals/leaderboard-poc-code/keys"
	"golang.project/go-fundamentals/leaderboard-poc-code/timettl"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/redis/go-redis/v9"
)

type leaderboard struct {
	table                  table.Model
	redisClient            *redis.Client
	projects               []string
	timeframes             []string
	projectIndex           int
	timeIndex              int
	leaderboardTimeRefresh time.Duration
}

type tickMsg time.Time

func NewLeaderboardModel(
	redisClient *adapter.Redis,
	columns []table.Column,
	leaderboardTimeRefresh time.Duration,
) *leaderboard {
	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
	)

	t.SetStyles(table.DefaultStyles())

	return &leaderboard{
		table:                  t,
		redisClient:            redisClient.GetClient(),
		projects:               []string{"global", "project1", "project2", "project3", "project4", "project5"},
		timeframes:             []string{"week", "month", "year"},
		projectIndex:           0,
		timeIndex:              0,
		leaderboardTimeRefresh: leaderboardTimeRefresh,
	}
}

func (l leaderboard) RunLeaderboard() {
	p := tea.NewProgram(l)
	if err := p.Start(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

func tick(timeRefresh time.Duration) tea.Cmd {
	return tea.Tick(timeRefresh, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (l leaderboard) Init() tea.Cmd {
	return tea.Batch(tick(l.leaderboardTimeRefresh), l.loadData())
}

func (l leaderboard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		return l, tea.Batch(tick(l.leaderboardTimeRefresh), l.loadData())
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return l, tea.Quit
		case "left":
			if l.projectIndex > 0 {
				l.projectIndex--
			}
			return l, l.loadData()
		case "right":
			if l.projectIndex < len(l.projects)-1 {
				l.projectIndex++
			}
			return l, l.loadData()
		case "up":
			if l.timeIndex > 0 {
				l.timeIndex--
			}
			return l, l.loadData()
		case "down":
			if l.timeIndex < len(l.timeframes)-1 {
				l.timeIndex++
			}
			return l, l.loadData()
		}
	case []table.Row:
		l.table.SetRows(msg)
	}

	var cmd tea.Cmd
	l.table, cmd = l.table.Update(msg)
	return l, cmd
}

func (l leaderboard) View() string {
	project := l.projects[l.projectIndex]
	timeframe := l.timeframes[l.timeIndex]

	header := lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("Leaderboard - %s / %s (Use ← → for project,"+
		" ↑ ↓ for timeframe, q to quit, (pgdn, pgup) for scrole leaderboard)", project, timeframe))
	return header + "\n" + l.table.View()
}

func (l leaderboard) loadData() tea.Cmd {
	ctx := context.Background()
	project := l.projects[l.projectIndex]
	timeframe := l.timeframes[l.timeIndex]

	var period string
	switch timeframe {
	case "week":
		period = timettl.GetWeek()
	case "month":
		period = timettl.GetMonth()
	case "year":
		period = timettl.GetYear()
	}

	var redisKey string
	if project == "global" {
		redisKey = keys.GetGlobalLeaderboardKey(timeframe, period)
	} else {
		redisKey = keys.GetPerProjectLeaderboardKey(project, timeframe, period)
	}

	return func() tea.Msg {
		data, err := l.redisClient.ZRevRangeWithScores(ctx, redisKey, 0, -1).Result()
		if err != nil {
			return []table.Row{{"Error", "-", "-"}}
		}

		var rows []table.Row
		for i, entry := range data {
			rows = append(rows, table.Row{
				fmt.Sprintf("%d", i+1),
				fmt.Sprintf("%v", entry.Member),
				fmt.Sprintf("%.0f", entry.Score),
			})
		}

		return rows
	}
}
