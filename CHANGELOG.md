# Changelog

本项目的所有重要变更都会记录在此文件中。

格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/)，
版本号遵循 [Semantic Versioning](https://semver.org/lang/zh-CN/)。

## [Unreleased]

## [2.1.0] - 2026-06-10

### Added
- **R-02 (Zread Integration)**：引入 Zread Trending 数据源。
- **Source 维度合并**：`GET /api/v1/repos?source=merged` 提供 GitHub 与 Zread 双榜合并去重能力。
- **Zread 扩展字段**：在 trending 扩展段中增加 `description_zh` 和 `zread_wiki_id` 等字段。

## [2.0.0] - 2026-06-09

### Added
- **三层架构**：spider（爬虫）→ store（SQLite）→ enricher（GitHub API 补全）
- **SQLite 持久化**：`trending_repos` + `trending_languages` 表，WAL 模式
- **GitHub API Enricher**：补全 `gh_repo_id`、topics、license、owner_avatar 等 18 个字段
- **Token Pool**：多 GitHub PAT 冗余 + Quota-aware `PickBest()` + 故障切换
- **Rate Limit 退避**：主动读 `X-RateLimit-Remaining` 头，低配额自动减速
- **Cron Scheduler**：daily/weekly/monthly 定时爬取 + 长尾 enrich + 过期清理
- **Enrich Queue**：优先级队列 + worker pool（`enrich_priority DESC`）
- **Bearer Token 鉴权**：所有 `/api/v1/*` 和 `/internal/*` 端点强制鉴权
- **Admin Endpoints**：`/internal/sync/{repos,languages,users}` 手动触发同步
- **Envelope 统一响应**：`schema_version` + `data` / `error` + `meta`
- **`.env` 配置**：godotenv 加载，敏感值走 fly secrets

### Changed
- **Breaking**: `GET /repo` → `GET /api/v1/repos`（旧路径直接删除，含 envelope）
- **Breaking**: `GET /lang` → `GET /api/v1/languages`
- **Breaking**: `GET /user` → `GET /api/v1/users`
- **Breaking**: 所有端点（除 `/healthz`）需 Bearer Token
- **Breaking**: 响应格式升级为 envelope（嵌套 `data` / `error`）
- `internal/models/` → `internal/model/` + spider 类型迁入 `internal/spider/types.go`

### Removed
- `GET /` 欢迎端点
- `GET /lang`、`GET /repo`、`GET /user` 旧端点
- `internal/models/models.go`（类型迁至 spider 包）

## [1.0.0] - 2026-06-08

### Added
- Go 语言重写版本（基于 [Python 原版](https://github.com/doforce/github-trending)）
- 标准库 `net/http` 实现 REST API
- 使用 `goquery` 进行 HTML 解析
- Docker 多阶段构建（`linux/amd64` + `linux/arm64`）
- GitHub Actions CI（Go 构建 + Docker 镜像）
- Dependabot 自动依赖更新
- Issue / PR 模板
- Security 政策
- 内部版本号包 (`internal/version`, 暴露 `version.Version` 常量)

### Changed
- **重写**：从 Python/FastAPI 重写为 Go
- **依赖简化**：从 5 个 Python 包精简为 1 个 Go 库（`goquery`）
- **性能提升**：原生 goroutine 并发、二进制启动毫秒级

### API 兼容性
- `GET /` - 完全兼容
- `GET /lang` - 完全兼容
- `GET /repo?lang=&since=` - 完全兼容
- `GET /user?lang=&since=&sponsorable=` - 完全兼容

[Unreleased]: https://github.com/dong4j/starcat-trending-api/compare/v2.0.0...HEAD
[2.0.0]: https://github.com/dong4j/starcat-trending-api/releases/tag/v2.0.0
