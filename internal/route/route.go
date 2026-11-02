// Package route records the ordered sequence of zones the local player enters
// and builds a route graph: nodes are zones (clusters, including Avalon roads —
// each road is its own cluster id on Join), edges are consecutive transitions
// between them with traversal counts. It is fed one Visit per zone change.
package route

import (
	"sort"
	"sync"
	"time"
)

// Visit is one entry into a zone, in order.
type Visit struct {
	Zone string    `json:"zone"`
	At   time.Time `json:"at"`
}

// Node is a distinct zone with how often it was entered.
type Node struct {
	Zone    string    `json:"zone"`
	Visits  int       `json:"visits"`
	FirstAt time.Time `json:"firstAt"`
	LastAt  time.Time `json:"lastAt"`
}

// Edge is a directed transition from one zone to the next, with how many times
// it was traversed.
type Edge struct {
	From   string    `json:"from"`
	To     string    `json:"to"`
	Count  int       `json:"count"`
	LastAt time.Time `json:"lastAt"`
}

// Tracker accumulates the route. Safe for concurrent use.
type Tracker struct {
	mu    sync.RWMutex
	path  []Visit
	nodes map[string]*Node
	edges map[string]*Edge // key "from\x00to"
	last  string           // last distinct zone
	max   int
}

// New keeps up to max ordered visits in the path (<=0 means 5000). Node/edge
// aggregates are never capped.
func New(max int) *Tracker {
	if max <= 0 {
		max = 5000
	}
	return &Tracker{nodes: map[string]*Node{}, edges: map[string]*Edge{}, max: max}
}

// Visit records entering zone at time at. A re-join of the same zone (zone ==
// last) is ignored so we don't create self-loops or phantom revisits. Empty zone
// names are ignored.
func (t *Tracker) Visit(zone string, at time.Time) {
	if zone == "" {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if zone == t.last {
		return
	}
	prev := t.last
	t.last = zone

	t.path = append(t.path, Visit{Zone: zone, At: at})
	if len(t.path) > t.max {
		t.path = t.path[len(t.path)-t.max:]
	}

	n := t.nodes[zone]
	if n == nil {
		n = &Node{Zone: zone, FirstAt: at}
		t.nodes[zone] = n
	}
	n.Visits++
	n.LastAt = at

	if prev != "" {
		key := prev + "\x00" + zone
		e := t.edges[key]
		if e == nil {
			e = &Edge{From: prev, To: zone}
			t.edges[key] = e
		}
		e.Count++
		e.LastAt = at
	}
}

// Snapshot is an immutable view. Path is ordered oldest→newest (capped to recent
// when recent>0); nodes are sorted by visit count desc, edges by traversal desc.
type Snapshot struct {
	Path        []Visit `json:"path"`
	Nodes       []Node  `json:"nodes"`
	Edges       []Edge  `json:"edges"`
	TotalVisits int     `json:"totalVisits"`
}

func (t *Tracker) Snapshot(recent int) Snapshot {
	t.mu.RLock()
	defer t.mu.RUnlock()

	n := len(t.path)
	start := 0
	if recent > 0 && recent < n {
		start = n - recent
	}
	path := make([]Visit, n-start)
	copy(path, t.path[start:])

	nodes := make([]Node, 0, len(t.nodes))
	for _, nd := range t.nodes {
		nodes = append(nodes, *nd)
	}
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].Visits != nodes[j].Visits {
			return nodes[i].Visits > nodes[j].Visits
		}
		return nodes[i].Zone < nodes[j].Zone
	})

	edges := make([]Edge, 0, len(t.edges))
	for _, e := range t.edges {
		edges = append(edges, *e)
	}
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].Count != edges[j].Count {
			return edges[i].Count > edges[j].Count
		}
		if edges[i].From != edges[j].From {
			return edges[i].From < edges[j].From
		}
		return edges[i].To < edges[j].To
	})

	return Snapshot{Path: path, Nodes: nodes, Edges: edges, TotalVisits: len(t.path)}
}

func (t *Tracker) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.path = nil
	t.nodes = map[string]*Node{}
	t.edges = map[string]*Edge{}
	t.last = ""
}
