package i18n

var en = map[string]string{
	"app_name":        "Tor Bridge Collector",
	"app_description": "Tor Bridge Collection Tool",

	"init_success":   "Initialization successful",
	"init_failed":    "Initialization failed",
	"config_created": "Configuration file created",
	"db_created":     "Database created",

	"fetch_success": "Fetch successful",
	"fetch_failed":  "Fetch failed",
	"fetching":      "Fetching bridge data...",
	"bridges_found": "Found %d bridges",

	"validate_success":  "Validation completed",
	"validate_failed":   "Validation failed",
	"validating":        "Validating bridge availability...",
	"validation_result": "Result: %d available, %d unavailable",

	"export_success": "Export successful",
	"export_failed":  "Export failed",
	"exporting":      "Exporting data...",
	"export_file":    "Exported to: %s",

	"stats_title":       "Statistics",
	"stats_total":       "Total bridges",
	"stats_available":   "Available bridges",
	"stats_unavailable": "Unavailable bridges",
	"stats_unknown":     "Unknown bridges",
	"stats_avg_time":    "Average response time",
	"stats_last_fetch":  "Last fetch",

	"error_network":  "Network error",
	"error_parse":    "Parse error",
	"error_database": "Database error",
	"error_timeout":  "Connection timeout",
	"error_proxy":    "Proxy error",

	"bridge_available":   "Available",
	"bridge_unavailable": "Unavailable",
	"bridge_unknown":     "Unknown",

	"query_success": "Query completed",
	"query_failed":  "Query failed",
	"querying":      "Querying bridges...",

	"import_success":     "Import completed",
	"import_failed":      "Import failed",
	"import_file_failed": "Import file failed",
	"importing":          "Importing data...",
	"import_total":       "Total in file",
	"import_imported":    "Imported",
	"import_skipped":     "Skipped (duplicate)",

	"cmd_init":     "Initialize config and database",
	"cmd_fetch":    "Fetch bridge data",
	"cmd_validate": "Validate bridge availability",
	"cmd_export":   "Export bridge data",
	"cmd_stats":    "Show statistics",
	"cmd_query":    "Query bridges with filters",
	"cmd_import":   "Import bridges from file",

	"flag_config":  "Config file path",
	"flag_lang":    "Language (en/zh)",
	"flag_proxy":   "Proxy server address",
	"flag_timeout": "Timeout (seconds)",
	"flag_workers": "Concurrent workers",
	"flag_format":  "Output format (torrc/json/all)",
	"flag_output":  "Output directory",
	"flag_period":  "Stats period (day/week/month)",

	"help_init":     "Run init command to initialize config and database",
	"help_fetch":    "Run fetch command to collect bridges from Tor servers",
	"help_validate": "Run validate command to test bridge connectivity",
	"help_export":   "Run export command to export bridge data",
	"help_stats":    "Run stats command to view statistics",
}
