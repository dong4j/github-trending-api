// Package model 定义 Envelope 统一响应结构。
//
// R-01 v1.2: 所有 /api/v1/* 数据接口的 200 + 错误响应统一走此 envelope，
// 与 trending / sharing byte-level 一致（详见 supports/docs/R-01-总体设计.md §3.2）。
//
// ⚠️ 跨项目共享代码同步约定（supports/docs/R-01-总体设计.md §4.1）：
//   - 本文件必须在 trending / weekly / sharing 三个 API 中 byte-level 一致（仅 package 名 / module path 可不同）。
//   - 任何字段 / 注释修改都必须同时同步 3 份，否则违反 R-01 设计约束。
package model

// Envelope 是 /api/v1/* 200 响应的顶层包装。
// T 可以是 StarcatRepoCardDTO、ShareResponseDTO、[]Project 等具体业务类型。
type Envelope[T any] struct {
	SchemaVersion int   `json:"schema_version"`
	Data          T     `json:"data"`
	Meta          *Meta `json:"meta,omitempty"`
}

// Meta 可选的分页/性能/限流元数据。
//
// 字段集合是 3 个 API 的并集：
//   - 分页（trending / weekly / sharing 都可能用）: Page / PageSize / Total / NextPage
//   - trending /api/v1/repos 专用: Since / Language
//   - 语言列表 / 通用时间戳: GeneratedAt / CacheStatus / FetchedAt
//
// 所有字段都 omitempty —— 不用的 API 自动不输出。
type Meta struct {
	Page        int    `json:"page,omitempty"`
	PageSize    int    `json:"page_size,omitempty"`
	Total       int    `json:"total,omitempty"`
	NextPage    *int   `json:"next_page,omitempty"`
	Since       string `json:"since,omitempty"`
	Language    string `json:"language,omitempty"`
	GeneratedAt string `json:"generated_at,omitempty"` // RFC3339 字符串（不要用 time.Time，跨 API 序列化语义要一致）
	CacheStatus string `json:"cache_status,omitempty"` // fresh / stale / cold
	FetchedAt   string `json:"fetched_at,omitempty"`
}

// ErrorResponse 统一错误响应体。
type ErrorResponse struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// ErrorEnvelope 所有非 2xx 响应的顶层包装。
type ErrorEnvelope struct {
	SchemaVersion int           `json:"schema_version"`
	Error         ErrorResponse `json:"error"`
}
