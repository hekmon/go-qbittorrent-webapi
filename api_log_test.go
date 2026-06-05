package qbtapi

import (
	"context"
	"testing"
)

func TestLogDomain(t *testing.T) {
	c := newTestClient(t)
	ctx := context.Background()

	// ── get log (no filters) ────────────────────────────────
	entries, err := c.GetLog(ctx, nil)
	if err != nil {
		t.Fatalf("GetLog: %v", err)
	}
	t.Logf("log entries: %d", len(entries))
	for i, e := range entries {
		if i >= 3 {
			break
		}
		t.Logf("log[%d]: id=%d type=%d ts=%v msg=%q", i, e.ID, e.Type, e.Timestamp, e.Message)
	}

	// ── get log with filters ────────────────────────────────
	filtered, err := c.GetLog(ctx, &LogFilters{
		Normal:   Bool(true),
		Info:     Bool(false),
		Warning:  Bool(false),
		Critical: Bool(false),
	})
	if err != nil {
		t.Fatalf("GetLog (filtered): %v", err)
	}
	t.Logf("filtered log entries (normal only): %d", len(filtered))
	for _, e := range filtered {
		if e.Type != LogMessageTypeNormal {
			t.Fatalf("expected only normal messages, got type %d", e.Type)
		}
	}

	// ── get peer log ────────────────────────────────────────
	peerEntries, err := c.GetPeerLog(ctx, nil)
	if err != nil {
		t.Fatalf("GetPeerLog: %v", err)
	}
	t.Logf("peer log entries: %d", len(peerEntries))
}
