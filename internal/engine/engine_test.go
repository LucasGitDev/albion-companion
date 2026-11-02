package engine

import (
	"testing"
	"time"

	"github.com/LucasGitDev/albion-companion/internal/capture"
	"github.com/LucasGitDev/albion-companion/internal/combat"
	"github.com/LucasGitDev/albion-companion/internal/itemdb"
	"github.com/LucasGitDev/albion-companion/internal/loot"
	"github.com/LucasGitDev/albion-companion/internal/photon/photontest"
	"github.com/LucasGitDev/albion-companion/internal/protocol"
)

type items map[int]itemdb.Item

func (i items) Resolve(n int) (itemdb.Item, bool) { it, ok := i[n]; return it, ok }

func guid(b byte) []byte { g := make([]byte, 16); g[0] = b; return g }

// TestEngine_EndToEnd drives synthetic Photon packets through the full
// pipeline: photon → protocol → events → state/loot/combat.
func TestEngine_EndToEnd(t *testing.T) {
	codes := protocol.DefaultCodes()
	code := func(name string) uint16 {
		c, ok := codes.EventCode(name)
		if !ok {
			t.Fatalf("missing event %s", name)
		}
		return c
	}
	join, _ := codes.OperationCode("Join")

	var topics []Topic
	e := New(Config{
		Codes:    codes,
		Items:    items{2888: {Index: 2888, UniqueName: "T4_BAG", Name: "Adept's Bag"}},
		Combat:   combat.Options{PartyOnly: true},
		OnChange: func(tp Topic) { topics = append(topics, tp) },
		OnDecodeError: func(name string, err error) {
			t.Errorf("decode error %s: %v", name, err)
		},
	})

	t0 := time.Unix(1000, 0)
	feed := func(payload []byte, dt time.Duration) {
		e.Feed(capture.Packet{At: t0.Add(dt), SrcIP: "5.188.125.1", Payload: payload})
	}

	feed(photontest.Response(join, map[byte]any{0: int32(1), 1: guid(1), 2: "Me", 8: "3008", 58: "MyGuild"}), 0)
	feed(photontest.Event(code("NewCharacter"), map[byte]any{0: int32(2), 1: "Friend", 7: guid(2), 8: "MyGuild"}), 0)
	feed(photontest.Event(code("PartyJoined"), map[byte]any{8: [][]byte{guid(1), guid(2)}, 9: []string{"Me", "Friend"}}), 0)
	feed(photontest.Event(code("OtherGrabbedLoot"), map[byte]any{0: int32(50), 1: "@MOB_X", 2: "Friend", 4: int32(2888), 5: int32(3)}), time.Second)
	feed(photontest.Event(code("HealthUpdate"), map[byte]any{0: int32(900), 2: float32(-120), 3: float32(500), 6: int32(1)}), time.Second)
	feed(photontest.Event(code("HealthUpdates"), map[byte]any{0: int32(900), 2: []float32{-80, -40}, 6: []int32{1, 2}}), 3*time.Second)

	if st := e.Stats(); st.DecodeErrors != 0 || st.Messages != 6 {
		t.Fatalf("stats = %+v", st)
	}
	if lp := e.State().LocalPlayer(); lp.Name != "Me" || e.State().Location() != "3008" {
		t.Fatalf("local player = %+v", lp)
	}
	if !e.State().InParty("Friend") {
		t.Fatal("Friend should be in party")
	}

	entries := e.Loot().Entries(loot.Filter{PartyOnly: true})
	if len(entries) != 1 || entries[0].ItemID != "T4_BAG" || entries[0].Quantity != 3 || entries[0].LooterGuild != "MyGuild" {
		t.Fatalf("loot = %+v", entries)
	}

	snap := e.Combat().Snapshot()
	if len(snap) != 2 || snap[0].Name != "Me" || snap[0].Damage != 200 || snap[1].Damage != 40 {
		t.Fatalf("combat = %+v", snap)
	}
	if snap[0].DPS() != 100 { // 200 dmg over 2s
		t.Fatalf("dps = %v", snap[0].DPS())
	}
	if len(topics) == 0 {
		t.Fatal("OnChange never called")
	}

	// Timeline merges the session newest-first: loot, two party joins, zone.
	tl := e.Timeline().Recent(0)
	if len(tl) != 4 {
		t.Fatalf("timeline len = %d, want 4: %+v", len(tl), tl)
	}
	if tl[0].Kind != "loot" || tl[0].Item != "Adept's Bag" || tl[0].Quantity != 3 {
		t.Errorf("timeline[0] = %+v", tl[0])
	}
	if tl[1].Kind != "party_joined" || tl[1].Player != "Friend" {
		t.Errorf("timeline[1] = %+v", tl[1])
	}
	if tl[2].Kind != "party_joined" || tl[2].Player != "Me" {
		t.Errorf("timeline[2] = %+v", tl[2])
	}
	if tl[3].Kind != "zone" || tl[3].Zone != "3008" {
		t.Errorf("timeline[3] = %+v", tl[3])
	}
	// Timestamps must be non-increasing (newest first).
	for i := 1; i < len(tl); i++ {
		if tl[i].At.After(tl[i-1].At) {
			t.Errorf("timeline out of order at %d: %v after %v", i, tl[i].At, tl[i-1].At)
		}
	}
}

