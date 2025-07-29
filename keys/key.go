package keys

import "fmt"

// GetGlobalLeaderboardKey  Global leaderboard (all projects combined) ---> key: leaderboard:global:{timeframe}:{period}
func GetGlobalLeaderboardKey(timeframe string, period string) string {
	return fmt.Sprintf("leaderboard:global:%s:%s", timeframe, period)
}

// GetPerProjectLeaderboardKey Per-Project leaderboard ---> key: leaderboard:{project_id}:{timeframe}:{period}
func GetPerProjectLeaderboardKey(project string, timeframe string, period string) string {
	return fmt.Sprintf("leaderboard:%s:%s:%s", project, timeframe, period)
}

// GetHashKey key: leaderboard:global:{timeframe}:{period}
func GetHashKey(project string, timeframe string, period string, userID string, contribID string) string {
	return fmt.Sprintf("user:%s:%s:%s:%s:%s", project, timeframe, period, userID, contribID)
}
