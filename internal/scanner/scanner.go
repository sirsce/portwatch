package scanner

import (
	"fmt"
	"net"
	"time"
)

// PortStatus represents the result of scanning a single port.
type PortStatus struct {
	Port    int
	Open    bool
	Latency time.Duration
	Err     error
}

// Scanner performs TCP port scans on a given host.
type Scanner struct {
	Host    string
	Timeout time.Duration
}

// New creates a new Scanner for the given host with the specified timeout.
func New(host string, timeout time.Duration) *Scanner {
	return &Scanner{
		Host:    host,
		Timeout: timeout,
	}
}

// ScanPort checks whether a single TCP port is open on the scanner's host.
func (s *Scanner) ScanPort(port int) PortStatus {
	address := fmt.Sprintf("%s:%d", s.Host, port)
	start := time.Now()

	conn, err := net.DialTimeout("tcp", address, s.Timeout)
	latency := time.Since(start)

	if err != nil {
		return PortStatus{Port: port, Open: false, Latency: latency, Err: err}
	}
	conn.Close()
	return PortStatus{Port: port, Open: true, Latency: latency}
}

// ScanPorts scans a list of ports concurrently and returns a slice of PortStatus.
func (s *Scanner) ScanPorts(ports []int) []PortStatus {
	results := make([]PortStatus, len(ports))
	ch := make(chan struct {
		idx    int
		status PortStatus
	}, len(ports))

	for i, port := range ports {
		go func(idx, p int) {
			ch <- struct {
				idx    int
				status PortStatus
			}{idx, s.ScanPort(p)}
		}(i, port)
	}

	for range ports {
		res := <-ch
		results[res.idx] = res.status
	}

	return results
}
