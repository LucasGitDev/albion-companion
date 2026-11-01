package prices

import (
	"context"
	"errors"
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
