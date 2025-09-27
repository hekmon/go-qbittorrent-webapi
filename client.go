package qbtapi

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"github.com/hashicorp/go-cleanhttp"
	"golang.org/x/net/publicsuffix"
)

const (
	// APIReferenceVersion contains the version of the API this libs is built against
	APIReferenceVersion = "2.11.3"
)

// New return a initialized and ready to use Client.
// customHTTPClient can be nil
func New(apiEndpoint *url.URL, user, password string, customHTTPClient *http.Client) (c *Client, err error) {
	// handle url
	if apiEndpoint == nil {
		err = errors.New("apiEndpoint can't be nil")
		return
	}
	copiedURL, err := url.Parse(apiEndpoint.String())
	if err != nil {
		err = fmt.Errorf("apiEndpoint can't be (re)parsed as URL: %v", err) // weird
		return
	}
	// handle http client
	if customHTTPClient == nil {
		customHTTPClient = cleanhttp.DefaultPooledClient()
	}
	// create the cookie jar if needed
	if customHTTPClient.Jar == nil {
		if customHTTPClient.Jar, err = cookiejar.New(&cookiejar.Options{
			PublicSuffixList: publicsuffix.List,
		}); err != nil {
			return
		}
	}
	// spawn the client
	c = &Client{
		user:     user,
		password: password,
		url:      copiedURL,
		client:   customHTTPClient,
	}
	return
}

// Client is a statefull object allowing to interface the qBittorrent Web API on a particular endpoint.
// Must be instanciated with New().
type Client struct {
	user     string
	password string
	url      *url.URL
	client   *http.Client
}

// String returns a pointer to the string value passed in.
// Useful for the many *string fields in the API model.
func String(value string) *string {
	return &value
}

// Int returns a pointer to the int value passed in.
// Useful for the many *int fields in the API model.
func Int(value int) *int {
	return &value
}

// Bool returns a pointer to the bool value passed in.
// Useful for the many *bool fields in the API model.
func Bool(value bool) *bool {
	return &value
}
