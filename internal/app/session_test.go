package app

import (
	"testing"
	"time"

	"github.com/LucasGitDev/albion-companion/internal/events"
	"github.com/LucasGitDev/albion-companion/internal/game"
	"github.com/LucasGitDev/albion-companion/internal/store"
)

// TestSession_PersistAndLabel exercises the full save→load round-trip with the
// new SessionDetail fields and user-set label/type.
func TestSession_PersistAndLabel(t *testing.T) {
	dir := t.TempDir()
	svc := NewService(Options{Items: guardItems, SessionsDir: dir})

	// Begin a labeled session, then seed loot + a chest withdrawal + timeline.
	svc.StartSession("Avalon run", "avalon")
	st := game.NewState()
	svc.Engine().Loot().Record(events.GrabbedLoot{Looter: "Rat", LootedFrom: "@MOB_T6", ItemIndex: 10, Quantity: 3}, time.Now(), st)

	if info := svc.SessionInfo(); info.Label != "Avalon run" || info.Type != "avalon" || !info.Active {
		t.Fatalf("session info = %+v", info)
	}

	svc.saveSession()

	metas, err := svc.HistoryList()
	if err != nil || len(metas) != 1 {
		t.Fatalf("history list = %v (err %v)", metas, err)
	}
	m := metas[0]
	if m.Label != "Avalon run" || m.Type != "avalon" {
		t.Fatalf("meta label/type = %q/%q", m.Label, m.Type)
	}

	d, err := svc.HistoryLoad(m.ID)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if d.Meta.Label != "Avalon run" || d.Meta.Type != "avalon" {
		t.Fatalf("detail meta = %+v", d.Meta)
	}
	if d.Loot.Total != 1 {
		t.Fatalf("loot total = %d", d.Loot.Total)
	}
	// The new forensic views are present (structs, not nil).
	if d.Timeline.Total < 0 || d.Dungeon.Total < 0 {
		t.Fatalf("unexpected negative totals: %+v %+v", d.Timeline, d.Dungeon)
	}
}

// TestSession_BackwardCompatLoad ensures an old session JSON lacking the new
// fields still unmarshals (zero values), so History keeps loading them.
func TestSession_BackwardCompatLoad(t *testing.T) {
	dir := t.TempDir()
	svc := NewService(Options{SessionsDir: dir})
	id := "20261101-101500"
	old := []byte(`{"meta":{"id":"` + id + `","me":"L0rdCap","lootCount":5},"loot":{"total":5},"combat":{"totalDamage":0}}`)
	if err := svc.store.Save(store.Meta{ID: id, Me: "L0rdCap", LootCount: 5}, old); err != nil {
		t.Fatalf("save: %v", err)
	}
	d, err := svc.HistoryLoad(id)
	if err != nil {
		t.Fatalf("load old session: %v", err)
	}
	if d.Meta.Me != "L0rdCap" || d.Loot.Total != 5 {
		t.Fatalf("old detail = %+v", d.Meta)
	}
	// New fields default cleanly.
	if d.Meta.Label != "" || d.Timeline.Total != 0 || d.Dungeon.Total != 0 {
		t.Fatalf("new fields should be zero on old session: %+v", d)
	}
}

// StartSession with no captured data should not persist an empty session.
func TestStartSession_SkipsEmpty(t *testing.T) {
	dir := t.TempDir()
	svc := NewService(Options{SessionsDir: dir})
	svc.StartSession("empty", "group") // nothing captured before
	svc.saveSession()                  // should be a no-op (no loot/combat/fame/gather)
	metas, _ := svc.HistoryList()
	if len(metas) != 0 {
		t.Fatalf("empty session should not be saved, got %d", len(metas))
	}
}
