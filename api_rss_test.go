package qbtapi

import (
	"context"
	"testing"
)

func TestRSSDomain(t *testing.T) {
	c := newTestClient(t)
	ctx := context.Background()

	// Cleanup leftovers from previous runs
	_ = c.RemoveRSSItem(ctx, "test-folder")
	_ = c.RemoveRSSItem(ctx, "test-folder-moved")
	_ = c.RemoveRSSItem(ctx, "test-feed")
	_ = c.RemoveRSSAutoDownloadingRule(ctx, "test-rule")
	_ = c.RemoveRSSAutoDownloadingRule(ctx, "test-rule-renamed")

	// ── add folder ──────────────────────────────────────────
	if err := c.AddRSSFolder(ctx, "test-folder"); err != nil {
		t.Fatalf("AddRSSFolder: %v", err)
	}

	// ── get all items (verify folder exists) ────────────────
	items, err := c.GetAllRSSItems(ctx, nil)
	if err != nil {
		t.Fatalf("GetAllRSSItems: %v", err)
	}
	if _, ok := items["test-folder"]; !ok {
		t.Fatalf("folder 'test-folder' not found in RSS items")
	}
	t.Logf("RSS items count: %d", len(items))

	// ── move folder ─────────────────────────────────────────
	if err := c.MoveRSSItem(ctx, "test-folder", "test-folder-moved"); err != nil {
		t.Fatalf("MoveRSSItem: %v", err)
	}
	items, err = c.GetAllRSSItems(ctx, nil)
	if err != nil {
		t.Fatalf("GetAllRSSItems (after move): %v", err)
	}
	if _, ok := items["test-folder"]; ok {
		t.Fatalf("old folder 'test-folder' still exists after move")
	}
	if _, ok := items["test-folder-moved"]; !ok {
		t.Fatalf("new folder 'test-folder-moved' not found after move")
	}

	// ── remove folder ───────────────────────────────────────
	if err := c.RemoveRSSItem(ctx, "test-folder-moved"); err != nil {
		t.Fatalf("RemoveRSSItem: %v", err)
	}
	items, err = c.GetAllRSSItems(ctx, nil)
	if err != nil {
		t.Fatalf("GetAllRSSItems (after remove): %v", err)
	}
	if _, ok := items["test-folder-moved"]; ok {
		t.Fatalf("folder 'test-folder-moved' still exists after removal")
	}

	// ── feed operations (best-effort) ───────────────────────
	feedURL := "https://releases.ubuntu.com/releases.rss"
	feedPath := "test-feed"
	if err := c.AddRSSFeed(ctx, feedURL, &feedPath); err != nil {
		t.Logf("AddRSSFeed skipped (might require valid feed): %v", err)
	} else {
		if err := c.RefreshRSSItem(ctx, feedPath); err != nil {
			t.Logf("RefreshRSSItem: %v", err)
		}
		if err := c.MarkRSSItemAsRead(ctx, feedPath, nil); err != nil {
			t.Logf("MarkRSSItemAsRead: %v", err)
		}
		if err := c.RemoveRSSItem(ctx, feedPath); err != nil {
			t.Fatalf("RemoveRSSItem (feed): %v", err)
		}
	}

	// ── auto-downloading rules CRUD ─────────────────────────
	ruleName := "test-rule"
	rule := RSSAutoDownloadingRule{
		Enabled:          true,
		MustContain:      "test",
		AddPaused:        true,
		AssignedCategory: "",
		SavePath:         "",
	}

	if err := c.SetRSSAutoDownloadingRule(ctx, ruleName, rule); err != nil {
		t.Fatalf("SetRSSAutoDownloadingRule: %v", err)
	}

	rules, err := c.GetAllRSSAutoDownloadingRules(ctx)
	if err != nil {
		t.Fatalf("GetAllRSSAutoDownloadingRules: %v", err)
	}
	if r, ok := rules[ruleName]; !ok {
		t.Fatalf("rule %q not found after creation", ruleName)
	} else if !r.Enabled {
		t.Fatalf("rule %q should be enabled", ruleName)
	}

	// Test matching articles (likely empty, just verify no error)
	articles, err := c.GetAllArticlesMatchingRule(ctx, ruleName)
	if err != nil {
		t.Fatalf("GetAllArticlesMatchingRule: %v", err)
	}
	t.Logf("articles matching rule: %d feeds", len(articles))

	newRuleName := "test-rule-renamed"
	if err := c.RenameRSSAutoDownloadingRule(ctx, ruleName, newRuleName); err != nil {
		t.Fatalf("RenameRSSAutoDownloadingRule: %v", err)
	}

	rules, err = c.GetAllRSSAutoDownloadingRules(ctx)
	if err != nil {
		t.Fatalf("GetAllRSSAutoDownloadingRules (after rename): %v", err)
	}
	if _, ok := rules[ruleName]; ok {
		t.Fatalf("old rule name %q still exists", ruleName)
	}
	if _, ok := rules[newRuleName]; !ok {
		t.Fatalf("new rule name %q not found after rename", newRuleName)
	}

	if err := c.RemoveRSSAutoDownloadingRule(ctx, newRuleName); err != nil {
		t.Fatalf("RemoveRSSAutoDownloadingRule: %v", err)
	}

	rules, err = c.GetAllRSSAutoDownloadingRules(ctx)
	if err != nil {
		t.Fatalf("GetAllRSSAutoDownloadingRules (after remove): %v", err)
	}
	if _, ok := rules[newRuleName]; ok {
		t.Fatalf("rule %q still exists after removal", newRuleName)
	}
}
