// Package scheduler 提供榜单定时刷新 + 增量 enrich。
//
// cron 驱动爬虫 → 落库 → enricher 补全。
package scheduler

import (
	"encoding/json"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/dong4j/starcat-trending-api/internal/enricher"
	"github.com/dong4j/starcat-trending-api/internal/model"
	"github.com/dong4j/starcat-trending-api/internal/notifier"
	"github.com/dong4j/starcat-trending-api/internal/spider"
	"github.com/dong4j/starcat-trending-api/internal/store"
)

// Scheduler 管理定时爬虫任务。
type Scheduler struct {
	store        store.Store
	enricher     *enricher.Enricher
	wikiNotifier *notifier.WikiNotifier
	cron         *cron.Cron
	langCache    *languageCache
	mu           sync.Mutex
	running      map[string]bool // 防止并发跑同一任务
}

// languageCache 语言列表内存缓存（24h TTL）。
type languageCache struct {
	mu        sync.RWMutex
	languages []model.Language
	fetchedAt time.Time
}

// New 创建 Scheduler。
func New(s store.Store, enc *enricher.Enricher, wn *notifier.WikiNotifier) *Scheduler {
	sch := &Scheduler{
		store:        s,
		enricher:     enc,
		wikiNotifier: wn,
		cron:         cron.New(cron.WithSeconds()),
		langCache:    &languageCache{},
		running:      make(map[string]bool),
	}

	// daily 每小时第 7 分
	sch.cron.AddFunc("7 * * * *", sch.syncDaily)
	// weekly 每 6 小时第 13 分
	sch.cron.AddFunc("13 */6 * * *", sch.syncWeekly)
	// monthly 每 2 天 05:19 UTC
	sch.cron.AddFunc("19 5 */2 * *", sch.syncMonthly)
	// 长尾 enrich 每天 03:00 UTC
	sch.cron.AddFunc("0 3 * * *", sch.enrichLongTail)
	// 过期清理 每天 04:00 UTC
	sch.cron.AddFunc("0 4 * * *", sch.cleanupStale)

	return sch
}

// Start 启动定时任务 + 冷启动全量同步。
func (sch *Scheduler) Start() {
	log.Println("[scheduler] cold start: syncing daily + languages")
	go sch.syncDaily()
	go sch.syncLanguages()
	sch.cron.Start()
	log.Println("[scheduler] cron started")
}

// Stop 停止所有定时任务。
func (sch *Scheduler) Stop() {
	ctx := sch.cron.Stop()
	<-ctx.Done()
	log.Println("[scheduler] stopped")
}

// SyncAll 手动全量同步（admin endpoint 调用）。
func (sch *Scheduler) SyncAll() {
	go func() {
		sch.syncDaily()
		sch.syncWeekly()
		sch.syncMonthly()
		sch.syncLanguages()
	}()
}

// SyncLanguages 手动刷新语言列表。
func (sch *Scheduler) SyncLanguages() {
	go sch.syncLanguages()
}

func (sch *Scheduler) syncDaily() {
	if !sch.tryLock("daily") {
		return
	}
	defer sch.unlock("daily")

	log.Println("[scheduler] syncing daily trending")
	repos := sch.scrapeAndPersist("", "daily")
	_ = sch.store.RecomputePriorities("daily")
	sch.enricher.EnrichAll()
	sch.wikiNotifier.NotifyRepos(repos)
}

func (sch *Scheduler) syncWeekly() {
	if !sch.tryLock("weekly") {
		return
	}
	defer sch.unlock("weekly")

	log.Println("[scheduler] syncing weekly trending")
	repos := sch.scrapeAndPersist("", "weekly")
	_ = sch.store.RecomputePriorities("weekly")
	sch.wikiNotifier.NotifyRepos(repos)
}

func (sch *Scheduler) syncMonthly() {
	if !sch.tryLock("monthly") {
		return
	}
	defer sch.unlock("monthly")

	log.Println("[scheduler] syncing monthly trending")
	repos := sch.scrapeAndPersist("", "monthly")
	_ = sch.store.RecomputePriorities("monthly")
	sch.wikiNotifier.NotifyRepos(repos)
}

