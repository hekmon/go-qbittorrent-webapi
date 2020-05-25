package qbtapi

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
)

const (
	apiPrefix = "api/v2"
)

func (c *Controller) requestAutoLogin(ctx context.Context, method, APIName, APIMethodName string, output interface{}) (err error) {
	return c.request(ctx, method, APIName, APIMethodName, output, false)
}

func (c *Controller) request(ctx context.Context, method, APIName, APIMethodName string, output interface{}, lastTry bool) (err error) {
	// build URL
	requestURL := *c.url
	requestURL.Path = fmt.Sprintf("%s/%s/%s/%s", requestURL.Path, apiPrefix, APIName, APIMethodName)
	// build request
	request, err := http.NewRequest(method, requestURL.String(), nil)
	if err != nil {
		return
	}
	if ctx != nil {
		request = request.WithContext(ctx)
	}
	// execute request
	response, err := c.client.Do(request)
	if err != nil {
		return
	}
	defer response.Body.Close()
	switch response.StatusCode {
	case http.StatusOK:
		// proceed
	case http.StatusForbidden:
		if lastTry {
			err = HTTPError(response.StatusCode)
			return
		}
		response.Body.Close() // don't leave it hanging, early close
		// try to login
		/// TODO
		// re issue request now that we are authed
		return c.request(ctx, method, APIName, APIMethodName, output, true)
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
			return
		}
		*typedOutput = string(bodyData)
	default:
		err = fmt.Errorf("output type is not supported: %v", reflect.TypeOf(output))
	}
	return
}

// HTTPError contains a HTTP status code which was not acceptable
type HTTPError int

func (he HTTPError) Error() string {
	return fmt.Sprintf("%d %s", int(he), http.StatusText(int(he)))
}
