package app

import (
	"bytes"
	"encoding/csv"
	"path/filepath"
	"testing"
	"time"

	"github.com/LucasGitDev/albion-companion/internal/events"
	"github.com/LucasGitDev/albion-companion/internal/game"
)

// guardItems resolves a couple of item indexes to names for the feed.
var guardItems = fakeItems{
	10: {Index: 10, UniqueName: "T6_RUNE", Name: "Rune"},
	11: {Index: 11, UniqueName: "T6_SOUL", Name: "Soul"},
}

// seedGuardLoot records: scout "Scout" grabs runes, thief "Rat" grabs runes +
// silver, teammate "Ally" grabs runes.
func seedGuardLoot(svc *Service) (time.Time, *game.State) {
	tr := svc.Engine().Loot()
	st := game.NewState()
	base := time.Date(2026, 11, 3, 20, 0, 0, 0, time.UTC)
	tr.Record(events.GrabbedLoot{Looter: "Scout", LootedFrom: "@MOB_T6", ItemIndex: 10, Quantity: 5}, base, st)
	tr.Record(events.GrabbedLoot{Looter: "Rat", LootedFrom: "@MOB_T6", ItemIndex: 10, Quantity: 3}, base.Add(time.Minute), st)
	tr.Record(events.GrabbedLoot{Looter: "Rat", LootedFrom: "@MOB_T6", ItemIndex: 11, Quantity: 2}, base.Add(2*time.Minute), st)
	tr.Record(events.GrabbedLoot{Looter: "Rat", IsSilver: true, Quantity: 100_0000}, base.Add(3*time.Minute), st)
	tr.Record(events.GrabbedLoot{Looter: "Ally", LootedFrom: "@MOB_T6", ItemIndex: 10, Quantity: 1}, base.Add(4*time.Minute), st)
	return base, st
}

func TestLootGuardView_RosterFlagsNonCollectors(t *testing.T) {
	svc := NewService(Options{Items: guardItems})
	seedGuardLoot(svc)
	svc.SetLootGuardConfig(GuardConfig{Collectors: []string{"Scout"}})

	v := svc.LootGuardView(0)
	// Scout is a collector → not flagged. Rat and Ally are flagged.
	names := map[string]GuardSuspect{}
	for _, s := range v.Suspects {
		names[s.Name] = s
	}
	if _, ok := names["Scout"]; ok {
		t.Fatal("collector Scout should not be flagged")
	}
	if len(v.Suspects) != 2 {
		t.Fatalf("want 2 suspects, got %d: %+v", len(v.Suspects), v.Suspects)
	}
	// Silver excluded by default: Rat's item qty = 3+2 = 5, no silver counted.
	rat := names["Rat"]
	if rat.ItemQty != 5 || rat.Silver != 0 {
		t.Fatalf("Rat = %+v (want itemQty 5, silver 0)", rat)
	}
	if rat.Grabs != 2 {
		t.Fatalf("Rat grabs = %d, want 2 (silver excluded)", rat.Grabs)
	}
	// Suspects sorted by item qty desc → Rat (5) before Ally (1).
	if v.Suspects[0].Name != "Rat" || v.Suspects[1].Name != "Ally" {
		t.Fatalf("suspect order = %s,%s", v.Suspects[0].Name, v.Suspects[1].Name)
	}
	// Feed newest first.
	if len(v.Feed) != 3 {
		t.Fatalf("feed len = %d, want 3 (3 flagged item grabs)", len(v.Feed))
	}
	if !v.Feed[0].At.After(v.Feed[len(v.Feed)-1].At) {
		t.Fatal("feed not newest-first")
	}
	// Rat's item roll-up: rune 3, soul 2.
	if len(rat.Items) != 2 || rat.Items[0].Quantity != 3 {
		t.Fatalf("rat items = %+v", rat.Items)
	}
}

