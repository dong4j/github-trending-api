// Package enricher 提供 GitHub API 字段补全 + Rate Limit 退避逻辑。
//
// queue.go: 优先级队列 + worker pool。
// 按 enrich_priority DESC 从 SQLite 取待处理 repo，分发到 worker 并发 enrich。
//
// R-01 v1.2 §4.2: 不是简单 FIFO，scheduler RecomputePriorities 后
// worker 每次取 ORDER BY enrich_priority DESC LIMIT N。
package enricher

import (
	"log"
	"sync"
	"time"
)

// EnrichQueue 基于 SQL 优先级排序的异步 enrich 队列。
//
// Worker pool 持续从 store 拉取 enrich_priority 最高的未处理 repo，
// 确保榜单前排优先补全。受 RateLimitHandler 约束，worker 在请求前
// 必须调 rateLimit.Wait()。
type EnrichQueue struct {
	enricher  *Enricher
	workerCnt int

	mu      sync.Mutex
	running bool
	stopCh  chan struct{}
	wg      sync.WaitGroup

	// 统计
	enriched int64
	failed   int64
}

// NewEnrichQueue 创建 worker pool。
// workerCnt 默认 2（受 GitHub API 5000 req/h 约束，2 worker 即够）。
func NewEnrichQueue(enc *Enricher, workerCnt int) *EnrichQueue {
	if workerCnt <= 0 {
		workerCnt = 2
	}
	return &EnrichQueue{
		enricher:  enc,
		workerCnt: workerCnt,
		stopCh:    make(chan struct{}),
	}
}

// Start 启动后台 worker pool，持续 enrich 待处理 repo。
// 重复调用安全（第二次调用无操作）。
func (q *EnrichQueue) Start() {
	q.mu.Lock()
	if q.running {
		q.mu.Unlock()
		return
	}
	q.running = true
	q.mu.Unlock()

	for i := 0; i < q.workerCnt; i++ {
		q.wg.Add(1)
		go q.worker(i)
	}
	log.Printf("[enricher-queue] %d workers started", q.workerCnt)
}

// Stop 优雅关闭所有 worker，等待当前 enrich 完成后退出。
func (q *EnrichQueue) Stop() {
	q.mu.Lock()
	if !q.running {
		q.mu.Unlock()
		return
	}
	q.running = false
	q.mu.Unlock()

	close(q.stopCh)
	q.wg.Wait()
	log.Printf("[enricher-queue] stopped (enriched=%d failed=%d)", q.enriched, q.failed)
}

// ProcessAll 同步全量 enrich（供 scheduler cron 任务使用）。
// 与后台 worker 互斥：ProcessAll 前后台 worker 不会同时处理同一批 repo，
// 因为 GetUnenrichedRepos 返回的 repo 会立刻被 UpdateEnriched 标记为已处理。
func (q *EnrichQueue) ProcessAll() {
	q.enricher.EnrichAll()
}

// Stats 返回 enrich 计数。
func (q *EnrichQueue) Stats() (enriched, failed int64) {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.enriched, q.failed
}

func (q *EnrichQueue) worker(id int) {
	defer q.wg.Done()

	for {
		select {
		case <-q.stopCh:
			return
		default:
		}

		// 取优先��最高的单个未处理 repo
		repos, err := q.enricher.store.GetUnenrichedRepos(1)
		if err != nil {
			log.Printf("[enricher-queue] worker %d: store error: %v", id, err)
			q.sleepOrStop(10 * time.Second)
			continue
		}

		if len(repos) == 0 {
			// 无待处理 repo，空闲等待
			q.sleepOrStop(30 * time.Second)
			continue
		}

		if err := q.enricher.EnrichOne(&repos[0]); err != nil {
			log.Printf("[enricher-queue] worker %d: %s failed: %v", id, repos[0].FullName, err)
			q.mu.Lock()
			q.failed++
			q.mu.Unlock()
		} else {
			q.mu.Lock()
			q.enriched++
			q.mu.Unlock()
		}
	}
}

// sleepOrStop 等待指定时长或收到停止信号。
func (q *EnrichQueue) sleepOrStop(d time.Duration) {
	select {
	case <-q.stopCh:
	case <-time.After(d):
	}
}
