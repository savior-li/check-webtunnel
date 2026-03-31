package database

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"tor-bridge-collector/pkg/bridge"
	"tor-bridge-collector/pkg/validator"
)

func TestHistoryRepository_InsertBridgeValidation(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewBridgeRepository(db)
	historyRepo := NewHistoryRepository(db)

	b := &bridge.Bridge{
		Hash:        "test-hash",
		Transport:   "webtunnel",
		Address:     "192.168.1.1",
		Port:        443,
		IsAvailable: -1,
	}
	id, _ := repo.Insert(b)

	result := &validator.ValidationResult{
		BridgeID:     id,
		IsAvailable:  true,
		ResponseTime: 100,
		ValidatedAt:  time.Now(),
	}

	err := historyRepo.InsertBridgeValidation(id, b, result)
	assert.NoError(t, err)
}

func TestHistoryRepository_GetByBridgeID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewBridgeRepository(db)
	historyRepo := NewHistoryRepository(db)

	b := &bridge.Bridge{
		Hash:        "test-hash",
		Transport:   "webtunnel",
		Address:     "192.168.1.1",
		Port:        443,
		IsAvailable: -1,
	}
	id, _ := repo.Insert(b)

	result := &validator.ValidationResult{
		BridgeID:     id,
		IsAvailable:  true,
		ResponseTime: 100,
		ValidatedAt:  time.Now(),
	}

	historyRepo.InsertBridgeValidation(id, b, result)
	historyRepo.InsertBridgeValidation(id, b, result)

	history, err := historyRepo.GetByBridgeID(id, 10)
	assert.NoError(t, err)
	assert.Len(t, history, 2)
}

func TestHistoryRepository_GetByBridgeID_Empty(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	historyRepo := NewHistoryRepository(db)

	history, err := historyRepo.GetByBridgeID(9999, 10)
	assert.NoError(t, err)
	assert.Empty(t, history)
}

func TestHistoryRepository_GetByBridgeID_Limit(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewBridgeRepository(db)
	historyRepo := NewHistoryRepository(db)

	b := &bridge.Bridge{
		Hash:        "test-hash",
		Transport:   "webtunnel",
		Address:     "192.168.1.1",
		Port:        443,
		IsAvailable: -1,
	}
	id, _ := repo.Insert(b)

	for i := 0; i < 5; i++ {
		result := &validator.ValidationResult{
			BridgeID:     id,
			IsAvailable:  true,
			ResponseTime: 100,
			ValidatedAt:  time.Now(),
		}
		historyRepo.InsertBridgeValidation(id, b, result)
	}

	history, err := historyRepo.GetByBridgeID(id, 3)
	assert.NoError(t, err)
	assert.Len(t, history, 3)
}

func TestHistoryRepository_InsertBridgeValidation_WithError(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewBridgeRepository(db)
	historyRepo := NewHistoryRepository(db)

	b := &bridge.Bridge{
		Hash:        "test-hash",
		Transport:   "webtunnel",
		Address:     "192.168.1.1",
		Port:        443,
		IsAvailable: -1,
	}
	id, _ := repo.Insert(b)

	result := &validator.ValidationResult{
		BridgeID:     id,
		IsAvailable:  false,
		ResponseTime: 0,
		Error:        assert.AnError,
		ValidatedAt:  time.Now(),
	}

	err := historyRepo.InsertBridgeValidation(id, b, result)
	assert.NoError(t, err)
}
