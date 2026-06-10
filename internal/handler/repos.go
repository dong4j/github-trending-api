// Package handler 包含 HTTP handler 实现。
//
// trending-api 只走 GitHub 单源；zread 数据请走 weekly-api /api/v1/trending/zread。
package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dong4j/starcat-trending-api/internal/model"
	"github.com/dong4j/starcat-trending-api/internal/store"
)

// HandleReposV1 GET /api/v1/repos - 返回 GitHub trending repo 卡片列表。
//
// query 参数：
//   - since: daily | weekly | monthly（默认 daily）
//   - lang: 语言过滤（如 go / python / swift）
//   - limit: 1-100（默认 100）
//
// 不接受 source=* 参数。trending-api 固定走 GitHub 单源；zread 数据请改用
// weekly-api /api/v1/trending/zread。
func HandleReposV1(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := r.URL.Query().Get("lang")
		since := r.URL.Query().Get("since")
		if since == "" {
			since = "daily"
		}

		if since != "daily" && since != "weekly" && since != "monthly" {
			writeError(w, http.StatusBadRequest, "BAD_REQUEST",
				"since must be one of: daily, weekly, monthly",
				map[string]interface{}{
					"param":   "since",
					"got":     since,
					"allowed": []string{"daily", "weekly", "monthly"},
				})
			return
		}

		// 显式拒绝任何 source=* 参数，引导客户端改用 weekly-api。
		if src := r.URL.Query().Get("source"); src != "" {
			writeError(w, http.StatusBadRequest, "BAD_REQUEST",
				fmt.Sprintf("source=%s is not supported by trending-api. "+
					"trending-api is github-only; use weekly-api /api/v1/trending/zread for zread data.", src),
				map[string]interface{}{
					"param":   "source",
					"got":     src,
					"allowed": []string{"(none — trending-api is github-only)"},
					"see":     "https://github.com/dong4j/starcat-weekly-api (GET /api/v1/trending/zread)",
				})
			return
		}

		limit := 100
		if l := r.URL.Query().Get("limit"); l != "" {
			fmt.Sscanf(l, "%d", &limit)
			if limit > 100 {
				limit = 100
			}
		}

		// GitHub 单源查询
		repos, err := s.GetRepos(since, lang, limit)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR",
				"failed to query repos: "+err.Error(), nil)
			return
		}

		cards := make([]model.StarcatRepoCardDTO, len(repos))
		for i, r := range repos {
			cards[i] = store.TrendingRepoToCardDTO(r)
		}

		cacheStatus := "fresh"
		if len(cards) == 0 {
			cacheStatus = "cold"
		}

		writeJSONWithMeta(w, cards, &model.Meta{
			Since:       since,
			Language:    lang,
			Total:       len(cards),
			GeneratedAt: time.Now().Format(time.RFC3339),
			CacheStatus: cacheStatus,
		})
	}
}

