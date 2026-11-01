// Package app is the UI-facing service: the exact contract the Wails frontend
// talks to. It has no Wails import — main.go binds *Service and passes an
// Emitter backed by runtime.EventsEmit — so it is fully unit-testable.
//
// Frontend contract:
//   - methods on *Service are Wails-bound (StartCapture, LootView, ...)
//   - push events (throttled): "loot:changed", "combat:changed", "party:changed",
//     "zone:changed", "capture:status", "loot:new" (payload = loot.Entry)
//   - the frontend re-fetches the view on a *:changed event (pull-on-notify),
//     so payloads stay small and the UI never drifts from the engine state.
package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/LucasGitDev/albion-companion/internal/capture"
	"github.com/LucasGitDev/albion-companion/internal/combat"
	"github.com/LucasGitDev/albion-companion/internal/dungeon"
	"github.com/LucasGitDev/albion-companion/internal/engine"
	"github.com/LucasGitDev/albion-companion/internal/events"
	"github.com/LucasGitDev/albion-companion/internal/fame"
	"github.com/LucasGitDev/albion-companion/internal/loot"
	"github.com/LucasGitDev/albion-companion/internal/protocol"
	"github.com/LucasGitDev/albion-companion/internal/rates"
	"github.com/LucasGitDev/albion-companion/internal/timeline"
)

// Emitter pushes an event to the frontend (Wails: runtime.EventsEmit).
type Emitter func(event string, payload any)

// SourceFactory builds capture sources; swapped in tests / dev replay mode.
type SourceFactory func() ([]capture.Source, error)

type Options struct {
	Codes         *protocol.Codes
	Items         loot.ItemResolver
	Sources       SourceFactory // default: live capture on all physical interfaces
	Throttle      time.Duration // min interval between *:changed events (default 250ms)
	Emit          Emitter
	PartyOnlyLoot bool
	Combat        combat.Options
}

type CaptureStatus struct {
	Running       bool         `json:"running"`
	Sources       []string     `json:"sources"`
	Error         string       `json:"error,omitempty"`
	Hint          string       `json:"hint,omitempty"` // e.g. permission instructions
	Stats         engine.Stats `json:"stats"`
	CodesOutdated bool         `json:"codesOutdated"` // too many unknown codes → codes.json likely stale after a patch
}

type Service struct {
	opts Options
	eng  *engine.Engine

	mu        sync.Mutex
	cancel    context.CancelFunc
	status    CaptureStatus
	pending   map[engine.Topic]bool
	lastFlush time.Time
	lootOpts  loot.Filter
}

func NewService(opts Options) *Service {
	if opts.Throttle == 0 {
		opts.Throttle = 250 * time.Millisecond
	}
	if opts.Emit == nil {
		opts.Emit = func(string, any) {}
	}
	if opts.Sources == nil {
		opts.Sources = func() ([]capture.Source, error) { return capture.LiveSources(0) }
	}
	s := &Service{opts: opts, pending: map[engine.Topic]bool{}, lootOpts: loot.Filter{PartyOnly: opts.PartyOnlyLoot}}
	s.eng = engine.New(engine.Config{
		Codes:    opts.Codes,
		Items:    opts.Items,
		Combat:   opts.Combat,
		OnChange: s.onChange,
		OnLoot:   func(e loot.Entry) { s.opts.Emit("loot:new", e) },
	})
	return s
}

// Engine exposes the engine (used by main.go for itemdb hot-swap, tests).
func (s *Service) Engine() *engine.Engine { return s.eng }

// ---- capture lifecycle ----------------------------------------------------

func (s *Service) StartCapture() (CaptureStatus, error) {
	s.mu.Lock()
	if s.cancel != nil {
		st := s.statusLocked()
		s.mu.Unlock()
		return st, nil
	}
	sources, err := s.opts.Sources()
	if err != nil {
		s.status.Error = err.Error()
		st := s.statusLocked()
		s.mu.Unlock()
		return st, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.status = CaptureStatus{Running: true}
	for _, src := range sources {
		s.status.Sources = append(s.status.Sources, src.Name())
	}
	s.mu.Unlock()

	go func() {
		err := s.eng.Run(ctx, sources...)
		s.mu.Lock()
		s.cancel = nil
		s.status.Running = false
		if err != nil {
			s.status.Error = err.Error()
			if errors.Is(err, capture.ErrPermission) {
				s.status.Hint = "permission"
			}
		}
		st := s.statusLocked()
		s.mu.Unlock()
		s.opts.Emit("capture:status", st)
	}()
	st := s.Status()
	s.opts.Emit("capture:status", st)
	return st, nil
}

func (s *Service) StopCapture() CaptureStatus {
	s.mu.Lock()
	if s.cancel != nil {
		s.cancel()
	}
	s.mu.Unlock()
	return s.Status()
}

func (s *Service) Status() CaptureStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.statusLocked()
}

