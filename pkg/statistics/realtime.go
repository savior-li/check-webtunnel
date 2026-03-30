package statistics

import (
	"fmt"
	"time"

	"tor-bridge-collector/pkg/database"
)

type RealtimeStats struct {
	TotalBridges       int
	AvailableBridges   int
	UnavailableBridges int
	UnknownBridges     int
	AvgResponseTime    float64
	LastFetchTime      time.Time
}

func GetRealtimeStats(db *database.DB) (*RealtimeStats, error) {
	repo := database.NewBridgeRepository(db)

	total, err := repo.Count()
	if err != nil {
		return nil, fmt.Errorf("count total failed: %w", err)
	}

	available, err := repo.CountAvailable()
	if err != nil {
		return nil, fmt.Errorf("count available failed: %w", err)
	}

	unavailable, err := repo.CountUnavailable()
	if err != nil {
		return nil, fmt.Errorf("count unavailable failed: %w", err)
	}

	unknown, err := repo.CountUnknown()
	if err != nil {
		return nil, fmt.Errorf("count unknown failed: %w", err)
	}

	avgTime, err := repo.GetAvgResponseTime()
	if err != nil {
		return nil, fmt.Errorf("get avg response time failed: %w", err)
	}

	lastFetch, _ := repo.GetLastFetchTime()

	return &RealtimeStats{
		TotalBridges:       total,
		AvailableBridges:   available,
		UnavailableBridges: unavailable,
		UnknownBridges:     unknown,
		AvgResponseTime:    avgTime,
		LastFetchTime:      lastFetch,
	}, nil
}

func (s *RealtimeStats) AvailableRate() float64 {
	if s.TotalBridges == 0 {
		return 0
	}
	return float64(s.AvailableBridges) / float64(s.TotalBridges) * 100
}
