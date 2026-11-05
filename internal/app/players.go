package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/LucasGitDev/albion-companion/internal/loot"
)

// knownPlayers is an accumulating registry of every player nick observed across
// sessions, persisted to disk. It powers the Loot Guard roster typeahead so the
// user can pick a nick they've seen before instead of retyping it. Case is
// folded for de-dup; the first-seen casing is kept for display.
type knownPlayers struct {
	mu  sync.Mutex
	set map[string]string // lower-case key -> display name
	dir string            // persistence dir ("" = memory only)
}

func newKnownPlayers(dir string) *knownPlayers {
	k := &knownPlayers{set: map[string]string{}, dir: dir}
	k.load()
	return k
}

// add records names, returning true if any were new. Blank names are ignored.
func (k *knownPlayers) add(names ...string) bool {
	k.mu.Lock()
	changed := false
	for _, n := range names {
		n = strings.TrimSpace(n)
		if n == "" {
			continue
		}
		key := strings.ToLower(n)
		if _, ok := k.set[key]; !ok {
			k.set[key] = n
			changed = true
		}
	}
	dir := k.dir
	k.mu.Unlock()
	if changed && dir != "" {
		_ = k.save()
	}
	return changed
}

// list returns the known nicks sorted case-insensitively.
func (k *knownPlayers) list() []string {
	k.mu.Lock()
	out := make([]string, 0, len(k.set))
	for _, v := range k.set {
		out = append(out, v)
	}
	k.mu.Unlock()
	sort.Slice(out, func(i, j int) bool { return strings.ToLower(out[i]) < strings.ToLower(out[j]) })
	return out
}

func knownPlayersFile(dir string) string { return filepath.Join(dir, "known-players.json") }

func (k *knownPlayers) load() {
	if k.dir == "" {
		return
	}
	b, err := os.ReadFile(knownPlayersFile(k.dir))
	if err != nil {
		return
	}
	var names []string
	if json.Unmarshal(b, &names) != nil {
		return
	}
	for _, n := range names {
		if n = strings.TrimSpace(n); n != "" {
			k.set[strings.ToLower(n)] = n
		}
	}
}

func (k *knownPlayers) save() error {
	k.mu.Lock()
	names := make([]string, 0, len(k.set))
	for _, v := range k.set {
		names = append(names, v)
	}
	dir := k.dir
	k.mu.Unlock()
	if dir == "" {
		return nil
	}
	sort.Slice(names, func(i, j int) bool { return strings.ToLower(names[i]) < strings.ToLower(names[j]) })
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(names, "", "  ")
	if err != nil {
		return err
	}
	tmp := knownPlayersFile(dir) + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, knownPlayersFile(dir))
}

// harvestKnownPlayers folds the players currently visible in the loot, combat,
// and party views into the registry. Called lazily so the set keeps growing as
// a session plays out, without wiring a hook into every event path.
func (s *Service) harvestKnownPlayers() {
	names := make([]string, 0, 32)
	for _, p := range s.eng.Loot().ByPlayer(loot.Filter{IncludeTrash: true}) {
		names = append(names, p.Name)
	}
	for _, p := range s.eng.Combat().Snapshot() {
		names = append(names, p.Name)
	}
	if st := s.eng.State(); st != nil {
		if me := st.LocalPlayer().Name; me != "" {
			names = append(names, me)
		}
		for _, m := range st.PartyMembers() {
			names = append(names, m.Name)
		}
	}
	s.known.add(names...)
}

// KnownPlayers returns every player nick observed across sessions (for the Loot
// Guard roster typeahead). It harvests the live views first so the list reflects
// the current session too.
func (s *Service) KnownPlayers() []string {
	s.harvestKnownPlayers()
	return s.known.list()
}
