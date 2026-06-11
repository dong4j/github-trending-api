// Package handler 的 endpoint 测试：repos.go
//
// 覆盖 HandleReposV1 的 7 个 query 场景：
//  1. 正常返回（默认 since=daily, 无 lang, 无 limit）
//  2. since=daily / weekly / monthly
//  3. since 非法值（yearly）→ 400
//  4. source=github 拒绝 → 400 + 引导 weekly-api
//  5. source=zread 拒绝 → 400 + 引导 weekly-api
//  6. lang 过滤透传
//  7. limit 上限 clamp 到 100
//  8. 缓存状态：返回 0 条时 cacheStatus=cold，否则 fresh
//  9. 内部 store 错误 → 500
//
// fakeStore 实现 store.Store interface,只覆盖 GetRepos 调用路径，
// 其他方法 panic（不调）。
package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/dong4j/starcat-trending-api/internal/model"
	"github.com/dong4j/starcat-trending-api/internal/store"
)

// fakeStore 是 store.Store 的最小实现,只支持 GetRepos 的可观测调用。
type fakeStore struct {
	repos        []model.TrendingRepo
	gotSince     string
	gotLang      string
	gotLimit     int
	callCount    int
	forceGetErr  error
}

func (f *fakeStore) GetRepos(since, lang string, limit int) ([]model.TrendingRepo, error) {
	f.callCount++
	f.gotSince = since
	f.gotLang = lang
	f.gotLimit = limit
	if f.forceGetErr != nil {
		return nil, f.forceGetErr
	}
	return f.repos, nil
}

// 其它方法不调,panic 提示
func (f *fakeStore) UpsertRepo(r model.TrendingRepo) error {
	panic(fmt.Sprintf("UpsertRepo should not be called in handler test, got %s/%s", r.FullName, r.Since))
}
func (f *fakeStore) GetUnenrichedRepos(limit int) ([]model.TrendingRepo, error) {
	panic("GetUnenrichedRepos not used in handler test")
}
func (f *fakeStore) UpdateEnriched(fullName, since string, r model.TrendingRepo) error {
	panic("UpdateEnriched not used in handler test")
}
func (f *fakeStore) MarkUnavailable(fullName, since string) error {
	panic("MarkUnavailable not used in handler test")
}
func (f *fakeStore) RecomputePriorities(since string) error {
	panic("RecomputePriorities not used in handler test")
}
func (f *fakeStore) ResetAllEnriched() error {
	panic("ResetAllEnriched not used in handler test")
}
func (f *fakeStore) UpsertLanguages(langs []model.Language) error {
	panic("UpsertLanguages not used in handler test")
}
func (f *fakeStore) GetLanguages() ([]model.Language, error) {
	panic("GetLanguages not used in handler test")
}
func (f *fakeStore) Close() error { return nil }

// 不实现的接口：让 fakeStore 类型上仍然实现 store.Store,必须补完所有方法
// 注：上面已经全部定义。

// 编译期断言 fakeStore 实现了 store.Store
var _ store.Store = (*fakeStore)(nil)

// 辅助：构造一条 enriched repo（UpsertRepo 不传 is_available,这里手工设置）
func makeRepo(name, lang string, stars int) model.TrendingRepo {
	langPtr := &lang
	desc := "desc of " + name
	enriched := time.Now()
	return model.TrendingRepo{
		FullName:   "owner/" + name,
		Owner:      "owner",
		Name:       name,
		DescText:   &desc,
		Stars:      stars,
		Forks:      stars / 10,
		Language:   langPtr,
		Change:     1,
		Since:      "daily",
		CapturedAt: time.Now(),
		EnrichedAt: &enriched,
		IsAvailable: true,
	}
}

// doReq 调 HandleReposV1 走一次 HTTP,返回 *httptest.ResponseRecorder。
func doReq(s store.Store, query string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/repos?"+query, nil)
	HandleReposV1(s)(w, r)
	return w
}

// decodeEnvelope 解码 envelope 到具体 data 类型。
func decodeEnvelope[T any](t *testing.T, w *httptest.ResponseRecorder) model.Envelope[T] {
	t.Helper()
	var env model.Envelope[T]
	if err := decodeJSON(w.Body.Bytes(), &env); err != nil {
		t.Fatalf("decode envelope: %v (body: %s)", err, w.Body.String())
	}
	return env
}

func decodeErrorEnv(t *testing.T, w *httptest.ResponseRecorder) model.ErrorEnvelope {
	t.Helper()
	var env model.ErrorEnvelope
	if err := decodeJSON(w.Body.Bytes(), &env); err != nil {
		t.Fatalf("decode error envelope: %v (body: %s)", err, w.Body.String())
	}
	return env
}

