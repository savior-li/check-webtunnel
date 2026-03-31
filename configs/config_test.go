package configs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.NotNil(t, cfg)
	assert.Equal(t, "./bridges.db", cfg.Database.Path)
	assert.Equal(t, "https://bridges.torproject.org/bridges?transport=webtunnel", cfg.Fetch.URL)
	assert.Equal(t, 30, cfg.Fetch.Timeout)
	assert.Equal(t, 10, cfg.Validate.Timeout)
	assert.Equal(t, 5, cfg.Validate.Workers)
	assert.Equal(t, "./output", cfg.Export.OutputDir)
	assert.Equal(t, "zh", cfg.App.Lang)
	assert.Equal(t, "info", cfg.App.LogLevel)
}

func TestLoad_Valid(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	content := `
database:
  path: "./test.db"
fetch:
  url: "https://example.com/bridges"
  timeout: 60
validate:
  timeout: 20
  workers: 10
export:
  output_dir: "./test_output"
app:
  lang: "en"
  log_level: "debug"
`

	err := os.WriteFile(configPath, []byte(content), 0644)
	assert.NoError(t, err)

	cfg, err := Load(configPath)

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "./test.db", cfg.Database.Path)
	assert.Equal(t, "https://example.com/bridges", cfg.Fetch.URL)
	assert.Equal(t, 60, cfg.Fetch.Timeout)
	assert.Equal(t, 20, cfg.Validate.Timeout)
	assert.Equal(t, 10, cfg.Validate.Workers)
	assert.Equal(t, "./test_output", cfg.Export.OutputDir)
	assert.Equal(t, "en", cfg.App.Lang)
	assert.Equal(t, "debug", cfg.App.LogLevel)
}

func TestLoad_NotExist(t *testing.T) {
	cfg, err := Load("/nonexistent/path/config.yaml")

	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestLoad_InvalidYAML(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	content := `
database:
  path: "./test.db
  invalid yaml content
`

	err := os.WriteFile(configPath, []byte(content), 0644)
	assert.NoError(t, err)

	cfg, err := Load(configPath)

	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestSave(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	cfg := &Config{
		Database: DatabaseConfig{
			Path: "./test.db",
		},
		Fetch: FetchConfig{
			URL:     "https://example.com",
			Timeout: 30,
		},
		Validate: ValidateConfig{
			Timeout: 10,
			Workers: 5,
		},
		Export: ExportConfig{
			OutputDir: "./output",
		},
		App: AppConfig{
			Lang:     "zh",
			LogLevel: "info",
		},
	}

	err := Save(configPath, cfg)
	assert.NoError(t, err)

	assert.FileExists(t, configPath)

	loadedCfg, err := Load(configPath)
	assert.NoError(t, err)
	assert.Equal(t, cfg.Database.Path, loadedCfg.Database.Path)
	assert.Equal(t, cfg.Fetch.URL, loadedCfg.Fetch.URL)
	assert.Equal(t, cfg.Fetch.Timeout, loadedCfg.Fetch.Timeout)
}

func TestConfig_Structure(t *testing.T) {
	cfg := DefaultConfig()

	assert.NotNil(t, cfg)
	assert.Equal(t, DatabaseConfig{Path: "./bridges.db"}, cfg.Database)
	assert.Equal(t, FetchConfig{URL: "https://bridges.torproject.org/bridges?transport=webtunnel", Timeout: 30}, cfg.Fetch)
	assert.Equal(t, ValidateConfig{Timeout: 10, Workers: 5}, cfg.Validate)
	assert.NotNil(t, cfg.Proxy)
	assert.Equal(t, ExportConfig{OutputDir: "./output"}, cfg.Export)
	assert.Equal(t, AppConfig{Lang: "zh", LogLevel: "info"}, cfg.App)
}

func TestDatabaseConfig(t *testing.T) {
	cfg := DatabaseConfig{Path: "./test.db"}
	assert.Equal(t, "./test.db", cfg.Path)
}

func TestFetchConfig(t *testing.T) {
	cfg := FetchConfig{URL: "https://example.com", Timeout: 30}
	assert.Equal(t, "https://example.com", cfg.URL)
	assert.Equal(t, 30, cfg.Timeout)
}

func TestValidateConfig(t *testing.T) {
	cfg := ValidateConfig{Timeout: 10, Workers: 5}
	assert.Equal(t, 10, cfg.Timeout)
	assert.Equal(t, 5, cfg.Workers)
}

func TestProxyConfig(t *testing.T) {
	cfg := ProxyConfig{
		Enabled: true,
		Proxies: []Proxy{
			{Type: "http", Address: "127.0.0.1", Port: 8080},
		},
	}
	assert.True(t, cfg.Enabled)
	assert.Len(t, cfg.Proxies, 1)
}

func TestProxy_Structure(t *testing.T) {
	p := Proxy{Type: "http", Address: "127.0.0.1", Port: 8080}
	assert.Equal(t, "http", p.Type)
	assert.Equal(t, "127.0.0.1", p.Address)
	assert.Equal(t, 8080, p.Port)
}

func TestExportConfig(t *testing.T) {
	cfg := ExportConfig{OutputDir: "./output"}
	assert.Equal(t, "./output", cfg.OutputDir)
}

func TestAppConfig(t *testing.T) {
	cfg := AppConfig{Lang: "zh", LogLevel: "info"}
	assert.Equal(t, "zh", cfg.Lang)
	assert.Equal(t, "info", cfg.LogLevel)
}
