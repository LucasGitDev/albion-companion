package rates

import (
	"testing"
	"time"
)

func TestTrackerTotalsAndPerHour(t *testing.T) {
	tr := New()
	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	// 30 minutes of accrual → per-hour is double the total.
	tr.AddSilver(1000, base)
	tr.AddSilver(500, base.Add(30*time.Minute))
	tr.AddRespec(2, base)
	tr.AddRespec(4, base.Add(30*time.Minute))
	tr.AddMight(10, base)
	tr.AddMight(20, base.Add(30*time.Minute))
	tr.AddFavor(100, base)
	tr.AddFavor(200, base.Add(30*time.Minute))

	s := tr.Snapshot()
	if s.Silver.Total != 1500 || s.Silver.Count != 2 {
		t.Fatalf("silver total %+v", s.Silver)
	}
	if s.Silver.PerHour != 3000 {
		t.Fatalf("silver perHour = %v want 3000", s.Silver.PerHour)
	}
	if s.Respec.Total != 6 || s.Respec.PerHour != 12 {
		t.Fatalf("respec %+v", s.Respec)
	}
	if s.Might.Total != 30 || s.Might.PerHour != 60 {
		t.Fatalf("might %+v", s.Might)
	}
	if s.Favor.Total != 300 || s.Favor.PerHour != 600 {
		t.Fatalf("favor %+v", s.Favor)
	}
	if !s.Silver.Since.Equal(base) {
		t.Fatalf("since = %v want %v", s.Silver.Since, base)
	}
}

func TestTrackerSingleSampleNoRate(t *testing.T) {
	tr := New()
	tr.AddSilver(1000, time.Now())
	if s := tr.Snapshot(); s.Silver.PerHour != 0 {
		t.Fatalf("perHour with one sample = %v want 0", s.Silver.PerHour)
	}
}

func TestTrackerReset(t *testing.T) {
	tr := New()
	now := time.Now()
	tr.AddSilver(1000, now)
	tr.AddMight(5, now)
	tr.AddFavor(50, now)
	tr.Reset()
	s := tr.Snapshot()
	if s.Silver.Total != 0 || s.Might.Total != 0 || s.Favor.Total != 0 || s.Silver.Count != 0 {
		t.Fatalf("reset left data: %+v", s)
	}
}
