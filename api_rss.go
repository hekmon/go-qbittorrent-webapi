package qbtapi

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
)

const rssAPIName = "rss"

// RSSAutoDownloadingRule represents an auto-downloading rule for RSS feeds.
type RSSAutoDownloadingRule struct {
	Enabled                   bool     `json:"enabled"`
	MustContain               string   `json:"mustContain"`
	MustNotContain            string   `json:"mustNotContain"`
	UseRegex                  bool     `json:"useRegex"`
	EpisodeFilter             string   `json:"episodeFilter"`
	SmartFilter               bool     `json:"smartFilter"`
	PreviouslyMatchedEpisodes []string `json:"previouslyMatchedEpisodes"`
	AffectedFeeds             []string `json:"affectedFeeds"`
	IgnoreDays                int      `json:"ignoreDays"`
	LastMatch                 string   `json:"lastMatch"`
	AddPaused                 bool     `json:"addPaused"`
	AssignedCategory          string   `json:"assignedCategory"`
	SavePath                  string   `json:"savePath"`
}

// RSSItems represents the tree structure of RSS feeds and folders returned by GetAllRSSItems.
// Values can be either string (feed URL) or RSSItems (subfolder).
type RSSItems map[string]interface{}

// AddRSSFolder adds a new RSS folder at the given path.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#add-folder
func (c *Client) AddRSSFolder(ctx context.Context, path string) (err error) {
	req, err := c.requestBuild(ctx, "POST", rssAPIName, "addFolder", map[string]string{
		"path": path,
	}, nil)
	if err != nil {
		err = fmt.Errorf("building request failed: %w", err)
		return
	}
	if err = c.requestExecute(req, nil, true); err != nil {
		err = fmt.Errorf("executing request failed: %w", err)
	}
	return
}

// AddRSSFeed adds a new RSS feed. The optional path parameter specifies the folder path.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#add-feed
func (c *Client) AddRSSFeed(ctx context.Context, url string, path *string) (err error) {
	params := map[string]string{
		"url": url,
	}
	if path != nil {
		params["path"] = *path
	}
	req, err := c.requestBuild(ctx, "POST", rssAPIName, "addFeed", params, nil)
	if err != nil {
		err = fmt.Errorf("building request failed: %w", err)
		return
	}
	if err = c.requestExecute(req, nil, true); err != nil {
		err = fmt.Errorf("executing request failed: %w", err)
	}
	return
}

// RemoveRSSItem removes a folder or feed at the given path.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#remove-item
func (c *Client) RemoveRSSItem(ctx context.Context, path string) (err error) {
	req, err := c.requestBuild(ctx, "POST", rssAPIName, "removeItem", map[string]string{
		"path": path,
	}, nil)
	if err != nil {
		err = fmt.Errorf("building request failed: %w", err)
		return
	}
	if err = c.requestExecute(req, nil, true); err != nil {
		err = fmt.Errorf("executing request failed: %w", err)
	}
	return
}

// MoveRSSItem moves or renames a folder or feed.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#move-item
func (c *Client) MoveRSSItem(ctx context.Context, itemPath, destPath string) (err error) {
	req, err := c.requestBuild(ctx, "POST", rssAPIName, "moveItem", map[string]string{
		"itemPath": itemPath,
		"destPath": destPath,
	}, nil)
	if err != nil {
		err = fmt.Errorf("building request failed: %w", err)
		return
	}
	if err = c.requestExecute(req, nil, true); err != nil {
		err = fmt.Errorf("executing request failed: %w", err)
	}
	return
}

// GetAllRSSItems returns all RSS items (feeds and folders).
// If withData is true, current feed articles are included.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#get-all-items
func (c *Client) GetAllRSSItems(ctx context.Context, withData *bool) (items RSSItems, err error) {
	params := make(map[string]string)
	if withData != nil {
		params["withData"] = strconv.FormatBool(*withData)
	}
	req, err := c.requestBuild(ctx, "GET", rssAPIName, "items", params, nil)
	if err != nil {
		err = fmt.Errorf("building request failed: %w", err)
		return
	}
	if err = c.requestExecute(req, &items, true); err != nil {
		err = fmt.Errorf("executing request failed: %w", err)
	}
	return
}

