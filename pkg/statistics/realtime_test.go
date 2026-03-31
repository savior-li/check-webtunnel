package statistics

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"tor-bridge-collector/pkg/bridge"
	"tor-bridge-collector/pkg/database"
)

func TestGetRealtimeStats_Empty(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	stats, err := GetRealtimeStats(db)

	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, 0, stats.TotalBridges)
	assert.Equal(t, 0, stats.AvailableBridges)
	assert.Equal(t, 0, stats.UnavailableBridges)
	assert.Equal(t, 0, stats.UnknownBridges)
}

func TestGetRealtimeStats_WithBridges(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := database.NewBridgeRepository(db)

	testBridges := []bridge.Bridge{
		{
			Hash:         "hash1",
			Transport:    "webtunnel",
			Address:      "192.168.1.1",
			Port:         443,
			IsAvailable:  1,
			ResponseTime: 100,
		},
		{
			Hash:        "hash2",
			Transport:   "webtunnel",
			Address:     "192.168.1.2",
			Port:        443,
			IsAvailable: 0,
		},
		{
			Hash:        "hash3",
			Transport:   "webtunnel",
			Address:     "192.168.1.3",
			Port:        443,
			IsAvailable: -1,
		},
	}

	for _, b := range testBridges {
		repo.Insert(&b)
	}

	stats, err := GetRealtimeStats(db)

	assert.NoError(t, err)
	assert.Equal(t, 3, stats.TotalBridges)
	assert.Equal(t, 1, stats.AvailableBridges)
	assert.Equal(t, 1, stats.UnavailableBridges)
	assert.Equal(t, 1, stats.UnknownBridges)
}

func TestRealtimeStats_AvailableRate(t *testing.T) {
	tests := []struct {
		name         string
		total        int
		available    int
		expectedRate float64
	}{
		{"no bridges", 0, 0, 0},
		{"all available", 10, 10, 100},
		{"none available", 10, 0, 0},
		{"half available", 10, 5, 50},
		{"mixed", 3, 1, 33.33333333333333},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := &RealtimeStats{
				TotalBridges:     tt.total,
				AvailableBridges: tt.available,
			}
			rate := stats.AvailableRate()
			assert.InDelta(t, tt.expectedRate, rate, 0.01)
		})
	}
}

func TestGetRealtimeStats_LastFetchTime(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	stats, err := GetRealtimeStats(db)
	assert.NoError(t, err)
	assert.True(t, stats.LastFetchTime.IsZero())
}

func setupTestDB(t *testing.T) (*database.DB, func()) {
	t.Helper()

	tempDir := t.TempDir()
	dbPath := tempDir + "/test.db"

	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("create test db failed: %v", err)
	}

	if err := db.InitSchema(); err != nil {
		db.Close()
		t.Fatalf("init test schema failed: %v", err)
	}

	cleanup := func() {
		db.Close()
	}

	return db, cleanup
}
