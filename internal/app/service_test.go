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
	if pv := svc.PartyView(); pv.Me != "Me" || pv.Location != "4300" {
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
