// Package main 是 starcat-trending-api 的入口。
//
// R-01 v1.2: 从无状态爬虫升级为三层架构
//
//	spider（爬虫）→ store（SQLite）→ enricher（GitHub API 补全）→ scheduler（cron）
//	+ Bearer Token 鉴权 + Token Pool + /api/v1/* 契约。
package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/dong4j/starcat-trending-api/internal/enricher"
	"github.com/dong4j/starcat-trending-api/internal/handler"
	"github.com/dong4j/starcat-trending-api/internal/middleware"
	"github.com/dong4j/starcat-trending-api/internal/notifier"
	"github.com/dong4j/starcat-trending-api/internal/scheduler"
	"github.com/dong4j/starcat-trending-api/internal/store"
	"github.com/dong4j/starcat-trending-api/internal/tokenpool"
)

func main() {
	// 加载 .env
	if err := godotenv.Load(); err != nil {
		log.Printf("[env] no .env file found, using OS environment only")
	} else {
		log.Printf("[env] .env loaded")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "5002"
	}

	storeFile := os.Getenv("STORE_FILE")
	if storeFile == "" {
		storeFile = "./trending.db"
	}

	apiKeysStr := os.Getenv("API_KEYS")
	if apiKeysStr == "" {
		log.Fatal("API_KEYS env is required")
	}
	apiKeys := strings.Split(apiKeysStr, ",")

	tokensStr := os.Getenv("GITHUB_TOKENS")
	if tokensStr == "" {
		log.Fatal("GITHUB_TOKENS env is required (at least 1 GitHub PAT)")
	}
	tokens := strings.Split(tokensStr, ",")

	// SQLite Store
	sqliteStore, err := store.NewSQLiteStore(storeFile)
	if err != nil {
		log.Fatalf("Failed to initialize SQLite: %v", err)
	}
	defer sqliteStore.Close()

	// Token Pool
	pool := tokenpool.New(tokens)

	// Rate Limit Handler（5000 req/h → 720ms/req 兜底间隔）
	rateLimitHandler := enricher.NewRateLimitHandler(720 * time.Millisecond)

	// Enricher
	enc := enricher.New(sqliteStore, pool, rateLimitHandler)

	// Enrich Queue（后台 worker pool，持续处理 priority 最高的待 enrich repo）
	enrichQueue := enricher.NewEnrichQueue(enc, 2)
	enrichQueue.Start()
	defer enrichQueue.Stop()

	// Wiki Notifier（增量预热 wiki-api 缓存，通过 WIKI_API_KEY 控制开关）
	wikiNotifier := notifier.NewWikiNotifier()

	// Scheduler
	sch := scheduler.New(sqliteStore, enc, wikiNotifier)

	// Bearer 鉴权中间件
	authMW := middleware.NewBearerAuth(apiKeys)

	// 路由
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthzHandler)
	mux.Handle("GET /api/v1/repos", authMW.Wrap(handler.HandleReposV1(sqliteStore)))
	mux.Handle("GET /api/v1/languages", authMW.Wrap(handler.HandleLanguagesV1(sch)))
	mux.Handle("GET /api/v1/users", authMW.Wrap(handler.HandleUsersV1()))
	mux.Handle("POST /internal/sync/repos", authMW.Wrap(handler.HandleAdminSyncRepos(sch)))
	mux.Handle("POST /internal/sync/languages", authMW.Wrap(handler.HandleAdminSyncLanguages(sch)))
	mux.Handle("POST /internal/sync/users", authMW.Wrap(handler.HandleAdminSyncUsers()))

	// 优雅关闭
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigCh
		log.Printf("Received %v, shutting down...", sig)
		sch.Stop()
		enrichQueue.Stop()
		sqliteStore.Close()
		os.Exit(0)
	}()

	// 冷启动：爬 daily + 语言列表
	go sch.Start()

	log.Printf("starcat-trending-api starting on port %s", port)
	log.Printf("Endpoints:")
	log.Printf("  GET  /api/v1/repos          - Trending repos (auth required)")
	log.Printf("  GET  /api/v1/languages      - Languages list (auth required)")
	log.Printf("  GET  /api/v1/users          - Trending developers (auth required)")
	log.Printf("  POST /internal/sync/repos    - Manual sync trigger (auth required)")
	log.Printf("  POST /internal/sync/languages - Languages refresh (auth required)")
	log.Printf("  POST /internal/sync/users    - Developers refresh (auth required)")
	log.Printf("  GET  /healthz               - Health check (public)")
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
