package database

import (
	"database/sql"
	"fmt"
	"time"

	"tor-bridge-collector/pkg/bridge"
)

type BridgeRepository struct {
	db *DB
}

func NewBridgeRepository(db *DB) *BridgeRepository {
	return &BridgeRepository{db: db}
}

func (r *BridgeRepository) Insert(b *bridge.Bridge) (int64, error) {
	result, err := r.db.Exec(`
		INSERT INTO bridges (hash, transport, address, port, fingerprint, discovered_at, is_available)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, b.Hash, b.Transport, b.Address, b.Port, b.Fingerprint, b.DiscoveredAt, b.IsAvailable)
	if err != nil {
		return 0, fmt.Errorf("insert bridge failed: %w", err)
	}
	return result.LastInsertId()
}

func (r *BridgeRepository) Upsert(b *bridge.Bridge) (int64, bool, error) {
	var id int64
	err := r.db.QueryRow(`
		SELECT id FROM bridges WHERE hash = ?
	`, b.Hash).Scan(&id)

	if err == sql.ErrNoRows {
		id, err = r.Insert(b)
		if err != nil {
			return 0, false, err
		}
		return id, true, nil
	} else if err != nil {
		return 0, false, fmt.Errorf("query bridge failed: %w", err)
	}

	_, err = r.db.Exec(`
		UPDATE bridges SET 
			fingerprint = COALESCE(NULLIF(?, ''), fingerprint),
			updated_at = ?
		WHERE id = ?
	`, b.Fingerprint, time.Now(), id)
	if err != nil {
		return 0, false, fmt.Errorf("update bridge failed: %w", err)
	}

	return id, false, nil
}

func (r *BridgeRepository) GetAll() ([]bridge.Bridge, error) {
	rows, err := r.db.Query(`
		SELECT id, hash, transport, address, port, fingerprint, 
		       discovered_at, last_validated, is_available, response_time_ms,
		       created_at, updated_at
		FROM bridges
		ORDER BY discovered_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("query bridges failed: %w", err)
	}
	defer rows.Close()

	var bridges []bridge.Bridge
	for rows.Next() {
		var b bridge.Bridge
		var fingerprint, lastValidated sql.NullString
		var responseTime sql.NullInt64

		err := rows.Scan(
			&b.ID, &b.Hash, &b.Transport, &b.Address, &b.Port,
			&fingerprint, &b.DiscoveredAt, &lastValidated, &b.IsAvailable,
			&responseTime, &b.CreatedAt, &b.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan bridge failed: %w", err)
		}

		if fingerprint.Valid {
			b.Fingerprint = fingerprint.String
		}
		if lastValidated.Valid {
			b.LastValidAt, _ = time.Parse("2006-01-02 15:04:05", lastValidated.String)
		}
		if responseTime.Valid {
			b.ResponseTime = int(responseTime.Int64)
		}

		bridges = append(bridges, b)
	}

	return bridges, nil
}

func (r *BridgeRepository) GetByHash(hash string) (*bridge.Bridge, error) {
	var b bridge.Bridge
	var fingerprint, lastValidated sql.NullString
	var responseTime sql.NullInt64

	err := r.db.QueryRow(`
		SELECT id, hash, transport, address, port, fingerprint,
		       discovered_at, last_validated, is_available, response_time_ms,
		       created_at, updated_at
		FROM bridges WHERE hash = ?
	`, hash).Scan(
		&b.ID, &b.Hash, &b.Transport, &b.Address, &b.Port,
		&fingerprint, &b.DiscoveredAt, &lastValidated, &b.IsAvailable,
		&responseTime, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if fingerprint.Valid {
		b.Fingerprint = fingerprint.String
	}
	if responseTime.Valid {
		b.ResponseTime = int(responseTime.Int64)
	}

	return &b, nil
}

func (r *BridgeRepository) UpdateAvailability(id int64, available bool, responseTime int) error {
	now := time.Now()
	avail := 0
	if available {
		avail = 1
	}

	_, err := r.db.Exec(`
		UPDATE bridges SET 
			is_available = ?,
			response_time_ms = ?,
			last_validated = ?,
			updated_at = ?
		WHERE id = ?
	`, avail, responseTime, now, now, id)

	return err
}

func (r *BridgeRepository) DeleteOld(days int) error {
	cutoff := time.Now().AddDate(0, 0, -days)
	_, err := r.db.Exec(`
		DELETE FROM bridges WHERE discovered_at < ?
	`, cutoff)
	return err
}

func (r *BridgeRepository) Count() (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM bridges`).Scan(&count)
	return count, err
}

func (r *BridgeRepository) CountAvailable() (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM bridges WHERE is_available = 1`).Scan(&count)
	return count, err
}

func (r *BridgeRepository) CountUnavailable() (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM bridges WHERE is_available = 0`).Scan(&count)
	return count, err
}

func (r *BridgeRepository) CountUnknown() (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM bridges WHERE is_available = -1`).Scan(&count)
	return count, err
}

func (r *BridgeRepository) GetLastFetchTime() (time.Time, error) {
	var t time.Time
	err := r.db.QueryRow(`SELECT MAX(discovered_at) FROM bridges`).Scan(&t)
	return t, err
}

func (r *BridgeRepository) GetAvgResponseTime() (float64, error) {
	var avg sql.NullFloat64
	err := r.db.QueryRow(`
		SELECT AVG(response_time_ms) FROM bridges 
		WHERE response_time_ms > 0 AND is_available = 1
	`).Scan(&avg)
	if err != nil {
		return 0, err
	}
	if avg.Valid {
		return avg.Float64, nil
	}
	return 0, nil
}
