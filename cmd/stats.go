package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"tor-bridge-collector/pkg/models"
)

var statsPeriod string

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show bridge statistics",
	Long:  `Display statistics about collected bridges.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		_, s, err := loadConfigAndStorage()
		if err != nil {
			return err
		}

		stats, err := s.GetStats()
		if err != nil {
			return fmt.Errorf("failed to get stats: %w", err)
		}

		fmt.Println("=== Bridge Statistics ===")
		fmt.Printf("Total Bridges:     %d\n", stats.TotalBridges)
		fmt.Printf("Valid Bridges:     %d\n", stats.ValidBridges)
		fmt.Printf("Invalid Bridges:   %d\n", stats.InvalidBridges)
		fmt.Printf("Average Latency:   %.2f ms\n", stats.AvgLatency)
		fmt.Printf("Success Rate:      %.2f %%\n", stats.SuccessRate)
		fmt.Printf("Total Validations: %d\n", stats.TotalValidations)

		return nil
	},
}

func init() {
	statsCmd.Flags().StringVar(&statsPeriod, "period", "24h", "Time period (24h, 7d, 30d)")
}

type StatsOutput struct {
	Stats models.Stats `json:"stats"`
}
