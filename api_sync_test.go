package qbtapi

import (
	"context"
	"testing"
)

func TestSyncDomain(t *testing.T) {
	c := newTestClient(t)
	ctx := context.Background()

	// ── get main data (full update) ─────────────────────────
	mainData, err := c.GetMainData(ctx, 0)
	if err != nil {
		t.Fatalf("GetMainData: %v", err)
	}
	t.Logf("main data: rid=%d full_update=%v torrents=%d categories=%d tags=%d",
		mainData.RID, mainData.FullUpdate, len(mainData.Torrents), len(mainData.Categories), len(mainData.Tags))
	if mainData.ServerState.ConnectionStatus == "" {
		t.Fatal("GetMainData returned empty connection status in server_state")
	}

	// ── incremental main data ───────────────────────────────
	incremental, err := c.GetMainData(ctx, mainData.RID)
	if err != nil {
		t.Fatalf("GetMainData (incremental): %v", err)
	}
	t.Logf("incremental main data: rid=%d full_update=%v", incremental.RID, incremental.FullUpdate)

	// ── get torrent peers data (requires a torrent) ─────────
	list, err := c.GetTorrentList(ctx, nil)
	if err != nil {
		t.Fatalf("GetTorrentList: %v", err)
	}
	if len(list) == 0 {
		t.Skip("no torrents available, skipping torrent peers sync test")
	}
	hash := list[0].Hash

	peersData, err := c.GetTorrentPeersData(ctx, hash, 0)
	if err != nil {
		t.Fatalf("GetTorrentPeersData: %v", err)
	}
	t.Logf("torrent peers data: rid=%d full_update=%v peers=%d",
		peersData.RID, peersData.FullUpdate, len(peersData.Peers))
}
