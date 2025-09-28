package qbtapi

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/hekmon/cunits/v3"
)

/*
	Torrent management
	https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#torrent-management
*/

const (
	torrentsAPIName = "torrents"
)

var (
	// UnlimitedSpeedLimit is a special value that can be used to set a torrent or global speed limit to unlimited.
	UnlimitedSpeedLimit = cunits.Speed{Bits: math.MaxUint64}
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
	AddedOn            time.Time     `json:"added_on"`           // Time when the torrent was added to the client
	AmountLeft         cunits.Bits   `json:"amount_left"`        // Amount of data left to download
	AutoTMM            bool          `json:"auto_tmm"`           // Whether this torrent is managed by Automatic Torrent Management
	Availability       float64       `json:"availability"`       // Percentage of file pieces currently available
	Category           string        `json:"category"`           // Category of the torrent
	Completed          cunits.Bits   `json:"completed"`          // Amount of transfer data completed
	CompletionOn       time.Time     `json:"completion_on"`      // Time when the torrent completed
	ContentPath        string        `json:"content_path"`       // Absolute path of torrent content (root path for multifile torrents, absolute file path for singlefile torrents)
	DownloadSpeedLimit cunits.Speed  `json:"dl_limit"`           // Torrent download speed limit per second. -1 if unlimited.
	DownloadSpeed      cunits.Speed  `json:"dlspeed"`            // Torrent download speed per second.
	Downloaded         cunits.Bits   `json:"downloaded"`         // Amount of data downloaded
	DownloadedSession  cunits.Bits   `json:"downloaded_session"` // Amount of data downloaded this session
	ETA                time.Duration `json:"eta"`                // Torrent ETA
	FirstLastPiecePrio bool          `json:"f_l_piece_prio"`     // True if first last piece are prioritized
	ForceStart         bool          `json:"force_start"`        // True if force start is enabled for this torrent
	Hash               string        `json:"hash"`               // Torrent hash
	IsPrivate          bool          `json:"isPrivate"`          // True if torrent is from a private tracker
	LastActivity       time.Time     `json:"last_activity"`      // Last time when a chunk was downloaded/uploaded
	MagnetURI          string        `json:"magnet_uri"`         // Magnet URI corresponding to this torrent
	MaxRatio           float64       `json:"max_ratio"`          // Maximum share ratio until torrent is stopped from seeding/uploading
	MaxSeedingTime     time.Duration `json:"max_seeding_time"`   // Maximum seeding time until torrent is stopped from seeding
	Name               string        `json:"name"`               // Torrent name
	NumComplete        int           `json:"num_complete"`       // Number of seeds in the swarm
	NumIncomplete      int           `json:"num_incomplete"`     // Number of leechers in the swarm
	NumLeechs          int           `json:"num_leechs"`         // Number of leechers connected to
	NumSeeds           int           `json:"num_seeds"`          // Number of seeds connected to
	Priority           int           `json:"priority"`           // Torrent priority. Returns -1 if queuing is disabled or torrent is in seed mode
	Progress           float64       `json:"progress"`           // Torrent progress (percentage/100)
	Ratio              float64       `json:"ratio"`              // Torrent share ratio. Max ratio value: 9999.
	RatioLimit         float64       `json:"ratio_limit"`        // TODO (what is different from max_ratio?)
	Reannounce         time.Duration `json:"reannounce"`         // Time until the next tracker reannounce
	SavePath           string        `json:"save_path"`          // Path where this torrent's data is stored
	SeedingTime        time.Duration `json:"seeding_time"`       // Torrent elapsed time while complete
	SeedingTimeLimit   time.Duration `json:"seeding_time_limit"` // TODO (what is different from max_seeding_time?)
	SeenComplete       time.Time     `json:"seen_complete"`      // Time when this torrent was last seen complete
	SequentialDownload bool          `json:"seq_dl"`             // True if sequential download is enabled
	Size               cunits.Bits   `json:"size"`               // Total size of files selected for download
	State              FilterState   `json:"state"`              // Torrent state
	SuperSeeding       bool          `json:"super_seeding"`      // True if super seeding is enabled
	Tags               []string      `json:"tags"`               // Comma-concatenated tag list of the torrent
	TimeActive         time.Duration `json:"time_active"`        // Total active time
	TotalSize          cunits.Bits   `json:"total_size"`         // Total size of all file in this torrent (including unselected ones)
	Tracker            string        `json:"tracker"`            // The first tracker with working status. Returns empty string if no tracker is working.
	UploadSpeedLimit   cunits.Speed  `json:"up_limit"`           // Torrent upload speed limit per second. -1 if unlimited.
	Uploaded           cunits.Bits   `json:"uploaded"`           // Amount of data uploaded
	UploadedSession    cunits.Bits   `json:"uploaded_session"`   // Amount of data uploaded this session
	UploadSpeed        cunits.Speed  `json:"upspeed"`            // Torrent upload speed (per second)
}

