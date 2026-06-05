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

		// ── extended endpoint tests ─────────────────────────────
		findTorrent := func(list []TorrentInfos, h string) *TorrentInfos {
			for i := range list {
				if list[i].Hash == h {
					return &list[i]
				}
			}
			return nil
		}

		// web seeds
		if seeds, err := c.GetTorrentWebSeeds(ctx, testHash); err != nil {
			t.Fatalf("GetTorrentWebSeeds: %v", err)
		} else {
			t.Logf("web seeds: %d", len(seeds))
		}

		// contents
		contents, err := c.GetTorrentContents(ctx, testHash, nil)
		if err != nil {
			t.Fatalf("GetTorrentContents: %v", err)
		}
		if len(contents) == 0 {
			t.Fatal("expected at least one file in torrent contents")
		}
		for i, f := range contents {
			if f.Name == "" {
				t.Fatalf("content[%d]: Name is empty", i)
			}
			if f.Size <= 0 {
				t.Fatalf("content[%d]: Size is zero", i)
			}
		}
		t.Logf("torrent contents: %d files", len(contents))

		// pieces states & hashes
		states, err := c.GetTorrentPiecesStates(ctx, testHash)
		if err != nil {
			t.Fatalf("GetTorrentPiecesStates: %v", err)
		}
		if len(states) == 0 {
			t.Fatal("expected at least one piece state")
		}
		pieceHashes, err := c.GetTorrentPiecesHashes(ctx, testHash)
		if err != nil {
			t.Fatalf("GetTorrentPiecesHashes: %v", err)
		}
		if len(pieceHashes) == 0 {
			t.Fatal("expected at least one piece hash")
		}
		if len(pieceHashes) != len(states) {
			t.Fatalf("piece hashes count (%d) != piece states count (%d)", len(pieceHashes), len(states))
		}
		t.Logf("pieces: %d states, %d hashes", len(states), len(pieceHashes))

		// start torrent (currently paused) and verify state changed
		if err := c.StartTorrents(ctx, []string{testHash}); err != nil {
			t.Fatalf("StartTorrents: %v", err)
		}
		listAfterStart, err := c.GetTorrentList(ctx, nil)
		if err != nil {
			t.Fatalf("GetTorrentList (after start): %v", err)
		}
		ti := findTorrent(listAfterStart, testHash)
		if ti == nil {
			t.Fatal("torrent not found after start")
		}
		if ti.State == TorrentStatePausedDownloading || ti.State == TorrentStatePausedUploading {
			t.Fatalf("expected torrent to be running after StartTorrents, got state %q", ti.State)
		}

		// stop torrent and verify state is paused
		if err := c.StopTorrents(ctx, []string{testHash}); err != nil {
			t.Fatalf("StopTorrents: %v", err)
		}
		listAfterStop, err := c.GetTorrentList(ctx, nil)
		if err != nil {
			t.Fatalf("GetTorrentList (after stop): %v", err)
		}
		ti = findTorrent(listAfterStop, testHash)
		if ti == nil {
			t.Fatal("torrent not found after stop")
		}
		if ti.State != TorrentStatePausedDownloading && ti.State != TorrentStatePausedUploading {
			t.Fatalf("expected torrent to be paused after StopTorrents, got state %q", ti.State)
		}

		// download / upload limits (set, verify, reset)
		limit := GetSpeedFromBytes(1024 * 1024) // 1 MiB/s
		if err := c.SetTorrentDownloadLimit(ctx, []string{testHash}, limit); err != nil {
			t.Fatalf("SetTorrentDownloadLimit: %v", err)
		}
		dlLimits, err := c.GetTorrentDownloadLimit(ctx, []string{testHash})
		if err != nil {
			t.Fatalf("GetTorrentDownloadLimit: %v", err)
		}
		if dlLimits[testHash] != limit.ToBytes() {
			t.Fatalf("download limit mismatch: expected %d, got %d", limit.ToBytes(), dlLimits[testHash])
		}
		if err := c.SetTorrentUploadLimit(ctx, []string{testHash}, limit); err != nil {
			t.Fatalf("SetTorrentUploadLimit: %v", err)
		}
		ulLimits, err := c.GetTorrentUploadLimit(ctx, []string{testHash})
		if err != nil {
			t.Fatalf("GetTorrentUploadLimit: %v", err)
		}
		if ulLimits[testHash] != limit.ToBytes() {
			t.Fatalf("upload limit mismatch: expected %d, got %d", limit.ToBytes(), ulLimits[testHash])
		}
		// reset to unlimited
		if err := c.SetTorrentDownloadLimit(ctx, []string{testHash}, UnlimitedSpeedLimit); err != nil {
			t.Fatalf("SetTorrentDownloadLimit (reset): %v", err)
		}
		if err := c.SetTorrentUploadLimit(ctx, []string{testHash}, UnlimitedSpeedLimit); err != nil {
			t.Fatalf("SetTorrentUploadLimit (reset): %v", err)
		}

		// share limits
		if err := c.SetTorrentShareLimits(ctx, []string{testHash}, 2.0, 60, -1); err != nil {
			t.Fatalf("SetTorrentShareLimits: %v", err)
		}

		// category lifecycle
		catName := "testcat_" + testHash[:8]
		catPath := t.TempDir()
		if err := c.CreateCategory(ctx, catName, catPath); err != nil {
			t.Fatalf("CreateCategory: %v", err)
		}
		if err := c.SetTorrentCategory(ctx, []string{testHash}, catName); err != nil {
			t.Fatalf("SetTorrentCategory: %v", err)
		}
		listByCat, err := c.GetTorrentList(ctx, &ListFilters{Category: &catName})
		if err != nil {
			t.Fatalf("GetTorrentList (by category): %v", err)
		}
		if len(listByCat) != 1 || listByCat[0].Hash != testHash {
			t.Fatalf("category filter: expected 1 torrent with hash %s, got %+v", testHash, listByCat)
		}
		newCatPath := t.TempDir()
		if err := c.EditCategory(ctx, catName, newCatPath); err != nil {
			t.Fatalf("EditCategory: %v", err)
		}
		cats, err := c.GetAllCategories(ctx)
		if err != nil {
			t.Fatalf("GetAllCategories: %v", err)
		}
		if cat, ok := cats[catName]; !ok || cat.SavePath != newCatPath {
			t.Fatalf("category edit not reflected: got %+v", cat)
		}
		if err := c.RemoveCategories(ctx, []string{catName}); err != nil {
			t.Fatalf("RemoveCategories: %v", err)
		}

		// tags lifecycle
		tagName := "testtag_" + testHash[:8]
		if err := c.CreateTags(ctx, []string{tagName}); err != nil {
			t.Fatalf("CreateTags: %v", err)
		}
		if err := c.AddTorrentTags(ctx, []string{testHash}, []string{tagName}); err != nil {
			t.Fatalf("AddTorrentTags: %v", err)
		}
		listWithTag, err := c.GetTorrentList(ctx, &ListFilters{Tag: &tagName})
		if err != nil {
			t.Fatalf("GetTorrentList (by tag): %v", err)
		}
		if len(listWithTag) != 1 || listWithTag[0].Hash != testHash {
			t.Fatalf("tag filter: expected 1 torrent with hash %s, got %+v", testHash, listWithTag)
		}
		if err := c.RemoveTorrentTags(ctx, []string{testHash}, []string{tagName}); err != nil {
			t.Fatalf("RemoveTorrentTags: %v", err)
		}
		listWithoutTag, err := c.GetTorrentList(ctx, &ListFilters{Tag: &tagName})
		if err != nil {
			t.Fatalf("GetTorrentList (after tag removal): %v", err)
		}
		if len(listWithoutTag) != 0 {
			t.Fatalf("expected 0 torrents with tag %q after removal, got %d", tagName, len(listWithoutTag))
		}
		if err := c.DeleteTags(ctx, []string{tagName}); err != nil {
			t.Fatalf("DeleteTags: %v", err)
		}

		// rename torrent and verify
		if err := c.RenameTorrent(ctx, testHash, "sintel_test"); err != nil {
			t.Fatalf("RenameTorrent: %v", err)
		}
		listAfterRename, err := c.GetTorrentList(ctx, nil)
		if err != nil {
			t.Fatalf("GetTorrentList (after rename): %v", err)
		}
		ti = findTorrent(listAfterRename, testHash)
		if ti == nil || ti.Name != "sintel_test" {
			t.Fatalf("rename not reflected: expected Name=%q, got %q", "sintel_test", ti.Name)
		}

		// toggles with verification
		// sequential download
		origSeq := ti.SequentialDownload
		if err := c.ToggleSequentialDownload(ctx, []string{testHash}); err != nil {
			t.Fatalf("ToggleSequentialDownload: %v", err)
		}
		listAfterToggle, err := c.GetTorrentList(ctx, nil)
		if err != nil {
			t.Fatalf("GetTorrentList (after seq toggle): %v", err)
		}
		ti = findTorrent(listAfterToggle, testHash)
		if ti.SequentialDownload == origSeq {
			t.Fatalf("sequential download toggle not reflected")
		}
		// toggle back
		if err := c.ToggleSequentialDownload(ctx, []string{testHash}); err != nil {
			t.Fatalf("ToggleSequentialDownload (back): %v", err)
		}

		// first/last piece priority
		origFL := ti.FirstLastPiecePrio
		if err := c.ToggleFirstLastPiecePrio(ctx, []string{testHash}); err != nil {
			t.Fatalf("ToggleFirstLastPiecePrio: %v", err)
		}
		listAfterFL, err := c.GetTorrentList(ctx, nil)
		if err != nil {
			t.Fatalf("GetTorrentList (after fl toggle): %v", err)
		}
		ti = findTorrent(listAfterFL, testHash)
		if ti.FirstLastPiecePrio == origFL {
			t.Fatalf("first/last piece prio toggle not reflected")
		}
		// toggle back
		if err := c.ToggleFirstLastPiecePrio(ctx, []string{testHash}); err != nil {
			t.Fatalf("ToggleFirstLastPiecePrio (back): %v", err)
		}

		// force start
		if err := c.SetForceStart(ctx, []string{testHash}, true); err != nil {
			t.Fatalf("SetForceStart: %v", err)
		}
		listAfterFS, err := c.GetTorrentList(ctx, nil)
		if err != nil {
			t.Fatalf("GetTorrentList (after force start): %v", err)
		}
		ti = findTorrent(listAfterFS, testHash)
		if !ti.ForceStart {
			t.Fatalf("force start not reflected")
		}
		if err := c.SetForceStart(ctx, []string{testHash}, false); err != nil {
			t.Fatalf("SetForceStart (reset): %v", err)
		}

		// auto management
		if err := c.SetAutoManagement(ctx, []string{testHash}, false); err != nil {
			t.Fatalf("SetAutoManagement: %v", err)
		}
		listAfterAM, err := c.GetTorrentList(ctx, nil)
		if err != nil {
			t.Fatalf("GetTorrentList (after auto mgmt): %v", err)
		}
		ti = findTorrent(listAfterAM, testHash)
		if ti.AutoTMM {
			t.Fatalf("auto management disable not reflected")
		}
		if err := c.SetAutoManagement(ctx, []string{testHash}, true); err != nil {
			t.Fatalf("SetAutoManagement (reset): %v", err)
		}

		// super seeding
		if err := c.SetSuperSeeding(ctx, []string{testHash}, true); err != nil {
			t.Fatalf("SetSuperSeeding: %v", err)
		}
		listAfterSS, err := c.GetTorrentList(ctx, nil)
		if err != nil {
			t.Fatalf("GetTorrentList (after super seeding): %v", err)
		}
		ti = findTorrent(listAfterSS, testHash)
		if !ti.SuperSeeding {
			t.Fatalf("super seeding not reflected")
		}
		if err := c.SetSuperSeeding(ctx, []string{testHash}, false); err != nil {
			t.Fatalf("SetSuperSeeding (reset): %v", err)
		}

		// location
		newLoc := t.TempDir()
		if err := c.SetTorrentLocation(ctx, []string{testHash}, newLoc); err != nil {
			t.Fatalf("SetTorrentLocation: %v", err)
		}

		// recheck / reannounce
		if err := c.RecheckTorrents(ctx, []string{testHash}); err != nil {
			t.Fatalf("RecheckTorrents: %v", err)
		}
		if err := c.ReannounceTorrents(ctx, []string{testHash}); err != nil {
			t.Fatalf("ReannounceTorrents: %v", err)
		}

		// file priority
		if len(contents) > 0 {
			firstID := contents[0].Index
			if err := c.SetFilePriority(ctx, testHash, []int{firstID}, 7); err != nil {
				t.Fatalf("SetFilePriority: %v", err)
			}
			contentsAfter, err := c.GetTorrentContents(ctx, testHash, nil)
			if err != nil {
				t.Fatalf("GetTorrentContents (after prio): %v", err)
			}
			var found bool
			for _, f := range contentsAfter {
				if f.Index == firstID {
					found = true
					if f.Priority != 7 {
						t.Fatalf("file priority not reflected: expected 7, got %d", f.Priority)
					}
					break
				}
			}
			if !found {
				t.Fatalf("file with index %d not found after priority change", firstID)
			}
		}

		// trackers add / edit / remove
		origTrackers, err := c.GetTorrentTrackers(ctx, testHash)
		if err != nil {
			t.Fatalf("GetTorrentTrackers (before): %v", err)
		}
		dummy := "http://127.0.0.1:9999/announce"
		if err := c.AddTrackers(ctx, testHash, []string{dummy}); err != nil {
			t.Fatalf("AddTrackers: %v", err)
		}
		trackersAfterAdd, err := c.GetTorrentTrackers(ctx, testHash)
		if err != nil {
			t.Fatalf("GetTorrentTrackers (after add): %v", err)
		}
		if len(trackersAfterAdd) != len(origTrackers)+1 {
			t.Fatalf("expected %d trackers after add, got %d", len(origTrackers)+1, len(trackersAfterAdd))
		}
		newDummy := "http://127.0.0.1:9998/announce"
		if err := c.EditTracker(ctx, testHash, dummy, newDummy); err != nil {
			t.Fatalf("EditTracker: %v", err)
		}
		trackersAfterEdit, err := c.GetTorrentTrackers(ctx, testHash)
		if err != nil {
			t.Fatalf("GetTorrentTrackers (after edit): %v", err)
		}
		foundEdited := false
		for _, tr := range trackersAfterEdit {
			if tr.URL.String() == newDummy {
				foundEdited = true
				break
			}
		}
		if !foundEdited {
			t.Fatalf("edited tracker %q not found", newDummy)
		}
		if err := c.RemoveTrackers(ctx, testHash, []string{newDummy}); err != nil {
			t.Fatalf("RemoveTrackers: %v", err)
		}
		trackersAfterRemove, err := c.GetTorrentTrackers(ctx, testHash)
		if err != nil {
			t.Fatalf("GetTorrentTrackers (after remove): %v", err)
		}
		if len(trackersAfterRemove) != len(origTrackers) {
			t.Fatalf("expected %d trackers after remove, got %d", len(origTrackers), len(trackersAfterRemove))
		}

		// queue priority
		if err := c.IncreaseTorrentPriority(ctx, []string{testHash}); err != nil {
			var he HTTPError
			if errors.As(err, &he) && he == 409 {
				t.Logf("torrent queueing not enabled, skipping priority tests")
			} else {
				t.Fatalf("IncreaseTorrentPriority: %v", err)
			}
		} else {
			if err := c.DecreaseTorrentPriority(ctx, []string{testHash}); err != nil {
				t.Fatalf("DecreaseTorrentPriority: %v", err)
			}
			if err := c.TopTorrentPriority(ctx, []string{testHash}); err != nil {
				t.Fatalf("TopTorrentPriority: %v", err)
			}
			if err := c.BottomTorrentPriority(ctx, []string{testHash}); err != nil {
				t.Fatalf("BottomTorrentPriority: %v", err)
			}
		}

		// add peers
		if err := c.AddPeers(ctx, []string{testHash}, []string{"127.0.0.1:12345"}); err != nil {
			var he HTTPError
			if errors.As(err, &he) && he == 400 {
				t.Logf("fake peer rejected by server, skipping")
			} else {
				t.Fatalf("AddPeers: %v", err)
			}
		}

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
