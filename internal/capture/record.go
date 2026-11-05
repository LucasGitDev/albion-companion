package capture

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"sync"
	"sync/atomic"
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

// ReplayControl drives an in-progress replay: live speed changes, pause/resume,
// and a per-packet progress callback. Zero value is not usable; use
// NewReplayControl. Safe for concurrent use.
type ReplayControl struct {
	mu        sync.Mutex
	cond      *sync.Cond
	paused    bool
	stopped   bool
	speedBits atomic.Uint64 // math.Float64bits of the speed multiplier (0 = max)
	// OnProgress is called with each packet's recorded time as it is played.
	OnProgress func(at time.Time)
}

func NewReplayControl(speed float64) *ReplayControl {
	c := &ReplayControl{}
	c.cond = sync.NewCond(&c.mu)
	c.SetSpeed(speed)
	return c
}

func (c *ReplayControl) SetSpeed(s float64) { c.speedBits.Store(math.Float64bits(s)) }
func (c *ReplayControl) Speed() float64     { return math.Float64frombits(c.speedBits.Load()) }

func (c *ReplayControl) Pause() {
	c.mu.Lock()
	c.paused = true
	c.mu.Unlock()
}

func (c *ReplayControl) Resume() {
	c.mu.Lock()
	c.paused = false
	c.mu.Unlock()
	c.cond.Broadcast()
}

func (c *ReplayControl) Paused() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.paused
}

// Stop wakes any paused wait so the replay goroutine can exit (used alongside
// cancelling the context).
func (c *ReplayControl) Stop() {
	c.mu.Lock()
	c.stopped = true
	c.mu.Unlock()
	c.cond.Broadcast()
}

// wait blocks while paused; returns false if the replay was stopped.
func (c *ReplayControl) wait() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	for c.paused && !c.stopped {
		c.cond.Wait()
	}
	return !c.stopped
}

// ReplaySource plays an .aorec file. Speed 0 = as fast as possible,
// 1 = real time, 2 = twice as fast, etc. When Control is set it overrides Speed
// and allows live pause/resume/speed changes and progress reporting.
type ReplaySource struct {
	Path    string
	Speed   float64
	Control *ReplayControl
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
	sinceSleep := 0 // packets processed since the last time we yielded
	for {
		p, err := readRecord(r)
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("%s: %w", s.Path, err)
		}
		if s.Control != nil && !s.Control.wait() {
			return nil // stopped while paused
		}
		speed := s.Speed
		if s.Control != nil {
			speed = s.Control.Speed()
		}
		slept := false
		if speed > 0 && !prev.IsZero() {
			if d := time.Duration(float64(p.At.Sub(prev)) / speed); d > 0 {
				select {
				case <-time.After(d):
					slept = true
				case <-ctx.Done():
					return nil
				}
			}
		}
		prev = p.At
		// Safety yield: bursts of same-timestamp packets (delta 0) and very high
		// speeds would otherwise be a tight CPU loop that can starve the UI/OS.
		// Yield briefly every maxBurst packets whenever we haven't slept, so the
		// scheduler always gets air regardless of speed.
		if slept {
			sinceSleep = 0
		} else if sinceSleep++; sinceSleep >= replayMaxBurst {
			sinceSleep = 0
			select {
			case <-time.After(time.Millisecond):
			case <-ctx.Done():
				return nil
			}
		}
		select {
		case out <- p:
		case <-ctx.Done():
			return nil
		}
		if s.Control != nil && s.Control.OnProgress != nil {
			s.Control.OnProgress(p.At)
		}
	}
}

// replayMaxBurst bounds how many packets replay processes without yielding, so
// same-timestamp bursts and high speeds never peg a core / freeze the UI.
const replayMaxBurst = 256

// AorecSpan scans a recording's headers and returns the first and last packet
// times and the packet count, without decoding payloads. Used to show replay
// progress (elapsed vs total).
func AorecSpan(path string) (first, last time.Time, count int, err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	r := bufio.NewReader(f)
	magic := make([]byte, len(recMagic))
	if _, e := io.ReadFull(r, magic); e != nil || string(magic) != recMagic {
		err = fmt.Errorf("%s: not an .aorec file", path)
		return
	}
	for {
		p, e := readRecord(r)
		if errors.Is(e, io.EOF) {
			return
		}
		if e != nil {
			err = fmt.Errorf("%s: %w", path, e)
			return
		}
		if count == 0 {
			first = p.At
		}
		last = p.At
		count++
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
