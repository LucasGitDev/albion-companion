// Command companion-cli runs the engine headless. It is the main dev/debug tool:
// list interfaces, capture live, record/replay sessions and dump decoded messages.
//
//	sudo go run ./cmd/companion-cli -dump                    # live, print every message
//	sudo go run ./cmd/companion-cli -record s.aorec          # live + record
//	go run ./cmd/companion-cli -replay s.aorec -dump -only OtherGrabbedLoot,HealthUpdate
//	go run ./cmd/companion-cli -replay s.aorec -csv loot.csv # export loot
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"math"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/LucasGitDev/albion-companion/internal/capture"
	"github.com/LucasGitDev/albion-companion/internal/combat"
	"github.com/LucasGitDev/albion-companion/internal/engine"
	"github.com/LucasGitDev/albion-companion/internal/itemdb"
	"github.com/LucasGitDev/albion-companion/internal/loot"
	"github.com/LucasGitDev/albion-companion/internal/protocol"
)

func main() {
	var (
		listIfaces = flag.Bool("ifaces", false, "list capture interfaces and exit")
		iface      = flag.String("iface", "", "capture only this interface (default: all physical)")
		pcapFile   = flag.String("pcap", "", "read packets from a .pcap/.pcapng file")
		replay     = flag.String("replay", "", "replay an .aorec recording")
		speed      = flag.Float64("speed", 0, "replay speed (0 = max, 1 = real time)")
		record     = flag.String("record", "", "record captured packets to an .aorec file")
		dump       = flag.Bool("dump", false, "print every decoded message")
		only       = flag.String("only", "", "with -dump: comma-separated message names to print")
		unknown    = flag.Bool("unknown", false, "with -dump: also print messages with unknown codes")
		codesPath  = flag.String("codes", "", "override codes.json")
		itemsDir   = flag.String("items-cache", defaultCacheDir(), "directory for the items.txt cache")
		csvOut     = flag.String("csv", "", "on exit, write loot CSV (ao-loot-logger format) here")
		partyOnly  = flag.Bool("party", true, "party-only filter for loot/DPS")
		every      = flag.Duration("every", 5*time.Second, "summary interval (0 = only at exit)")
	)
	flag.Parse()

	if *listIfaces {
		devs, err := capture.Interfaces()
		exitIf(err)
		for _, d := range devs {
			fmt.Println(d)
		}
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	codes := protocol.DefaultCodes()
	if *codesPath != "" {
		c, err := protocol.LoadCodes(*codesPath)
		exitIf(err)
		codes = c
	}
	fmt.Fprintf(os.Stderr, "codes: %s\n", codes.Source)

	filter := map[string]bool{}
	for _, n := range strings.Split(*only, ",") {
		if n = strings.TrimSpace(n); n != "" {
			filter[n] = true
		}
	}

	cfg := engine.Config{
		Codes:  codes,
		Combat: combat.Options{PartyOnly: *partyOnly},
		OnLoot: func(e loot.Entry) {
			if e.IsSilver {
				fmt.Printf("[loot] %s %s +%.0f silver\n", e.At.Format("15:04:05"), e.Looter, math.Floor(e.Silver))
				return
			}
			fmt.Printf("[loot] %s %s x%d %s (%s) from %s\n", e.At.Format("15:04:05"), e.Looter, e.Quantity, e.ItemName, e.ItemID, e.LootedFrom)
		},
		OnDecodeError: func(name string, err error) {
			fmt.Fprintf(os.Stderr, "[decode-error] %s: %v\n", name, err)
		},
	}
	if *dump {
		cfg.OnMessage = func(m protocol.Message) {
			if m.Name == "" && !*unknown {
				return
			}
			if len(filter) > 0 && !filter[m.Name] {
				return
			}
			fmt.Printf("[%s] %s %d %s %s\n", m.At.Format("15:04:05.000"), m.Kind, m.Code, m.Name, formatParams(m.Params))
		}
	}
	if *record != "" {
		rec, err := capture.NewRecorder(*record)
		exitIf(err)
		defer rec.Close()
		cfg.Recorder = rec
	}

	eng := engine.New(cfg)
	if db, err := itemdb.LoadCached(ctx, *itemsDir, 24*time.Hour); err != nil {
		fmt.Fprintf(os.Stderr, "itemdb: %v (item names unavailable)\n", err)
	} else {
		eng.Loot().SetItems(db)
		fmt.Fprintf(os.Stderr, "itemdb: %d items\n", db.Len())
	}

	var sources []capture.Source
	switch {
	case *replay != "":
		sources = []capture.Source{capture.ReplaySource{Path: *replay, Speed: *speed}}
	case *pcapFile != "":
		sources = []capture.Source{capture.PcapFileSource{Path: *pcapFile}}
	case *iface != "":
		sources = []capture.Source{capture.LiveSource{Device: *iface}}
	default:
		s, err := capture.LiveSources(0)
		exitIf(err)
		sources = s
	}
	for _, s := range sources {
		fmt.Fprintf(os.Stderr, "source: %s\n", s.Name())
	}

	if *every > 0 {
		go func() {
			t := time.NewTicker(*every)
			defer t.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-t.C:
					printSummary(eng, *partyOnly)
				}
			}
		}()
	}

	err := eng.Run(ctx, sources...)
	printSummary(eng, *partyOnly)
	if *csvOut != "" {
		f, ferr := os.Create(*csvOut)
		exitIf(ferr)
		exitIf(eng.Loot().WriteCSV(f, loot.Filter{PartyOnly: *partyOnly}))
		exitIf(f.Close())
		fmt.Fprintf(os.Stderr, "loot csv: %s\n", *csvOut)
	}
	if errors.Is(err, capture.ErrPermission) {
		fmt.Fprintln(os.Stderr, "hint: run with sudo, or grant access to /dev/bpf* (macOS) / CAP_NET_RAW (Linux). See backlog doc: Platforms and capture permissions")
	}
	exitIf(err)
}

