package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	App        AppConfig        `mapstructure:"app"`
	Server     ServerConfig     `mapstructure:"server"`
	Proxy      ProxyConfig      `mapstructure:"proxy"`
	Fetch      FetchConfig      `mapstructure:"fetch"`
	Validation ValidationConfig `mapstructure:"validation"`
	Export     ExportConfig     `mapstructure:"export"`
}

type AppConfig struct {
	Language string `mapstructure:"language"`
	DBPath   string `mapstructure:"db_path"`
	LogLevel string `mapstructure:"log_level"`
	LogFile  string `mapstructure:"log_file"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type ProxyConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Type     string `mapstructure:"type"`
	Address  string `mapstructure:"address"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

type FetchConfig struct {
	URL      string `mapstructure:"url"`
	Interval int    `mapstructure:"interval"`
	Timeout  int    `mapstructure:"timeout"`
}

type ValidationConfig struct {
	Timeout     int `mapstructure:"timeout"`
	Concurrency int `mapstructure:"concurrency"`
	Retry       int `mapstructure:"retry"`
}

type ExportConfig struct {
	TorrcPath string `mapstructure:"torrc_path"`
	JSONPath  string `mapstructure:"json_path"`
	CSVPath   string `mapstructure:"csv_path"`
}

var cfg *Config

func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	cfg = &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return cfg, nil
}

func LoadDefault() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.tor-bridge-collector")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	cfg = &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return cfg, nil
}

func Get() *Config {
	return cfg
}

func InitConfig(force bool) error {
	defaultConfig := `app:
  language: "en"
  db_path: "./data/bridges.db"
  log_level: "info"
  log_file: "./logs/app.log"

server:
  host: "0.0.0.0"
  port: 8080

proxy:
  enabled: false
  type: "http"
  address: ""
  port: 0
  username: ""
  password: ""

fetch:
  url: "https://bridges.torproject.org/bridges?transport=webtunnel"
  interval: 3600
  timeout: 30

validation:
  timeout: 10
  concurrency: 5
  retry: 2

export:
  torrc_path: "./output/bridges.txt"
  json_path: "./output/bridges.json"
  csv_path: "./output/bridges.csv"
`

	configPath := "config.yaml"
	if _, err := os.Stat(configPath); err == nil && !force {
		return fmt.Errorf("config.yaml already exists, use --force to overwrite")
	}

	if err := os.WriteFile(configPath, []byte(defaultConfig), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
