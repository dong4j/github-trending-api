package spider

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// UserSpider 开发者 trending 爬虫
// 对应 Python 版本的 UserSpider
type UserSpider struct {
	*BaseRequest
	Since       string
	Lang        string
	Sponsorable string
}

// NewUserSpider 创建开发者爬虫
func NewUserSpider(since, lang, sponsorable string) *UserSpider {
	return &UserSpider{
		BaseRequest: NewBaseRequest(),
		Since:       since,
		Lang:        lang,
		Sponsorable: sponsorable,
	}
}

// GetURL 获取开发者列表 URL
// URL 格式: https://github.com/trending/developers/{lang}?since={since}&sponsorable={sponsorable}
func (u *UserSpider) GetURL() string {
	base := u.BaseURL
	if u.Lang != "" {
		langEncoded := encodeLangParam(u.Lang)
		base = fmt.Sprintf("%s/developers/%s", base, langEncoded)
	} else {
		base = fmt.Sprintf("%s/developers", base)
	}

	url := fmt.Sprintf("%s?since=%s", base, u.Since)
	if u.Sponsorable != "" {
		url = fmt.Sprintf("%s&sponsorable=%s", url, u.Sponsorable)
	}
	return url
}

// Parse 解析开发者列表页
func (u *UserSpider) Parse(html string) []UserItem {
	var items []UserItem

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return items
	}

	// 找到 article[id^="pa-"] 的元素
	doc.Find("article[id^=pa-]").Each(func(i int, article *goquery.Selection) {
		item := UserItem{}

		// 获取头像
		imgSel := article.Find("div > a:has(img) img")
		if imgSel.Length() > 0 {
			item.Avatar, _ = imgSel.Attr("src")
		}

		// 获取 name (可能在 h1 中)
		nameSel := article.Find("div > h1:has(a)")
		if nameSel.Length() > 0 {
			item.Name = strings.TrimSpace(nameSel.Find("a").Text())
		}

		// 获取 github_name
		githubSel := article.Find("div > p:has(a)")
		if githubSel.Length() > 0 {
			item.GitHubName, _ = githubSel.Find("a").Attr("href")
		}

		// 获取热门仓库
		// 嵌套的 article article 中的 h1 a 是仓库链接
		// h1 + div 是描述
		innerArticle := article.Find("article article")
		if innerArticle.Length() > 0 {
			repoLink := innerArticle.Find("h1 a")
			repo := ""
			if repoLink.Length() > 0 {
				repo, _ = repoLink.Attr("href")
			}

			// 描述在 h1 + div 中
			desc := ""
			innerArticle.Find("h1 + div").Each(func(_ int, s *goquery.Selection) {
				desc = strings.TrimSpace(s.Text())
			})

			item.Popular = &PopItem{
				Repo: repo,
				Desc: desc,
			}
		}

		items = append(items, item)
	})

	return items
}

// GetItems 获取 trending 开发者列表
func (u *UserSpider) GetItems() []UserItem {
	html, err := u.Fetch(u.GetURL())
	if err != nil {
		return []UserItem{}
	}
	return u.Parse(html)
}
