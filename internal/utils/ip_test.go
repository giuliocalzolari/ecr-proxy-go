package utils

import (
	"testing"
)

func TestIsIPAllowed(t *testing.T) {
	tests := []struct {
		name        string
		remoteAddr  string
		ipWhitelist string
		want        bool
	}{
		{
			name:        "IPv4 exact match",
			remoteAddr:  "192.168.1.10:12345",
			ipWhitelist: "192.168.1.10",
			want:        true,
		},
		{
			name:        "IPv4 in CIDR",
			remoteAddr:  "10.0.0.5:54321",
			ipWhitelist: "10.0.0.0/8",
			want:        true,
		},
		{
			name:        "IPv4 not in CIDR",
			remoteAddr:  "172.16.0.1:80",
			ipWhitelist: "192.168.0.0/16",
			want:        false,
		},
		{
			name:        "IPv6 exact match",
			remoteAddr:  "[2001:db8::1]:443",
			ipWhitelist: "2001:db8::1",
			want:        true,
		},
		{
			name:        "IPv6 in CIDR",
			remoteAddr:  "[2001:db8::2]:443",
			ipWhitelist: "2001:db8::/32",
			want:        true,
		},
		{
			name:        "IPv6 not in CIDR",
			remoteAddr:  "[2001:db9::1]:443",
			ipWhitelist: "2001:db8::/32",
			want:        false,
		},
		{
			name:        "Multiple whitelist entries, match second",
			remoteAddr:  "10.1.2.3:8080",
			ipWhitelist: "192.168.1.1,10.0.0.0/8",
			want:        true,
		},
		{
			name:        "Multiple whitelist entries, no match",
			remoteAddr:  "8.8.8.8:53",
			ipWhitelist: "192.168.1.1,10.0.0.0/8",
			want:        false,
		},
		{
			name:        "IPv4 without port",
			remoteAddr:  "127.0.0.1",
			ipWhitelist: "127.0.0.1",
			want:        true,
		},
		{
			name:        "Empty whitelist",
			remoteAddr:  "127.0.0.1:1234",
			ipWhitelist: "",
			want:        false,
		},
		{
			name:        "Malformed remoteAddr",
			remoteAddr:  "not_an_ip",
			ipWhitelist: "127.0.0.1",
			want:        false,
		},
		{
			name:        "Malformed whitelist entry",
			remoteAddr:  "127.0.0.1:1234",
			ipWhitelist: "bad_cidr,127.0.0.1",
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsIPAllowed(tt.remoteAddr, tt.ipWhitelist)
			if got != tt.want {
				t.Errorf("IsIPAllowed(%q, %q) = %v; want %v", tt.remoteAddr, tt.ipWhitelist, got, tt.want)
			}
		})
	}
}
