// Package gather accumulates resources harvested during a session (wood, ore,
// fiber, hide, stone, fish). Fed from HarvestFinished for the local player.
package gather

import (
	"sort"
	"sync"
	"time"

	"github.com/LucasGitDev/albion-companion/internal/itemdb"
)

// ItemResolver maps an items.txt index to its names (same shape as loot's).
type ItemResolver interface {
	Resolve(index int) (itemdb.Item, bool)
}

// Tracker is the running gathered-resource total. Safe for concurrent use.
type Tracker struct {
	mu     sync.RWMutex
	byItem map[int]int64 // items.txt index → total amount
	total  int64
	count  int
	first  time.Time
	last   time.Time
	items  ItemResolver
}

func New() *Tracker { return &Tracker{byItem: map[int]int64{}} }

// SetItems swaps the item resolver (itemdb loads asynchronously).
func (t *Tracker) SetItems(items ItemResolver) {
	t.mu.Lock()
	t.items = items
	t.mu.Unlock()
}

// Add records one harvest of `amount` units of item `index`.
func (t *Tracker) Add(index int, amount int64, at time.Time) {
	if amount <= 0 {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.byItem[index] += amount
	t.total += amount
	t.count++
	if t.first.IsZero() {
		t.first = at
	}
	t.last = at
}

// Item is one gathered resource line. UniqueName/Name are filled when the item
// resolver knows the index (empty otherwise; the UI falls back to the index).
type Item struct {
	ItemIndex  int    `json:"itemIndex"`
	Amount     int64  `json:"amount"`
	UniqueName string `json:"uniqueName"`
	Name       string `json:"name"`
}

// Snapshot is an immutable view, items sorted by amount desc.
type Snapshot struct {
	Items   []Item    `json:"items"`
	Total   int64     `json:"total"`
	Count   int       `json:"count"`
	PerHour float64   `json:"perHour"`
	Since   time.Time `json:"since"`
}

func (t *Tracker) Snapshot() Snapshot {
	t.mu.RLock()
	defer t.mu.RUnlock()
	items := make([]Item, 0, len(t.byItem))
	for idx, amt := range t.byItem {
		it := Item{ItemIndex: idx, Amount: amt}
		if t.items != nil {
			if res, ok := t.items.Resolve(idx); ok {
				it.UniqueName = res.UniqueName
				if res.LocalName != "" {
					it.Name = res.LocalName
				} else {
					it.Name = res.Name
				}
			}
		}
		items = append(items, it)
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Amount != items[j].Amount {
			return items[i].Amount > items[j].Amount
		}
		return items[i].ItemIndex < items[j].ItemIndex
	})
	return Snapshot{Items: items, Total: t.total, Count: t.count, PerHour: perHour(t.total, t.first, t.last), Since: t.first}
}

func perHour(total int64, first, last time.Time) float64 {
	if first.IsZero() || !last.After(first) {
		return 0
	}
	h := last.Sub(first).Hours()
	if h <= 0 {
		return 0
	}
	return float64(total) / h
}

func (t *Tracker) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.byItem = map[int]int64{}
	t.total = 0
	t.count = 0
	t.first = time.Time{}
	t.last = time.Time{}
}
