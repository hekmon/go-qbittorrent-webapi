package qbtapi

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

const logAPIName = "log"

// LogMessageType represents the type of a log message.
type LogMessageType int

const (
	LogMessageTypeNormal   LogMessageType = 1
	LogMessageTypeInfo     LogMessageType = 2
	LogMessageTypeWarning  LogMessageType = 4
	LogMessageTypeCritical LogMessageType = 8
)

// LogEntry represents a single entry in the qBittorrent log.
type LogEntry struct {
	ID        int            `json:"id"`
	Message   string         `json:"message"`
	Timestamp time.Time      `json:"timestamp"`
	Type      LogMessageType `json:"type"`
}

// UnmarshalJSON implements json.Unmarshaler to convert the Unix timestamp into a time.Time.
func (le *LogEntry) UnmarshalJSON(data []byte) error {
	type mask LogEntry
	tmp := struct {
		*mask
		Timestamp int64 `json:"timestamp"`
	}{
		mask: (*mask)(le),
	}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	le.Timestamp = time.Unix(tmp.Timestamp, 0)
	return nil
}

// MarshalJSON implements json.Marshaler to convert the time.Time into a Unix timestamp.
func (le LogEntry) MarshalJSON() ([]byte, error) {
	type mask LogEntry
	tmp := struct {
		mask
		Timestamp int64 `json:"timestamp"`
	}{
		mask: mask(le),
	}
	if !le.Timestamp.IsZero() {
		tmp.Timestamp = le.Timestamp.Unix()
	}
	return json.Marshal(tmp)
}

// PeerLogEntry represents a single entry in the peer log.
type PeerLogEntry struct {
	ID        int       `json:"id"`
	IP        string    `json:"ip"`
	Timestamp time.Time `json:"timestamp"`
	Blocked   bool      `json:"blocked"`
	Reason    string    `json:"reason"`
}

// UnmarshalJSON implements json.Unmarshaler to convert the Unix timestamp into a time.Time.
func (ple *PeerLogEntry) UnmarshalJSON(data []byte) error {
	type mask PeerLogEntry
	tmp := struct {
		*mask
		Timestamp int64 `json:"timestamp"`
	}{
		mask: (*mask)(ple),
	}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	ple.Timestamp = time.Unix(tmp.Timestamp, 0)
	return nil
}

// MarshalJSON implements json.Marshaler to convert the time.Time into a Unix timestamp.
func (ple PeerLogEntry) MarshalJSON() ([]byte, error) {
	type mask PeerLogEntry
	tmp := struct {
		mask
		Timestamp int64 `json:"timestamp"`
	}{
		mask: mask(ple),
	}
	if !ple.Timestamp.IsZero() {
		tmp.Timestamp = ple.Timestamp.Unix()
	}
	return json.Marshal(tmp)
}

// LogFilters contains optional filters for the GetLog request.
type LogFilters struct {
	Normal      *bool // Include normal messages (default: true)
	Info        *bool // Include info messages (default: true)
	Warning     *bool // Include warning messages (default: true)
	Critical    *bool // Include critical messages (default: true)
	LastKnownID *int  // Exclude messages with "message id" <= last_known_id (default: -1)
}

func (lf LogFilters) getLowLevelRepr() map[string]string {
	params := make(map[string]string)
	if lf.Normal != nil {
		params["normal"] = strconv.FormatBool(*lf.Normal)
	}
	if lf.Info != nil {
		params["info"] = strconv.FormatBool(*lf.Info)
	}
	if lf.Warning != nil {
		params["warning"] = strconv.FormatBool(*lf.Warning)
	}
	if lf.Critical != nil {
		params["critical"] = strconv.FormatBool(*lf.Critical)
	}
	if lf.LastKnownID != nil {
		params["last_known_id"] = strconv.Itoa(*lf.LastKnownID)
	}
	return params
}

// PeerLogFilters contains optional filters for the GetPeerLog request.
type PeerLogFilters struct {
	LastKnownID *int // Exclude messages with "message id" <= last_known_id (default: -1)
}

func (plf PeerLogFilters) getLowLevelRepr() map[string]string {
	params := make(map[string]string)
	if plf.LastKnownID != nil {
		params["last_known_id"] = strconv.Itoa(*plf.LastKnownID)
	}
	return params
}

// GetLog returns the qBittorrent log entries.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#get-log
func (c *Client) GetLog(ctx context.Context, filters *LogFilters) (entries []LogEntry, err error) {
	var params map[string]string
	if filters != nil {
		params = filters.getLowLevelRepr()
	}
	req, err := c.requestBuild(ctx, "GET", logAPIName, "main", params, nil)
	if err != nil {
		err = fmt.Errorf("building request failed: %w", err)
		return
	}
	if err = c.requestExecute(req, &entries, true); err != nil {
		err = fmt.Errorf("executing request failed: %w", err)
	}
	return
}

// GetPeerLog returns the peer log entries.
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#get-peer-log
func (c *Client) GetPeerLog(ctx context.Context, filters *PeerLogFilters) (entries []PeerLogEntry, err error) {
	var params map[string]string
	if filters != nil {
		params = filters.getLowLevelRepr()
	}
	req, err := c.requestBuild(ctx, "GET", logAPIName, "peers", params, nil)
	if err != nil {
		err = fmt.Errorf("building request failed: %w", err)
		return
	}
	if err = c.requestExecute(req, &entries, true); err != nil {
		err = fmt.Errorf("executing request failed: %w", err)
	}
	return
}