func TestLootGuardView_IncludeSilver(t *testing.T) {
	svc := NewService(Options{Items: guardItems})
	seedGuardLoot(svc)
	svc.SetLootGuardConfig(GuardConfig{Collectors: []string{"Scout"}, IncludeSilver: true})

	v := svc.LootGuardView(0)
	var rat GuardSuspect
	for _, s := range v.Suspects {
		if s.Name == "Rat" {
			rat = s
		}
	}
	if rat.Grabs != 3 { // 2 item grabs + 1 silver
		t.Fatalf("Rat grabs = %d, want 3 with silver included", rat.Grabs)
	}
	if rat.Silver != 100 { // 100_0000 / 10000
		t.Fatalf("Rat silver = %v, want 100", rat.Silver)
	}
}

func TestLootGuardView_ScoutMe(t *testing.T) {
	svc := NewService(Options{Items: guardItems})
	seedGuardLoot(svc)
	// The local player is "Scout"; ScoutMe should exempt them without listing.
	svc.Engine().State().ApplyJoin(events.JoinResponse{ObjectID: 1, Name: "Scout"})
	svc.SetLootGuardConfig(GuardConfig{ScoutMe: true})

	v := svc.LootGuardView(0)
	if v.Me != "Scout" {
		t.Fatalf("me = %q, want Scout", v.Me)
	}
	for _, s := range v.Suspects {
		if s.Name == "Scout" {
			t.Fatal("ScoutMe should exempt the local player")
		}
	}
	if len(v.Suspects) != 2 {
		t.Fatalf("want Rat+Ally flagged, got %d", len(v.Suspects))
	}
}

func TestGuardConfig_PersistRoundTrip(t *testing.T) {
	dir := t.TempDir()
	cfg := GuardConfig{Collectors: []string{"Scout", "Helper"}, ScoutMe: true, IncludeSilver: true}
	if err := saveGuardConfig(dir, cfg); err != nil {
		t.Fatalf("save: %v", err)
	}
	got := loadGuardConfig(dir)
	if len(got.Collectors) != 2 || got.Collectors[0] != "Scout" || !got.ScoutMe || !got.IncludeSilver {
		t.Fatalf("round-trip = %+v", got)
	}
	// A fresh service loads it.
	svc := NewService(Options{ConfigDir: dir})
	if svc.LootGuardConfig().Collectors[1] != "Helper" {
		t.Fatalf("service did not load config: %+v", svc.LootGuardConfig())
	}
	// Missing dir → zero config, no panic.
	if c := loadGuardConfig(filepath.Join(dir, "nope")); len(c.Collectors) != 0 {
		t.Fatalf("missing config should be empty, got %+v", c)
	}
}

func TestSetLootGuardConfig_NormalizesRoster(t *testing.T) {
	svc := NewService(Options{})
	svc.SetLootGuardConfig(GuardConfig{Collectors: []string{" Scout ", "scout", "", "Helper"}})
	got := svc.LootGuardConfig().Collectors
	// Trimmed, de-duped case-insensitively (first form kept), blanks dropped.
	if len(got) != 2 || got[0] != "Scout" || got[1] != "Helper" {
		t.Fatalf("normalized = %+v", got)
	}
}

func TestExportLootGuardCSV(t *testing.T) {
	svc := NewService(Options{Items: guardItems})
	seedGuardLoot(svc)
	svc.SetLootGuardConfig(GuardConfig{Collectors: []string{"Scout"}})

	var buf bytes.Buffer
	if err := svc.ExportLootGuardCSV(&buf); err != nil {
		t.Fatalf("export: %v", err)
	}
	rows, err := csv.NewReader(&buf).ReadAll()
	if err != nil {
		t.Fatalf("parse csv: %v", err)
	}
	// header + 3 flagged item grabs.
	if len(rows) != 4 {
		t.Fatalf("csv rows = %d, want 4", len(rows))
	}
	if rows[0][1] != "looter" || rows[0][4] != "item_name" {
		t.Fatalf("header = %v", rows[0])
	}
}
