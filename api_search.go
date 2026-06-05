package qbtapi

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

const searchAPIName = "search"

// SearchJob represents the status of a search job.
type SearchJob struct {
	ID     int    `json:"id"`
	Status string `json:"status"`
	Total  int    `json:"total"`
}

// SearchResult represents a single search result.
type SearchResult struct {
	DescrLink  string `json:"descrLink"`
	FileName   string `json:"fileName"`
	FileSize   int64  `json:"fileSize"`
	FileURL    string `json:"fileUrl"`
	NbLeechers int    `json:"nbLeechers"`
	NbSeeders  int    `json:"nbSeeders"`
	SiteURL    string `json:"siteUrl"`
}

// SearchPluginCategory represents a category supported by a search plugin.
type SearchPluginCategory struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// SearchPlugin represents an installed search plugin.
type SearchPlugin struct {
	Enabled             bool                   `json:"enabled"`
	FullName            string                 `json:"fullName"`
	Name                string                 `json:"name"`
	SupportedCategories []SearchPluginCategory `json:"supportedCategories"`
	URL                 string                 `json:"url"`
	Version             string                 `json:"version"`
}

// SearchResults contains the results of a search query.
type SearchResults struct {
	Results []SearchResult `json:"results"`
	Status  string         `json:"status"`
	Total   int            `json:"total"`
}

// StartSearch starts a new search and returns the job ID.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#start-search
func (c *Client) StartSearch(ctx context.Context, pattern, plugins, category string) (jobID int, err error) {
	req, err := c.requestBuild(ctx, "POST", searchAPIName, "start", map[string]string{
		"pattern":  pattern,
		"plugins":  plugins,
		"category": category,
	}, nil)
	if err != nil {
		err = fmt.Errorf("building request failed: %w", err)
		return
	}
	var resp struct {
		ID int `json:"id"`
	}
	if err = c.requestExecute(req, &resp, true); err != nil {
		err = fmt.Errorf("executing request failed: %w", err)
		return
	}
	jobID = resp.ID
	return
}

// StopSearch stops a running search job.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#stop-search
func (c *Client) StopSearch(ctx context.Context, id int) (err error) {
	req, err := c.requestBuild(ctx, "POST", searchAPIName, "stop", map[string]string{
		"id": strconv.Itoa(id),
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

// GetSearchStatus returns the status of search jobs.
// If id is nil, all jobs are returned.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#get-search-status
func (c *Client) GetSearchStatus(ctx context.Context, id *int) (jobs []SearchJob, err error) {
	params := make(map[string]string)
	if id != nil {
		params["id"] = strconv.Itoa(*id)
	}
	req, err := c.requestBuild(ctx, "GET", searchAPIName, "status", params, nil)
	if err != nil {
		err = fmt.Errorf("building request failed: %w", err)
		return
	}
	if err = c.requestExecute(req, &jobs, true); err != nil {
		err = fmt.Errorf("executing request failed: %w", err)
	}
	return
}

// GetSearchResults returns the results of a search job.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#get-search-results
func (c *Client) GetSearchResults(ctx context.Context, id int, limit, offset *int) (results SearchResults, err error) {
	params := map[string]string{
		"id": strconv.Itoa(id),
	}
	if limit != nil {
		params["limit"] = strconv.Itoa(*limit)
	}
	if offset != nil {
		params["offset"] = strconv.Itoa(*offset)
	}
	req, err := c.requestBuild(ctx, "GET", searchAPIName, "results", params, nil)
	if err != nil {
		err = fmt.Errorf("building request failed: %w", err)
		return
	}
	if err = c.requestExecute(req, &results, true); err != nil {
		err = fmt.Errorf("executing request failed: %w", err)
	}
	return
}

// DeleteSearch deletes a search job.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#delete-search
func (c *Client) DeleteSearch(ctx context.Context, id int) (err error) {
	req, err := c.requestBuild(ctx, "POST", searchAPIName, "delete", map[string]string{
		"id": strconv.Itoa(id),
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

// GetSearchPlugins returns all installed search plugins.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#get-search-plugins
func (c *Client) GetSearchPlugins(ctx context.Context) (plugins []SearchPlugin, err error) {
	req, err := c.requestBuild(ctx, "GET", searchAPIName, "plugins", nil, nil)
	if err != nil {
		err = fmt.Errorf("building request failed: %w", err)
		return
	}
	if err = c.requestExecute(req, &plugins, true); err != nil {
		err = fmt.Errorf("executing request failed: %w", err)
	}
	return
}

// InstallSearchPlugin installs one or more search plugins from the given sources.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#install-search-plugin
func (c *Client) InstallSearchPlugin(ctx context.Context, sources []string) (err error) {
	req, err := c.requestBuild(ctx, "POST", searchAPIName, "installPlugin", map[string]string{
		"sources": strings.Join(sources, "|"),
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

// UninstallSearchPlugin uninstalls one or more search plugins by name.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#uninstall-search-plugin
func (c *Client) UninstallSearchPlugin(ctx context.Context, names []string) (err error) {
	req, err := c.requestBuild(ctx, "POST", searchAPIName, "uninstallPlugin", map[string]string{
		"names": strings.Join(names, "|"),
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

// EnableSearchPlugin enables or disables one or more search plugins by name.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#enable-search-plugin
func (c *Client) EnableSearchPlugin(ctx context.Context, names []string, enable bool) (err error) {
	req, err := c.requestBuild(ctx, "POST", searchAPIName, "enablePlugin", map[string]string{
		"names":  strings.Join(names, "|"),
		"enable": strconv.FormatBool(enable),
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

// UpdateSearchPlugins updates all search plugins.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#update-search-plugins
func (c *Client) UpdateSearchPlugins(ctx context.Context) (err error) {
	req, err := c.requestBuild(ctx, "POST", searchAPIName, "updatePlugins", nil, nil)
	if err != nil {
		err = fmt.Errorf("building request failed: %w", err)
		return
	}
	if err = c.requestExecute(req, nil, true); err != nil {
		err = fmt.Errorf("executing request failed: %w", err)
	}
	return
}
