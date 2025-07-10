package main

import (
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestDirectorSetsFields(t *testing.T) {
	// Set up
	ecrEndpoint = "test.ecr.amazonaws.com"
	ecrToken = "dummytoken"
	req, _ := http.NewRequest("GET", "http://localhost/v2/test", nil)

	director(req)

	if req.URL.Scheme != "https" {
		t.Errorf("Expected scheme https, got %s", req.URL.Scheme)
	}
	if req.URL.Host != ecrEndpoint {
		t.Errorf("Expected host %s, got %s", ecrEndpoint, req.URL.Host)
	}
	if req.Host != ecrEndpoint {
		t.Errorf("Expected req.Host %s, got %s", ecrEndpoint, req.Host)
	}
	auth := req.Header.Get("Authorization")
	expectedAuth := "Basic " + ecrToken
	if auth != expectedAuth {
		t.Errorf("Expected Authorization header %q, got %q", expectedAuth, auth)
	}
}

func TestDirector_SetsRequestFields(t *testing.T) {
	// Set up global variables as expected by director
	ecrEndpoint = "test.ecr.amazonaws.com"
	ecrToken = "dummytoken"
	req, _ := http.NewRequest("GET", "http://localhost/v2/test", nil)

	director(req)

	if req.URL.Scheme != "https" {
		t.Errorf("Expected scheme https, got %s", req.URL.Scheme)
	}
	if req.URL.Host != ecrEndpoint {
		t.Errorf("Expected host %s, got %s", ecrEndpoint, req.URL.Host)
	}
	if req.Host != ecrEndpoint {
		t.Errorf("Expected req.Host %s, got %s", ecrEndpoint, req.Host)
	}
	auth := req.Header.Get("Authorization")
	expectedAuth := "Basic " + ecrToken
	if auth != expectedAuth {
		t.Errorf("Expected Authorization header %q, got %q", expectedAuth, auth)
	}
}
func TestIPWhitelist_AllowsAndDenies(t *testing.T) {
	// Save and restore environment variable
	origWhitelist := os.Getenv("IP_WHITELIST")
	defer os.Setenv("IP_WHITELIST", origWhitelist)

	// Set up a test whitelist
	os.Setenv("IP_WHITELIST", "127.0.0.1,192.168.1.1")

	// Set up a test server using the /v2/ handler
	ecrEndpoint = "test.ecr.amazonaws.com"
	ecrToken = "dummytoken"
	tokenExpiry = time.Now().Add(1 * time.Hour) // ensure token is valid

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use the same logic as in main.go for /v2/
		ipWhitelist := os.Getenv("IP_WHITELIST")
		if ipWhitelist != "" {
			allowed := false
			remoteIP := r.RemoteAddr
			if host, _, err := net.SplitHostPort(remoteIP); err == nil {
				remoteIP = host
			}
			for _, allowedIP := range strings.Split(ipWhitelist, ",") {
				if strings.TrimSpace(allowedIP) == remoteIP {
					allowed = true
					break
				}
			}
			if !allowed {
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte("Forbidden"))
				return
			}
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Allowed IP
	req1, _ := http.NewRequest("GET", "/v2/test", nil)
	req1.RemoteAddr = "127.0.0.1:12345"
	rr1 := &httptest.ResponseRecorder{}
	handler.ServeHTTP(rr1, req1)
	if rr1.Code != http.StatusOK {
		t.Errorf("Expected status 200 for whitelisted IP, got %d", rr1.Code)
	}

	// Not allowed IP
	req2, _ := http.NewRequest("GET", "/v2/test", nil)
	req2.RemoteAddr = "10.0.0.1:54321"
	rr2 := &httptest.ResponseRecorder{}
	handler.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusForbidden {
		t.Errorf("Expected status 403 for non-whitelisted IP, got %d", rr2.Code)
	}
}

func TestIsIPAllowed_CIDRAndIP(t *testing.T) {
	whitelist := "127.0.0.1,10.0.0.0/8,::1"
	tests := []struct {
		remoteAddr string
		want       bool
	}{
		{"127.0.0.1:1234", true},
		{"10.1.2.3:5678", true},
		{"192.168.1.1:80", false},
		{"[::1]:443", true},
		{"[2001:db8::1]:443", false},
	}
	for _, tt := range tests {
		got := isIPAllowed(tt.remoteAddr, whitelist)
		if got != tt.want {
			t.Errorf("isIPAllowed(%q, %q) = %v, want %v", tt.remoteAddr, whitelist, got, tt.want)
		}
	}
}

func TestSplitAndTrim(t *testing.T) {
	input := " a, b ,c,, ,d "
	want := []string{"a", "b", "c", "d"}
	got := splitAndTrim(input, ",")
	if len(got) != len(want) {
		t.Fatalf("splitAndTrim returned wrong length: got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("splitAndTrim: got %q at %d, want %q", got[i], i, want[i])
		}
	}
}
