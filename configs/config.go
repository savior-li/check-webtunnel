package configs

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Database DatabaseConfig `yaml:"database"`
	Fetch    FetchConfig    `yaml:"fetch"`
	Validate ValidateConfig `yaml:"validate"`
	Proxy    ProxyConfig    `yaml:"proxy"`
	Export   ExportConfig   `yaml:"export"`
	App      AppConfig      `yaml:"app"`
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}

type FetchConfig struct {
	URL     string `yaml:"url"`
	Timeout int    `yaml:"timeout"`
}

type ValidateConfig struct {
	Timeout int `yaml:"timeout"`
	Workers int `yaml:"workers"`
}

type ProxyConfig struct {
	Enabled bool    `yaml:"enabled"`
	Proxies []Proxy `yaml:"proxies"`
}

type Proxy struct {
	Type    string `yaml:"type"`
	Address string `yaml:"address"`
	Port    int    `yaml:"port"`
}

type ExportConfig struct {
	OutputDir string `yaml:"output_dir"`
}

type AppConfig struct {
	Lang     string `yaml:"lang"`
	LogLevel string `yaml:"log_level"`
	Debug    bool   `yaml:"debug"`
}

func DefaultConfig() *Config {
	return &Config{
		Database: DatabaseConfig{
			Path: "./bridges.db",
		},
		Fetch: FetchConfig{
			URL:     "https://bridges.torproject.org/bridges?transport=webtunnel",
			Timeout: 30,
		},
		Validate: ValidateConfig{
			Timeout: 10,
			Workers: 5,
		},
		Proxy: ProxyConfig{
			Enabled: false,
			Proxies: []Proxy{},
		},
		Export: ExportConfig{
			OutputDir: "./output",
		},
		App: AppConfig{
			Lang:     "zh",
			LogLevel: "info",
		},
	}
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func Save(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
