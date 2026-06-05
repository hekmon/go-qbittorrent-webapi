package qbtapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
	"net/url"
	"os"
	"path/filepath"
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
	torrentsAPIName   = "torrents"
	tagListSeparator  = ","
	hashListSeparator = "|"
)

/*
	Listing
*/

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
		filters["hashes"] = strings.Join(lf.Hashes, hashListSeparator)
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
	req, err := c.requestBuild(ctx, "GET", torrentsAPIName, "info", preparedFilters, nil)
	if err != nil {
		err = fmt.Errorf("building request failed: %w", err)
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
	DownloadSpeedLimit Speed         `json:"dl_limit"`           // Torrent download speed limit per second. -1 if unlimited.
	DownloadSpeed      Speed         `json:"dlspeed"`            // Torrent download speed per second.
	Downloaded         cunits.Bits   `json:"downloaded"`         // Amount of data downloaded
	DownloadedSession  cunits.Bits   `json:"downloaded_session"` // Amount of data downloaded this session
	ETA                time.Duration `json:"eta"`                // Torrent ETA
	FirstLastPiecePrio bool          `json:"f_l_piece_prio"`     // True if first last piece are prioritized
	ForceStart         bool          `json:"force_start"`        // True if force start is enabled for this torrent
	Hash               string        `json:"hash"`               // Torrent hash
	Private            bool          `json:"isPrivate"`          // True if torrent is from a private tracker
	LastActivity       time.Time     `json:"last_activity"`      // Last time when a chunk was downloaded/uploaded
	MagnetURI          *url.URL      `json:"magnet_uri"`         // Magnet URI corresponding to this torrent (use .Query() to access values)
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
	RatioLimit         float64       `json:"ratio_limit"`        // Per-torrent configured share ratio limit. -1 means no limit, -2 means use global limit
	Reannounce         time.Duration `json:"reannounce"`         // Time until the next tracker reannounce
	SavePath           string        `json:"save_path"`          // Path where this torrent's data is stored
	SeedingTime        time.Duration `json:"seeding_time"`       // Torrent elapsed time while complete
	SeedingTimeLimit   time.Duration `json:"seeding_time_limit"` // Per-torrent configured seeding time limit. -1 means no limit, -2 means use global limit
	SeenComplete       time.Time     `json:"seen_complete"`      // Time when this torrent was last seen complete
	SequentialDownload bool          `json:"seq_dl"`             // True if sequential download is enabled
	Size               cunits.Bits   `json:"size"`               // Total size of files selected for download
	State              TorrentState  `json:"state"`              // Torrent state
	SuperSeeding       bool          `json:"super_seeding"`      // True if super seeding is enabled
	Tags               []string      `json:"tags"`               // Comma-concatenated tag list of the torrent
	TimeActive         time.Duration `json:"time_active"`        // Total active time
	TotalSize          cunits.Bits   `json:"total_size"`         // Total size of all file in this torrent (including unselected ones)
	Tracker            string        `json:"tracker"`            // The first tracker with working status. Returns empty string if no tracker is working.
	UploadSpeedLimit   Speed         `json:"up_limit"`           // Torrent upload speed limit per second. -1 if unlimited.
	Uploaded           cunits.Bits   `json:"uploaded"`           // Amount of data uploaded
	UploadedSession    cunits.Bits   `json:"uploaded_session"`   // Amount of data uploaded this session
	UploadSpeed        Speed         `json:"upspeed"`            // Torrent upload speed (per second)
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
		MagnetURI          string `json:"magnet_uri"`         // Magnet URI corresponding to this torrent
		MaxSeedingTime     int    `json:"max_seeding_time"`   // Maximum seeding time (seconds) until torrent is stopped from seeding
		Reannounce         int    `json:"reannounce"`         // Time until the next tracker reannounce
		SeedingTime        int    `json:"seeding_time"`       // Torrent elapsed time while complete (seconds)
		SeedingTimeLimit   int    `json:"seeding_time_limit"` // Per-torrent configured seeding time limit (seconds). -1 means no limit, -2 means use global limit
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
	ti.DownloadSpeedLimit = GetSpeedFromBytes(tmp.DownloadSpeedLimit)
	ti.DownloadSpeed = GetSpeedFromBytes(tmp.DownloadSpeed)
	ti.Downloaded = cunits.ImportInBytes(float64(tmp.Downloaded))
	ti.DownloadedSession = cunits.ImportInBytes(float64(tmp.DownloadedSession))
	ti.ETA = time.Duration(tmp.ETA) * time.Second
	ti.LastActivity = time.Unix(tmp.LastActivity, 0)
	if ti.MagnetURI, err = url.Parse(tmp.MagnetURI); err != nil {
		err = fmt.Errorf("parsing magnet URI failed: %w", err)
		return
	}
	ti.MaxSeedingTime = time.Duration(tmp.MaxSeedingTime) * time.Second
	ti.Reannounce = time.Duration(tmp.Reannounce) * time.Second
	ti.SeedingTime = time.Duration(tmp.SeedingTime) * time.Second
	ti.SeedingTimeLimit = time.Duration(tmp.SeedingTimeLimit) * time.Second
	ti.SeenComplete = time.Unix(tmp.SeenComplete, 0)
	ti.Size = cunits.ImportInBytes(float64(tmp.Size))
	if tmp.Tags != "" {
		ti.Tags = strings.Split(tmp.Tags, tagListSeparator)
		for i := range ti.Tags {
			ti.Tags[i] = strings.TrimSpace(ti.Tags[i])
		}
	}
	ti.TimeActive = time.Duration(tmp.TimeActive) * time.Second
	ti.TotalSize = cunits.ImportInBytes(float64(tmp.TotalSize))
	ti.UploadSpeedLimit = GetSpeedFromBytes(tmp.UploadSpeedLimit)
	ti.Uploaded = cunits.ImportInBytes(float64(tmp.Uploaded))
	ti.UploadedSession = cunits.ImportInBytes(float64(tmp.UploadedSession))
	ti.UploadSpeed = GetSpeedFromBytes(tmp.UploadSpeed)
	return
}

