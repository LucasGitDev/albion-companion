// Package prices estimates the market value of looted items using the Albion
// Online Data project's public price API (https://www.albion-online-data.com).
//
// This is the ONE feature that relies on an EXTERNAL network API rather than
// passive game packets. It is kept strictly separate and optional: nothing here
// ever sends game packets, and every failure degrades gracefully — a missing
// price simply makes a value unknown; the engine is never blocked or affected.
//
// The project runs region-specific markets (Americas/west, Asia/east,
// Europe/eu), selected by subdomain. The capture source IP of the game server
// identifies the region (see RegionFromIP, ported from the albiondata-client
// lineage's albionState.GetServer).
package prices

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Region is one Albion Online Data market server.
type Region struct {
	ID      int    `json:"id"`      // AOData server id: 1 west, 2 east, 3 europe
	Key     string `json:"key"`     // url/setting key: "west", "east", "europe"
	Label   string `json:"label"`   // human label: "Americas", "Asia", "Europe"
	BaseURL string `json:"baseUrl"` // https://west.albion-online-data.com
}

// The three live AO servers. BaseURLs use the region subdomains the Albion
// Online Data project exposes for region-specific market data.
var (
	West   = Region{ID: 1, Key: "west", Label: "Americas", BaseURL: "https://west.albion-online-data.com"}
	East   = Region{ID: 2, Key: "east", Label: "Asia", BaseURL: "https://east.albion-online-data.com"}
	Europe = Region{ID: 3, Key: "europe", Label: "Europe", BaseURL: "https://europe.albion-online-data.com"}
)

// Regions lists all selectable regions (order = Settings dropdown order).
var Regions = []Region{West, East, Europe}

// DefaultRegion is used until a region is detected or chosen (task default:
// Americas, as the gap note in the task calls for a configurable default).
var DefaultRegion = West

// RegionFromIP maps a game-server source IP to its market region. The class-C
// ranges are the known AO server ranges, ported from the albiondata-client
// lineage (client/albion_state.go GetServer). Returns false for any other IP
// (local traffic, other services), so only real game-server packets match.
func RegionFromIP(ip string) (Region, bool) {
	switch {
	case strings.HasPrefix(ip, "5.188.125."): // west server class-C range
		return West, true
	case strings.HasPrefix(ip, "5.45.187."): // east server class-C range
		return East, true
	case strings.HasPrefix(ip, "193.169.238."): // eu server class-C range
		return Europe, true
	}
	return Region{}, false
}

// RegionByKey resolves a region from its key ("west"/"east"/"europe").
func RegionByKey(key string) (Region, bool) {
	for _, r := range Regions {
		if r.Key == key {
			return r, true
		}
	}
	return Region{}, false
}

// HTTPDoer is the subset of *http.Client the client needs; injected in tests.
type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

// maxBatch caps items per request to keep the URL short and the AO Data API
// happy. The API itself accepts large batches, but we stay polite.
const maxBatch = 100

type entry struct {
	price int64     // estimated unit price in silver (0 = no market data)
	at    time.Time // when it was fetched
}

// Client caches item prices per region in memory with a TTL and fails soft on
// any network error. It is safe for concurrent use.
//
// Lookup never blocks on the network: it reads the cache and schedules missing
// items for a background fetch. Refresh fetches synchronously (manual button).
type Client struct {
	http HTTPDoer
	ttl  time.Duration
	now  func() time.Time

	mu       sync.Mutex
	region   Region
	cache    map[string]entry // key: region.Key + "|" + itemID
	inflight map[string]bool  // items currently being fetched (per region key)
	onUpdate func()           // optional: called after a background fetch stores results
}

// Options configure a Client.
type Options struct {
	HTTP     HTTPDoer      // default: &http.Client{Timeout: 15s}
	TTL      time.Duration // cache lifetime (default 30m)
	Region   Region        // initial region (default DefaultRegion)
	OnUpdate func()        // optional: notified after background fetches land
	Now      func() time.Time
}

// New builds a price client. All options are optional.
func New(o Options) *Client {
	if o.HTTP == nil {
		o.HTTP = &http.Client{Timeout: 15 * time.Second}
	}
	if o.TTL == 0 {
		o.TTL = 30 * time.Minute
	}
	if o.Region.Key == "" {
		o.Region = DefaultRegion
	}
	if o.Now == nil {
		o.Now = time.Now
	}
	return &Client{
		http:     o.HTTP,
		ttl:      o.TTL,
		now:      o.Now,
		region:   o.Region,
		cache:    map[string]entry{},
		inflight: map[string]bool{},
		onUpdate: o.OnUpdate,
	}
}

// Region returns the current market region.
func (c *Client) Region() Region {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.region
}

// SetRegion switches the market region. Prices are cached per region, so a
// switch simply starts resolving against the new region's cache.
func (c *Client) SetRegion(r Region) {
	if r.Key == "" {
		return
	}
	c.mu.Lock()
	c.region = r
	c.mu.Unlock()
}

func cacheKey(regionKey, itemID string) string { return regionKey + "|" + itemID }

