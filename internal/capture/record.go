package capture

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// .aorec format: magic "AOREC1\n" then records of
// [int64 unix-nanos BE][uint8 len(srcIP)][srcIP][uint32 len(payload) BE][payload].
// Recordings let us develop and test decoders without the game running.
const recMagic = "AOREC1\n"

// Recorder appends packets to an .aorec file. Safe for concurrent use.
type Recorder struct {
	mu   sync.Mutex
	f    *os.File
	w    *bufio.Writer
	path string
}

func NewRecorder(path string) (*Recorder, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	w := bufio.NewWriter(f)
	if _, err := w.WriteString(recMagic); err != nil {
		f.Close()
		return nil, err
	}
	return &Recorder{f: f, w: w, path: path}, nil
}

// Path is the file the recorder writes to.
func (r *Recorder) Path() string { return r.path }

func (r *Recorder) Write(p Packet) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	var hdr [8]byte
	binary.BigEndian.PutUint64(hdr[:], uint64(p.At.UnixNano()))
	r.w.Write(hdr[:])
	r.w.WriteByte(byte(len(p.SrcIP)))
	r.w.WriteString(p.SrcIP)
	var l [4]byte
	binary.BigEndian.PutUint32(l[:], uint32(len(p.Payload)))
	r.w.Write(l[:])
	_, err := r.w.Write(p.Payload)
	return err
}

func (r *Recorder) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := r.w.Flush(); err != nil {
		r.f.Close()
		return err
	}
	return r.f.Close()
}

// ReplaySource plays an .aorec file. Speed 0 = as fast as possible,
// 1 = real time, 2 = twice as fast, etc.
type ReplaySource struct {
	Path  string
	Speed float64
}

func (s ReplaySource) Name() string { return "replay:" + s.Path }

func (s ReplaySource) Run(ctx context.Context, out chan<- Packet) error {
	f, err := os.Open(s.Path)
	if err != nil {
		return err
	}
	defer f.Close()
	r := bufio.NewReader(f)
	magic := make([]byte, len(recMagic))
	if _, err := io.ReadFull(r, magic); err != nil || string(magic) != recMagic {
		return fmt.Errorf("%s: not an .aorec file", s.Path)
	}
	var prev time.Time
	for {
		p, err := readRecord(r)
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("%s: %w", s.Path, err)
		}
		if s.Speed > 0 && !prev.IsZero() {
			if d := time.Duration(float64(p.At.Sub(prev)) / s.Speed); d > 0 {
				select {
				case <-time.After(d):
				case <-ctx.Done():
					return nil
				}
			}
		}
		prev = p.At
		select {
		case out <- p:
		case <-ctx.Done():
			return nil
		}
	}
}

func readRecord(r *bufio.Reader) (Packet, error) {
	var hdr [8]byte
	if _, err := io.ReadFull(r, hdr[:]); err != nil {
		return Packet{}, err
	}
	ipLen, err := r.ReadByte()
	if err != nil {
		return Packet{}, io.ErrUnexpectedEOF
	}
	ip := make([]byte, ipLen)
	if _, err := io.ReadFull(r, ip); err != nil {
		return Packet{}, io.ErrUnexpectedEOF
	}
	var l [4]byte
	if _, err := io.ReadFull(r, l[:]); err != nil {
		return Packet{}, io.ErrUnexpectedEOF
	}
	payload := make([]byte, binary.BigEndian.Uint32(l[:]))
	if _, err := io.ReadFull(r, payload); err != nil {
		return Packet{}, io.ErrUnexpectedEOF
	}
	return Packet{At: time.Unix(0, int64(binary.BigEndian.Uint64(hdr[:]))), SrcIP: string(ip), Payload: payload}, nil
}
