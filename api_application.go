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
