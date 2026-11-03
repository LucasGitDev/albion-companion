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

func TestTimelineNewestFirst(t *testing.T) {
	tr := New()
	base := time.Unix(2000, 0)
	tr.Add(424, 3, base)                    // oldest
	tr.Add(100, 5, base.Add(time.Minute))   // middle
	tr.Add(424, 1, base.Add(2*time.Minute)) // newest
	tr.Add(100, 0, base.Add(3*time.Minute)) // ignored (amount 0)

	s := tr.Snapshot()
	if len(s.Timeline) != 3 {
		t.Fatalf("Timeline len = %d want 3 (zero-amount harvest excluded)", len(s.Timeline))
	}
	// Newest first.
	if s.Timeline[0].ItemIndex != 424 || s.Timeline[0].Amount != 1 {
		t.Fatalf("Timeline[0] = %+v want newest {424,1}", s.Timeline[0])
	}
	if !s.Timeline[0].At.After(s.Timeline[1].At) || !s.Timeline[1].At.After(s.Timeline[2].At) {
		t.Fatalf("timeline not strictly newest-first: %v", s.Timeline)
	}
	if s.Timeline[2].Amount != 3 { // oldest
		t.Fatalf("Timeline[2] = %+v want oldest amount 3", s.Timeline[2])
	}
}

func TestTimelineBounded(t *testing.T) {
	tr := New()
	base := time.Unix(3000, 0)
	const n = maxTimeline + 50
	for i := 0; i < n; i++ {
		tr.Add(1, 1, base.Add(time.Duration(i)*time.Second))
	}
	s := tr.Snapshot()
	if len(s.Timeline) != maxTimeline {
		t.Fatalf("Timeline len = %d want cap %d", len(s.Timeline), maxTimeline)
	}
	// Totals are unaffected by the timeline bound.
	if s.Total != int64(n) || s.Count != n {
		t.Fatalf("totals Total=%d Count=%d want %d/%d", s.Total, s.Count, n, n)
	}
	// Newest retained, oldest dropped: the newest event is the last added.
	wantNewest := base.Add(time.Duration(n-1) * time.Second)
	if !s.Timeline[0].At.Equal(wantNewest) {
		t.Fatalf("Timeline[0].At = %v want newest %v", s.Timeline[0].At, wantNewest)
	}
}

func TestResetClearsTimeline(t *testing.T) {
	tr := New()
	tr.Add(1, 2, time.Unix(4000, 0))
	tr.Reset()
	s := tr.Snapshot()
	if len(s.Timeline) != 0 || s.Total != 0 {
		t.Fatalf("after reset: timeline=%d total=%d, want empty", len(s.Timeline), s.Total)
	}
}