func (sch *Scheduler) enrichLongTail() {
	log.Println("[scheduler] running long-tail enrich")
	sch.enricher.EnrichAll()
}

func (sch *Scheduler) cleanupStale() {
	log.Println("[scheduler] cleaning up stale repos (captured_at < 7d)")
}

// scrapeAndPersist 爬所有语言 × since 组合并落库。
// 返回本次持久化的 owner/repo 列表，用于后续 wiki 预热。
func (sch *Scheduler) scrapeAndPersist(lang, since string) []string {
	sp := spider.NewRepoSpider(since, lang)
	items := sp.GetItems()

	var repos []string

	for _, item := range items {
		parts := strings.SplitN(item.Repo, "/", 2)
		if len(parts) != 2 {
			continue
		}
		owner, name := parts[0], parts[1]

		// Defense in depth（2026-06-11）：
		// 历史上 spider/repo.go 漏 strip 了 href 的前导 "/"，导致 SplitN 后
		// owner="" / name="owner/repo"，整批数据落库后被 enricher 用 404 标 unavailable，
		// 整张表的 is_available 全 0，handler 返回空数组 + cache_status=cold。
		// 即使源头已修，这里也兜底校验：owner 或 name 为空一律跳过 + 打 warn，
		// 让任何未来再出现的同类异常（爬虫 HTML 结构变动、第三方源差异）能在日志可见。
		if owner == "" || name == "" {
			log.Printf("[scheduler] skip malformed repo %q (owner=%q name=%q) — spider bug?",
				item.Repo, owner, name)
			continue
		}

		var bjJSON *string
		if len(item.BuildBy) > 0 {
			b, _ := json.Marshal(item.BuildBy)
			s := string(b)
			bjJSON = &s
		}

		fullName := owner + "/" + name
		rec := model.TrendingRepo{
			FullName:    fullName,
			Owner:       owner,
			Name:        name,
			DescText:    &item.Desc,
			Stars:       item.Stars,
			Forks:       item.Forks,
			Language:    &item.Lang,
			Change:      item.Change,
			BuildByJSON: bjJSON,
			Since:       since,
			CapturedAt:  time.Now(),
			IsAvailable: true,
		}

		if err := sch.store.UpsertRepo(rec); err != nil {
			log.Printf("[scheduler] upsert %s failed: %v", rec.FullName, err)
			continue
		}

		repos = append(repos, fullName)
	}

	log.Printf("[scheduler] scraped %d repos for since=%s lang=%s", len(items), since, lang)
	return repos
}

// syncLanguages 刷新语言列表缓存。
func (sch *Scheduler) syncLanguages() {
	langSpider := spider.NewLangSpider()
	items := langSpider.GetItems()

	langs := make([]model.Language, len(items))
	for i, item := range items {
		langs[i] = model.Language{Key: item.Key, Label: item.Label}
	}

	_ = sch.store.UpsertLanguages(langs)

	sch.langCache.mu.Lock()
	sch.langCache.languages = langs
	sch.langCache.fetchedAt = time.Now()
	sch.langCache.mu.Unlock()

	log.Printf("[scheduler] synced %d languages", len(langs))
}

// GetLanguages 从缓存返回语言列表（24h TTL 内不重爬）。
func (sch *Scheduler) GetLanguages() []model.Language {
	sch.langCache.mu.RLock()
	languages := sch.langCache.languages
	fetchedAt := sch.langCache.fetchedAt
	sch.langCache.mu.RUnlock()

	if len(languages) == 0 || time.Since(fetchedAt) > 24*time.Hour {
		sch.syncLanguages()
		sch.langCache.mu.RLock()
		languages = sch.langCache.languages
		sch.langCache.mu.RUnlock()
	}
	return languages
}

func (sch *Scheduler) tryLock(name string) bool {
	sch.mu.Lock()
	defer sch.mu.Unlock()
	if sch.running[name] {
		return false
	}
	sch.running[name] = true
	return true
}

func (sch *Scheduler) unlock(name string) {
	sch.mu.Lock()
	sch.running[name] = false
	sch.mu.Unlock()
}