// TestRepos_DefaultSince 验证默认 since=daily。
func TestRepos_DefaultSince(t *testing.T) {
	f := &fakeStore{
		repos: []model.TrendingRepo{makeRepo("r1", "go", 100)},
	}
	w := doReq(f, "")
	env := decodeEnvelope[[]model.StarcatRepoCardDTO](t, w)
	if env.Meta == nil || env.Meta.Since != "daily" {
		t.Errorf("default since: want daily, got %+v", env.Meta)
	}
	if f.gotSince != "daily" {
		t.Errorf("store got since: want daily, got %s", f.gotSince)
	}
}

// TestRepos_ValidSince 验证 3 个合法 since 都通过。
func TestRepos_ValidSince(t *testing.T) {
	for _, since := range []string{"daily", "weekly", "monthly"} {
		t.Run(since, func(t *testing.T) {
			f := &fakeStore{}
			w := doReq(f, "since="+since)
			if w.Code != http.StatusOK {
				t.Errorf("since=%s should be accepted, got %d", since, w.Code)
			}
			if f.gotSince != since {
				t.Errorf("store got since: want %s, got %s", since, f.gotSince)
			}
		})
	}
}

// TestRepos_InvalidSince 验证非法 since → 400。
func TestRepos_InvalidSince(t *testing.T) {
	f := &fakeStore{}
	w := doReq(f, "since=yearly")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("since=yearly: want 400, got %d", w.Code)
	}
	env := decodeErrorEnv(t, w)
	if env.Error.Code != "BAD_REQUEST" {
		t.Errorf("code: want BAD_REQUEST, got %s", env.Error.Code)
	}
	details, ok := env.Error.Details.(map[string]interface{})
	if !ok {
		t.Fatalf("details type: %T", env.Error.Details)
	}
	if details["param"] != "since" {
		t.Errorf("details.param: want since, got %v", details["param"])
	}
	if details["got"] != "yearly" {
		t.Errorf("details.got: want yearly, got %v", details["got"])
	}
	if f.callCount != 0 {
		t.Errorf("store should not be called on invalid since, got %d calls", f.callCount)
	}
}

// TestRepos_SourceRejected 验证 source= 任何值都拒绝。
func TestRepos_SourceRejected(t *testing.T) {
	cases := []string{"github", "zread", "merged", ""} // 注意 "" 是默认值,不会拒绝
	for _, src := range cases[:3] {                      // github / zread / merged
		t.Run("source="+src, func(t *testing.T) {
			f := &fakeStore{}
			w := doReq(f, "source="+src)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("source=%s: want 400, got %d", src, w.Code)
			}
			env := decodeErrorEnv(t, w)
			if env.Error.Code != "BAD_REQUEST" {
				t.Errorf("code: want BAD_REQUEST, got %s", env.Error.Code)
			}
			// 错误信息必须引导 weekly-api
			if !strings.Contains(env.Error.Message, "weekly-api") {
				t.Errorf("error message should mention weekly-api, got: %q", env.Error.Message)
			}
			if f.callCount != 0 {
				t.Errorf("store should not be called on rejected source, got %d calls", f.callCount)
			}
		})
	}
}

// TestRepos_LangFilter 透传 lang 参数。
func TestRepos_LangFilter(t *testing.T) {
	f := &fakeStore{}
	w := doReq(f, "lang=swift")
	if w.Code != http.StatusOK {
		t.Fatalf("lang=swift: want 200, got %d", w.Code)
	}
	if f.gotLang != "swift" {
		t.Errorf("store got lang: want swift, got %s", f.gotLang)
	}
	env := decodeEnvelope[[]model.StarcatRepoCardDTO](t, w)
	if env.Meta == nil || env.Meta.Language != "swift" {
		t.Errorf("meta.language: want swift, got %+v", env.Meta)
	}
}

// TestRepos_LimitClampTo100 验证 limit > 100 被 clamp。
func TestRepos_LimitClampTo100(t *testing.T) {
	f := &fakeStore{}
	w := doReq(f, "limit=500")
	if w.Code != http.StatusOK {
		t.Fatalf("limit=500: want 200, got %d", w.Code)
	}
	if f.gotLimit != 100 {
		t.Errorf("limit should clamp to 100, store got %d", f.gotLimit)
	}
}

