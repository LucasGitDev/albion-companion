package itemdb

import (
	_ "embed"
	"strings"
)

// localizedTSV is the trimmed PT-BR/EN-US name table generated from
// ao-bin-dumps formatted/items.json by `go run ./tools/genlocalized`.
// Each line is "UNIQUE_NAME\tPT-BR\tEN-US" (base unique name, no @enchant).
// Embedding keeps name resolution fully offline at runtime.
//
//go:embed localized.tsv
var localizedTSV string

type localizedName struct{ pt, en string }

var localizedNames = parseLocalized(localizedTSV)

func parseLocalized(s string) map[string]localizedName {
	m := make(map[string]localizedName, 12000)
	for _, line := range strings.Split(s, "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) != 3 || parts[0] == "" {
			continue
		}
		m[parts[0]] = localizedName{pt: parts[1], en: parts[2]}
	}
	return m
}

// LocalizedName resolves a unique item name to its display name using the
// fallback chain PT-BR -> EN-US -> humanized unique id. The raw unique id is
// never returned verbatim as the name. The @enchant suffix is stripped before
// the lookup because LocalizedNames are keyed by the base unique name.
func LocalizedName(uniqueName string) string {
	base := uniqueName
	if i := strings.IndexByte(base, '@'); i >= 0 {
		base = base[:i]
	}
	if l, ok := localizedNames[base]; ok {
		if l.pt != "" {
			return l.pt
		}
		if l.en != "" {
			return l.en
		}
	}
	return humanize(uniqueName)
}

// humanize turns a raw unique id into a readable fallback label, e.g.
// "T4_2H_TOOL_TRACKING@1" -> "T4 2H Tool Tracking". It is only used when no
// localized name exists, so a player still never sees the raw id as the name.
func humanize(uniqueName string) string {
	s := uniqueName
	if i := strings.IndexByte(s, '@'); i >= 0 {
		s = s[:i]
	}
	s = strings.ReplaceAll(s, "_", " ")
	fields := strings.Fields(s)
	for i, f := range fields {
		if len(f) == 0 {
			continue
		}
		// Keep all-caps tokens/tiers (T4, 2H) as-is; title-case the rest.
		lower := strings.ToLower(f)
		fields[i] = strings.ToUpper(f[:1]) + lower[1:]
	}
	out := strings.Join(fields, " ")
	if out == "" {
		return uniqueName
	}
	return out
}
