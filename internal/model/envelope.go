// Package model 定义 Envelope 统一响应结构。
//
// 所有 /api/v1/* 数据接口的 200 + 错误响应统一走此 envelope。
// 各 API 自治 envelope 的 Meta 字段（不做跨项目共享约束）。
package model

// Envelope 是 /api/v1/* 200 响应的顶层包装。
// T 可以是 StarcatRepoCardDTO、[]Project 等具体业务类型。
type Envelope[T any] struct {
	SchemaVersion int   `json:"schema_version"`
	Data          T     `json:"data"`
	Meta          *Meta `json:"meta,omitempty"`
}

// Meta 可选的分页/性能/限流元数据。
//
// trending-api 当前仅用到 Since / Language / Total / GeneratedAt / CacheStatus。
// 所有字段都 omitempty —— 不填的字段不输出。
type Meta struct {
	Page        int    `json:"page,omitempty"`
	PageSize    int    `json:"page_size,omitempty"`
	Total       int    `json:"total,omitempty"`
	NextPage    *int   `json:"next_page,omitempty"`
	Since       string `json:"since,omitempty"`
	Language    string `json:"language,omitempty"`
	GeneratedAt string `json:"generated_at,omitempty"`
	CacheStatus string `json:"cache_status,omitempty"`
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
