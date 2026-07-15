# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目状态

Phase 1 进行中：**gin 骨架 + Steam 档案查询已搭好并跑通**——`GET /api/profile/:id` 能查到真实档案（请求流：handler → `core.GetProfile` → `steam.GetPlayerSummaries`）。完整开发计划见 `.claude/plan.md`——动工前必读。
module path 已定为 `github.com/0x3ea/SteamPulse`（`go 1.25.0` + gin v1.12.0 + godotenv）；项目名 SteamPulse。
**尚未做**：账号价值估算 / 时长分析 / Redis 缓存 / 令牌桶限速 / Docker 部署。

## 这是什么项目

SteamPulse：比小黑盒更轻、更快、**Web 端免登录**的 Steam 玩家数据查询 + 实时监控工具。
数据源是**官方 Steam Web API**（合法，非爬虫；玩家 profile 须设"公开"才能查到）。

## 架构铁律（最重要，从 Phase 1 第一行代码就要守）

### core / adapter 分层
- **core 厚，adapter 薄。core 不知道外面是 gin 还是 bot。**
- 三个对外通道复用同一份 core：
  - Web: gin → `adapter/http` → `core.GetProfile(id)` → JSON
  - Telegram: `adapter/telegram` → `core.GetProfile(id)` → 文本+图
  - QQ/NapCat (OneBot 11): `adapter/onebot` → `core.GetProfile(id)` → 文本+图
- **gin handler 必须写薄**：只做「解析参数 → 调 core → 返回结果」，业务逻辑**全部**放 `internal/core`。
  在 handler 里写业务 → 将来接 bot 时逻辑要复制粘贴 = 烂摊子。这是 Phase 1 唯一要守的纪律，多花 10 分钟分层省两天重构。
- **core 对外部依赖走接口**：core 包内**自己定义** `SteamClient` 接口（消费者侧、极小、隐式实现），依赖抽象而非具体 `*steam.Client` → 单测能塞 fake、不绑死实现。adapter 通过 `*core.Service` 这个**唯一入口**访问 core 全部能力（`main.go` 里 `core.NewService(steamClient)` 装配）。

### 目标目录结构
```
cmd/server/main.go   # 启动 gin +(将来)bot adapter，各自 goroutine 并行
internal/
  core/      # Steam 查询 / 价值计算 / 监控 diff，传输无关（核心）
  steam/     # Steam API client
  adapter/
    http/    # gin 网页 + REST API   (Phase 1)
    telegram/  # TG bot               (后续)
    onebot/    # QQ/NapCat OneBot 11  (后续)
  store/     # MySQL / Redis 数据访问
```

## 开发顺序铁律

**MVP 先上线拿用户，再做监控。顺序不可反。**
- **Phase 1 (MVP)**：输入 Steam ID → 出一张「玩家档案卡」（档案 + 账号价值 + 时长分析 + Redis 缓存 + 令牌桶限速）→ Docker 部署 → **发社区拿真实用户**。
- **Phase 2 (监控)**：调度器轮询 + worker pool + 全局令牌桶 + 三个 diff（在线状态机 / 会话时长 / 成就解锁）+ 游戏日记 / 年度报告。
- 不要在 Phase 1 就动手做 Phase 2 的监控/轮询。

## 技术栈（每项都有真实理由，不是 buzzword 堆叠）

Go + gin + MySQL + Redis + 令牌桶限流 + goroutine pool + Docker。
- **Redis**：缓存 profile + 存「上次快照」做 diff（Steam API 限速 → 真刚需）。
- **令牌桶**：Steam Web API 限速 + Phase 2 多用户轮询放大。
- **Steam 无官方推送 webhook** → 监控只能轮询（这是 Phase 2 设计的前提）。

## 两个杀手锏功能
1. **账号价值估算**（账号值 ¥X / 玩了 Y 小时 / 每小时 ¥Z）——最爱被分享。
2. **实时监控 + 游戏日记 / 年度报告**（Wrapped 风格）——Phase 2 护城河。

## 讲深点（做的时候要能讲清，不能是 AI 黑盒）
令牌桶 + Redis 缓存如何把响应延迟降下来；Steam 无推送为何必须轮询；大玩家几千款游戏的批量价格拉取 + 价格表缓存；**成就 diff 用 bitmap 存 O(n/64)** 为什么不用 hash set；自适应轮询间隔；时间序列增长 → 事件表 + Redis 近态 + 老数据归档；core/adapter 分层为何接 bot 时不用碰 core。

## 命令

标准 Go module 命令：
- 编译并跑：`go build ./...`，或 `go build -o bin/server ./cmd/server && ./bin/server`
- 本地验证：`./bin/server`，另开终端 `curl localhost:8080/healthz` → `{"status":"ok"}`；`curl localhost:8080/api/profile/<17位SteamID>` → 档案 JSON
- 测试：`go test ./...`，单个测试 `go test ./internal/core -run TestName -v`

## 运行环境与约定

- 配置走环境变量：`STEAM_API_KEY`（Steam Web API key，放 `.env`，**已被 `.gitignore` 忽略，勿提交**）、`ADDR`（监听地址，默认 `:8080`）。
- **`os.Getenv` 本身不读 `.env`**，但 `main.go` 已用 `godotenv.Load()` 自动加载——直接 `go run ./cmd/server` 就能读到 `.env` 里的 key。生产环境改由容器注入 env（不依赖 godotenv）。
- `/healthz` 是**零依赖**的存活探针：不查 DB、不调 Steam API；200 = 进程在跑、HTTP 层通。Docker / 负载均衡 / 本地 curl 都靠它。新功能路由在 `internal/adapter/http/router.go` 的 `NewRouter()` 里加。

## 避坑（项目相关，来自 plan.md §10）
- ❌ 做成模板项目（限流/秒杀/IM/博客风）——每项技术在本项目都有真实理由。
- ❌ buzzword 堆叠（Redis+MQ+K8s+Prometheus 全家桶 = 教程感）。
- ❌ 业务逻辑写死在 gin handler。
- ❌ 一上来就做监控 / 无限加功能不上线——先 ship MVP 拿用户。
- 别碰商标 / 商业化；个人项目 + 开源 OK。
