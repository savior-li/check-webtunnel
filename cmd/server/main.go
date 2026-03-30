package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/urfave/cli/v2"

	"tor-bridge-collector/configs"
	"tor-bridge-collector/pkg/bridge"
	"tor-bridge-collector/pkg/database"
	"tor-bridge-collector/pkg/exporter"
	"tor-bridge-collector/pkg/i18n"
	"tor-bridge-collector/pkg/statistics"
	"tor-bridge-collector/pkg/validator"
)

var (
	version = "1.0.0"
)

func main() {
	app := &cli.App{
		Name:    "tor-bridge-collector",
		Usage:   "Tor Bridge Collector - Fetch and manage Tor bridges",
		Version: version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Value:   "config.yaml",
				Usage:   "Config file path",
			},
			&cli.StringFlag{
				Name:    "lang",
				Aliases: []string{"l"},
				Value:   "zh",
				Usage:   "Language (en/zh)",
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "init",
				Aliases: []string{"i"},
				Usage:   "Initialize config and database",
				Action:  initAction,
			},
			{
				Name:    "fetch",
				Aliases: []string{"f"},
				Usage:   "Fetch bridge data",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "proxy",
						Aliases: []string{"p"},
						Usage:   "Proxy server address",
					},
				},
				Action: fetchAction,
			},
			{
				Name:    "validate",
				Aliases: []string{"v"},
				Usage:   "Validate bridge availability",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:    "timeout",
						Aliases: []string{"t"},
						Value:   10,
						Usage:   "Timeout in seconds",
					},
					&cli.IntFlag{
						Name:    "workers",
						Aliases: []string{"w"},
						Value:   5,
						Usage:   "Concurrent workers",
					},
				},
				Action: validateAction,
			},
			{
				Name:    "export",
				Aliases: []string{"e"},
				Usage:   "Export bridge data",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "format",
						Aliases: []string{"f"},
						Value:   "torrc",
						Usage:   "Output format (torrc/json/all)",
					},
					&cli.StringFlag{
						Name:    "output",
						Aliases: []string{"o"},
						Value:   "./output",
						Usage:   "Output directory",
					},
				},
				Action: exportAction,
			},
			{
				Name:    "stats",
				Aliases: []string{"s"},
				Usage:   "Show statistics",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "period",
						Aliases: []string{"p"},
						Value:   "day",
						Usage:   "Stats period (day/week/month)",
					},
				},
				Action: statsAction,
			},
			{
				Name:    "query",
				Aliases: []string{"q"},
				Usage:   "Query bridges with validation stats",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:  "min-validation",
						Usage: "Minimum validation count",
					},
					&cli.IntFlag{
						Name:  "min-success",
						Usage: "Minimum success count",
					},
					&cli.BoolFlag{
						Name:  "available",
						Usage: "Only available bridges",
					},
					&cli.StringFlag{
						Name:  "order-by",
						Value: "validation_count",
						Usage: "Order by (validation_count/success_rate/last_validated)",
					},
					&cli.StringFlag{
						Name:  "order",
						Value: "desc",
						Usage: "Order direction (asc/desc)",
					},
					&cli.IntFlag{
						Name:  "limit",
						Usage: "Result limit",
					},
					&cli.StringFlag{
						Name:  "format",
						Value: "torrc",
						Usage: "Output format (torrc/json/all)",
					},
					&cli.StringFlag{
						Name:    "output",
						Aliases: []string{"o"},
						Value:   "./output",
						Usage:   "Output directory",
					},
				},
				Action: queryAction,
			},
			{
				Name:    "import",
				Aliases: []string{"imp"},
				Usage:   "Import bridges from file",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "file",
						Aliases:  []string{"f"},
						Required: true,
						Usage:    "Import file path",
					},
					&cli.StringFlag{
						Name:  "format",
						Value: "text",
						Usage: "File format (text/csv)",
					},
					&cli.StringFlag{
						Name:  "transport",
						Value: "webtunnel",
						Usage: "Default transport type",
					},
					&cli.BoolFlag{
						Name:  "validate",
						Usage: "Validate after import",
					},
				},
				Action: importAction,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func getConfig(c *cli.Context) *configs.Config {
	configPath := c.String("config")
	cfg, err := configs.Load(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			cfg = configs.DefaultConfig()
		} else {
			log.Printf("Warning: failed to load config: %v, using defaults", err)
			cfg = configs.DefaultConfig()
		}
	}
	return cfg
}

func getLang(c *cli.Context) string {
	lang := c.String("lang")
	if lang == "" {
		cfg := getConfig(c)
		lang = cfg.App.Lang
	}
	return lang
}

func t(c *cli.Context, key string) string {
	lang := getLang(c)
	translator := i18n.NewTranslator(lang)
	return translator.T(key)
}

