package main

import (
    "testing"
    "strings"
)

func TestParseMessage(t *testing.T) {
    s := "<86> aiops9242 sudo: pam_unix(sudo:session): session opened for user root(uid=0) by motadata(uid=1000)"
    p := parseMessage(s)
    if p.Username == "" {
        t.Fatal("expected username to be parsed")
    }
    if p.EventCategory != "login.audit" {
        t.Fatal("expected login.audit category")
    }
    if !strings.Contains(p.RawMessage, "session opened") {
        t.Fatal("raw message mismatch")
    }
}
