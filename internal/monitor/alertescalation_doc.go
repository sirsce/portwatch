// Package monitor provides port monitoring primitives for portwatch.
//
// # Alert Escalation
//
// AlertEscalation tracks how many consecutive alerts have fired for a given
// port and promotes the notification to a higher-priority level once a
// configured threshold is crossed.
//
// Escalation levels are evaluated from highest threshold to lowest so that
// the most severe applicable level is always returned. Each level carries its
// own cooldown to prevent alert storms at that tier.
//
// Typical usage:
//
//	levels := []monitor.EscalationLevel{
//		{Name: "warn",     Threshold: 2, Cooldown: 10 * time.Minute},
//		{Name: "critical", Threshold: 5, Cooldown:  5 * time.Minute},
//	}
//	esc := monitor.NewAlertEscalation(levels)
//
//	if lvl, err := esc.Record(port); err == nil {
//		log.Printf("escalated to %s for port %d", lvl.Name, port)
//	}
//
// Call Reset to clear state when a port returns to a healthy condition.
package monitor
