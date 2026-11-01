// Command genlocalized regenerates internal/itemdb/localized.tsv from the
// ao-bin-dumps formatted/items.json dump.
//
// items.json carries a LocalizedNames block per item keyed by language. We trim
// it to a compact "UNIQUE_NAME\tPT-BR\tEN-US" table (base unique name, no
// @enchant) that internal/itemdb embeds, so localized name resolution needs no
// network at runtime. Items without a PT-BR or EN-US name are skipped.
//
// Usage:
//
//	go run ./tools/genlocalized -out internal/itemdb/localized.tsv
//	go run ./tools/genlocalized -in /path/to/items.json -out internal/itemdb/localized.tsv
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const itemsJSONURL = "https://raw.githubusercontent.com/ao-data/ao-bin-dumps/master/formatted/items.json"

type item struct {
	UniqueName     string            `json:"UniqueName"`
	LocalizedNames map[string]string `json:"LocalizedNames"`
}

func main() {
	in := flagString("-in", "", os.Args)
	out := flagString("-out", "internal/itemdb/localized.tsv", os.Args)

	var r io.ReadCloser
	if in != "" {
		f, err := os.Open(in)
		if err != nil {
			fail(err)
		}
		r = f
	} else {
		fmt.Fprintf(os.Stderr, "downloading %s ...\n", itemsJSONURL)
		rc, err := fetch(itemsJSONURL)
		if err != nil {
			fail(err)
		}
		r = rc
	}
	defer r.Close()

	var items []item
	if err := json.NewDecoder(r).Decode(&items); err != nil {
		fail(fmt.Errorf("decode items.json: %w", err))
	}

	lines := make([]string, 0, len(items))
	seen := map[string]bool{}
	for _, it := range items {
		if it.UniqueName == "" || seen[it.UniqueName] {
			continue
		}
		pt := clean(it.LocalizedNames["PT-BR"])
		en := clean(it.LocalizedNames["EN-US"])
		if pt == "" && en == "" {
			continue
		}
		seen[it.UniqueName] = true
		lines = append(lines, it.UniqueName+"\t"+pt+"\t"+en)
	}
	sort.Strings(lines)

	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		fail(err)
	}
	if err := os.WriteFile(out, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		fail(err)
	}
	fmt.Fprintf(os.Stderr, "wrote %d entries to %s\n", len(lines), out)
}

func clean(s string) string {
	s = strings.ReplaceAll(s, "\t", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return strings.TrimSpace(s)
}

func fetch(url string) (io.ReadCloser, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		cancel()
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		cancel()
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		cancel()
		return nil, fmt.Errorf("GET %s: %s", url, resp.Status)
	}
	return &cancelReader{ReadCloser: resp.Body, cancel: cancel}, nil
}

type cancelReader struct {
	io.ReadCloser
	cancel context.CancelFunc
}

func (c *cancelReader) Close() error {
	err := c.ReadCloser.Close()
	c.cancel()
	return err
}

// flagString is a tiny positional/space-separated flag reader so the tool keeps
// the gencodes-style zero-dependency main without pulling in flag parsing order
// assumptions. Supports "-name value" and "-name=value".
func flagString(name, def string, args []string) string {
	for i := 1; i < len(args); i++ {
		a := args[i]
		if a == name && i+1 < len(args) {
			return args[i+1]
		}
		if strings.HasPrefix(a, name+"=") {
			return strings.TrimPrefix(a, name+"=")
		}
	}
	return def
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "genlocalized:", err)
	os.Exit(1)
}
