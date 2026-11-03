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

// maxTimeline bounds the retained harvest history. A long session can produce
// thousands of harvests; the UI only shows a recent feed, so older events are
// dropped oldest-first. Totals (byItem/total/count) are unaffected.
const maxTimeline = 1000

// event is one recorded harvest, kept to build the timeline feed.
type event struct {
	at     time.Time
	index  int
	amount int64
}

// Tracker is the running gathered-resource total. Safe for concurrent use.
type Tracker struct {
	mu       sync.RWMutex
	byItem   map[int]int64 // items.txt index → total amount
	total    int64
	count    int
	first    time.Time
	last     time.Time
	timeline []event // chronological, bounded to maxTimeline (oldest dropped)
	items    ItemResolver
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
	t.timeline = append(t.timeline, event{at: at, index: index, amount: amount})
	if len(t.timeline) > maxTimeline {
		// Drop the oldest; copy to a fresh slice so the backing array does not
		// grow without bound across a long session.
		t.timeline = append([]event(nil), t.timeline[len(t.timeline)-maxTimeline:]...)
	}
}

// Item is one gathered resource line. UniqueName/Name are filled when the item
// resolver knows the index (empty otherwise; the UI falls back to the index).
type Item struct {
	ItemIndex  int    `json:"itemIndex"`
	Amount     int64  `json:"amount"`
	UniqueName string `json:"uniqueName"`
	Name       string `json:"name"`
}

// Event is one harvest in the timeline feed (newest first in a Snapshot).
type Event struct {
	At         time.Time `json:"at"`
	ItemIndex  int       `json:"itemIndex"`
	Amount     int64     `json:"amount"`
	UniqueName string    `json:"uniqueName"`
	Name       string    `json:"name"`
}

// Snapshot is an immutable view, items sorted by amount desc and a newest-first
// timeline of individual harvests.
type Snapshot struct {
	Items    []Item    `json:"items"`
	Total    int64     `json:"total"`
	Count    int       `json:"count"`
	PerHour  float64   `json:"perHour"`
	Since    time.Time `json:"since"`
	Timeline []Event   `json:"timeline"`
}

func (t *Tracker) Snapshot() Snapshot {
	t.mu.RLock()
	defer t.mu.RUnlock()
	// resolve fills localized/unique names for an item index (empty when the
	// resolver has not loaded the item yet).
	resolve := func(idx int) (unique, name string) {
		if t.items == nil {
			return "", ""
		}
		res, ok := t.items.Resolve(idx)
		if !ok {
			return "", ""
		}
		name = res.Name
		if res.LocalName != "" {
			name = res.LocalName
		}
		return res.UniqueName, name
	}

	items := make([]Item, 0, len(t.byItem))
	for idx, amt := range t.byItem {
		it := Item{ItemIndex: idx, Amount: amt}
		it.UniqueName, it.Name = resolve(idx)
		items = append(items, it)
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Amount != items[j].Amount {
			return items[i].Amount > items[j].Amount
		}
		return items[i].ItemIndex < items[j].ItemIndex
	})

	timeline := make([]Event, 0, len(t.timeline))
	for i := len(t.timeline) - 1; i >= 0; i-- { // newest first
		e := t.timeline[i]
		ev := Event{At: e.at, ItemIndex: e.index, Amount: e.amount}
		ev.UniqueName, ev.Name = resolve(e.index)
		timeline = append(timeline, ev)
	}

	return Snapshot{
		Items:    items,
		Total:    t.total,
		Count:    t.count,
		PerHour:  perHour(t.total, t.first, t.last),
		Since:    t.first,
		Timeline: timeline,
	}
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
	t.timeline = nil
}
