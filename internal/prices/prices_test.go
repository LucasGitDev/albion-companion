package prices

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestRegionFromIP(t *testing.T) {
	cases := []struct {
		ip   string
		want Region
		ok   bool
	}{
		{"5.188.125.10", West, true},
		{"5.45.187.3", East, true},
		{"193.169.238.200", Europe, true},
		{"192.168.0.1", Region{}, false}, // LAN, not a game server
		{"8.8.8.8", Region{}, false},
		{"", Region{}, false},
	}
	for _, c := range cases {
		got, ok := RegionFromIP(c.ip)
		if ok != c.ok || got.Key != c.want.Key {
			t.Errorf("RegionFromIP(%q) = %v,%v want %v,%v", c.ip, got.Key, ok, c.want.Key, c.ok)
		}
	}
}

func TestRegionByKey(t *testing.T) {
	if r, ok := RegionByKey("europe"); !ok || r.ID != 3 {
		t.Fatalf("RegionByKey(europe) = %+v,%v", r, ok)
	}
	if _, ok := RegionByKey("mars"); ok {
		t.Fatal("RegionByKey(mars) should be false")
	}
}

// pricesJSON is a minimal AO Data /stats/prices response: two cities for T4_BAG
// (best sell 1200), and T5_SOUL with no sell orders (falls back to buy 800).
const pricesJSON = `[
	{"item_id":"T4_BAG","city":"Caerleon","quality":1,"sell_price_min":1000,"buy_price_max":400},
	{"item_id":"T4_BAG","city":"Martlock","quality":1,"sell_price_min":1200,"buy_price_max":500},
	{"item_id":"T5_SOUL","city":"Bridgewatch","quality":1,"sell_price_min":0,"buy_price_max":800}
]`

// testServer records the item ids requested and returns the canned response.
func testServer(t *testing.T, requested *[]string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const prefix = "/api/v2/stats/prices/"
		path := strings.TrimPrefix(r.URL.Path, prefix)
		path = strings.TrimSuffix(path, ".json")
		if requested != nil {
			*requested = append(*requested, path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(pricesJSON))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func regionAt(url string) Region {
	return Region{ID: 1, Key: "west", Label: "Test", BaseURL: url}
}

func TestRefreshAndLookup(t *testing.T) {
	var requested []string
	srv := testServer(t, &requested)
	c := New(Options{HTTP: srv.Client(), Region: regionAt(srv.URL)})

	if err := c.Refresh(context.Background(), []string{"T4_BAG", "T5_SOUL"}); err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	prices, complete := c.Lookup([]string{"T4_BAG", "T5_SOUL"})
	if !complete {
		t.Fatal("expected complete lookup after refresh")
	}
	if prices["T4_BAG"] != 1200 { // best sell_price_min across cities
		t.Errorf("T4_BAG = %d, want 1200", prices["T4_BAG"])
	}
	if prices["T5_SOUL"] != 800 { // fallback to buy_price_max
		t.Errorf("T5_SOUL = %d, want 800", prices["T5_SOUL"])
	}
	if len(requested) != 1 {
		t.Fatalf("expected 1 request, got %d: %v", len(requested), requested)
	}
}

func TestLookupSchedulesBackgroundFetch(t *testing.T) {
	var mu sync.Mutex
	var requested []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requested = append(requested, r.URL.Path)
		mu.Unlock()
		_, _ = w.Write([]byte(pricesJSON))
	}))
	defer srv.Close()

	done := make(chan struct{}, 1)
	c := New(Options{
		HTTP:   srv.Client(),
		Region: regionAt(srv.URL),
		OnUpdate: func() {
			select {
			case done <- struct{}{}:
			default:
			}
		},
	})

	// First lookup: cache cold, returns nothing but schedules a fetch.
	prices, complete := c.Lookup([]string{"T4_BAG"})
	if complete || len(prices) != 0 {
		t.Fatalf("cold lookup = %v,%v, want empty,false", prices, complete)
	}
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("background fetch did not complete")
	}
	// Second lookup: now served from cache.
	prices, complete = c.Lookup([]string{"T4_BAG"})
	if !complete || prices["T4_BAG"] != 1200 {
		t.Fatalf("warm lookup = %v,%v", prices, complete)
	}
}

