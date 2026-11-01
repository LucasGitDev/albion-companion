package engine

import (
	"testing"
	"time"

	"github.com/LucasGitDev/albion-companion/internal/capture"
	"github.com/LucasGitDev/albion-companion/internal/dungeon"
	"github.com/LucasGitDev/albion-companion/internal/photon/photontest"
	"github.com/LucasGitDev/albion-companion/internal/protocol"
)

// TestEngine_DungeonSegmentation drives synthetic Join (zone change) packets
// through the full pipeline and asserts the session is segmented into dungeon
// records: a city → dungeon → city sequence yields one closed record carrying
// the fame and chest seen while inside, and entering a second dungeon opens a
// fresh one.
func TestEngine_DungeonSegmentation(t *testing.T) {
	codes := protocol.DefaultCodes()
	code := func(name string) uint16 {
		c, ok := codes.EventCode(name)
		if !ok {
			t.Fatalf("missing event %s", name)
		}
		return c
	}
	join, _ := codes.OperationCode("Join")

	var dungeonChanges int
	e := New(Config{
		Codes: codes,
		OnChange: func(tp Topic) {
			if tp == TopicDungeon {
				dungeonChanges++
			}
		},
		OnDecodeError: func(name string, err error) { t.Errorf("decode error %s: %v", name, err) },
	})

	t0 := time.Unix(1000, 0)
	feed := func(payload []byte, dt time.Duration) {
		e.Feed(capture.Packet{At: t0.Add(dt), Payload: payload})
	}

	// Start in a city (numeric cluster id): no dungeon.
	feed(photontest.Response(join, map[byte]any{0: int32(1), 2: "Me", 8: "1002"}), 0)
	// Enter a random dungeon.
	feed(photontest.Response(join, map[byte]any{0: int32(1), 2: "Me", 8: "@RANDOMDUNGEON@SOLO_T6"}), time.Minute)
	// A chest appears and the local player gains fame inside the dungeon.
	feed(photontest.Event(code("NewLootChest"), map[byte]any{0: int32(500), 3: "TREASURE_COFFER"}), 90*time.Second)
	feed(photontest.Event(code("UpdateFame"), map[byte]any{0: int32(1), 1: int64(9_000_000), 2: int32(5_000_000), 17: float32(1)}), 2*time.Minute)
	// Leave back to a city: the random dungeon record closes.
	feed(photontest.Response(join, map[byte]any{0: int32(1), 2: "Me", 8: "1002"}), 5*time.Minute)
	// Enter a corrupted dungeon: a second record opens.
	feed(photontest.Response(join, map[byte]any{0: int32(1), 2: "Me", 8: "CORRUPTED_LAIR_STANDARD"}), 6*time.Minute)

	recs := e.Dungeon().Records()
	if len(recs) != 2 {
		t.Fatalf("expected 2 dungeon records, got %d: %+v", len(recs), recs)
	}

	first := recs[0]
	if first.Type != dungeon.TypeRandom {
		t.Errorf("first type = %q, want random", first.Type)
	}
	if first.Open() {
		t.Error("first dungeon should be closed after leaving")
	}
	if first.Chests != 1 {
		t.Errorf("first.Chests = %d, want 1", first.Chests)
	}
	if first.Fame <= 0 {
		t.Errorf("first.Fame = %v, want > 0", first.Fame)
	}
	if !first.EnteredAt.Equal(t0.Add(time.Minute)) || !first.ExitedAt.Equal(t0.Add(5*time.Minute)) {
		t.Errorf("first window wrong: %v..%v", first.EnteredAt, first.ExitedAt)
	}

	second := recs[1]
	if second.Type != dungeon.TypeCorrupted || !second.Open() {
		t.Errorf("second record wrong: %+v", second)
	}
	if dungeonChanges == 0 {
		t.Error("TopicDungeon never signalled")
	}
}
