package handler

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dong4j/starcat-trending-api/internal/scheduler"
)

// HandleAdminSyncRepos POST /internal/sync/repos - 手动触发全量同步。
func HandleAdminSyncRepos(sch *scheduler.Scheduler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		taskID := fmt.Sprintf("task-%s-repos-%d", time.Now().Format("2006-01-02T15:04:05Z"), time.Now().UnixNano()%1000)

		go sch.SyncAll()

		writeJSON(w, map[string]string{
			"task_id":    taskID,
			"started_at": time.Now().Format(time.RFC3339),
			"status":     "running",
		})
	}
}

// HandleAdminSyncLanguages POST /internal/sync/languages - 手动触发热刷新语言列表。
func HandleAdminSyncLanguages(sch *scheduler.Scheduler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		taskID := fmt.Sprintf("task-%s-languages-%d", time.Now().Format("2006-01-02T15:04:05Z"), time.Now().UnixNano()%1000)

		go sch.SyncLanguages()

		writeJSON(w, map[string]string{
			"task_id":    taskID,
			"started_at": time.Now().Format(time.RFC3339),
			"status":     "running",
		})
	}
}

// HandleAdminSyncUsers POST /internal/sync/users - 「重爬开发者榜单」管理触发器。
//
// ⚠️ 已知约束（R-01 v1.2 设计意图）：
//
// trending 开发者数据（GET /api/v1/users）当前是**按需实时爬取**（无落库 / 无缓存），
// 因此本 admin endpoint 没有「触发后台爬虫 + 写库」的工作可做，只回 task_id 表示请求被接受。
//
// 之所以保留这个 endpoint，是为了：
//  1. 与 /internal/sync/{repos,languages} 的接口形态一致（前端 / 运维脚本可统一处理）
//  2. 未来如果把开发者数据改为「定时爬 + 落库」（前端 N4 演进或运维需求），不需要改 endpoint 路径
//  3. 监控 / 日志侧能区分「手动 sync users」与「业务请求 users」两条流量
//
// 后续如果接入真实爬取，把下面注释里的 spider.UserSync(...) 启动起来即可。
// 详见 supports/docs/R-01-trending-api-改造方案.md §6.2.3 + TODO.md「中优」条目。
func HandleAdminSyncUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		taskID := fmt.Sprintf("task-%s-users-%d", time.Now().Format("2006-01-02T15:04:05Z"), time.Now().UnixNano()%1000)

		// TODO(R-01 / P2): 若 users 数据改为落库模式，在此处启动后台爬取 goroutine。
		// 例：go sch.SyncUsers()

		log.Printf("[admin] /internal/sync/users invoked (no-op: users are fetched on-demand, not persisted)")

		writeJSON(w, map[string]string{
			"task_id":    taskID,
			"started_at": time.Now().Format(time.RFC3339),
			"status":     "accepted", // 不是 "running"——实际没有后台任务在跑
		})
	}
}
