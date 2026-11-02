// Package chestlog is the local player's personal container-activity log: items
// the local client took out of a chest / loot container, in chronological order.
//
// Scope and limits (see decision-7): chest put/take events are CLIENT-LOCAL, so
// this only ever reflects the local player — never other guild/party members.
// The deposit direction (inventory → chest) is not implemented: the only signal
// is the slot-based InventoryMoveItem request, whose item/quantity fields are
// not documented by StatisticsAnalysis, so decoding it would mean guessing
// packet fields. Withdrawals are what the game lets us observe passively.
package chestlog

import (
	"sync"
	"time"
)

// Kind is the direction of a container operation. Only withdrawals are observable
// today; Deposit is reserved for when a validated capture makes it decodable.
type Kind string

const (
	KindWithdraw Kind = "withdraw"
	KindDeposit  Kind = "deposit"
)

// Entry is one observed container operation by the local player.
type Entry struct {
	At        time.Time `json:"at"`
	Kind      Kind      `json:"kind"`
	Container string    `json:"container"` // chest unique name or loot-body name
	ItemID    string    `json:"itemId"`    // item unique name (e.g. T5_SOUL)
	ItemName  string    `json:"itemName"`
	Quantity  int64     `json:"quantity"`
}

// Tracker keeps the activity log for the current session. Safe for concurrent
// use, though the engine drives it from the single decode goroutine.
type Tracker struct {
	mu      sync.RWMutex
	entries []Entry
	max     int
}

// New returns a tracker keeping up to max recent entries (<=0 means 2000).
func New(max int) *Tracker {
	if max <= 0 {
		max = 2000
	}
	return &Tracker{max: max}
}

// Add appends one operation, trimming the oldest past the cap.
func (t *Tracker) Add(e Entry) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.entries = append(t.entries, e)
	if len(t.entries) > t.max {
		t.entries = t.entries[len(t.entries)-t.max:]
	}
}

// Snapshot is an immutable view, newest first, capped at recent (<=0 = all).
type Snapshot struct {
	Entries    []Entry `json:"entries"`
	Total      int     `json:"total"`      // withdraw + deposit
	Withdrawn  int64   `json:"withdrawn"`  // total quantity withdrawn
	Deposited  int64   `json:"deposited"`  // total quantity deposited (0 until supported)
	Containers int     `json:"containers"` // distinct container names seen
}

func (t *Tracker) Snapshot(recent int) Snapshot {
	t.mu.RLock()
	defer t.mu.RUnlock()
	var s Snapshot
	s.Total = len(t.entries)
	seen := map[string]struct{}{}
	for _, e := range t.entries {
		switch e.Kind {
		case KindDeposit:
			s.Deposited += e.Quantity
		default:
			s.Withdrawn += e.Quantity
		}
		if e.Container != "" {
			seen[e.Container] = struct{}{}
		}
	}
	s.Containers = len(seen)

	n := len(t.entries)
	if recent > 0 && recent < n {
		n = recent
	}
	out := make([]Entry, 0, n)
	for i := len(t.entries) - 1; i >= len(t.entries)-n; i-- {
		out = append(out, t.entries[i])
	}
	s.Entries = out
	return s
}

func (t *Tracker) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.entries = nil
}
