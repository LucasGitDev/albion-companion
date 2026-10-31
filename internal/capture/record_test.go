package capture

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestRecordReplayRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "s.aorec")
	rec, err := NewRecorder(path)
	if err != nil {
		t.Fatal(err)
	}
	in := []Packet{
		{At: time.Unix(10, 5), SrcIP: "5.188.125.10", Payload: []byte{1, 2, 3}},
		{At: time.Unix(11, 0), SrcIP: "5.188.125.10", Payload: []byte{4}},
	}
	for _, p := range in {
		if err := rec.Write(p); err != nil {
			t.Fatal(err)
		}
	}
	if err := rec.Close(); err != nil {
		t.Fatal(err)
	}

	out := make(chan Packet, 10)
	if err := (ReplaySource{Path: path}).Run(context.Background(), out); err != nil {
		t.Fatal(err)
	}
	close(out)
	var got []Packet
	for p := range out {
		got = append(got, p)
	}
	if len(got) != len(in) {
		t.Fatalf("got %d packets", len(got))
	}
	for i := range in {
		if !got[i].At.Equal(in[i].At) || got[i].SrcIP != in[i].SrcIP || !bytes.Equal(got[i].Payload, in[i].Payload) {
			t.Fatalf("packet %d: got %+v want %+v", i, got[i], in[i])
		}
	}
}
