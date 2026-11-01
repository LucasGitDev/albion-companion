// Package itemdb resolves the numeric item index sent by the game to the item's
// unique name (T4_BAG@1) and display name, using ao-bin-dumps formatted/items.txt.
package itemdb

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	ItemsURL = "https://raw.githubusercontent.com/ao-data/ao-bin-dumps/master/formatted/items.txt"
	iconURL  = "https://render.albiononline.com/v1/item/%s.png"
)

type Item struct {
	Index      int
	UniqueName string // e.g. T6_ARMOR_CLOTH_KEEPER@1
	Name       string // English display name from items.txt, e.g. Master's Druid Robe
	LocalName  string // localized display name: PT-BR -> EN-US -> humanized id
}

// IconURL returns the official render URL for the item icon.
func (i Item) IconURL() string { return fmt.Sprintf(iconURL, i.UniqueName) }

// IsTrash reports whether the item is trash loot (SA: ItemController.IsTrash).
func (i Item) IsTrash() bool { return strings.Contains(i.UniqueName, "_TRASH") }

type DB struct{ items map[int]Item }

// Parse reads the items.txt format: "   12: UNIQUE_NAME     : Display Name".
func Parse(r io.Reader) (*DB, error) {
	db := &DB{items: map[int]Item{}}
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		idxStr, rest, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		idx, err := strconv.Atoi(strings.TrimSpace(idxStr))
		if err != nil {
			continue
		}
		unique, name, _ := strings.Cut(rest, " : ")
		unique = strings.TrimSpace(unique)
		name = strings.TrimSpace(name)
		if name == "" {
			name = unique
		}
		db.items[idx] = Item{Index: idx, UniqueName: unique, Name: name, LocalName: LocalizedName(unique)}
	}
	if len(db.items) == 0 {
		return nil, fmt.Errorf("itemdb: no items parsed")
	}
	return db, sc.Err()
}

func (db *DB) Resolve(index int) (Item, bool) {
	it, ok := db.items[index]
	return it, ok
}

func (db *DB) Len() int { return len(db.items) }

// LoadCached loads items.txt from cacheDir, downloading it when missing or
// older than maxAge. A stale cache is used if the download fails.
func LoadCached(ctx context.Context, cacheDir string, maxAge time.Duration) (*DB, error) {
	path := filepath.Join(cacheDir, "items.txt")
	if st, err := os.Stat(path); err != nil || time.Since(st.ModTime()) > maxAge {
		if dlErr := download(ctx, path); dlErr != nil && err != nil {
			return nil, fmt.Errorf("itemdb: no cache and download failed: %w", dlErr)
		}
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Parse(f)
}

func download(ctx context.Context, dst string) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ItemsURL, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: %s", ItemsURL, resp.Status)
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	tmp := dst + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, dst)
}