func initAction(c *cli.Context) error {
	cfg := getConfig(c)

	fmt.Println(t(c, "init_success"))
	fmt.Printf("  %s: %s\n", t(c, "config_created"), cfg.Database.Path)

	if err := os.MkdirAll("./", 0755); err != nil {
		return fmt.Errorf("create dir failed: %w", err)
	}

	configPath := c.String("config")
	if err := configs.Save(configPath, cfg); err != nil {
		return fmt.Errorf("save config failed: %w", err)
	}
	fmt.Printf("  %s: %s\n", t(c, "config_created"), configPath)

	db, err := database.New(cfg.Database.Path)
	if err != nil {
		return fmt.Errorf("create db failed: %w", err)
	}
	defer db.Close()

	if err := db.InitSchema(); err != nil {
		return fmt.Errorf("init schema failed: %w", err)
	}
	fmt.Printf("  %s: %s\n", t(c, "db_created"), cfg.Database.Path)

	return nil
}

func fetchAction(c *cli.Context) error {
	cfg := getConfig(c)
	lang := getLang(c)
	translator := i18n.NewTranslator(lang)

	fmt.Println(translator.T("fetching"))

	db, err := database.New(cfg.Database.Path)
	if err != nil {
		return fmt.Errorf("open db failed: %w", err)
	}
	defer db.Close()

	repo := database.NewBridgeRepository(db)

	proxyStr := c.String("proxy")
	fetcher := bridge.NewFetcher(cfg.Fetch.URL, time.Duration(cfg.Fetch.Timeout)*time.Second)

	if proxyStr != "" {
		if err := fetcher.SetProxy(proxyStr); err != nil {
			return fmt.Errorf("set proxy failed: %w", err)
		}
	}

	bridges, err := fetcher.Fetch()
	if err != nil {
		return fmt.Errorf("%s: %w", translator.T("fetch_failed"), err)
	}

	newCount := 0
	for _, b := range bridges {
		id, isNew, err := repo.Upsert(&b)
		if err != nil {
			log.Printf("Warning: upsert bridge failed: %v", err)
			continue
		}
		if isNew {
			newCount++
			_ = id
		}
	}

	fmt.Printf("%s: %d, %s: %d\n",
		translator.T("bridges_found"), len(bridges),
		translator.T("stats_total"), newCount)
	fmt.Println(translator.T("fetch_success"))

	return nil
}

func validateAction(c *cli.Context) error {
	cfg := getConfig(c)
	lang := getLang(c)
	translator := i18n.NewTranslator(lang)

	fmt.Println(translator.T("validating"))

	db, err := database.New(cfg.Database.Path)
	if err != nil {
		return fmt.Errorf("open db failed: %w", err)
	}
	defer db.Close()

	repo := database.NewBridgeRepository(db)
	historyRepo := database.NewHistoryRepository(db)

	bridges, err := repo.GetAll()
	if err != nil {
		return fmt.Errorf("get bridges failed: %w", err)
	}

	timeout := c.Int("timeout")
	workers := c.Int("workers")
	v := validator.NewValidator(timeout, workers)

	available := 0
	unavailable := 0

	err = v.ValidateConcurrent(bridges, func(result *validator.ValidationResult) {
		for i, b := range bridges {
			if b.ID == result.BridgeID {
				repo.UpdateAvailability(b.ID, result.IsAvailable, result.ResponseTime)
				historyRepo.InsertBridgeValidation(b.ID, &bridges[i], result)

				if result.IsAvailable {
					available++
				} else {
					unavailable++
				}
				break
			}
		}
	})

	if err != nil {
		return fmt.Errorf("%s: %w", translator.T("validate_failed"), err)
	}

	fmt.Printf("%s: %d, %s: %d\n",
		translator.T("bridge_available"), available,
		translator.T("bridge_unavailable"), unavailable)
	fmt.Println(translator.T("validate_success"))

	return nil
}

func exportAction(c *cli.Context) error {
	cfg := getConfig(c)
	lang := getLang(c)
	translator := i18n.NewTranslator(lang)

	fmt.Println(translator.T("exporting"))

	db, err := database.New(cfg.Database.Path)
	if err != nil {
		return fmt.Errorf("open db failed: %w", err)
	}
	defer db.Close()

	repo := database.NewBridgeRepository(db)

	bridges, err := repo.GetAll()
	if err != nil {
		return fmt.Errorf("get bridges failed: %w", err)
	}

	format := exporter.ExportFormat(c.String("format"))
	outputDir := c.String("output")

	if err := exporter.Export(bridges, format, outputDir); err != nil {
		return fmt.Errorf("%s: %w", translator.T("export_failed"), err)
	}

	fmt.Printf("%s: %s\n", translator.T("export_file"), outputDir)
	fmt.Println(translator.T("export_success"))

	return nil
}

