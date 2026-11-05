package app

import (
	"bytes"
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/LucasGitDev/albion-companion/internal/capture"
	"github.com/LucasGitDev/albion-companion/internal/combat"
	"github.com/LucasGitDev/albion-companion/internal/engine"
	"github.com/LucasGitDev/albion-companion/internal/photon/photontest"
	"github.com/LucasGitDev/albion-companion/internal/protocol"
)

// sliceSource emits a fixed list of payloads.
type sliceSource [][]byte

func (s sliceSource) Name() string { return "test" }
func (s sliceSource) Run(ctx context.Context, out chan<- capture.Packet) error {
	for i, p := range s {
		out <- capture.Packet{At: time.Unix(int64(1000+i), 0), Payload: p}
	}
	return nil
}

func TestService_CaptureAndViews(t *testing.T) {
	codes := protocol.DefaultCodes()
	ev := func(n string) uint16 { c, _ := codes.EventCode(n); return c }
	join, _ := codes.OperationCode("Join")
	src := sliceSource{
		photontest.Response(join, map[byte]any{0: int32(1), 2: "Me", 8: "4300"}),
		photontest.Event(ev("OtherGrabbedLoot"), map[byte]any{1: "@MOB_X", 2: "Me", 4: int32(10), 5: int32(2)}),
		photontest.Event(ev("HealthUpdate"), map[byte]any{0: int32(99), 2: float32(-500), 6: int32(1)}),
	}

	var mu sync.Mutex
	got := map[string]int{}
	svc := NewService(Options{
		Codes:   codes,
		Sources: func() ([]capture.Source, error) { return []capture.Source{src}, nil },
		Combat:  combat.Options{PartyOnly: true},
		Emit: func(e string, _ any) {
			mu.Lock()
			got[e]++
			mu.Unlock()
		},
	})
	if _, err := svc.StartCapture(); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for svc.Status().Running && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	svc.Flush()

	lv := svc.LootView(10)
	if lv.Total != 1 || lv.Recent[0].Quantity != 2 || lv.Players[0].Name != "Me" {
		t.Fatalf("loot view = %+v", lv)
	}
	cv := svc.CombatView()
	if len(cv.Rows) != 1 || cv.Rows[0].Damage != 500 || cv.Rows[0].Share != 100 {
		t.Fatalf("combat view = %+v", cv)
	}
	if pv := svc.PartyView(); pv.Me != "Me" || pv.Location != "Arthur's Rest" {
		t.Fatalf("party view = %+v", pv)
	}
	mu.Lock()
	if got["loot:new"] != 1 || got["capture:status"] < 2 {
		t.Fatalf("emitted = %v", got)
	}
	mu.Unlock()

	var buf bytes.Buffer
	if err := svc.ExportLootCSV(&buf); err != nil || !strings.Contains(buf.String(), ";Me;") {
		t.Fatalf("csv = %q err=%v", buf.String(), err)
	}
}

// TestService_SessionPersistence covers TASK-13: a captured session is saved on
// stop and survives an app "restart" (a fresh Service over the same directory
// still lists, loads, and exports it).
func TestService_SessionPersistence(t *testing.T) {
	codes := protocol.DefaultCodes()
	ev := func(n string) uint16 { c, _ := codes.EventCode(n); return c }
	join, _ := codes.OperationCode("Join")
	src := sliceSource{
		photontest.Response(join, map[byte]any{0: int32(1), 2: "Me", 8: "4300"}),
		photontest.Event(ev("OtherGrabbedLoot"), map[byte]any{1: "@MOB_X", 2: "Me", 4: int32(10), 5: int32(3)}),
	}
	dir := t.TempDir()
	newSvc := func() *Service {
		return NewService(Options{
			Codes:       codes,
			SessionsDir: dir,
			Sources:     func() ([]capture.Source, error) { return []capture.Source{src}, nil },
		})
	}

	svc := newSvc()
	if _, err := svc.StartCapture(); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for svc.Status().Running && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}

	// Fresh Service = simulated restart; the session must still be there.
	restarted := newSvc()
	metas, err := restarted.HistoryList()
	if err != nil {
		t.Fatal(err)
	}
	if len(metas) != 1 {
		t.Fatalf("expected 1 persisted session, got %d: %+v", len(metas), metas)
	}
	m := metas[0]
	if m.Me != "Me" || m.Zone != "Arthur's Rest" || m.LootCount != 1 {
		t.Fatalf("session meta = %+v", m)
	}

	detail, err := restarted.HistoryLoad(m.ID)
	if err != nil {
		t.Fatal(err)
	}
	if detail.Loot.Total != 1 || len(detail.Loot.Recent) != 1 || detail.Loot.Recent[0].Quantity != 3 {
		t.Fatalf("loaded detail loot = %+v", detail.Loot)
	}

	var buf bytes.Buffer
	if err := restarted.WriteSessionCSV(&buf, m.ID); err != nil || !strings.Contains(buf.String(), ";Me;") {
		t.Fatalf("session csv = %q err=%v", buf.String(), err)
	}

	// Delete removes it.
	if err := restarted.DeleteSession(m.ID); err != nil {
		t.Fatal(err)
	}
	if metas, _ := restarted.HistoryList(); len(metas) != 0 {
		t.Fatalf("session not deleted: %+v", metas)
	}
}

func TestCodesOutdated(t *testing.T) {
	cases := []struct {
		name string
		st   engine.Stats
		want bool
	}{
		{"cold start", engine.Stats{Messages: 100, Unknown: 100}, false},      // below warmup
		{"healthy", engine.Stats{Messages: 10000, Unknown: 0}, false},         // codes match
		{"a few unknown", engine.Stats{Messages: 10000, Unknown: 50}, false},  // 0.5% < threshold
		{"shifted codes", engine.Stats{Messages: 10000, Unknown: 6000}, true}, // 60% unknown
		{"at threshold", engine.Stats{Messages: 1000, Unknown: 50}, true},     // exactly 5%
	}
	for _, c := range cases {
		if got := codesOutdated(c.st); got != c.want {
			t.Errorf("%s: codesOutdated(%+v) = %v, want %v", c.name, c.st, got, c.want)
		}
	}
}