func (ti *TorrentInfos) MarshalJSON() ([]byte, error) {
	type mask TorrentInfos
	tmp := struct {
		mask
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
		MagnetURI          string `json:"magnet_uri"`         // Magnet URI corresponding to this torrent
		MaxSeedingTime     int    `json:"max_seeding_time"`   // Maximum seeding time (seconds) until torrent is stopped from seeding
		Reannounce         int    `json:"reannounce"`         // Time until the next tracker reannounce
		SeedingTime        int    `json:"seeding_time"`       // Torrent elapsed time while complete (seconds)
		SeedingTimeLimit   int    `json:"seeding_time_limit"` // Per-torrent configured seeding time limit (seconds). -1 means no limit, -2 means use global limit
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
		mask: mask(*ti),
	}
	// Adapt to JSON types
	tmp.AddedOn = ti.AddedOn.Unix()
	tmp.AmountLeft = int(ti.AmountLeft.Bytes())
	tmp.Completed = int(ti.Completed.Bytes())
	tmp.CompletionOn = ti.CompletionOn.Unix()
	tmp.DownloadSpeedLimit = ti.DownloadSpeedLimit.ToBytes()
	tmp.DownloadSpeed = ti.DownloadSpeed.ToBytes()
	tmp.Downloaded = int(ti.Downloaded.Bytes())
	tmp.DownloadedSession = int(ti.DownloadedSession.Bytes())
	tmp.ETA = int(ti.ETA.Seconds())
	tmp.LastActivity = ti.LastActivity.Unix()
	tmp.MagnetURI = ti.MagnetURI.String()
	tmp.MaxSeedingTime = int(ti.MaxSeedingTime.Seconds())
	tmp.Reannounce = int(ti.Reannounce.Seconds())
	tmp.SeedingTime = int(ti.SeedingTime.Seconds())
	tmp.SeedingTimeLimit = int(ti.SeedingTimeLimit.Seconds())
	tmp.SeenComplete = ti.SeenComplete.Unix()
	tmp.Size = int(ti.Size.Bytes())
	tmp.Tags = strings.Join(ti.Tags, tagListSeparator)
	tmp.TimeActive = int(ti.TimeActive.Seconds())
	tmp.TotalSize = int(ti.TotalSize.Bytes())
	tmp.UploadSpeedLimit = ti.UploadSpeedLimit.ToBytes()
	tmp.Uploaded = int(ti.Uploaded.Bytes())
	tmp.UploadedSession = int(ti.UploadedSession.Bytes())
	tmp.UploadSpeed = ti.UploadSpeed.ToBytes()
	return json.Marshal(tmp)
}

type TorrentState string

const (
	TorrentStateError               TorrentState = "error"              // Some error occurred, applies to paused torrents
	TorrentStateMissingFiles        TorrentState = "missingFiles"       // Torrent data files is missing
	TorrentStateUploading           TorrentState = "uploading"          // Torrent is being seeded and data is being transferred
	TorrentStatePausedUploading     TorrentState = "pausedUP"           // Torrent is paused and has finished downloading
	TorrentStateQueuedUploading     TorrentState = "queuedUP"           // Queuing is enabled and torrent is queued for upload
	TorrentStateStalledUploading    TorrentState = "stalledUP"          // Torrent is being seeded, but no connection were made
	TorrentStateCheckingUploading   TorrentState = "checkingUP"         // Torrent has finished downloading and is being checked
	TorrentStateForcedUploading     TorrentState = "forcedUP"           // Torrent is forced to uploading and ignore queue limit
	TorrentStateAllocating          TorrentState = "allocating"         // Torrent is allocating disk space for download
	TorrentStateDownloading         TorrentState = "downloading"        // Torrent is being downloaded and data is being transferred
	TorrentStateMetadataDownloading TorrentState = "metaDL"             // Torrent has just started downloading and is fetching metadata
	TorrentStatePausedDownloading   TorrentState = "pausedDL"           // Torrent is paused and has not finished downloading
	TorrentStateQueuedDownloading   TorrentState = "queuedDL"           // Queuing is enabled and torrent is queued for download
	TorrentStateStalledDownloading  TorrentState = "stalledDL"          // Torrent is being downloaded, but no connection were made
	TorrentStateCheckingDownloading TorrentState = "checkingDL"         // Same as TorrentStateCheckingUploading, but torrent has NOT finished downloading
	TorrentStateForcedDownloading   TorrentState = "forcedDL"           // Torrent is forced to downloading to ignore queue limit
	TorrentStateCheckingResumeData  TorrentState = "checkingResumeData" // Checking resume data on qBt startup
	TorrentStateMoving              TorrentState = "moving"             // Torrent is moving to another location
	TorrentStateUnknown             TorrentState = "unknown"            // Unknown torrent state
)

/*
	Generic properties
*/

// GetTorrentGenericProperties returns the generic properties of a torrent identified by its hash.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#get-torrent-generic-properties
func (c *Client) GetTorrentGenericProperties(ctx context.Context, hash string) (torrentProperties TorrentGenericProperties, err error) {
	// build request
	req, err := c.requestBuild(ctx, "GET", torrentsAPIName, "properties", map[string]string{
		"hash": hash,
	}, nil)
	if err != nil {
		err = fmt.Errorf("building request failed: %w", err)
		return
	}
	// execute request
	if err = c.requestExecute(req, &torrentProperties, true); err != nil {
		err = fmt.Errorf("executing request failed: %w", err)
	}
	return
}

// TorrentGenericProperties represents the generic properties of a torrent.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#get-torrent-generic-properties
type TorrentGenericProperties struct {
	SavePath               string        `json:"save_path"`                // Torrent save path
	CreationDate           time.Time     `json:"creation_date"`            // Torrent creation date
	PieceSize              cunits.Bits   `json:"piece_size"`               // Torrent piece size
	Comment                string        `json:"comment"`                  // Torrent comment
	TotalWasted            cunits.Bits   `json:"total_wasted"`             // Total data wasted for torrent
	TotalUploaded          cunits.Bits   `json:"total_uploaded"`           // Total data uploaded for torrent
	TotalUploadedSession   cunits.Bits   `json:"total_uploaded_session"`   // Total data uploaded this session
	TotalDownloaded        cunits.Bits   `json:"total_downloaded"`         // Total data downloaded for torrent
	TotalDownloadedSession cunits.Bits   `json:"total_downloaded_session"` // Total data downloaded this session
	UploadLimit            Speed         `json:"up_limit"`                 // Torrent upload limit
	DownloadLimit          Speed         `json:"dl_limit"`                 // Torrent download limit
	TimeElapsed            time.Duration `json:"time_elapsed"`             // Torrent elapsed time
	SeedingTime            time.Duration `json:"seeding_time"`             // Torrent elapsed time while complete
	NbConnections          int           `json:"nb_connections"`           // Torrent connection count
	NbConnectionsLimit     int           `json:"nb_connections_limit"`     // Torrent connection count limit
	ShareRatio             float64       `json:"share_ratio"`              // Torrent share ratio
	AdditionDate           time.Time     `json:"addition_date"`            // When this torrent was added
	CompletionDate         time.Time     `json:"completion_date"`          // Torrent completion date
	CreatedBy              string        `json:"created_by"`               // Torrent creator
	DownloadSpeedAvg       Speed         `json:"dl_speed_avg"`             // Torrent average download speed
	DownloadSpeed          Speed         `json:"dl_speed"`                 // Torrent download speed
	ETA                    time.Duration `json:"eta"`                      // Torrent ETA
	LastSeen               time.Time     `json:"last_seen"`                // Last seen complete date (unix timestamp)
	Peers                  int           `json:"peers"`                    // Number of peers connected to
	PeersTotal             int           `json:"peers_total"`              // Number of peers in the swarm
	PiecesHave             int           `json:"pieces_have"`              // Number of pieces owned
	PiecesNum              int           `json:"pieces_num"`               // Number of pieces of the torrent
	Reannounce             time.Duration `json:"reannounce"`               // Duration until the next announce
	Seeds                  int           `json:"seeds"`                    // Number of seeds connected to
	SeedsTotal             int           `json:"seeds_total"`              // Number of seeds in the swarm
	TotalSize              cunits.Bits   `json:"total_size"`               // Torrent total size
	UploadSpeedAvg         Speed         `json:"up_speed_avg"`             // Torrent average upload speed
	UploadSpeed            Speed         `json:"up_speed"`                 // Torrent upload speed
	Private                bool          `json:"isPrivate"`                // True if torrent is from a private tracker
}

