package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"tor-bridge-collector/internal/validator"
	"tor-bridge-collector/pkg/models"
)

var validateAll bool
var validateID uint

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate bridge connectivity",
	Long:  `Test TCP connectivity to bridges and update their status.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, s, err := loadConfigAndStorage()
		if err != nil {
			return err
		}

		fmt.Println("Validating bridges...")

		v := validator.New(s, cfg.Validation.Timeout, cfg.Validation.Concurrency, cfg.Validation.Retry)
		ctx := context.Background()

		var bridges []models.Bridge
		if validateID > 0 {
			bridge, err := s.GetBridgeByID(validateID)
			if err != nil {
				return fmt.Errorf("bridge not found: %d", validateID)
			}
			bridges = []models.Bridge{*bridge}
		} else {
			filter := &models.BridgeFilter{}
			if !validateAll {
				valid := false
				filter.IsValid = &valid
			}
			var total int64
			bridges, total, err = s.GetAllBridges(filter)
			if err != nil {
				return err
			}
			if total == 0 {
				fmt.Println("No bridges to validate.")
				return nil
			}
		}

		valid, invalid, err := v.ValidateAndSave(ctx, bridges)
		if err != nil {
			return err
		}

		fmt.Printf("Validation completed: %d valid, %d invalid\n", valid, invalid)
		return nil
	},
}

func init() {
	validateCmd.Flags().BoolVar(&validateAll, "all", false, "Validate all bridges")
	validateCmd.Flags().UintVar(&validateID, "id", 0, "Validate specific bridge by ID")
}
