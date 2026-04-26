package scanner_test

import (
	"net"
	"testing"
	"time"

	"github.com/yourorg/portwatch/internal/scanner"
)

// startTCPServer starts a local TCP listener and returns its port and a stop func.
func startTCPServer(t *testing.T) (int, func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start test server: %v", err)
	}
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()
	port := ln.Addr().(*net.TCPAddr).Port
	return port, func() { ln.Close() }
}

func TestScanPort_Open(t *testing.T) {
	port, stop := startTCPServer(t)
	defer stop()

	s := scanner.New("127.0.0.1", 2*time.Second)
	status := s.ScanPort(port)

	if !status.Open {
		t.Errorf("expected port %d to be open, got closed", port)
	}
	if status.Err != nil {
		t.Errorf("unexpected error: %v", status.Err)
	}
}

func TestScanPort_Closed(t *testing.T) {
	s := scanner.New("127.0.0.1", 500*time.Millisecond)
	// Port 1 is almost certainly closed/refused in test environments.
	status := s.ScanPort(1)

	if status.Open {
		t.Errorf("expected port 1 to be closed")
	}
	if status.Err == nil {
		t.Errorf("expected an error for closed port")
	}
}

func TestScanPorts_MultipleResults(t *testing.T) {
	port1, stop1 := startTCPServer(t)
	defer stop1()
	port2, stop2 := startTCPServer(t)
	defer stop2()

	s := scanner.New("127.0.0.1", 2*time.Second)
	results := s.ScanPorts([]int{port1, port2})

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if !r.Open {
			t.Errorf("expected port %d to be open", r.Port)
		}
	}
}
