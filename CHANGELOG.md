# Changelog

本项目的所有重要变更都会记录在此文件中。

格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/)，
版本号遵循 [Semantic Versioning](https://semver.org/lang/zh-CN/)。

## [Unreleased]

### Added
- Go 语言重写版本（基于 [Python 原版](https://github.com/doforce/github-trending)）
- 标准库 `net/http` 实现 REST API
- 使用 `goquery` 进行 HTML 解析
- Docker 多阶段构建（`linux/amd64` + `linux/arm64`）
- GitHub Actions CI（Go 构建 + Docker 镜像）
- Dependabot 自动依赖更新
- Issue / PR 模板
- Security 政策

## [2.0.0] - 2024-XX-XX

### Changed
- **重写**：从 Python/FastAPI 重写为 Go
- **依赖简化**：从 5 个 Python 包精简为 1 个 Go 库（`goquery`）
- **性能提升**：原生 goroutine 并发、二进制启动毫秒级

### API 兼容性
- `GET /` - 完全兼容
- `GET /lang` - 完全兼容
- `GET /repo?lang=&since=` - 完全兼容
- `GET /user?lang=&since=&sponsorable=` - 完全兼容

[Unreleased]: https://github.com/dong4j/github-trending-api/compare/v2.0.0...HEAD
[2.0.0]: https://github.com/dong4j/github-trending-api/releases/tag/v2.0.0
