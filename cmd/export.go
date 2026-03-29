package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"tor-bridge-collector/internal/config"
	"tor-bridge-collector/internal/exporter"
	"tor-bridge-collector/pkg/models"
)

var exportFormat string
var exportPath string

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export bridges to file",
	Long:  `Export bridges in various formats (torrc, json, csv).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		_, s, err := loadConfigAndStorage()
		if err != nil {
			return err
		}

		cfg, err := config.LoadDefault()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		format := models.ExportFormat(exportFormat)
		path := exportPath

		if path == "" {
			switch format {
			case models.FormatTorrc:
				path = cfg.Export.TorrcPath
			case models.FormatJSON:
				path = cfg.Export.JSONPath
			case models.FormatCSV:
				path = cfg.Export.CSVPath
			}
		}

		exp := exporter.New(s)
		count, err := exp.Export(path, format, nil)
		if err != nil {
			return fmt.Errorf("export failed: %w", err)
		}

		fmt.Printf("Exported %d bridges to %s\n", count, path)
		return nil
	},
}

func init() {
	exportCmd.Flags().StringVar(&exportFormat, "format", "torrc", "Export format (torrc/json/csv)")
	exportCmd.Flags().StringVar(&exportPath, "path", "", "Output file path")
}
