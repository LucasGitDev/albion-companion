package engine

import (
	"context"
	"testing"

	"github.com/LucasGitDev/albion-companion/internal/capture"
	"github.com/LucasGitDev/albion-companion/internal/combat"
	"github.com/LucasGitDev/albion-companion/internal/loot"
	"github.com/LucasGitDev/albion-companion/internal/protocol"
)

// TestEngine_ReplayRealSession replays a trimmed real macOS capture
// (testdata/party-dungeon.aorec, produced by tools/trimrec from the TASK-1
// session) and asserts the decoded loot/damage outcome. It is the regression
// guard for TASK-2: any decoder or pipeline regression shifts these numbers.
//
// The fixture splices two windows of the original session: a party fight
// (zone 4333) and a later looting burst, so it exercises Join, PartyJoined,
// NewCharacter, HealthUpdate(s) and OtherGrabbedLoot (items + silver).
func TestEngine_ReplayRealSession(t *testing.T) {
	e := New(Config{
		Codes:  protocol.DefaultCodes(),
		Combat: combat.Options{PartyOnly: true},
		OnDecodeError: func(name string, err error) {
			t.Errorf("decode error %s: %v", name, err)
		},
	})

	if err := e.Run(context.Background(), capture.ReplaySource{Path: "../../testdata/party-dungeon.aorec", Speed: 0}); err != nil {
		t.Fatalf("replay: %v", err)
	}

	// Clean decode of the whole slice.
	st := e.Stats()
	if st.DecodeErrors != 0 {
		t.Fatalf("DecodeErrors = %d, want 0", st.DecodeErrors)
	}
	if st.Unknown != 0 {
		t.Fatalf("Unknown = %d, want 0", st.Unknown)
	}
	if st.Packets != 6636 || st.Messages != 15234 {
		t.Fatalf("packets/messages = %d/%d, want 6636/15234", st.Packets, st.Messages)
	}

	// Local player resolved from the Join response.
	if me := e.State().LocalPlayer().Name; me != "L0rdCap" {
		t.Fatalf("local player = %q, want L0rdCap", me)
	}

	// Damage meter: ranking and absolute damage are stable for this fixture.
	snap := e.Combat().Snapshot()
	if len(snap) < 2 {
		t.Fatalf("combat snapshot has %d players, want >= 2", len(snap))
	}
	if snap[0].Name != "slug88" || snap[0].Damage != 5485 {
		t.Fatalf("top damage = %s/%d, want slug88/5485", snap[0].Name, snap[0].Damage)
	}
	if snap[1].Name != "L0rdCap" || snap[1].Damage != 2413 {
		t.Fatalf("second damage = %s/%d, want L0rdCap/2413", snap[1].Name, snap[1].Damage)
	}
	// Snapshot must be sorted by damage descending.
	for i := 1; i < len(snap); i++ {
		if snap[i-1].Damage < snap[i].Damage {
			t.Fatalf("snapshot not sorted by damage: %+v", snap)
		}
	}

	// Loot: ThalisDON is the top looter with 79 items in the fixture; silver
	// was distributed to party members.
	byPlayer := e.Loot().ByPlayer(loot.Filter{})
	var thalis *loot.PlayerSummary
	var totalItems int64
	var totalSilver float64
	for i := range byPlayer {
		totalItems += byPlayer[i].Items
		totalSilver += byPlayer[i].Silver
		if byPlayer[i].Name == "ThalisDON" {
			thalis = &byPlayer[i]
		}
	}
	if thalis == nil || thalis.Items != 79 {
		t.Fatalf("ThalisDON loot = %+v, want 79 items", thalis)
	}
	if totalItems == 0 || totalSilver == 0 {
		t.Fatalf("expected items and silver recorded, got items=%d silver=%v", totalItems, totalSilver)
	}

	// Timeline merged the whole session and is ordered newest-first.
	tl := e.Timeline().Recent(0)
	if len(tl) == 0 {
		t.Fatal("timeline is empty after replay")
	}
	var kinds int
	seen := map[string]bool{}
	for _, ev := range tl {
		seen[string(ev.Kind)] = true
	}
	kinds = len(seen)
	if !seen["loot"] || !seen["zone"] {
		t.Fatalf("timeline missing expected kinds, saw %v", seen)
	}
	if kinds < 2 {
		t.Fatalf("expected multiple event kinds in timeline, saw %v", seen)
	}
	for i := 1; i < len(tl); i++ {
		if tl[i].At.After(tl[i-1].At) {
			t.Fatalf("timeline not newest-first at %d: %v after %v", i, tl[i].At, tl[i-1].At)
		}
	}
}
