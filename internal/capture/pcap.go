package capture

import (
	"context"
	"fmt"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

// LiveSource captures from one network interface.
type LiveSource struct {
	Device string
	Port   int // 0 = GamePort
}

func (s LiveSource) Name() string { return "live:" + s.Device }

func (s LiveSource) Run(ctx context.Context, out chan<- Packet) error {
	port := s.Port
	if port == 0 {
		port = GamePort
	}
	// Short read timeout so cancellation is noticed even when idle.
	h, err := pcap.OpenLive(s.Device, 65535, false, 250*time.Millisecond)
	if err != nil {
		return classify(fmt.Errorf("open %s: %w", s.Device, err))
	}
	if err := h.SetBPFFilter(fmt.Sprintf("udp port %d or tcp port %d", port, port)); err != nil {
		h.Close()
		return fmt.Errorf("bpf filter on %s: %w", s.Device, err)
	}
	return pump(ctx, h, out)
}

// LiveSources builds one LiveSource per physical interface.
func LiveSources(port int) ([]Source, error) {
	devs, err := Interfaces()
	if err != nil {
		return nil, err
	}
	out := make([]Source, 0, len(devs))
	for _, d := range devs {
		out = append(out, LiveSource{Device: d, Port: port})
	}
	return out, nil
}

// PcapFileSource replays a .pcap/.pcapng file (e.g. recorded with Wireshark).
type PcapFileSource struct{ Path string }

func (s PcapFileSource) Name() string { return "pcap:" + s.Path }

func (s PcapFileSource) Run(ctx context.Context, out chan<- Packet) error {
	h, err := pcap.OpenOffline(s.Path)
	if err != nil {
		return err
	}
	return pump(ctx, h, out)
}

func pump(ctx context.Context, h *pcap.Handle, out chan<- Packet) error {
	defer h.Close()
	src := gopacket.NewPacketSource(h, h.LinkType())
	packets := src.Packets()
	for {
		select {
		case <-ctx.Done():
			return nil
		case p, ok := <-packets:
			if !ok {
				return nil
			}
			if pkt, ok := toPacket(p); ok {
				select {
				case out <- pkt:
				case <-ctx.Done():
					return nil
				}
			}
		}
	}
}

func toPacket(p gopacket.Packet) (Packet, bool) {
	var src string
	if l, ok := p.Layer(layers.LayerTypeIPv4).(*layers.IPv4); ok {
		src = l.SrcIP.String()
	} else if l, ok := p.Layer(layers.LayerTypeIPv6).(*layers.IPv6); ok {
		src = l.SrcIP.String()
	} else {
		return Packet{}, false
	}
	var payload []byte
	if l, ok := p.Layer(layers.LayerTypeUDP).(*layers.UDP); ok {
		payload = l.Payload
	} else if l, ok := p.Layer(layers.LayerTypeTCP).(*layers.TCP); ok {
		payload = l.Payload
	}
	if len(payload) == 0 {
		return Packet{}, false
	}
	at := p.Metadata().Timestamp
	if at.IsZero() {
		at = time.Now()
	}
	return Packet{At: at, SrcIP: src, Payload: append([]byte(nil), payload...)}, true
}
