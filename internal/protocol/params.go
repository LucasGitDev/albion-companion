package protocol

import (
	"encoding/hex"
	"math"
)

// Params is a decoded Photon parameter table.
//
// The Protocol18 deserializer picks the smallest wire type for each value, so
// the same logical field may arrive as byte, int16, int32 or int64 depending on
// its magnitude. Always read through these helpers, never type-assert directly.
type Params map[byte]any

func (p Params) Has(k byte) bool { _, ok := p[k]; return ok }

func (p Params) Int64(k byte) (int64, bool) { return toInt64(p[k]) }

func (p Params) Float64(k byte) (float64, bool) {
	switch v := p[k].(type) {
	case float32:
		return float64(v), true
	case float64:
		return v, true
	}
	i, ok := toInt64(p[k])
	return float64(i), ok
}

func (p Params) String(k byte) (string, bool) {
	s, ok := p[k].(string)
	return s, ok
}

func (p Params) Bool(k byte) (bool, bool) {
	switch v := p[k].(type) {
	case bool:
		return v, true
	case nil:
		return false, false
	}
	i, ok := toInt64(p[k])
	return i != 0, ok
}

// GUID returns a 16-byte array parameter as a lowercase hex string. Only
// equality matters for us, so byte order is irrelevant.
func (p Params) GUID(k byte) (string, bool) { return toGUID(p[k]) }

func (p Params) Int64s(k byte) ([]int64, bool) {
	switch v := p[k].(type) {
	case []byte:
		return mapSlice(v, func(x byte) int64 { return int64(x) }), true
	case []int16:
		return mapSlice(v, func(x int16) int64 { return int64(x) }), true
	case []int32:
		return mapSlice(v, func(x int32) int64 { return int64(x) }), true
	case []int64:
		return v, true
	case []any:
		out := make([]int64, 0, len(v))
		for _, x := range v {
			i, ok := toInt64(x)
			if !ok {
				return nil, false
			}
			out = append(out, i)
		}
		return out, true
	}
	return nil, false
}

func (p Params) Float64s(k byte) ([]float64, bool) {
	switch v := p[k].(type) {
	case []float32:
		return mapSlice(v, func(x float32) float64 { return float64(x) }), true
	case []float64:
		return v, true
	}
	ints, ok := p.Int64s(k)
	if !ok {
		return nil, false
	}
	return mapSlice(ints, func(x int64) float64 { return float64(x) }), true
}

func (p Params) Strings(k byte) ([]string, bool) {
	switch v := p[k].(type) {
	case []string:
		return v, true
	case []any:
		out := make([]string, 0, len(v))
		for _, x := range v {
			s, ok := x.(string)
			if !ok {
				return nil, false
			}
			out = append(out, s)
		}
		return out, true
	}
	return nil, false
}

func (p Params) GUIDs(k byte) ([]string, bool) {
	var items []any
	switch v := p[k].(type) {
	case []any:
		items = v
	case [][]byte:
		for _, b := range v {
			items = append(items, b)
		}
	case []byte:
		// Flat concatenation of 16-byte GUIDs (SA: GetPartyUserGuidsFromBytes).
		if len(v)%16 != 0 {
			return nil, false
		}
		for i := 0; i < len(v); i += 16 {
			items = append(items, v[i:i+16])
		}
	case []int8:
		if len(v)%16 != 0 {
			return nil, false
		}
		for i := 0; i < len(v); i += 16 {
			items = append(items, v[i:i+16])
		}
	default:
		return nil, false
	}
	out := make([]string, 0, len(items))
	for _, x := range items {
		g, ok := toGUID(x)
		if !ok {
			return nil, false
		}
		out = append(out, g)
	}
	return out, true
}

func toInt64(v any) (int64, bool) {
	switch n := v.(type) {
	case byte:
		return int64(n), true
	case int8:
		return int64(n), true
	case int16:
		return int64(n), true
	case uint16:
		return int64(n), true
	case int32:
		return int64(n), true
	case uint32:
		return int64(n), true
	case int64:
		return n, true
	case int:
		return int64(n), true
	case float32:
		return int64(math.Round(float64(n))), true
	case float64:
		return int64(math.Round(n)), true
	}
	return 0, false
}

func toGUID(v any) (string, bool) {
	switch b := v.(type) {
	case []byte:
		if len(b) == 16 {
			return hex.EncodeToString(b), true
		}
	case []int8:
		if len(b) == 16 {
			raw := make([]byte, 16)
			for i, x := range b {
				raw[i] = byte(x)
			}
			return hex.EncodeToString(raw), true
		}
	}
	return "", false
}

func mapSlice[T, U any](in []T, f func(T) U) []U {
	out := make([]U, len(in))
	for i, x := range in {
		out[i] = f(x)
	}
	return out
}
