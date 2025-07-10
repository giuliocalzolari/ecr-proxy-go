package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"
)

type LogEntry struct {
	Method     string `json:"method"`
	SourceIp   string `json:"ip"`
	SourcePort string `json:"port"`
	Time       string `json:"time"`
	Msg        string `json:"msg"`
	Path       string `json:"path"`
}

func Log(r *http.Request, msg string) {
	// Log the request details
	remoteAddr := r.RemoteAddr
	host, port, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr // Fallback if no port is present
		port = ""
	}
	logEntry := LogEntry{
		Method:     r.Method,
		SourceIp:   host,
		SourcePort: port,
		Time:       time.Now().Format(time.RFC3339),
		Msg:        msg,
		Path:       r.URL.Path,
	}
	logData, _ := json.Marshal(logEntry)
	fmt.Println(string(logData))
}