// MarkRSSItemAsRead marks an RSS feed or a specific article as read.
// If articleID is provided, only that article is marked as read.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#mark-as-read
func (c *Client) MarkRSSItemAsRead(ctx context.Context, itemPath string, articleID *string) (err error) {
	params := map[string]string{
		"itemPath": itemPath,
	}
	if articleID != nil {
		params["articleId"] = *articleID
	}
	req, err := c.requestBuild(ctx, "POST", rssAPIName, "markAsRead", params, nil)
	if err != nil {
		err = fmt.Errorf("building request failed: %w", err)
		return
	}
	if err = c.requestExecute(req, nil, true); err != nil {
		err = fmt.Errorf("executing request failed: %w", err)
	}
	return
}

// RefreshRSSItem refreshes a folder or feed.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#refresh-item
func (c *Client) RefreshRSSItem(ctx context.Context, itemPath string) (err error) {
	req, err := c.requestBuild(ctx, "POST", rssAPIName, "refreshItem", map[string]string{
		"itemPath": itemPath,
	}, nil)
	if err != nil {
		err = fmt.Errorf("building request failed: %w", err)
		return
	}
	if err = c.requestExecute(req, nil, true); err != nil {
		err = fmt.Errorf("executing request failed: %w", err)
	}
	return
}

// SetRSSAutoDownloadingRule creates or updates an auto-downloading rule.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#set-auto-downloading-rule
func (c *Client) SetRSSAutoDownloadingRule(ctx context.Context, ruleName string, rule RSSAutoDownloadingRule) (err error) {
	ruleJSON, err := json.Marshal(rule)
	if err != nil {
		err = fmt.Errorf("marshaling rule failed: %w", err)
		return
	}
	req, err := c.requestBuild(ctx, "POST", rssAPIName, "setRule", map[string]string{
		"ruleName": ruleName,
		"ruleDef":  string(ruleJSON),
	}, nil)
	if err != nil {
		err = fmt.Errorf("building request failed: %w", err)
		return
	}
	if err = c.requestExecute(req, nil, true); err != nil {
		err = fmt.Errorf("executing request failed: %w", err)
	}
	return
}

// RenameRSSAutoDownloadingRule renames an existing auto-downloading rule.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#rename-auto-downloading-rule
func (c *Client) RenameRSSAutoDownloadingRule(ctx context.Context, ruleName, newRuleName string) (err error) {
	req, err := c.requestBuild(ctx, "POST", rssAPIName, "renameRule", map[string]string{
		"ruleName":    ruleName,
		"newRuleName": newRuleName,
	}, nil)
	if err != nil {
		err = fmt.Errorf("building request failed: %w", err)
		return
	}
	if err = c.requestExecute(req, nil, true); err != nil {
		err = fmt.Errorf("executing request failed: %w", err)
	}
	return
}

// RemoveRSSAutoDownloadingRule removes an auto-downloading rule.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#remove-auto-downloading-rule
func (c *Client) RemoveRSSAutoDownloadingRule(ctx context.Context, ruleName string) (err error) {
	req, err := c.requestBuild(ctx, "POST", rssAPIName, "removeRule", map[string]string{
		"ruleName": ruleName,
	}, nil)
	if err != nil {
		err = fmt.Errorf("building request failed: %w", err)
		return
	}
	if err = c.requestExecute(req, nil, true); err != nil {
		err = fmt.Errorf("executing request failed: %w", err)
	}
	return
}

// GetAllRSSAutoDownloadingRules returns all auto-downloading rules.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#get-all-auto-downloading-rules
func (c *Client) GetAllRSSAutoDownloadingRules(ctx context.Context) (rules map[string]RSSAutoDownloadingRule, err error) {
	req, err := c.requestBuild(ctx, "GET", rssAPIName, "rules", nil, nil)
	if err != nil {
		err = fmt.Errorf("building request failed: %w", err)
		return
	}
	var raw map[string]RSSAutoDownloadingRule
	if err = c.requestExecute(req, &raw, true); err != nil {
		err = fmt.Errorf("executing request failed: %w", err)
		return
	}
	rules = raw
	return
}

// GetAllArticlesMatchingRule returns all articles that match a given rule, grouped by feed name.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#get-all-articles-matching-a-rule
func (c *Client) GetAllArticlesMatchingRule(ctx context.Context, ruleName string) (articles map[string][]string, err error) {
	req, err := c.requestBuild(ctx, "GET", rssAPIName, "matchingArticles", map[string]string{
		"ruleName": ruleName,
	}, nil)
	if err != nil {
		err = fmt.Errorf("building request failed: %w", err)
		return
	}
	var raw map[string][]string
	if err = c.requestExecute(req, &raw, true); err != nil {
		err = fmt.Errorf("executing request failed: %w", err)
		return
	}
	articles = raw
	return
}
