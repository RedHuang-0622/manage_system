# 实现方案：Go HTTP Server 资源管理加固，修复 Vite 代理连接僵死

## 设计目标

Go HTTP Server 当前**裸奔**：五个超时参数全部为零，无 Header 大小限制，无连接数上限。导致 CLOSE_WAIT 连接堆积、Vite 代理连接池耗尽。

本次补齐 HTTP Server 层的资源管理缺口。

## 各层资源池现状（改造前）

| 层 | 资源池/限制 | 当前配置 | 状态 |
|----|-----------|---------|------|
| MySQL (GORM) | 连接池 | `MaxIdleConns=10` `MaxOpenConns=100` `ConnMaxLifetime=3600s` | ✅ 完善 |
| Redis (go-redis) | 连接池 | `PoolSize=100` `MinIdleConns=10` `DialTimeout=5s` `ReadTimeout=3s` `WriteTimeout=3s` | ✅ 完善 |
| **Go HTTP Server** | **无连接池概念（goroutine-per-conn）** | **五个超时全为零，无任何限制** | ❌ 裸奔 |
| Vite proxy | Node.js http.Agent | `keepAlive: false`（默认，每次新建连接） | ⚠️ 安全但低效 |

Go `http.Server` 是服务端，没有客户端那种"连接池"，每个连接是一个独立 goroutine。但 `http.Server` 提供了 5 个关键超时参数 + 1 个大小限制，全部缺失：

| 缺失参数 | 作用 | 为 0 时的后果 |
|---------|------|-------------|
| `ReadTimeout` | 读完整请求(Header+Body)的最长时间 | 慢客户端可无限占用连接 |
| `ReadHeaderTimeout` | **只读 Header 阶段**的超时 | Slowloris 攻击：每秒发 1 字节 Header 永不超时 |
| `WriteTimeout` | 写完整响应的最长时间 | 慢客户端可无限挂起响应 |
| `IdleTimeout` | Keep-Alive 空闲连接最大存活时间 | **本次根因**：CLOSE_WAIT 永不过期 |
| `MaxHeaderBytes` | 请求头最大字节数 | Header 炸弹 |

> **为什么 `ReadTimeout` 不能替代 `ReadHeaderTimeout`**：`ReadTimeout` 覆盖 Header + Body 整个过程。Slowloris 攻击者在 Header 阶段每秒发 1 字节，发送 9 秒后开始发 Body，此时 `ReadTimeout: 10s` 还剩 1 秒给 Body，看起来"正常"。`ReadHeaderTimeout: 5s` 独立切断 Header 慢速攻击。

## 设计模式选择

| 模式 | 语言实现 | 应用位置 | 理由 |
|------|---------|---------|------|
| Builder (Functional Options) | 已有 — `setDefaults()` 模式 | `pkg/config/config.go` | 沿用项目既有的默认值注入风格，不引入新范式 |
| Strategy | 硬编码的 `net/http` 参数 | `cmd/main.go` — `http.Server{}` | Go 标准库 `http.Server` 超时配置无抽象必要，直接使用 |

## 方案对比

### 方案 A：硬编码常量（最小改动）

在 `cmd/main.go` 构造 `http.Server` 时直接写入 6 个常量值。

