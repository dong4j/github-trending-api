package store

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

// createSchema 初始化 trending 数据表与索引。
//
// trending-api 是 GitHub 单源服务，仅维护一张 trending_repos + 一张 trending_languages。
// 主键 (full_name, since)：同一 repo 在不同 since 区间（daily / weekly / monthly）独立落库。
//
// 关键索引：
//   - idx_trending_since_captured：API 查询主路径（since + 最新捕获时间）
//   - idx_trending_gh_repo_id：enricher 按 gh_repo_id 反查
//   - idx_trending_unenriched：partial index，仅索引待 enrich 的可用行
//   - idx_trending_language_since：语言过滤查询
func createSchema(db *sql.DB) error {
	log.Println("[migrate] createSchema: trending_repos + trending_languages")
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS trending_repos (
			full_name       TEXT NOT NULL,
			owner           TEXT NOT NULL,
			name            TEXT NOT NULL,
			desc_text       TEXT,
			stars           INTEGER NOT NULL DEFAULT 0,
			forks           INTEGER NOT NULL DEFAULT 0,
			language        TEXT,
			change          INTEGER NOT NULL DEFAULT 0,
			build_by_json   TEXT,
			gh_repo_id      INTEGER,
			description     TEXT,
			homepage        TEXT,
			license_spdx    TEXT,
			topics_json     TEXT,
			watchers        INTEGER DEFAULT 0,
			subscribers     INTEGER DEFAULT 0,
			owner_avatar    TEXT,
			is_archived     INTEGER NOT NULL DEFAULT 0,
			is_fork         INTEGER NOT NULL DEFAULT 0,
			is_private      INTEGER NOT NULL DEFAULT 0,
			default_branch  TEXT,
			open_issues     INTEGER DEFAULT 0,
			pushed_at       TEXT,
			updated_at      TEXT,
			created_at      TEXT,
			since           TEXT NOT NULL,
			captured_at     TEXT NOT NULL,
			enriched_at     TEXT,
			is_available    INTEGER NOT NULL DEFAULT 1,
			enrich_priority INTEGER NOT NULL DEFAULT 0,
			PRIMARY KEY (full_name, since)
		);

		CREATE INDEX IF NOT EXISTS idx_trending_since_captured ON trending_repos(since, captured_at DESC);
		CREATE INDEX IF NOT EXISTS idx_trending_gh_repo_id ON trending_repos(gh_repo_id) WHERE gh_repo_id IS NOT NULL;
		CREATE INDEX IF NOT EXISTS idx_trending_unenriched ON trending_repos(enriched_at) WHERE enriched_at IS NULL AND is_available = 1;
		CREATE INDEX IF NOT EXISTS idx_trending_language_since ON trending_repos(language, since, captured_at DESC);

		CREATE TABLE IF NOT EXISTS trending_languages (
			key         TEXT PRIMARY KEY,
			label       TEXT NOT NULL,
			captured_at TEXT NOT NULL
		);
	`)
	return err
}
