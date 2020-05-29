package qbtapi

import (
	"context"
	"fmt"
)

/*
	Application
	https://github.com/qbittorrent/qBittorrent/wiki/Web-API-Documentation#application
*/

const (
	applicationAPIName = "app"
)

// GetApplicationVersion returns the application version.
// https://github.com/qbittorrent/qBittorrent/wiki/Web-API-Documentation#get-application-version
func (c *Controller) GetApplicationVersion(ctx context.Context) (version string, err error) {
	req, err := c.requestBuild(ctx, "GET", applicationAPIName, "version", nil)
	if err != nil {
		err = fmt.Errorf("building request failed: %w", err)
		return
	}
	if err = c.requestExecute(ctx, req, &version, true); err != nil {
		err = fmt.Errorf("executing request failed: %w", err)
	}
	return
}

// GetAPIVersion returns the WebAPI version.
// https://github.com/qbittorrent/qBittorrent/wiki/Web-API-Documentation#get-api-version
func (c *Controller) GetAPIVersion(ctx context.Context) (version string, err error) {
	req, err := c.requestBuild(ctx, "GET", applicationAPIName, "webapiVersion", nil)
	if err != nil {
		err = fmt.Errorf("building request failed: %w", err)
		return
	}
	if err = c.requestExecute(ctx, req, &version, true); err != nil {
		err = fmt.Errorf("executing request failed: %w", err)
	}
	return
}

// BuildInfo contains all the compilation informations of a given remote server
type BuildInfo struct {
	QT         string `json:"qt"`         // QT version
	LibTorrent string `json:"libtorrent"` // libtorrent version
	Boost      string `json:"boost"`      // Boost version
	OpenSSL    string `json:"openssl"`    // OpenSSL version
	Bitness    int    `json:"bitness"`    // Application bitness (e.g. 64-bit)
}

// GetBuildInfo returns compilation informations of target server.
// https://github.com/qbittorrent/qBittorrent/wiki/Web-API-Documentation#get-build-info
func (c *Controller) GetBuildInfo(ctx context.Context) (infos BuildInfo, err error) {
	req, err := c.requestBuild(ctx, "GET", applicationAPIName, "buildInfo", nil)
	if err != nil {
		err = fmt.Errorf("building request failed: %w", err)
		return
	}
	if err = c.requestExecute(ctx, req, &infos, true); err != nil {
		err = fmt.Errorf("executing request failed: %w", err)
	}
	return
}
