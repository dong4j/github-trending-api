// Package store 定义 trending 数据持久化接口。
//
// R-01 v1.2: 从零建立 SQLite 持久化层。
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
	GetRepos(since, lang string, limit int) ([]model.TrendingRepo, error)

	// GetUnenrichedRepos 获取待 enrich 的 repo（按 priority desc）。
	GetUnenrichedRepos(limit int) ([]model.TrendingRepo, error)

	// UpdateEnriched 写入 enricher 补全字段 + 标记 enriched_at。
	UpdateEnriched(fullName, since string, repo model.TrendingRepo) error

	// MarkUnavailable 标记 repo 不可用（404）。
	MarkUnavailable(fullName, since string) error

	// RecomputePriorities 重新计算 enrich 优先级。
	RecomputePriorities(since string) error

	// --- Languages ---

	// UpsertLanguages 批量 upsert 语言列表。
	UpsertLanguages(langs []model.Language) error

	// GetLanguages 获取语言列表。
	GetLanguages() ([]model.Language, error)

	// Close 关闭数据库连接。
	Close() error
}
