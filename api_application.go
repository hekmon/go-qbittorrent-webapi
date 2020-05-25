package qbtapi

import (
	"context"
)

/*
	Application
	https://github.com/qbittorrent/qBittorrent/wiki/Web-API-Documentation#application
*/

const (
	applicationPrefix = "app"
)

// GetApplicationVersion returns the application version. Ctx can be nil.
// https://github.com/qbittorrent/qBittorrent/wiki/Web-API-Documentation#get-application-version
func (c *Controller) GetApplicationVersion(ctx context.Context) (version string, err error) {
	err = c.requestAutoLogin(ctx, "GET", applicationPrefix, "version", &version)
	return
}