// TestRepos_LimitCustom 验证 limit 正常值透传。
func TestRepos_LimitCustom(t *testing.T) {
	f := &fakeStore{}
	doReq(f, "limit=30")
	if f.gotLimit != 30 {
		t.Errorf("limit=30: store got %d, want 30", f.gotLimit)
	}
}

// TestRepos_CacheStatusFresh 验证有数据时 cacheStatus=fresh。
func TestRepos_CacheStatusFresh(t *testing.T) {
	f := &fakeStore{
		repos: []model.TrendingRepo{makeRepo("r1", "go", 100)},
	}
	w := doReq(f, "")
	env := decodeEnvelope[[]model.StarcatRepoCardDTO](t, w)
	if env.Meta == nil || env.Meta.CacheStatus != "fresh" {
		t.Errorf("cacheStatus: want fresh, got %+v", env.Meta)
	}
}

// TestRepos_CacheStatusCold 验证 0 条时 cacheStatus=cold。
func TestRepos_CacheStatusCold(t *testing.T) {
	f := &fakeStore{repos: nil}
	w := doReq(f, "")
	env := decodeEnvelope[[]model.StarcatRepoCardDTO](t, w)
	if env.Meta == nil || env.Meta.CacheStatus != "cold" {
		t.Errorf("cacheStatus: want cold, got %+v", env.Meta)
	}
}

// TestRepos_StoreError 验证 store 错误 → 500。
func TestRepos_StoreError(t *testing.T) {
	f := &fakeStore{forceGetErr: errors.New("db locked")}
	w := doReq(f, "")
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("store error: want 500, got %d", w.Code)
	}
	env := decodeErrorEnv(t, w)
	if env.Error.Code != "INTERNAL_ERROR" {
		t.Errorf("code: want INTERNAL_ERROR, got %s", env.Error.Code)
	}
}

// TestRepos_CardDTOConversion 验证 repo → card DTO 转换（含 contributors 解析）。
func TestRepos_CardDTOConversion(t *testing.T) {
	buildBy := `[{"by":"/alice","avatar":"https://x/a.png"},{"by":"/bob","avatar":"https://x/b.png"}]`
	r := makeRepo("r1", "go", 100)
	r.BuildByJSON = &buildBy
	r.Description = ptrStr("github official desc") // enricher 写的 description
	f := &fakeStore{repos: []model.TrendingRepo{r}}
	w := doReq(f, "")
	env := decodeEnvelope[[]model.StarcatRepoCardDTO](t, w)
	if len(env.Data) != 1 {
		t.Fatalf("want 1 card, got %d", len(env.Data))
	}
	card := env.Data[0]
	if card.FullName != "owner/r1" {
		t.Errorf("full_name: want owner/r1, got %s", card.FullName)
	}
	if card.Language == nil || *card.Language != "go" {
		t.Errorf("language: want go, got %v", card.Language)
	}
	if card.Description == nil || *card.Description != "github official desc" {
		t.Errorf("description should use github desc, got %v", card.Description)
	}
	if card.HTMLURL == nil || *card.HTMLURL != "https://github.com/owner/r1" {
		t.Errorf("html_url: want https://github.com/owner/r1, got %v", card.HTMLURL)
	}
	// Trending 扩展段
	if card.Trending == nil {
		t.Fatalf("trending extension should be present")
	}
	if card.Trending.Change != 1 {
		t.Errorf("trending.change: want 1, got %d", card.Trending.Change)
	}
	if len(card.Trending.Contributors) != 2 {
		t.Errorf("contributors: want 2, got %d", len(card.Trending.Contributors))
	}
	// By 字段 "/alice" 去掉前缀成 "alice"
	if card.Trending.Contributors[0].Login != "alice" {
		t.Errorf("contributor[0].login: want alice, got %s", card.Trending.Contributors[0].Login)
	}
}

// TestRepos_DescriptionFallbackToDescText 验证 description 空时回退到 desc_text。
func TestRepos_DescriptionFallbackToDescText(t *testing.T) {
	r := makeRepo("r1", "go", 100)
	r.Description = nil // enricher 还没写
	f := &fakeStore{repos: []model.TrendingRepo{r}}
	w := doReq(f, "")
	env := decodeEnvelope[[]model.StarcatRepoCardDTO](t, w)
	card := env.Data[0]
	if card.Description == nil || *card.Description != "desc of r1" {
		t.Errorf("description should fall back to desc_text, got %v", card.Description)
	}
}

// --- helpers ---

func ptrStr(s string) *string { return &s }

func decodeJSON(b []byte, v interface{}) error {
	return json.Unmarshal(b, v)
}
