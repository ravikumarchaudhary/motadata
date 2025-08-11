package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

type RawPayload struct {
	Message   string `json:"message"`
	Timestamp string `json:"timestamp,omitempty"`
}

type Parsed struct {
	Timestamp     time.Time `json:"timestamp"`
	EventCategory string    `json:"event.category"`
	Username      string    `json:"username,omitempty"`
	Hostname      string    `json:"hostname,omitempty"`
	Severity      string    `json:"severity,omitempty"`
	RawMessage    string    `json:"raw.message,omitempty"`
	IsBlacklisted bool      `json:"is.blacklisted,omitempty"`
}

var blacklist = map[string]bool{
	"baduser":   true,
	"192.0.2.1": true,
}

var wordUserRegex = regexp.MustCompile(`by\s+([A-Za-z0-9_\-]+)`)
var hostRegex = regexp.MustCompile(`^<\d+>\s*([^\s]+)`)
var severityRegex = regexp.MustCompile(`^<(\d+)>`)

func parseMessage(s string) Parsed {
	p := Parsed{RawMessage: s}
	p.Timestamp = time.Now().UTC()
	if m := severityRegex.FindStringSubmatch(s); len(m) > 1 {
		code := m[1]
		p.Severity = mapSeverity(code)
	}
	if m := hostRegex.FindStringSubmatch(s); len(m) > 1 {
		p.Hostname = m[1]
	}
	if m := wordUserRegex.FindStringSubmatch(s); len(m) > 1 {
		p.Username = m[1]
	}
	if strings.Contains(strings.ToLower(s), "login") || strings.Contains(strings.ToLower(s), "logged on") || strings.Contains(strings.ToLower(s), "session opened") {
		p.EventCategory = "login.audit"
	} else if strings.Contains(strings.ToLower(s), "logout") || strings.Contains(strings.ToLower(s), "session closed") || strings.Contains(strings.ToLower(s), "terminated") {
		p.EventCategory = "logout.audit"
	} else {
		p.EventCategory = "event"
	}

	// blacklisted check
	if p.Username != "" && blacklist[strings.ToLower(p.Username)] {
		p.IsBlacklisted = true
	}
	for ip := range blacklist { //check if IPs in message
		if strings.Contains(s, ip) {
			p.IsBlacklisted = true
		}
	}
	return p
}

func mapSeverity(code string) string {
	if code == "86" {
		return "INFO"
	} // simple mapping based on lowest 3 bits / static mapping for demo
	if code == "134" {
		return "WARN"
	}
	return "INFO"
}

func forwardToServer(parsed Parsed) error {
	server := os.Getenv("LOG_SERVER_URL") //get the env of the log server URL.....
	if server == "" {
		server = "http://log-server:8081/ingest"
	}
	b, _ := json.Marshal(parsed)
	resp, err := http.Post(server, "application/json", strings.NewReader(string(b)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("remote error: %s", string(body))
	}
	return nil
}

func handleConn(conn net.Conn) {
	defer conn.Close()
	r := bufio.NewReader(conn)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				log.Println("read error:", err)
			}
			break
		}
		line = strings.TrimSpace(line)
		var raw RawPayload
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			log.Println("invalid json:", err, "line:", line)
			continue
		}
		parsed := parseMessage(raw.Message)
		if raw.Timestamp != "" {
			if t, err := time.Parse(time.RFC3339, raw.Timestamp); err == nil {
				parsed.Timestamp = t
			}
		}
		go func(p Parsed) { // dont wait for the response and forward(asynchronously)
			if err := forwardToServer(p); err != nil {
				log.Println("forward error:", err)
			}
		}(parsed)
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "6000"
	} // Listen port 6000.....
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("log-collector listening on", port)
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("accept error:", err)
			continue
		}
		go handleConn(conn)
	}
}
