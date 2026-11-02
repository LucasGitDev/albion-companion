// Package rates accumulates the progression/economy currencies that Albion's
// dashboard reports as both session totals and per-hour rates: silver, respec
// points and might. It mirrors the internal/fame Tracker pattern (total +
// actual rate over the elapsed window) so the UI can show one consistent
// "rates" view. Fame itself stays in internal/fame; the app service combines
// the two. Ported from SA's LiveStatsTracker (ValueType.Silver/ReSpec/Might).
package rates

import (
	"sync"
	"time"
)

// metric is one running currency total with its own elapsed window.
type metric struct {
	total float64
	count int
	first time.Time
	last  time.Time
}

func (m *metric) add(v float64, at time.Time) {
	m.total += v
	m.count++
	if m.first.IsZero() {
		m.first = at
	}
	m.last = at
}

func (m *metric) snapshot() Metric {
	return Metric{
		Total:   m.total,
		PerHour: perHour(m.total, m.first, m.last),
		Count:   m.count,
		Since:   m.first,
	}
}

// Tracker is the running silver/respec/might totals for a session. Safe for
// concurrent use.
type Tracker struct {
	mu     sync.RWMutex
	silver metric
	respec metric
	might  metric
	favor  metric
}

func New() *Tracker { return &Tracker{} }

// AddSilver records silver looted by the local player (already floored per
// pickup by loot.GrabbedLoot.SilverAmount, so totals match the loot views).
func (t *Tracker) AddSilver(v float64, at time.Time) {
	t.mu.Lock()
	t.silver.add(v, at)
	t.mu.Unlock()
}

// AddRespec records respec points gained (UpdateReSpecPoints, local-only).
func (t *Tracker) AddRespec(v float64, at time.Time) {
	t.mu.Lock()
	t.respec.add(v, at)
	t.mu.Unlock()
}

// AddMight records might gained (MightAndFavorReceived, local-only).
func (t *Tracker) AddMight(v float64, at time.Time) {
	t.mu.Lock()
	t.might.add(v, at)
	t.mu.Unlock()
}

// AddFavor records favor gained (MightAndFavorReceived, local-only).
func (t *Tracker) AddFavor(v float64, at time.Time) {
	t.mu.Lock()
	t.favor.add(v, at)
	t.mu.Unlock()
}

// Metric is an immutable view of one currency: session total and actual rate.
type Metric struct {
	Total   float64   `json:"total"`
	PerHour float64   `json:"perHour"`
	Count   int       `json:"count"`
	Since   time.Time `json:"since"`
}

// Snapshot is an immutable view of all tracked rates.
type Snapshot struct {
	Silver Metric `json:"silver"`
	Respec Metric `json:"respec"`
	Might  Metric `json:"might"`
	Favor  Metric `json:"favor"`
}

func (t *Tracker) Snapshot() Snapshot {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return Snapshot{
		Silver: t.silver.snapshot(),
		Respec: t.respec.snapshot(),
		Might:  t.might.snapshot(),
		Favor:  t.favor.snapshot(),
	}
}

func (t *Tracker) Reset() {
	t.mu.Lock()
	t.silver, t.respec, t.might, t.favor = metric{}, metric{}, metric{}, metric{}
	t.mu.Unlock()
}

// perHour is the session rate: total ÷ elapsed time, expressed per hour. Early
// in a session the short window makes it a projection; it converges on the
// actual rate as the session grows. Matches internal/fame.perHour.
func perHour(total float64, first, last time.Time) float64 {
	if first.IsZero() || !last.After(first) {
		return 0
	}
	hours := last.Sub(first).Hours()
	if hours <= 0 {
		return 0
	}
	return total / hours
}
