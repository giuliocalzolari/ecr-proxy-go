package main

import (
	"net/http"
	"testing"
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
