package database

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"tor-bridge-collector/pkg/bridge"
)

func setupTestDB(t *testing.T) (*DB, func()) {
	t.Helper()

	tempDir := t.TempDir()
	dbPath := tempDir + "/test.db"

	db, err := New(dbPath)
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

func TestNew(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := tempDir + "/test.db"

	db, err := New(dbPath)
	assert.NoError(t, err)
	assert.NotNil(t, db)
	assert.Equal(t, dbPath, db.Path())

	db.Close()
}

func TestDB_InitSchema(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	err := db.InitSchema()
	assert.NoError(t, err)
}

func TestDB_Close(t *testing.T) {
	db, cleanup := setupTestDB(t)

	err := db.Close()
	assert.NoError(t, err)

	cleanup()
}

func TestDB_Path(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	assert.Equal(t, db.Path(), db.path)
}

func TestBridgeRepository_Insert(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewBridgeRepository(db)

	b := &bridge.Bridge{
		Hash:         "test-hash-123",
		Transport:    "webtunnel",
		Address:      "192.168.1.1",
		Port:         443,
		Fingerprint:  "ABCD1234",
		DiscoveredAt: time.Now(),
		IsAvailable:  -1,
	}

	id, err := repo.Insert(b)
	assert.NoError(t, err)
	assert.Greater(t, id, int64(0))
}

func TestBridgeRepository_Insert_Duplicate(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewBridgeRepository(db)

	b := &bridge.Bridge{
		Hash:        "test-hash-123",
		Transport:   "webtunnel",
		Address:     "192.168.1.1",
		Port:        443,
		IsAvailable: -1,
	}

	_, err := repo.Insert(b)
	assert.NoError(t, err)

	_, err = repo.Insert(b)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "UNIQUE constraint")
}

func TestBridgeRepository_Upsert_New(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewBridgeRepository(db)

	b := &bridge.Bridge{
		Hash:        "test-hash-upsert",
		Transport:   "webtunnel",
		Address:     "192.168.1.1",
		Port:        443,
		Fingerprint: "ABCD1234",
		IsAvailable: -1,
	}

	id, isNew, err := repo.Upsert(b)
	assert.NoError(t, err)
	assert.True(t, isNew)
	assert.Greater(t, id, int64(0))
}

func TestBridgeRepository_Upsert_Existing(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewBridgeRepository(db)

	b := &bridge.Bridge{
		Hash:        "test-hash-upsert",
		Transport:   "webtunnel",
		Address:     "192.168.1.1",
		Port:        443,
		Fingerprint: "ORIGINAL",
		IsAvailable: -1,
	}

	id, isNew, err := repo.Upsert(b)
	assert.NoError(t, err)
	assert.True(t, isNew)
	assert.Greater(t, id, int64(0))
}

func TestBridgeRepository_GetAll(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewBridgeRepository(db)

	bridges := []bridge.Bridge{
		{Hash: "hash1", Transport: "webtunnel", Address: "192.168.1.1", Port: 443, IsAvailable: -1},
		{Hash: "hash2", Transport: "webtunnel", Address: "192.168.1.2", Port: 8080, IsAvailable: -1},
	}

	for _, b := range bridges {
		repo.Insert(&b)
	}

	all, err := repo.GetAll()
	assert.NoError(t, err)
	assert.Len(t, all, 2)
}

func TestBridgeRepository_GetAll_Empty(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewBridgeRepository(db)

	all, err := repo.GetAll()
	assert.NoError(t, err)
	assert.Empty(t, all)
}

func TestBridgeRepository_GetByHash(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewBridgeRepository(db)

	b := &bridge.Bridge{
		Hash:        "unique-hash",
		Transport:   "webtunnel",
		Address:     "192.168.1.1",
		Port:        443,
		Fingerprint: "ABCD1234",
		IsAvailable: -1,
	}

	repo.Insert(b)

	found, err := repo.GetByHash("unique-hash")
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, b.Address, found.Address)
	assert.Equal(t, b.Port, found.Port)
}

