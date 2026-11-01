package dungeon

import (
	"testing"
	"time"
)

func at(s int) time.Time { return time.Unix(int64(s), 0) }

func TestClassify(t *testing.T) {
	cases := map[string]Type{
		"4333":                        TypeNone, // hideout / numeric map id
		"":                            TypeNone,
		"1002":                        TypeNone, // open-world zone id
		"@RANDOMDUNGEON@SOLO_T6_guid": TypeRandom,
		"CORRUPTED_LAIR_STANDARD@abc": TypeCorrupted,
		"HELLGATE_2V2_guid":           TypeHellgate,
		"@HG@xyz":                     TypeHellgate,
		"AVALON_ROADS_DUNGEON":        TypeAvalon,
		"TUNNEL_MISTS_GROUP":          TypeMists,
		"EXPEDITION_SOLO":             TypeRandom,
		"GROUPDUNGEON_T7":             TypeGroup,
		"SOME_DUNGEON_UNKNOWN":        TypeOther,
	}
	for in, want := range cases {
		if got := Classify(in); got != want {
			t.Errorf("Classify(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestOpenCloseSegment(t *testing.T) {
	tr := New()
	snap := func() []CombatSnap { return []CombatSnap{{Name: "me", Damage: 100, DPS: 10}} }

	// Start in a city (numeric id): nothing opens.
	if tr.OnZone("1002", at(0), snap) {
		t.Fatal("city zone should not open a record")
	}
	if len(tr.Records()) != 0 {
		t.Fatal("no records expected yet")
	}

	// Enter a random dungeon: a record opens.
	if !tr.OnZone("@RANDOMDUNGEON@SOLO", at(10), snap) {
		t.Fatal("entering dungeon should report change")
	}
	tr.AddFame(500)
	tr.AddSilver(1200)
	tr.AddLoot(3)
	tr.AddChest()
	tr.AddChest()

	recs := tr.Records()
	if len(recs) != 1 || !recs[0].Open() {
		t.Fatalf("expected 1 open record, got %+v", recs)
	}

	// Leave to a city: the record closes with a combat snapshot.
	if !tr.OnZone("1002", at(70), snap) {
		t.Fatal("leaving dungeon should report change")
	}
	recs = tr.Records()
	r := recs[0]
	if r.Open() {
		t.Fatal("record should be closed after leaving")
	}
	if r.Fame != 500 || r.Silver != 1200 || r.Loot != 3 || r.Chests != 2 {
		t.Fatalf("totals wrong: %+v", r)
	}
	if !r.ExitedAt.Equal(at(70)) || r.EnteredAt != at(10) {
		t.Fatalf("times wrong: entered=%v exited=%v", r.EnteredAt, r.ExitedAt)
	}
	if len(r.Combat) != 1 || r.Combat[0].DPS != 10 {
		t.Fatalf("combat snapshot missing: %+v", r.Combat)
	}

	// Fame while outside a dungeon is dropped (not attributed to the closed one).
	tr.AddFame(999)
	if tr.Records()[0].Fame != 500 {
		t.Fatal("fame leaked into a closed record")
	}
}

func TestReJoinSameDungeonDoesNotSegment(t *testing.T) {
	tr := New()
	tr.OnZone("@RANDOMDUNGEON@A", at(0), nil)
	if tr.OnZone("@RANDOMDUNGEON@A", at(5), nil) {
		t.Fatal("re-join into the same cluster should not segment")
	}
	if len(tr.Records()) != 1 {
		t.Fatalf("expected 1 record, got %d", len(tr.Records()))
	}
}

func TestDungeonToDungeonSegments(t *testing.T) {
	tr := New()
	tr.OnZone("@RANDOMDUNGEON@A", at(0), nil)
	tr.AddFame(10)
	// Move straight into a different dungeon: close the first, open the second.
	if !tr.OnZone("@RANDOMDUNGEON@B", at(30), nil) {
		t.Fatal("dungeon-to-dungeon should report change")
	}
	recs := tr.Records()
	if len(recs) != 2 {
		t.Fatalf("expected 2 records, got %d", len(recs))
	}
	if recs[0].Open() || !recs[1].Open() {
		t.Fatalf("first should be closed, second open: %+v", recs)
	}
	if recs[0].Fame != 10 || recs[1].Fame != 0 {
		t.Fatalf("fame mis-segmented: %+v", recs)
	}
}

func TestRenameAndTier(t *testing.T) {
	tr := New()
	tr.OnZone("@RANDOMDUNGEON@A", at(0), nil)
	id := tr.Records()[0].ID
	if !tr.Rename(id, "Morgana Run") || !tr.SetTier(id, "T6") {
		t.Fatal("rename/tier should succeed for a known id")
	}
	if tr.Rename(999, "x") || tr.SetTier(999, "T8") {
		t.Fatal("rename/tier should fail for unknown id")
	}
	r := tr.Records()[0]
	if r.Name != "Morgana Run" || r.Tier != "T6" {
		t.Fatalf("rename/tier not applied: %+v", r)
	}
}

func TestReset(t *testing.T) {
	tr := New()
	tr.OnZone("@RANDOMDUNGEON@A", at(0), nil)
	tr.Reset()
	if len(tr.Records()) != 0 {
		t.Fatal("reset should clear records")
	}
	// Still usable, ids restart.
	tr.OnZone("@RANDOMDUNGEON@B", at(1), nil)
	if tr.Records()[0].ID != 1 {
		t.Fatalf("id should restart at 1, got %d", tr.Records()[0].ID)
	}
}
