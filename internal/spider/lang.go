package spider

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// LangSpider 语言列表爬虫
// 对应 Python 版本的 LangSpider
type LangSpider struct {
	*BaseRequest
}

// NewLangSpider 创建语言列表爬虫
func NewLangSpider() *LangSpider {
	return &LangSpider{
		BaseRequest: NewBaseRequest(),
	}
}

// GetURL 获取语言列表页 URL
func (l *LangSpider) GetURL() string {
	return l.BaseURL
}

// Parse 解析语言列表页
func (l *LangSpider) Parse(html string) []LangItem {
	var items []LangItem

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return items
	}

	// 找到 #languages-menuitems 下的 [data-filter-list] 中的 a 标签
	doc.Find("#languages-menuitems [data-filter-list] a").Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		span := s.Find("span")
		label := strings.TrimSpace(span.Text())

		// 提取 key: 移除 /trending 前缀和 /，然后按 ? 分割取第一部分
		// 例如: "/trending/python?since=daily" -> "python"
		// "/trending/c%23?since=daily" -> "c%23"
		key := regexp.MustCompile(`^/trending|/`).ReplaceAllString(href, "")
		key = strings.Split(key, "?")[0]

		items = append(items, LangItem{
			Label: label,
			Key:   key,
		})
	})

	return items
}

// GetItems 获取语言列表
func (l *LangSpider) GetItems() []LangItem {
	html, err := l.Fetch(l.GetURL())
	if err != nil {
		return []LangItem{}
	}
	return l.Parse(html)
}
