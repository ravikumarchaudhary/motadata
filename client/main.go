package main

// All the things will run on the localhost server
import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"time"
)
// time of event hostnamen source type, category, and actual event message.....
type Msg struct {
	Timestamp     string `json:"timestamp"`
	Hostname      string `json:"hostname"`
	EventType     string `json:"event.source.type"`
	EventCategory string `json:"event.category"`
	Message       string `json:"message"`
}

// Created a slice to containing sample log of messages
var messages = []string{
	"<86> host1 sudo: pam_unix(sudo:session): session opened for user root(uid=0) by motadata(uid=1000)",
	"<86> host2 sshd: Accepted password for alice from 198.51.100.23 port 51234 ssh2",
	"<134> WIN-EQ5V3RA5F7H Microsoft-Windows-Security-Auditing: A user account was successfully logged on. Account Name: Motadata",
	"<86> host3 CRON[1234]: (root) CMD (run-parts /etc/cron.daily)",
	"<86> host1 pam_unix: session closed for user motadata",
}

// random Interval generates a random delay....
func randomInterval() time.Duration {
	return time.Duration(1000+rand.Intn(1000)) * time.Millisecond
}

func main() {
	r := rand.New(rand.NewSource(time.Now().UnixNano())) // Create a new random number generator with a seed based on current time
	collector := os.Getenv("COLLECTOR_ADDR")             // Get the collector server address from environment variable, or use default
	if collector == "" {
		collector = "localhost:6000" // This is the Default collector address
	}
	hostname, _ := os.Hostname() // Get the hostname of the current machine
	flagHostname := flag.String("hostname", hostname, "hostname to identify")
	envCat := flag.String("category", "linux", "event.source.type")
	flag.Parse() // Parse the flags

	// Infinite loop to send log messages continuously....
	for {
		msg := messages[r.Intn(len(messages))] //select random message from the defined list

		// Create a Msg struct with event details like (hostname, eventtype, fixed category random log)
		m := Msg{
			Timestamp:     time.Now().UTC().Format(time.RFC3339),
			Hostname:      *flagHostname,
			EventType:     *envCat,
			EventCategory: "login.audit",
			Message:       msg,
		}
		b, _ := json.Marshal(m)
		conn, err := net.Dial("tcp", collector)
		if err != nil {
			log.Println("dial error:", err)
			time.Sleep(2 * time.Second)
			continue
		}
		fmt.Fprintln(conn, string(b))
		conn.Close()
		time.Sleep(randomInterval()) // Wait for a random interval before sending the next message....
	}
}
