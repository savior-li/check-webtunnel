# Tor Bridge Collector

A Golang-based tool for fetching, validating, and managing Tor webtunnel bridges.

## Features

- **Data Collection**: Fetch webtunnel bridges from Tor Project
- **Proxy Support**: HTTP/HTTPS/SOCKS5 proxy servers supported
- **Data Persistence**: SQLite storage with deduplication and history tracking
- **Validation**: TCP connectivity testing with latency measurement
- **Statistics**: Real-time and historical analytics
- **Export**: Multiple formats (torrc, JSON, CSV)
- **Web UI**: Browser-based management interface
- **Internationalization**: English and Chinese (Simplified) support

## Requirements

- Go 1.21 or later
- SQLite3

## Quick Start

### Initialize

```bash
./tor-bridge-collector init
```

### Fetch Bridges

```bash
./tor-bridge-collector fetch
```

### Validate Bridges

```bash
./tor-bridge-collector validate --all
```

### Export Bridges

```bash
./tor-bridge-collector export --format torrc
./tor-bridge-collector export --format json
./tor-bridge-collector export --format csv
```

### View Statistics

```bash
./tor-bridge-collector stats
```

### Start Web UI

```bash
./tor-bridge-collector serve --port 8080
```

## Configuration

Edit `config.yaml` to customize settings:

```yaml
app:
  language: "en"           # en/zh
  db_path: "./data/bridges.db"
  log_level: "info"

server:
  host: "0.0.0.0"
  port: 8080

proxy:
  enabled: false
  type: "http"             # http/https/socks5
  address: ""
  port: 0

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
```

## CLI Commands

| Command | Description |
|---------|-------------|
| `init` | Initialize config and database |
| `fetch` | Fetch bridges from Tor project |
| `validate` | Validate bridge connectivity |
| `export` | Export bridges to file |
| `stats` | Show statistics |
| `serve` | Start web UI |

## Building

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o tor-bridge-collector-linux-amd64

# Windows
GOOS=windows GOARCH=amd64 go build -o tor-bridge-collector-windows-amd64.exe

# macOS
GOOS=darwin GOARCH=arm64 go build -o tor-bridge-collector-darwin-arm64
```

## Project Structure

```
tor-bridge-collector/
├── cmd/                    # CLI commands
├── internal/
│   ├── config/            # Configuration management
│   ├── fetcher/           # Bridge data fetching
│   ├── validator/         # Bridge validation
│   ├── storage/           # SQLite persistence
│   ├── exporter/          # Data export
│   ├── proxy/             # Proxy support
│   ├── i18n/              # Internationalization
│   └── web/               # Web UI
├── pkg/models/            # Data models
├── config.yaml            # Configuration file
└── main.go                # Entry point
```

## License

MIT License
