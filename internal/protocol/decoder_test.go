package protocol

import (
	"encoding/binary"
	"testing"
	"time"

	"github.com/LucasGitDev/albion-companion/internal/photon"
)

// buildPacket wraps one reliable message into a minimal Photon UDP packet.
func buildPacket(msgType, code byte, body []byte) []byte {
	hdr := []byte{0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0}
	cmdData := append([]byte{0x00, msgType, code}, body...)
	cmdHdr := make([]byte, 12)
	cmdHdr[0] = 6 // SendReliable
	binary.BigEndian.PutUint32(cmdHdr[4:], uint32(12+len(cmdData)))
	return append(append(hdr, cmdHdr...), cmdData...)
}

func TestDecoder_EventNameFromParam252(t *testing.T) {
	codes := DefaultCodes()
	want, ok := codes.EventCode("OtherGrabbedLoot")
	if !ok {
		t.Fatal("OtherGrabbedLoot missing from codes.json")
	}

	// params: count=2 | 2:String "Bob" | 252:Short <code>
	body := []byte{2, 2, photon.TypeString, 3, 'B', 'o', 'b', 252, 4}
	body = binary.LittleEndian.AppendUint16(body, want)

	var got []Message
	d := NewDecoder(codes, func(m Message) { got = append(got, m) })
	at := time.Unix(100, 0)
	d.Feed(buildPacket(photon.MsgEvent, 1, body), at)

	if len(got) != 1 {
		t.Fatalf("got %d messages, want 1", len(got))
	}
	m := got[0]
	if m.Kind != KindEvent || m.Name != "OtherGrabbedLoot" || m.Code != want || !m.At.Equal(at) {
		t.Fatalf("unexpected message: %+v", m)
	}
	if s, _ := m.Params.String(2); s != "Bob" {
		t.Fatalf("param 2 = %q, want Bob", s)
	}
}

func TestNormalize_ByteSwapped(t *testing.T) {
	known := func(c uint16) bool { return c == 0x0116 }
	if got := normalize(0x1601, known); got != 0x0116 {
		t.Fatalf("normalize = %#x, want 0x0116", got)
	}
}

func TestParams_IntWidths(t *testing.T) {
	p := Params{1: byte(7), 2: int16(-3), 3: int32(70000), 4: int64(1 << 40), 5: float32(2.6)}
	for k, want := range map[byte]int64{1: 7, 2: -3, 3: 70000, 4: 1 << 40, 5: 3} {
		if got, ok := p.Int64(k); !ok || got != want {
			t.Errorf("Int64(%d) = %d,%v want %d", k, got, ok, want)
		}
	}
}
