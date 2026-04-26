// Package scanner provides utilities for performing concurrent TCP port scans
// against a target host. It is used by portwatch to determine which configured
// ports are currently open or closed, feeding results into the alerting pipeline.
//
// Basic usage:
//
//	s := scanner.New("localhost", 2*time.Second)
//
//	// Scan a single port
//	status := s.ScanPort(8080)
//	fmt.Println(status.Open, status.Latency)
//
//	// Scan multiple ports concurrently
//	results := s.ScanPorts([]int{80, 443, 8080, 9090})
//	for _, r := range results {
//		fmt.Printf("port %d open=%v latency=%v\n", r.Port, r.Open, r.Latency)
//	}
package scanner
