package store

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

// migrate 执行所有未应用的 schema 迁移。
func migrate(db *sql.DB) error {
	var version int
	if err := db.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		return err
	}
	log.Printf("[migrate] current schema version: %d", version)

	if version < 1 {
		if err := migrateV1(db); err != nil {
			return err
		}
		version = 1
	}
	if version < 2 {
		if err := migrateV2(db); err != nil {
			return err
		}
		version = 2
	}
	return nil
}

func setUserVersion(tx *sql.Tx, version int) error {
	_, err := tx.Exec(fmt.Sprintf("PRAGMA user_version = %d", version))
	return err
}

// migrateV1 创建 trending_repos + trending_languages 表。
func migrateV1(db *sql.DB) error {
	log.Println("[migrate] running migrateV1: create trending tables")

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
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
	if err != nil {
		return err
	}

	if err := setUserVersion(tx, 1); err != nil {
		return err
	}

	return tx.Commit()
}

// migrateV2 添加 zread / wiki 相关字段，并重构主键（SQLite 不支持直接 ALTER PRIMARY KEY，这里采用重建表模式）。
func migrateV2(db *sql.DB) error {
	log.Println("[migrate] running migrateV2: add source and zread dimensions")

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. 创建 V2 表
	_, err = tx.Exec(`
		CREATE TABLE IF NOT EXISTS trending_repos_v2 (
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

			-- v2 new fields
			source              TEXT NOT NULL DEFAULT 'github',
			description_zh      TEXT,
			zread_week_start    TEXT,
			zread_week_end      TEXT,
			zread_week_label    TEXT,
			zread_rank_in_week  INTEGER,
			zread_wiki_id       TEXT,
			zread_week_start_raw TEXT,
			zread_week_end_raw   TEXT,
			zread_year_inferred  INTEGER,
			PRIMARY KEY (full_name, since, source)
		);
	`)
	if err != nil {
		return err
	}

	// 2. 数据迁移
	_, err = tx.Exec(`
		INSERT INTO trending_repos_v2 (
			full_name, owner, name, desc_text, stars, forks, language, change, build_by_json,
			gh_repo_id, description, homepage, license_spdx, topics_json, watchers, subscribers,
			owner_avatar, is_archived, is_fork, is_private, default_branch, open_issues,
			pushed_at, updated_at, created_at, since, captured_at, enriched_at, is_available, enrich_priority
		) SELECT
			full_name, owner, name, desc_text, stars, forks, language, change, build_by_json,
			gh_repo_id, description, homepage, license_spdx, topics_json, watchers, subscribers,
			owner_avatar, is_archived, is_fork, is_private, default_branch, open_issues,
			pushed_at, updated_at, created_at, since, captured_at, enriched_at, is_available, enrich_priority
		FROM trending_repos;
	`)
	if err != nil {
		return err
	}

	// 3. 删旧表，重命名新表
	_, err = tx.Exec(`DROP TABLE trending_repos;`)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`ALTER TABLE trending_repos_v2 RENAME TO trending_repos;`)
	if err != nil {
		return err
	}

	// 4. 重建索引 (v2)
	_, err = tx.Exec(`
		CREATE INDEX idx_trending_lookup ON trending_repos(source, since, captured_at DESC);
		CREATE INDEX idx_trending_gh_repo_id ON trending_repos(gh_repo_id) WHERE gh_repo_id IS NOT NULL;
		CREATE INDEX idx_trending_unenriched ON trending_repos(enriched_at) WHERE enriched_at IS NULL AND is_available = 1;
		CREATE INDEX idx_trending_language_since ON trending_repos(source, language, since, captured_at DESC);
	`)
	if err != nil {
		return err
	}

	if err := setUserVersion(tx, 2); err != nil {
		return err
	}

	return tx.Commit()
}