func TestCacheTTLExpiry(t *testing.T) {
	var requested []string
	srv := testServer(t, &requested)
	now := time.Unix(1000, 0)
	c := New(Options{
		HTTP:   srv.Client(),
		Region: regionAt(srv.URL),
		TTL:    time.Minute,
		Now:    func() time.Time { return now },
	})
	if err := c.Refresh(context.Background(), []string{"T4_BAG"}); err != nil {
		t.Fatal(err)
	}
	if _, complete := c.Lookup([]string{"T4_BAG"}); !complete {
		t.Fatal("fresh entry should be complete")
	}
	now = now.Add(2 * time.Minute) // expire
	if _, complete := c.Lookup([]string{"T4_BAG"}); complete {
		t.Fatal("stale entry should not be complete")
	}
}

// errDoer always fails, simulating offline / no network.
type errDoer struct{}

func (errDoer) Do(*http.Request) (*http.Response, error) { return nil, errors.New("offline") }

func TestFailSoftOffline(t *testing.T) {
	c := New(Options{HTTP: errDoer{}, Region: West})
	// Refresh surfaces the error for logging but does not panic.
	if err := c.Refresh(context.Background(), []string{"T4_BAG"}); err == nil {
		t.Fatal("expected an error from offline refresh")
	}
	// Lookup must still return cleanly with no prices — engine unaffected.
	prices, complete := c.Lookup([]string{"T4_BAG"})
	if complete || len(prices) != 0 {
		t.Fatalf("offline lookup = %v,%v, want empty,false", prices, complete)
	}
}

func TestSetRegionIsolatesCache(t *testing.T) {
	srv := testServer(t, nil)
	c := New(Options{HTTP: srv.Client(), Region: regionAt(srv.URL)})
	if err := c.Refresh(context.Background(), []string{"T4_BAG"}); err != nil {
		t.Fatal(err)
	}
	// Switch to a different region key pointing at the same test server: the
	// west cache must not satisfy the new region.
	c.SetRegion(Region{ID: 2, Key: "east", BaseURL: srv.URL})
	if _, complete := c.Lookup([]string{"T4_BAG"}); complete {
		t.Fatal("east region should not be satisfied by west cache")
	}
}

func TestChunk(t *testing.T) {
	got := chunk([]string{"a", "b", "c", "d", "e"}, 2)
	if len(got) != 3 || len(got[2]) != 1 {
		t.Fatalf("chunk = %v", got)
	}
}

// codeDoer returns a fixed status code with optional headers, recording how many
// requests it saw, to drive the 429 backoff tests deterministically.
type codeDoer struct {
	code    int
	headers map[string]string
	mu      sync.Mutex
	calls   int
}

func (d *codeDoer) Do(*http.Request) (*http.Response, error) {
	d.mu.Lock()
	d.calls++
	d.mu.Unlock()
	h := http.Header{}
	for k, v := range d.headers {
		h.Set(k, v)
	}
	return &http.Response{
		StatusCode: d.code,
		Status:     http.StatusText(d.code),
		Header:     h,
		Body:       io.NopCloser(strings.NewReader("")),
	}, nil
}

func (d *codeDoer) count() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.calls
}

func TestRateLimitBackoff(t *testing.T) {
	now := time.Unix(1000, 0)
	doer := &codeDoer{code: http.StatusTooManyRequests, headers: map[string]string{"Retry-After": "120"}}
	c := New(Options{HTTP: doer, Region: West, Now: func() time.Time { return now }})

	// A refresh hits 429 and records the backoff from Retry-After.
	if err := c.Refresh(context.Background(), []string{"T4_BAG"}); err == nil {
		t.Fatal("expected a rate-limit error")
	}
	st := c.Status()
	if !st.RateLimited || st.RetryAfterSec != 120 {
		t.Fatalf("status = %+v, want rateLimited with ~120s", st)
	}

	// While limited, Lookup must NOT schedule another fetch (no new request) and
	// returns no prices for an uncached item.
	before := doer.count()
	if prices, _ := c.Lookup([]string{"T5_SOUL"}); len(prices) != 0 {
		t.Fatalf("rate-limited cold lookup returned prices: %v", prices)
	}
	time.Sleep(50 * time.Millisecond) // give any (wrongly) scheduled goroutine a chance
	if doer.count() != before {
		t.Fatalf("Lookup fired a fetch while rate-limited: %d -> %d", before, doer.count())
	}

	// After the backoff window the client is free again.
	now = now.Add(121 * time.Second)
	if c.Status().RateLimited {
		t.Fatal("backoff should have lifted")
	}
}

