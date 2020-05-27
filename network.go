package qbtapi

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

const (
	apiPrefix = "api/v2"
)

type request struct {
	request *http.Request
	payload *strings.Reader
}

func (c *Controller) requestBuild(ctx context.Context, method, APIName, APIMethodName string, input map[string]string) (req request, err error) {
	// build URL
	requestURL := *c.url
	requestURL.Path = fmt.Sprintf("%s/%s/%s/%s", requestURL.Path, apiPrefix, APIName, APIMethodName)
	// build payload
	var reqPayloadSize int
	if method == "POST" && input != nil {
		payloadValues := url.Values{}
		for key, value := range input {
			payloadValues.Set(key, value)
		}
		payloadSerialized := payloadValues.Encode()
		reqPayloadSize = len(payloadSerialized)
		req.payload = strings.NewReader(payloadSerialized)
	}
	// build http request
	if ctx == nil {
		ctx = context.Background()
	}
	if req.request, err = http.NewRequestWithContext(ctx, method, requestURL.String(), req.payload); err != nil {
		return
	}
	if req.payload != nil {
		req.request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.request.Header.Set("Content-Length", strconv.Itoa(reqPayloadSize))
	}
	return
}

func (c *Controller) requestExecute(ctx context.Context, req request, output interface{}, autoAuth bool) (err error) {
	// execute request
	response, err := c.client.Do(req.request)
	if err != nil {
		err = fmt.Errorf("HTTP request failure: %w", err)
		return
	}
	defer response.Body.Close()
	switch response.StatusCode {
	case http.StatusOK:
		// proceed
	case http.StatusForbidden:
		if !autoAuth {
			err = HTTPError(response.StatusCode)
			return
		}
		response.Body.Close() // don't leave it hanging, early close
		// try to login
		if err = c.Login(ctx); err != nil {
			err = fmt.Errorf("auto login failed: %w", err)
			return
		}
		// reset payload reader & reissue request now that we are auth
		if _, err = req.payload.Seek(0, io.SeekStart); err != nil {
			err = fmt.Errorf("auto login succeeded but reseting original request payload failed: %w", err)
			return
		}
		return c.requestExecute(ctx, req, output, false)
	default:
		err = HTTPError(response.StatusCode)
		return
	}
	// handle body if needed
	if output == nil {
		return
	}
	switch typedOutput := output.(type) {
	case *string:
		var bodyData []byte
		if bodyData, err = ioutil.ReadAll(response.Body); err != nil {
			err = fmt.Errorf("reading answer body failed: %w", err)
			return
		}
		*typedOutput = string(bodyData)
	default:
		err = fmt.Errorf("request succeeded but output type is not supported: %v", reflect.TypeOf(output))
	}
	return
}

// HTTPError contains a HTTP status code which was not acceptable
type HTTPError int

func (he HTTPError) Error() string {
	return fmt.Sprintf("%d %s", int(he), http.StatusText(int(he)))
}
