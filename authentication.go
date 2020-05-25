package qbtapi

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

/*
	Authentication
	https://github.com/qbittorrent/qBittorrent/wiki/Web-API-Documentation#authentication
*/

const (
	authenticationPrefix = "auth"
)

// Login performs a login against the remote qBittorrent server.
// If successfull a cookie will be set in the http client and will be used for any other methods.
// Note that you do not need to call login yourself as it is called automatically if necessary.
func (c *Controller) Login(ctx context.Context) (err error) {
	// build URL
	authURL := *c.url
	authURL.Path = fmt.Sprintf("%s/%s/%s/%s", authURL.Path, apiPrefix, authenticationPrefix, "login")
	// build payload
	payload := url.Values{}
	payload.Set("username", c.user)
	payload.Set("password", c.password)
	payloadSerialized := payload.Encode()
	// build request
	request, err := http.NewRequest("POST", authURL.String(), strings.NewReader(payloadSerialized))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Content-Length", strconv.Itoa(len(payloadSerialized)))
	referer := fmt.Sprintf("%s://%s", authURL.Scheme, authURL.Hostname())
	if authURL.Port() != "" {
		referer += ":" + authURL.Port()
	}
	request.Header.Set("Referer", referer)
	if ctx != nil {
		request = request.WithContext(ctx)
	}
	// execute request
	response, err := c.client.Do(request)
	if err != nil {
		return
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		err = HTTPError(response.StatusCode)
	}
	return
}
