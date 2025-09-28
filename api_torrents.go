package qbtapi

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

/*
	Torrent management
	https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#torrent-management
*/

const (
	torrentsAPIName = "torrents"
)

// ListFilters contains all filters that can be applied to torrent listing.
// All values will be automatically url encoded and so does not need pre encoding.
type ListFilters struct {
	State       *FilterState // Filter torrent list by state
	Category    *string      // Get torrents with the given category (empty string means "without category"; nil parameter means "any category").
	Tag         *string      // Get torrents with the given tag (empty string means "without tag"; nil parameter means "any tag"
	Sort        *string      // Sort torrents by given key. They can be sorted using any field of the response's JSON array (which are documented below) as the sort key
	ReverseSort *bool        // Enable reverse sorting. Defaults to false.
	Limit       *int         // Limit the number of torrents returned.
	Offset      *int         // Set offset (if less than 0, offset from end)
	Hashes      []string     // Filter by hashes
}

func (lf ListFilters) getLowLevelRepr() (filters map[string]string) {
	filters = make(map[string]string, reflect.TypeOf(lf).NumField())
	if lf.State != nil {
		filters["filter"] = string(*lf.State)
	}
	if lf.Category != nil {
		filters["category"] = *lf.Category
	}
	if lf.Tag != nil {
		filters["tag"] = *lf.Tag
	}
	if lf.Sort != nil {
		filters["sort"] = *lf.Sort
	}
	if lf.ReverseSort != nil {
		filters["reverse"] = strconv.FormatBool(*lf.ReverseSort)
	}
	if lf.Limit != nil {
		filters["limit"] = strconv.Itoa(*lf.Limit)
	}
	if lf.Offset != nil {
		filters["offset"] = strconv.Itoa(*lf.Offset)
	}
	if lf.Hashes != nil {
		filters["hashes"] = strings.Join(lf.Hashes, "|")
	}
	return
}

// FilterState represent a filtering on a specific state.
type FilterState string

const (
	FilterStateAll                FilterState = "all"
	FilterStateDownloading        FilterState = "downloading"
	FilterStateSeeding            FilterState = "seeding"
	FilterStateCompleted          FilterState = "completed"
	FilterStateStopped            FilterState = "stopped"
	FilterStateActive             FilterState = "active"
	FilterStateInactive           FilterState = "inactive"
	FilterStateRunning            FilterState = "running"
	FilterStateStalled            FilterState = "stalled"
	FilterStateStalledUploading   FilterState = "stalled_uploading"
	FilterStateStalledDownloading FilterState = "stalled_downloading"
	FilterStateErrored            FilterState = "errored"
)

func (fs FilterState) Ptr() *FilterState {
	return &fs
}

// GetTorrentList returns a torrent listing. filters are optional and can be nil.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#get-torrent-list
func (c *Client) GetTorrentList(ctx context.Context, filters *ListFilters) (list []TorrentInfos, err error) {
	var preparedFilters map[string]string
	if filters != nil {
		preparedFilters = filters.getLowLevelRepr()
	}
	// build request
	req, err := c.requestBuild(ctx, "GET", torrentsAPIName, "info", preparedFilters)
	if err != nil {
		err = fmt.Errorf("request building failure: %w", err)
		return
	}
	// execute request
	if err = c.requestExecute(req, &list, true); err != nil {
		err = fmt.Errorf("executing request failed: %w", err)
	}
	return
}