func (ti *TorrentInfos) UnmarshalJSON(data []byte) (err error) {
	type mask TorrentInfos
	tmp := struct {
		*mask
		// Custom unmarshaling
		AddedOn            int64  `json:"added_on"`           // Time (Unix Epoch) when the torrent was added to the client
		AmountLeft         int    `json:"amount_left"`        // Amount of data left to download (bytes)
		Completed          int    `json:"completed"`          // Amount of transfer data completed (bytes)
		CompletionOn       int64  `json:"completion_on"`      // Time (Unix Epoch) when the torrent completed
		DownloadSpeedLimit int    `json:"dl_limit"`           // Torrent download speed limit (bytes/s). -1 if unlimited.
		DownloadSpeed      int    `json:"dlspeed"`            // Torrent download speed (bytes/s)
		Downloaded         int    `json:"downloaded"`         // Amount of data downloaded
		DownloadedSession  int    `json:"downloaded_session"` // Amount of data downloaded this session
		ETA                int    `json:"eta"`                // Torrent ETA (seconds)
		LastActivity       int64  `json:"last_activity"`      // Last time (Unix Epoch) when a chunk was downloaded/uploaded
		MaxSeedingTime     int    `json:"max_seeding_time"`   // Maximum seeding time (seconds) until torrent is stopped from seeding
		Reannounce         int    `json:"reannounce"`         // Time until the next tracker reannounce
		SeedingTime        int    `json:"seeding_time"`       // Torrent elapsed time while complete (seconds)
		SeedingTimeLimit   int    `json:"seeding_time_limit"` // TODO (what is different from max_seeding_time?)
		SeenComplete       int64  `json:"seen_complete"`      // Time (Unix Epoch) when this torrent was last seen complete
		Size               int    `json:"size"`               // Total size (bytes) of files selected for download
		Tags               string `json:"tags"`               // Comma-concatenated tag list of the torrent
		TimeActive         int    `json:"time_active"`        // Total active time (seconds)
		TotalSize          int    `json:"total_size"`         // Total size (bytes) of all file in this torrent (including unselected ones)
		UploadSpeedLimit   int    `json:"up_limit"`           // Torrent upload speed limit (bytes/s). -1 if unlimited.
		Uploaded           int    `json:"uploaded"`           // Amount of data uploaded
		UploadedSession    int    `json:"uploaded_session"`   // Amount of data uploaded this session
		UploadSpeed        int    `json:"upspeed"`            // Torrent upload speed (bytes/s)
	}{
		mask: (*mask)(ti),
	}
	// Unmarshall to tmp struct
	if err = json.Unmarshal(data, &tmp); err != nil {
		return
	}
	// Adapt to golang types
	ti.AddedOn = time.Unix(tmp.AddedOn, 0)
	ti.AmountLeft = cunits.ImportInBytes(float64(tmp.AmountLeft))
	ti.Completed = cunits.ImportInBytes(float64(tmp.Completed))
	ti.CompletionOn = time.Unix(tmp.CompletionOn, 0)
	switch tmp.DownloadSpeedLimit {
	case -1:
		// special value
		ti.DownloadSpeedLimit = UnlimitedSpeedLimit
	default:
		ti.DownloadSpeedLimit = cunits.Speed{Bits: cunits.ImportInBytes(float64(tmp.DownloadSpeedLimit))}
	}
	ti.DownloadSpeed = cunits.Speed{Bits: cunits.ImportInBytes(float64(tmp.DownloadSpeed))}
	ti.Downloaded = cunits.ImportInBytes(float64(tmp.Downloaded))
	ti.DownloadedSession = cunits.ImportInBytes(float64(tmp.DownloadedSession))
	ti.ETA = time.Duration(tmp.ETA) * time.Second
	ti.LastActivity = time.Unix(tmp.LastActivity, 0)
	ti.MaxSeedingTime = time.Duration(tmp.MaxSeedingTime) * time.Second
	ti.Reannounce = time.Duration(tmp.Reannounce) * time.Second
	ti.SeedingTime = time.Duration(tmp.SeedingTime) * time.Second
	ti.SeedingTimeLimit = time.Duration(tmp.SeedingTimeLimit) * time.Second
	ti.SeenComplete = time.Unix(tmp.SeenComplete, 0)
	ti.Size = cunits.ImportInBytes(float64(tmp.Size))
	ti.Tags = strings.Split(tmp.Tags, ", ")
	ti.TimeActive = time.Duration(tmp.TimeActive) * time.Second
	ti.TotalSize = cunits.ImportInBytes(float64(tmp.TotalSize))
	switch tmp.UploadSpeedLimit {
	case -1:
		// special value
		ti.UploadSpeedLimit = UnlimitedSpeedLimit
	default:
		ti.UploadSpeedLimit = cunits.Speed{Bits: cunits.ImportInBytes(float64(tmp.UploadSpeedLimit))}
	}
	ti.Uploaded = cunits.ImportInBytes(float64(tmp.Uploaded))
	ti.UploadedSession = cunits.ImportInBytes(float64(tmp.UploadedSession))
	ti.UploadSpeed = cunits.Speed{Bits: cunits.ImportInBytes(float64(tmp.UploadSpeed))}
	return
}

