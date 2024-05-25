package client

import (
	"bytes"
	"encoding/json"
	be "errors"
	"fmt"
	"github.com/pkg/errors"
	"github.com/vinr-eu/go-framework/log"
	"io"
	"net/http"
	"strings"
	"time"
)

var ErrDataNotFound = be.New("data not found")

func Get(url string, response interface{}, headers ...string) error {
	return doHTTP("GET", url, nil, response, headers...)
}

func Post(url string, request interface{}, response interface{}, headers ...string) error {
	return doHTTP("POST", url, request, response, headers...)
}

func doHTTP(method string, url string, request interface{}, response interface{}, headers ...string) error {
	// Should be parameterized some day... over the rainbow.
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	var req *http.Request
	contentType := "application/json"

	if request != nil {
		body := &bytes.Buffer{}
		var requestBody []byte
		// Marshal the object into JSON.
		var err error
		requestBody, err = json.Marshal(request)
		if err != nil {
			return errors.WithStack(err)
		}
		body = bytes.NewBuffer(requestBody)
		req, err = http.NewRequest(method, url, body)
		if err != nil {
			return errors.WithStack(err)
		}
	} else {
		// Make the GET request.
		var err error
		req, err = http.NewRequest(method, url, nil)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	// Set the headers.
	for i := 0; i < len(headers); i += 2 {
		key := headers[i]
		value := headers[i+1]
		req.Header.Set(key, value)
	}
	if request != nil {
		req.Header.Set("Content-Type", contentType)
	}

	// Hit it.
	resp, err := client.Do(req)
	if err != nil {
		logger := log.NewLogger()
		logger.Error("client request failed", "err", err, "resp", resp)
		return errors.WithStack(err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	// Check the response status.
	if resp.StatusCode == http.StatusNotFound {
		return errors.WithStack(ErrDataNotFound)
	} else if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		var bodyBytes []byte
		if resp.ContentLength != 0 {
			bodyBytes, err = io.ReadAll(resp.Body)
			if err != nil {
				return errors.WithStack(err)
			}
		}
		return errors.WithStack(fmt.Errorf("call response failed with status= %v body= %v",
			resp.StatusCode, string(bodyBytes)))
	}

	// Parse the JSON response.
	if resp.ContentLength != 0 {
		if strings.HasPrefix(resp.Header.Get("Content-Type"), "text/plain") {
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return errors.WithStack(err)
			}
			switch r := response.(type) {
			case *string:
				*r = string(b)
			}
		} else {
			err = json.NewDecoder(resp.Body).Decode(response)
			if err != nil {
				return errors.WithStack(err)
			}
		}
	}

	return nil
}
