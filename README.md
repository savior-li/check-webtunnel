# Tor Bridge Collector

Tor 桥接节点采集工具 - 用于从 Tor Project 官方服务器采集 webtunnel 类型的桥接节点信息。

## 功能特性

- 数据采集：从 Tor Project 官方桥梁服务器获取 webtunnel 信息
- 代理支持：支持 HTTP/HTTPS/SOCKS5 代理服务器
- 数据持久化：SQLite 存储历史记录，支持去重和历史追溯
- 有效性验证：对 bridge 进行连接测试，记录响应速度
- 统计分析：实时统计 + 历史聚合分析
- 多格式输出：torrc 可用原始文本、JSON、分类文件
- 中英文双语界面

## 系统要求

- Golang >= 1.21
- SQLite3

## 安装

### 从源码编译

```bash
git clone <repository-url>
cd tor-bridge-collector
go mod download
./scripts/build.sh
```

### 使用预编译二进制

从 `build/` 目录下载对应平台的二进制文件。

## 快速开始

### 1. 初始化

```bash
./tor-bridge-collector init
```

这将创建：
- `config.yaml` - 配置文件
- `bridges.db` - SQLite 数据库

### 2. 采集桥梁数据

```bash
./tor-bridge-collector fetch
```

使用代理：

```bash
./tor-bridge-collector fetch --proxy http://127.0.0.1:7890
```

### 3. 验证桥梁可用性

```bash
./tor-bridge-collector validate
```

自定义超时和并发数：

```bash
./tor-bridge-collector validate --timeout 15 --workers 10
```

### 4. 导出数据

导出为 torrc 格式（可直接用于 Tor 配置）：

```bash
./tor-bridge-collector export --format torrc --output ./output
```

导出为 JSON 格式：

```bash
./tor-bridge-collector export --format json --output ./output
```

### 5. 查看统计信息

```bash
./tor-bridge-collector stats
./tor-bridge-collector stats --period week
```

## Debug 模式

启用 debug 模式可以查看详细的执行步骤日志：

```bash
./tor-bridge-collector -d fetch
./tor-bridge-collector -d fetch --proxy socks5://127.0.0.1:1080
./tor-bridge-collector -d validate --timeout 5
./tor-bridge-collector -d stats
```

Debug 输出示例：
```
[DEBUG] [14:57:53] Starting fetch action
[DEBUG] [14:57:53] Fetch URL: https://bridges.torproject.org/bridges?transport=webtunnel
[DEBUG] [14:57:53] Using proxy: socks5://127.0.0.1:1080
[DEBUG] [14:57:55] Received 2 bridges from server
[DEBUG] [14:57:55] Processing bridge: 192.168.1.1:443
[DEBUG] [14:57:55] New bridge inserted with ID: 1
```

## 命令行选项

### 全局选项

| 选项 | 描述 | 默认值 |
|------|------|--------|
| `--config, -c` | 配置文件路径 | config.yaml |
| `--lang, -l` | 语言 (en/zh) | zh |
| `--debug, -d` | 启用 debug 模式 | false |

### 子命令

#### init
初始化配置文件和数据库。

```
./tor-bridge-collector init
```

#### fetch
采集桥梁数据。

| 选项 | 描述 | 默认值 |
|------|------|--------|
| `--proxy, -p` | 代理服务器地址 | 空 |

#### validate
验证桥梁可用性。

| 选项 | 描述 | 默认值 |
|------|------|--------|
| `--timeout, -t` | 超时时间(秒) | 10 |
| `--workers, -w` | 并发数 | 5 |

#### export
导出桥梁数据。

| 选项 | 描述 | 默认值 |
|------|------|--------|
| `--format, -f` | 输出格式 (torrc/json/all) | torrc |
| `--output, -o` | 输出目录 | ./output |

#### stats
显示统计信息。

| 选项 | 描述 | 默认值 |
|------|------|--------|
| `--period, -p` | 统计周期 (day/week/month) | day |

## 配置文件

配置文件 `config.yaml` 示例：

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
  debug: false
```

## 数据库 Schema

### bridges 表

| 字段 | 类型 | 描述 |
|------|------|------|
| id | INTEGER | 主键 |
| hash | TEXT | 唯一标识 (address+port+transport) |
| transport | TEXT | 传输类型 (webtunnel) |
| address | TEXT | IP 或域名 |
| port | INTEGER | 端口 |
| fingerprint | TEXT | 指纹 (可选) |
| discovered_at | DATETIME | 发现时间 |
| last_validated | DATETIME | 最后验证时间 |
| is_available | INTEGER | 可用性 (-1:未知, 0:不可用, 1:可用) |
| response_time_ms | INTEGER | 响应时间(毫秒) |

### validation_history 表

| 字段 | 类型 | 描述 |
|------|------|------|
| id | INTEGER | 主键 |
| bridge_id | INTEGER | 关联的 bridge ID |
| validated_at | DATETIME | 验证时间 |
| is_available | INTEGER | 是否可用 |
| response_time_ms | INTEGER | 响应时间 |
| error_message | TEXT | 错误信息 |

## 输出示例

### torrc 格式输出

```
# Tor Bridge Collector Export
# Exported at: 2026-03-30T10:00:00Z
# Total bridges: 5

BridgeRelay 1

Bridge webtunnel 192.168.1.1:443 fingerprint=ABCD1234...
Bridge webtunnel 192.168.1.2:443 fingerprint=EFGH5678...
```

### JSON 格式输出

```json
{
  "exported_at": "2026-03-30T10:00:00Z",
  "total_count": 5,
  "bridges": [
    {
      "id": 1,
      "hash": "abc123...",
      "transport": "webtunnel",
      "address": "192.168.1.1",
      "port": 443,
      "fingerprint": "ABCD1234...",
      "is_available": 1,
      "response_time_ms": 120
    }
  ]
}
```

## 构建多平台二进制

```bash
chmod +x scripts/build.sh
./scripts/build.sh
```

支持的平台：
- Linux (amd64, arm64)
- Windows (amd64)
- macOS/Darwin (amd64, arm64)

## 许可证

MIT License
