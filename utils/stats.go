package utils

import (
	"sync"
	"time"
)

var blockStats = struct {
	mu          sync.RWMutex
	totalBlocks int64
	perUser     map[string]int64
	startTime   time.Time
}{
	perUser:   make(map[string]int64),
	startTime: time.Now(),
}

type StatsSnapshot struct {
	TotalBlocks int64            `json:"total_blocks"`
	PerUser     map[string]int64 `json:"per_user"`
	StartedAt   time.Time        `json:"started_at"`
}

func RecordBlock(username string) {
	cleanUsername := processUsernameForWebhook(username)

	blockStats.mu.Lock()
	defer blockStats.mu.Unlock()

	blockStats.totalBlocks++
	blockStats.perUser[cleanUsername]++
}

func GetStats() StatsSnapshot {
	blockStats.mu.RLock()
	defer blockStats.mu.RUnlock()

	perUser := make(map[string]int64, len(blockStats.perUser))
	for user, count := range blockStats.perUser {
		perUser[user] = count
	}

	return StatsSnapshot{
		TotalBlocks: blockStats.totalBlocks,
		PerUser:     perUser,
		StartedAt:   blockStats.startTime,
	}
}
