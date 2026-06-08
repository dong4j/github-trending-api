# Fly.io 部署指南

> 适用场景：海外访问快、永久免费、不休眠

## 📋 前置条件

1. 拥有 GitHub 账号
2. 安装 [flyctl CLI](https://fly.io/docs/hands-on/install-flyctl/)

```bash
# macOS
brew install flyctl

# Linux
curl -L https://fly.io/install.sh | sh

# Windows
iwr https://fly.io/install.ps1 -useb | iex
```

## 🚀 部署步骤

### 1. 注册 Fly.io 账号

```bash
fly auth signup
# 按提示输入邮箱、密码、信用卡 (信用卡仅用于验证身份, 不会扣费)
```

### 2. 登录

```bash
fly auth login
```

### 3. 首次部署

在项目根目录执行：

```bash
fly launch
```

**交互选项**：
- `? Would you like to copy its configuration to the new app?` → **No**（我们已有 `fly.toml`）
- `? Do you want to tweak these settings before proceeding?` → **No**（使用默认）

或**完全跳过交互**（推荐）：

```bash
fly launch --no-deploy
fly deploy
```

### 4. 验证部署

```bash
# 查看应用状态
fly status

# 查看日志
fly logs

# 打开应用
fly open
```

部署成功后，你会得到：
- **默认域名**：`https://dong4j-starcat-trending-api.fly.dev`
- **自定义域名**：可在 Dashboard 配置

## 🧪 测试 API

```bash
# 健康检查
curl https://dong4j-starcat-trending-api.fly.dev/

# 测试 /lang
curl https://dong4j-starcat-trending-api.fly.dev/lang | head -c 200

# 测试 /repo
curl https://dong4j-starcat-trending-api.fly.dev/repo?lang=go | head -c 200

# 测试 /user
curl https://dong4j-starcat-trending-api.fly.dev/user | head -c 200
```

## 🔧 常用命令

```bash
# 部署
fly deploy

# 查看状态
fly status

# 实时日志
fly logs

# SSH 进容器
fly ssh console

# 扩展 (免费层仅支持 1 个实例)
fly scale count 1

# 查看环境变量
fly config env

# 设置环境变量
fly secrets set MY_VAR=xxx

# 销毁应用
fly apps destroy dong4j-starcat-trending-api
```

## 💰 费用

| 资源 | 免费额度 | 超出后 |
|------|---------|--------|
| VM | 3 个 shared-cpu-1x 256MB | $1.94/月/VM |
| 带宽 | 160GB/月 | $0.02/GB |
| 存储 | 1GB | $0.15/GB/月 |

**对于本项目**：完全在免费额度内。

## ⚠️ 注意事项

1. **256MB 内存**：本项目二进制 ~8.5MB，运行时占用 ~10MB，足够
2. **GitHub 访问**：默认 region 是 `nrt`（东京），GitHub 访问延迟低
3. **冷启动**：无冷启动，第一个请求立即响应
4. **持久化**：免费层无持久存储，每次重启数据丢失（本项目无状态，无影响）

## 🌍 修改区域

如需修改部署区域，编辑 `fly.toml`：

```toml
primary_region = "nrt"  # 东京
# 常用 region:
# nrt   - 东京 (日本)
# sin   - 新加坡
# hkg   - 香港
# sjc   - 圣何塞 (美国西海岸)
# fra   - 法兰克福 (欧洲)
# cdg   - 巴黎
```

## 🐛 故障排查

```bash
# 查看详细部署日志
fly deploy --verbose

# 查看应用健康状态
fly checks list

# 重启应用
fly apps restart dong4j-starcat-trending-api
```

## 📚 参考

- [Fly.io 官方文档](https://fly.io/docs/)
- [fly.toml 参考](https://fly.io/docs/reference/configuration/)
- [Fly.io Go 示例](https://fly.io/docs/languages-and-frameworks/golang/)
