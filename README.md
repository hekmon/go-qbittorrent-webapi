# qBittorrent Web API

[![Go Reference](https://pkg.go.dev/badge/github.com/hekmon/go-qbittorrent-webapi.svg)](https://pkg.go.dev/github.com/hekmon/go-qbittorrent-webapi)

Golang bindings for [qBittorrent v5 Web API](https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)).

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
	client, err := qbtapi.New(endpoint, "admin", "password", nil)
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

	// Default save path
	path, err := client.GetDefaultSavePath(context.TODO())
	if err != nil {
		panic(err)
	}
	fmt.Printf("Default save path: %s\n", path)

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

    // etc...
}
```

## Endpoints implementation

Not all endpoints are implemented, only the ones I needed for my own usage.
But global software architecture is ready and adding more endpoints should be easy.
If you need more, feel free to open an issue or a PR.

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
- [x] Set cookies (implemented but not working, expected payload on server side is unclear)

### Log

- [ ] Get log
- [ ] Get peer log

### Sync

- [ ] Get main data
- [ ] Get torrent peers data

### Transfer info

- [ ] Get global transfer info
- [ ] Get alternative speed limits state
- [ ] Toggle alternative speed limits
- [ ] Get global download limit
- [ ] Set global download limit
- [ ] Get global upload limit
- [ ] Set global upload limit
- [ ] Ban peers

### Torrent management

- [x] Get torrent list
- [x] Get torrent generic properties
- [x] Get torrent trackers
- [ ] Get torrent web seeds
- [ ] Get torrent contents
- [ ] Get torrent pieces' states
- [ ] Get torrent pieces' hashes
- [ ] Pause torrents
- [ ] Resume torrents
- [ ] Delete torrents
- [ ] Recheck torrents
- [ ] Reannounce torrents
- [ ] Edit trackers
- [ ] Remove trackers
- [ ] Add peers
- [ ] Add new torrent
- [ ] Add trackers to torrent
- [ ] Increase torrent priority
- [ ] Decrease torrent priority
- [ ] Maximal torrent priority
- [ ] Minimal torrent priority
- [ ] Set file priority
- [ ] Get torrent download limit
- [ ] Set torrent download limit
- [ ] Set torrent share limit
- [ ] Get torrent upload limit
- [ ] Set torrent upload limit
- [ ] Set torrent location
- [ ] Set torrent name
- [ ] Set torrent category
- [ ] Get all categories
- [ ] Add new category
- [ ] Edit category
- [ ] Remove categories
- [ ] Add torrent tags
- [ ] Remove torrent tags
- [ ] Get all tags
- [ ] Create tags
- [ ] Delete tags
- [ ] Set automatic torrent management
- [ ] Toggle sequential download
- [ ] Set first/last piece priority
- [ ] Set force start
- [ ] Set super seeding
- [ ] Rename file
- [ ] Rename folder

### RSS (experimental)

- [ ] Add folder
- [ ] Add feed
- [ ] Remove item
- [ ] Move item
- [ ] Get all items
- [ ] Mark as read
- [ ] Refresh item
- [ ] Set auto-downloading rule
- [ ] Rename auto-downloading rule
- [ ] Remove auto-downloading rule
- [ ] Get all auto-downloading rules
- [ ] Get all articles matching a rule

### Search

- [ ] Start search
- [ ] Stop search
- [ ] Get search status
- [ ] Get search results
- [ ] Delete search
- [ ] Get search plugins
- [ ] Install search plugin
- [ ] Uninstall search plugin
- [ ] Enable search plugin
- [ ] Update search plugins
