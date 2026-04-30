// Package monitor provides port monitoring, health checking, and alerting
// primitives for the portwatch daemon.
//
// # AlertBatch
//
// AlertBatch collects individual alert events (port + state label) within a
// configurable time window and delivers them as a single grouped batch to a
// flush callback. This reduces notification noise when multiple ports change
// state in rapid succession (e.g. during a network partition or host restart).
//
// Events are flushed either:
//   - when the window duration elapses after the first Add call, or
//   - immediately when the accumulated event count reaches maxSize, or
//   - explicitly via Flush.
//
// The flush callback is invoked in a separate goroutine to avoid blocking the
// caller. AlertBatch is safe for concurrent use.
package monitor
