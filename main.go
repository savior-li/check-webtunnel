package main

import (
	"log"
	"os"

	"tor-bridge-collector/cmd"
	"tor-bridge-collector/internal/i18n"
)

func main() {
	if err := i18n.Init(); err != nil {
		log.Printf("Warning: i18n init failed: %v", err)
	}

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
