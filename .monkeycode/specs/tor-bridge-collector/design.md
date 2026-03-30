# Tor Bridge Collector 技术设计文档

## 1. 项目结构

```
tor-bridge-collector/
├── cmd/
│   └── server/
│       └── main.go           # 程序入口
├── pkg/
│   ├── bridge/
│   │   ├── fetcher.go        # 数据采集
│   │   ├── parser.go         # 数据解析
│   │   └── types.go          # 桥梁数据结构
│   ├── proxy/
│   │   ├── handler.go        # 代理处理
│   │   └── types.go          # 代理类型定义
│   ├── database/
│   │   ├── db.go             # 数据库连接和初始化
│   │   ├── bridge_repo.go    # 桥梁数据仓库
│   │   └── history_repo.go   # 历史记录仓库
│   ├── validator/
│   │   ├── checker.go        # 连接验证
│   │   └── types.go          # 验证结果类型
│   ├── statistics/
│   │   ├── realtime.go       # 实时统计
│   │   └── aggregator.go     # 历史聚合
│   ├── exporter/
│   │   ├── torrc.go          # torrc 格式导出
│   │   ├── json.go           # JSON 格式导出
│   │   └── types.go          # 导出类型定义
│   └── i18n/
│       ├── i18n.go           # 国际化支持
│       ├── zh.go             # 中文翻译
│       └── en.go             # 英文翻译
├── configs/
│   └── config.go            # 配置管理
├── scripts/
│   └── build.sh             # 构建脚本
├── docs/
│   └── ...                   # 文档目录
├── build/
│   └── ...                   # 构建输出
├── go.mod
├── go.sum
└── README.md
```

---

## 2. 模块设计

### 2.1 bridge 模块

#### 数据结构 (types.go)
```go
type Bridge struct {
    ID           int64
    Hash         string    // 唯一标识 hash(address+port+transport)
    Transport    string    // 传输类型: webtunnel
    Address      string    // IP 或域名
    Port         int       // 端口
    Fingerprint  string    // 指纹 (可选)
    DiscoveredAt time.Time // 发现时间
    LastValidAt  time.Time // 最后验证时间
    IsAvailable  int       // -1:未知, 0:不可用, 1:可用
    ResponseTime  int       // 响应时间(ms)
}

type BridgeResponse struct {
    Bridges []Bridge
    FetchedAt time.Time
}
```

#### 数据采集 (fetcher.go)
- 使用 `net/http` 发起 HTTP GET 请求
- 支持自定义 HTTP Transport
- 解析 HTML 响应提取桥梁信息
- 错误处理和重试机制

#### 数据解析 (parser.go)
- 解析 Tor Project 桥梁页面 HTML
- 使用正则表达式提取 bridge line
- 格式: `webtunnel ip:port [fingerprint]`

### 2.2 proxy 模块

#### 代理类型 (types.go)
```go
type Proxy struct {
    Type    string // http, https, socks5
    Address string
    Port    int
}

type ProxyConfig struct {
    Proxies []Proxy
    Index   int // 当前使用索引
    Mu      sync.Mutex
}
```

#### 代理处理 (handler.go)
- 支持 HTTP/HTTPS/SOCKS5 代理
- Round-robin 代理轮询
- 代理连接测试

### 2.3 database 模块

#### 数据库连接 (db.go)
```go
type DB struct {
    *sql.DB
}

func New(dbPath string) (*DB, error)
func (db *DB) InitSchema() error
```

#### 桥梁仓库 (bridge_repo.go)
```go
type BridgeRepository struct {
    db *DB
}

func (r *BridgeRepository) Insert(bridge *Bridge) (int64, error)
func (r *BridgeRepository) Upsert(bridge *Bridge) (int64, error)
func (r *BridgeRepository) GetAll() ([]Bridge, error)
func (r *BridgeRepository) GetByHash(hash string) (*Bridge, error)
func (r *BridgeRepository) UpdateAvailability(id int64, available bool, responseTime int) error
func (r *BridgeRepository) DeleteOld(days int) error
func (r *BridgeRepository) Count() (int, error)
func (r *BridgeRepository) CountAvailable() (int, error)
```

#### 历史记录仓库 (history_repo.go)
```go
type HistoryRepository struct {
    db *DB
}

func (r *HistoryRepository) Insert(record *ValidationRecord) error
func (r *HistoryRepository) GetByBridgeID(bridgeID int64, limit int) ([]ValidationRecord, error)
func (r *HistoryRepository) GetStatsByPeriod(period string) ([]DailyStats, error)
```

### 2.4 validator 模块

#### 验证器 (checker.go)
```go
type Validator struct {
    timeout  time.Duration
    workers  int
}

func (v *Validator) Validate(bridge *Bridge) (*ValidationResult, error)
func (v *Validator) ValidateAll(bridges []Bridge) ([]ValidationResult, error)
```

#### 验证结果 (types.go)
```go
type ValidationResult struct {
    BridgeID     int64
    IsAvailable  bool
    ResponseTime int // ms
    Error        error
    ValidatedAt  time.Time
}
```

### 2.5 statistics 模块

#### 实时统计 (realtime.go)
```go
type RealtimeStats struct {
    TotalBridges   int
    AvailableBridges int
    UnavailableBridges int
    UnknownBridges  int
    AvgResponseTime float64
    LastFetchTime   time.Time
}

func GetRealtimeStats(db *database.DB) (*RealtimeStats, error)
```

