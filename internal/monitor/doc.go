// Package monitor provides the core port-monitoring loop for portwatch.
//
// It periodically scans a configured list of TCP ports on a target host,
// tracks state transitions using a History, and dispatches alerts via the
// alert.Dispatcher when a port's availability changes.
//
// # Components
//
//   - Monitor: orchestrates the scan loop and wires together the scanner,
//     history, and dispatcher.
//   - History: records per-port boolean states and surfaces transitions so
//     that alerts are only emitted on change.
//   - Report: a point-in-time snapshot of all monitored ports, built from
//     the latest History entries; useful for status endpoints or log lines.
package monitor
