## Real-Time Leaderboard with Redis & RediSearch

This guide demonstrates how to build a **real-time, filterable leaderboard** using:

- **Redis** with `ZSET` and `HASH` data structures
- **RediSearch** for advanced queries and filtering

---

## Redis Stack Setup (with RediSearch)

RediSearch is not included in the default Redis image. You must use:

- [`redis/redis-stack`](https://hub.docker.com/r/redis/redis-stack) (**Recommended**)
- [`redislabs/redisearch`](https://hub.docker.com/r/redislabs/redisearch) (Deprecated)

### ➔ Start Redis with Docker Compose

```bash
  docker compose up -d
```

Access Redis UI at: [http://127.0.0.1:8001](http://127.0.0.1:8001)

- Username: _(leave blank)_
- Password: `password1999`

---

## Running the Project

### 1. Start the TUI (Leaderboard UI)

This should be launched **before inserting data** to observe leaderboard changes in real time.

```bash
  go run startLeaderboard.go
```

You will see the leaderboard interface in your terminal. Use the following keys:

- `←` / `→` to switch between projects
- `↑` / `↓` to change timeframes (`week`, `month`, `year`)
- `PgUp` / `PgDn` to scroll leaderboard vertically
- `q` to quit

### 2. Insert Sample Data into Redis

Once the leaderboard UI is running, insert the simulated data:

```bash
  go run insert.go
```

This generates:

- 100 users
- 5 projects
- 100,000 contributions across the last year

---

## System Design Overview

### Why use ZSET, HASH, and RediSearch?

For displaying the necessary leaderboard data, including global rankings over the last week, month, and year, **ZSET**
is preferred for real-time ranking. This allows us to efficiently maintain a leaderboard across different timeframes.

For more complex queries and detailed reporting or analytics, **HASH** and **RediSearch** are used. RediSearch allows us
to:

- Query specific data (e.g., which user contributed on a particular date)
- Calculate total scores for users within specific timeframes
- Filter contributions by time period
- Aggregate and sort data more efficiently

For example, using RediSearch, we can answer questions like:

1. Which user contributed on a specific date?
2. What is the total score for a user on a project within a specific date range?
3. How many contributions were made in the past week?
4. What is the average score for users on project3?

In summary:

- **ZSET** is used to display global and per-project leaderboard rankings.
- **RediSearch** is used for detailed user activity insights, project analysis, and team management.

---

## ZSET Key Structure

Each contribution updates **6 leaderboard keys**:

### Global Leaderboard Keys

```
leaderboard:global:{timeframe}:{period}
```

Example:

```
leaderboard:global:week:2025-W31
```

### Per-Project Leaderboard Keys

```
leaderboard:{project}:{timeframe}:{period}
```

Example:

```
leaderboard:project2:month:2025-07
```

---

## HASH Key Structure for Contributions

Each contribution is stored in 3 `HASH` records (week/month/year):

```
user:{project}:{timeframe}:{period}:{user_id}:{contrib_id}
```

Example:

```
user:project1:year:2025:user9:contrib000001
```

Each hash contains:

- `user_id`
- `project`
- `score`
- `timeframe`, `period`
- `contribute_id`

---

## Create RediSearch Index

```bash
  FT.CREATE contrib_idx ON HASH PREFIX 1 "user:" SCHEMA \
    project TAG SEPARATOR , \
    user_id TAG SEPARATOR , \
    timeframe TAG \
    period TEXT \
    score NUMERIC SORTABLE
```

---

## Query Examples

### 1. Get Global Rankings (ZSET)

```bash
  ZREVRANGE leaderboard:global:month:2025-07 0 -1 WITHSCORES
```

### 2. Get Project Rankings

```bash
  ZREVRANGE leaderboard:project2:year:2025 0 -1 WITHSCORES
```

### 3. Get Rank or Score of a User

```bash
  ZREVRANK leaderboard:global:week:2025-W31 user23
  ZSCORE leaderboard:project3:month:2025-07 user23
```

---

## RediSearch Aggregations

### 1. Total Contributions Last Week

```bash
  FT.AGGREGATE contrib_idx "@timeframe:{week}" GROUPBY 0 REDUCE COUNT 0 AS total_contributions
```

### 2. Average Score in Project3

```bash
  FT.AGGREGATE contrib_idx "@project:{project3}" GROUPBY 0 REDUCE AVG 1 @score AS avg_score
```

### 3. Top Projects by Total Score

```bash
  FT.AGGREGATE contrib_idx "*" \
    GROUPBY 1 @project \
    REDUCE SUM 1 @score AS total_score \
    SORTBY 2 @total_score DESC
```

---

This design ensures your leaderboard system is **real-time**, **filterable**, and **scalable** across teams and
timeframes.