func (tgp *TorrentGenericProperties) UnmarshalJSON(data []byte) (err error) {
	type mask TorrentGenericProperties
	tmp := struct {
		*mask
		// Custom unmarshaling
		CreationDate           int64 `json:"creation_date"`            // Torrent creation date (Unix timestamp)
		PieceSize              int   `json:"piece_size"`               // Torrent piece size (bytes)
		TotalWasted            int   `json:"total_wasted"`             // Total data wasted for torrent (bytes)
		TotalUploaded          int   `json:"total_uploaded"`           // Total data uploaded for torrent (bytes)
		TotalUploadedSession   int   `json:"total_uploaded_session"`   // Total data uploaded this session (bytes)
		TotalDownloaded        int   `json:"total_downloaded"`         // Total data downloaded for torrent (bytes)
		TotalDownloadedSession int   `json:"total_downloaded_session"` // Total data downloaded this session (bytes)
		UploadLimit            int   `json:"up_limit"`                 // Torrent upload limit (bytes/s)
		DownloadLimit          int   `json:"dl_limit"`                 // Torrent download limit (bytes/s)
		TimeElapsed            int64 `json:"time_elapsed"`             // Torrent elapsed time (seconds)
		SeedingTime            int   `json:"seeding_time"`             // Torrent elapsed time while complete (seconds)
		AdditionDate           int64 `json:"addition_date"`            // When this torrent was added (unix timestamp)
		CompletionDate         int64 `json:"completion_date"`          // Torrent completion date (unix timestamp)
		DownloadSpeedAvg       int   `json:"dl_speed_avg"`             // Torrent average download speed (bytes/second)
		DownloadSpeed          int   `json:"dl_speed"`                 // Torrent download speed (bytes/second)
		ETA                    int   `json:"eta"`                      // Torrent ETA (seconds)
		LastSeen               int64 `json:"last_seen"`                // Last seen complete date (unix timestamp)
		Reannounce             int   `json:"reannounce"`               // Number of seconds until the next announce
		TotalSize              int   `json:"total_size"`               // Torrent total size (bytes)
		UploadSpeedAvg         int   `json:"up_speed_avg"`             // Torrent average upload speed (bytes/second)
		UploadSpeed            int   `json:"up_speed"`                 // Torrent upload speed (bytes/second)
	}{
		mask: (*mask)(tgp),
	}
	// Unmarshall to tmp struct
	if err = json.Unmarshal(data, &tmp); err != nil {
		return
	}
	// Adapt to golang types
	tgp.CreationDate = time.Unix(tmp.CreationDate, 0)
	tgp.PieceSize = cunits.ImportInBytes(float64(tmp.PieceSize))
	tgp.TotalWasted = cunits.ImportInBytes(float64(tmp.TotalWasted))
	tgp.TotalUploaded = cunits.ImportInBytes(float64(tmp.TotalUploaded))
	tgp.TotalUploadedSession = cunits.ImportInBytes(float64(tmp.TotalUploadedSession))
	tgp.TotalDownloaded = cunits.ImportInBytes(float64(tmp.TotalDownloaded))
	tgp.TotalDownloadedSession = cunits.ImportInBytes(float64(tmp.TotalDownloadedSession))
	tgp.UploadLimit = GetSpeedFromBytes(tmp.UploadLimit)
	tgp.DownloadLimit = GetSpeedFromBytes(tmp.DownloadLimit)
	tgp.TimeElapsed = time.Duration(tmp.TimeElapsed) * time.Second
	tgp.SeedingTime = time.Duration(tmp.SeedingTime) * time.Second
	tgp.AdditionDate = time.Unix(tmp.AdditionDate, 0)
	tgp.CompletionDate = time.Unix(tmp.CompletionDate, 0)
	tgp.DownloadSpeedAvg = GetSpeedFromBytes(tmp.DownloadSpeedAvg)
	tgp.DownloadSpeed = GetSpeedFromBytes(tmp.DownloadSpeed)
	tgp.ETA = time.Duration(tmp.ETA) * time.Second
	tgp.LastSeen = time.Unix(tmp.LastSeen, 0)
	tgp.Reannounce = time.Duration(tmp.Reannounce) * time.Second
	tgp.TotalSize = cunits.ImportInBytes(float64(tmp.TotalSize))
	tgp.UploadSpeedAvg = GetSpeedFromBytes(tmp.UploadSpeedAvg)
	tgp.UploadSpeed = GetSpeedFromBytes(tmp.UploadSpeed)
	return
}

func (tgp *TorrentGenericProperties) MarshalJSON() ([]byte, error) {
	type mask TorrentGenericProperties
	tmp := struct {
		mask
		// Custom marshaling
		CreationDate           int64 `json:"creation_date"`            // Torrent creation date (Unix timestamp)
		PieceSize              int   `json:"piece_size"`               // Torrent piece size (bytes)
		TotalWasted            int   `json:"total_wasted"`             // Total data wasted for torrent (bytes)
		TotalUploaded          int   `json:"total_uploaded"`           // Total data uploaded for torrent (bytes)
		TotalUploadedSession   int   `json:"total_uploaded_session"`   // Total data uploaded this session (bytes)
		TotalDownloaded        int   `json:"total_downloaded"`         // Total data downloaded for torrent (bytes)
		TotalDownloadedSession int   `json:"total_downloaded_session"` // Total data downloaded this session (bytes)
		UploadLimit            int   `json:"up_limit"`                 // Torrent upload limit (bytes/s)
		DownloadLimit          int   `json:"dl_limit"`                 // Torrent download limit (bytes/s)
		TimeElapsed            int   `json:"time_elapsed"`             // Torrent elapsed time (seconds)
		SeedingTime            int   `json:"seeding_time"`             // Torrent elapsed time while complete (seconds)
		AdditionDate           int64 `json:"addition_date"`            // When this torrent was added (unix timestamp)
		CompletionDate         int64 `json:"completion_date"`          // Torrent completion date (unix timestamp)
		DownloadSpeedAvg       int   `json:"dl_speed_avg"`             // Torrent average download speed (bytes/second)
		DownloadSpeed          int   `json:"dl_speed"`                 // Torrent download speed (bytes/second)
		ETA                    int   `json:"eta"`                      // Torrent ETA (seconds)
		LastSeen               int64 `json:"last_seen"`                // Last seen complete date (unix timestamp)
		Reannounce             int   `json:"reannounce"`               // Number of seconds until the next announce
		TotalSize              int   `json:"total_size"`               // Torrent total size (bytes)
		UploadSpeedAvg         int   `json:"up_speed_avg"`             // Torrent average upload speed (bytes/second)
		UploadSpeed            int   `json:"up_speed"`                 // Torrent upload speed (bytes/second)
	}{
		mask:                   mask(*tgp),
		CreationDate:           tgp.CreationDate.Unix(),
		PieceSize:              int(tgp.PieceSize.Bytes()),
		TotalWasted:            int(tgp.TotalWasted.Bytes()),
		TotalUploaded:          int(tgp.TotalUploaded.Bytes()),
		TotalUploadedSession:   int(tgp.TotalUploadedSession.Bytes()),
		TotalDownloaded:        int(tgp.TotalDownloaded.Bytes()),
		TotalDownloadedSession: int(tgp.TotalDownloadedSession.Bytes()),
		UploadLimit:            tgp.UploadLimit.ToBytes(),
		DownloadLimit:          tgp.DownloadLimit.ToBytes(),
		TimeElapsed:            int(tgp.TimeElapsed.Seconds()),
		SeedingTime:            int(tgp.SeedingTime.Seconds()),
		AdditionDate:           tgp.AdditionDate.Unix(),
		CompletionDate:         tgp.CompletionDate.Unix(),
		DownloadSpeedAvg:       tgp.DownloadSpeedAvg.ToBytes(),
		DownloadSpeed:          tgp.DownloadSpeed.ToBytes(),
		ETA:                    int(tgp.ETA.Seconds()),
		LastSeen:               tgp.LastSeen.Unix(),
		Reannounce:             int(tgp.Reannounce.Seconds()),
		TotalSize:              int(tgp.TotalSize.Bytes()),
		UploadSpeedAvg:         tgp.UploadSpeedAvg.ToBytes(),
		UploadSpeed:            tgp.UploadSpeed.ToBytes(),
	}
	return json.Marshal(tmp)
}

