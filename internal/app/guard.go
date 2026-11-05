package app

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/LucasGitDev/albion-companion/internal/loot"
)

// Loot Guard is a forensic, anti-theft view for runs where one or more players
// are the designated loot collectors (a scout outside the party, say). Party
// filtering fails for that case — a scout sits outside the party so every squad
// member reads inParty=false — so the roster is explicit instead.
//
// Attribution comes only from the broadcast OtherGrabbedLoot event (looter +
// item + source + timestamp), which is what feeds loot.Tracker. Dungeon
// treasure-chest "take-all" is a private request and is NOT observable for
// other players (a protocol limit, same as StatisticsAnalysis). The view says
// so and never implies chest takes are tracked.

// GuardConfig is the user's anti-theft roster. Persisted to disk so it survives
// restarts.
type GuardConfig struct {
	// Collectors are the players allowed to loot (grabs by anyone else are
	// flagged). Matched case-insensitively.
	Collectors []string `json:"collectors"`
	// ScoutMe treats the local player as a collector without having to type
	// their own name into the roster.
	ScoutMe bool `json:"scoutMe"`
	// IncludeSilver also flags silver pickups. Off by default: silver pickup is
	// routine (auto-distributed off corpses), not theft.
	IncludeSilver bool `json:"includeSilver"`
}

// GuardGrab is one flagged pickup by a non-collector.
type GuardGrab struct {
	At          time.Time `json:"at"`
	Looter      string    `json:"looter"`
	LooterGuild string    `json:"looterGuild"`
	ItemID      string    `json:"itemId"`
	ItemName    string    `json:"itemName"` // localized display name
	Quantity    int64     `json:"quantity"`
	Source      string    `json:"source"` // lootedFrom (mob / body)
	FromMob     bool      `json:"fromMob"`
	IsSilver    bool      `json:"isSilver"`
	Silver      float64   `json:"silver"`
	Location    string    `json:"location"`
}

// GuardItemLine is a per-item roll-up inside a suspect.
type GuardItemLine struct {
	ItemID   string `json:"itemId"`
	ItemName string `json:"itemName"`
	Quantity int64  `json:"quantity"`
}

// GuardSuspect groups a non-collector's flagged activity.
type GuardSuspect struct {
	Name    string          `json:"name"`
	Guild   string          `json:"guild"`
	Grabs   int             `json:"grabs"`   // number of flagged pickups
	ItemQty int64           `json:"itemQty"` // total item quantity grabbed
	Silver  float64         `json:"silver"`  // total silver grabbed (when included)
	First   time.Time       `json:"first"`
	Last    time.Time       `json:"last"`
	Items   []GuardItemLine `json:"items"` // item roll-up, qty desc
}

// GuardView is the forensic feed plus roster/config and the protocol boundary.
type GuardView struct {
	Config         GuardConfig    `json:"config"`
	Me             string         `json:"me"`           // local player (resolved live), may be empty before capture
	Suspects       []GuardSuspect `json:"suspects"`     // flagged players, most grabs first
	Feed           []GuardGrab    `json:"feed"`         // flagged grabs, newest first, capped
	FlaggedGrabs   int            `json:"flaggedGrabs"` // total flagged pickups
	FlaggedItemQty int64          `json:"flaggedItemQty"`
}

// normCollectors returns the effective collector set (lower-cased, trimmed),
// including the local player when ScoutMe is on.
func (c GuardConfig) collectorSet(me string) map[string]bool {
	set := make(map[string]bool, len(c.Collectors)+1)
	for _, n := range c.Collectors {
		if n = strings.TrimSpace(n); n != "" {
			set[strings.ToLower(n)] = true
		}
	}
	if c.ScoutMe {
		if me = strings.TrimSpace(me); me != "" {
			set[strings.ToLower(me)] = true
		}
	}
	return set
}

// LootGuardConfig returns the current anti-theft roster.
func (s *Service) LootGuardConfig() GuardConfig {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.guardCfg
}

// SetLootGuardConfig replaces the roster and persists it. Nudges the guard view
// to re-pull.
func (s *Service) SetLootGuardConfig(cfg GuardConfig) {
	// Normalize: trim, drop blanks, de-dup (case-insensitive, keep first form).
	seen := map[string]bool{}
	clean := make([]string, 0, len(cfg.Collectors))
	for _, n := range cfg.Collectors {
		n = strings.TrimSpace(n)
		if n == "" {
			continue
		}
		k := strings.ToLower(n)
		if seen[k] {
			continue
		}
		seen[k] = true
		clean = append(clean, n)
	}
	cfg.Collectors = clean

	s.mu.Lock()
	s.guardCfg = cfg
	dir := s.opts.ConfigDir
	s.mu.Unlock()
	if dir != "" {
		_ = saveGuardConfig(dir, cfg)
	}
	s.opts.Emit("guard:changed", nil)
}