func TestBridgeRepository_GetByHash_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewBridgeRepository(db)

	found, err := repo.GetByHash("nonexistent")
	assert.Error(t, err)
	assert.Nil(t, found)
}

func TestBridgeRepository_UpdateAvailability(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewBridgeRepository(db)

	b := &bridge.Bridge{
		Hash:        "test-hash",
		Transport:   "webtunnel",
		Address:     "192.168.1.1",
		Port:        443,
		IsAvailable: -1,
	}

	id, _ := repo.Insert(b)

	err := repo.UpdateAvailability(id, true, 100)
	assert.NoError(t, err)

	updated, _ := repo.GetByHash("test-hash")
	assert.Equal(t, 1, updated.IsAvailable)
	assert.Equal(t, 100, updated.ResponseTime)
}

func TestBridgeRepository_Count(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewBridgeRepository(db)

	count, err := repo.Count()
	assert.NoError(t, err)
	assert.Equal(t, 0, count)

	b := &bridge.Bridge{
		Hash:        "test-hash",
		Transport:   "webtunnel",
		Address:     "192.168.1.1",
		Port:        443,
		IsAvailable: -1,
	}
	repo.Insert(b)

	count, err = repo.Count()
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestBridgeRepository_CountAvailable(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewBridgeRepository(db)

	count, err := repo.CountAvailable()
	assert.NoError(t, err)
	assert.Equal(t, 0, count)

	b1 := &bridge.Bridge{Hash: "hash1", Transport: "webtunnel", Address: "192.168.1.1", Port: 443, IsAvailable: 1}
	b2 := &bridge.Bridge{Hash: "hash2", Transport: "webtunnel", Address: "192.168.1.2", Port: 443, IsAvailable: 0}
	repo.Insert(b1)
	repo.Insert(b2)

	count, err = repo.CountAvailable()
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestBridgeRepository_CountUnavailable(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewBridgeRepository(db)

	count, _ := repo.CountUnavailable()
	assert.Equal(t, 0, count)

	b1 := &bridge.Bridge{Hash: "hash1", Transport: "webtunnel", Address: "192.168.1.1", Port: 443, IsAvailable: 1}
	b2 := &bridge.Bridge{Hash: "hash2", Transport: "webtunnel", Address: "192.168.1.2", Port: 443, IsAvailable: 0}
	repo.Insert(b1)
	repo.Insert(b2)

	count, _ = repo.CountUnavailable()
	assert.Equal(t, 1, count)
}

func TestBridgeRepository_CountUnknown(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewBridgeRepository(db)

	count, _ := repo.CountUnknown()
	assert.Equal(t, 0, count)

	b := &bridge.Bridge{Hash: "hash1", Transport: "webtunnel", Address: "192.168.1.1", Port: 443, IsAvailable: -1}
	repo.Insert(b)

	count, _ = repo.CountUnknown()
	assert.Equal(t, 1, count)
}

func TestBridgeRepository_DeleteOld(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewBridgeRepository(db)

	oldBridge := &bridge.Bridge{
		Hash:        "old-hash",
		Transport:   "webtunnel",
		Address:     "192.168.1.1",
		Port:        443,
		IsAvailable: -1,
	}
	repo.Insert(oldBridge)

	err := repo.DeleteOld(0)
	assert.NoError(t, err)

	count, _ := repo.Count()
	assert.Equal(t, 0, count)
}

func TestBridgeRepository_GetLastFetchTime_Empty(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewBridgeRepository(db)

	t1, err := repo.GetLastFetchTime()
	assert.Error(t, err)
	assert.True(t, t1.IsZero())
}

func TestBridgeRepository_GetAvgResponseTime_Empty(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewBridgeRepository(db)

	avg, err := repo.GetAvgResponseTime()
	assert.NoError(t, err)
	assert.Equal(t, 0.0, avg)
}
