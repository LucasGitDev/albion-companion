package world

import "testing"

func TestName(t *testing.T) {
	// A known royal city resolves (embedded from ao-bin-dumps world.json).
	if got := Name("1000"); got != "Lymhurst" {
		t.Fatalf(`Name("1000") = %q, want "Lymhurst"`, got)
	}
	if !Known("1000") {
		t.Fatal("1000 should be a known cluster")
	}
	// Lymhurst Portal is a distinct cluster from the city.
	if got := Name("1301"); got != "Lymhurst Portal" {
		t.Fatalf(`Name("1301") = %q, want "Lymhurst Portal"`, got)
	}
	// Unknown ids and instanced locations pass through unchanged.
	for _, id := range []string{"@RANDOMDUNGEON@abc-123", "@ISLAND@xyz", "zzz-not-a-cluster", ""} {
		if got := Name(id); got != id {
			t.Fatalf("Name(%q) = %q, want passthrough", id, got)
		}
		if id != "" && Known(id) {
			t.Fatalf("%q should not be known", id)
		}
	}
}
