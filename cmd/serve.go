package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"tor-bridge-collector/internal/config"
	"tor-bridge-collector/internal/web"
)

var servePort int
var serveHost string

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start web UI server",
	Long:  `Start the web-based user interface for bridge management.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		_, s, err := loadConfigAndStorage()
		if err != nil {
			return err
		}

		cfg, err := config.LoadDefault()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		host := serveHost
		if host == "" {
			host = cfg.Server.Host
		}

		port := servePort
		if port == 0 {
			port = cfg.Server.Port
		}

		addr := fmt.Sprintf("%s:%d", host, port)
		fmt.Printf("Starting server on %s\n", addr)

		router := web.NewRouter(s, cfg)
		return router.Run(addr)
	},
}

func init() {
	serveCmd.Flags().IntVar(&servePort, "port", 0, "Server port")
	serveCmd.Flags().StringVar(&serveHost, "bind", "", "Bind address")
}
