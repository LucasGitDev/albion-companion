package app

import (
	"path/filepath"
	"testing"

	"github.com/LucasGitDev/albion-companion/internal/protocol"
)

// TestReprocessAorec replays a recorded fixture into a fresh history session,
// without any live capture running.
func TestReprocessAorec(t *testing.T) {
	rec := filepath.Join("..", "..", "testdata", "party-dungeon.aorec")
	dir := t.TempDir()
	svc := NewService(Options{Codes: protocol.DefaultCodes(), SessionsDir: dir})

	if err := svc.ReprocessAorec(rec, "Replay test", "avalon"); err != nil {
		t.Fatalf("reprocess: %v", err)
	}

	metas, err := svc.HistoryList()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(metas) != 1 {
		t.Fatalf("want 1 saved session, got %d", len(metas))
	}
	m := metas[0]
	if m.Label != "Replay test" || m.Type != "avalon" {
		t.Fatalf("meta label/type = %q/%q", m.Label, m.Type)
	}
	// The recording carried combat and/or loot, so bounds come from the data
	// (not the zero time), and the session is non-empty.
	if m.StartedAt.IsZero() {
		t.Fatal("startedAt should be set from the recording")
	}
	if m.LootCount == 0 && m.Damage == 0 {
		t.Fatalf("expected captured loot or damage from the fixture, meta=%+v", m)
	}

	d, err := svc.HistoryLoad(m.ID)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if d.Meta.Label != "Replay test" {
		t.Fatalf("detail meta = %+v", d.Meta)
	}
}

// Empty label defaults to the recording's base file name.
func TestReprocessAorec_DefaultLabel(t *testing.T) {
	rec := filepath.Join("..", "..", "testdata", "party-dungeon.aorec")
	dir := t.TempDir()
	svc := NewService(Options{Codes: protocol.DefaultCodes(), SessionsDir: dir})
	if err := svc.ReprocessAorec(rec, "", ""); err != nil {
		t.Fatalf("reprocess: %v", err)
	}
	metas, _ := svc.HistoryList()
	if len(metas) != 1 || metas[0].Label != "party-dungeon" {
		t.Fatalf("default label = %q (want party-dungeon)", metas[0].Label)
	}
}

// Reprocess with history disabled errors cleanly, and a blank path too.
func TestReprocessAorec_Guards(t *testing.T) {
	noStore := NewService(Options{Codes: protocol.DefaultCodes()})
	if err := noStore.ReprocessAorec("x.aorec", "", ""); err == nil {
		t.Fatal("expected error when history is disabled")
	}
	withStore := NewService(Options{Codes: protocol.DefaultCodes(), SessionsDir: t.TempDir()})
	if err := withStore.ReprocessAorec("  ", "", ""); err == nil {
		t.Fatal("expected error for blank path")
	}
}