// TestEngine_ChestLoot drives the four-event chest chain (NewLootChest →
// NewSimpleItem → AttachItemContainer → InventoryPutItem) and asserts the taken
// item is recorded as loot from the chest for the local player. Items in a
// non-chest container (the player's own inventory) must not be logged.
func TestEngine_ChestLoot(t *testing.T) {
	codes := protocol.DefaultCodes()
	code := func(name string) uint16 {
		c, ok := codes.EventCode(name)
		if !ok {
			t.Fatalf("missing event %s", name)
		}
		return c
	}
	join, _ := codes.OperationCode("Join")

	var looted []loot.Entry
	e := New(Config{
		Codes:  codes,
		Items:  items{2014: {Index: 2014, UniqueName: "T5_SOUL", Name: "Expert's Soul"}},
		OnLoot: func(en loot.Entry) { looted = append(looted, en) },
		OnDecodeError: func(name string, err error) {
			t.Errorf("decode error %s: %v", name, err)
		},
	})

	t0 := time.Unix(2000, 0)
	feed := func(payload []byte, dt time.Duration) {
		e.Feed(capture.Packet{At: t0.Add(dt), Payload: payload})
	}

	feed(photontest.Response(join, map[byte]any{0: int32(1), 2: "Me", 8: "9", 54: guid(9)}), 0)
	feed(photontest.Event(code("NewLootChest"), map[byte]any{0: int32(500), 3: "TREASURE_COFFER_SOLO"}), time.Second)
	feed(photontest.Event(code("NewSimpleItem"), map[byte]any{0: int32(900), 1: int16(2014), 2: int16(1)}), time.Second)
	// An item that lands in the player's own inventory (container 5), not a chest.
	feed(photontest.Event(code("NewSimpleItem"), map[byte]any{0: int32(901), 1: int16(2014), 2: int16(1)}), time.Second)
	feed(photontest.Event(code("AttachItemContainer"), map[byte]any{0: int32(500), 3: []int32{900}, 5: int16(1)}), time.Second)
	feed(photontest.Event(code("AttachItemContainer"), map[byte]any{0: int32(5), 3: []int32{901}, 5: int16(1)}), time.Second)
	feed(photontest.Event(code("InventoryPutItem"), map[byte]any{0: int32(901), 2: guid(9)}), 2*time.Second) // inventory item: ignored
	feed(photontest.Event(code("InventoryPutItem"), map[byte]any{0: int32(900), 2: guid(9)}), 3*time.Second) // chest item: logged

	if st := e.Stats(); st.DecodeErrors != 0 {
		t.Fatalf("decode errors = %d", st.DecodeErrors)
	}
	if len(looted) != 1 {
		t.Fatalf("expected exactly 1 chest loot entry, got %d: %+v", len(looted), looted)
	}
	got := looted[0]
	if got.Looter != "Me" || got.LootedFrom != "TREASURE_COFFER_SOLO" || got.ItemID != "T5_SOUL" || got.Quantity != 1 {
		t.Fatalf("unexpected chest loot: %+v", got)
	}

	// The same withdrawal lands in the personal chest-activity log (TASK-14),
	// and the inventory item (container 5) must NOT appear there.
	cl := e.ChestLog().Snapshot(0)
	if cl.Total != 1 || cl.Withdrawn != 1 {
		t.Fatalf("chest log = %+v want 1 withdrawal", cl)
	}
	if en := cl.Entries[0]; en.Kind != "withdraw" || en.Container != "TREASURE_COFFER_SOLO" || en.ItemID != "T5_SOUL" || en.Quantity != 1 {
		t.Fatalf("chest log entry = %+v", en)
	}
}

// TestEngine_CorpseBagLoot covers taking an item from a corpse bag / loot body
// (NewLoot container), not a chest — the user's "open my death bag, take the log"
// case. It must be recorded as loot with lootedFrom = the body name.
func TestEngine_CorpseBagLoot(t *testing.T) {
	codes := protocol.DefaultCodes()
	code := func(name string) uint16 {
		c, ok := codes.EventCode(name)
		if !ok {
			t.Fatalf("missing event %s", name)
		}
		return c
	}
	join, _ := codes.OperationCode("Join")

	var looted []loot.Entry
	e := New(Config{
		Codes:         codes,
		Items:         items{566: {Index: 566, UniqueName: "T1_WOOD", Name: "Log"}},
		OnLoot:        func(en loot.Entry) { looted = append(looted, en) },
		OnDecodeError: func(name string, err error) { t.Errorf("decode error %s: %v", name, err) },
	})

	t0 := time.Unix(3000, 0)
	feed := func(payload []byte, dt time.Duration) {
		e.Feed(capture.Packet{At: t0.Add(dt), Payload: payload})
	}

	feed(photontest.Response(join, map[byte]any{0: int32(1), 2: "Me", 8: "9", 54: guid(9)}), 0)
	// A corpse bag (NewLoot), its item, the attach, and the take.
	feed(photontest.Event(code("NewLoot"), map[byte]any{0: int32(700), 3: "react01"}), time.Second)
	feed(photontest.Event(code("NewSimpleItem"), map[byte]any{0: int32(910), 1: int16(566), 2: int16(1)}), time.Second)
	feed(photontest.Event(code("AttachItemContainer"), map[byte]any{0: int32(700), 3: []int32{910}, 5: int16(1)}), time.Second)
	feed(photontest.Event(code("InventoryPutItem"), map[byte]any{0: int32(910), 2: guid(9)}), 2*time.Second)

	if st := e.Stats(); st.DecodeErrors != 0 {
		t.Fatalf("decode errors = %d", st.DecodeErrors)
	}
	if len(looted) != 1 {
		t.Fatalf("expected 1 corpse-bag loot entry, got %d: %+v", len(looted), looted)
	}
	if got := looted[0]; got.Looter != "Me" || got.LootedFrom != "react01" || got.ItemID != "T1_WOOD" || got.Quantity != 1 {
		t.Fatalf("unexpected corpse loot: %+v", got)
	}
}

