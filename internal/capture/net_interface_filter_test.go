package capture

import (
	"net"
	"testing"
)

func TestIsUsableCaptureIP(t *testing.T) {
	tests := []struct {
		name string
		ip   net.IP
		want bool
	}{
		{"nil", nil, false},
		{"ipv4 loopback", net.ParseIP("127.0.0.1"), false},
		{"ipv6 loopback", net.ParseIP("::1"), false},
		{"ipv4 link-local", net.ParseIP("169.254.1.5"), false},
		{"ipv6 link-local", net.ParseIP("fe80::1"), false},
		{"unspecified v4", net.ParseIP("0.0.0.0"), false},
		{"unspecified v6", net.ParseIP("::"), false},
		{"private v4 (en0/LAN)", net.ParseIP("192.168.1.11"), true},
		{"cgnat v4", net.ParseIP("100.64.0.2"), true},
		{"public v4", net.ParseIP("85.234.70.222"), true},
		{"global v6", net.ParseIP("2001:db8::1"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isUsableCaptureIP(tt.ip); got != tt.want {
				t.Fatalf("isUsableCaptureIP(%v) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}

// Sanity: the real host enumeration never panics and only returns interfaces
// that actually carry a usable IP (the whole point of the TASK-38 filter).
func TestGetAllPhysicalInterface_OnlyAddressed(t *testing.T) {
	names, err := getAllPhysicalInterface()
	if err != nil {
		t.Skipf("interface enumeration unavailable: %v", err)
	}
	byName := map[string]net.Interface{}
	ifaces, _ := net.Interfaces()
	for _, in := range ifaces {
		byName[in.Name] = in
	}
	for _, n := range names {
		in, ok := byName[n]
		if !ok {
			// Windows returns \Device\NPF_ names that won't match net.Interfaces.
			continue
		}
		if !ifaceHasUsableIP(in) {
			t.Errorf("interface %q returned but has no usable IP", n)
		}
	}
}
