package timeline

import (
	"testing"
	"time"
)

func ev(i int) Event {
	return Event{At: time.Unix(int64(i), 0), Kind: KindLoot, Quantity: int64(i)}
}

func TestRing_OrderNewestFirst(t *testing.T) {
	r := NewRing(10)
	for i := 1; i <= 5; i++ {
		r.Add(ev(i))
	}
	if got := r.Len(); got != 5 {
		t.Fatalf("Len = %d, want 5", got)
	}
	rec := r.Recent(0)
	if len(rec) != 5 {
		t.Fatalf("Recent len = %d, want 5", len(rec))
	}
	// newest first: 5,4,3,2,1
	for i, e := range rec {
		want := int64(5 - i)
		if e.Quantity != want {
			t.Errorf("rec[%d].Quantity = %d, want %d", i, e.Quantity, want)
		}
	}
}

func TestRing_Bounded(t *testing.T) {
	r := NewRing(3)
	for i := 1; i <= 7; i++ {
		r.Add(ev(i))
	}
	if got := r.Len(); got != 3 {
		t.Fatalf("Len = %d, want 3 (bounded)", got)
	}
	rec := r.Recent(0)
	want := []int64{7, 6, 5} // oldest (1..4) evicted, newest first
	if len(rec) != len(want) {
		t.Fatalf("Recent len = %d, want %d", len(rec), len(want))
	}
	for i, w := range want {
		if rec[i].Quantity != w {
			t.Errorf("rec[%d] = %d, want %d", i, rec[i].Quantity, w)
		}
	}
}

func TestRing_RecentLimit(t *testing.T) {
	r := NewRing(10)
	for i := 1; i <= 6; i++ {
		r.Add(ev(i))
	}
	rec := r.Recent(2)
	if len(rec) != 2 || rec[0].Quantity != 6 || rec[1].Quantity != 5 {
		t.Fatalf("Recent(2) = %+v", rec)
	}
}

func TestRing_Reset(t *testing.T) {
	r := NewRing(4)
	for i := 1; i <= 4; i++ {
		r.Add(ev(i))
	}
	r.Reset()
	if r.Len() != 0 {
		t.Fatalf("Len after reset = %d, want 0", r.Len())
	}
	if len(r.Recent(0)) != 0 {
		t.Fatalf("Recent after reset not empty")
	}
	// still usable, order intact after wrap
	for i := 10; i <= 16; i++ {
		r.Add(ev(i))
	}
	rec := r.Recent(0)
	want := []int64{16, 15, 14, 13}
	if len(rec) != len(want) {
		t.Fatalf("Recent len = %d, want %d", len(rec), len(want))
	}
	for i, w := range want {
		if rec[i].Quantity != w {
			t.Errorf("rec[%d] = %d, want %d", i, rec[i].Quantity, w)
		}
	}
}

func TestNewRing_DefaultCap(t *testing.T) {
	r := NewRing(0)
	if len(r.buf) != DefaultCap {
		t.Fatalf("default cap = %d, want %d", len(r.buf), DefaultCap)
	}
}
