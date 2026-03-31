package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tor-bridge-collector/pkg/bridge"
	"tor-bridge-collector/pkg/database"
	"tor-bridge-collector/pkg/exporter"
)

func TestIntegration_BridgeFetchAndStore(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		html := `
			<html>
				<body>
					Bridge webtunnel 192.168.1.1:443 fingerprint ABCD1234
					Bridge webtunnel 10.0.0.1:8080 fingerprint EFGH5678
				</body>
			</html>
		`
		w.Write([]byte(html))
	}))
	defer ts.Close()

	db, err := database.New(":memory:")
	require.NoError(t, err)
	defer db.Close()

	err = db.InitSchema()
	require.NoError(t, err)

	fetcher := bridge.NewFetcher(ts.URL, 5*time.Second)
	bridges, err := fetcher.Fetch()
	require.NoError(t, err)
	assert.Len(t, bridges, 2)

	repo := database.NewBridgeRepository(db)
	for _, b := range bridges {
		_, _, err := repo.Upsert(&b)
		require.NoError(t, err)
	}

	count, err := repo.Count()
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestIntegration_DuplicateBridge(t *testing.T) {
	db, err := database.New(":memory:")
	require.NoError(t, err)
	defer db.Close()

	db.InitSchema()

	repo := database.NewBridgeRepository(db)

	b := &bridge.Bridge{
		Hash:        "test-hash",
		Transport:   "webtunnel",
		Address:     "192.168.1.1",
		Port:        443,
		IsAvailable: -1,
	}

	id1, isNew1, err := repo.Upsert(b)
	require.NoError(t, err)
	assert.True(t, isNew1)
	assert.Greater(t, id1, int64(0))

	id2, isNew2, err := repo.Upsert(b)
	require.NoError(t, err)
	assert.False(t, isNew2)
	assert.Equal(t, id1, id2)

	count, err := repo.Count()
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestIntegration_ValidationFlow(t *testing.T) {
	db, err := database.New(":memory:")
	require.NoError(t, err)
	defer db.Close()

	db.InitSchema()

	repo := database.NewBridgeRepository(db)
	historyRepo := database.NewHistoryRepository(db)

	b1 := &bridge.Bridge{Hash: "hash1", Address: "192.168.1.1", Port: 443, IsAvailable: -1}
	b2 := &bridge.Bridge{Hash: "hash2", Address: "192.168.1.2", Port: 443, IsAvailable: -1}
	repo.Insert(b1)
	repo.Insert(b2)

	bridges, err := repo.GetAll()
	require.NoError(t, err)
	assert.Len(t, bridges, 2)

	for _, b := range bridges {
		err := repo.UpdateAvailability(b.ID, false, 0)
		require.NoError(t, err)
	}

	availableCount, err := repo.CountAvailable()
	require.NoError(t, err)
	assert.Equal(t, 0, availableCount)

	unavailableCount, err := repo.CountUnavailable()
	require.NoError(t, err)
	assert.Equal(t, 2, unavailableCount)

	history, err := historyRepo.GetByBridgeID(bridges[0].ID, 10)
	assert.NoError(t, err)
	assert.Empty(t, history)
}

func TestIntegration_ExportFlow(t *testing.T) {
	tempDir := t.TempDir()

	bridges := []bridge.Bridge{
		{ID: 1, Hash: "hash1", Transport: "webtunnel", Address: "192.168.1.1", Port: 443, Fingerprint: "ABCD1234"},
		{ID: 2, Hash: "hash2", Transport: "webtunnel", Address: "10.0.0.1", Port: 8080},
	}

	err := exporter.Export(bridges, exporter.FormatTorrc, tempDir)
	require.NoError(t, err)

	torrcPath := filepath.Join(tempDir, "bridges.txt")
	assert.FileExists(t, torrcPath)

	content, err := os.ReadFile(torrcPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "Bridge webtunnel 192.168.1.1:443 fingerprint=ABCD1234")
	assert.Contains(t, string(content), "Bridge webtunnel 10.0.0.1:8080")
}

func TestIntegration_JSONExportFlow(t *testing.T) {
	tempDir := t.TempDir()

	bridges := []bridge.Bridge{
		{ID: 1, Hash: "hash1", Transport: "webtunnel", Address: "192.168.1.1", Port: 443},
	}

	err := exporter.Export(bridges, exporter.FormatJSON, tempDir)
	require.NoError(t, err)

	jsonPath := filepath.Join(tempDir, "bridges.json")
	assert.FileExists(t, jsonPath)

	content, err := os.ReadFile(jsonPath)
	require.NoError(t, err)

	var exportData exporter.ExportData
	err = json.Unmarshal(content, &exportData)
	require.NoError(t, err)
	assert.Equal(t, 1, exportData.TotalCount)
	assert.Len(t, exportData.Bridges, 1)
}

func TestIntegration_StatisticsFlow(t *testing.T) {
	db, err := database.New(":memory:")
	require.NoError(t, err)
	defer db.Close()

	db.InitSchema()

	repo := database.NewBridgeRepository(db)

	b1 := &bridge.Bridge{Hash: "hash1", Address: "192.168.1.1", Port: 443, IsAvailable: 1, ResponseTime: 100}
	b2 := &bridge.Bridge{Hash: "hash2", Address: "192.168.1.2", Port: 443, IsAvailable: 0}
	b3 := &bridge.Bridge{Hash: "hash3", Address: "192.168.1.3", Port: 443, IsAvailable: -1}
	repo.Insert(b1)
	repo.Insert(b2)
	repo.Insert(b3)

	total, err := repo.Count()
	require.NoError(t, err)
	assert.Equal(t, 3, total)

	available, err := repo.CountAvailable()
	require.NoError(t, err)
	assert.Equal(t, 1, available)

	unavailable, err := repo.CountUnavailable()
	require.NoError(t, err)
	assert.Equal(t, 1, unavailable)

	unknown, err := repo.CountUnknown()
	require.NoError(t, err)
	assert.Equal(t, 1, unknown)

	avg, err := repo.GetAvgResponseTime()
	require.NoError(t, err)
	assert.Equal(t, 0.0, avg)
}

func TestIntegration_FetcherWithProxy(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		html := `<html><body>Bridge webtunnel 192.168.1.1:443</body></html>`
		w.Write([]byte(html))
	}))
	defer ts.Close()

	fetcher := bridge.NewFetcher(ts.URL, 5*time.Second)

	err := fetcher.SetProxy("http://127.0.0.1:1080")
	assert.NoError(t, err)

	bridges, err := fetcher.Fetch()
	assert.NoError(t, err)
	assert.Len(t, bridges, 1)
}
