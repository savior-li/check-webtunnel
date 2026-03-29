package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"tor-bridge-collector/internal/config"
	"tor-bridge-collector/internal/storage"
)

var initForce bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration and database",
	Long:  `Create configuration file and initialize SQLite database in the current directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ensureOutputDirs()

		if err := config.InitConfig(initForce); err != nil {
			if os.IsExist(err) {
				fmt.Println("Config file already exists. Use --force to overwrite.")
				return nil
			}
			return fmt.Errorf("failed to init config: %w", err)
		}

		fmt.Println("Config file created: config.yaml")

		if err := storage.InitDB("./data/bridges.db"); err != nil {
			return fmt.Errorf("failed to init database: %w", err)
		}

		fmt.Println("Database initialized: ./data/bridges.db")
		fmt.Println("Initialization completed successfully!")

		return nil
	},
}

func init() {
	initCmd.Flags().BoolVar(&initForce, "force", false, "Force overwrite existing config and database")
}
