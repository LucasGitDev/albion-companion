package app

import (
	"testing"
	"time"

	"github.com/LucasGitDev/albion-companion/internal/events"
	"github.com/LucasGitDev/albion-companion/internal/game"
)

func TestKnownPlayers_AddDedupSorted(t *testing.T) {
	k := newKnownPlayers("")
	if !k.add("Rat", "Scout") {
		t.Fatal("first adds should report changed")
	}
	// Case-insensitive de-dup keeps the first casing; blanks ignored.
	if k.add("rat", "  ") {
		t.Fatal("duplicate/blank should not report changed")
	}
	got := k.list()
	if len(got) != 2 || got[0] != "Rat" || got[1] != "Scout" {
		t.Fatalf("list = %v (want [Rat Scout] sorted ci)", got)
	}
}

func TestKnownPlayers_PersistRoundTrip(t *testing.T) {
	dir := t.TempDir()
	k := newKnownPlayers(dir)
	k.add("Alpha", "bravo")
	// A fresh instance loads what the first persisted.
	k2 := newKnownPlayers(dir)
	got := k2.list()
	if len(got) != 2 || got[0] != "Alpha" || got[1] != "bravo" {
		t.Fatalf("reloaded = %v", got)
	}
	// Re-adding a known name (any casing) does not grow the set.
	if k2.add("ALPHA") {
		t.Fatal("known name should not be new")
	}
}

func TestKnownPlayers_HarvestFromViews(t *testing.T) {
	svc := NewService(Options{Items: guardItems})
	seedGuardLoot(svc) // records Scout/Rat/Ally as looters
	svc.Engine().State().ApplyJoin(events.JoinResponse{ObjectID: 1, Name: "Me"})

	got := svc.KnownPlayers()
	has := func(n string) bool {
		for _, x := range got {
			if x == n {
				return true
			}
		}
		return false
	}
	for _, n := range []string{"Scout", "Rat", "Ally", "Me"} {
		if !has(n) {
			t.Fatalf("KnownPlayers missing %q: %v", n, got)
		}
	}
}

// guard against a nil-known panic on the OnLoot path before the engine runs.
func TestKnownPlayers_OnLootFeed(t *testing.T) {
	svc := NewService(Options{Items: guardItems})
	st := game.NewState()
	svc.Engine().Loot().Record(events.GrabbedLoot{Looter: "Sneaky", ItemIndex: 10, Quantity: 1}, time.Now(), st)
	// Record alone doesn't invoke OnLoot (that's engine-driven), so harvest picks
	// it up from the loot view.
	if got := svc.KnownPlayers(); len(got) == 0 {
		t.Fatalf("expected harvested players, got %v", got)
	}
}
