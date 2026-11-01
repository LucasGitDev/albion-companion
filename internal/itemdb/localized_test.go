package itemdb

import "testing"

func TestLocalizedName(t *testing.T) {
	// Seed a known table so the test does not depend on the embedded dump.
	localizedNames = map[string]localizedName{
		"T4_BAG":                {pt: "Bolsa do Adepto", en: "Adept's Bag"},
		"T6_SOUL":               {pt: "", en: "Master's Soul"}, // EN-US only -> EN fallback
		"T5_EMPTY":              {pt: "", en: ""},              // present but blank -> humanize
		"T6_ARMOR_CLOTH_KEEPER": {pt: "Robe de Druida do Mestre", en: "Master's Druid Robe"},
	}

	cases := []struct {
		name   string
		unique string
		want   string
	}{
		{"PT-BR hit", "T4_BAG", "Bolsa do Adepto"},
		{"PT-BR hit ignores enchant suffix", "T4_BAG@3", "Bolsa do Adepto"},
		{"EN-US fallback when PT-BR missing", "T6_SOUL", "Master's Soul"},
		{"humanize when entry blank", "T5_EMPTY", "T5 Empty"},
		{"humanize when entry absent", "T2_2H_TOOL_TRACKING@1", "T2 2h Tool Tracking"},
		{"enchant lookup on known base", "T6_ARMOR_CLOTH_KEEPER@2", "Robe de Druida do Mestre"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := LocalizedName(c.unique); got != c.want {
				t.Fatalf("LocalizedName(%q) = %q, want %q", c.unique, got, c.want)
			}
		})
	}
}

// TestLocalizedEmbed is a smoke test over the real embedded table.
func TestLocalizedEmbed(t *testing.T) {
	embedded := parseLocalized(localizedTSV)
	if len(embedded) < 5000 {
		t.Fatalf("embedded localized table too small: %d entries", len(embedded))
	}
	if l, ok := embedded["T4_BAG"]; !ok || l.pt == "" {
		t.Fatalf("expected a PT-BR name for T4_BAG, got %+v (ok=%v)", l, ok)
	}
}
