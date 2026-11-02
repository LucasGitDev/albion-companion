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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/LucasGitDev/albion-companion/internal/capture"
	"github.com/LucasGitDev/albion-companion/internal/chestlog"
	"github.com/LucasGitDev/albion-companion/internal/combat"
	"github.com/LucasGitDev/albion-companion/internal/dungeon"
	"github.com/LucasGitDev/albion-companion/internal/engine"
	"github.com/LucasGitDev/albion-companion/internal/events"
	"github.com/LucasGitDev/albion-companion/internal/fame"
	"github.com/LucasGitDev/albion-companion/internal/gather"
	"github.com/LucasGitDev/albion-companion/internal/loot"
	"github.com/LucasGitDev/albion-companion/internal/prices"
	"github.com/LucasGitDev/albion-companion/internal/protocol"
	"github.com/LucasGitDev/albion-companion/internal/rates"
	"github.com/LucasGitDev/albion-companion/internal/route"
	"github.com/LucasGitDev/albion-companion/internal/store"
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
	Prices        *prices.Client    // optional: price client (nil = built with defaults)
	Recorder      *capture.Recorder // optional: record every captured packet to an .aorec
	SessionsDir   string            // optional: persist finished sessions here (empty = no history)
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

	prices *prices.Client

	mu           sync.Mutex
	cancel       context.CancelFunc
	status       CaptureStatus
	pending      map[engine.Topic]bool
	lastFlush    time.Time
	lootOpts     loot.Filter
	regionManual bool   // user explicitly chose the price region (stops auto-detect)
	regionSource string // "default" | "detected" | "manual"

	store        *store.Store // nil when SessionsDir is unset
	sessionStart time.Time    // start of the current capture session (for persistence)
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
	s := &Service{opts: opts, pending: map[engine.Topic]bool{}, lootOpts: loot.Filter{PartyOnly: opts.PartyOnlyLoot}, regionSource: "default"}
	if opts.SessionsDir != "" {
		s.store = store.New(opts.SessionsDir)
	}
	s.prices = opts.Prices
	if s.prices == nil {
		// Prices land asynchronously; nudge the UI to re-pull the loot view so
		// values fill in without a user action. OnUpdate runs off a background
		// goroutine, so this never blocks the engine.
		s.prices = prices.New(prices.Options{OnUpdate: func() { s.opts.Emit("loot:changed", nil) }})
	}
	s.eng = engine.New(engine.Config{
		Codes:    opts.Codes,
		Items:    opts.Items,
		Combat:   opts.Combat,
		Recorder: opts.Recorder,
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
	s.sessionStart = time.Now()
	for _, src := range sources {
		s.status.Sources = append(s.status.Sources, src.Name())
	}
	s.mu.Unlock()

	go func() {
		err := s.eng.Run(ctx, sources...)
		// Persist the finished session before clearing state, so History keeps it.
		s.saveSession()
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
		s.opts.Emit("history:changed", nil)
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
	Value   ValueView            `json:"value"`
}

// ValueView is the estimated market value of the (filtered) loot, from the
// Albion Online Data project. It is best-effort and external: Available is
// false when no price data has arrived (offline, or still fetching), in which
// case the UI shows values as unknown. Nothing here affects the engine.
type ValueView struct {
	Region        string           `json:"region"`        // region key: "west"/"east"/"europe"
	RegionLabel   string           `json:"regionLabel"`   // "Americas"/"Asia"/"Europe"
	RegionSource  string           `json:"regionSource"`  // "default" | "detected" | "manual"
	Available     bool             `json:"available"`     // at least one item has a price
	Pending       bool             `json:"pending"`       // some prices are still being fetched
	Total         int64            `json:"total"`         // estimated silver value of looted items
	SilverPerHour float64          `json:"silverPerHour"` // (item value + looted silver) per session hour
	PerPlayer     map[string]int64 `json:"perPlayer"`     // looter -> estimated item value
	PerItem       map[string]int64 `json:"perItem"`       // itemId -> estimated unit price
	Priced        int              `json:"priced"`        // distinct items with a price
	Missing       int              `json:"missing"`       // distinct items without a price yet
}

// reconcileRegion keeps the price region in sync with the detected game server
// unless the user has manually chosen one. Caller must not hold s.mu.
func (s *Service) reconcileRegion() {
	s.mu.Lock()
	manual := s.regionManual
	s.mu.Unlock()
	if manual {
		return
	}
	if r, ok := prices.RegionFromIP(s.eng.ServerIP()); ok {
		if s.prices.Region().Key != r.Key {
			s.prices.SetRegion(r)
		}
		s.mu.Lock()
		s.regionSource = "detected"
		s.mu.Unlock()
	}
}

// valueView computes the estimated value of the given (already filtered)
// entries. It reads prices from cache only and schedules missing ones in the
// background; it never blocks on the network.
func (s *Service) valueView(entries []loot.Entry) ValueView {
	s.reconcileRegion()
	region := s.prices.Region()
	s.mu.Lock()
	src := s.regionSource
	s.mu.Unlock()

	v := ValueView{
		Region:       region.Key,
		RegionLabel:  region.Label,
		RegionSource: src,
		PerPlayer:    map[string]int64{},
		PerItem:      map[string]int64{},
	}

	ids := make([]string, 0, len(entries))
	seen := map[string]bool{}
	var lootedSilver float64
	for _, e := range entries {
		if e.IsSilver {
			lootedSilver += e.Silver
			continue
		}
		if e.ItemID == "" || seen[e.ItemID] {
			continue
		}
		seen[e.ItemID] = true
		ids = append(ids, e.ItemID)
	}

	priceByID, complete := s.prices.Lookup(ids)
	v.PerItem = priceByID
	v.Priced = len(priceByID)
	v.Missing = len(ids) - len(priceByID)
	v.Pending = !complete
	v.Available = len(priceByID) > 0

	for _, e := range entries {
		if e.IsSilver {
			continue
		}
		if p, ok := priceByID[e.ItemID]; ok {
			val := p * e.Quantity
			v.Total += val
			v.PerPlayer[e.Looter] += val
		}
	}

	// silver/hour equivalent = (estimated item value + looted silver) / hours.
	if hours := time.Since(s.eng.Loot().StartedAt()).Hours(); hours > 0 {
		v.SilverPerHour = (float64(v.Total) + lootedSilver) / hours
	}
	return v
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
	return LootView{
		Filter:  f,
		Players: tr.ByPlayer(f),
		Items:   tr.ByItem(f),
		Recent:  rec,
		Total:   len(all),
		Value:   s.valueView(all),
	}
}

// ---- prices (external, optional) ------------------------------------------

// PriceRegions lists the selectable market regions for the Settings dropdown.
func (s *Service) PriceRegions() []prices.Region { return prices.Regions }

// SetPriceRegion overrides the market region and stops IP auto-detection. An
// empty or unknown key re-enables auto-detection (back to the detected/default
// region). Triggers a loot re-pull so values refresh against the new region.
func (s *Service) SetPriceRegion(key string) {
	if r, ok := prices.RegionByKey(key); ok {
		s.prices.SetRegion(r)
		s.mu.Lock()
		s.regionManual = true
		s.regionSource = "manual"
		s.mu.Unlock()
	} else {
		s.mu.Lock()
		s.regionManual = false
		s.regionSource = "default"
		s.mu.Unlock()
	}
	s.opts.Emit("loot:changed", nil)
}

// RefreshPrices force-fetches prices for every distinct item currently in the
// (unfiltered) loot session, bypassing the cache TTL. It fails soft: a network
// error is returned for logging, the UI simply keeps showing what it has. Runs
// synchronously, so the Wails binding should call it off the UI thread (Wails
// already dispatches bound calls on their own goroutine).
func (s *Service) RefreshPrices() error {
	s.reconcileRegion()
	all := s.eng.Loot().Entries(loot.Filter{IncludeTrash: true})
	ids := make([]string, 0, len(all))
	for _, e := range all {
		if !e.IsSilver && e.ItemID != "" {
			ids = append(ids, e.ItemID)
		}
	}
	err := s.prices.Refresh(context.Background(), ids)
	s.opts.Emit("loot:changed", nil)
	return err
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

// ResetGather clears the gathered-resource totals.
func (s *Service) ResetGather() { s.eng.Gather().Reset(); s.opts.Emit("gather:changed", nil) }

// GatherView is the current session's gathered resources, by item and total.
func (s *Service) GatherView() gather.Snapshot { return s.eng.Gather().Snapshot() }

// ResetChestLog clears the personal chest-activity log.
func (s *Service) ResetChestLog() { s.eng.ChestLog().Reset(); s.opts.Emit("chestlog:changed", nil) }

// ResetRoute clears the zone route graph.
func (s *Service) ResetRoute() { s.eng.Route().Reset(); s.opts.Emit("route:changed", nil) }

// RouteView is the session's zone route graph (ordered path + nodes + edges).
func (s *Service) RouteView(recent int) route.Snapshot {
	if recent <= 0 {
		recent = 500
	}
	return s.eng.Route().Snapshot(recent)
}

// ChestLogView is the local player's container-activity log (withdrawals), newest
// first. It is a personal log only — chest put/take is client-local, so it never
// reflects other members, and deposits are not passively observable (decision-7).
func (s *Service) ChestLogView(recent int) chestlog.Snapshot {
	if recent <= 0 {
		recent = 200
	}
	return s.eng.ChestLog().Snapshot(recent)
}

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

// ---- session persistence / history ---------------------------------------

// SessionDetail is a finished session's full reopenable state. It is marshaled
// into the store's opaque detail blob and returned verbatim to the History tab,
// which renders it read-only with the same components as the live views.
type SessionDetail struct {
	Meta   store.Meta      `json:"meta"`
	Loot   LootView        `json:"loot"`
	Combat CombatView      `json:"combat"`
	Rates  RatesView       `json:"rates"`
	Gather gather.Snapshot `json:"gather"`
	Fame   fame.Snapshot   `json:"fame"`
}

// lootViewFull is an unfiltered loot view (all players, trash included) used for
// persistence and session CSV export, independent of the live filter.
func (s *Service) lootViewFull() LootView {
	f := loot.Filter{IncludeTrash: true}
	tr := s.eng.Loot()
	all := tr.Entries(f)
	rec := make([]loot.Entry, 0, len(all))
	for i := len(all) - 1; i >= 0; i-- {
		rec = append(rec, all[i])
	}
	return LootView{Filter: f, Players: tr.ByPlayer(f), Items: tr.ByItem(f), Recent: rec, Total: len(all), Value: s.valueView(all)}
}

// saveSession persists the just-ended session. No-op when history is disabled or
// the session captured nothing worth keeping.
func (s *Service) saveSession() {
	s.mu.Lock()
	st := s.store
	start := s.sessionStart
	s.sessionStart = time.Time{}
	s.mu.Unlock()
	if st == nil || start.IsZero() {
		return
	}

	lv := s.lootViewFull()
	cv := s.CombatView()
	rv := s.RatesView()
	gv := s.GatherView()
	fv := s.FameView()
	if lv.Total == 0 && cv.Total == 0 && fv.Total == 0 && gv.Total == 0 {
		return // nothing captured; don't clutter History with empty sessions
	}

	gs := s.eng.State()
	meta := store.Meta{
		ID:        store.NewID(start),
		StartedAt: start,
		EndedAt:   time.Now(),
		Zone:      gs.Location(),
		Me:        gs.LocalPlayer().Name,
		LootCount: lv.Total,
		Silver:    rv.Silver.Total,
		Fame:      fv.Total,
		Damage:    cv.Total,
	}
	detail, err := json.Marshal(SessionDetail{Meta: meta, Loot: lv, Combat: cv, Rates: rv, Gather: gv, Fame: fv})
	if err != nil {
		return
	}
	_ = st.Save(meta, detail) // best-effort; a write error must not crash capture
}

// HistoryList returns saved sessions, newest first (empty when disabled).
func (s *Service) HistoryList() ([]store.Meta, error) {
	if s.store == nil {
		return nil, nil
	}
	return s.store.List()
}

// HistoryLoad returns one saved session's full reopenable detail.
func (s *Service) HistoryLoad(id string) (SessionDetail, error) {
	var d SessionDetail
	if s.store == nil {
		return d, fmt.Errorf("history: persistence disabled")
	}
	_, raw, err := s.store.Load(id)
	if err != nil {
		return d, err
	}
	if err := json.Unmarshal(raw, &d); err != nil {
		return d, err
	}
	return d, nil
}

// DeleteSession removes one saved session.
func (s *Service) DeleteSession(id string) error {
	if s.store == nil {
		return nil
	}
	err := s.store.Delete(id)
	s.opts.Emit("history:changed", nil)
	return err
}

// WriteSessionCSV writes a saved session's loot as CSV (ao-loot-logger format).
func (s *Service) WriteSessionCSV(w io.Writer, id string) error {
	d, err := s.HistoryLoad(id)
	if err != nil {
		return err
	}
	return loot.WriteEntriesCSV(w, d.Loot.Recent)
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
