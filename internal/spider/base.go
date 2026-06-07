// Package spider 实现 GitHub Trending 页面爬虫
package spider

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// BaseRequest 爬虫基类
type BaseRequest struct {
	BaseURL string
	Client  *http.Client
}

// NewBaseRequest 创建新的爬虫请求实例
func NewBaseRequest() *BaseRequest {
	return &BaseRequest{
		BaseURL: "https://github.com/trending",
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetURL 获取请求 URL
func (b *BaseRequest) GetURL() string {
	return b.BaseURL
}

// Fetch 发起 HTTP 请求并返回响应内容
func (b *BaseRequest) Fetch(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	// 模拟浏览器 User-Agent
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	resp, err := b.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("HTTP status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
