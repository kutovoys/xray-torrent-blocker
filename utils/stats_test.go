package utils

import (
	"testing"
)

func TestRecordBlockAndGetStats(t *testing.T) {
	before := GetStats()

	RecordBlock("statsuser1")
	RecordBlock("statsuser1")
	RecordBlock("statsuser2")

	stats := GetStats()

	if stats.TotalBlocks != before.TotalBlocks+3 {
		t.Errorf("Expected total blocks to increase by 3, got %d -> %d", before.TotalBlocks, stats.TotalBlocks)
	}

	if stats.PerUser["statsuser1"] != before.PerUser["statsuser1"]+2 {
		t.Errorf("Expected statsuser1 count to increase by 2, got %d", stats.PerUser["statsuser1"])
	}

	if stats.PerUser["statsuser2"] != before.PerUser["statsuser2"]+1 {
		t.Errorf("Expected statsuser2 count to increase by 1, got %d", stats.PerUser["statsuser2"])
	}

	if stats.StartedAt.IsZero() {
		t.Error("Expected StartedAt to be set")
	}
}

func TestGetStatsReturnsCopy(t *testing.T) {
	RecordBlock("copyuser")

	snapshot := GetStats()
	snapshot.PerUser["copyuser"] = 9999

	fresh := GetStats()
	if fresh.PerUser["copyuser"] == 9999 {
		t.Error("Expected GetStats to return a copy of the per-user map")
	}
}
