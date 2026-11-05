// Command genworld regenerates internal/world/clusters.tsv from the ao-bin-dumps
// formatted/world.json dump.
//
// world.json is an array of {Index, UniqueName}; UniqueName is the human cluster
// name (e.g. Index "1000" -> "Lymhurst"). We trim it to a compact "id\tname"
// table that internal/world embeds, so cluster-id resolution needs no network at
// runtime. Entries without a numeric/usable id or name are skipped.
//
// Usage:
//
//	go run ./tools/genworld -out internal/world/clusters.tsv
//	go run ./tools/genworld -in /path/to/world.json -out internal/world/clusters.tsv
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

const worldJSONURL = "https://raw.githubusercontent.com/ao-data/ao-bin-dumps/master/formatted/world.json"

type cluster struct {
	Index      string `json:"Index"`
	UniqueName string `json:"UniqueName"`
}

func main() {
	in := flagString("-in", "", os.Args)
	out := flagString("-out", "internal/world/clusters.tsv", os.Args)

	var r io.ReadCloser
	if in != "" {
		f, err := os.Open(in)
		if err != nil {
			fail(err)
		}
		r = f
	} else {
		fmt.Fprintf(os.Stderr, "downloading %s ...\n", worldJSONURL)
		rc, err := fetch(worldJSONURL)
		if err != nil {
			fail(err)
		}
		r = rc
	}
	defer r.Close()

	var clusters []cluster
	if err := json.NewDecoder(r).Decode(&clusters); err != nil {
		fail(fmt.Errorf("decode world.json: %w", err))
	}

	lines := make([]string, 0, len(clusters))
	seen := map[string]bool{}
	for _, c := range clusters {
		id := clean(c.Index)
		name := clean(c.UniqueName)
		if id == "" || name == "" || name == id || seen[id] {
			continue
		}
		seen[id] = true
		lines = append(lines, id+"\t"+name)
	}
	sort.Strings(lines)

	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		fail(err)
	}
	if err := os.WriteFile(out, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		fail(err)
	}
	fmt.Fprintf(os.Stderr, "wrote %d clusters to %s\n", len(lines), out)
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
	fmt.Fprintln(os.Stderr, "genworld:", err)
	os.Exit(1)
}
