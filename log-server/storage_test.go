package main

import (
    "os"
    "testing"
    "time"
)

func TestFileStorageSaveQuery(t *testing.T) {
    path := "/tmp/test_logs.jsonl"
    os.Remove(path)
    s := NewFileStorage(path)
    e := LogEntry{
        Timestamp: time.Now().UTC(),
        EventCategory: "linux_login",
        Username: "alice",
        Hostname: "host1",
        Severity: "INFO",
        RawMessage: "test",
        IsBlacklisted: false,
    }
    if err := s.Save(e); err != nil {
        t.Fatal("save error:", err)
    }
    res, err := s.Query(map[string]string{"username":"alice"}, 10, "timestamp")
    if err != nil {
        t.Fatal(err)
    }
    if len(res) == 0 {
        t.Fatal("expected at least one result")
    }
    os.Remove(path)
}
