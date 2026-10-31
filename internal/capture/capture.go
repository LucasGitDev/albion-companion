// Package capture produces raw Albion UDP/TCP payloads from live interfaces
// (libpcap/Npcap), offline .pcap files or our own .aorec recordings.
//
// It is strictly passive: it never sends packets nor touches the game process.
package capture

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"
)

// GamePort is the Photon port used by Albion game servers.
const GamePort = 5056

// Packet is one transport payload coming from (or going to) the game server.
type Packet struct {
	At      time.Time
	SrcIP   string
	Payload []byte
}

// Source produces packets until ctx is cancelled or the source is exhausted.
type Source interface {
	Name() string
	Run(ctx context.Context, out chan<- Packet) error
}

// ErrPermission is returned when the OS denies packet capture.
// macOS: /dev/bpf* needs read access (sudo, or a ChmodBPF-style LaunchDaemon).
// Linux: needs CAP_NET_RAW (setcap) or root. Windows: Npcap must be installed.
var ErrPermission = errors.New("capture: permission denied")

// Interfaces lists physical, up, non-loopback interfaces (VMs/tunnels filtered).
func Interfaces() ([]string, error) { return getAllPhysicalInterface() }

// RunAll runs every source concurrently, fanning packets into out. It returns
// when all sources stop; the first non-nil error is returned.
func RunAll(ctx context.Context, out chan<- Packet, sources ...Source) error {
	var (
		wg    sync.WaitGroup
		once  sync.Once
		first error
	)
	for _, s := range sources {
		wg.Add(1)
		go func(s Source) {
			defer wg.Done()
			if err := s.Run(ctx, out); err != nil {
				once.Do(func() { first = err })
			}
		}(s)
	}
	wg.Wait()
	return first
}

func classify(err error) error {
	if err == nil {
		return nil
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "permission") || strings.Contains(msg, "operation not permitted") || strings.Contains(msg, "/dev/bpf") {
		return errors.Join(ErrPermission, err)
	}
	return err
}
