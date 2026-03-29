package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"tor-bridge-collector/internal/config"
	"tor-bridge-collector/internal/fetcher"
	"tor-bridge-collector/internal/storage"
)

var fetchProxy string

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch bridges from Tor project",
	Long:  `Fetch webtunnel bridges from https://bridges.torproject.org`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, s, err := loadConfigAndStorage()
		if err != nil {
			return err
		}

		fmt.Println("Fetching bridges...")

		var f *fetcher.Fetcher
		if fetchProxy != "" {
			f, err = fetcher.NewWithProxy(fetchProxy, cfg.Fetch.Timeout)
			if err != nil {
				return fmt.Errorf("failed to create fetcher with proxy: %w", err)
			}
		} else if cfg.Proxy.Enabled {
			proxyURL := fmt.Sprintf("%s://%s:%d", cfg.Proxy.Type, cfg.Proxy.Address, cfg.Proxy.Port)
			f, err = fetcher.NewWithProxy(proxyURL, cfg.Fetch.Timeout)
			if err != nil {
				return fmt.Errorf("failed to create fetcher with proxy: %w", err)
			}
		} else {
			f = fetcher.New(cfg.Fetch.Timeout)
		}

		bridges, err := f.FetchWithRetry(cfg.Validation.Retry)
		if err != nil {
			return fmt.Errorf("failed to fetch bridges: %w", err)
		}

		newCount := 0
		updateCount := 0
		for _, bridge := range bridges {
			existing, _ := s.GetBridgeByHash(bridge.Hash)
			if existing != nil {
				existing.LastSeen = bridge.LastSeen
				s.UpdateBridge(existing)
				updateCount++
			} else {
				s.CreateBridge(&bridge)
				newCount++
			}
		}

		fmt.Printf("Fetch completed: %d new, %d updated\n", newCount, updateCount)
		return nil
	},
}

func init() {
	fetchCmd.Flags().StringVar(&fetchProxy, "proxy", "", "Proxy URL (e.g., http://proxy:8080)")
}

func fetchWithContext(ctx context.Context, s *storage.Storage, cfg *config.Config) error {
	return nil
}
