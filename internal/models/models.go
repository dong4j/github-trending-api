// Package models 定义数据结构，与 Python 原版保持一致
package models

// BuildBy 表示仓库的构建者信息
type BuildBy struct {
	Avatar string `json:"avatar"`
	By     string `json:"by"`
}

// RepoItem 表示 trending 仓库项
type RepoItem struct {
	Repo    string    `json:"repo"`
	Desc    string    `json:"desc"`
	Lang    string    `json:"lang"`
	Stars   int       `json:"stars"`
	Forks   int       `json:"forks"`
	BuildBy []BuildBy `json:"build_by"`
	Change  int       `json:"change"`
}

// PopItem 表示开发者的热门仓库
type PopItem struct {
	Repo string `json:"repo"`
	Desc string `json:"desc"`
}

// UserItem 表示 trending 开发者
type UserItem struct {
	Avatar     string   `json:"avatar"`
	Name       string   `json:"name"`
	GitHubName string   `json:"github_name"`
	Popular    *PopItem `json:"popular"`
}

// LangItem 表示可用语言
type LangItem struct {
	Label string `json:"label"`
	Key   string `json:"key"`
}
