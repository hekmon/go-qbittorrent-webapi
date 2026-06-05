package qbtapi

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

func TestTorrentsDomain(t *testing.T) {
	c := newTestClient(t)
	ctx := context.Background()

	// ── list all torrents ───────────────────────────────────
	list, err := c.GetTorrentList(ctx, nil)
	if err != nil {
		t.Fatalf("GetTorrentList (no filters): %v", err)
	}
	t.Logf("torrent count (no filters): %d", len(list))

	for i, ti := range list {
		validateTorrentInfo(t, ti, "list["+string(rune('0'+i))+"]")
	}

	// ── list with filters ───────────────────────────────────
	listFiltered, err := c.GetTorrentList(ctx, &ListFilters{
		State: FilterStateAll.Ptr(),
	})
	if err != nil {
		t.Fatalf("GetTorrentList (with filters): %v", err)
	}
	if len(listFiltered) != len(list) {
		t.Fatalf("filtered list length mismatch: expected %d, got %d", len(list), len(listFiltered))
	}

	// ── hash filter ─────────────────────────────────────────
	if len(list) > 0 {
		hash := list[0].Hash
		byHash, err := c.GetTorrentList(ctx, &ListFilters{Hashes: []string{hash}})
		if err != nil {
			t.Fatalf("GetTorrentList (hash filter): %v", err)
		}
		if len(byHash) != 1 {
			t.Fatalf("hash filter: expected 1 result, got %d", len(byHash))
		}
		if byHash[0].Hash != hash {
			t.Fatalf("hash filter: expected hash %s, got %s", hash, byHash[0].Hash)
		}
	}

	// ── generic properties & trackers (existing torrents) ───
	if len(list) > 0 {
		hash := list[0].Hash

		t.Run("generic_properties_existing", func(t *testing.T) {
			props, err := c.GetTorrentGenericProperties(ctx, hash)
			if err != nil {
				t.Fatalf("GetTorrentGenericProperties: %v", err)
			}
			validateTorrentGenericProperties(t, props, "existing")
			t.Logf("properties for %s: save_path=%s total_size=%s", hash, props.SavePath, props.TotalSize)
		})

		t.Run("trackers_existing", func(t *testing.T) {
			trackers, err := c.GetTorrentTrackers(ctx, hash)
			if err != nil {
				t.Fatalf("GetTorrentTrackers: %v", err)
			}
			if len(trackers) == 0 {
				t.Fatal("expected at least one tracker")
			}
			for i, tr := range trackers {
				validateTorrentTracker(t, tr, "existing["+string(rune('0'+i))+"]")
			}
			t.Logf("trackers for %s: %d", hash, len(trackers))
		})
	}

	// ── add via magnet, inspect, and delete ─────────────────
	t.Run("add_magnet_inspect_delete", func(t *testing.T) {
		const knownHash = "dd8255ecdc7ca55fb0bbf81323d87062db1f6d1c"
		magnet, err := url.Parse("magnet:?xt=urn:btih:" + knownHash + "&dn=Big+Buck+Bunny")
		if err != nil {
			t.Fatalf("parsing magnet link: %v", err)
		}

		// check if the test torrent is already present
		wasPreExisting := false
		for _, ti := range list {
			if ti.Hash == knownHash {
				wasPreExisting = true
				break
			}
		}

		// if pre-existing, remove it so we can test the full lifecycle
		if wasPreExisting {
			if err := c.DeleteTorrents(ctx, []string{knownHash}, false); err != nil {
				t.Fatalf("pre-test cleanup delete failed: %v", err)
			}
		}

		// add paused so we don't actually download anything
		if err := c.AddNewTorrents(ctx, nil, []*url.URL{magnet}, &AddNewTorrentsOptions{
			Paused: Bool(true),
		}); err != nil {
			var httpErr HTTPError
			if errors.As(err, &httpErr) && httpErr == 409 {
				t.Skipf("AddNewTorrents returned 409 (torrent may already exist despite cleanup): %v", err)
			}
			t.Skipf("AddNewTorrents failed (magnet may be unavailable): %v", err)
		}

		// find the newly added torrent
		listAfterAdd, err := c.GetTorrentList(ctx, nil)
		if err != nil {
			t.Fatalf("GetTorrentList (after add): %v", err)
		}

		var testHash string
		for _, ti := range listAfterAdd {
			if ti.Hash == knownHash {
				testHash = ti.Hash
				break
			}
		}
		if testHash == "" {
			t.Fatal("could not find added torrent in list")
		}
		t.Logf("magnet torrent hash: %s", testHash)

		// validate list entry
		for _, ti := range listAfterAdd {
			if ti.Hash == testHash {
				validateTorrentInfo(t, ti, "magnet")
				break
			}
		}

		// inspect generic properties
		props, err := c.GetTorrentGenericProperties(ctx, testHash)
		if err != nil {
			t.Fatalf("GetTorrentGenericProperties (magnet): %v", err)
		}
		// magnet may not have metadata yet, so only validate core fields
		if props.SavePath == "" {
			t.Fatal("GetTorrentGenericProperties (magnet) returned empty save path")
		}
		if props.AdditionDate.IsZero() {
			t.Fatal("GetTorrentGenericProperties (magnet) returned zero addition date")
		}
		t.Logf("magnet torrent save path: %s", props.SavePath)

		// inspect trackers
		trackers, err := c.GetTorrentTrackers(ctx, testHash)
		if err != nil {
			t.Fatalf("GetTorrentTrackers (magnet): %v", err)
		}
		if len(trackers) == 0 {
			t.Fatal("expected at least one tracker for magnet torrent")
		}
		for i, tr := range trackers {
			validateTorrentTracker(t, tr, "magnet["+string(rune('0'+i))+"]")
		}
		t.Logf("magnet torrent trackers: %d", len(trackers))

		// delete
		if err := c.DeleteTorrents(ctx, []string{testHash}, false); err != nil {
			t.Fatalf("DeleteTorrents: %v", err)
		}

		// verify deletion
		listAfterDelete, err := c.GetTorrentList(ctx, nil)
		if err != nil {
			t.Fatalf("GetTorrentList (after delete): %v", err)
		}
		for _, ti := range listAfterDelete {
			if ti.Hash == testHash {
				t.Fatal("torrent still present after deletion")
			}
		}

		// restore the torrent if it was pre-existing
		if wasPreExisting {
			if err := c.AddNewTorrents(ctx, nil, []*url.URL{magnet}, &AddNewTorrentsOptions{
				Paused: Bool(true),
			}); err != nil {
				t.Logf("restoring pre-existing torrent failed: %v", err)
			}
		}
	})

	// ── add via file, inspect, and delete ───────────────────
	t.Run("add_file_inspect_delete", func(t *testing.T) {
		const torrentURL = "https://webtorrent.io/torrents/sintel.torrent"

		// download the torrent file
		resp, err := http.Get(torrentURL)
		if err != nil {
			t.Skipf("downloading torrent file failed: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Skipf("downloading torrent file returned status %d", resp.StatusCode)
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("reading torrent file body: %v", err)
		}

		// write to a temporary file so we can exercise ReadTorrentsFiles
		tmpDir := t.TempDir()
		tmpPath := filepath.Join(tmpDir, "sintel.torrent")
		if err := os.WriteFile(tmpPath, data, 0644); err != nil {
			t.Fatalf("writing temp torrent file: %v", err)
		}

		files, err := ReadTorrentsFiles([]string{tmpPath})
		if err != nil {
			t.Fatalf("ReadTorrentsFiles: %v", err)
		}
		if len(files) != 1 {
			t.Fatalf("expected 1 file, got %d", len(files))
		}

		// capture list before add
		listBefore, err := c.GetTorrentList(ctx, nil)
		if err != nil {
			t.Fatalf("GetTorrentList (before): %v", err)
		}

		// add paused
		if err := c.AddNewTorrents(ctx, files, nil, &AddNewTorrentsOptions{
			Paused: Bool(true),
		}); err != nil {
			var httpErr HTTPError
			if errors.As(err, &httpErr) && httpErr == 409 {
				t.Skipf("AddNewTorrents returned 409 (torrent may already exist): %v", err)
			}
			t.Skipf("AddNewTorrents failed: %v", err)
		}

		// find the newly added torrent
		listAfter, err := c.GetTorrentList(ctx, nil)
		if err != nil {
			t.Fatalf("GetTorrentList (after): %v", err)
		}

		beforeHashes := make(map[string]struct{}, len(listBefore))
		for _, ti := range listBefore {
			beforeHashes[ti.Hash] = struct{}{}
		}

		var testHash string
		for _, ti := range listAfter {
			if _, ok := beforeHashes[ti.Hash]; !ok {
				testHash = ti.Hash
				break
			}
		}
		if testHash == "" {
			t.Fatal("could not find added torrent in list")
		}
		t.Logf("file-added torrent hash: %s", testHash)

		// validate list entry (file torrents have metadata immediately)
		for _, ti := range listAfter {
			if ti.Hash == testHash {
				validateTorrentInfo(t, ti, "file")
				// file torrent should have known size
				if ti.TotalSize.Bytes() <= 0 {
					t.Fatalf("file torrent TotalSize is zero: %v", ti.TotalSize)
				}
				break
			}
		}

		// inspect generic properties
		props, err := c.GetTorrentGenericProperties(ctx, testHash)
		if err != nil {
			t.Fatalf("GetTorrentGenericProperties (file): %v", err)
		}
		validateTorrentGenericProperties(t, props, "file")
		// file torrent must have metadata
		if props.TotalSize.Bytes() <= 0 {
			t.Fatalf("file torrent generic TotalSize is zero: %v", props.TotalSize)
		}
		if props.PiecesNum <= 0 {
			t.Fatalf("file torrent PiecesNum is zero: %d", props.PiecesNum)
		}
		if props.PieceSize.Bytes() <= 0 {
			t.Fatalf("file torrent PieceSize is zero: %v", props.PieceSize)
		}
		t.Logf("file-added torrent save path: %s size: %s pieces: %d", props.SavePath, props.TotalSize, props.PiecesNum)

		// inspect trackers
		trackers, err := c.GetTorrentTrackers(ctx, testHash)
		if err != nil {
			t.Fatalf("GetTorrentTrackers (file): %v", err)
		}
		if len(trackers) == 0 {
			t.Fatal("expected at least one tracker for file torrent")
		}
		for i, tr := range trackers {
			validateTorrentTracker(t, tr, "file["+string(rune('0'+i))+"]")
		}
		t.Logf("file-added torrent trackers: %d", len(trackers))

		// delete
		if err := c.DeleteTorrents(ctx, []string{testHash}, false); err != nil {
			t.Fatalf("DeleteTorrents (file): %v", err)
		}

		// verify deletion
		listFinal, err := c.GetTorrentList(ctx, nil)
		if err != nil {
			t.Fatalf("GetTorrentList (final): %v", err)
		}
		for _, ti := range listFinal {
			if ti.Hash == testHash {
				t.Fatal("file-added torrent still present after deletion")
			}
		}
	})
}

// validateTorrentInfo checks that key fields on a TorrentInfos entry were actually
// unmarshaled (non-zero). A silent json tag typo would leave these empty.
func validateTorrentInfo(t *testing.T, ti TorrentInfos, label string) {
	t.Helper()
	if ti.Hash == "" {
		t.Fatalf("%s: Hash is empty", label)
	}
	if ti.Name == "" {
		t.Fatalf("%s: Name is empty", label)
	}
	if string(ti.State) == "" {
		t.Fatalf("%s: State is empty", label)
	}
	if ti.SavePath == "" {
		t.Fatalf("%s: SavePath is empty", label)
	}
	if ti.AddedOn.IsZero() {
		t.Fatalf("%s: AddedOn is zero", label)
	}
}

// validateTorrentGenericProperties checks core fields that should always be populated.
func validateTorrentGenericProperties(t *testing.T, props TorrentGenericProperties, label string) {
	t.Helper()
	if props.SavePath == "" {
		t.Fatalf("%s: SavePath is empty", label)
	}
	if props.AdditionDate.IsZero() {
		t.Fatalf("%s: AdditionDate is zero", label)
	}
	if props.CreationDate.IsZero() {
		t.Fatalf("%s: CreationDate is zero", label)
	}
}

// validateTorrentTracker checks that tracker fields were unmarshaled correctly.
func validateTorrentTracker(t *testing.T, tr TorrentTracker, label string) {
	t.Helper()
	if tr.URL == nil {
		t.Fatalf("%s: URL is nil", label)
	}
	if tr.URL.String() == "" {
		t.Fatalf("%s: URL is empty", label)
	}
}
