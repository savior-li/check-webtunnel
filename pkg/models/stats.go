package models

type Stats struct {
	TotalBridges     int64   `json:"total_bridges"`
	ValidBridges     int64   `json:"valid_bridges"`
	InvalidBridges   int64   `json:"invalid_bridges"`
	AvgLatency       float64 `json:"avg_latency"`
	SuccessRate      float64 `json:"success_rate"`
	TotalValidations int64   `json:"total_validations"`
	LastFetchTime    string  `json:"last_fetch_time,omitempty"`
	LastValidateTime string  `json:"last_validate_time,omitempty"`
}

type BridgeFilter struct {
	Transport  string  `form:"transport"`
	IsValid    *bool   `form:"is_valid"`
	MinLatency float64 `form:"min_latency"`
	MaxLatency float64 `form:"max_latency"`
	Page       int     `form:"page"`
	PageSize   int     `form:"page_size"`
}

type ExportFormat string

const (
	FormatTorrc ExportFormat = "torrc"
	FormatJSON  ExportFormat = "json"
	FormatCSV   ExportFormat = "csv"
)
