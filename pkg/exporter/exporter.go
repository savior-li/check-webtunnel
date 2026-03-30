package exporter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"tor-bridge-collector/pkg/bridge"
)

type Exporter interface {
	Export(bridges []bridge.Bridge, outputDir string) error
}

type ExportFormat string

const (
	FormatTorrc ExportFormat = "torrc"
	FormatJSON  ExportFormat = "json"
	FormatAll   ExportFormat = "all"
)

type ExportData struct {
	ExportedAt time.Time       `json:"exported_at"`
	TotalCount int             `json:"total_count"`
	Bridges    []bridge.Bridge `json:"bridges"`
}

type JSONExporter struct{}

func (e *JSONExporter) Export(bridges []bridge.Bridge, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("create output dir failed: %w", err)
	}

	data := ExportData{
		ExportedAt: time.Now(),
		TotalCount: len(bridges),
		Bridges:    bridges,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json failed: %w", err)
	}

	filename := filepath.Join(outputDir, "bridges.json")
	if err := os.WriteFile(filename, jsonData, 0644); err != nil {
		return fmt.Errorf("write json file failed: %w", err)
	}

	return nil
}

type TorrcExporter struct{}

func (e *TorrcExporter) Export(bridges []bridge.Bridge, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("create output dir failed: %w", err)
	}

	var torrcLines []string
	for _, b := range bridges {
		torrcLines = append(torrcLines, bridge.FormatTorrcLine(&b))
	}

	content := fmt.Sprintf("# Tor Bridge Collector Export\n")
	content += fmt.Sprintf("# Exported at: %s\n", time.Now().Format(time.RFC3339))
	content += fmt.Sprintf("# Total bridges: %d\n\n", len(bridges))
	content += "BridgeRelay 1\n\n"
	for _, line := range torrcLines {
		content += line + "\n"
	}

	filename := filepath.Join(outputDir, "bridges.txt")
	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		return fmt.Errorf("write torrc file failed: %w", err)
	}

	return nil
}

type AllExporter struct{}

func (e *AllExporter) Export(bridges []bridge.Bridge, outputDir string) error {
	jsonExp := &JSONExporter{}
	if err := jsonExp.Export(bridges, outputDir); err != nil {
		return err
	}

	torrcExp := &TorrcExporter{}
	if err := torrcExp.Export(bridges, outputDir); err != nil {
		return err
	}

	return nil
}

func NewExporter(format ExportFormat) Exporter {
	switch format {
	case FormatJSON:
		return &JSONExporter{}
	case FormatTorrc:
		return &TorrcExporter{}
	case FormatAll:
		return &AllExporter{}
	default:
		return &TorrcExporter{}
	}
}

func Export(bridges []bridge.Bridge, format ExportFormat, outputDir string) error {
	exporter := NewExporter(format)
	return exporter.Export(bridges, outputDir)
}
