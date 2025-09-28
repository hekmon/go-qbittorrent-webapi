package qbtapi

import (
	"context"
	"encoding/json"
	"fmt"
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
	ti.DownloadSpeedLimit = GetSpeedFromBytes(tmp.DownloadSpeedLimit)
	ti.DownloadSpeed = GetSpeedFromBytes(tmp.DownloadSpeed)
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
	ti.UploadSpeedLimit = GetSpeedFromBytes(tmp.UploadSpeedLimit)
	ti.Uploaded = cunits.ImportInBytes(float64(tmp.Uploaded))
	ti.UploadedSession = cunits.ImportInBytes(float64(tmp.UploadedSession))
	ti.UploadSpeed = GetSpeedFromBytes(tmp.UploadSpeed)
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
	tmp.DownloadSpeedLimit = ti.DownloadSpeedLimit.ToBytes()
	tmp.DownloadSpeed = ti.DownloadSpeed.ToBytes()
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
	req, err := c.requestBuild(ctx, "GET", torrentsAPIName, "properties", map[string]string{"hash": hash})
	if err != nil {
		err = fmt.Errorf("request building failure: %w", err)
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
	Reannounce             int           `json:"reannounce"`               // Number of seconds until the next announce
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
	tgp.TotalSize = cunits.ImportInBytes(float64(tmp.TotalSize))
	tgp.UploadSpeedAvg = GetSpeedFromBytes(tmp.UploadSpeedAvg)
	tgp.UploadSpeed = GetSpeedFromBytes(tmp.UploadSpeed)
	return
}

func (tgp *TorrentGenericProperties) MarshalJSON() ([]byte, error) {
	type mask TorrentGenericProperties
	tmp := struct {
		*mask
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
		TotalSize              int   `json:"total_size"`               // Torrent total size (bytes)
		UploadSpeedAvg         int   `json:"up_speed_avg"`             // Torrent average upload speed (bytes/second)
		UploadSpeed            int   `json:"up_speed"`                 // Torrent upload speed (bytes/second)
	}{
		mask:                   (*mask)(tgp),
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
		TotalSize:              int(tgp.TotalSize.Bytes()),
		UploadSpeedAvg:         tgp.UploadSpeedAvg.ToBytes(),
		UploadSpeed:            tgp.UploadSpeed.ToBytes(),
	}
	return json.Marshal(tmp)
}