#### 历史聚合 (aggregator.go)
```go
type DailyStats struct {
    Date          string
    TotalCount    int
    AvailableCount int
    AvgResponseTime float64
}

func GetDailyStats(db *database.DB, days int) ([]DailyStats, error)
func GetWeeklyStats(db *database.DB, weeks int) ([]DailyStats, error)
func GetMonthlyStats(db *database.DB, months int) ([]DailyStats, error)
```

### 2.6 exporter 模块

#### torrc 导出 (torrc.go)
```go
type TorrcExporter struct {
    outputDir string
}

func (e *TorrcExporter) Export(bridges []Bridge) error
func formatBridgeLine(bridge *Bridge) string
```

#### JSON 导出 (json.go)
```go
type JSONExporter struct {
    outputDir string
}

type ExportData struct {
    ExportedAt  time.Time `json:"exported_at"`
    TotalCount  int        `json:"total_count"`
    Bridges     []Bridge   `json:"bridges"`
}

func (e *JSONExporter) Export(bridges []Bridge) error
```

### 2.7 i18n 模块

#### 国际化 (i18n.go)
```go
type Translator struct {
    lang string
    t    map[string]string
}

func NewTranslator(lang string) *Translator
func (t *Translator) T(key string) string
func (t *Translator) SetLang(lang string)
```

---

## 3. CLI 设计

使用 `github.com/urfave/cli/v2` 构建命令行应用。

### 3.1 命令结构
```
tor-bridge-collector (root)
├── init      # 初始化
├── fetch     # 采集
├── validate  # 验证
├── export    # 导出
└── stats     # 统计
```

### 3.2 全局标志
```go
var (
    configFlag = cli.StringFlag{
        Name:    "config",
        Value:   "config.yaml",
        Usage:   "配置文件路径",
    }
    langFlag = cli.StringFlag{
        Name:    "lang",
        Value:   "zh",
        Usage:   "语言: en/zh",
    }
)
```

---

## 4. 配置设计

### 4.1 config.yaml 结构
```yaml
database:
  path: "./bridges.db"

fetch:
  url: "https://bridges.torproject.org/bridges?transport=webtunnel"
  timeout: 30

validate:
  timeout: 10
  workers: 5

proxy:
  enabled: false
  proxies:
    - type: "http"
      address: "127.0.0.1"
      port: 7890

export:
  output_dir: "./output"

app:
  lang: "zh"
  log_level: "info"
```

---

## 5. 构建设计

### 5.1 构建脚本 (build.sh)
```bash
#!/bin/bash
PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "windows/amd64"
    "darwin/amd64"
    "darwin/arm64"
)

for PLATFORM in "${PLATFORMS[@]}"; do
    GOOS=${PLATFORM%/*}
    GOARCH=${PLATFORM#*/}
    OUTPUT="tor-bridge-collector-${GOOS}-${GOARCH}"
    [ "$GOOS" == "windows" ] && OUTPUT="${OUTPUT}.exe"
    
    GOOS=$GOOS GOARCH=$GOARCH go build -o "./build/$OUTPUT" ./cmd/server
done
```

### 5.2 go.mod 依赖
```
module tor-bridge-collector

go 1.21

require (
    github.com/mattn/go-sqlite3 v1.14.18
    github.com/urfave/cli/v2 v2.27.1
    gopkg.in/yaml.v3 v3.0.1
)
```

---

## 6. 数据库 Schema

### 6.1 bridges 表
```sql
CREATE TABLE IF NOT EXISTS bridges (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    hash TEXT NOT NULL UNIQUE,
    transport TEXT NOT NULL DEFAULT 'webtunnel',
    address TEXT NOT NULL,
    port INTEGER NOT NULL,
    fingerprint TEXT,
    discovered_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_validated DATETIME,
    is_available INTEGER DEFAULT -1,
    response_time_ms INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_bridges_hash ON bridges(hash);
CREATE INDEX IF NOT EXISTS idx_bridges_available ON bridges(is_available);
CREATE INDEX IF NOT EXISTS idx_bridges_discovered ON bridges(discovered_at);
```

### 6.2 validation_history 表
```sql
CREATE TABLE IF NOT EXISTS validation_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    bridge_id INTEGER NOT NULL,
    validated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    is_available INTEGER NOT NULL,
    response_time_ms INTEGER,
    error_message TEXT,
    FOREIGN KEY (bridge_id) REFERENCES bridges(id)
);

CREATE INDEX IF NOT EXISTS idx_history_bridge ON validation_history(bridge_id);
CREATE INDEX IF NOT EXISTS idx_history_validated ON validation_history(validated_at);
```

---

## 7. 错误处理策略

### 7.1 错误分类
- 网络错误: 重试 + 代理切换
- 解析错误: 记录日志，跳过该条
- 数据库错误: 回滚事务，返回错误
- 验证超时: 标记为不可用，继续下一个

### 7.2 日志级别
- ERROR: 操作失败
- WARN: 潜在问题
- INFO: 正常操作
- DEBUG: 调试信息

---

## 8. 安全性考虑

- 不记录敏感信息到日志
- 数据库文件权限设置为 600
- 配置文件不包含明文密码 (使用占位符)
