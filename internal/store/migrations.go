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
