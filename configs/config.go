package configs

import (
	"fmt"
	"os"
	"strings"

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

func (p *Proxy) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var raw string
	if err := unmarshal(&raw); err != nil {
		return err
	}

	if strings.HasPrefix(raw, "socks5h://") {
		p.Type = "socks5h"
		raw = strings.TrimPrefix(raw, "socks5h://")
	} else if strings.HasPrefix(raw, "socks5://") {
		p.Type = "socks5"
		raw = strings.TrimPrefix(raw, "socks5://")
	} else if strings.HasPrefix(raw, "socks://") {
		p.Type = "socks5"
		raw = strings.TrimPrefix(raw, "socks://")
	} else if strings.HasPrefix(raw, "http://") {
		p.Type = "http"
		raw = strings.TrimPrefix(raw, "http://")
	} else if strings.HasPrefix(raw, "https://") {
		p.Type = "https"
		raw = strings.TrimPrefix(raw, "https://")
	} else {
		p.Type = "http"
	}

	parts := strings.Split(raw, ":")
	if len(parts) == 2 {
		p.Address = parts[0]
		fmt.Sscanf(parts[1], "%d", &p.Port)
	}

	return nil
}

type ExportConfig struct {
	OutputDir string `yaml:"output_dir"`
}

type AppConfig struct {
	Lang     string `yaml:"lang"`
	LogLevel string `yaml:"log_level"`
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
