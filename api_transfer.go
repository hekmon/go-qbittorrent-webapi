package qbtapi

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

const transferAPIName = "transfer"

// GlobalTransferInfo contains the global transfer information.
type GlobalTransferInfo struct {
	DlInfoSpeed      int    `json:"dl_info_speed"`     // Global download rate (bytes/s)
	DlInfoData       int    `json:"dl_info_data"`      // Data downloaded this session (bytes)
	UpInfoSpeed      int    `json:"up_info_speed"`     // Global upload rate (bytes/s)
	UpInfoData       int    `json:"up_info_data"`      // Data uploaded this session (bytes)
	DlRateLimit      int    `json:"dl_rate_limit"`     // Download rate limit (bytes/s)
	UpRateLimit      int    `json:"up_rate_limit"`     // Upload rate limit (bytes/s)
	DHTNodes         int    `json:"dht_nodes"`         // DHT nodes connected to
	ConnectionStatus string `json:"connection_status"` // Connection status: connected, firewalled, disconnected
}

// GetGlobalTransferInfo returns info you usually see in qBt status bar.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#get-global-transfer-info
func (c *Client) GetGlobalTransferInfo(ctx context.Context) (info GlobalTransferInfo, err error) {
	req, err := c.requestBuild(ctx, "GET", transferAPIName, "info", nil, nil)
	if err != nil {
		err = fmt.Errorf("building request failed: %w", err)
		return
	}
	if err = c.requestExecute(req, &info, true); err != nil {
		err = fmt.Errorf("executing request failed: %w", err)
	}
	return
}

// GetAlternativeSpeedLimitsState returns true if alternative speed limits are enabled.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#get-alternative-speed-limits-state
func (c *Client) GetAlternativeSpeedLimitsState(ctx context.Context) (enabled bool, err error) {
	req, err := c.requestBuild(ctx, "GET", transferAPIName, "speedLimitsMode", nil, nil)
	if err != nil {
		err = fmt.Errorf("building request failed: %w", err)
		return
	}
	var raw string
	if err = c.requestExecute(req, &raw, true); err != nil {
		err = fmt.Errorf("executing request failed: %w", err)
		return
	}
	switch raw {
	case "1":
		enabled = true
	case "0":
		enabled = false
	default:
		err = fmt.Errorf("unexpected speed limits mode value: %q", raw)
	}
	return
}

// ToggleAlternativeSpeedLimits toggles the alternative speed limits.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#toggle-alternative-speed-limits
func (c *Client) ToggleAlternativeSpeedLimits(ctx context.Context) (err error) {
	req, err := c.requestBuild(ctx, "POST", transferAPIName, "toggleSpeedLimitsMode", nil, nil)
	if err != nil {
		err = fmt.Errorf("building request failed: %w", err)
		return
	}
	if err = c.requestExecute(req, nil, true); err != nil {
		err = fmt.Errorf("executing request failed: %w", err)
	}
	return
}

// GetGlobalDownloadLimit returns the value of current global download speed limit in bytes/second.
// This value will be zero if no limit is applied.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#get-global-download-limit
func (c *Client) GetGlobalDownloadLimit(ctx context.Context) (limit int, err error) {
	req, err := c.requestBuild(ctx, "GET", transferAPIName, "downloadLimit", nil, nil)
	if err != nil {
		err = fmt.Errorf("building request failed: %w", err)
		return
	}
	var raw string
	if err = c.requestExecute(req, &raw, true); err != nil {
		err = fmt.Errorf("executing request failed: %w", err)
		return
	}
	limit, err = strconv.Atoi(raw)
	if err != nil {
		err = fmt.Errorf("parsing download limit failed: %w", err)
	}
	return
}

// SetGlobalDownloadLimit sets the global download speed limit in bytes/second.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#set-global-download-limit
func (c *Client) SetGlobalDownloadLimit(ctx context.Context, limit int) (err error) {
	req, err := c.requestBuild(ctx, "POST", transferAPIName, "setDownloadLimit", map[string]string{
		"limit": strconv.Itoa(limit),
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

// GetGlobalUploadLimit returns the value of current global upload speed limit in bytes/second.
// This value will be zero if no limit is applied.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#get-global-upload-limit
func (c *Client) GetGlobalUploadLimit(ctx context.Context) (limit int, err error) {
	req, err := c.requestBuild(ctx, "GET", transferAPIName, "uploadLimit", nil, nil)
	if err != nil {
		err = fmt.Errorf("building request failed: %w", err)
		return
	}
	var raw string
	if err = c.requestExecute(req, &raw, true); err != nil {
		err = fmt.Errorf("executing request failed: %w", err)
		return
	}
	limit, err = strconv.Atoi(raw)
	if err != nil {
		err = fmt.Errorf("parsing upload limit failed: %w", err)
	}
	return
}

// SetGlobalUploadLimit sets the global upload speed limit in bytes/second.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#set-global-upload-limit
func (c *Client) SetGlobalUploadLimit(ctx context.Context, limit int) (err error) {
	req, err := c.requestBuild(ctx, "POST", transferAPIName, "setUploadLimit", map[string]string{
		"limit": strconv.Itoa(limit),
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

// BanPeers bans the given peers. Each peer must be in the format host:port.
// Multiple peers can be banned at once.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#ban-peers
func (c *Client) BanPeers(ctx context.Context, peers []string) (err error) {
	req, err := c.requestBuild(ctx, "POST", transferAPIName, "banPeers", map[string]string{
		"peers": strings.Join(peers, "|"),
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
