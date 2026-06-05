package qbtapi

import (
	"context"
	"testing"
)

func TestSearchDomain(t *testing.T) {
	c := newTestClient(t)
	ctx := context.Background()

	// Clean up any leftover search jobs from previous runs
	jobs, _ := c.GetSearchStatus(ctx, nil)
	for _, job := range jobs {
		_ = c.DeleteSearch(ctx, job.ID)
	}

	// ── get search plugins ──────────────────────────────────
	plugins, err := c.GetSearchPlugins(ctx)
	if err != nil {
		t.Fatalf("GetSearchPlugins: %v", err)
	}
	t.Logf("search plugins: %d", len(plugins))
	for i, p := range plugins {
		if i >= 3 {
			break
		}
		t.Logf("plugin[%d]: %s (%s) enabled=%v", i, p.Name, p.FullName, p.Enabled)
	}

	// ── search job lifecycle (requires at least one plugin) ─
	if len(plugins) == 0 {
		t.Skip("no search plugins available, skipping search job tests")
	}

	jobID, err := c.StartSearch(ctx, "ubuntu", "all", "all")
	if err != nil {
		t.Fatalf("StartSearch: %v", err)
	}
	t.Logf("started search job: %d", jobID)

	// Get status for specific job
	status, err := c.GetSearchStatus(ctx, &jobID)
	if err != nil {
		t.Fatalf("GetSearchStatus: %v", err)
	}
	if len(status) != 1 {
		t.Fatalf("expected 1 status entry, got %d", len(status))
	}
	if status[0].ID != jobID {
		t.Fatalf("status ID mismatch: expected %d, got %d", jobID, status[0].ID)
	}
	t.Logf("search status: %s, total: %d", status[0].Status, status[0].Total)

	// Get all statuses (should include our job)
	allStatus, err := c.GetSearchStatus(ctx, nil)
	if err != nil {
		t.Fatalf("GetSearchStatus (all): %v", err)
	}
	found := false
	for _, s := range allStatus {
		if s.ID == jobID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("job %d not found in global status list", jobID)
	}

	// Get results
	results, err := c.GetSearchResults(ctx, jobID, nil, nil)
	if err != nil {
		t.Fatalf("GetSearchResults: %v", err)
	}
	t.Logf("search results: %d returned (status: %s, total: %d)", len(results.Results), results.Status, results.Total)

	// Stop search
	if err := c.StopSearch(ctx, jobID); err != nil {
		t.Fatalf("StopSearch: %v", err)
	}

	// Delete search
	if err := c.DeleteSearch(ctx, jobID); err != nil {
		t.Fatalf("DeleteSearch: %v", err)
	}

	// Verify deletion
	allStatus, err = c.GetSearchStatus(ctx, nil)
	if err != nil {
		t.Fatalf("GetSearchStatus (after delete): %v", err)
	}
	for _, s := range allStatus {
		if s.ID == jobID {
			t.Fatalf("job %d still exists after deletion", jobID)
		}
	}

	// ── plugin management (safe ops) ────────────────────────
	if err := c.UpdateSearchPlugins(ctx); err != nil {
		t.Logf("UpdateSearchPlugins: %v", err)
	}
}
