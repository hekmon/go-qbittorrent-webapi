package qbtapi

import (
	"context"
	"fmt"
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

func (c *Controller) requestBuild(ctx context.Context, method, APIName, APIMethodName string, input map[string]string) (request *http.Request, err error) {
	// build URL
	requestURL := *c.url
	requestURL.Path = fmt.Sprintf("%s/%s/%s/%s", requestURL.Path, apiPrefix, APIName, APIMethodName)
	// build payload
	var reqPayload string
	if method == "POST" && input != nil {
		payloadValues := make(url.Values, len(input))
		for key, value := range input {
			payloadValues.Set(key, value)
		}
		reqPayload = payloadValues.Encode()
	}
	// build http request
	if ctx == nil {
		ctx = context.Background()
	}
	if request, err = http.NewRequestWithContext(ctx, method, requestURL.String(), strings.NewReader(reqPayload)); err != nil {
		return
	}
	if reqPayload != "" {
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		request.Header.Set("Content-Length", strconv.Itoa(len(reqPayload)))
	}
	return
}

func (c *Controller) requestExecute(ctx context.Context, request *http.Request, output interface{}, autoAuth bool) (err error) {
	// execute request
	response, err := c.client.Do(request)
	if err != nil {
		err = fmt.Errorf("HTTP request failure: %w", err)
		return
	}
	defer response.Body.Close()
	switch response.StatusCode {
	case http.StatusOK:
		// proceed
	case http.StatusForbidden:
		// is this iteration allow to auto login ?
		if !autoAuth {
			err = HTTPError(response.StatusCode)
			return
		}
		// try to login
		response.Body.Close() // don't leave it hanging, early close
		if err = c.Login(ctx); err != nil {
			err = fmt.Errorf("auto login failed: %w", err)
			return
		}
		fmt.Println("autologged!")
		// reset payload reader & reissue request now that we are auth
		if request.Body, err = request.GetBody(); err != nil {
			err = fmt.Errorf("can't reset body of original query after successfull autologin: %w", err)
			return
		}
		return c.requestExecute(ctx, request, output, false)
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