func (s *Service) statusLocked() CaptureStatus {
	st := s.status
	st.Stats = s.eng.Stats()
	st.CodesOutdated = codesOutdated(st.Stats)
	return st
}

// codesOutdated flags a likely patch break: with a matching codes.json the
// unknown-code ratio is ~0, so once a session has warmed up a sustained spike
// means the event/operation codes moved and codes.json needs regenerating.
func codesOutdated(st engine.Stats) bool {
	const warmup = 300
	const thresholdPct = 5
	if st.Messages < warmup {
		return false
	}
	return st.Unknown*100/st.Messages >= thresholdPct
}

// ---- views (pull) ---------------------------------------------------------

type LootView struct {
	Filter  loot.Filter          `json:"filter"`
	Players []loot.PlayerSummary `json:"players"`
	Items   []loot.ItemSummary   `json:"items"`
	Recent  []loot.Entry         `json:"recent"` // newest first, capped
	Total   int                  `json:"total"`
}

func (s *Service) SetLootFilter(f loot.Filter) {
	s.mu.Lock()
	s.lootOpts = f
	s.mu.Unlock()
	s.opts.Emit("loot:changed", nil)
}

func (s *Service) LootView(recent int) LootView {
	s.mu.Lock()
	f := s.lootOpts
	s.mu.Unlock()
	tr := s.eng.Loot()
	all := tr.Entries(f)
	if recent <= 0 || recent > len(all) {
		recent = len(all)
	}
	rec := make([]loot.Entry, 0, recent)
	for i := len(all) - 1; i >= len(all)-recent; i-- {
		rec = append(rec, all[i])
	}
	return LootView{Filter: f, Players: tr.ByPlayer(f), Items: tr.ByItem(f), Recent: rec, Total: len(all)}
}

type CombatRow struct {
	combat.PlayerStats
	DPS   float64 `json:"dps"`
	HPS   float64 `json:"hps"`
	Share float64 `json:"share"` // % of total party damage
}

type CombatView struct {
	Options combat.Options `json:"options"`
	Rows    []CombatRow    `json:"rows"`
	Total   int64          `json:"totalDamage"`
}

func (s *Service) SetCombatOptions(o combat.Options) {
	s.mu.Lock()
	s.opts.Combat = o
	s.mu.Unlock()
	s.eng.Combat().SetOptions(o)
}

func (s *Service) CombatView() CombatView {
	snap := s.eng.Combat().Snapshot()
	var total int64
	for _, p := range snap {
		total += p.Damage
	}
	rows := make([]CombatRow, len(snap))
	for i, p := range snap {
		r := CombatRow{PlayerStats: p, DPS: p.DPS(), HPS: p.HPS()}
		if total > 0 {
			r.Share = float64(p.Damage) * 100 / float64(total)
		}
		rows[i] = r
	}
	s.mu.Lock()
	o := s.opts.Combat
	s.mu.Unlock()
	return CombatView{Options: o, Rows: rows, Total: total}
}

type PartyView struct {
	Me       string               `json:"me"`
	Location string               `json:"location"`
	Members  []events.PartyMember `json:"members"`
}

func (s *Service) PartyView() PartyView {
	st := s.eng.State()
	return PartyView{Me: st.LocalPlayer().Name, Location: st.Location(), Members: st.PartyMembers()}
}

// ---- actions --------------------------------------------------------------

func (s *Service) ResetLoot()     { s.eng.Loot().Reset(); s.opts.Emit("loot:changed", nil) }
func (s *Service) ResetCombat()   { s.eng.Combat().Reset(); s.opts.Emit("combat:changed", nil) }
func (s *Service) ResetFame()     { s.eng.Fame().Reset(); s.opts.Emit("fame:changed", nil) }
func (s *Service) ResetTimeline() { s.eng.Timeline().Reset(); s.opts.Emit("timeline:changed", nil) }

// ResetRates clears the silver/respec/might accumulators (fame is reset
// separately via ResetFame, which the rates dashboard also shows).
func (s *Service) ResetRates() { s.eng.Rates().Reset(); s.opts.Emit("rates:changed", nil) }

