package app

import (
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
