# qBittorrent Web API

[![Go Reference](https://pkg.go.dev/badge/github.com/hekmon/go-qbittorrent-webapi.svg)](https://pkg.go.dev/github.com/hekmon/go-qbittorrent-webapi)

Golang bindings for [qBittorrent v5 Web API](https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)).

> **A note on how this library was built**: development happened in three phases. First, the core architecture and low-level HTTP plumbing (`client.go`, `request.go`) were hand-crafted around a small set of initial endpoints to establish patterns for error handling, custom JSON marshaling, and data modeling. Second, **HITL** (human in the loop) introduced agentic engineering: the validated patterns were codified into [`AGENTS.md`](AGENTS.md), and the first agent-driven endpoints were implemented under close supervision to ensure the rules worked in practice. Third, **HOTL** (human on the loop) took over as agents filled in the remaining endpoints autonomously, with human validation and adjustments to `AGENTS.md` when edge cases surfaced.

## Installation

```bash
go get -u github.com/hekmon/go-qbittorrent-webapi
```

## Usage

```go
package main

import (
	"context"
	"fmt"
	"net/url"
	"time"

	qbtapi "github.com/hekmon/go-qbittorrent-webapi"
)

func main() {
	// Parse endpoint
	endpoint, err := url.Parse("http://127.0.0.1:8080")
	if err != nil {
		panic(err)
	}
	client, err := qbtapi.New(endpoint, "admin", "password")
	if err != nil {
		panic(err)
	}
	// Login (manual login is not needed as client will automatically try to login if not logged)
	if err = client.Login(context.TODO()); err != nil {
		panic(err)
	}
	defer func() {
        // But it is recommended that you logout once done to clear session on server side
		if err = client.Logout(context.TODO()); err != nil {
			panic(err)
		}
	}()

	// Get versions
	appVersion, err := client.GetApplicationVersion(context.TODO())
	if err != nil {
		panic(err)
	}
	fmt.Printf("qBittorrent application version: %s\n", appVersion)
	apiVersion, err := client.GetAPIVersion(context.TODO())
	if err != nil {
		panic(err)
	}
	fmt.Printf("qBittorrent API version: %s\n", apiVersion)
	buildInfos, err := client.GetBuildInfo(context.TODO())
	if err != nil {
		panic(err)
	}
	fmt.Printf("qBittorrent build info: %+v\n", buildInfos)

    // App prefs
	appPrefs, err := client.GetApplicationPreferences(context.TODO())
	if err != nil {
		panic(err)
	}
	fmt.Printf("qBittorrent application preferences:\n%s\n\n", appPrefs)
	prefs := qbtapi.ApplicationPreferences{
		AlternativeWebuiEnabled: qbtapi.Bool(true),
	}
	fmt.Println(prefs)
	err = client.SetApplicationPreferences(context.TODO(), prefs)
	if err != nil {
		panic(err)
	}
	fmt.Println("Prefs set.")
	appPrefs, err = client.GetApplicationPreferences(context.TODO())
	if err != nil {
		panic(err)
	}
	fmt.Printf("qBittorrent new application preferences:\n%s\n\n", appPrefs)

	// Torrents listing
	torrents, err := client.GetTorrentList(context.TODO(), &qbtapi.ListFilters{
		State: qbtapi.FilterStateActive.Ptr(),
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("%d active torrents:\n", len(torrents))
	for _, torrent := range torrents {
		fmt.Printf("\t * %+v\n", torrent)
	}

    // Add torrents
	files, err := qbtapi.ReadTorrentsFiles([]string{"/mnt/d/Downloads/ubuntu-25.04-desktop-amd64.iso.torrent"})
	if err != nil {
		panic(err)
	}
	trURL, err := url.Parse("https://releases.ubuntu.com/25.04/ubuntu-25.04-live-server-amd64.iso.torrent")
	if err != nil {
		panic(err)
	}
	err = client.AddNewTorrents(context.TODO(), files, []*url.URL{trURL}, &qbtapi.AddNewTorrentsOptions{
		Paused: qbtapi.Bool(true),
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("torrents added")
	time.Sleep(3 * time.Second)
	
	// List all torrents for deletion
	torrents, err := client.GetTorrentList(context.TODO(), nil)
	if err != nil {
		panic(err)
	}
	hashes := make([]string, len(torrents))
	for index, torrent := range torrents {
		hashes[index] = torrent.Hash
	}
	if err = client.DeleteTorrents(context.TODO(), hashes, true); err != nil {
		panic(err)
	}
	fmt.Printf("Deleted %d torrents\n", len(hashes))

	// etc...
}
```

## Endpoints implementation

All documented endpoints of the qBittorrent v5 Web API are implemented.

### Authentication

- [x] Login
- [x] Logout

### Application

- [x] Get application version
- [x] Get API version
- [x] Get build info
- [x] Shutdown application
- [x] Get application preferences
- [x] Set application preferences
- [x] Get default save path
- [x] Get cookies
- [x] Set cookies

### Log

- [x] Get log
- [x] Get peer log

### Sync

- [x] Get main data
- [x] Get torrent peers data

### Transfer info

- [x] Get global transfer info
- [x] Get alternative speed limits state
- [x] Toggle alternative speed limits
- [x] Get global download limit
- [x] Set global download limit
- [x] Get global upload limit
- [x] Set global upload limit
- [x] Ban peers

### Torrent management

- [x] Get torrent list
- [x] Get torrent generic properties
- [x] Get torrent trackers
- [x] Get torrent web seeds
- [x] Get torrent contents
- [x] Get torrent pieces' states
- [x] Get torrent pieces' hashes
- [x] Pause torrents
- [x] Resume torrents
- [x] Delete torrents
- [x] Recheck torrents
- [x] Reannounce torrents
- [x] Edit trackers
- [x] Remove trackers
- [x] Add peers
- [x] Add new torrent
- [x] Add trackers to torrent
- [x] Increase torrent priority
- [x] Decrease torrent priority
- [x] Maximal torrent priority
- [x] Minimal torrent priority
- [x] Set file priority
- [x] Get torrent download limit
- [x] Set torrent download limit
- [x] Set torrent share limit
- [x] Get torrent upload limit
- [x] Set torrent upload limit
- [x] Set torrent location
- [x] Set torrent name
- [x] Set torrent category
- [x] Get all categories
- [x] Add new category
- [x] Edit category
- [x] Remove categories
- [x] Add torrent tags
- [x] Remove torrent tags
- [x] Get all tags
- [x] Create tags
- [x] Delete tags
- [x] Set automatic torrent management
- [x] Toggle sequential download
- [x] Set first/last piece priority
- [x] Set force start
- [x] Set super seeding
- [x] Rename file
- [x] Rename folder

### RSS (experimental)

- [x] Add folder
- [x] Add feed
- [x] Remove item
- [x] Move item
- [x] Get all items
- [x] Mark as read
- [x] Refresh item
- [x] Set auto-downloading rule
- [x] Rename auto-downloading rule
- [x] Remove auto-downloading rule
- [x] Get all auto-downloading rules
- [x] Get all articles matching a rule

### Search

- [x] Start search
- [x] Stop search
- [x] Get search status
- [x] Get search results
- [x] Delete search
- [x] Get search plugins
- [x] Install search plugin
- [x] Uninstall search plugin
- [x] Enable search plugin
- [x] Update search plugins
