// Command gencodes regenerates internal/protocol/codes.json from the C# enums
// of AlbionOnline-StatisticsAnalysis (EventCodes.cs / OperationCodes.cs).
//
// Albion patches shift event/operation codes. StatisticsAnalysis is usually the
// first project to update them, so we treat it as the upstream source of truth.
//
// Usage:
//
//	go run ./tools/gencodes -sa refs/StatisticsAnalysis -out internal/protocol/codes.json
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var memberRe = regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_]*)\s*(?:=\s*(-?\d+))?\s*,?$`)

func main() {
	saDir := flag.String("sa", "refs/StatisticsAnalysis", "path to a clone of Triky313/AlbionOnline-StatisticsAnalysis")
	out := flag.String("out", "internal/protocol/codes.json", "output JSON file")
	flag.Parse()

	netDir := filepath.Join(*saDir, "src", "StatisticsAnalysisTool", "Network")
	events, err := parseEnum(filepath.Join(netDir, "EventCodes.cs"))
	if err != nil {
		fail(err)
	}
	ops, err := parseEnum(filepath.Join(netDir, "OperationCodes.cs"))
	if err != nil {
		fail(err)
	}

	doc := map[string]any{
		"source":       "Triky313/AlbionOnline-StatisticsAnalysis@" + gitRev(*saDir),
		"generated_at": time.Now().UTC().Format(time.RFC3339),
		"events":       events,
		"operations":   ops,
	}
	b, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		fail(err)
	}
	if err := os.WriteFile(*out, append(b, '\n'), 0o644); err != nil {
		fail(err)
	}
	fmt.Printf("wrote %s (%d events, %d operations)\n", *out, len(events), len(ops))
}

// parseEnum reads the first `enum X { ... }` block of a C# file and returns
// name -> value, following C# implicit numbering rules.
func parseEnum(path string) (map[string]int, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	result := map[string]int{}
	inside, next := false, 0
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if i := strings.Index(line, "//"); i >= 0 {
			line = line[:i]
		}
		line = strings.TrimSpace(line)
		if !inside {
			if strings.Contains(line, "{") {
				inside = true
			}
			continue
		}
		if strings.HasPrefix(line, "}") {
			break
		}
		if line == "" || line == "{" {
			continue
		}
		m := memberRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		if m[2] != "" {
			v, err := strconv.Atoi(m[2])
			if err != nil {
				return nil, fmt.Errorf("%s: %w", line, err)
			}
			next = v
		}
		result[m[1]] = next
		next++
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("no enum members found in %s", path)
	}
	return result, sc.Err()
}

func gitRev(dir string) string {
	out, err := exec.Command("git", "-C", dir, "rev-parse", "--short", "HEAD").Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "gencodes:", err)
	os.Exit(1)
}
