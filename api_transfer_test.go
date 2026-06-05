package qbtapi

import (
	"context"
	"testing"
)

func TestTransferDomain(t *testing.T) {
	c := newTestClient(t)
	ctx := context.Background()

	// ── global transfer info ────────────────────────────────
	info, err := c.GetGlobalTransferInfo(ctx)
	if err != nil {
		t.Fatalf("GetGlobalTransferInfo: %v", err)
	}
	if info.ConnectionStatus == "" {
		t.Fatal("GetGlobalTransferInfo returned empty connection status")
	}
	t.Logf("transfer info: dl=%d bytes/s, ul=%d bytes/s, dht=%d, status=%s",
		info.DlInfoSpeed, info.UpInfoSpeed, info.DHTNodes, info.ConnectionStatus)

	// ── alternative speed limits toggle round-trip ──────────
	originalAltState, err := c.GetAlternativeSpeedLimitsState(ctx)
	if err != nil {
		t.Fatalf("GetAlternativeSpeedLimitsState: %v", err)
	}
	t.Logf("alternative speed limits original state: %v", originalAltState)

	if err := c.ToggleAlternativeSpeedLimits(ctx); err != nil {
		t.Fatalf("ToggleAlternativeSpeedLimits: %v", err)
	}

	altStateAfterToggle, err := c.GetAlternativeSpeedLimitsState(ctx)
	if err != nil {
		t.Fatalf("GetAlternativeSpeedLimitsState (after toggle): %v", err)
	}
	if altStateAfterToggle == originalAltState {
		t.Fatalf("alternative speed limits state did not toggle: expected %v, got %v",
			!originalAltState, altStateAfterToggle)
	}

	// restore
	if err := c.ToggleAlternativeSpeedLimits(ctx); err != nil {
		t.Fatalf("ToggleAlternativeSpeedLimits (restore): %v", err)
	}
	altStateRestored, err := c.GetAlternativeSpeedLimitsState(ctx)
	if err != nil {
		t.Fatalf("GetAlternativeSpeedLimitsState (after restore): %v", err)
	}
	if altStateRestored != originalAltState {
		t.Fatalf("alternative speed limits state not restored: expected %v, got %v",
			originalAltState, altStateRestored)
	}

	// ── global download limit round-trip ────────────────────
	originalDlLimit, err := c.GetGlobalDownloadLimit(ctx)
	if err != nil {
		t.Fatalf("GetGlobalDownloadLimit: %v", err)
	}
	t.Logf("global download limit original: %d bytes/s", originalDlLimit)

	testDlLimit := 102400 // 100 KiB/s
	if originalDlLimit == testDlLimit {
		testDlLimit = 204800
	}

	if err := c.SetGlobalDownloadLimit(ctx, testDlLimit); err != nil {
		t.Fatalf("SetGlobalDownloadLimit: %v", err)
	}

	dlLimitAfterSet, err := c.GetGlobalDownloadLimit(ctx)
	if err != nil {
		t.Fatalf("GetGlobalDownloadLimit (after set): %v", err)
	}
	if dlLimitAfterSet != testDlLimit {
		t.Fatalf("global download limit mismatch: expected %d, got %d",
			testDlLimit, dlLimitAfterSet)
	}

	// restore
	if err := c.SetGlobalDownloadLimit(ctx, originalDlLimit); err != nil {
		t.Fatalf("SetGlobalDownloadLimit (restore): %v", err)
	}
	dlLimitRestored, err := c.GetGlobalDownloadLimit(ctx)
	if err != nil {
		t.Fatalf("GetGlobalDownloadLimit (after restore): %v", err)
	}
	if dlLimitRestored != originalDlLimit {
		t.Fatalf("global download limit not restored: expected %d, got %d",
			originalDlLimit, dlLimitRestored)
	}

	// ── global upload limit round-trip ──────────────────────
	originalUlLimit, err := c.GetGlobalUploadLimit(ctx)
	if err != nil {
		t.Fatalf("GetGlobalUploadLimit: %v", err)
	}
	t.Logf("global upload limit original: %d bytes/s", originalUlLimit)

	testUlLimit := 102400 // 100 KiB/s
	if originalUlLimit == testUlLimit {
		testUlLimit = 204800
	}

	if err := c.SetGlobalUploadLimit(ctx, testUlLimit); err != nil {
		t.Fatalf("SetGlobalUploadLimit: %v", err)
	}

	ulLimitAfterSet, err := c.GetGlobalUploadLimit(ctx)
	if err != nil {
		t.Fatalf("GetGlobalUploadLimit (after set): %v", err)
	}
	if ulLimitAfterSet != testUlLimit {
		t.Fatalf("global upload limit mismatch: expected %d, got %d",
			testUlLimit, ulLimitAfterSet)
	}

	// restore
	if err := c.SetGlobalUploadLimit(ctx, originalUlLimit); err != nil {
		t.Fatalf("SetGlobalUploadLimit (restore): %v", err)
	}
	ulLimitRestored, err := c.GetGlobalUploadLimit(ctx)
	if err != nil {
		t.Fatalf("GetGlobalUploadLimit (after restore): %v", err)
	}
	if ulLimitRestored != originalUlLimit {
		t.Fatalf("global upload limit not restored: expected %d, got %d",
			originalUlLimit, ulLimitRestored)
	}

	// ── ban peers ───────────────────────────────────────────
	// The API returns 200 for all scenarios, so we can test with a dummy peer.
	if err := c.BanPeers(ctx, []string{"1.2.3.4:5678"}); err != nil {
		t.Fatalf("BanPeers: %v", err)
	}
}