func TestRateLimitServesStaleCache(t *testing.T) {
	now := time.Unix(1000, 0)
	ok := testServer(t, nil)
	c := New(Options{HTTP: ok.Client(), Region: regionAt(ok.URL), TTL: time.Minute, Now: func() time.Time { return now }})
	if err := c.Refresh(context.Background(), []string{"T4_BAG"}); err != nil {
		t.Fatal(err)
	}
	// Expire the entry, then enter a rate-limited state.
	now = now.Add(2 * time.Minute)
	c.setBackoff(now.Add(time.Minute))
	// Stale price is still served (better than nothing) while limited.
	prices, _ := c.Lookup([]string{"T4_BAG"})
	if prices["T4_BAG"] != 1200 {
		t.Fatalf("stale lookup while limited = %v, want cached 1200", prices)
	}
}

func TestParseRetryAfter(t *testing.T) {
	now := time.Unix(1000, 0)
	if got := parseRetryAfter("30", now); !got.Equal(now.Add(30*time.Second)) {
		t.Errorf("seconds form = %v", got)
	}
	if got := parseRetryAfter("", now); !got.IsZero() {
		t.Errorf("empty = %v, want zero", got)
	}
	if got := parseRetryAfter("garbage", now); !got.IsZero() {
		t.Errorf("garbage = %v, want zero", got)
	}
	httpDate := "Wed, 21 Oct 2026 07:28:00 GMT"
	if got := parseRetryAfter(httpDate, now); got.IsZero() {
		t.Errorf("http-date form should parse, got zero")
	}
}

func TestDiskCacheRoundTrip(t *testing.T) {
	dir := t.TempDir()
	now := time.Unix(1_000_000, 0)
	srv := testServer(t, nil)

	// First client fetches and persists.
	c1 := New(Options{HTTP: srv.Client(), Region: regionAt(srv.URL), CacheDir: dir, Now: func() time.Time { return now }})
	if err := c1.Refresh(context.Background(), []string{"T4_BAG"}); err != nil {
		t.Fatal(err)
	}

	// Second client (simulated restart) with an offline doer must serve the
	// persisted price without any network call.
	c2 := New(Options{HTTP: errDoer{}, Region: regionAt(srv.URL), CacheDir: dir, Now: func() time.Time { return now }})
	prices, complete := c2.Lookup([]string{"T4_BAG"})
	if !complete || prices["T4_BAG"] != 1200 {
		t.Fatalf("restart lookup = %v,%v, want cached 1200 complete", prices, complete)
	}
}

func TestDiskCacheDropsExpired(t *testing.T) {
	dir := t.TempDir()
	base := time.Unix(1_000_000, 0)
	srv := testServer(t, nil)
	c1 := New(Options{HTTP: srv.Client(), Region: regionAt(srv.URL), CacheDir: dir, TTL: time.Minute, Now: func() time.Time { return base }})
	if err := c1.Refresh(context.Background(), []string{"T4_BAG"}); err != nil {
		t.Fatal(err)
	}
	// Restart well after the TTL: the persisted entry is too old to load.
	later := base.Add(time.Hour)
	c2 := New(Options{HTTP: errDoer{}, Region: regionAt(srv.URL), CacheDir: dir, TTL: time.Minute, Now: func() time.Time { return later }})
	if _, complete := c2.Lookup([]string{"T4_BAG"}); complete {
		t.Fatal("expired persisted entry should not be loaded")
	}
}

func TestSendsUserAgent(t *testing.T) {
	var gotUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		_, _ = w.Write([]byte(pricesJSON))
	}))
	defer srv.Close()
	c := New(Options{HTTP: srv.Client(), Region: regionAt(srv.URL)})
	if err := c.Refresh(context.Background(), []string{"T4_BAG"}); err != nil {
		t.Fatal(err)
	}
	if gotUA == "" || gotUA == "Go-http-client/1.1" {
		t.Fatalf("User-Agent = %q, want a descriptive app UA", gotUA)
	}
}
