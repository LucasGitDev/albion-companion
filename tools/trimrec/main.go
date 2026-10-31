// Command trimrec extracts one or more wall-clock time windows from an .aorec
// recording into a smaller .aorec, preserving order. Used to turn a long real
// session into a compact regression fixture (see TASK-2).
//
// Usage:
//
//	go run ./tools/trimrec -in refs/session-1.aorec -out testdata/party-dungeon.aorec \
//	  -keep 16:39:25.000-16:41:10.000,16:55:40.000-16:57:10.000
//
// Window bounds are local-clock HH:MM:SS(.mmm), matched inclusively against each
// packet's timestamp formatted as 15:04:05.000 (same rendering as the CLI dump).
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/LucasGitDev/albion-companion/internal/capture"
)

type window struct{ lo, hi string }

func main() {
	in := flag.String("in", "", "input .aorec")
	out := flag.String("out", "", "output .aorec")
	keep := flag.String("keep", "", "comma-separated HH:MM:SS.mmm-HH:MM:SS.mmm windows to keep")
	flag.Parse()
	if *in == "" || *out == "" || *keep == "" {
		fmt.Fprintln(os.Stderr, "trimrec: -in, -out and -keep are required")
		os.Exit(2)
	}

	var windows []window
	for part := range strings.SplitSeq(*keep, ",") {
		lohi := strings.SplitN(part, "-", 2)
		if len(lohi) != 2 {
			fmt.Fprintf(os.Stderr, "trimrec: bad window %q\n", part)
			os.Exit(2)
		}
		windows = append(windows, window{strings.TrimSpace(lohi[0]), strings.TrimSpace(lohi[1])})
	}

	kept, total, err := trim(*in, *out, windows)
	if err != nil {
		fmt.Fprintf(os.Stderr, "trimrec: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("trimrec: kept %d/%d packets -> %s\n", kept, total, *out)
}

func trim(in, out string, windows []window) (kept, total int, err error) {
	rec, err := capture.NewRecorder(out)
	if err != nil {
		return 0, 0, err
	}
	defer rec.Close()

	src := capture.ReplaySource{Path: in, Speed: 0}
	// ReplaySource streams packets; collect through a channel synchronously.
	ch := make(chan capture.Packet, 256)
	done := make(chan error, 1)
	go func() { done <- src.Run(context.Background(), ch); close(ch) }()

	for p := range ch {
		total++
		ts := p.At.Format("15:04:05.000")
		if inAny(ts, windows) {
			if werr := rec.Write(p); werr != nil {
				return kept, total, werr
			}
			kept++
		}
	}
	return kept, total, <-done
}

func inAny(ts string, windows []window) bool {
	for _, w := range windows {
		if ts >= w.lo && ts <= w.hi {
			return true
		}
	}
	return false
}