// LootGuardView builds the forensic feed from the current loot session.
func (s *Service) LootGuardView(recent int) GuardView {
	s.mu.Lock()
	cfg := s.guardCfg
	s.mu.Unlock()

	me := ""
	if st := s.eng.State(); st != nil {
		me = st.LocalPlayer().Name
	}
	collectors := cfg.collectorSet(me)

	// Unfiltered: party filter is meaningless here (scout is outside the party).
	all := s.eng.Loot().Entries(loot.Filter{IncludeTrash: true})

	type agg struct {
		s     GuardSuspect
		items map[string]*GuardItemLine
	}
	byPlayer := map[string]*agg{}
	feed := make([]GuardGrab, 0, len(all))

	for _, e := range all {
		if e.IsSilver && !cfg.IncludeSilver {
			continue
		}
		if collectors[strings.ToLower(strings.TrimSpace(e.Looter))] {
			continue
		}
		g := GuardGrab{
			At: e.At, Looter: e.Looter, LooterGuild: e.LooterGuild,
			ItemID: e.ItemID, ItemName: displayName(e), Quantity: e.Quantity,
			Source: e.LootedFrom, FromMob: e.FromMob,
			IsSilver: e.IsSilver, Silver: e.Silver, Location: e.Location,
		}
		feed = append(feed, g)

		a := byPlayer[e.Looter]
		if a == nil {
			a = &agg{s: GuardSuspect{Name: e.Looter, Guild: e.LooterGuild, First: e.At, Last: e.At}, items: map[string]*GuardItemLine{}}
			byPlayer[e.Looter] = a
		}
		a.s.Grabs++
		if e.At.Before(a.s.First) {
			a.s.First = e.At
		}
		if e.At.After(a.s.Last) {
			a.s.Last = e.At
		}
		if e.LooterGuild != "" {
			a.s.Guild = e.LooterGuild
		}
		if e.IsSilver {
			a.s.Silver += e.Silver
		} else {
			a.s.ItemQty += e.Quantity
			key := e.ItemID
			if key == "" {
				key = displayName(e)
			}
			li := a.items[key]
			if li == nil {
				li = &GuardItemLine{ItemID: e.ItemID, ItemName: displayName(e)}
				a.items[key] = li
			}
			li.Quantity += e.Quantity
		}
	}

	// Build suspects, newest-grab players implicit; sort by grabs desc then qty.
	suspects := make([]GuardSuspect, 0, len(byPlayer))
	var totalGrabs int
	var totalQty int64
	for _, a := range byPlayer {
		lines := make([]GuardItemLine, 0, len(a.items))
		for _, li := range a.items {
			lines = append(lines, *li)
		}
		sort.Slice(lines, func(i, j int) bool {
			if lines[i].Quantity != lines[j].Quantity {
				return lines[i].Quantity > lines[j].Quantity
			}
			return lines[i].ItemName < lines[j].ItemName
		})
		a.s.Items = lines
		suspects = append(suspects, a.s)
		totalGrabs += a.s.Grabs
		totalQty += a.s.ItemQty
	}
	sort.Slice(suspects, func(i, j int) bool {
		if suspects[i].ItemQty != suspects[j].ItemQty {
			return suspects[i].ItemQty > suspects[j].ItemQty
		}
		if suspects[i].Grabs != suspects[j].Grabs {
			return suspects[i].Grabs > suspects[j].Grabs
		}
		return suspects[i].Name < suspects[j].Name
	})

	// Feed newest first, capped.
	sort.Slice(feed, func(i, j int) bool { return feed[i].At.After(feed[j].At) })
	if recent > 0 && recent < len(feed) {
		feed = feed[:recent]
	}

	return GuardView{
		Config: cfg, Me: me, Suspects: suspects, Feed: feed,
		FlaggedGrabs: totalGrabs, FlaggedItemQty: totalQty,
	}
}

// ExportLootGuardCSV writes the full flagged feed (uncapped) as CSV.
func (s *Service) ExportLootGuardCSV(w io.Writer) error {
	v := s.LootGuardView(0)
	cw := csv.NewWriter(w)
	defer cw.Flush()
	if err := cw.Write([]string{"time", "looter", "guild", "item_id", "item_name", "quantity", "is_silver", "silver", "source", "from_mob", "location"}); err != nil {
		return err
	}
	for _, g := range v.Feed {
		rec := []string{
			g.At.Format(time.RFC3339Nano),
			g.Looter, g.LooterGuild, g.ItemID, g.ItemName,
			strconv.FormatInt(g.Quantity, 10),
			strconv.FormatBool(g.IsSilver),
			strconv.FormatFloat(g.Silver, 'f', -1, 64),
			g.Source, strconv.FormatBool(g.FromMob), g.Location,
		}
		if err := cw.Write(rec); err != nil {
			return err
		}
	}
	return cw.Error()
}

// displayName prefers the localized name, falling back to the English name then
// the item id.
func displayName(e loot.Entry) string {
	if e.ItemLocalName != "" {
		return e.ItemLocalName
	}
	if e.ItemName != "" {
		return e.ItemName
	}
	return e.ItemID
}

// --- config persistence -----------------------------------------------------

func guardConfigFile(dir string) string { return filepath.Join(dir, "loot-guard.json") }

// loadGuardConfig reads the roster from disk; a missing/invalid file yields the
// zero config (no collectors), never an error the caller must handle.
func loadGuardConfig(dir string) GuardConfig {
	var cfg GuardConfig
	if dir == "" {
		return cfg
	}
	b, err := os.ReadFile(guardConfigFile(dir))
	if err != nil {
		return cfg
	}
	_ = json.Unmarshal(b, &cfg)
	return cfg
}

// saveGuardConfig writes the roster atomically (temp + rename).
func saveGuardConfig(dir string, cfg GuardConfig) error {
	if dir == "" {
		return nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	tmp := guardConfigFile(dir) + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, guardConfigFile(dir))
}