// FameView is the current session fame totals and rate.
func (s *Service) FameView() fame.Snapshot { return s.eng.Fame().Snapshot() }

// RatesView is the full dashboard: fame, silver, respec and might as session
// totals plus per-hour rates. Fame comes from the fame tracker (already split
// into base/premium/yellow), the rest from the rates tracker. The frontend
// picks total vs per-hour with a session/hour toggle.
type RatesView struct {
	Fame   rates.Metric `json:"fame"`
	Silver rates.Metric `json:"silver"`
	Respec rates.Metric `json:"respec"`
	Might  rates.Metric `json:"might"`
}

func (s *Service) RatesView() RatesView {
	f := s.eng.Fame().Snapshot()
	r := s.eng.Rates().Snapshot()
	return RatesView{
		Fame:   rates.Metric{Total: f.Total, PerHour: f.PerHour, Count: f.Count, Since: f.Since},
		Silver: r.Silver,
		Respec: r.Respec,
		Might:  r.Might,
	}
}

// TimelineView is the merged chronological feed, newest first, capped at recent.
type TimelineView struct {
	Events []timeline.Event `json:"events"`
	Total  int              `json:"total"`
}

// TimelineView returns the most recent timeline events (newest first).
func (s *Service) TimelineView(recent int) TimelineView {
	tl := s.eng.Timeline()
	return TimelineView{Events: tl.Recent(recent), Total: tl.Len()}
}

// ---- dungeon tracker ------------------------------------------------------

// DungeonRecord is one visited dungeon, enriched for the UI with a live
// duration and an open flag so the frontend renders without recomputing from
// timestamps.
type DungeonRecord struct {
	dungeon.Record
	DurationSec float64 `json:"durationSec"`
	Open        bool    `json:"open"`
}

// DungeonView is the dungeon list, newest first.
type DungeonView struct {
	Records []DungeonRecord `json:"records"`
	Total   int             `json:"total"`
}

// DungeonView returns every dungeon record, newest first.
func (s *Service) DungeonView() DungeonView {
	recs := s.eng.Dungeon().Records()
	now := time.Now()
	out := make([]DungeonRecord, len(recs))
	// Reverse into newest-first order.
	for i, r := range recs {
		end := r.ExitedAt
		if end.IsZero() {
			end = now
		}
		out[len(recs)-1-i] = DungeonRecord{
			Record:      r,
			DurationSec: end.Sub(r.EnteredAt).Seconds(),
			Open:        r.Open(),
		}
	}
	return DungeonView{Records: out, Total: len(recs)}
}

// RenameDungeon sets a user label on a dungeon record.
func (s *Service) RenameDungeon(id int, name string) {
	if s.eng.Dungeon().Rename(id, name) {
		s.opts.Emit("dungeon:changed", nil)
	}
}

// SetDungeonTier tags a dungeon record's tier (free text, e.g. "T6").
func (s *Service) SetDungeonTier(id int, tier string) {
	if s.eng.Dungeon().SetTier(id, tier) {
		s.opts.Emit("dungeon:changed", nil)
	}
}

// ResetDungeons clears the dungeon list.
func (s *Service) ResetDungeons() { s.eng.Dungeon().Reset(); s.opts.Emit("dungeon:changed", nil) }

// ExportLootCSV writes the current (filtered) loot in ao-loot-logger format.
// main.go wraps it with a native save dialog.
func (s *Service) ExportLootCSV(w io.Writer) error {
	s.mu.Lock()
	f := s.lootOpts
	s.mu.Unlock()
	if err := s.eng.Loot().WriteCSV(w, f); err != nil {
		return fmt.Errorf("export csv: %w", err)
	}
	return nil
}

// ---- throttled change notifications --------------------------------------

func (s *Service) onChange(t engine.Topic) {
	s.mu.Lock()
	s.pending[t] = true
	if time.Since(s.lastFlush) < s.opts.Throttle {
		s.mu.Unlock()
		return
	}
	s.lastFlush = time.Now()
	topics := s.pending
	s.pending = map[engine.Topic]bool{}
	s.mu.Unlock()
	for tp := range topics {
		s.opts.Emit(string(tp)+":changed", nil)
	}
}

// Flush emits pending change notifications; main.go calls it on a ticker so
// the last burst before a quiet period is never lost.
func (s *Service) Flush() {
	s.mu.Lock()
	topics := s.pending
	s.pending = map[engine.Topic]bool{}
	s.lastFlush = time.Now()
	s.mu.Unlock()
	for tp := range topics {
		s.opts.Emit(string(tp)+":changed", nil)
	}
}