func (ti *TorrentInfos) MarshalJSON() ([]byte, error) {
	type mask TorrentInfos
	tmp := struct {
		*mask
		// Custom marshaling
		AddedOn            int64  `json:"added_on"`           // Time (Unix Epoch) when the torrent was added to the client
		AmountLeft         int    `json:"amount_left"`        // Amount of data left to download (bytes)
		Completed          int    `json:"completed"`          // Amount of transfer data completed (bytes)
		CompletionOn       int64  `json:"completion_on"`      // Time (Unix Epoch) when the torrent completed
		DownloadSpeedLimit int    `json:"dl_limit"`           // Torrent download speed limit (bytes/s). -1 if unlimited.
		DownloadSpeed      int    `json:"dlspeed"`            // Torrent download speed (bytes/s)
		Downloaded         int    `json:"downloaded"`         // Amount of data downloaded
		DownloadedSession  int    `json:"downloaded_session"` // Amount of data downloaded this session
		ETA                int    `json:"eta"`                // Torrent ETA (seconds)
		LastActivity       int64  `json:"last_activity"`      // Last time (Unix Epoch) when a chunk was downloaded/uploaded
		MaxSeedingTime     int    `json:"max_seeding_time"`   // Maximum seeding time (seconds) until torrent is stopped from seeding
		Reannounce         int    `json:"reannounce"`         // Time until the next tracker reannounce
		SeedingTime        int    `json:"seeding_time"`       // Torrent elapsed time while complete (seconds)
		SeedingTimeLimit   int    `json:"seeding_time_limit"` // TODO (what is different from max_seeding_time?)
		SeenComplete       int64  `json:"seen_complete"`      // Time (Unix Epoch) when this torrent was last seen complete
		Size               int    `json:"size"`               // Total size (bytes) of files selected for download
		Tags               string `json:"tags"`               // Comma-concatenated tag list of the torrent
		TimeActive         int    `json:"time_active"`        // Total active time (seconds)
		TotalSize          int    `json:"total_size"`         // Total size (bytes) of all file in this torrent (including unselected ones)
		UploadSpeedLimit   int    `json:"up_limit"`           // Torrent upload speed limit (bytes/s). -1 if unlimited.
		Uploaded           int    `json:"uploaded"`           // Amount of data uploaded
		UploadedSession    int    `json:"uploaded_session"`   // Amount of data uploaded this session
		UploadSpeed        int    `json:"upspeed"`            // Torrent upload speed (bytes/s)
	}{
		mask: (*mask)(ti),
	}
	// Adapt to JSON types
	tmp.AddedOn = ti.AddedOn.Unix()
	tmp.AmountLeft = int(ti.AmountLeft.Bytes())
	tmp.Completed = int(ti.Completed.Bytes())
	tmp.CompletionOn = ti.CompletionOn.Unix()
	switch ti.DownloadSpeedLimit {
	case UnlimitedSpeedLimit:
		// special value
		tmp.DownloadSpeedLimit = int(-1)
	default:
		tmp.DownloadSpeedLimit = int(ti.DownloadSpeedLimit.Bytes())
	}
	tmp.DownloadSpeed = int(ti.DownloadSpeed.Bytes())
	tmp.Downloaded = int(ti.Downloaded.Bytes())
	tmp.DownloadedSession = int(ti.DownloadedSession.Bytes())
	tmp.ETA = int(ti.ETA.Seconds())
	tmp.LastActivity = ti.LastActivity.Unix()
	tmp.MaxSeedingTime = int(ti.MaxSeedingTime.Seconds())
	tmp.Reannounce = int(ti.Reannounce.Seconds())
	tmp.SeedingTime = int(ti.SeedingTime.Seconds())
	tmp.SeedingTimeLimit = int(ti.SeedingTimeLimit.Seconds())
	tmp.SeenComplete = ti.SeenComplete.Unix()
	tmp.Size = int(ti.Size.Bytes())
	tmp.Tags = strings.Join(ti.Tags, ", ")
	tmp.TimeActive = int(ti.TimeActive.Seconds())
	tmp.TotalSize = int(ti.TotalSize.Bytes())
	switch ti.UploadSpeedLimit {
	case UnlimitedSpeedLimit:
		// special value
		tmp.UploadSpeedLimit = int(-1)
	default:
		tmp.UploadSpeedLimit = int(ti.UploadSpeedLimit.Bytes())
	}
	tmp.Uploaded = int(ti.Uploaded.Bytes())
	tmp.UploadedSession = int(ti.UploadedSession.Bytes())
	tmp.UploadSpeed = int(ti.UploadSpeed.Bytes())
	return json.Marshal(tmp)
}
