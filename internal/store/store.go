// Package store 定义 trending 数据持久化接口。
//
// R-01 v1.2: 从零建立 SQLite 持久化层。
// R-02 v2.1（P2 清理,2026-06-10）: 删 v0.3 zread 集成残留的 source 参数。trending-api 重新
// 收敛到 GitHub 单源数据，所有 Store 方法不再接受 source 入参。
package store

import (
	"github.com/dong4j/starcat-trending-api/internal/model"
)

// Store 定义 trending 数据访问接口。
type Store interface {
	// --- Trending Repos ---

	// UpsertRepo 按 full_name + since 幂等 upsert。
	UpsertRepo(repo model.TrendingRepo) error

	// GetRepos 按 since / lang / limit 查询 repo 列表。
	//
	// R-02 v2.1：删除 source 参数，trending-api 固定走 GitHub 单源。
	GetRepos(since, lang string, limit int) ([]model.TrendingRepo, error)

	// GetUnenrichedRepos 获取待 enrich 的 repo（按 priority desc）。
	GetUnenrichedRepos(limit int) ([]model.TrendingRepo, error)

	// UpdateEnriched 写入 enricher 补全字段 + 标记 enriched_at。
	//
	// R-02 v2.1：删除 source 参数。
	UpdateEnriched(fullName, since string, repo model.TrendingRepo) error

	// MarkUnavailable 标记 repo 不可用（404）。
	//
	// R-02 v2.1：删除 source 参数。
	MarkUnavailable(fullName, since string) error

	// RecomputePriorities 重新计算 enrich 优先级。
	//
	// R-02 v2.1：删除 source 参数，固定走全表 since 维度。
	RecomputePriorities(since string) error

	// ResetAllEnriched 把所有 repo 的 enriched_at 置为 NULL，让下一次
	// `enricher.EnrichAll()` 把全表当成"未 enrich"状态重新跑一遍 GitHub API。
	//
	// 设计动机（dong4j 2026-06-11）：
	// enricher 字段映射 / GitHub API 返回字段集会随版本演进（如 R-05 加 10 个
	// 详情字段），但 EnrichAll 流程是「只补 enriched_at IS NULL 的行」—— 已经
	// 用老版 enricher 处理过的历史数据**永远不会被新版字段覆盖**，新加字段会
	// 长期为空。这个方法配合 HandleEnrichForce admin endpoint 提供「一键全表重
	// enrich」能力，省得每次升级 enricher 字段都要手动删 trending.db。
	//
	// 关键约束：
	//   - 只清 enriched_at，不动 spider 抓的 stars / forks / change / language 等
	//     基础字段，避免在重 enrich 间隙客户端看到空卡片
	//   - 不重置 enrich_priority，让 enricher 仍按现有优先级倒序处理
	//   - 调用方需注意：会让所有行重新进入 enricher 队列，token 消耗 = 全表行数 × 1 req
	ResetAllEnriched() error

	// --- Languages ---

	// UpsertLanguages 批量 upsert 语言列表。
	UpsertLanguages(langs []model.Language) error

	// GetLanguages 获取语言列表。
	GetLanguages() ([]model.Language, error)

	// Close 关闭数据库连接。
	Close() error
}
