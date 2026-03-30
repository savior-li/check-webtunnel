# Tor Bridge Collector 需求规格说明书

## 1. 项目概述

### 1.1 项目名称
Tor Bridge Collector

### 1.2 项目类型
命令行工具 (CLI)

### 1.3 核心功能概述
从 Tor Project 官方桥梁服务器采集 webtunnel 类型的桥接信息，支持代理连接、有效性验证、数据持久化、统计分析和多格式导出。

### 1.4 目标用户
- Tor 网络用户需要获取最新桥接信息
- 网络安全研究人员
- 需要绕过网络限制的用户

---

## 2. 功能需求

### 2.1 数据采集

| 需求ID | 需求描述 | 优先级 | 类别 |
|--------|----------|--------|------|
| FR-001 | 从 `https://bridges.torproject.org/bridges?transport=webtunnel` 获取 webtunnel 桥梁信息 | 必须 | 数据采集 |
| FR-002 | 支持通过 HTTP/HTTPS 代理服务器发起请求 | 必须 | 代理支持 |
| FR-003 | 支持 SOCKS5 代理协议 | 必须 | 代理支持 |
| FR-004 | 支持配置多个代理服务器进行轮询 | 可选 | 代理支持 |

### 2.2 数据持久化

| 需求ID | 需求描述 | 优先级 | 类别 |
|--------|----------|--------|------|
| FR-005 | 使用 SQLite 数据库存储桥梁历史记录 | 必须 | 数据持久化 |
| FR-006 | 支持桥梁信息去重 (基于 hash) | 必须 | 数据持久化 |
| FR-007 | 记录每次采集的时间戳 | 必须 | 数据持久化 |
| FR-008 | 支持历史数据查询和追溯 | 必须 | 数据持久化 |

### 2.3 有效性验证

| 需求ID | 需求描述 | 优先级 | 类别 |
|--------|----------|--------|------|
| FR-009 | 对采集到的 bridge 进行 TCP 连接测试 | 必须 | 有效性验证 |
| FR-010 | 记录每个 bridge 的响应延迟 (ms) | 必须 | 有效性验证 |
| FR-011 | 标记 bridge 的可用性状态 (可用/不可用/未知) | 必须 | 有效性验证 |
| FR-012 | 支持并发验证以提高效率 | 可选 | 有效性验证 |

### 2.4 统计分析

| 需求ID | 需求描述 | 优先级 | 类别 |
|--------|----------|--------|------|
| FR-013 | 实时统计：当前桥梁总数、可用率、平均响应时间 | 必须 | 统计分析 |
| FR-014 | 历史聚合：按日/周/月统计桥梁数量变化趋势 | 必须 | 统计分析 |
| FR-015 | 输出统计报表 (控制台输出) | 必须 | 统计分析 |

### 2.5 多格式输出

| 需求ID | 需求描述 | 优先级 | 类别 |
|--------|----------|--------|------|
| FR-016 | 输出 torrc 格式的原始桥梁信息 (可直接用于 Tor 配置) | 必须 | 输出 |
| FR-017 | 输出 JSON 格式的完整桥梁数据 | 必须 | 输出 |
| FR-018 | 按类型分类输出到不同文件 | 可选 | 输出 |
| FR-019 | 支持指定输出目录 | 必须 | 输出 |

### 2.6 初始化

| 需求ID | 需求描述 | 优先级 | 类别 |
|--------|----------|--------|------|
| FR-020 | 初始化命令创建默认配置文件 `config.yaml` | 必须 | 初始化 |
| FR-021 | 初始化命令创建 SQLite 数据库文件 `bridges.db` | 必须 | 初始化 |
| FR-022 | 配置文件支持自定义数据库路径、代理设置、语言偏好 | 必须 | 初始化 |

### 2.7 用户界面

| 需求ID | 需求描述 | 优先级 | 类别 |
|--------|----------|--------|------|
| FR-023 | 支持中英文双语界面 | 必须 | 用户界面 |
| FR-024 | 中文为默认语言 | 可选 | 用户界面 |
| FR-025 | 通过配置文件或命令行参数切换语言 | 必须 | 用户界面 |

---

## 3. 技术需求

### 3.1 编程语言
- **Golang** (版本 >= 1.21)

### 3.2 数据库
- **SQLite3** (使用 go-sqlite3 驱动)