func statsAction(c *cli.Context) error {
	cfg := getConfig(c)
	lang := getLang(c)
	translator := i18n.NewTranslator(lang)

	fmt.Println(translator.T("stats_title"))

	db, err := database.New(cfg.Database.Path)
	if err != nil {
		return fmt.Errorf("open db failed: %w", err)
	}
	defer db.Close()

	stats, err := statistics.GetRealtimeStats(db)
	if err != nil {
		return fmt.Errorf("get stats failed: %w", err)
	}

	fmt.Printf("  %s: %d\n", translator.T("stats_total"), stats.TotalBridges)
	fmt.Printf("  %s: %d\n", translator.T("stats_available"), stats.AvailableBridges)
	fmt.Printf("  %s: %d\n", translator.T("stats_unavailable"), stats.UnavailableBridges)
	fmt.Printf("  %s: %d\n", translator.T("stats_unknown"), stats.UnknownBridges)
	fmt.Printf("  %s: %.2f ms\n", translator.T("stats_avg_time"), stats.AvgResponseTime)
	if !stats.LastFetchTime.IsZero() {
		fmt.Printf("  %s: %s\n", translator.T("stats_last_fetch"), stats.LastFetchTime.Format(time.RFC3339))
	}

	period := c.String("period")
	limit := 7

	if period != "day" {
		historical, err := statistics.GetStatsByPeriod(db, period, limit)
		if err != nil {
			log.Printf("Warning: get historical stats failed: %v", err)
		} else {
			fmt.Printf("\n--- %s ---\n", translator.T("stats_title"))
			for _, s := range historical {
				fmt.Printf("  %s: total=%d available=%d avg=%.2fms\n",
					s.Date, s.TotalCount, s.AvailableCount, s.AvgResponseTime)
			}
		}
	}

	return nil
}

func queryAction(c *cli.Context) error {
	cfg := getConfig(c)
	lang := getLang(c)
	translator := i18n.NewTranslator(lang)

	fmt.Println(translator.T("querying"))

	db, err := database.New(cfg.Database.Path)
	if err != nil {
		return fmt.Errorf("open db failed: %w", err)
	}
	defer db.Close()

	repo := database.NewBridgeRepository(db)

	opts := &database.BridgeQueryOption{
		MinValidationCount: c.Int("min-validation"),
		MinSuccessCount:    c.Int("min-success"),
		OrderBy:            c.String("order-by"),
		OrderDesc:          c.String("order") == "desc",
		Limit:              c.Int("limit"),
	}

	if c.Bool("available") {
		avail := true
		opts.IsAvailable = &avail
	}

	bridgesWithStats, err := repo.GetBridgesWithStats(opts)
	if err != nil {
		return fmt.Errorf("%s: %w", translator.T("query_failed"), err)
	}

	var bridges []bridge.Bridge
	for _, s := range bridgesWithStats {
		bridges = append(bridges, s.Bridge)
	}

	fmt.Printf("%s: %d\n", translator.T("bridges_found"), len(bridges))

	format := exporter.ExportFormat(c.String("format"))
	outputDir := c.String("output")

	if len(bridges) > 0 {
		if err := exporter.Export(bridges, format, outputDir); err != nil {
			return fmt.Errorf("%s: %w", translator.T("export_failed"), err)
		}
		fmt.Printf("%s: %s\n", translator.T("export_file"), outputDir)
	}

	fmt.Println(translator.T("query_success"))
	return nil
}

func importAction(c *cli.Context) error {
	cfg := getConfig(c)
	lang := getLang(c)
	translator := i18n.NewTranslator(lang)

	filePath := c.String("file")
	format := c.String("format")
	transport := c.String("transport")

	fmt.Println(translator.T("importing"))

	db, err := database.New(cfg.Database.Path)
	if err != nil {
		return fmt.Errorf("open db failed: %w", err)
	}
	defer db.Close()

	repo := database.NewBridgeRepository(db)

	fileImporter := bridge.NewFileImporter(format, transport)
	bridges, err := fileImporter.Import(filePath)
	if err != nil {
		return fmt.Errorf("%s: %w", translator.T("import_file_failed"), err)
	}

	imported := 0
	skipped := 0
	for _, b := range bridges {
		id, isNew, err := repo.Upsert(&b)
		if err != nil {
			log.Printf("Warning: upsert bridge failed: %v", err)
			continue
		}
		if isNew {
			imported++
			_ = id
		} else {
			skipped++
		}
	}

	fmt.Printf("%s: %d, %s: %d\n",
		translator.T("import_total"), len(bridges),
		translator.T("import_imported"), imported)
	fmt.Printf("%s: %d\n", translator.T("import_skipped"), skipped)

	if c.Bool("validate") {
		fmt.Println(translator.T("validating"))
	}

	fmt.Println(translator.T("import_success"))
	return nil
}
