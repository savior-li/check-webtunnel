package database

import (
	"database/sql"
	"fmt"
	"time"

	"tor-bridge-collector/pkg/bridge"
	"tor-bridge-collector/pkg/validator"
)

type ValidationRecord struct {
	ID           int64
	BridgeID     int64
	ValidatedAt  time.Time
	IsAvailable  bool
	ResponseTime int
	ErrorMessage string
}

type DailyStats struct {
	Date            string
	TotalCount      int
	AvailableCount  int
	AvgResponseTime float64
}

type HistoryRepository struct {
	db *DB
}

func NewHistoryRepository(db *DB) *HistoryRepository {
	return &HistoryRepository{db: db}
}

func (r *HistoryRepository) Insert(record *ValidationRecord) error {
	avail := 0
	if record.IsAvailable {
		avail = 1
	}

	var errorMsg interface{} = nil
	if record.ErrorMessage != "" {
		errorMsg = record.ErrorMessage
	}

	_, err := r.db.Exec(`
		INSERT INTO validation_history (bridge_id, validated_at, is_available, response_time_ms, error_message)
		VALUES (?, ?, ?, ?, ?)
	`, record.BridgeID, record.ValidatedAt, avail, record.ResponseTime, errorMsg)

	if err != nil {
		return fmt.Errorf("insert history failed: %w", err)
	}
	return nil
}

func (r *HistoryRepository) GetByBridgeID(bridgeID int64, limit int) ([]ValidationRecord, error) {
	rows, err := r.db.Query(`
		SELECT id, bridge_id, validated_at, is_available, response_time_ms, error_message
		FROM validation_history
		WHERE bridge_id = ?
		ORDER BY validated_at DESC
		LIMIT ?
	`, bridgeID, limit)
	if err != nil {
		return nil, fmt.Errorf("query history failed: %w", err)
	}
	defer rows.Close()

	var records []ValidationRecord
	for rows.Next() {
		var rec ValidationRecord
		var isAvailable sql.NullInt64
		var responseTime sql.NullInt64
		var errorMsg sql.NullString

		err := rows.Scan(&rec.ID, &rec.BridgeID, &rec.ValidatedAt, &isAvailable, &responseTime, &errorMsg)
		if err != nil {
			return nil, fmt.Errorf("scan history failed: %w", err)
		}

		if isAvailable.Valid {
			rec.IsAvailable = isAvailable.Int64 == 1
		}
		if responseTime.Valid {
			rec.ResponseTime = int(responseTime.Int64)
		}
		if errorMsg.Valid {
			rec.ErrorMessage = errorMsg.String
		}

		records = append(records, rec)
	}

	return records, nil
}

func (r *HistoryRepository) GetStatsByPeriod(period string, days int) ([]DailyStats, error) {
	cutoff := time.Now().AddDate(0, 0, -days)

	rows, err := r.db.Query(`
		SELECT 
			date(validated_at) as date,
			COUNT(DISTINCT bridge_id) as total_count,
			SUM(CASE WHEN is_available = 1 THEN 1 ELSE 0 END) as available_count,
			AVG(CASE WHEN response_time_ms > 0 THEN response_time_ms ELSE NULL END) as avg_response
		FROM validation_history
		WHERE validated_at >= ?
		GROUP BY date(validated_at)
		ORDER BY date DESC
	`, cutoff)
	if err != nil {
		return nil, fmt.Errorf("query stats failed: %w", err)
	}
	defer rows.Close()

	var stats []DailyStats
	for rows.Next() {
		var s DailyStats
		var avgResponse sql.NullFloat64

		err := rows.Scan(&s.Date, &s.TotalCount, &s.AvailableCount, &avgResponse)
		if err != nil {
			return nil, fmt.Errorf("scan stats failed: %w", err)
		}

		if avgResponse.Valid {
			s.AvgResponseTime = avgResponse.Float64
		}

		stats = append(stats, s)
	}

	return stats, nil
}

func (r *HistoryRepository) InsertBridgeValidation(bridgeID int64, b *bridge.Bridge, result *validator.ValidationResult) error {
	record := &ValidationRecord{
		BridgeID:     bridgeID,
		ValidatedAt:  result.ValidatedAt,
		IsAvailable:  result.IsAvailable,
		ResponseTime: result.ResponseTime,
	}
	if result.Error != nil {
		record.ErrorMessage = result.Error.Error()
	}
	return r.Insert(record)
}