| 维度 | 评价 |
|------|------|
| 耦合度 | 低 — 仅 main.go 一处改动 |
| 内聚性 | 高 — 服务启动参数集中一处 |
| 可测试性 | 无退化 |
| 实现成本 | 极低 — 6 行代码 |
| 改动面 | [main.go:110-113](backend/cmd/main.go#L110-L113) 一处 |
| 可回滚性 | 极好 — 删除 6 行即可 |
| 配置硬度 | 🟠 默认+覆盖 — 常量，不可环境覆盖 |

### 方案 B：配置驱动（推荐）

将超时参数纳入 `ServerConfig` → `config.yaml` → `setDefaults()`，`main.go` 读取配置。符合项目 MySQL/Redis 连接池参数的一致模式。

| 维度 | 评价 |
|------|------|
| 耦合度 | 低 — 仅增 struct 字段 |
| 内聚性 | 高 — 超时与 server 配置同属 `ServerConfig` |
| 可测试性 | 无退化 |
| 实现成本 | 低 — 约 40 行代码，4 个文件 |
| 改动面 | [config.go](backend/pkg/config/config.go) + [config.yaml](backend/conf/config.yaml) + [config.example.yaml](backend/conf/config.example.yaml) + [main.go](backend/cmd/main.go) |
| 可回滚性 | 好 — 删除字段 + 恢复 main.go 即可 |
| 配置硬度 | 🟡 环境变量 — `LAB_SERVER_READ_TIMEOUT` 等可覆盖 |

### 方案 C：方案 B + Vite 代理 keepAlive

在方案 B 基础上，Vite 代理显式开启 `keepAlive` + `maxSockets`，减少 TCP 握手开销。

| 维度 | 评价 |
|------|------|
| 耦合度 | 低 |
| 内聚性 | 中 — 跨前后端协调 |
| 可测试性 | 好 |
| 实现成本 | 中 — +1 文件，~5 行 |
| 改动面 | + [vite.config.ts](frontend/vite.config.ts) |
| 可回滚性 | 好 |
| 额外收益 | 代理层连接复用，减少 TCP 握手 |

## 推荐：方案 B

**理由**：
1. 与项目现有的 `setDefaults()` + YAML 驱动模式完全一致（参见 `MySQLConfig.ConnMaxLifetime`、`RedisConfig.PoolSize` 等）
2. 支持环境变量覆盖，部署灵活
3. `IdleTimeout` 是修复 CLOSE_WAIT 的核心参数；`ReadHeaderTimeout` 是 Slowloris 防御；其余为防御性加成
4. 改动面清晰，4 个文件，无架构变更

**最大风险**：`ReadTimeout` 设得过小会截断慢上传。当前系统无文件上传 API，默认 10s 安全。

## 关于"连接池"

Go `http.Server` 没有连接池 — 它是服务端，每个连接对应一个 goroutine，由 Go runtime 调度。不需要也无法配置"HTTP 连接池"。

真正需要连接池的两层（MySQL、Redis）已经配好了，不需要改动。

Go `http.Server` 唯一与"池"相关的概念是 **`IdleTimeout`**：Keep-Alive 连接在这个时间后被 GC，相当于连接回收的 TTL。这是本次修复的核心参数。

如需限制**并发连接数**（防止 goroutine 爆炸），Go 标准库不直接支持，需要通过 `netutil.LimitListener` 或中间件实现。但实验室管理系统并发量低，暂不需要。

## 循环依赖检查

```
config.go → (无内部依赖，纯数据 struct)
main.go   → config.go（已有）
```

无新增 import 依赖，无循环。

## 核心接口/Protocol 定义

无新增 interface — 本次改动是配置值的传递，不涉及抽象边界。`ServerConfig` 是纯 DTO struct，沿用既有模式。

## 具体实现

### 1. `pkg/config/config.go` — 扩展 ServerConfig

```go
// ServerConfig 服务配置
type ServerConfig struct {
    Port              int    `mapstructure:"port"`
    Mode              string `mapstructure:"mode"`
    ReadTimeout       int    `mapstructure:"read_timeout"`        // 读取完整请求超时(秒)
    ReadHeaderTimeout int    `mapstructure:"read_header_timeout"` // 仅读请求头超时(秒)，防 Slowloris
    WriteTimeout      int    `mapstructure:"write_timeout"`       // 写入响应超时(秒)
    IdleTimeout       int    `mapstructure:"idle_timeout"`        // Keep-Alive 空闲超时(秒)
    MaxHeaderBytes    int    `mapstructure:"max_header_bytes"`    // 请求头最大字节数
}
```

在 `setDefaults()` 中追加：

```go
if cfg.Server.ReadTimeout == 0 {
    cfg.Server.ReadTimeout = 10       // 10s 覆盖所有 API
}
if cfg.Server.ReadHeaderTimeout == 0 {
    cfg.Server.ReadHeaderTimeout = 5  // 5s 切断慢速 Header 攻击
}
if cfg.Server.WriteTimeout == 0 {
    cfg.Server.WriteTimeout = 10
}
if cfg.Server.IdleTimeout == 0 {
    cfg.Server.IdleTimeout = 30       // 30s 回收僵死连接
}
if cfg.Server.MaxHeaderBytes == 0 {
    cfg.Server.MaxHeaderBytes = 1 << 20  // 1MB，Go 默认值
}
```

### 2. `cmd/main.go` — 应用超时配置

```go
// 11. 启动服务
srv := &http.Server{
    Addr:              fmt.Sprintf(":%d", cfg.Server.Port),
    Handler:           r,
    ReadTimeout:       time.Duration(cfg.Server.ReadTimeout) * time.Second,
    ReadHeaderTimeout: time.Duration(cfg.Server.ReadHeaderTimeout) * time.Second,
    WriteTimeout:      time.Duration(cfg.Server.WriteTimeout) * time.Second,
    IdleTimeout:       time.Duration(cfg.Server.IdleTimeout) * time.Second,
    MaxHeaderBytes:    cfg.Server.MaxHeaderBytes,
}
```

### 3. `conf/config.yaml` — 添加超时配置节

```yaml
server:
  port: 8080
  mode: debug
  read_timeout: 10
  read_header_timeout: 5
  write_timeout: 10
  idle_timeout: 30
  max_header_bytes: 1048576
```

### 4. `conf/config.example.yaml` — 同步模板

```yaml
server:
  port: 8080
  mode: debug
  read_timeout: 10
  read_header_timeout: 5
  write_timeout: 10
  idle_timeout: 30
  max_header_bytes: 1048576
```

## 参数取值依据

| 参数 | 值 | 理由 |
|------|----|------|
| `ReadTimeout` | 10s | 最慢 API（login bcrypt）约 0.8ms，10s 余量极大 |
| `ReadHeaderTimeout` | 5s | Header 读取正常 <1ms，5s 只拦攻击不误杀 |
| `WriteTimeout` | 10s | 响应体均 <100KB，内网远超此值 |
| `IdleTimeout` | **30s** | 🔑 **修复 CLOSE_WAIT 的核心**。Vite proxy timeout 12s，30s 覆盖其重试窗口 |
| `MaxHeaderBytes` | 1MB | Go 默认值，显式声明便于审计 |

## 实现步骤

| # | 步骤 | 文件 | 设计模式 |
|---|------|------|---------|
| 1 | `ServerConfig` 增加 6 个字段 | [config.go:25-28](backend/pkg/config/config.go#L25-L28) | DTO 扩展 |
| 2 | `setDefaults()` 追加 6 个默认值 | [config.go:101-132](backend/pkg/config/config.go#L101-L132) | Factory Method |
| 3 | `http.Server{}` 读取配置并应用 | [main.go:110-113](backend/cmd/main.go#L110-L113) | 直接注入 |
| 4 | `config.yaml` 添加 server 配置节 | [config.yaml:1-3](backend/conf/config.yaml#L1-L3) | 🟡 环境变量 |
| 5 | `config.example.yaml` 同步 | [config.example.yaml:1-3](backend/conf/config.example.yaml#L1-L3) | 文档 |
| 6 | 编译 + 重启后端验证 | — | — |

## 测试策略

| 层级 | 测试内容 | 方式 |
|------|---------|------|
| 单元 | `setDefaults()` 6 个默认值生效 | `go test ./pkg/config/` |
| 集成 | 慢请求（sleep 15s）验证 ReadTimeout 10s 切断 | curl |
| 集成 | 慢 Header（逐字节发 >5s）验证 ReadHeaderTimeout 切断 | netcat |
| 回归 | 正常 API（login / listUsers）不受影响 | 浏览器 |
| 连接泄漏 | 连续 100 次请求后 `netstat` CLOSE_WAIT = 0 | 脚本 |
| 大 Header | 发送 2MB Header 验证 MaxHeaderBytes 拒绝 | curl |

## 回滚方案

```bash
git revert <commit>
# 或手动：
# 1. 删除 main.go 中新增的 5 行 Server 字段
# 2. 删除 config.go 中 6 个字段 + setDefaults 6 行
# 3. config.yaml / example.yaml 删除 server 下新增字段
```
