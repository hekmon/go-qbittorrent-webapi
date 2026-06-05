package qbtapi

import (
	"context"
	"fmt"
	"strconv"
)

const syncAPIName = "sync"

// SyncMainData contains the response from the sync maindata endpoint.
type SyncMainData struct {
	RID               int                     `json:"rid"`
	FullUpdate        bool                    `json:"full_update"`
	Torrents          map[string]TorrentInfos `json:"torrents"`
	TorrentsRemoved   []string                `json:"torrents_removed"`
	Categories        map[string]Category     `json:"categories"`
	CategoriesRemoved []string                `json:"categories_removed"`
	Tags              []string                `json:"tags"`
	TagsRemoved       []string                `json:"tags_removed"`
	ServerState       SyncServerState         `json:"server_state"`
}

// SyncServerState represents the global server state returned in sync updates.
// It contains the same fields as GlobalTransferInfo plus additional partial-update fields.
type SyncServerState struct {
	GlobalTransferInfo
	Queueing          bool `json:"queueing"`             // True if torrent queueing is enabled
	UseAltSpeedLimits bool `json:"use_alt_speed_limits"` // True if alternative speed limits are enabled
	RefreshInterval   int  `json:"refresh_interval"`     // Transfer list refresh interval (milliseconds)
}

// TorrentPeerData represents information about a single peer in a torrent.
type TorrentPeerData struct {
	Client      string  `json:"client"`
	Connection  string  `json:"connection"`
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	DlSpeed     int     `json:"dl_speed"`
	Downloaded  int     `json:"downloaded"`
	Files       string  `json:"files"`
	Flags       string  `json:"flags"`
	FlagsDesc   string  `json:"flags_desc"`
	IP          string  `json:"ip"`
	Port        int     `json:"port"`
	Progress    float64 `json:"progress"`
	Relevance   float64 `json:"relevance"`
	UpSpeed     int     `json:"up_speed"`
	Uploaded    int     `json:"uploaded"`
}

// SyncTorrentPeersData contains the response from the sync torrentPeers endpoint.
type SyncTorrentPeersData struct {
	RID          int                        `json:"rid"`
	FullUpdate   bool                       `json:"full_update"`
	Peers        map[string]TorrentPeerData `json:"peers"`
	PeersRemoved []string                   `json:"peers_removed"`
}

// GetMainData returns changes since the last request.
// Pass rid=0 for a full update.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#get-main-data
func (c *Client) GetMainData(ctx context.Context, rid int) (data SyncMainData, err error) {
	req, err := c.requestBuild(ctx, "GET", syncAPIName, "maindata", map[string]string{
		"rid": strconv.Itoa(rid),
	}, nil)
	if err != nil {
		err = fmt.Errorf("building request failed: %w", err)
		return
	}
	if err = c.requestExecute(req, &data, true); err != nil {
		err = fmt.Errorf("executing request failed: %w", err)
	}
	return
}

// GetTorrentPeersData returns peer data for a specific torrent since the last request.
// Pass rid=0 for a full update.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#get-torrent-peers-data
func (c *Client) GetTorrentPeersData(ctx context.Context, hash string, rid int) (data SyncTorrentPeersData, err error) {
	req, err := c.requestBuild(ctx, "GET", syncAPIName, "torrentPeers", map[string]string{
		"hash": hash,
		"rid":  strconv.Itoa(rid),
	}, nil)
	if err != nil {
		err = fmt.Errorf("building request failed: %w", err)
		return
	}
	if err = c.requestExecute(req, &data, true); err != nil {
		err = fmt.Errorf("executing request failed: %w", err)
	}
	return
}