// Lookup returns cached unit prices (silver) for the given item ids in the
// current region. Items without a fresh cached price are omitted and scheduled
// for a background fetch; complete is true only when every requested id had a
// fresh price. It never performs network IO on the calling goroutine.
func (c *Client) Lookup(items []string) (prices map[string]int64, complete bool) {
	prices = map[string]int64{}
	var missing []string

	c.mu.Lock()
	region := c.region
	cutoff := c.now().Add(-c.ttl)
	seen := map[string]bool{}
	for _, id := range items {
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		e, ok := c.cache[cacheKey(region.Key, id)]
		if ok && e.at.After(cutoff) {
			if e.price > 0 {
				prices[id] = e.price
			}
			continue
		}
		if !c.inflight[cacheKey(region.Key, id)] {
			missing = append(missing, id)
		}
	}
	for _, id := range missing {
		c.inflight[cacheKey(region.Key, id)] = true
	}
	c.mu.Unlock()

	if len(missing) > 0 {
		go c.fetchInto(context.Background(), region, missing, true)
	}
	return prices, len(missing) == 0
}

// Refresh fetches prices for the given items synchronously in the current
// region, bypassing cache freshness, and stores the results. It fails soft:
// a network error is returned for logging but the cache simply stays as-is.
func (c *Client) Refresh(ctx context.Context, items []string) error {
	c.mu.Lock()
	region := c.region
	c.mu.Unlock()
	return c.fetchInto(ctx, region, dedup(items), false)
}

// fetchInto batches items, fetches each batch and stores the estimates. When
// clearInflight is set it clears the inflight guard for the items afterwards
// (used by the background path). Returns the first error encountered.
func (c *Client) fetchInto(ctx context.Context, region Region, items []string, clearInflight bool) error {
	if clearInflight {
		defer func() {
			c.mu.Lock()
			for _, id := range items {
				delete(c.inflight, cacheKey(region.Key, id))
			}
			c.mu.Unlock()
		}()
	}
	var firstErr error
	updated := false
	for _, batch := range chunk(items, maxBatch) {
		got, err := c.fetchBatch(ctx, region, batch)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue // fail soft: skip this batch, keep going
		}
		now := c.now()
		c.mu.Lock()
		for _, id := range batch {
			// Store even a zero price (no market data) so we don't refetch it
			// every pass until the TTL expires.
			c.cache[cacheKey(region.Key, id)] = entry{price: got[id], at: now}
		}
		c.mu.Unlock()
		updated = true
	}
	if updated && c.onUpdate != nil {
		c.onUpdate()
	}
	return firstErr
}

// priceRow is one row of the AO Data /stats/prices response.
type priceRow struct {
	ItemID       string `json:"item_id"`
	City         string `json:"city"`
	Quality      int    `json:"quality"`
	SellPriceMin int64  `json:"sell_price_min"`
	BuyPriceMax  int64  `json:"buy_price_max"`
}

// fetchBatch requests one batch of item prices and reduces them to a single
// estimated unit price per item id: the highest sell_price_min across cities
// and qualities (what the item realistically sells for), falling back to the
// highest buy_price_max when no sell order exists.
func (c *Client) fetchBatch(ctx context.Context, region Region, items []string) (map[string]int64, error) {
	if len(items) == 0 {
		return map[string]int64{}, nil
	}
	// Item ids may contain '@' (enchant), which is URL-safe in a path segment,
	// but we escape each id defensively to be safe against future id shapes.
	escaped := make([]string, len(items))
	for i, id := range items {
		escaped[i] = url.PathEscape(id)
	}
	endpoint := region.BaseURL + "/api/v2/stats/prices/" + strings.Join(escaped, ",") + ".json"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		// Drain a little so the connection can be reused.
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<16))
		return nil, fmt.Errorf("prices: GET %s: %s", region.Key, resp.Status)
	}
	var rows []priceRow
	if err := json.NewDecoder(resp.Body).Decode(&rows); err != nil {
		return nil, fmt.Errorf("prices: decode %s: %w", region.Key, err)
	}
	sell := map[string]int64{}
	buy := map[string]int64{}
	for _, r := range rows {
		if r.SellPriceMin > sell[r.ItemID] {
			sell[r.ItemID] = r.SellPriceMin
		}
		if r.BuyPriceMax > buy[r.ItemID] {
			buy[r.ItemID] = r.BuyPriceMax
		}
	}
	out := make(map[string]int64, len(items))
	for _, id := range items {
		if sell[id] > 0 {
			out[id] = sell[id]
		} else {
			out[id] = buy[id] // 0 when there's no market data at all
		}
	}
	return out, nil
}

func dedup(in []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(in))
	for _, s := range in {
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}

func chunk(in []string, size int) [][]string {
	if size <= 0 {
		size = len(in)
	}
	var out [][]string
	for i := 0; i < len(in); i += size {
		end := i + size
		if end > len(in) {
			end = len(in)
		}
		out = append(out, in[i:end])
	}
	return out
}
