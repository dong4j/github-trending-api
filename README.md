# GitHub Trending API (Go 版本)

> 使用 Go 语言重写的 GitHub Trending API，与 [Python 原版](https://github.com/doforce/github-trending) 保持 **100% 兼容** 的 REST API 接口。

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/docker-supported-2496ED?style=flat&logo=docker)](Dockerfile)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

## ✨ 特性

- 🚀 **零依赖框架**：使用 Go 标准库 `net/http`
- 📦 **单二进制部署**：无运行时依赖
- 🐳 **多架构 Docker**：`linux/amd64` + `linux/arm64`
- ✅ **100% API 兼容**：与 Python 版本字段完全一致
- 🔧 **易于扩展**：清晰的目录结构
- 🤖 **CI/CD 就绪**：内置 GitHub Actions

## 📂 项目结构

```
github-trending-api/
├── cmd/server/main.go          # HTTP 服务器入口
├── internal/
│   ├── models/models.go        # 数据模型定义
│   └── spider/
│       ├── base.go            # 爬虫基类 (HTTP 请求)
│       ├── lang.go            # 语言列表爬虫
│       ├── repo.go            # 仓库 trending 爬虫
│       └── user.go            # 开发者 trending 爬虫
├── pkg/utils/number.go         # 工具函数 (数字解析)
├── .github/                    # GitHub 配置
│   ├── ISSUE_TEMPLATE/         # Issue 模板
│   ├── workflows/              # CI/CD
│   ├── dependabot.yml          # 依赖更新
│   └── PULL_REQUEST_TEMPLATE.md
├── go.mod / go.sum            # Go 模块
├── Dockerfile                 # 多阶段构建
├── Makefile                   # 统一命令入口
├── .gitignore / .gitattributes / .editorconfig
├── CONTRIBUTING.md
├── CODE_OF_CONDUCT.md
├── SECURITY.md
├── CHANGELOG.md
└── README.md
```

## 🚀 快速开始

### 方式 1：使用 Makefile（推荐）

```bash
# 查看所有可用命令
make help

# 开发模式运行
make run

# 编译
make build

# 完整检查
make check
```

### 方式 2：直接使用 Go

```bash
go mod tidy
go build -o bin/server ./cmd/server/
./bin/server
```

### 方式 3：Docker

```bash
# 构建镜像
make docker-build
# 或
docker build -t github-trending-api .

# 运行
make docker-run
# 或
docker run --rm -p 8080:8080 github-trending-api
```

服务器将在 `http://localhost:8080` 启动。

## 📚 API 接口

所有接口与 Python 原版保持 **完全一致**。

### `GET /`

欢迎信息

```bash
curl http://localhost:8080/
```

响应:
```json
{"message":"Hello GitHub trending"}
```

### `GET /lang`

获取所有可用语言（约 814 种）

```bash
curl http://localhost:8080/lang
```

响应:
```json
[
  {"label":"Unknown languages","key":"unknown"},
  {"label":"Python","key":"python"},
  {"label":"C#","key":"c%23"}
]
```

### `GET /repo`

获取 trending 仓库列表

| 参数 | 类型   | 描述                                      |
|------|--------|-------------------------------------------|
| lang | string | 可选，语言过滤 (如 `python`, `go`, `java`)  |
| since| string | 可选，默认 `daily`，可选: `daily`/`weekly`/`monthly` |

```bash
# 所有语言的每日 trending
curl "http://localhost:8080/repo"

# Python 的每周 trending
curl "http://localhost:8080/repo?lang=python&since=weekly"
```

响应: (最多 25 条)
```json
[
  {
    "repo": "/StarRocks/starrocks",
    "desc": "StarRocks 描述...",
    "lang": "Java",
    "stars": 6338,
    "forks": 1437,
    "build_by": [
      {"avatar": "https://avatars.githubusercontent.com/...", "by": "/amber-create"}
    ],
    "change": 619
  }
]
```

### `GET /user`

获取 trending 开发者列表

| 参数        | 类型   | 描述                                              |
|-------------|--------|--------------------------------------------------|
| lang        | string | 可选，语言过滤                                    |
| since       | string | 可选，默认 `daily`，可选: `daily`/`weekly`/`monthly` |
| sponsorable | string | 可选，`"1"` 表示仅显示有赞助选项的开发者             |

```bash
curl "http://localhost:8080/user?lang=python&since=weekly"
```

响应: (最多 25 条)
```json
[
  {
    "avatar": "https://avatars.githubusercontent.com/u/322311?s=96&v=4",
    "name": "Ben McCann",
    "github_name": "/benmccann",
    "popular": {
      "repo": "/benmccann/NameMatching",
      "desc": "MITRE's name matching competition"
    }
  }
]
```

## 🆚 对比原版 (Python)

| 特性         | Python 原版           | Go 版本                    |
|--------------|----------------------|----------------------------|
| Web 框架     | FastAPI              | 标准库 `net/http`           |
| HTML 解析    | parsel + Scrapy      | goquery                    |
| 依赖数量     | 5 个包               | 1 个库                      |
| 并发         | 异步                 | goroutine (原生)            |
| 部署         | Vercel / Docker      | 单二进制 / Docker            |
| 启动时间     | 数百毫秒              | < 10ms                      |
| 内存占用     | ~50MB+               | ~5MB                        |

## 🛠 技术选型说明

- **net/http**: 使用 Go 标准库，无需额外框架依赖
- **goquery**: 类似于 Python 的 `parsel`，选择器语法相似
- **无反射 ORM**: 直接使用 `encoding/json`，无性能损耗

## ☁️ 部署

支持多种免费云平台部署：

| 平台 | 永久免费 | 不休眠 | 难度 | 适用场景 |
|------|---------|--------|------|---------|
| 🏆 [Fly.io](docs/DEPLOY_FLY.md) | ✅ | ✅ | ⭐⭐ | **海外访问快、长期运行** |
| [Render](docs/DEPLOY_RENDER.md) | ⚠️ 有限制 | ❌ 15 分钟休眠 | ⭐ | 演示、GitHub 集成 |
| Oracle Cloud | ✅ | ✅ | ⭐⭐⭐ | 国内访问稳定 |
| Railway | ❌ $5 试用 | ✅ | ⭐ | 简单测试 |

### 🚀 快速部署到 Fly.io

```bash
brew install flyctl  # 安装 CLI
fly auth signup       # 注册（需信用卡验证，不扣费）
fly launch            # 一键部署
fly open              # 打开应用
```

详细步骤见 [docs/DEPLOY_FLY.md](docs/DEPLOY_FLY.md)。

### 🚀 快速部署到 Render

1. 在 Render Dashboard 点击 **New** → **Blueprint**
2. 连接 `dong4j/github-trending-api` 仓库
3. Render 自动读取 `render.yaml` 并部署

详细步骤见 [docs/DEPLOY_RENDER.md](docs/DEPLOY_RENDER.md)。

## 🤝 贡献

欢迎贡献！请阅读 [CONTRIBUTING.md](CONTRIBUTING.md) 了解贡献流程。

## 📜 许可证

本项目基于 MIT 协议开源 - 详见 [LICENSE](LICENSE) 文件。

## 🙏 致谢

- 原 Python 版本作者：[Edgar Xie](https://github.com/doforce)
- HTML 解析：[goquery](https://github.com/PuerkitoBio/goquery)
