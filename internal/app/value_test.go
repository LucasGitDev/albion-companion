package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/LucasGitDev/albion-companion/internal/events"
	"github.com/LucasGitDev/albion-companion/internal/game"
	"github.com/LucasGitDev/albion-companion/internal/itemdb"
	"github.com/LucasGitDev/albion-companion/internal/prices"
)

type fakeItems map[int]itemdb.Item

func (f fakeItems) Resolve(i int) (itemdb.Item, bool) { it, ok := f[i]; return it, ok }

const valueJSON = `[
	{"item_id":"T4_BAG","city":"Martlock","quality":1,"sell_price_min":1000,"buy_price_max":400}
]`

func TestLootView_EstimatedValue(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(valueJSON))
	}))
	defer srv.Close()

	pc := prices.New(prices.Options{
		HTTP:   srv.Client(),
		Region: prices.Region{ID: 1, Key: "west", Label: "Americas", BaseURL: srv.URL},
	})
	svc := NewService(Options{Prices: pc, Items: fakeItems{10: {Index: 10, UniqueName: "T4_BAG", Name: "Bag"}}})

	tr := svc.Engine().Loot()
	st := game.NewState()
	at := time.Now()
	tr.Record(events.GrabbedLoot{Looter: "Me", ItemIndex: 10, Quantity: 3}, at, st)
	tr.Record(events.GrabbedLoot{Looter: "Rat", ItemIndex: 10, Quantity: 1}, at, st)

	// Before any price fetch, values are unavailable (offline-style); engine fine.
	if cold := svc.LootView(10).Value; cold.Available {
		t.Fatal("expected no value before prices are fetched")
	}

	if err := svc.RefreshPrices(); err != nil {
		t.Fatalf("RefreshPrices: %v", err)
	}
	v := svc.LootView(10).Value
	if !v.Available || v.PerItem["T4_BAG"] != 1000 {
		t.Fatalf("value = %+v", v)
	}
	if v.PerPlayer["Me"] != 3000 || v.PerPlayer["Rat"] != 1000 {
		t.Fatalf("perPlayer = %+v", v.PerPlayer)
	}
	if v.Total != 4000 || v.Region != "west" {
		t.Fatalf("value = %+v", v)
	}
}

func TestSetPriceRegion_Manual(t *testing.T) {
	svc := NewService(Options{})
	svc.SetPriceRegion("europe")
	if svc.prices.Region().Key != "europe" || !svc.regionManual {
		t.Fatalf("region = %s manual=%v", svc.prices.Region().Key, svc.regionManual)
	}
	// Unknown/empty key re-enables auto-detection.
	svc.SetPriceRegion("")
	if svc.regionManual {
		t.Fatal("empty key should re-enable auto-detect")
	}
}

// gatherPricesJSON prices T4_WOOD at 120 and leaves T4_ROCK with no market data
// (sell/buy both 0), so the gather value must count wood and skip rock.
const gatherPricesJSON = `[
	{"item_id":"T4_WOOD","city":"Martlock","quality":1,"sell_price_min":120,"buy_price_max":50},
	{"item_id":"T4_ROCK","city":"Martlock","quality":1,"sell_price_min":0,"buy_price_max":0}
]`

func TestGatherView_EstimatedValue(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(gatherPricesJSON))
	}))
	defer srv.Close()

	pc := prices.New(prices.Options{
		HTTP:   srv.Client(),
		Region: prices.Region{ID: 1, Key: "west", Label: "Americas", BaseURL: srv.URL},
	})
	svc := NewService(Options{
		Prices: pc,
		Items: fakeItems{
			424: {Index: 424, UniqueName: "T4_WOOD", Name: "Wood"},
			100: {Index: 100, UniqueName: "T4_ROCK", Name: "Rock"},
		},
	})

	g := svc.Engine().Gather()
	base := time.Unix(5000, 0)
	g.Add(424, 10, base)                  // wood x10 @120 = 1200
	g.Add(100, 4, base.Add(time.Minute))  // rock x4, no market
	g.Add(424, 5, base.Add(2*time.Minute)) // wood x5 more → 15 total @120 = 1800

	// Cold: no value yet, but the snapshot (items + timeline) is already there.
	cold := svc.GatherView()
	if cold.Value.Available {
		t.Fatal("expected no value before prices fetched")
	}
	if len(cold.Timeline) != 3 || cold.Total != 19 {
		t.Fatalf("snapshot cold = total %d timeline %d", cold.Total, len(cold.Timeline))
	}

	if err := pc.Refresh(context.Background(), []string{"T4_WOOD", "T4_ROCK"}); err != nil {
		t.Fatalf("refresh: %v", err)
	}

	v := svc.GatherView().Value
	if !v.Available || v.Region != "west" {
		t.Fatalf("value = %+v", v)
	}
	if v.PerItem["T4_WOOD"] != 120 {
		t.Fatalf("wood unit price = %d want 120", v.PerItem["T4_WOOD"])
	}
	// Rock has no market data → not priced, must not appear as a misleading 0.
	if _, ok := v.PerItem["T4_ROCK"]; ok {
		t.Fatalf("rock should have no price: %+v", v.PerItem)
	}
	if v.Missing != 1 {
		t.Fatalf("missing = %d want 1 (rock)", v.Missing)
	}
	if v.Total != 1800 { // only wood counts: 15 * 120
		t.Fatalf("total = %d want 1800", v.Total)
	}
}
