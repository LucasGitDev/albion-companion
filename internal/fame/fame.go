// Package fame accumulates the fame gained during a session. Fame is Albion's
// core progression currency, so the tool tracks total fame, the premium
// ("yellow") bonus, and fame per hour. Ported from SA's UpdateFameEvent model.
package fame

import (
	"sync"
	"time"

	"github.com/LucasGitDev/albion-companion/internal/events"
)

// Tracker is the running fame total for a session. Safe for concurrent use.
type Tracker struct {
	mu      sync.RWMutex
	total   float64
	base    float64
	premium float64
	satchel float64
	count   int
	latest  float64 // last seen lifetime TotalPlayerFame
	first   time.Time
	last    time.Time
}

func New() *Tracker { return &Tracker{} }

// Add records one fame-gain event.
func (t *Tracker) Add(f events.FameUpdate, at time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.total += f.Gained()
	t.base += f.Base * f.BonusFactor
	t.premium += f.Premium() * f.BonusFactor
	t.satchel += f.Satchel * f.BonusFactor
	t.count++
	if f.TotalPlayerFame > 0 {
		t.latest = f.TotalPlayerFame
	}
	if t.first.IsZero() {
		t.first = at
	}
	t.last = at
}

// Snapshot is an immutable view of the current fame totals.
type Snapshot struct {
	Total           float64   `json:"total"`           // fame gained this session
	Base            float64   `json:"base"`            // base fame (no premium)
	Premium         float64   `json:"premium"`         // premium account bonus fame (+50%), not Quick-Learner yellow fame
	Satchel         float64   `json:"satchel"`         // satchel / bag insight fame
	PerHour         float64   `json:"perHour"`         // actual rate: total ÷ elapsed
	Count           int       `json:"count"`           // number of fame events
	LatestTotalFame float64   `json:"latestTotalFame"` // last seen lifetime player fame
	Since           time.Time `json:"since"`
}

func (t *Tracker) Snapshot() Snapshot {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return Snapshot{
		Total:           t.total,
		Base:            t.base,
		Premium:         t.premium,
		Satchel:         t.satchel,
		PerHour:         perHour(t.total, t.first, t.last),
		Count:           t.count,
		LatestTotalFame: t.latest,
		Since:           t.first,
	}
}

// perHour is the session fame rate: total fame gained ÷ elapsed time, expressed
// per hour. Early in a session this short window makes it a projection of the
// hourly rate; as the session grows it converges on the actual rate.
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

func (t *Tracker) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	*t = Tracker{}
}
