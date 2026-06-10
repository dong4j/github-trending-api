// Package model 定义 trending 数据库层结构。
//
// TrendingRepo 直接映射 SQLite trending_repos 表，与 DB 列一一对齐。
// R-01 v1.2: 替代旧的 internal/models/models.go 中的 RepoItem。
package model

import "time"

// TrendingRepo 是 trending_repos 表中一条记录的 Go 表示。
type TrendingRepo struct {
	FullName    string
	Owner       string
	Name        string
	DescText    *string // 爬虫原始描述
	Stars       int
	Forks       int
	Language    *string
	Change      int
	BuildByJSON *string // contributors JSON array

	// enricher 补全字段
	GhRepoID      *int64
	Description   *string // GitHub 官方 description，覆盖 desc_text
	Homepage      *string
	LicenseSpdx   *string
	TopicsJSON    *string
	Watchers      int
	Subscribers   int
	OwnerAvatar   *string
	IsArchived    bool
	IsFork        bool
	IsPrivate     bool
	DefaultBranch *string
	OpenIssues    int
	PushedAt      *string // RFC3339
	UpdatedAt     *string
	CreatedAt     *string

	// 元数据
	Since          string
	CapturedAt     time.Time
	EnrichedAt     *time.Time
	IsAvailable    bool
	EnrichPriority int

	// v0.4 source 维度
	Source string // github|zread

	// zread 独有
	DescriptionZh     *string
	ZreadWeekStart    *string
	ZreadWeekEnd      *string
	ZreadWeekLabel    *string
	ZreadRankInWeek   int
	ZreadWikiID       *string
	ZreadWeekStartRaw *string
	ZreadWeekEndRaw   *string
	ZreadYearInferred int
}
// Language 语言列表项。
type Language struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

// Developer 开发者列表项（保持与旧 UserItem 兼容）。
type Developer struct {
	Avatar     string `json:"avatar"`
	Name       string `json:"name"`
	GitHubName string `json:"github_name"`
	Popular    *struct {
		Repo string `json:"repo"`
		Desc string `json:"desc"`
	} `json:"popular,omitempty"`
}
