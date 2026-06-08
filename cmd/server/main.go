// Package main is the entry point for starcat-trending-api.
// It serves a GitHub Trending API backed by goquery HTML scraping.
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/dong4j/starcat-trending-api/internal/spider"
)

// ResponseWriter wraps http.ResponseWriter with a JSON helper.
type ResponseWriter struct {
	http.ResponseWriter
}

func (rw *ResponseWriter) JSON(data any) {
	rw.Header().Set("Content-Type", "application/json")
	json.NewEncoder(rw.ResponseWriter).Encode(data)
}

// healthzHandler health check (used by Fly.io http_service.checks)
func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// langHandler lists all available languages.
func langHandler(w http.ResponseWriter, r *http.Request) {
	spiderInstance := spider.NewLangSpider()
	items := spiderInstance.GetItems()
	(&ResponseWriter{ResponseWriter: w}).JSON(items)
}

// repoHandler lists trending repositories.
// Query params: lang (optional), since (optional, default daily)
func repoHandler(w http.ResponseWriter, r *http.Request) {
	lang := r.URL.Query().Get("lang")
	since := r.URL.Query().Get("since")
	if since == "" {
		since = "daily"
	}

	spiderInstance := spider.NewRepoSpider(since, lang)
	items := spiderInstance.GetItems()
	(&ResponseWriter{ResponseWriter: w}).JSON(items)
}

// userHandler lists trending developers.
// Query params: lang (optional), since (optional, default daily), sponsorable (optional)
func userHandler(w http.ResponseWriter, r *http.Request) {
	lang := r.URL.Query().Get("lang")
	since := r.URL.Query().Get("since")
	if since == "" {
		since = "daily"
	}
	sponsorable := r.URL.Query().Get("sponsorable")

	spiderInstance := spider.NewUserSpider(since, lang, sponsorable)
	items := spiderInstance.GetItems()
	(&ResponseWriter{ResponseWriter: w}).JSON(items)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "5002"
	}

	// Register routes (Go 1.22+ style: custom mux + method-aware paths)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthzHandler)
	mux.HandleFunc("GET /lang", langHandler)
	mux.HandleFunc("GET /repo", repoHandler)
	mux.HandleFunc("GET /user", userHandler)

	// Graceful shutdown on SIGINT / SIGTERM
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("Received shutdown signal, closing service...")
		os.Exit(0)
	}()

	log.Printf("starcat-trending-api starting on port %s", port)
	log.Printf("Endpoints:")
	log.Printf("  GET /healthz  - Health check")
	log.Printf("  GET /lang     - Get all available languages")
	log.Printf("  GET /repo     - Get trending repositories (params: lang, since)")
	log.Printf("  GET /user     - Get trending developers (params: lang, since, sponsorable)")

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
