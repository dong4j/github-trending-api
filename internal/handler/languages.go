package handler

import (
	"net/http"
	"time"

	"github.com/dong4j/starcat-trending-api/internal/model"
	"github.com/dong4j/starcat-trending-api/internal/scheduler"
)

// HandleLanguagesV1 GET /api/v1/languages - 返回可选语言列表。
func HandleLanguagesV1(sch *scheduler.Scheduler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		languages := sch.GetLanguages()

		writeJSONWithMeta(w, languages, &model.Meta{
			GeneratedAt: time.Now().Format(time.RFC3339),
			CacheStatus: "fresh",
		})
	}
}
