package exporter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"tor-bridge-collector/pkg/bridge"
)

func TestJSONExporter_Export(t *testing.T) {
	tempDir := t.TempDir()

	bridges := []bridge.Bridge{
		{
			ID:          1,
			Hash:        "hash1",
			Transport:   "webtunnel",
			Address:     "192.168.1.1",
			Port:        443,
			Fingerprint: "ABCD1234",
		},
		{
			ID:        2,
			Hash:      "hash2",
			Transport: "webtunnel",
			Address:   "10.0.0.1",
			Port:      8080,
		},
	}

	exporter := &JSONExporter{}
	err := exporter.Export(bridges, tempDir)

	assert.NoError(t, err)

	filePath := filepath.Join(tempDir, "bridges.json")
	assert.FileExists(t, filePath)

	data, err := os.ReadFile(filePath)
	assert.NoError(t, err)

	var exportData ExportData
	err = json.Unmarshal(data, &exportData)
	assert.NoError(t, err)
	assert.Equal(t, 2, exportData.TotalCount)
	assert.Len(t, exportData.Bridges, 2)
}

func TestTorrcExporter_Export(t *testing.T) {
	tempDir := t.TempDir()

	bridges := []bridge.Bridge{
		{
			ID:          1,
			Hash:        "hash1",
			Transport:   "webtunnel",
			Address:     "192.168.1.1",
			Port:        443,
			Fingerprint: "ABCD1234",
		},
		{
			ID:        2,
			Hash:      "hash2",
			Transport: "webtunnel",
			Address:   "10.0.0.1",
			Port:      8080,
		},
	}

	exporter := &TorrcExporter{}
	err := exporter.Export(bridges, tempDir)

	assert.NoError(t, err)

	filePath := filepath.Join(tempDir, "bridges.txt")
	assert.FileExists(t, filePath)

	content, err := os.ReadFile(filePath)
	assert.NoError(t, err)

	assert.Contains(t, string(content), "# Tor Bridge Collector Export")
	assert.Contains(t, string(content), "Bridge webtunnel 192.168.1.1:443 fingerprint=ABCD1234")
	assert.Contains(t, string(content), "Bridge webtunnel 10.0.0.1:8080")
	assert.Contains(t, string(content), "BridgeRelay 1")
}

func TestAllExporter_Export(t *testing.T) {
	tempDir := t.TempDir()

	bridges := []bridge.Bridge{
		{
			ID:        1,
			Hash:      "hash1",
			Transport: "webtunnel",
			Address:   "192.168.1.1",
			Port:      443,
		},
	}

	exporter := &AllExporter{}
	err := exporter.Export(bridges, tempDir)

	assert.NoError(t, err)

	jsonPath := filepath.Join(tempDir, "bridges.json")
	torrcPath := filepath.Join(tempDir, "bridges.txt")

	assert.FileExists(t, jsonPath)
	assert.FileExists(t, torrcPath)
}

func TestExport_CreatesDirectory(t *testing.T) {
	tempDir := t.TempDir()
	newDir := filepath.Join(tempDir, "nested", "output")

	bridges := []bridge.Bridge{
		{ID: 1, Hash: "hash1", Transport: "webtunnel", Address: "192.168.1.1", Port: 443},
	}

	exporter := &TorrcExporter{}
	err := exporter.Export(bridges, newDir)

	assert.NoError(t, err)
	assert.DirExists(t, newDir)
}

func TestExport_EmptyBridges(t *testing.T) {
	tempDir := t.TempDir()

	bridges := []bridge.Bridge{}

	exporter := &TorrcExporter{}
	err := exporter.Export(bridges, tempDir)

	assert.NoError(t, err)

	filePath := filepath.Join(tempDir, "bridges.txt")
	assert.FileExists(t, filePath)

	content, err := os.ReadFile(filePath)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "# Total bridges: 0")
}

func TestNewExporter(t *testing.T) {
	tests := []struct {
		format   ExportFormat
		expected interface{}
	}{
		{FormatJSON, &JSONExporter{}},
		{FormatTorrc, &TorrcExporter{}},
		{FormatAll, &AllExporter{}},
		{"unknown", &TorrcExporter{}},
	}

	for _, tt := range tests {
		t.Run(string(tt.format), func(t *testing.T) {
			exp := NewExporter(tt.format)
			assert.IsType(t, tt.expected, exp)
		})
	}
}

func TestExport_Function(t *testing.T) {
	tempDir := t.TempDir()

	bridges := []bridge.Bridge{
		{ID: 1, Hash: "hash1", Transport: "webtunnel", Address: "192.168.1.1", Port: 443},
	}

	err := Export(bridges, FormatTorrc, tempDir)
	assert.NoError(t, err)

	err = Export(bridges, FormatJSON, tempDir)
	assert.NoError(t, err)

	err = Export(bridges, FormatAll, tempDir)
	assert.NoError(t, err)
}

func TestExportData_JSON(t *testing.T) {
	data := ExportData{
		TotalCount: 1,
		Bridges: []bridge.Bridge{
			{ID: 1, Hash: "hash1", Transport: "webtunnel", Address: "192.168.1.1", Port: 443},
		},
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	assert.NoError(t, err)
	assert.Contains(t, string(jsonData), `"total_count": 1`)
	assert.Contains(t, string(jsonData), `"bridges"`)
}
