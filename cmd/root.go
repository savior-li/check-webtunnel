package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"tor-bridge-collector/internal/config"
	"tor-bridge-collector/internal/storage"
)

var rootCmd = &cobra.Command{
	Use:   "tor-bridge-collector",
	Short: "Tor Bridge Collector - Fetch and manage Tor bridges",
	Long: `A tool to fetch, validate and manage Tor webtunnel bridges.
Supports multiple export formats (torrc, JSON, CSV) and provides
both CLI and Web UI interfaces.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(fetchCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(statsCmd)
	rootCmd.AddCommand(serveCmd)
}

func loadConfigAndStorage() (*config.Config, *storage.Storage, error) {
	cfg, err := config.LoadDefault()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load config: %w", err)
	}

	s, err := storage.New(cfg.App.DBPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to init storage: %w", err)
	}

	return cfg, s, nil
}

func ensureOutputDirs() {
	os.MkdirAll("data", 0755)
	os.MkdirAll("logs", 0755)
	os.MkdirAll("output", 0755)
}
