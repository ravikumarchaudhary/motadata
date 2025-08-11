package main

import (
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "sort"
    "strconv"
    "strings"
    "sync"
    "time"
)

type LogEntry struct {
    Timestamp     time.Time            `json:"timestamp"`
    EventCategory string               `json:"event.category"`
    Username      string               `json:"username,omitempty"`
    Hostname      string               `json:"hostname,omitempty"`
    Severity      string               `json:"severity,omitempty"`
    RawMessage    string               `json:"raw.message,omitempty"`
    IsBlacklisted bool                 `json:"is.blacklisted,omitempty"`
    Meta          map[string]string    `json:"meta,omitempty"`
}

type Storage interface {
    Save(entry LogEntry) error
    Query(params map[string]string, limit int, sortKey string) ([]LogEntry, error)
    Count() int
    GroupByCategory() map[string]int
    GroupBySeverity() map[string]int
}

type FileStorage struct{
    mu sync.Mutex
    path string
    cache []LogEntry
}

func NewFileStorage(path string) *FileStorage {
    s := &FileStorage{path: path}
    f, err := os.Open(path) // load existing if present
    if err == nil {
        defer f.Close()
        dec := json.NewDecoder(f)
        for {
            var e LogEntry
            if err := dec.Decode(&e); err == io.EOF {
                break
            } else if err != nil {
                break
            }
            s.cache = append(s.cache, e)
        }
    }
    return s
}

func (s *FileStorage) Save(entry LogEntry) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    f, err := os.OpenFile(s.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
    if err != nil {
        return err
    }
    defer f.Close()
    b, _ := json.Marshal(entry)
    _, err = f.Write(append(b, '\n'))
    if err == nil {
        s.cache = append(s.cache, entry)
    }
    return err
}

func matchQuery(e LogEntry, params map[string]string) bool {
    for k, v := range params {
        switch k {
        case "service":
            if v != e.EventCategory { return false }
        case "level":
            if !strings.EqualFold(v, e.Severity) { return false }
        case "username":
            if v != e.Username { return false }
        case "is.blacklisted":
            want := (v == "true")
            if want != e.IsBlacklisted { return false }
        }
    }
    return true
}

func (s *FileStorage) Query(params map[string]string, limit int, sortKey string) ([]LogEntry, error) {
    s.mu.Lock()
    defer s.mu.Unlock()
    res := make([]LogEntry, 0)
    for _, e := range s.cache {
        if matchQuery(e, params) {
            res = append(res, e)
        }
    }
    // sort
    if sortKey == "timestamp" {
        sort.Slice(res, func(i,j int) bool { return res[i].Timestamp.Before(res[j].Timestamp) })
    }
    if limit > 0 && len(res) > limit {
        res = res[:limit]
    }
    return res, nil
}

func (s *FileStorage) Count() int {
    s.mu.Lock()
    defer s.mu.Unlock()
    return len(s.cache)
}
func (s *FileStorage) GroupByCategory() map[string]int {
    s.mu.Lock()
    defer s.mu.Unlock()
    m := map[string]int{}
    for _, e := range s.cache { m[e.EventCategory]++ }
    return m
}
func (s *FileStorage) GroupBySeverity() map[string]int {
    s.mu.Lock()
    defer s.mu.Unlock()
    m := map[string]int{}
    for _, e := range s.cache { m[e.Severity]++ }
    return m
}

var store Storage
// r => object which collect all the information about object like path URl etc...
func ingestHandler(w http.ResponseWriter, r *http.Request) { //https response writer
    if r.Method != http.MethodPost {
        http.Error(w, "only POST allowed", http.StatusMethodNotAllowed)
        return
    }
    var e LogEntry
    dec := json.NewDecoder(r.Body)
    if err := dec.Decode(&e); err != nil {
        http.Error(w, "invalid payload: "+err.Error(), http.StatusBadRequest)
        return
    }
    if e.Timestamp.IsZero() {
        e.Timestamp = time.Now().UTC()
    }
    if e.EventCategory == "" { e.EventCategory = "unknown" }
    if e.RawMessage == "" { e.RawMessage = "" }
    if err := store.Save(e); err != nil {
        http.Error(w, "save error: "+err.Error(), http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusAccepted)
}

func logsHandler(w http.ResponseWriter, r *http.Request) {
    q := r.URL.Query()
    params := map[string]string{}
    if v := q.Get("service"); v != "" { params["service"] = v }
    if v := q.Get("level"); v != "" { params["level"] = v }
    if v := q.Get("username"); v != "" { params["username"] = v }
    if v := q.Get("is.blacklisted"); v != "" { params["is.blacklisted"] = v }

    limit := 0
    if v := q.Get("limit"); v != "" {
        if i, err := strconv.Atoi(v); err == nil { limit = i }
    }
    sortKey := q.Get("sort")
    res, err := store.Query(params, limit, sortKey)
    if err != nil {
        http.Error(w, "query error", http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type","application/json")
    enc := json.NewEncoder(w)
    enc.Encode(res)
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
    m := map[string]interface{}{
        "total_logs": store.Count(),
        "by_category": store.GroupByCategory(),
        "by_severity": store.GroupBySeverity(),
    }
    w.Header().Set("Content-Type","application/json")
    json.NewEncoder(w).Encode(m)
}

func main() {
    dataPath := os.Getenv("STORAGE_FILE")
    if dataPath == "" { dataPath = "/data/logs.jsonl" }
    os.MkdirAll("/data", 0755)
    fs := NewFileStorage(dataPath)
    store = fs

    http.HandleFunc("/ingest", ingestHandler)
    http.HandleFunc("/logs", logsHandler)
    http.HandleFunc("/metrics", metricsHandler)

    port := os.Getenv("PORT")
    if port == "" { port = "8081" }
    addr := ":" + port
    fmt.Println("log-server listening on", addr)
    log.Fatal(http.ListenAndServe(addr, nil))//default mltiplexar....
}
