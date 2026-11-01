// Package timeline keeps a single chronological "combat log" for a session:
// loot drops, silver and fame gains, zone changes and party joins/leaves merged
// into one ordered feed. It is UI-agnostic — the engine feeds it typed events
// from the same handlers that drive loot/combat/fame, and the UI reads a
// newest-first slice. The ring buffer is bounded so a long session stays cheap.
package timeline

import (
	"sync"
	"time"
)

// Kind is the type of a timeline entry. Values are stable (JSON) so the UI can
// map each to an icon/colour and offer a type filter.
type Kind string

const (
	KindLoot        Kind = "loot"         // an item was picked up
	KindSilver      Kind = "silver"       // silver was picked up
	KindFame        Kind = "fame"         // fame gained (local player)
	KindZone        Kind = "zone"         // entered a new zone (Join)
	KindPartyJoined Kind = "party_joined" // a player joined the party
	KindPartyLeft   Kind = "party_left"   // a player left the party
)

// Event is one entry in the feed. It is a flat, JSON-friendly projection of the
// already-decoded domain structs (loot.Entry, events.FameUpdate, …) — the
// timeline never decodes packets itself. Unused fields stay zero and are
// omitted from JSON so each Kind carries only what it needs.
type Event struct {
	At       time.Time `json:"at"`
	Kind     Kind      `json:"kind"`
	Player   string    `json:"player,omitempty"`   // looter / fame subject / party member
	Item     string    `json:"item,omitempty"`     // display name (localized once TASK-18 lands)
	ItemID   string    `json:"itemId,omitempty"`   // unique name, for the item icon
	Quantity int64     `json:"quantity,omitempty"` // loot stack size
	Silver   float64   `json:"silver,omitempty"`   // floored silver (matches loot views)
	Fame     float64   `json:"fame,omitempty"`     // fame gained in this event
	Zone     string    `json:"zone,omitempty"`     // destination zone / location
	Source   string    `json:"source,omitempty"`   // loot source (body/chest/mob)
}

// Ring is a bounded, ordered buffer of events. Newest events are appended; when
// full the oldest is overwritten. Safe for concurrent use: the engine goroutine
// Adds while the UI goroutine reads.
type Ring struct {
	mu   sync.RWMutex
	buf  []Event
	next int // index of the next write (and, once full, the oldest entry)
	size int // number of entries currently held (<= cap)
}

// DefaultCap is the number of events retained when NewRing is called with n<=0.
const DefaultCap = 2000

// NewRing returns a ring that retains the most recent n events (DefaultCap if
// n <= 0).
func NewRing(n int) *Ring {
	if n <= 0 {
		n = DefaultCap
	}
	return &Ring{buf: make([]Event, n)}
}

// Add appends one event as the newest, evicting the oldest when full.
func (r *Ring) Add(e Event) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.buf[r.next] = e
	r.next = (r.next + 1) % len(r.buf)
	if r.size < len(r.buf) {
		r.size++
	}
}

// Len reports how many events are currently held.
func (r *Ring) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.size
}

// Recent returns up to n events, newest first. n <= 0 or n > Len returns all
// held events. The result is a copy, safe to hand to the UI.
func (r *Ring) Recent(n int) []Event {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if n <= 0 || n > r.size {
		n = r.size
	}
	out := make([]Event, 0, n)
	// Walk backwards from the most recently written slot.
	for i := 0; i < n; i++ {
		idx := (r.next - 1 - i + len(r.buf)) % len(r.buf)
		out = append(out, r.buf[idx])
	}
	return out
}

// Reset clears every held event, keeping the capacity.
func (r *Ring) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.buf {
		r.buf[i] = Event{}
	}
	r.next = 0
	r.size = 0
}
