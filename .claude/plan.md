# SteamPulse 项目计划

> 最后更新：2026-07-15
> 当前进度与开发命令见仓库根目录 `CLAUDE.md`；本文件描述产品设计与开发路线图。

## 1. 项目概述

SteamPulse 是一个轻量、快速的 Steam 玩家数据查询与实时监控工具。核心差异为 **Web 端免登录**：输入 Steam ID 即可查询，无需注册或登录。

- **数据源**：官方 Steam Web API（合法合规，非爬虫）。
- **前提**：玩家 Steam 个人资料须设为"公开"才能查到（Steam 平台机制，前端提示即可）。
- **定位**：更轻、更快、免登录，区别于需要登录或功能较重的同类工具。

**核心产品闭环**：输入 Steam ID → 输出「玩家档案卡」（基础档案 + 账号价值 + 时长分析）。
后续：关注玩家 → 实时监控其状态变化 → 自动生成游戏日记与年度报告。

## 2. 核心功能

1. **账号价值估算**：账号价值 ¥X / 已玩时长 Y 小时 / 单位时长成本 ¥Z。高传播性，适合分享。
2. **实时监控 + 游戏日记 / 年度报告**：基于采集的状态数据自动生成玩家活动时间线（Wrapped 风格）。Phase 2 的核心壁垒。

## 3. 架构

**分层铁律：core 厚，adapter 薄。core 传输无关——不感知外部是 gin、Telegram 还是 OneBot。**
后续接入新通道 = 新增一个 adapter，不修改 core、不修改 gin、不复制业务逻辑。

### 目录结构

```
cmd/server/main.go              # HTTP 入口（将来 bot adapter 并行 goroutine 启动）
internal/
  core/                       # 核心业务：Steam 查询 / 价值计算 / 监控 diff（传输无关）
    profile.go
    value.go
    monitor.go
  steam/                      # Steam API client
  adapter/                   # 传输层：每个对外通道一个，统一调 core
    http/                   # gin：网页 + REST API       (Phase 1)
    telegram/               # TG bot：webhook / polling (后续)
    onebot/                # QQ / NapCat OneBot 11  (后续)
  store/                    # MySQL / Redis 数据访问
```

### 三通道复用同一份 core

以"查玩家档案"为例：

```
Web : GET /api/profile/:id   → http adapter     → core.GetProfile(id) → JSON
TG  : /profile <id>          → telegram adapter → core.GetProfile(id) → 文本+图
QQ  : /profile <id>          → onebot adapter   → core.GetProfile(id) → 文本+图
```

### 通道接入方式

| 通道        | 推荐接法                                      | 与 gin 关系 |
| ----------- | ------------------------------------------- | ----------- |
| Telegram    | webhook：`POST /tg/webhook` 接收 TG 推送    | 复用 gin     |
| Telegram    | 或 polling：独立 goroutine 轮询                 | 独立         |
| QQ / NapCat | OneBot 11 WebSocket：独立连接接收事件          | 独立         |
| QQ / NapCat | 或 HTTP：NapCat POST 事件进来                 | 复用 gin     |

- Telegram：`go-telegram-bot-api/telegram-bot-api`
- NapCat：OneBot 11 协议（WebSocket + JSON），有 Go SDK，或自行实现 WS

### main.go 多通道并行启动

```
go runHTTPServer()     // gin（Phase 1 即有）
// 后续：
go runTelegramBot()     // webhook 复用 gin，或 polling 独立 goroutine
go runOneBotWS()        // NapCat WebSocket 独立连接
```

### Phase 1 纪律

gin handler 必须薄：**解析参数 → 调 core → 返回结果**，业务逻辑全部置于 `internal/core`。
在 handler 中写业务会导致接 bot 时逻辑被复制一遍。多花 10 分钟分层，省两天重构。

## 4. 技术栈

| 技术                    | 用途                                                |
| ----------------------- | --------------------------------------------------- |
| Go + gin               | HTTP/API 服务（Web 为一等公民）；bot 可复用或并行        |
| MySQL                  | 关注列表、游戏库快照、事件历史                     |
| Redis                  | 缓存 profile；存"上次快照"做 diff（Steam API 限速真刚需） |
| 令牌桶限流              | 应对 Steam API 限速 + Phase 2 多用户轮询放大          |
| goroutine pool          | Phase 2 批量轮询多用户                              |
| Docker / docker-compose | 部署上线                                            |

## 5. 开发路线图

**顺序铁律：MVP 先上线拿用户，再做监控。不可颠倒。**

### Phase 1 — MVP

输入 Steam ID → 玩家档案卡。严格控范围，不做监控。

- Steam Web API 封装（`GetPlayerSummaries` / `GetRecentlyPlayedGames` / `GetOwnedGames` / `GetPlayerAchievements`）
- 基础档案：头像、昵称、等级、注册天数、游戏总数、总时长
- 账号价值估算
- 时长分析：最常玩 5 个游戏、最爱类型、简单图表
- Redis 缓存 + 令牌桶限速
- Docker 部署上线

### Phase 1.5 — 上线

部署 + 获取真实用户 + 整理技术叙事。

### Phase 2 — 监控引擎

本质：状态机 + 增量 diff 引擎。

- 调度器：定时轮询关注用户（Steam 无推送，必须轮询）
- worker pool + 全局令牌桶（多用户并发轮询，限速）
- 三个 diff：
  - 在线状态机：offline → 玩 A → 玩 B → offline，触发事件
  - 会话时长：`本次 playtime − 上次 playtime`（playtime 为累计值）
  - 成就解锁：bitmap 存成就状态，diff O(n/64)
- 自适应轮询（活跃用户多查、僵尸号少查）
- 事件存储（时间序列）
- Steam 游戏日记 / 年度报告（Wrapped 风格）

### Phase 2.5 — 监控上线

持续迭代。

## 6. 关键技术决策与依据

- **令牌桶 + Redis 缓存**：Steam API 限速下，缓存 profile + 令牌桶控速，把响应延迟压到最低。
- **必须轮询**：Steam 无官方推送 webhook，监控只能轮询（Phase 2 设计前提）。
- **批量价格拉取 + 价格表缓存**：大玩家几千款游戏，批量拉价格并缓存价格表，避免逐个请求。
- **profile 缓存策略**：TTL + 主动失效。
- **成就 diff 用 bitmap**：成就状态用 bitmap 存储，diff 复杂度 O(n/64)；相比 hash set 空间省半且位运算快。
- **自适应轮询间隔**：活跃用户多查、僵尸号少查，平衡 API 压力与数据新鲜度。
- **时间序列归档**：事件表 + Redis 存近态 + 老数据归档，控制存储增长。
- **汇率**：每日拉取一次并缓存。
- **core/adapter 分层**：接 bot 时不碰 core。

## 7. 约束与避坑

- 使用官方 Steam Web API + key，合法、可讲清。
- 玩家 profile 须公开才能查到（正常机制，前端提示）。
- 不碰商标 / 不商业化；个人项目 + 开源 OK。
- 顺序铁律：MVP 先上线拿用户，再做监控。
- 不做模板项目（限流 / 秒杀 / IM / 博客风）——每项技术在本项目均有真实理由。
- 不 buzzword 堆叠（Redis + MQ + K8s + Prometheus 全家桶 = 教程感）。
- 不做"AI 黑盒"：代码须能讲清原理，不能是写了却讲不懂。
- 业务逻辑不写死在 gin handler。
- 不一上来就做监控 / 不无限加功能不上线——先 ship MVP。