// TestEngine_Gathering covers HarvestFinished: only the local player's harvests
// accumulate; another player's gathering nearby is ignored.
func TestEngine_Gathering(t *testing.T) {
	codes := protocol.DefaultCodes()
	code := func(name string) uint16 {
		c, ok := codes.EventCode(name)
		if !ok {
			t.Fatalf("missing event %s", name)
		}
		return c
	}
	join, _ := codes.OperationCode("Join")

	e := New(Config{
		Codes:         codes,
		OnDecodeError: func(name string, err error) { t.Errorf("decode error %s: %v", name, err) },
	})

	t0 := time.Unix(4000, 0)
	feed := func(payload []byte, dt time.Duration) {
		e.Feed(capture.Packet{At: t0.Add(dt), Payload: payload})
	}

	// Local player is object id 1.
	feed(photontest.Response(join, map[byte]any{0: int32(1), 2: "Me", 8: "9", 54: guid(9)}), 0)
	// Local player harvests wood twice (base + collector + premium bonuses).
	feed(photontest.Event(code("HarvestFinished"), map[byte]any{0: int32(1), 4: int32(566), 5: int16(3), 6: int16(1), 7: int16(2)}), time.Second)
	feed(photontest.Event(code("HarvestFinished"), map[byte]any{0: int32(1), 4: int32(566), 5: int16(4)}), 2*time.Second)
	// Another player (object id 2) harvests ore — must be ignored.
	feed(photontest.Event(code("HarvestFinished"), map[byte]any{0: int32(2), 4: int32(100), 5: int16(99)}), 3*time.Second)

	if st := e.Stats(); st.DecodeErrors != 0 {
		t.Fatalf("decode errors = %d", st.DecodeErrors)
	}
	s := e.Gather().Snapshot()
	if s.Total != 10 { // 3+1+2 then 4
		t.Fatalf("Total = %d want 10 (other player's 99 must be ignored)", s.Total)
	}
	if s.Count != 2 {
		t.Fatalf("Count = %d want 2", s.Count)
	}
	if len(s.Items) != 1 || s.Items[0].ItemIndex != 566 || s.Items[0].Amount != 10 {
		t.Fatalf("unexpected gather items: %+v", s.Items)
	}
}

// TestEngine_Route covers the zone-transition route graph (TASK-34): each Join
// records a visit; distinct cluster ids (incl. Avalon-road-like ids) become
// distinct nodes, consecutive joins become edges, re-joining the same zone is
// ignored.
func TestEngine_Route(t *testing.T) {
	codes := protocol.DefaultCodes()
	join, _ := codes.OperationCode("Join")
	e := New(Config{Codes: codes, OnDecodeError: func(n string, err error) { t.Errorf("decode %s: %v", n, err) }})

	t0 := time.Unix(5000, 0)
	joinZone := func(loc string, dt time.Duration) {
		e.Feed(capture.Packet{At: t0.Add(dt), Payload: photontest.Response(join, map[byte]any{0: int32(1), 2: "Me", 8: loc, 54: guid(9)})})
	}
	joinZone("Lymhurst", 0)
	joinZone("3005-Avalon-Road", time.Minute)    // distinct cluster id (Avalon road)
	joinZone("3005-Avalon-Road", 90*time.Second) // re-join: ignored
	joinZone("Caerleon", 2*time.Minute)

	s := e.Route().Snapshot(0)
	if s.TotalVisits != 3 {
		t.Fatalf("TotalVisits = %d want 3", s.TotalVisits)
	}
	if len(s.Nodes) != 3 {
		t.Fatalf("nodes = %d want 3 (avalon road distinct): %+v", len(s.Nodes), s.Nodes)
	}
	if len(s.Edges) != 2 {
		t.Fatalf("edges = %d want 2: %+v", len(s.Edges), s.Edges)
	}
	if s.Path[0].Zone != "Lymhurst" || s.Path[1].Zone != "3005-Avalon-Road" || s.Path[2].Zone != "Caerleon" {
		t.Fatalf("path wrong: %+v", s.Path)
	}
}
