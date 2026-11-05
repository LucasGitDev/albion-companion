// Package world resolves Albion cluster/zone ids to their display names
// (e.g. "1000" -> "Lymhurst"), from a table generated off the ao-bin-dumps
// formatted/world.json dump (see tools/genworld). Lookups need no network.
package world

import (
	_ "embed"
	"strings"
)

//go:embed clusters.tsv
var clustersTSV string

var names = parse(clustersTSV)

func parse(tsv string) map[string]string {
	m := make(map[string]string, 1300)
	for _, line := range strings.Split(tsv, "\n") {
		id, name, ok := strings.Cut(line, "\t")
		if !ok {
			continue
		}
		if id = strings.TrimSpace(id); id != "" {
			m[id] = strings.TrimSpace(name)
		}
	}
	return m
}

// Name returns the display name for a cluster id. Unknown ids and non-cluster
// location strings (instanced zones like "@RANDOMDUNGEON@<guid>" or already-
// resolved names) are returned unchanged, so callers can resolve blindly.
func Name(id string) string {
	if id == "" {
		return ""
	}
	if n, ok := names[id]; ok {
		return n
	}
	return id
}

// Known reports whether the id is a known cluster.
func Known(id string) bool {
	_, ok := names[id]
	return ok
}
