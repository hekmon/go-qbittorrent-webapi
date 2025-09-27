package qbtapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"strconv"
	"strings"
)

const (
	apiPrefix                     = "api/v2"
	originHeader                  = "Origin"
	contentLenHeader              = "Content-Length"
	contentTypeHeader             = "Content-Type"
	contentTypeHeaderFormURL      = "application/x-www-form-urlencoded"
	contentTypeHeaderTextPlain    = "text/plain"
	contentTypeHeaderTextPlanUTF8 = "text/plain; charset=UTF-8"
	contentTypeHeaderJSON         = "application/json"
)

func (c *Client) requestBuild(ctx context.Context, method, APIName, APIMethodName string, input map[string]string) (request *http.Request, err error) {
	// build URL
	requestURL := *c.url
	requestURL.Path = path.Join(requestURL.Path, apiPrefix, APIName, APIMethodName)
	// build payload
	var (
		body       io.Reader
		reqPayload string
	)
	if method == "POST" && input != nil {
		switch {
		case len(input) == 1 && input[""] != "":
			// weird qbittorrent implementation: we need to put the json data without encoding it (set cookies ?)
			reqPayload = input[""]
			fmt.Println(reqPayload)
		case len(input) == 1 && input["json"] != "":
			// weird qbittorrent implementation: we need to put the json data without encoding it (set app prefs)
			reqPayload = "json=" + input["json"]
			fmt.Println(reqPayload)
		default:
			// regulard url encoded values
			payloadValues := make(url.Values, len(input))
			for key, value := range input {
				payloadValues.Set(key, value)
			}
			reqPayload = payloadValues.Encode()
		}
		body = strings.NewReader(reqPayload)
	}
	// build http request
	if request, err = http.NewRequestWithContext(ctx, method, requestURL.String(), body); err != nil {
		return
	}
	if body != nil {
		request.Header.Set(contentTypeHeader, contentTypeHeaderFormURL)
		request.Header.Set(contentLenHeader, strconv.Itoa(len(reqPayload)))
	}
	return
}

func (c *Client) requestExecute(ctx context.Context, request *http.Request, output any, autoAuth bool) (err error) {
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
	// handle body
	return c.requestExtract(response, output)
}

func (c *Client) requestExtract(response *http.Response, output any) (err error) {
	// Pre checks
	if output == nil {
		// caller does not care about body
		return
	}
	if reflect.TypeOf(output).Kind() != reflect.Ptr {
		return InternalError(fmt.Sprintf("output must be a pointer (currentlyu: %v)",
			reflect.TypeOf(output)))
	}
	// Given the response body content type
	switch response.Header.Get(contentTypeHeader) {
	// text-plain
	case contentTypeHeaderTextPlain, contentTypeHeaderTextPlanUTF8:
		// output must be a string pointer
		if reflect.Indirect(reflect.ValueOf(output)).Kind() != reflect.String {
			return InternalError(fmt.Sprintf("output should be a string pointer when %s is '%s' (currently: %v)",
				contentTypeHeader, contentTypeHeaderTextPlain, reflect.TypeOf(output)))
		}
		// extract it
		var bodyData []byte
		if bodyData, err = io.ReadAll(response.Body); err != nil {
			err = fmt.Errorf("reading answer body failed: %w", err)
			return
		}
		*output.(*string) = string(bodyData)
	// application/json
	case contentTypeHeaderJSON:
		// output must be a struct or a slice pointer
		switch reflect.Indirect(reflect.ValueOf(output)).Kind() {
		case reflect.Struct, reflect.Slice:
			// ok
		default:
			return InternalError(fmt.Sprintf("when %s is '%s' output should be a struct pointer or a slice pointer (currently: %v)",
				contentTypeHeader, contentTypeHeaderJSON, reflect.TypeOf(output)))
		}
		// decode as JSON
		if err = json.NewDecoder(response.Body).Decode(output); err != nil {
			return fmt.Errorf("decody response body as JSON failed: %w", err)
		}
	default:
		return InternalError(fmt.Sprintf("%s value '%s' is not supported",
			contentTypeHeader, response.Header.Get(contentTypeHeader)))
	}
	return
}

// HTTPError contains a HTTP status code which was not acceptable
type HTTPError int

func (he HTTPError) Error() string {
	return fmt.Sprintf("%d %s", int(he), http.StatusText(int(he)))
}

// InternalError is an error that should not happen: if you encounter one, please open a bug report !
type InternalError string

func (ie InternalError) Error() string {
	return fmt.Sprintf("internal error: %s", string(ie))
}
