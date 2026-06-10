package spider

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dong4j/starcat-trending-api/internal/model"
	"github.com/dong4j/starcat-trending-api/internal/store"
)

type ZreadSpider struct {
	client *http.Client
	store  *store.SQLiteStore
	url    string
}

func NewZreadSpider(s *store.SQLiteStore) *ZreadSpider {
	return &ZreadSpider{
		client: &http.Client{Timeout: 30 * time.Second},
		store:  s,
		url:    "https://zread.ai/api/v1/public/repo/trending",
	}
}

func (s *ZreadSpider) RunOnce(ctx context.Context) error {
	log.Println("[zread] starting fetch trending...")

	result, err := s.fetchAndParse(ctx)
	if err != nil {
		return err
	}

	if s.store == nil {
		return nil // For testing
	}

	now := time.Now()
	for _, group := range result.Data {
		// 推断年份
		year, err := InferYear(group.TimeSpan.Start, now)
		if err != nil {
			log.Printf("[zread] failed to infer year for %s: %v", group.TimeSpan.Start, err)
			continue
		}

		for i, r := range group.Repos {
			fullName := r.Owner + "/" + r.Name
			topics, _ := json.Marshal(r.Topics)
			topicsStr := string(topics)

			repo := model.TrendingRepo{
				FullName:    fullName,
				Owner:       r.Owner,
				Name:        r.Name,
				DescText:    &r.Description,
				Stars:       r.StarCount,
				Language:    &r.Language,
				Since:       "weekly",
				Source:      "zread",
				CapturedAt:  now,
				IsAvailable: true,

				DescriptionZh:     &r.DescriptionZh,
				ZreadWeekStart:    ptrStr(fmt.Sprintf("%d-%s", year, convertMMDD(group.TimeSpan.Start))),
				ZreadWeekEnd:      ptrStr(fmt.Sprintf("%d-%s", year, convertMMDD(group.TimeSpan.End))),
				ZreadWeekLabel:    &group.Title,
				ZreadRankInWeek:   i + 1,
				ZreadWikiID:       &r.WikiID,
				ZreadWeekStartRaw: &group.TimeSpan.Start,
				ZreadWeekEndRaw:   &group.TimeSpan.End,
				ZreadYearInferred: year,
				TopicsJSON:        &topicsStr,
			}

			if err := s.store.UpsertRepo(repo); err != nil {
				log.Printf("[zread] upsert %s error: %v", fullName, err)
			}
		}
	}

	log.Printf("[zread] finished fetch %d groups", len(result.Data))
	return nil
}

func (s *ZreadSpider) fetchAndParse(ctx context.Context) (*ZreadFetchResult, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", s.url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json")
	// 防御 Cloudflare / WAF，模拟真实浏览器请求 JSON
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("zread status: %d", resp.StatusCode)
	}

	var result ZreadFetchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("zread api error: %d %s", result.Code, result.Msg)
	}

	return &result, nil
}

func convertMMDD(mmdd string) string {
	// "08/06" -> "08-06"
	return fmt.Sprintf("%s-%s", mmdd[0:2], mmdd[3:5])
}

func ptrStr(s string) *string {
	return &s
}
