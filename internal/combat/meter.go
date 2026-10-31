// Package combat implements the damage meter (damage, heal, damage taken, DPS)
// following StatisticsAnalysis' CombatController rules.
package combat

import (
	"sort"
	"sync"
	"time"

	"github.com/LucasGitDev/albion-companion/internal/events"
	"github.com/LucasGitDev/albion-companion/internal/game"
)

type Options struct {
	// PartyOnly counts only the local player and party members as causers (SA default).
	PartyOnly bool
	// PlayerTargetsOnly ignores damage dealt to mobs (PvP mode).
	PlayerTargetsOnly bool
}

type PlayerStats struct {
	Name    string    `json:"name"`
	Damage  int64     `json:"damage"`
	Heal    int64     `json:"heal"`
	Taken   int64     `json:"taken"`
	Hits    int       `json:"hits"`
	FirstAt time.Time `json:"firstAt"`
	LastAt  time.Time `json:"lastAt"`
}

// DPS is damage over the player's active combat window (min 1s).
func (p PlayerStats) DPS() float64 { return float64(p.Damage) / p.window() }

// HPS is healing over the same window.
func (p PlayerStats) HPS() float64 { return float64(p.Heal) / p.window() }

func (p PlayerStats) window() float64 {
	d := p.LastAt.Sub(p.FirstAt).Seconds()
	if d < 1 {
		return 1
	}
	return d
}

// Meter is safe for concurrent use.
type Meter struct {
	mu    sync.RWMutex
	opts  Options
	stats map[string]*PlayerStats // by player name
}

func NewMeter(opts Options) *Meter {
	return &Meter{opts: opts, stats: map[string]*PlayerStats{}}
}

func (m *Meter) SetOptions(o Options) {
	m.mu.Lock()
	m.opts = o
	m.mu.Unlock()
}

func (m *Meter) stat(name string) *PlayerStats {
	s := m.stats[name]
	if s == nil {
		s = &PlayerStats{Name: name}
		m.stats[name] = s
	}
	return s
}

// touch widens the player's active window (damage/heal dealt, not taken).
func (s *PlayerStats) touch(at time.Time) {
	if s.FirstAt.IsZero() || at.Before(s.FirstAt) {
		s.FirstAt = at
	}
	if at.After(s.LastAt) {
		s.LastAt = at
	}
}

// Apply records one health change. Rules (SA CombatController.AddDamage):
//   - self-damage is ignored
//   - causer must be a known player (and in party when PartyOnly)
//   - damage taken is tracked for affected party players
//
// TODO: overheal needs max health (SA IsMaxHealthReached) — not tracked yet.
func (m *Meter) Apply(u events.HealthUpdate, at time.Time, st *game.State) {
	amount := u.Amount()
	if amount == 0 {
		return
	}
	causer, causerOK := st.Entity(u.CauserID)
	affected, affectedOK := st.Entity(u.AffectedID)

	m.mu.Lock()
	defer m.mu.Unlock()

	if u.IsDamage() {
		if u.CauserID == u.AffectedID {
			return
		}
		if affectedOK && affected.Kind == game.KindPlayer && st.InParty(affected.Name) {
			m.stat(affected.Name).Taken += amount
		}
		if m.opts.PlayerTargetsOnly && !(affectedOK && affected.Kind == game.KindPlayer) {
			return
		}
	}
	if !causerOK || causer.Kind != game.KindPlayer {
		return
	}
	if m.opts.PartyOnly && !st.InParty(causer.Name) {
		return
	}
	s := m.stat(causer.Name)
	s.touch(at)
	if u.IsDamage() {
		s.Damage += amount
		s.Hits++
	} else {
		s.Heal += amount
	}
}

// Snapshot returns players sorted by damage (desc).
func (m *Meter) Snapshot() []PlayerStats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]PlayerStats, 0, len(m.stats))
	for _, s := range m.stats {
		if s.Damage == 0 && s.Heal == 0 && s.Taken == 0 {
			continue
		}
		out = append(out, *s)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Damage != out[j].Damage {
			return out[i].Damage > out[j].Damage
		}
		return out[i].Name < out[j].Name
	})
	return out
}

func (m *Meter) Reset() {
	m.mu.Lock()
	m.stats = map[string]*PlayerStats{}
	m.mu.Unlock()
}
