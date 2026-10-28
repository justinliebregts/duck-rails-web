package model

import "testing"

func TestPlayerFillIPFromIPs(t *testing.T) {
	cases := []struct {
		name string
		ips  []string
		want string
	}{
		{"IPv4", []string{"203.0.113.5"}, "203.0.113.5"},
		{"IPv6 and proxy exceed column", []string{"2001:0db8:1234:5678:90ab:cdef:1234:5678", "2001:db8:ffff:ffff:ffff:ffff:ffff:ffff"}, "2001:db8:1234:5678:90ab:cdef:1234:5678"},
		{"raw chain", []string{" 2001:db8::1, 10.0.0.1"}, "2001:db8::1"},
		{"invalid entry", []string{"unknown", "203.0.113.5"}, "203.0.113.5"},
		{"invalid only preserves existing", []string{"not an IP"}, "192.0.2.1"},
		{"no addresses preserves existing", nil, "192.0.2.1"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := Player{IP: "192.0.2.1", IPs: tc.ips}
			p.FillIPFromIPs()
			if p.IP != tc.want {
				t.Fatalf("IP = %q, want %q", p.IP, tc.want)
			}
			if len(p.IP) > 45 {
				t.Fatalf("IP exceeds schema length: %d", len(p.IP))
			}
		})
	}
}
