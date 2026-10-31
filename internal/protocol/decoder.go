// Package protocol turns raw Photon payloads into named Albion messages.
//
// Pipeline: capture.Packet → photon.PhotonParser → Decoder → Message{Name, Params}.
// Typed decoding of Params into structs lives in internal/events.
package protocol

import (
	"time"

	"github.com/LucasGitDev/albion-companion/internal/photon"
)

type Kind uint8

const (
	KindRequest Kind = iota + 1
	KindResponse
	KindEvent
)

func (k Kind) String() string {
	switch k {
	case KindRequest:
		return "request"
	case KindResponse:
		return "response"
	case KindEvent:
		return "event"
	}
	return "unknown"
}

// Albion puts the real event code in param 252 and the operation code in 253;
// the Photon-level code byte is not reliable.
const (
	paramEventCode     byte = 252
	paramOperationCode byte = 253
)

// Message is one decoded Photon request, response or event.
type Message struct {
	Kind       Kind
	Code       uint16
	Name       string // "" when the code is not in codes.json
	ReturnCode int16  // responses only
	Params     Params
	At         time.Time
}

// Decoder feeds payloads to the Photon parser and emits named Messages.
// Not safe for concurrent use: call Feed from a single goroutine (the engine does).
type Decoder struct {
	codes     *Codes
	parser    *photon.PhotonParser
	emit      func(Message)
	at        time.Time
	Encrypted int // count of encrypted commands seen (diagnostics)
}

func NewDecoder(codes *Codes, emit func(Message)) *Decoder {
	d := &Decoder{codes: codes, emit: emit}
	d.parser = photon.NewPhotonParser(d.onRequest, d.onResponse, d.onEvent)
	d.parser.OnEncrypted = func() { d.Encrypted++ }
	return d
}

// SetCodes swaps the code table (e.g. after loading an override file).
func (d *Decoder) SetCodes(c *Codes) { d.codes = c }

// Feed parses one UDP/TCP payload captured at time at.
func (d *Decoder) Feed(payload []byte, at time.Time) {
	d.at = at
	d.parser.ReceivePacket(payload)
}

func (d *Decoder) onRequest(code byte, params map[byte]any) {
	c := d.resolve(params, paramOperationCode, uint16(code), d.isOp)
	name, _ := d.codes.OperationName(c)
	d.emit(Message{Kind: KindRequest, Code: c, Name: name, Params: params, At: d.at})
}

func (d *Decoder) onResponse(code byte, ret int16, _ string, params map[byte]any) {
	c := d.resolve(params, paramOperationCode, uint16(code), d.isOp)
	name, _ := d.codes.OperationName(c)
	d.emit(Message{Kind: KindResponse, Code: c, Name: name, ReturnCode: ret, Params: params, At: d.at})
}

func (d *Decoder) onEvent(code byte, params map[byte]any) {
	c := d.resolve(params, paramEventCode, uint16(code), d.isEvent)
	name, _ := d.codes.EventName(c)
	d.emit(Message{Kind: KindEvent, Code: c, Name: name, Params: params, At: d.at})
}

func (d *Decoder) isOp(c uint16) bool    { _, ok := d.codes.OperationName(c); return ok }
func (d *Decoder) isEvent(c uint16) bool { _, ok := d.codes.EventName(c); return ok }

func (d *Decoder) resolve(params map[byte]any, key byte, fallback uint16, known func(uint16) bool) uint16 {
	v, ok := toInt64(params[key])
	if !ok {
		return fallback
	}
	return normalize(uint16(v), known)
}

// normalize undoes byte-order artifacts seen after some Albion updates
// (ported from the albiondata-client fork).
func normalize(code uint16, known func(uint16) bool) uint16 {
	if known(code) {
		return code
	}
	if swapped := code<<8 | code>>8; known(swapped) {
		return swapped
	}
	if code > 0x00FF && code&0x00FF == 0 {
		return code >> 8
	}
	return code
}
