package qbtapi

import (
	"context"
	"fmt"
)

/*
	Authentication
	https://github.com/qbittorrent/qBittorrent/wiki/Web-API-Documentation#authentication
*/

const (
	authenticationAPIName = "auth"
)

// Login performs a login against the remote qBittorrent server.
// If successfull, a cookie will be set within the http client and will be used for any further methods calls.
// Note that you do not need to call login yourself as it is called automatically if necessary.
// https://github.com/qbittorrent/qBittorrent/wiki/Web-API-Documentation#login
func (c *Client) Login(ctx context.Context) (err error) {
	// Build request
	req, err := c.requestBuild(ctx, "POST", authenticationAPIName, "login", map[string]string{
		"username": c.user,
		"password": c.password,
	})
	if err != nil {
		return fmt.Errorf("building request failed: %w", err)
	}
	// Add custom header for login
	origin := fmt.Sprintf("%s://%s", c.url.Scheme, c.url.Hostname())
	if c.url.Port() != "" {
		origin += ":" + c.url.Port()
	}
	req.Header.Set(originHeader, origin)
	// execute auth request
	if err = c.requestExecute(ctx, req, nil, false); err != nil {
		err = fmt.Errorf("executing request failed: %w", err)
	}
	return
}

// Logout performs a clean logout against the server, effectively cleaning upstream (and local) cookie.
// Recommended to call before exiting to leave a clean server state.
// https://github.com/qbittorrent/qBittorrent/wiki/Web-API-Documentation#logout
func (c *Client) Logout(ctx context.Context) (err error) {
	// Build request
	req, err := c.requestBuild(ctx, "GET", authenticationAPIName, "logout", nil)
	if err != nil {
		return fmt.Errorf("request building failure: %w", err)
	}
	// execute auth request
	if err = c.requestExecute(ctx, req, nil, false); err != nil {
		err = fmt.Errorf("executing request failed: %w", err)
	}
	return
}