// TorrentsInfos contains a torrent properties.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#get-torrent-list
type TorrentInfos struct {
	AddedOn            int     `json:"added_on"`           // Time (Unix Epoch) when the torrent was added to the client
	AmountLeft         int     `json:"amount_left"`        // Amount of data left to download (bytes)
	AutoTMM            bool    `json:"auto_tmm"`           // Whether this torrent is managed by Automatic Torrent Management
	Availability       float64 `json:"availability"`       // Percentage of file pieces currently available
	Category           string  `json:"category"`           // Category of the torrent
	Completed          int     `json:"completed"`          // Amount of transfer data completed (bytes)
	CompletionOn       int     `json:"completion_on"`      // Time (Unix Epoch) when the torrent completed
	ContentPath        string  `json:"content_path"`       // Absolute path of torrent content (root path for multifile torrents, absolute file path for singlefile torrents)
	DownloadSpeedLimit int     `json:"dl_limit"`           // Torrent download speed limit (bytes/s). -1 if unlimited.
	DownloadSpeed      int     `json:"dlspeed"`            // Torrent download speed (bytes/s)
	Downloaded         int     `json:"downloaded"`         // Amount of data downloaded
	DownloadedSession  int     `json:"downloaded_session"` // Amount of data downloaded this session
	ETA                int     `json:"eta"`                // Torrent ETA (seconds)
	FirstLastPiecePrio bool    `json:"f_l_piece_prio"`     // True if first last piece are prioritized
	ForceStart         bool    `json:"force_start"`        // True if force start is enabled for this torrent
	Hash               string  `json:"hash"`               // Torrent hash
	IsPrivate          bool    `json:"isPrivate"`          // True if torrent is from a private tracker
	LastActivity       int     `json:"last_activity"`      // Last time (Unix Epoch) when a chunk was downloaded/uploaded
	MagnetURI          string  `json:"magnet_uri"`         // Magnet URI corresponding to this torrent
	MaxRatio           float64 `json:"max_ratio"`          // Maximum share ratio until torrent is stopped from seeding/uploading
	MaxSeedingTime     int     `json:"max_seeding_time"`   // Maximum seeding time (seconds) until torrent is stopped from seeding
	Name               string  `json:"name"`               // Torrent name
	NumComplete        int     `json:"num_complete"`       // Number of seeds in the swarm
	NumIncomplete      int     `json:"num_incomplete"`     // Number of leechers in the swarm
	NumLeechs          int     `json:"num_leechs"`         // Number of leechers connected to
	NumSeeds           int     `json:"num_seeds"`          // Number of seeds connected to
	Priority           int     `json:"priority"`           // Torrent priority. Returns -1 if queuing is disabled or torrent is in seed mode
	Progress           float64 `json:"progress"`           // Torrent progress (percentage/100)
	Ratio              float64 `json:"ratio"`              // Torrent share ratio. Max ratio value: 9999.
	RatioLimit         float64 `json:"ratio_limit"`        // TODO (what is different from max_ratio?)
	Reannounce         int     `json:"reannounce"`         // Time until the next tracker reannounce
	SavePath           string  `json:"save_path"`          // Path where this torrent's data is stored
	SeedingTime        int     `json:"seeding_time"`       // Torrent elapsed time while complete (seconds)
	SeedingTimeLimit   int     `json:"seeding_time_limit"` // TODO (what is different from max_seeding_time?)
	SeenComplete       int     `json:"seen_complete"`      // Time (Unix Epoch) when this torrent was last seen complete
	SequentialDownload bool    `json:"seq_dl"`             // True if sequential download is enabled
	Size               int     `json:"size"`               // Total size (bytes) of files selected for download
	State              string  `json:"state"`              // Torrent state. See table here below for the possible values
	SuperSeeding       bool    `json:"super_seeding"`      // True if super seeding is enabled
	Tags               string  `json:"tags"`               // Comma-concatenated tag list of the torrent
	TimeActive         int     `json:"time_active"`        // Total active time (seconds)
	TotalSize          int     `json:"total_size"`         // Total size (bytes) of all file in this torrent (including unselected ones)
	Tracker            string  `json:"tracker"`            // The first tracker with working status. Returns empty string if no tracker is working.
	UploadSpeedLimit   int     `json:"up_limit"`           // Torrent upload speed limit (bytes/s). -1 if unlimited.
	Uploaded           int     `json:"uploaded"`           // Amount of data uploaded
	UploadedSession    int     `json:"uploaded_session"`   // Amount of data uploaded this session
	UploadSpeed        int     `json:"upspeed"`            // Torrent upload speed (bytes/s)
}