/*
	torrent trackers
*/

// GetTorrentTrackers returns the trackers for a given torrent.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#get-torrent-trackers
func (c *Client) GetTorrentTrackers(ctx context.Context, hash string) (trackers []TorrentTracker, err error) {
	// build request
	req, err := c.requestBuild(ctx, "GET", torrentsAPIName, "trackers", map[string]string{
		"hash": hash,
	}, nil)
	if err != nil {
		err = fmt.Errorf("building request failed: %w", err)
		return
	}
	// execute request
	if err = c.requestExecute(req, &trackers, true); err != nil {
		err = fmt.Errorf("executing request failed: %w", err)
	}
	return
}

// TorrentTracker represents a single tracker for a torrent
type TorrentTracker struct {
	URL           *url.URL             `json:"url"`            // Tracker url
	Status        TorrentTrackerStatus `json:"status"`         // Tracker status. See the TorrentTrackerStatus constants below for possible values.
	Tier          int                  `json:"tier"`           // Tracker priority tier. Lower tier trackers are tried before higher tiers. Tier numbers are valid when >= 0, < 0 is used as placeholder when tier does not exist for special entries (such as DHT).
	NumPeers      int                  `json:"num_peers"`      // Number of peers for current torrent, as reported by the tracker
	NumSeeds      int                  `json:"num_seeds"`      // Number of seeds for current torrent, as reported by the tracker
	NumLeeches    int                  `json:"num_leeches"`    // Number of leeches for current torrent, as reported by the tracker
	NumDownloaded int                  `json:"num_downloaded"` // Number of completed downloads for current torrent, as reported by the tracker
	Message       string               `json:"msg"`            // Tracker message (there is no way of knowing what this message is - it's up to tracker admins)
}

func (tt *TorrentTracker) UnmarshalJSON(data []byte) (err error) {
	type mask TorrentTracker
	tmp := struct {
		*mask
		// Custom unmarshaling
		URL string `json:"url"` // Tracker url
	}{
		mask: (*mask)(tt),
	}
	// Unmarshall to tmp struct
	if err = json.Unmarshal(data, &tmp); err != nil {
		return
	}
	// Adapt to golang types
	if tt.URL, err = url.Parse(tmp.URL); err != nil {
		err = fmt.Errorf("parsing tracker URL failed: %w", err)
		return
	}
	return
}

func (tt *TorrentTracker) MarshalJSON() ([]byte, error) {
	type mask TorrentTracker
	tmp := struct {
		*mask
		// Custom marshaling
		URL string `json:"url"` // Tracker url
	}{
		mask: (*mask)(tt),
		URL:  tt.URL.String(), // Format URL as string
	}
	return json.Marshal(tmp)
}

// TorrentTrackerStatus represents the status of a tracker
type TorrentTrackerStatus uint8

const (
	TorrentTrackerDisabled     TorrentTrackerStatus = 0 // Tracker is disabled (used for DHT, PeX, and LSD)
	TorrentTrackerNotContacted TorrentTrackerStatus = 1 // Tracker has not been contacted yet
	TorrentTrackerWorking      TorrentTrackerStatus = 2 // Tracker has been contacted and is working
	TorrentTrackerUpdating     TorrentTrackerStatus = 3 // Tracker is updating
	TorrentTrackerNotWorking   TorrentTrackerStatus = 4 // Tracker has been contacted, but it is not working (or doesn't send proper replies)
)

func (tts TorrentTrackerStatus) String() string {
	switch tts {
	case TorrentTrackerDisabled:
		return "disabled"
	case TorrentTrackerNotContacted:
		return "not contacted"
	case TorrentTrackerWorking:
		return "working"
	case TorrentTrackerUpdating:
		return "updating"
	case TorrentTrackerNotWorking:
		return "not working"
	default:
		return strconv.Itoa(int(tts))
	}
}

/*
	delete torrents
*/

// DeleteTorrents delete one or more torrents.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#delete-torrents
func (c *Client) DeleteTorrents(ctx context.Context, hashes []string, deleteFiles bool) (err error) {
	// build request
	req, err := c.requestBuild(ctx, "POST", torrentsAPIName, "delete", map[string]string{
		"hashes":      strings.Join(hashes, hashListSeparator),
		"deleteFiles": strconv.FormatBool(deleteFiles),
	}, nil)
	if err != nil {
		err = fmt.Errorf("building request failed: %w", err)
		return
	}
	// execute request
	if err = c.requestExecute(req, nil, true); err != nil {
		err = fmt.Errorf("executing request failed: %w", err)
		return
	}
	return
}

/*
	new torrents
*/

// ReadTorrentsFiles reads the content of all files in the provided paths and returns a map of filename to content.
// Helper function for AddNewTorrents().
func ReadTorrentsFiles(paths []string) (content map[string][]byte, err error) {
	content = make(map[string][]byte, len(paths))
	for _, path := range paths {
		var fileContent []byte
		if fileContent, err = os.ReadFile(path); err != nil {
			err = fmt.Errorf("reading file %q failed: %w", path, err)
			return
		}
		content[filepath.Base(path)] = fileContent
	}
	return
}

// AddNewTorrentsOptions holds options for adding new torrents.
type AddNewTorrentsOptions struct {
	SavePath               *string        // Download folder
	Category               *string        // Category for the torrent
	Tags                   []string       // Tags for the torrent
	SkipChecking           *bool          // Skip hash checking
	Paused                 *bool          // Add torrents in the paused state
	RootFolder             *bool          // Create the root folder
	Rename                 *string        // Rename torrent
	UploadLimit            *Speed         // Set torrent upload speed limit (/sec)
	DownloadLimit          *Speed         // Set torrent download speed limit (/sec)
	RatioLimit             *float64       // Set torrent share ratio limit (since 2.8.1)
	SeedingTimeLimit       *time.Duration // Set torrent seeding time limit (since 2.8.1) (will be converted to minutes)
	AutoTMM                *bool          // Whether Automatic Torrent Management should be used
	SequentialDownload     *bool          // Enable sequential download
	FirstLastPiecePriority *bool          // Prioritize download first last piece
}

