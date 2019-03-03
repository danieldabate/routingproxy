package routingproxy

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
)

// RequestModifier defines a request and response modifying functions
// and the regex for the paths for which it should be applied
type RequestModifier struct {
	MatchingPath string

	DisableEncoding bool

	RequestModifier      func(*http.Request)
	ResponseModifier     func(*http.Response) error
	ResponseBodyModifier func(*http.Response, []byte) []byte

	pathRegex *regexp.Regexp
}

// MatchesPath evaluates a given path against the MatchingPath
func (rm *RequestModifier) matchesPath(p string) bool {
	return rm.pathRegex.MatchString(p)
}

// ModifyRequest modifies a request if the path matches MatchingPath
func (rm *RequestModifier) modifyRequest(r *http.Request) {
	if rm.RequestModifier != nil && rm.matchesPath(r.URL.Path) {
		rm.RequestModifier(r)
	}

	if rm.DisableEncoding {
		r.Header.Set("Accept-Encoding", "")
	}
}

// ModifyResponse modifies a request if the path matches MatchingPath
func (rm *RequestModifier) modifyResponse(r *http.Response) error {

	if rm.matchesPath(r.Request.URL.Path) {
		if rm.ResponseModifier != nil {
			if err := rm.ResponseModifier(r); err != nil {
				return err
			}
		}

		if rm.ResponseBodyModifier != nil {
			bodyBytes, err := ioutil.ReadAll(r.Body)
			if err != nil {
				return err
			}

			err = r.Body.Close()
			if err != nil {
				return err
			}

			bodyBytes = rm.ResponseBodyModifier(r, bodyBytes)

			r.Body = ioutil.NopCloser(bytes.NewReader(bodyBytes))

			bodyLength := len(bodyBytes)
			r.ContentLength = int64(bodyLength)
			r.Header.Set("Content-Length", strconv.Itoa(bodyLength))
		}
	}

	return nil
}
