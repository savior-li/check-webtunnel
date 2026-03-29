package exporter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"tor-bridge-collector/internal/fetcher"
	"tor-bridge-collector/internal/storage"
	"tor-bridge-collector/pkg/models"
)

type Exporter struct {
	storage *storage.Storage
}

func New(s *storage.Storage) *Exporter {
	return &Exporter{storage: s}
}

func (e *Exporter) ExportTorrc(path string, filter *models.BridgeFilter) (int, error) {
	bridges, _, err := e.storage.GetAllBridges(filter)
	if err != nil {
		return 0, err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return 0, err
	}

	content := fetcher.FormatAsTorrc(bridges)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return 0, err
	}

	return len(bridges), nil
}

func (e *Exporter) ExportJSON(path string, filter *models.BridgeFilter) (int, error) {
	bridges, _, err := e.storage.GetAllBridges(filter)
	if err != nil {
		return 0, err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return 0, err
	}

	var output []map[string]interface{}
	for _, b := range bridges {
		output = append(output, b.ToJSON())
	}

	content, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return 0, err
	}

	if err := os.WriteFile(path, content, 0644); err != nil {
		return 0, err
	}

	return len(bridges), nil
}

func (e *Exporter) ExportCSV(path string, filter *models.BridgeFilter) (int, error) {
	bridges, _, err := e.storage.GetAllBridges(filter)
	if err != nil {
		return 0, err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return 0, err
	}

	content := fetcher.FormatAsCSV(bridges)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return 0, err
	}

	return len(bridges), nil
}

func (e *Exporter) Export(path string, format models.ExportFormat, filter *models.BridgeFilter) (int, error) {
	switch format {
	case models.FormatTorrc:
		return e.ExportTorrc(path, filter)
	case models.FormatJSON:
		return e.ExportJSON(path, filter)
	case models.FormatCSV:
		return e.ExportCSV(path, filter)
	default:
		return 0, fmt.Errorf("unsupported format: %s", format)
	}
}
