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
// Use opts to customize the client behavior, e.g. WithHTTPClient or WithUserAgent.
func New(apiEndpoint *url.URL, user, password string, opts ...ClientOption) (c *Client, err error) {
	// handle url
	if apiEndpoint == nil {
		err = errors.New("apiEndpoint can't be nil")
		return
	}
	copiedURL, err := url.Parse(apiEndpoint.String())
	if err != nil {
		err = fmt.Errorf("parsing API endpoint URL failed: %w", err)
		return
	}
	// spawn the client with defaults
	c = &Client{
		user:      user,
		password:  password,
		url:       copiedURL,
		client:    cleanhttp.DefaultPooledClient(),
		userAgent: userAgentValue,
	}
	// apply options
	for _, opt := range opts {
		opt(c)
	}
	// create the cookie jar if needed
	if c.client.Jar == nil {
		if c.client.Jar, err = cookiejar.New(&cookiejar.Options{
			PublicSuffixList: publicsuffix.List,
		}); err != nil {
			err = fmt.Errorf("creating cookie jar failed: %w", err)
			return
		}
	}
	return
}

// ClientOption configures a Client.
type ClientOption func(*Client)

// WithHTTPClient replaces the default HTTP client with a custom one.
// If the provided client does not have a cookie jar, one will be created automatically.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.client = httpClient
	}
}

// WithUserAgent sets a custom User-Agent header for all requests.
// If not used, the default "github.com/hekmon/go-qbittorrent-webapi" is sent.
func WithUserAgent(userAgent string) ClientOption {
	return func(c *Client) {
		c.userAgent = userAgent
	}
}

// Client is a statefull object allowing to interface the qBittorrent Web API on a particular endpoint.
// Must be instanciated with New().
type Client struct {
	user      string
	password  string
	url       *url.URL
	client    *http.Client
	userAgent string
}
