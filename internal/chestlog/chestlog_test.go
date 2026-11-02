package chestlog

import (
	"testing"
	"time"
)

func TestAddSnapshotReset(t *testing.T) {
	tr := New(0)
	base := time.Unix(1000, 0)
	tr.Add(Entry{At: base, Kind: KindWithdraw, Container: "TREASURE_COFFER_SOLO", ItemID: "T5_SOUL", ItemName: "Soul", Quantity: 2})
	tr.Add(Entry{At: base.Add(time.Minute), Kind: KindWithdraw, Container: "TREASURE_COFFER_SOLO", ItemID: "T4_BAG", ItemName: "Bag", Quantity: 1})
	tr.Add(Entry{At: base.Add(2 * time.Minute), Kind: KindWithdraw, Container: "@MOB_SKELETON", ItemID: "T3_RUNE", ItemName: "Rune", Quantity: 3})

	s := tr.Snapshot(0)
	if s.Total != 3 {
		t.Fatalf("Total = %d want 3", s.Total)
	}
	if s.Withdrawn != 6 {
		t.Fatalf("Withdrawn = %d want 6", s.Withdrawn)
	}
	if s.Deposited != 0 {
		t.Fatalf("Deposited = %d want 0", s.Deposited)
	}
	if s.Containers != 2 {
		t.Fatalf("Containers = %d want 2", s.Containers)
	}
	// newest first
	if s.Entries[0].ItemID != "T3_RUNE" || s.Entries[2].ItemID != "T5_SOUL" {
		t.Fatalf("order wrong: %+v", s.Entries)
	}

	// recent cap
	if r := tr.Snapshot(1); len(r.Entries) != 1 || r.Entries[0].ItemID != "T3_RUNE" {
		t.Fatalf("recent=1 => %+v", r.Entries)
	}

	tr.Reset()
	if z := tr.Snapshot(0); z.Total != 0 || len(z.Entries) != 0 {
		t.Fatalf("after reset = %+v", z)
	}
}

func TestMaxCap(t *testing.T) {
	tr := New(2)
	for i := range 5 {
		tr.Add(Entry{At: time.Unix(int64(i), 0), Kind: KindWithdraw, Quantity: 1})
	}
	s := tr.Snapshot(0)
	if s.Total != 2 {
		t.Fatalf("cap not enforced: Total=%d", s.Total)
	}
	// oldest trimmed: the two kept are the last two added (At 3 and 4), newest first
	if !s.Entries[0].At.Equal(time.Unix(4, 0)) || !s.Entries[1].At.Equal(time.Unix(3, 0)) {
		t.Fatalf("trim kept wrong entries: %+v", s.Entries)
	}
}
