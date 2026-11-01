package gather

import (
	"testing"
	"time"
)

func TestTrackerAddAndSnapshot(t *testing.T) {
	tr := New()
	base := time.Unix(1000, 0)
	tr.Add(424, 3, base)                     // wood
	tr.Add(100, 5, base.Add(time.Minute))    // ore
	tr.Add(424, 2, base.Add(2*time.Minute))  // more wood
	tr.Add(100, 0, base.Add(3*time.Minute))  // ignored (amount 0)
	tr.Add(100, -1, base.Add(3*time.Minute)) // ignored (amount < 0)

	s := tr.Snapshot()
	if s.Total != 10 {
		t.Fatalf("Total = %d want 10", s.Total)
	}
	if s.Count != 3 {
		t.Fatalf("Count = %d want 3", s.Count) // 3 valid harvest events
	}
	// sorted by amount desc: wood(5) before ore(5)? both 5 → tie broken by index asc
	if len(s.Items) != 2 {
		t.Fatalf("Items = %d want 2", len(s.Items))
	}
	if s.Items[0].ItemIndex != 100 || s.Items[0].Amount != 5 {
		t.Fatalf("Items[0] = %+v want {100,5}", s.Items[0])
	}
	if s.Items[1].ItemIndex != 424 || s.Items[1].Amount != 5 {
		t.Fatalf("Items[1] = %+v want {424,5}", s.Items[1])
	}
	if !s.Since.Equal(base) {
		t.Fatalf("Since = %v want %v", s.Since, base)
	}
	// 10 units over 2 min = 300/h
	if s.PerHour != 300 {
		t.Fatalf("PerHour = %v want 300", s.PerHour)
	}
}

func TestTrackerEmpty(t *testing.T) {
	s := New().Snapshot()
	if s.Total != 0 || s.Count != 0 || len(s.Items) != 0 || s.PerHour != 0 {
		t.Fatalf("empty snapshot = %+v", s)
	}
	if !s.Since.IsZero() {
		t.Fatalf("Since = %v want zero", s.Since)
	}
}

func TestTrackerReset(t *testing.T) {
	tr := New()
	tr.Add(1, 10, time.Unix(1000, 0))
	tr.Reset()
	s := tr.Snapshot()
	if s.Total != 0 || s.Count != 0 || len(s.Items) != 0 {
		t.Fatalf("after reset = %+v", s)
	}
	// still usable after reset
	tr.Add(2, 5, time.Unix(2000, 0))
	if tr.Snapshot().Total != 5 {
		t.Fatal("tracker unusable after reset")
	}
}
