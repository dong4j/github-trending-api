# Starcat Trending API

GitHub Trending API，使用 Go 语言实现。

> **R-01 v1.2**（2026-06-09）：从无状态 HTML 爬虫升级为三层架构（spider → store → enricher → scheduler），加 Bearer Token 鉴权、Token Pool、SQLite 持久化，API 升级到 `/api/v1/*`。

## 特性

- **三层架构**：爬虫（goquery HTML 解析）→ SQLite 落库 → GitHub API 字段补全 → cron 定时调度
- **Bearer Token 鉴权**：所有 `/api/v1/*` 和 `/internal/*` 端点强制鉴权
- **Token Pool**：多 GitHub PAT 冗余 + Quota-aware 选择 + 故障切换
- **Rate Limit 退避**：主动读 `X-RateLimit-Remaining` 头，低配额时自动减速
- **优先级队列**：榜单前排优先 enrich（`enrich_priority DESC`）
- **Admin endpoint**：手动触发同步（`/internal/sync/*`）

## 快速开始

### 环境要求

- Go 1.25+

### 本地运行

```bash
cp .env.example .env
# 编辑 .env，填入 API_KEYS 和 GITHUB_TOKENS
cd starcat-trending-api
go run ./cmd/server/
```

默认端口 `5002`。

### .env 配置

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `PORT` | 服务端口 | `5002` |
| `STORE_FILE` | SQLite 数据库路径 | `./trending.db` |
| `API_KEYS` | Bearer Token 白名单（逗号分隔） | 必填 |
| `GITHUB_TOKENS` | GitHub PAT 池（逗号分隔） | 必填 |

## API 接口

所有数据接口需要 `Authorization: Bearer <api-key>` 头。

### `GET /api/v1/repos?lang=&since=&limit=`（需鉴权）

返回 trending 仓库列表（含 GitHub API 补全字段）。

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `lang` | string | — | 语言过滤（如 `Go`、`Python`） |
| `since` | string | `daily` | `daily` / `weekly` / `monthly` |
| `limit` | int | 100 | 返回数量上限（最大 100） |

响应示例见 `internal/model/repo_card.go` 中的 `StarcatRepoCardDTO`。

### `GET /api/v1/languages`（需鉴权）

返回可用语言列表（24h 内存缓存）。

### `GET /api/v1/users?lang=&since=&sponsorable=`（需鉴权）

返回 trending 开发者列表。

### Admin Endpoints（需鉴权）

| 端点 | 说明 |
|------|------|
| `POST /internal/sync/repos` | 手动触发全量重爬 + enrich |
| `POST /internal/sync/languages` | 手动刷��语言列表缓存 |
| `POST /internal/sync/users` | 手动触发重爬开发者榜单 |

### `GET /healthz`（公开）

健康检查，返回 `ok`。

## 鉴权

所有 `/api/v1/*` 和 `/internal/*` 端点需要 `Authorization: Bearer <api-key>` 头。

生成新 key：

```bash
bash ../scripts/gen-api-key.sh
```

## 项目结构

```
starcat-trending-api/
├── cmd/server/main.go              # 入口：装配三层 + scheduler + 鉴权
├── internal/
│   ├── spider/                     # HTML 爬虫（goquery）
│   ├── store/                      # SQLite 持久化
│   ├── enricher/                   # GitHub API 字段补全 + Rate Limit
│   ├── tokenpool/                  # GitHub Token Pool
│   ├── scheduler/                  # cron 定时调度
│   ├── handler/                    # HTTP handler（v1 + admin）
│   ├── middleware/                 # Bearer 鉴权中间件
│   └── model/                      # 数据模型（DB + DTO + Envelope）
├── .env.example                    # 配置模板
├── fly.toml                        # Fly.io 部署配置
├── Dockerfile
└── Makefile
```

## 部署（Fly.io）

```bash
fly secrets set \
  API_KEYS="sk-starcat-prodKey1,sk-starcat-prodKey2" \
  GITHUB_TOKENS="ghp_token1,ghp_token2,ghp_token3" \
  STORE_FILE="/data/trending.db" \
  -a starcat-trending-api

fly deploy -a starcat-trending-api
```

## 技术选型

- **net/http**：Go 标准库，无框架依赖
- **goquery**：HTML 解析（类 jQuery 选择器）
- **SQLite**：modernc.org/sqlite（纯 Go，无 CGO）
- **cron**：robfig/cron/v3（定时调度）
- **godotenv**：.env 文件加载
