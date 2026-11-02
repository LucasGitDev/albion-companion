package reconcile

import (
	"strings"
	"testing"
	"time"

	"github.com/LucasGitDev/albion-companion/internal/loot"
)

// sampleExport is a trimmed real guild export (tab-separated, quoted fields):
// one deposit batch at 19:41 and a withdrawal batch at 19:43.
const sampleExport = `"Data"` + "\t" + `"Jogador"` + "\t" + `"Item"` + "\t" + `"Encantamento"` + "\t" + `"Qualidade"` + "\t" + `"Quantidade"` + "\n" +
	`"11/02/2026 19:43:52"` + "\t" + `"1plimplim"` + "\t" + `"Robe de Erudito do Iniciante"` + "\t" + `"0"` + "\t" + `"2"` + "\t" + `"-1"` + "\n" +
	`"11/02/2026 19:41:53"` + "\t" + `"1plimplim"` + "\t" + `"Elmo de Soldado do Novato"` + "\t" + `"0"` + "\t" + `"1"` + "\t" + `"1"` + "\n" +
	`"11/02/2026 19:41:53"` + "\t" + `"1plimplim"` + "\t" + `"Robe de Erudito do Iniciante"` + "\t" + `"0"` + "\t" + `"2"` + "\t" + `"2"` + "\n"

func TestParseGuildExport(t *testing.T) {
	deps, err := ParseGuildExport(strings.NewReader(sampleExport), time.Time{})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(deps) != 3 {
		t.Fatalf("expected 3 rows (header skipped), got %d: %+v", len(deps), deps)
	}
	// Signed quantity: withdrawal is negative, deposits positive.
	if deps[0].Quantity != -1 || deps[0].Player != "1plimplim" || deps[0].ItemName != "Robe de Erudito do Iniciante" {
		t.Fatalf("row0 = %+v", deps[0])
	}
	if deps[1].Quantity != 1 || deps[2].Quantity != 2 {
		t.Fatalf("deposit quantities = %d,%d", deps[1].Quantity, deps[2].Quantity)
	}
	if deps[2].Quality != 2 {
		t.Fatalf("quality = %d, want 2", deps[2].Quality)
	}
}

func TestParseGuildExport_FromFilter(t *testing.T) {
	// Keep only rows at/after 19:42 → drops the 19:41 deposit batch, keeps 19:43.
	from := time.Date(2026, 11, 2, 19, 42, 0, 0, time.UTC)
	deps, err := ParseGuildExport(strings.NewReader(sampleExport), from)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(deps) != 1 || deps[0].Quantity != -1 {
		t.Fatalf("from filter kept %+v", deps)
	}
}

func TestParseGuildExport_BadRow(t *testing.T) {
	bad := `"11/02/2026 19:41:53"` + "\t" + `"p"` + "\t" + `"item"` + "\t" + `"0"` + "\t" + `"1"` + "\t" + `"notanumber"` + "\n"
	if _, err := ParseGuildExport(strings.NewReader(bad), time.Time{}); err == nil {
		t.Fatal("expected error on non-numeric quantity")
	}
}

func TestReconcile(t *testing.T) {
	loots := []loot.Entry{
		{Looter: "1plimplim", ItemLocalName: "Robe de Erudito do Iniciante", Quantity: 3},
		{Looter: "1plimplim", ItemLocalName: "Arco do Novato", Quantity: 1},
		{Looter: "Ladrao", ItemLocalName: "Tocha do Iniciante", Quantity: 2},
		{Looter: "1plimplim", IsSilver: true, ItemID: "SILVER", Quantity: 999}, // ignored
	}
	deps := []Deposit{
		// 1plimplim deposited 2 robes net (deposited 2, net), handed in the bow.
		{Player: "1plimplim", ItemName: "Robe de Erudito do Iniciante", Quantity: 2},
		{Player: "1plimplim", ItemName: "Arco do Novato", Quantity: 1},
		// Ladrao deposited nothing.
	}
	rows := Reconcile(loots, deps)

	get := func(player, item string) (Row, bool) {
		for _, r := range rows {
			if r.Player == player && r.Item == item {
				return r, true
			}
		}
		return Row{}, false
	}

	// Robe: looted 3, deposited 2 → delta 1, flagged.
	if r, ok := get("1plimplim", "Robe de Erudito do Iniciante"); !ok || r.Delta != 1 || !r.Flagged {
		t.Fatalf("robe row = %+v (ok=%v)", r, ok)
	}
	// Bow: looted 1, deposited 1 → delta 0, not flagged.
	if r, _ := get("1plimplim", "Arco do Novato"); r.Delta != 0 || r.Flagged {
		t.Fatalf("bow row = %+v", r)
	}
	// Ladrao: looted 2, deposited 0 → delta 2, flagged.
	if r, ok := get("Ladrao", "Tocha do Iniciante"); !ok || r.Delta != 2 || !r.Flagged {
		t.Fatalf("ladrao row = %+v (ok=%v)", r, ok)
	}
	// Silver loot must not appear.
	if _, ok := get("1plimplim", "Silver"); ok {
		t.Fatal("silver should be excluded from reconciliation")
	}
	// Flagged rows sort first, worst delta on top.
	if !rows[0].Flagged || rows[0].Delta != 2 {
		t.Fatalf("expected worst flagged row first, got %+v", rows[0])
	}
}

func TestReconcile_WithdrawalWithoutLootNotFlagged(t *testing.T) {
	// Player withdrew an item they never looted this run: net out is negative,
	// Looted is 0. Delta is positive, but this is not undeposited loot.
	loots := []loot.Entry{}
	deps := []Deposit{{Player: "A", ItemName: "Capuz", Quantity: -1}}
	rows := Reconcile(loots, deps)
	if len(rows) != 1 {
		t.Fatalf("want 1 row, got %+v", rows)
	}
	if rows[0].Flagged {
		t.Fatalf("pure withdrawal (Looted=0) must not be flagged: %+v", rows[0])
	}
}

func TestReconcile_NameNormalization(t *testing.T) {
	loots := []loot.Entry{{Looter: "A", ItemLocalName: "Arco  do Novato", Quantity: 1}}
	deps := []Deposit{{Player: "a", ItemName: "arco do novato", Quantity: 1}}
	rows := Reconcile(loots, deps)
	if len(rows) != 1 || rows[0].Delta != 0 || rows[0].Flagged {
		t.Fatalf("normalization should match across case/spacing: %+v", rows)
	}
}
