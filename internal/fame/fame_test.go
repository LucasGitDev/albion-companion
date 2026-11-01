package fame

import (
	"testing"
	"time"

	"github.com/LucasGitDev/albion-companion/internal/events"
)

// TestReset covers the regression where Reset did "*t = Tracker{}" under its own
// lock, overwriting the mutex and panicking on the deferred Unlock.
func TestReset(t *testing.T) {
	tr := New()
	tr.Add(events.FameUpdate{Base: 100, BonusFactor: 1, TotalPlayerFame: 5000}, time.Unix(1000, 0))

	tr.Reset() // must not panic

	s := tr.Snapshot()
	if s.Total != 0 || s.Base != 0 || s.Count != 0 || s.LatestTotalFame != 0 || !s.Since.IsZero() {
		t.Fatalf("after reset, snapshot not zeroed: %+v", s)
	}

	// Still usable after reset.
	tr.Add(events.FameUpdate{Base: 50, BonusFactor: 1}, time.Unix(2000, 0))
	if got := tr.Snapshot(); got.Base != 50 || got.Count != 1 {
		t.Fatalf("tracker unusable after reset: %+v", got)
	}
}
