package main

import (
	"context"
	"fmt"
	"golang.project/go-fundamentals/leaderboard-poc-code/keys"
	"golang.project/go-fundamentals/leaderboard-poc-code/timettl"
	"math/rand"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var projects = []string{"project1", "project2", "project3", "project4", "project5"}
var timeframes = []string{"year", "month", "week"}
var rdb = redis.NewClient(&redis.Options{
	Network:  "tcp",
	Addr:     "localhost:6379",
	Password: "password1999",
	DB:       0,
})

func main() {
	InsertData()
}

func InsertData() {
	start := time.Now()
	wg := sync.WaitGroup{}
	contribCount := 100_000
	workers := 10
	batch := contribCount / workers

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(offset int) {
			defer wg.Done()
			for i := offset; i < offset+batch; i++ {
				userID := fmt.Sprintf("user%d", i%100) // 100 unique users
				contribID := fmt.Sprintf("contrib%06d", i)
				project := projects[rand.Intn(len(projects))]
				score := rand.Intn(10) + 1

				addGlobalLeaderboard(userID, score)
				addPerProjectLeaderboard(project, userID, score)

				setContribution(project, userID, contribID, score)

				time.Sleep(90 * time.Millisecond)
			}
		}(w * batch)
	}

	wg.Wait()
	fmt.Printf("Inserted %d contributions in %s\\n", contribCount, time.Since(start))
}

// ZINCRBY leaderboard:global:{timeframe}:{period} score user_id
func addGlobalLeaderboard(userID string, score int) {
	for _, timeframe := range timeframes {
		var period string
		var ttl time.Duration

		switch timeframe {
		case "year":
			period = timettl.GetYear()
			ttl = timettl.GetTTLToEndOfYear()
		case "month":
			period = timettl.GetMonth()
			ttl = timettl.GetTTLToEndOfMonth()
		case "week":
			period = timettl.GetWeek()
			ttl = timettl.GetTTLToEndOfWeek()
		}

		key := keys.GetGlobalLeaderboardKey(timeframe, period)

		exists, _ := rdb.Exists(ctx, key).Result()

		rdb.ZIncrBy(ctx, key, float64(score), userID)

		if exists == 0 {
			rdb.Expire(ctx, key, ttl)
		}

		fmt.Println(key)
	}
}

// ZINCRBY leaderboard:{project_id}:{timeframe}:{period}
func addPerProjectLeaderboard(project string, userID string, score int) {
	for _, timeframe := range timeframes {
		var period string
		var ttl time.Duration

		switch timeframe {
		case "year":
			period = timettl.GetYear()
			ttl = timettl.GetTTLToEndOfYear()
		case "month":
			period = timettl.GetMonth()
			ttl = timettl.GetTTLToEndOfMonth()
		case "week":
			period = timettl.GetWeek()
			ttl = timettl.GetTTLToEndOfWeek()
		}

		key := keys.GetPerProjectLeaderboardKey(project, timeframe, period)

		exists, _ := rdb.Exists(ctx, key).Result()

		rdb.ZIncrBy(ctx, key, float64(score), userID)

		if exists == 0 {
			rdb.Expire(ctx, key, ttl)
		}

		fmt.Println(key)
	}
}

// HSET user:{project}:{timeframe}:{period}:{user_id}:{contrib_id}
func setContribution(project string, userID string, contribID string, score int) {
	for _, timeframe := range timeframes {
		var period string

		var ttl time.Duration

		switch timeframe {
		case "year":
			period = timettl.GetYear()
			ttl = timettl.GetTTLToEndOfYear()
		case "month":
			period = timettl.GetMonth()
			ttl = timettl.GetTTLToEndOfMonth()
		case "week":
			period = timettl.GetWeek()
			ttl = timettl.GetTTLToEndOfWeek()
		}

		values := map[string]interface{}{
			"name":          "Team Member " + userID,
			"project":       project,
			"timeframe":     timeframe,
			"period":        period,
			"user_id":       userID,
			"score":         score,
			"contribute_id": contribID,
		}

		key := keys.GetHashKey(project, timeframe, period, userID, contribID)

		exists, _ := rdb.Exists(ctx, key).Result()

		rdb.HSet(ctx, key, values)

		if exists == 0 {
			rdb.Expire(ctx, key, ttl)
		}
	}
}