func printSummary(eng *engine.Engine, partyOnly bool) {
	st := eng.Stats()
	fmt.Printf("\n== packets=%d msgs=%d unknown=%d decodeErr=%d encrypted=%d zone=%s me=%s\n",
		st.Packets, st.Messages, st.Unknown, st.DecodeErrors, st.Encrypted, eng.State().Location(), eng.State().LocalPlayer().Name)
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "DPS\tPLAYER\tDAMAGE\tHEAL\tTAKEN\tDPS")
	for _, p := range eng.Combat().Snapshot() {
		fmt.Fprintf(w, "\t%s\t%d\t%d\t%d\t%.0f\n", p.Name, p.Damage, p.Heal, p.Taken, p.DPS())
	}
	fmt.Fprintln(w, "LOOT\tPLAYER\tITEMS\tSILVER\tPICKUPS\t")
	for _, p := range eng.Loot().ByPlayer(loot.Filter{PartyOnly: partyOnly}) {
		fmt.Fprintf(w, "\t%s\t%d\t%.0f\t%d\t\n", p.Name, p.Items, math.Floor(p.Silver), p.Pickups)
	}
	w.Flush()

	f := eng.Fame().Snapshot()
	if f.Count > 0 {
		fmt.Printf("FAME  total=%.0f (yellow=%.0f premium=%.0f) %.0f/h over %d events\n", f.Total, f.Yellow, f.Premium, f.PerHour, f.Count)
	}
}

func formatParams(p protocol.Params) string {
	keys := make([]int, 0, len(p))
	for k := range p {
		if k == 252 || k == 253 {
			continue
		}
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	var b strings.Builder
	for _, k := range keys {
		fmt.Fprintf(&b, "%d:%v ", k, p[byte(k)])
	}
	return b.String()
}

func defaultCacheDir() string {
	dir, err := os.UserCacheDir()
	if err != nil {
		return ".cache"
	}
	return filepath.Join(dir, "albion-companion")
}

func exitIf(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
