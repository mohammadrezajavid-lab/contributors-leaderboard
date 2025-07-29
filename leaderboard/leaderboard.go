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

func (m leaderboard) RunLeaderboard() {
	p := tea.NewProgram(m)
	if err := p.Start(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

func tick(timeRefresh time.Duration) tea.Cmd {
	return tea.Tick(timeRefresh, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m leaderboard) Init() tea.Cmd {
	return tea.Batch(tick(m.leaderboardTimeRefresh), m.loadData())
}

func (m leaderboard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		return m, tea.Batch(tick(m.leaderboardTimeRefresh), m.loadData())
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "left":
			if m.projectIndex > 0 {
				m.projectIndex--
			}
			return m, m.loadData()
		case "right":
			if m.projectIndex < len(m.projects)-1 {
				m.projectIndex++
			}
			return m, m.loadData()
		case "up":
			if m.timeIndex > 0 {
				m.timeIndex--
			}
			return m, m.loadData()
		case "down":
			if m.timeIndex < len(m.timeframes)-1 {
				m.timeIndex++
			}
			return m, m.loadData()
		}
	case []table.Row:
		m.table.SetRows(msg)
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m leaderboard) View() string {
	project := m.projects[m.projectIndex]
	timeframe := m.timeframes[m.timeIndex]

	header := lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("Leaderboard - %s / %s (Use ← → for project, ↑ ↓ for timeframe, q to quit)", project, timeframe))
	return header + "\n" + m.table.View()
}

func (m leaderboard) loadData() tea.Cmd {
	ctx := context.Background()
	project := m.projects[m.projectIndex]
	timeframe := m.timeframes[m.timeIndex]

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
		data, err := m.redisClient.ZRevRangeWithScores(ctx, redisKey, 0, -1).Result()
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
