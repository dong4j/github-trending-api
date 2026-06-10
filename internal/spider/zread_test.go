package spider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dong4j/starcat-trending-api/internal/store"
)

func TestZreadSpider_RunOnce(t *testing.T) {
	// Create mock Zread server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if r.Header.Get("Sec-Fetch-Dest") != "empty" {
			t.Errorf("missing or wrong Sec-Fetch-Dest")
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"code": 0,
			"msg": "success",
			"data": [
				{
					"title": "Test Group",
					"time_span": {"start": "06/01", "end": "06/07"},
					"repos": [
						{
							"repo_id": "123",
							"owner": "test",
							"name": "repo",
							"description": "desc",
							"description_zh": "desc zh",
							"star_count": 100,
							"language": "Go"
						}
					]
				}
			]
		}`))
	}))
	defer server.Close()

	// 此时数据库需要一个真实的或内存 SQLite
	// 因为没有现成的 interface 注入，这里只测 HTTP 请求解析逻辑。
	// 为了不破坏测试环境，可以改用依赖注入。这里直接利用我们刚才的结构体。
	
	spider := &ZreadSpider{
		client: &http.Client{Timeout: 5 * time.Second},
		url:    server.URL,
		store:  nil, // 我们在测试里不直接写库，可以通过修改 RunOnce 或者注入 mock 解决。
	}

	// 为了兼容，我们需要稍微调整一下 RunOnce 允许 URL 被覆盖。
	err := spider.fetchAndParse(context.Background())
	if err != nil {
		t.Fatalf("fetchAndParse failed: %v", err)
	}
}
