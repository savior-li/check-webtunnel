package statistics

import (
	"fmt"

	"tor-bridge-collector/pkg/database"
)

type DailyStats = database.DailyStats

func GetDailyStats(db *database.DB, days int) ([]DailyStats, error) {
	repo := database.NewHistoryRepository(db)
	return repo.GetStatsByPeriod("day", days)
}

func GetWeeklyStats(db *database.DB, weeks int) ([]DailyStats, error) {
	repo := database.NewHistoryRepository(db)
	days := weeks * 7
	return repo.GetStatsByPeriod("week", days)
}

func GetMonthlyStats(db *database.DB, months int) ([]DailyStats, error) {
	repo := database.NewHistoryRepository(db)
	days := months * 30
	return repo.GetStatsByPeriod("month", days)
}

func GetStatsByPeriod(db *database.DB, period string, limit int) ([]DailyStats, error) {
	repo := database.NewHistoryRepository(db)

	switch period {
	case "week":
		limit *= 7
	case "month":
		limit *= 30
	}

	stats, err := repo.GetStatsByPeriod(period, limit)
	if err != nil {
		return nil, fmt.Errorf("get stats by period failed: %w", err)
	}

	return stats, nil
}
