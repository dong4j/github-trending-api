package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dong4j/starcat-trending-api/internal/model"
	"github.com/dong4j/starcat-trending-api/internal/store"
)

// HandleReposV1 GET /api/v1/repos - 返回 trending repo 卡片列表。
func HandleReposV1(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := r.URL.Query().Get("lang")
		since := r.URL.Query().Get("since")
		source := r.URL.Query().Get("source") // github|zread|merged
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

		limit := 100
		if l := r.URL.Query().Get("limit"); l != "" {
			fmt.Sscanf(l, "%d", &limit)
			if limit > 100 {
				limit = 100
			}
		}

		var repos []model.TrendingRepo
		var err error

		if source == "merged" {
			// merged 逻辑: github 优先, zread 补缺
			ghRepos, err := s.GetRepos(since, lang, "github", limit)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR",
					"failed to query github repos: "+err.Error(), nil)
				return
			}

			// 只有 weekly 支持 zread
			var zRepos []model.TrendingRepo
			if since == "weekly" {
				zRepos, _ = s.GetRepos(since, lang, "zread", limit)
			}

			// 去重合并
			seen := make(map[string]bool)
			for _, r := range ghRepos {
				seen[r.FullName] = true
				repos = append(repos, r)
			}
			mergedFromZread := 0
			for _, r := range zRepos {
				if !seen[r.FullName] {
					repos = append(repos, r)
					mergedFromZread++
				}
			}
			
			mergedDedupRemoved := len(zRepos) - mergedFromZread

			cards := make([]model.StarcatRepoCardDTO, len(repos))
			for i, r := range repos {
				cards[i] = store.TrendingRepoToCardDTO(r)
			}

			writeJSONWithMeta(w, cards, &model.Meta{
				Since:              since,
				Language:           lang,
				Source:             source,
				Total:              len(cards),
				MergedFromGithub:   len(ghRepos),
				MergedFromZread:    mergedFromZread,
				MergedDedupRemoved: mergedDedupRemoved,
				GeneratedAt:        time.Now().Format(time.RFC3339),
				CacheStatus:        "fresh",
			})
			return
		}

		// 单源查询
		repos, err = s.GetRepos(since, lang, source, limit)
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
			Source:      source,
			Total:       len(cards),
			GeneratedAt: time.Now().Format(time.RFC3339),
			CacheStatus: cacheStatus,
		})
	}
}

