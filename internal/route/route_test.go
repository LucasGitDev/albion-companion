package route

import (
	"testing"
	"time"
)

func TestRouteGraph(t *testing.T) {
	tr := New(0)
	base := time.Unix(1000, 0)
	// A → B → A → C, with a re-join of C (ignored), then C → B.
	seq := []struct {
		zone string
		dt   time.Duration
	}{
		{"Lymhurst", 0},
		{"Avalon-Road-1", time.Minute},
		{"Lymhurst", 2 * time.Minute},
		{"Avalon-Road-2", 3 * time.Minute},
		{"Avalon-Road-2", 3*time.Minute + 10*time.Second}, // re-join: ignored
		{"Avalon-Road-1", 4 * time.Minute},
	}
	for _, s := range seq {
		tr.Visit(s.zone, base.Add(s.dt))
	}
	tr.Visit("", base) // empty ignored

	s := tr.Snapshot(0)
	if s.TotalVisits != 5 { // the re-join and the empty are not counted
		t.Fatalf("TotalVisits = %d want 5", s.TotalVisits)
	}
	// Path order preserved.
	wantPath := []string{"Lymhurst", "Avalon-Road-1", "Lymhurst", "Avalon-Road-2", "Avalon-Road-1"}
	if len(s.Path) != len(wantPath) {
		t.Fatalf("path len = %d want %d (%v)", len(s.Path), len(wantPath), s.Path)
	}
	for i, z := range wantPath {
		if s.Path[i].Zone != z {
			t.Fatalf("path[%d] = %q want %q", i, s.Path[i].Zone, z)
		}
	}
	// Nodes: Lymhurst visited twice (top), Avalon-Road-1 twice, Avalon-Road-2 once.
	byZone := map[string]Node{}
	for _, n := range s.Nodes {
		byZone[n.Zone] = n
	}
	if byZone["Lymhurst"].Visits != 2 || byZone["Avalon-Road-1"].Visits != 2 || byZone["Avalon-Road-2"].Visits != 1 {
		t.Fatalf("node visits wrong: %+v", s.Nodes)
	}
	if len(s.Nodes) != 3 {
		t.Fatalf("nodes = %d want 3 (Avalon roads are distinct)", len(s.Nodes))
	}
	// Edges: L→A1, A1→L, L→A2, A2→A1 (each once). No self-loop from the re-join.
	edgeKey := map[string]int{}
	for _, e := range s.Edges {
		if e.From == e.To {
			t.Fatalf("self-loop edge: %+v", e)
		}
		edgeKey[e.From+"->"+e.To] = e.Count
	}
	for _, k := range []string{"Lymhurst->Avalon-Road-1", "Avalon-Road-1->Lymhurst", "Lymhurst->Avalon-Road-2", "Avalon-Road-2->Avalon-Road-1"} {
		if edgeKey[k] != 1 {
			t.Fatalf("edge %q = %d want 1 (all: %+v)", k, edgeKey[k], s.Edges)
		}
	}
	if len(s.Edges) != 4 {
		t.Fatalf("edges = %d want 4", len(s.Edges))
	}

	// recent cap on path
	if r := tr.Snapshot(2); len(r.Path) != 2 || r.Path[1].Zone != "Avalon-Road-1" {
		t.Fatalf("recent=2 path = %+v", r.Path)
	}
}

func TestResetAndRepeatEdge(t *testing.T) {
	tr := New(0)
	base := time.Unix(0, 0)
	// Traverse A→B twice to bump the edge count.
	tr.Visit("A", base)
	tr.Visit("B", base.Add(time.Minute))
	tr.Visit("A", base.Add(2*time.Minute))
	tr.Visit("B", base.Add(3*time.Minute))
	s := tr.Snapshot(0)
	var ab *Edge
	for i := range s.Edges {
		if s.Edges[i].From == "A" && s.Edges[i].To == "B" {
			ab = &s.Edges[i]
		}
	}
	if ab == nil || ab.Count != 2 {
		t.Fatalf("A->B count = %+v want 2", ab)
	}

	tr.Reset()
	if z := tr.Snapshot(0); z.TotalVisits != 0 || len(z.Nodes) != 0 || len(z.Edges) != 0 {
		t.Fatalf("after reset = %+v", z)
	}
	// Usable after reset; last is cleared so the first visit starts fresh.
	tr.Visit("X", base)
	tr.Visit("Y", base.Add(time.Minute))
	if s := tr.Snapshot(0); s.TotalVisits != 2 || len(s.Edges) != 1 {
		t.Fatalf("after reset reuse = %+v", s)
	}
}
