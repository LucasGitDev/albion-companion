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
