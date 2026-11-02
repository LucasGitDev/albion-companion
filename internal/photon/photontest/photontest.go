// Package photontest builds synthetic Protocol18 Photon packets for tests.
// Supported value types: string, bool, int16, int32/int (as Short when it fits,
// else compressed int), int64, float32, []byte, []string, []float32, []int32.
package photontest

import (
	"encoding/binary"
	"fmt"
	"maps"
	"math"
	"sort"
)

// Protocol18 type codes (subset).
const (
	tBool          = 2
	tShort         = 4
	tFloat         = 5
	tString        = 7
	tCompressedInt = 9
	tCompLong      = 10
	tBoolTrue      = 28
	tBoolFalse     = 27
	tArray         = 0x40
	tByte          = 3
)

// Event returns a full UDP payload carrying one event. Param 252 (event code)
// is added automatically.
func Event(code uint16, params map[byte]any) []byte {
	p := clone(params)
	p[252] = int16(code)
	return packet(4, 1, Table(p))
}

// Response returns a full UDP payload carrying one operation response.
func Response(opCode uint16, params map[byte]any) []byte {
	p := clone(params)
	p[253] = int16(opCode)
	body := []byte{0, 0, 8} // returnCode=0 LE, debug message = null
	return packet(3, 1, append(body, Table(p)...))
}

// Request returns a full UDP payload carrying one operation request. The op code
// goes in param 253 (the decoder resolves it from there, as for responses).
func Request(opCode uint16, params map[byte]any) []byte {
	p := clone(params)
	p[253] = int16(opCode)
	return packet(2, 1, Table(p))
}

// Table encodes a Protocol18 parameter table (keys sorted for determinism).
func Table(params map[byte]any) []byte {
	keys := make([]int, 0, len(params))
	for k := range params {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	out := varint(uint32(len(keys)))
	for _, k := range keys {
		out = append(out, byte(k))
		out = append(out, value(params[byte(k)])...)
	}
	return out
}

func value(v any) []byte {
	switch x := v.(type) {
	case string:
		return append(append([]byte{tString}, varint(uint32(len(x)))...), x...)
	case bool:
		if x {
			return []byte{tBoolTrue}
		}
		return []byte{tBoolFalse}
	case int16:
		return binary.LittleEndian.AppendUint16([]byte{tShort}, uint16(x))
	case int:
		return value(int32(x))
	case int32:
		if x >= math.MinInt16 && x <= math.MaxInt16 {
			return value(int16(x))
		}
		return append([]byte{tCompressedInt}, varint(zigzag32(x))...)
	case int64:
		return append([]byte{tCompLong}, varint64(zigzag64(x))...)
	case float32:
		return binary.LittleEndian.AppendUint32([]byte{tFloat}, math.Float32bits(x))
	case []byte:
		return append(append([]byte{tArray | tByte}, varint(uint32(len(x)))...), x...)
	case []string:
		out := append([]byte{tArray | tString}, varint(uint32(len(x)))...)
		for _, s := range x {
			out = append(append(out, varint(uint32(len(s)))...), s...)
		}
		return out
	case [][]byte: // array-of-arrays, e.g. GUID lists
		out := append([]byte{tArray}, varint(uint32(len(x)))...)
		for _, b := range x {
			out = append(out, value(b)...)
		}
		return out
	case []float32:
		out := append([]byte{tArray | tFloat}, varint(uint32(len(x)))...)
		for _, f := range x {
			out = binary.LittleEndian.AppendUint32(out, math.Float32bits(f))
		}
		return out
	case []int32:
		out := append([]byte{tArray | tCompressedInt}, varint(uint32(len(x)))...)
		for _, n := range x {
			out = append(out, varint(zigzag32(n))...)
		}
		return out
	}
	panic(fmt.Sprintf("photontest: unsupported type %T", v))
}

func packet(msgType, code byte, body []byte) []byte {
	hdr := []byte{0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0}
	cmdData := append([]byte{0x00, msgType, code}, body...)
	cmdHdr := make([]byte, 12)
	cmdHdr[0] = 6 // SendReliable
	binary.BigEndian.PutUint32(cmdHdr[4:], uint32(12+len(cmdData)))
	return append(append(hdr, cmdHdr...), cmdData...)
}

func clone(m map[byte]any) map[byte]any {
	out := make(map[byte]any, len(m)+1)
	maps.Copy(out, m)
	return out
}

func varint(v uint32) []byte { return varint64(uint64(v)) }

func varint64(v uint64) []byte {
	var out []byte
	for v >= 0x80 {
		out = append(out, byte(v)|0x80)
		v >>= 7
	}
	return append(out, byte(v))
}

func zigzag32(n int32) uint32 { return uint32((n << 1) ^ (n >> 31)) }
func zigzag64(n int64) uint64 { return uint64((n << 1) ^ (n >> 63)) }
