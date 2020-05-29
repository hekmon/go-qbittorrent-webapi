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
	APIReferenceVersion = "2.5.1"
)

// New return a initialized and ready to use Controller.
// customHTTPClient can be nil
func New(apiEndpoint *url.URL, user, password string, customHTTPClient *http.Client) (c *Controller, err error) {
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
		customHTTPClient.Jar, err = cookiejar.New(&cookiejar.Options{
			PublicSuffixList: publicsuffix.List,
		})
		if err != nil {
			return
		}
	}
	// spawn the controller
	c = &Controller{
		user:     user,
		password: password,
		url:      copiedURL,
		client:   customHTTPClient,
	}
	return
}

// Controller is a statefull object allowing to interface the qBittorrent Web API on a particular endpoint.
// Must be instanciated with New().
type Controller struct {
	user     string
	password string
	url      *url.URL
	client   *http.Client
}
