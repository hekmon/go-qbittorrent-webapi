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

// GetApplicationVersionCtx returns the application version. Ctx can be nil.
// https://github.com/qbittorrent/qBittorrent/wiki/Web-API-Documentation#get-application-version
func (c *Controller) GetApplicationVersionCtx(ctx context.Context) (version string, err error) {
	err = c.request(ctx, "GET", applicationPrefix, "version", &version)
	return
}
