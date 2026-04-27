// Package monitor watches a set of TCP ports on the local host at a
// configurable interval and tracks state transitions (open ↔ closed).
//
// Core types:
//
//   - Monitor   – drives the scan loop and emits alerts via the dispatcher.
//   - History   – records per-port state transitions with eviction.
//   - Report    – aggregates current port states into a human-readable summary.
//   - Snapshot  – a point-in-time view of all port states that can be
//     persisted to disk and reloaded across daemon restarts via
//     SnapshotStore.
package monitor