// AddNewTorrents adds new torrents. There must be at least one file content or URL.
// Check the ReadTorrentsFiles() helper for files content. options can be nil.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#add-new-torrent
func (c *Client) AddNewTorrents(ctx context.Context, files map[string][]byte, urls []*url.URL, options *AddNewTorrentsOptions) (err error) {
	if len(files) == 0 && len(urls) == 0 {
		err = fmt.Errorf("no files or URLs provided")
		return
	}
	// build payload
	payload, contentType, err := torrentAddGeneratePayload(files, urls, options)
	if err != nil {
		err = fmt.Errorf("generating payload failed: %w", err)
		return
	}
	// build request
	req, err := c.requestBuild(ctx, "POST", torrentsAPIName, "add", nil, &payload)
	if err != nil {
		err = fmt.Errorf("building request failed: %w", err)
		return
	}
	req.Header.Set(contentTypeHeader, contentType)
	// execute request
	// output is nil because the server response format varies across versions:
	// some return text/plain ("Ok."), others application/json. Relying on the
	// HTTP 200 status is sufficient and avoids version-specific decoding issues.
	if err = c.requestExecute(req, nil, true); err != nil {
		err = fmt.Errorf("executing request failed: %w", err)
	}
	return
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func createBtFormFile(w *multipart.Writer, filename string) (io.Writer, error) {
	h := make(textproto.MIMEHeader, 2)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="torrents"; filename="%s"`, quoteEscaper.Replace(filename)))
	h.Set("Content-Type", "application/x-bittorrent")
	return w.CreatePart(h)
}

/*
	torrent web seeds
*/

// GetTorrentWebSeeds returns the list of web seeds for a torrent.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#get-torrent-web-seeds
func (c *Client) GetTorrentWebSeeds(ctx context.Context, hash string) (webSeeds []TorrentWebSeed, err error) {
	req, err := c.requestBuild(ctx, "GET", torrentsAPIName, "webseeds", map[string]string{
		"hash": hash,
	}, nil)
	if err != nil {
		err = fmt.Errorf("building request failed: %w", err)
		return
	}
	if err = c.requestExecute(req, &webSeeds, true); err != nil {
		err = fmt.Errorf("executing request failed: %w", err)
	}
	return
}

// TorrentWebSeed represents a single web seed for a torrent.
type TorrentWebSeed struct {
	URL *url.URL `json:"url"` // Web seed URL
}

func (tw *TorrentWebSeed) UnmarshalJSON(data []byte) (err error) {
	type mask TorrentWebSeed
	tmp := struct {
		*mask
		URL string `json:"url"`
	}{
		mask: (*mask)(tw),
	}
	if err = json.Unmarshal(data, &tmp); err != nil {
		return
	}
	if tw.URL, err = url.Parse(tmp.URL); err != nil {
		err = fmt.Errorf("parsing web seed URL failed: %w", err)
		return
	}
	return
}

func (tw *TorrentWebSeed) MarshalJSON() ([]byte, error) {
	type mask TorrentWebSeed
	tmp := struct {
		*mask
		URL string `json:"url"`
	}{
		mask: (*mask)(tw),
		URL:  tw.URL.String(),
	}
	return json.Marshal(tmp)
}

/*
	torrent contents
*/

// GetTorrentContents returns the contents of a torrent.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#get-torrent-contents
func (c *Client) GetTorrentContents(ctx context.Context, hash string, indexes []int) (contents []TorrentContent, err error) {
	params := map[string]string{
		"hash": hash,
	}
	if len(indexes) > 0 {
		indexStrs := make([]string, len(indexes))
		for i, idx := range indexes {
			indexStrs[i] = strconv.Itoa(idx)
		}
		params["indexes"] = strings.Join(indexStrs, hashListSeparator)
	}
	req, err := c.requestBuild(ctx, "GET", torrentsAPIName, "files", params, nil)
	if err != nil {
		err = fmt.Errorf("building request failed: %w", err)
		return
	}
	if err = c.requestExecute(req, &contents, true); err != nil {
		err = fmt.Errorf("executing request failed: %w", err)
	}
	return
}

// TorrentContent represents a single file within a torrent.
type TorrentContent struct {
	Index        int     `json:"index"`        // File index
	Name         string  `json:"name"`         // File name (including relative path)
	Size         int64   `json:"size"`         // File size (bytes)
	Progress     float64 `json:"progress"`     // File progress (percentage/100)
	Priority     int     `json:"priority"`     // File priority
	IsSeed       bool    `json:"is_seed"`      // True if file is seeding/complete
	PieceRange   []int   `json:"piece_range"`  // First and last piece index (inclusive)
	Availability float64 `json:"availability"` // Percentage of file pieces currently available
}

/*
	torrent pieces states
*/

// GetTorrentPiecesStates returns the states of all pieces of a torrent.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#get-torrent-pieces-states
func (c *Client) GetTorrentPiecesStates(ctx context.Context, hash string) (states []int, err error) {
	req, err := c.requestBuild(ctx, "GET", torrentsAPIName, "pieceStates", map[string]string{
		"hash": hash,
	}, nil)
	if err != nil {
		err = fmt.Errorf("building request failed: %w", err)
		return
	}
	if err = c.requestExecute(req, &states, true); err != nil {
		err = fmt.Errorf("executing request failed: %w", err)
	}
	return
}

/*
	torrent pieces hashes
*/

// GetTorrentPiecesHashes returns the hashes of all pieces of a torrent.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#get-torrent-pieces-hashes
func (c *Client) GetTorrentPiecesHashes(ctx context.Context, hash string) (hashes []string, err error) {
	req, err := c.requestBuild(ctx, "GET", torrentsAPIName, "pieceHashes", map[string]string{
		"hash": hash,
	}, nil)
	if err != nil {
		err = fmt.Errorf("building request failed: %w", err)
		return
	}
	if err = c.requestExecute(req, &hashes, true); err != nil {
		err = fmt.Errorf("executing request failed: %w", err)
	}
	return
}

/*
	pause torrents
*/

// StopTorrents pauses one or more torrents.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#pause-torrents
func (c *Client) StopTorrents(ctx context.Context, hashes []string) (err error) {
	req, err := c.requestBuild(ctx, "POST", torrentsAPIName, "stop", map[string]string{
		"hashes": strings.Join(hashes, hashListSeparator),
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

/*
	resume torrents
*/

// StartTorrents resumes one or more torrents.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#resume-torrents
func (c *Client) StartTorrents(ctx context.Context, hashes []string) (err error) {
	req, err := c.requestBuild(ctx, "POST", torrentsAPIName, "start", map[string]string{
		"hashes": strings.Join(hashes, hashListSeparator),
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

/*
	recheck torrents
*/

// RecheckTorrents rechecks one or more torrents.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#recheck-torrents
func (c *Client) RecheckTorrents(ctx context.Context, hashes []string) (err error) {
	req, err := c.requestBuild(ctx, "POST", torrentsAPIName, "recheck", map[string]string{
		"hashes": strings.Join(hashes, hashListSeparator),
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

/*
	reannounce torrents
*/

// ReannounceTorrents reannounces one or more torrents.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#reannounce-torrents
func (c *Client) ReannounceTorrents(ctx context.Context, hashes []string) (err error) {
	req, err := c.requestBuild(ctx, "POST", torrentsAPIName, "reannounce", map[string]string{
		"hashes": strings.Join(hashes, hashListSeparator),
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

/*
	increase torrent priority
*/

// IncreaseTorrentPriority increases the priority of one or more torrents.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#increase-torrent-priority
func (c *Client) IncreaseTorrentPriority(ctx context.Context, hashes []string) (err error) {
	req, err := c.requestBuild(ctx, "POST", torrentsAPIName, "increasePrio", map[string]string{
		"hashes": strings.Join(hashes, hashListSeparator),
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

/*
	decrease torrent priority
*/

// DecreaseTorrentPriority decreases the priority of one or more torrents.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#decrease-torrent-priority
func (c *Client) DecreaseTorrentPriority(ctx context.Context, hashes []string) (err error) {
	req, err := c.requestBuild(ctx, "POST", torrentsAPIName, "decreasePrio", map[string]string{
		"hashes": strings.Join(hashes, hashListSeparator),
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

/*
	maximal torrent priority
*/

// TopTorrentPriority sets the maximum priority for one or more torrents.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#maximal-torrent-priority
func (c *Client) TopTorrentPriority(ctx context.Context, hashes []string) (err error) {
	req, err := c.requestBuild(ctx, "POST", torrentsAPIName, "topPrio", map[string]string{
		"hashes": strings.Join(hashes, hashListSeparator),
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

/*
	minimal torrent priority
*/

// BottomTorrentPriority sets the minimum priority for one or more torrents.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#minimal-torrent-priority
func (c *Client) BottomTorrentPriority(ctx context.Context, hashes []string) (err error) {
	req, err := c.requestBuild(ctx, "POST", torrentsAPIName, "bottomPrio", map[string]string{
		"hashes": strings.Join(hashes, hashListSeparator),
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

/*
	toggle sequential download
*/

// ToggleSequentialDownload toggles sequential download for one or more torrents.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#toggle-sequential-download
func (c *Client) ToggleSequentialDownload(ctx context.Context, hashes []string) (err error) {
	req, err := c.requestBuild(ctx, "POST", torrentsAPIName, "toggleSequentialDownload", map[string]string{
		"hashes": strings.Join(hashes, hashListSeparator),
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

/*
	toggle first/last piece priority
*/

// ToggleFirstLastPiecePrio toggles first/last piece priority for one or more torrents.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#set-firstlast-piece-priority
func (c *Client) ToggleFirstLastPiecePrio(ctx context.Context, hashes []string) (err error) {
	req, err := c.requestBuild(ctx, "POST", torrentsAPIName, "toggleFirstLastPiecePrio", map[string]string{
		"hashes": strings.Join(hashes, hashListSeparator),
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

/*
	set force start
*/

// SetForceStart sets the force start state for one or more torrents.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#set-force-start
func (c *Client) SetForceStart(ctx context.Context, hashes []string, value bool) (err error) {
	req, err := c.requestBuild(ctx, "POST", torrentsAPIName, "setForceStart", map[string]string{
		"hashes": strings.Join(hashes, hashListSeparator),
		"value":  strconv.FormatBool(value),
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

/*
	set super seeding
*/

// SetSuperSeeding sets the super seeding state for one or more torrents.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#set-super-seeding
func (c *Client) SetSuperSeeding(ctx context.Context, hashes []string, value bool) (err error) {
	req, err := c.requestBuild(ctx, "POST", torrentsAPIName, "setSuperSeeding", map[string]string{
		"hashes": strings.Join(hashes, hashListSeparator),
		"value":  strconv.FormatBool(value),
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

/*
	set automatic torrent management
*/

// SetAutoManagement sets the automatic torrent management state for one or more torrents.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#set-automatic-torrent-management
func (c *Client) SetAutoManagement(ctx context.Context, hashes []string, value bool) (err error) {
	req, err := c.requestBuild(ctx, "POST", torrentsAPIName, "setAutoManagement", map[string]string{
		"hashes": strings.Join(hashes, hashListSeparator),
		"enable": strconv.FormatBool(value),
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

/*
	add trackers
*/

// AddTrackers adds trackers to a torrent.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#add-trackers-to-torrent
func (c *Client) AddTrackers(ctx context.Context, hash string, urls []string) (err error) {
	req, err := c.requestBuild(ctx, "POST", torrentsAPIName, "addTrackers", map[string]string{
		"hash": hash,
		"urls": strings.Join(urls, "\n"),
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

/*
	edit trackers
*/

// EditTracker edits a tracker URL for a torrent.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#edit-trackers
func (c *Client) EditTracker(ctx context.Context, hash string, origUrl string, newUrl string) (err error) {
	req, err := c.requestBuild(ctx, "POST", torrentsAPIName, "editTracker", map[string]string{
		"hash":    hash,
		"origUrl": origUrl,
		"newUrl":  newUrl,
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

/*
	remove trackers
*/

// RemoveTrackers removes trackers from a torrent.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#remove-trackers
func (c *Client) RemoveTrackers(ctx context.Context, hash string, urls []string) (err error) {
	req, err := c.requestBuild(ctx, "POST", torrentsAPIName, "removeTrackers", map[string]string{
		"hash": hash,
		"urls": strings.Join(urls, "|"),
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

/*
	add peers
*/

// AddPeers adds peers to torrents.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#add-peers
func (c *Client) AddPeers(ctx context.Context, hashes []string, peers []string) (err error) {
	req, err := c.requestBuild(ctx, "POST", torrentsAPIName, "addPeers", map[string]string{
		"hashes": strings.Join(hashes, hashListSeparator),
		"peers":  strings.Join(peers, "|"),
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

/*
	get torrent download limit
*/

// GetTorrentDownloadLimit returns the download limit for the given torrents.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#get-torrent-download-limit
func (c *Client) GetTorrentDownloadLimit(ctx context.Context, hashes []string) (limits map[string]int, err error) {
	req, err := c.requestBuild(ctx, "POST", torrentsAPIName, "downloadLimit", map[string]string{
		"hashes": strings.Join(hashes, hashListSeparator),
	}, nil)
	if err != nil {
		err = fmt.Errorf("building request failed: %w", err)
		return
	}
	if err = c.requestExecute(req, &limits, true); err != nil {
		err = fmt.Errorf("executing request failed: %w", err)
	}
	return
}

/*
	set torrent download limit
*/

// GetTorrentDownloadLimit
// SetTorrentDownloadLimit sets the download limit for the given torrents.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#set-torrent-download-limit
func (c *Client) SetTorrentDownloadLimit(ctx context.Context, hashes []string, limit Speed) (err error) {
	req, err := c.requestBuild(ctx, "POST", torrentsAPIName, "setDownloadLimit", map[string]string{
		"hashes": strings.Join(hashes, hashListSeparator),
		"limit":  strconv.Itoa(limit.ToBytes()),
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

/*
	get torrent upload limit
*/

// GetTorrentUploadLimit returns the upload limit for the given torrents.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#get-torrent-upload-limit
func (c *Client) GetTorrentUploadLimit(ctx context.Context, hashes []string) (limits map[string]int, err error) {
	req, err := c.requestBuild(ctx, "POST", torrentsAPIName, "uploadLimit", map[string]string{
		"hashes": strings.Join(hashes, hashListSeparator),
	}, nil)
	if err != nil {
		err = fmt.Errorf("building request failed: %w", err)
		return
	}
	if err = c.requestExecute(req, &limits, true); err != nil {
		err = fmt.Errorf("executing request failed: %w", err)
	}
	return
}

/*
	set torrent upload limit
*/

// SetTorrentUploadLimit sets the upload limit for the given torrents.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#set-torrent-upload-limit
func (c *Client) SetTorrentUploadLimit(ctx context.Context, hashes []string, limit Speed) (err error) {
	req, err := c.requestBuild(ctx, "POST", torrentsAPIName, "setUploadLimit", map[string]string{
		"hashes": strings.Join(hashes, hashListSeparator),
		"limit":  strconv.Itoa(limit.ToBytes()),
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

/*
	set torrent share limit
*/

// SetTorrentShareLimits sets the share limits for the given torrents.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#set-torrent-share-limit
func (c *Client) SetTorrentShareLimits(ctx context.Context, hashes []string, ratioLimit float64, seedingTimeLimit int, inactiveSeedingTimeLimit int) (err error) {
	req, err := c.requestBuild(ctx, "POST", torrentsAPIName, "setShareLimits", map[string]string{
		"hashes":                   strings.Join(hashes, hashListSeparator),
		"ratioLimit":               strconv.FormatFloat(ratioLimit, 'f', -1, 64),
		"seedingTimeLimit":         strconv.Itoa(seedingTimeLimit),
		"inactiveSeedingTimeLimit": strconv.Itoa(inactiveSeedingTimeLimit),
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

/*
	set torrent location
*/

// SetTorrentLocation sets the download location for the given torrents.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#set-torrent-location
func (c *Client) SetTorrentLocation(ctx context.Context, hashes []string, location string) (err error) {
	req, err := c.requestBuild(ctx, "POST", torrentsAPIName, "setLocation", map[string]string{
		"hashes":   strings.Join(hashes, hashListSeparator),
		"location": location,
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

/*
	set torrent category
*/

// SetTorrentCategory sets the category for one or more torrents.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#set-torrent-category
func (c *Client) SetTorrentCategory(ctx context.Context, hashes []string, category string) (err error) {
	req, err := c.requestBuild(ctx, "POST", torrentsAPIName, "setCategory", map[string]string{
		"hashes":   strings.Join(hashes, hashListSeparator),
		"category": category,
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

/*
	get all categories
*/

// Category represents a torrent category.
type Category struct {
	Name     string `json:"name"`     // Category name
	SavePath string `json:"savePath"` // Category save path
}

// GetAllCategories returns all categories.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#get-all-categories
func (c *Client) GetAllCategories(ctx context.Context) (categories map[string]Category, err error) {
	req, err := c.requestBuild(ctx, "GET", torrentsAPIName, "categories", nil, nil)
	if err != nil {
		err = fmt.Errorf("building request failed: %w", err)
		return
	}
	if err = c.requestExecute(req, &categories, true); err != nil {
		err = fmt.Errorf("executing request failed: %w", err)
	}
	return
}

/*
	add new category
*/

// CreateCategory creates a new category.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#create-category
func (c *Client) CreateCategory(ctx context.Context, category string, savePath string) (err error) {
	req, err := c.requestBuild(ctx, "POST", torrentsAPIName, "createCategory", map[string]string{
		"category": category,
		"savePath": savePath,
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

/*
	edit category
*/

// EditCategory edits an existing category.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#edit-category
func (c *Client) EditCategory(ctx context.Context, category string, savePath string) (err error) {
	req, err := c.requestBuild(ctx, "POST", torrentsAPIName, "editCategory", map[string]string{
		"category": category,
		"savePath": savePath,
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

/*
	remove categories
*/

// RemoveCategories removes one or more categories.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#remove-categories
func (c *Client) RemoveCategories(ctx context.Context, categories []string) (err error) {
	req, err := c.requestBuild(ctx, "POST", torrentsAPIName, "removeCategories", map[string]string{
		"categories": strings.Join(categories, "\n"),
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

/*
	get all tags
*/

// GetAllTags returns all tags.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#get-all-tags
func (c *Client) GetAllTags(ctx context.Context) (tags []string, err error) {
	req, err := c.requestBuild(ctx, "GET", torrentsAPIName, "tags", nil, nil)
	if err != nil {
		err = fmt.Errorf("building request failed: %w", err)
		return
	}
	if err = c.requestExecute(req, &tags, true); err != nil {
		err = fmt.Errorf("executing request failed: %w", err)
	}
	return
}

/*
	create tags
*/

// CreateTags creates new tags.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#create-tags
func (c *Client) CreateTags(ctx context.Context, tags []string) (err error) {
	req, err := c.requestBuild(ctx, "POST", torrentsAPIName, "createTags", map[string]string{
		"tags": strings.Join(tags, tagListSeparator),
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

/*
	delete tags
*/

// DeleteTags deletes one or more tags.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#delete-tags
func (c *Client) DeleteTags(ctx context.Context, tags []string) (err error) {
	req, err := c.requestBuild(ctx, "POST", torrentsAPIName, "deleteTags", map[string]string{
		"tags": strings.Join(tags, tagListSeparator),
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

/*
	add torrent tags
*/

// AddTorrentTags adds tags to one or more torrents.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#add-torrent-tags
func (c *Client) AddTorrentTags(ctx context.Context, hashes []string, tags []string) (err error) {
	req, err := c.requestBuild(ctx, "POST", torrentsAPIName, "addTags", map[string]string{
		"hashes": strings.Join(hashes, hashListSeparator),
		"tags":   strings.Join(tags, tagListSeparator),
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

/*
	remove torrent tags
*/

// RemoveTorrentTags removes tags from one or more torrents.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#remove-torrent-tags
func (c *Client) RemoveTorrentTags(ctx context.Context, hashes []string, tags []string) (err error) {
	req, err := c.requestBuild(ctx, "POST", torrentsAPIName, "removeTags", map[string]string{
		"hashes": strings.Join(hashes, hashListSeparator),
		"tags":   strings.Join(tags, tagListSeparator),
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

/*
	set torrent name
*/

// RenameTorrent renames a torrent.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#set-torrent-name
func (c *Client) RenameTorrent(ctx context.Context, hash string, name string) (err error) {
	req, err := c.requestBuild(ctx, "POST", torrentsAPIName, "rename", map[string]string{
		"hash": hash,
		"name": name,
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

/*
	rename file
*/

// RenameFile renames a file in a torrent.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#rename-file
func (c *Client) RenameFile(ctx context.Context, hash string, oldPath string, newPath string) (err error) {
	req, err := c.requestBuild(ctx, "POST", torrentsAPIName, "renameFile", map[string]string{
		"hash":    hash,
		"oldPath": oldPath,
		"newPath": newPath,
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

/*
	rename folder
*/

// RenameFolder renames a folder in a torrent.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#rename-folder
func (c *Client) RenameFolder(ctx context.Context, hash string, oldPath string, newPath string) (err error) {
	req, err := c.requestBuild(ctx, "POST", torrentsAPIName, "renameFolder", map[string]string{
		"hash":    hash,
		"oldPath": oldPath,
		"newPath": newPath,
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

/*
	set file priority
*/

// SetFilePriority sets the priority for files in a torrent.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#set-file-priority
func (c *Client) SetFilePriority(ctx context.Context, hash string, ids []int, priority int) (err error) {
	idStrs := make([]string, len(ids))
	for i, id := range ids {
		idStrs[i] = strconv.Itoa(id)
	}
	req, err := c.requestBuild(ctx, "POST", torrentsAPIName, "filePrio", map[string]string{
		"hash":     hash,
		"id":       strings.Join(idStrs, hashListSeparator),
		"priority": strconv.Itoa(priority),
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

/*
	new torrents payload helper
*/

func torrentAddGeneratePayload(files map[string][]byte, urls []*url.URL, options *AddNewTorrentsOptions) (payload bytes.Buffer, contentType string, err error) {
	mp := multipart.NewWriter(&payload)
	defer func() {
		if err == nil {
			if err = mp.Close(); err != nil {
				err = fmt.Errorf("closing multipart writer failed: %w", err)
			}
		}
	}()
	contentType = mp.FormDataContentType()
	// Add raw files
	var mpw io.Writer
	for filename, content := range files {
		if filepath.Ext(filename) != ".torrent" {
			err = fmt.Errorf("file %q is not a .torrent file", filename)
			return
		}
		if len(content) == 0 {
			err = fmt.Errorf("file %q is empty", filename)
			return
		}
		if mpw, err = createBtFormFile(mp, filename); err != nil {
			err = fmt.Errorf("creating form file %s failed: %w", filename, err)
			return
		}
		if _, err = mpw.Write(content); err != nil {
			err = fmt.Errorf("writing file %q content failed: %w", filename, err)
			return
		}
	}
	// Add URLs
	strURLs := make([]string, len(urls))
	for index, tURL := range urls {
		// check URL
		if tURL == nil {
			err = fmt.Errorf("nil URL")
			return
		}
		switch tURL.Scheme {
		case "http", "https":
			if tURL.Host == "" {
				err = fmt.Errorf("invalid URL %q: missing host", tURL.String())
				return
			}
			if tURL.Path == "" {
				err = fmt.Errorf("invalid URL %q: missing path", tURL.String())
				return
			}
		case "magnet":
			if tURL.Host != "" {
				err = fmt.Errorf("invalid URL %q: magnet URL should not have a host", tURL.String())
				return
			}
			if tURL.Path != "" {
				err = fmt.Errorf("invalid URL %q: magnet URL should not have a path", tURL.String())
				return
			}
			if !tURL.Query().Has("xt") {
				err = fmt.Errorf("invalid URI %q: magnet URI should have an xt parameter", tURL.String())
				return
			}
		case "bc":
			// bc ??
			if tURL.Host != "bt" {
				err = fmt.Errorf("invalid URL %q: invalid host for bc protocol", tURL.String())
				return
			}
		default:
			err = fmt.Errorf("invalid URL %q: unsupported protocol", tURL.String())
			return
		}
		strURLs[index] = tURL.String()
	}
	if err = mp.WriteField("urls", strings.Join(strURLs, "\n")); err != nil {
		err = fmt.Errorf("writing urls to form field failed: %w", err)
		return
	}
	// Handles options
	if options == nil {
		return
	}
	if options.SavePath != nil {
		if err = mp.WriteField("savepath", *options.SavePath); err != nil {
			err = fmt.Errorf("writing savepath to form field failed: %w", err)
			return
		}
	}
	if options.Category != nil {
		if err = mp.WriteField("category", *options.Category); err != nil {
			err = fmt.Errorf("writing category to form field failed: %w", err)
			return
		}
	}
	if options.Tags != nil {
		if err = mp.WriteField("tags", strings.Join(options.Tags, ",")); err != nil {
			err = fmt.Errorf("writing tags to form field failed: %w", err)
			return
		}
	}
	if options.SkipChecking != nil {
		if err = mp.WriteField("skip_checking", strconv.FormatBool(*options.SkipChecking)); err != nil {
			err = fmt.Errorf("writing skip_checking to form field failed: %w", err)
			return
		}
	}
	if options.Paused != nil {
		if err = mp.WriteField("paused", strconv.FormatBool(*options.Paused)); err != nil {
			err = fmt.Errorf("writing paused to form field failed: %w", err)
			return
		}
	}
	if options.RootFolder != nil {
		if err = mp.WriteField("root_folder", strconv.FormatBool(*options.RootFolder)); err != nil {
			err = fmt.Errorf("writing root_folder to form field failed: %w", err)
			return
		}
	}
	if options.Rename != nil {
		if err = mp.WriteField("rename", *options.Rename); err != nil {
			err = fmt.Errorf("writing rename to form field failed: %w", err)
			return
		}
	}
	if options.UploadLimit != nil {
		if err = mp.WriteField("upLimit", strconv.Itoa(options.UploadLimit.ToBytes())); err != nil {
			err = fmt.Errorf("writing upLimit to form field failed: %w", err)
			return
		}
	}
	if options.DownloadLimit != nil {
		if err = mp.WriteField("dlLimit", strconv.Itoa(options.DownloadLimit.ToBytes())); err != nil {
			err = fmt.Errorf("writing dlLimit to form field failed: %w", err)
			return
		}
	}
	if options.RatioLimit != nil {
		if err = mp.WriteField("ratioLimit", strconv.FormatFloat(*options.RatioLimit, 'f', -1, 64)); err != nil {
			err = fmt.Errorf("writing ratioLimit to form field failed: %w", err)
			return
		}
	}
	if options.SeedingTimeLimit != nil {
		if err = mp.WriteField("seedingTimeLimit", strconv.Itoa(int(options.SeedingTimeLimit.Minutes()))); err != nil {
			err = fmt.Errorf("writing seedingTimeLimit to form field failed: %w", err)
			return
		}
	}
	if options.AutoTMM != nil {
		if err = mp.WriteField("autoTMM", strconv.FormatBool(*options.AutoTMM)); err != nil {
			err = fmt.Errorf("writing autoTMM to form field failed: %w", err)
			return
		}
	}
	if options.SequentialDownload != nil {
		if err = mp.WriteField("sequentialDownload", strconv.FormatBool(*options.SequentialDownload)); err != nil {
			err = fmt.Errorf("writing sequentialDownload to form field failed: %w", err)
			return
		}
	}
	if options.FirstLastPiecePriority != nil {
		if err = mp.WriteField("firstLastPiecePrio", strconv.FormatBool(*options.FirstLastPiecePriority)); err != nil {
			err = fmt.Errorf("writing firstLastPiecePrio to form field failed: %w", err)
			return
		}
	}
	return
}
