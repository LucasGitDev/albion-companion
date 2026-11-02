// Package reconcile compares the items a group looted (IN, from the passive loot
// log) against what was deposited into the guild vault (OUT, from the guild's
// official account/vault export), per player and item, to surface loot that was
// never handed in — the "who pocketed the drops" check.
//
// Why an import and not packets: the passive client only ever sees the LOCAL
// player's own deposits, never other members'. The OUT side for a whole group
// can only come from the guild's own export, so it is imported here.
package reconcile

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/LucasGitDev/albion-companion/internal/loot"
)

// Deposit is one row of the guild account/vault export. Quantity is signed: a
// positive value is a deposit into the vault, a negative value a withdrawal.
type Deposit struct {
	At       time.Time
	Player   string
	ItemName string // localized item name, exactly as exported (e.g. PT-BR)
	Enchant  int
	Quality  int
	Quantity int64
}

// exportLayout is the timestamp format used by the guild export: MM/DD/YYYY and
// a 24-hour clock, e.g. "11/02/2026 19:43:52".
const exportLayout = "01/02/2006 15:04:05"

// ParseGuildExport reads the guild account/vault export: tab-separated, each
// field wrapped in double quotes, with a header row
//
//	Data  Jogador  Item  Encantamento  Qualidade  Quantidade
//
// Rows whose timestamp is before `from` are skipped (a zero `from` keeps every
// row), which is how a caller limits reconciliation to a single run. The header
// row and blank lines are ignored. A malformed data row is an error rather than
// silently dropped, so a bad export never looks like missing deposits.
func ParseGuildExport(r io.Reader, from time.Time) ([]Deposit, error) {
	cr := csv.NewReader(r)
	cr.Comma = '\t'
	cr.FieldsPerRecord = -1 // tolerate trailing empty columns
	cr.LazyQuotes = true

	var out []Deposit
	for line := 1; ; line++ {
		rec, err := cr.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", line, err)
		}
		if len(rec) == 0 {
			continue
		}
		first := strings.TrimSpace(rec[0])
		if first == "" {
			continue
		}
		// Header row (or any non-data row): the first column is not a timestamp.
		at, err := time.Parse(exportLayout, first)
		if err != nil {
			if line == 1 || strings.EqualFold(first, "data") {
				continue // header
			}
			return nil, fmt.Errorf("line %d: bad timestamp %q: %w", line, first, err)
		}
		if len(rec) < 6 {
			return nil, fmt.Errorf("line %d: expected 6 columns, got %d", line, len(rec))
		}
		qty, err := strconv.ParseInt(strings.TrimSpace(rec[5]), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("line %d: bad quantity %q: %w", line, rec[5], err)
		}
		enchant, _ := strconv.Atoi(strings.TrimSpace(rec[3]))
		quality, _ := strconv.Atoi(strings.TrimSpace(rec[4]))
		if !from.IsZero() && at.Before(from) {
			continue
		}
		out = append(out, Deposit{
			At:       at,
			Player:   strings.TrimSpace(rec[1]),
			ItemName: strings.TrimSpace(rec[2]),
			Enchant:  enchant,
			Quality:  quality,
			Quantity: qty,
		})
	}
	return out, nil
}

// Row is the reconciliation of one (player, item) pair. Delta = Looted - NetOut.
// Flagged marks the anti-theft case: the player looted this item (Looted > 0) and
// handed in less than they looted (Delta > 0). A row with Looted == 0 is only a
// vault movement of an item not looted this run (e.g. a withdrawal, or a deposit
// from an earlier stash); it is shown for context but never flagged, so a pure
// withdrawal does not masquerade as undeposited loot.
type Row struct {
	Player  string
	Item    string // localized display name
	Looted  int64  // IN: quantity looted (from the loot log)
	NetOut  int64  // OUT: net quantity deposited (deposits minus withdrawals)
	Delta   int64  // Looted - NetOut
	Flagged bool   // Looted > 0 && Delta > 0: looted more than handed in
}

// Reconcile matches looted items (IN) against net guild deposits (OUT) by
// (player, item). Matching is on the localized item name: the export and
// loot.Entry.ItemLocalName both carry the game's localized name, and the name
// already encodes tier and type. Enchantment and quality are summed over, as the
// passive loot log does not yet record them. Silver loot entries are ignored
// (silver is deposited through a separate path, not the item vault).
//
// Names are matched case-insensitively with internal whitespace collapsed, so a
// stray double space or casing difference between the export and the game's
// localization table does not look like a missing deposit.
func Reconcile(loots []loot.Entry, deposits []Deposit) []Row {
	type key struct{ player, item string }
	looted := map[key]int64{}
	netOut := map[key]int64{}
	display := map[key]struct{ player, item string }{}

	remember := func(k key, player, item string) {
		if _, ok := display[k]; !ok {
			display[k] = struct{ player, item string }{player, item}
		}
	}

	for _, e := range loots {
		if e.IsSilver {
			continue
		}
		name := e.ItemLocalName
		if name == "" {
			name = e.ItemName
		}
		if name == "" {
			name = e.ItemID
		}
		k := key{normalize(e.Looter), normalize(name)}
		looted[k] += e.Quantity
		remember(k, e.Looter, name)
	}
	for _, d := range deposits {
		k := key{normalize(d.Player), normalize(d.ItemName)}
		netOut[k] += d.Quantity
		remember(k, d.Player, d.ItemName)
	}

	rows := make([]Row, 0, len(display))
	for k, disp := range display {
		in, out := looted[k], netOut[k]
		rows = append(rows, Row{
			Player:  disp.player,
			Item:    disp.item,
			Looted:  in,
			NetOut:  out,
			Delta:   in - out,
			Flagged: in > 0 && in-out > 0,
		})
	}
	// Flagged first (worst discrepancy on top), then by player and item for a
	// stable, readable order.
	sort.Slice(rows, func(i, j int) bool {
		a, b := rows[i], rows[j]
		if a.Flagged != b.Flagged {
			return a.Flagged
		}
		if a.Flagged && a.Delta != b.Delta {
			return a.Delta > b.Delta
		}
		if a.Player != b.Player {
			return a.Player < b.Player
		}
		return a.Item < b.Item
	})
	return rows
}

func normalize(s string) string {
	return strings.ToLower(strings.Join(strings.Fields(s), " "))
}
