// Package spider 定义 GitHub Trending 页面爬虫的数据结构。
//
// R-01 v1.2: 从 internal/models/models.go 迁移至此，
// 这些类型是爬虫阶段的输出，不应独立为一个包。
package spider

// BuildBy 表示仓库的构建者信息（contributor avatar + profile link）。
type BuildBy struct {
	Avatar string `json:"avatar"`
	By     string `json:"by"`
}

// RepoItem 表示 trending 仓库列表中的单个仓库。
type RepoItem struct {
	Repo    string    `json:"repo"`
	Desc    string    `json:"desc"`
	Lang    string    `json:"lang"`
	Stars   int       `json:"stars"`
	Forks   int       `json:"forks"`
	BuildBy []BuildBy `json:"build_by"`
	Change  int       `json:"change"`
}

// PopItem 表示 trending 开发者的热门仓库。
type PopItem struct {
	Repo string `json:"repo"`
	Desc string `json:"desc"`
}

// UserItem 表示 trending 开发者列表中的单个开发者。
type UserItem struct {
	Avatar     string   `json:"avatar"`
	Name       string   `json:"name"`
	GitHubName string   `json:"github_name"`
	Popular    *PopItem `json:"popular"`
}

// LangItem 表示 GitHub Trending 页面的可选语言。
type LangItem struct {
	Label string `json:"label"`
	Key   string `json:"key"`
}
