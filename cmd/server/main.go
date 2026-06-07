// Package main 提供 GitHub Trending API 服务器
// 使用 Go 标准库实现 REST API，与 Python FastAPI 版本保持相同的端点和响应格式
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github-trending-api/internal/spider"
)

// ResponseWriter 包装 http.ResponseWriter 以提供 JSON 编码方法
type ResponseWriter struct {
	http.ResponseWriter
}

func (rw *ResponseWriter) JSON(data interface{}) {
	rw.Header().Set("Content-Type", "application/json")
	json.NewEncoder(rw.ResponseWriter).Encode(data)
}

// rootHandler 根路径处理函数
func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"message": "Hello GitHub trending"})
}

// langHandler 获取所有可用语言
func langHandler(w http.ResponseWriter, r *http.Request) {
	spiderInstance := spider.NewLangSpider()
	items := spiderInstance.GetItems()
	(&ResponseWriter{ResponseWriter: w}).JSON(items)
}

// repoHandler 获取 trending 仓库列表
// Query 参数: lang (可选), since (可选，默认 daily)
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

// userHandler 获取 trending 开发者列表
// Query 参数: lang (可选), since (可选，默认 daily), sponsorable (可选)
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

	// 注册路由
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/lang", langHandler)
	http.HandleFunc("/repo", repoHandler)
	http.HandleFunc("/user", userHandler)

	log.Printf("GitHub Trending API server starting on port %s", port)
	log.Printf("Endpoints:")
	log.Printf("  GET /         - Welcome message")
	log.Printf("  GET /lang     - Get all available languages")
	log.Printf("  GET /repo     - Get trending repositories (params: lang, since)")
	log.Printf("  GET /user     - Get trending developers (params: lang, since, sponsorable)")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