### 3.3 外部依赖
- 无特殊外部依赖
- 标准库为主

### 3.4 目标平台
- Linux (amd64, arm64)
- Windows (amd64)
- macOS/Darwin (amd64, arm64)

---

## 4. 非功能需求

### 4.1 性能需求
- 单一桥梁验证超时时间: 10 秒 (可配置)
- 并发验证数量: 5 (可配置)
- 数据库查询响应时间: < 100ms

### 4.2 可用性需求
- 程序应能在网络受限环境下运行 (通过代理)
- 初始化失败时应给出明确错误提示

### 4.3 可维护性需求
- 模块化设计，便于维护和扩展
- 配置文件使用 YAML 格式，便于编辑

---

## 5. 命令行接口

### 5.1 主命令
```
tor-bridge-collector [global options] command [command options]
```

### 5.2 全局选项
| 选项 | 描述 | 默认值 |
|------|------|--------|
| --config | 配置文件路径 | ./config.yaml |
| --lang | 语言 (en/zh) | zh |

### 5.3 子命令

#### init
初始化配置文件和数据库
```
tor-bridge-collector init
```

#### fetch
采集桥梁数据
```
tor-bridge-collector fetch [options]
```
| 选项 | 描述 | 默认值 |
|------|------|--------|
| --proxy | 代理服务器地址 | 空 |

#### validate
验证桥梁有效性
```
tor-bridge-collector validate [options]
```
| 选项 | 描述 | 默认值 |
|------|------|--------|
| --timeout | 超时时间(秒) | 10 |
| --workers | 并发数 | 5 |

#### export
导出桥梁数据
```
tor-bridge-collector export [options]
```
| 选项 | 描述 | 默认值 |
|------|------|--------|
| --format | 输出格式 (torrc/json/all) | torrc |
| --output | 输出目录 | ./output |

#### stats
显示统计信息
```
tor-bridge-collector stats [options]
```
| 选项 | 描述 | 默认值 |
|------|------|--------|
| --period | 统计周期 (day/week/month) | day |

---

## 6. 数据库 Schema

### 6.1 bridges 表
```sql
CREATE TABLE bridges (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    hash TEXT NOT NULL UNIQUE,
    transport TEXT NOT NULL,
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
```

### 6.2 validation_history 表
```sql
CREATE TABLE validation_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    bridge_id INTEGER NOT NULL,
    validated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    is_available INTEGER NOT NULL,
    response_time_ms INTEGER,
    error_message TEXT,
    FOREIGN KEY (bridge_id) REFERENCES bridges(id)
);
```

---

## 7. 验收标准

### 7.1 功能验收
- [ ] `init` 命令能正确创建 config.yaml 和 bridges.db
- [ ] `fetch` 命令能从指定 URL 获取 webtunnel 桥梁信息
- [ ] `fetch --proxy` 能通过代理服务器获取数据
- [ ] 采集的数据能正确存储到 SQLite 数据库
- [ ] 重复的桥梁信息不会重复插入 (去重)
- [ ] `validate` 命令能测试桥梁连通性
- [ ] `validate` 命令能记录响应延迟
- [ ] `export --format torrc` 输出能被 Tor 直接使用
- [ ] `export --format json` 输出完整桥梁数据
- [ ] `stats` 命令能显示实时统计
- [ ] `stats --period week/month` 能显示历史聚合
- [ ] 所有命令支持中英文切换

### 7.2 技术验收
- [ ] 代码能成功编译为 Linux amd64 二进制
- [ ] 代码能成功编译为 Windows amd64 二进制
- [ ] 代码能成功编译为 Darwin amd64/arm64 二进制
- [ ] 无外部依赖 (纯标准库 + go-sqlite3)
- [ ] 数据库 schema 符合设计
- [ ] 配置文件格式为 YAML

---

## 8. 术语表

| 术语 | 定义 |
|------|------|
| Bridge | Tor 桥接节点，用于绕过网络封锁 |
| WebTunnel | 一种 Tor 桥接传输类型 |
| torrc | Tor 配置文件 |
| Fingerprint | 桥接节点的唯一标识 |
| Proxy | 代理服务器 |
| SOCKS5 | 一种代理协议 |

---

## 9. 变更历史

| 版本 | 日期 | 描述 |
|------|------|------|
| 1.0.0 | 2026-03-30 | 初始版本 |
