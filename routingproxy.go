package routingproxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
)

// RoutingProxy is an HTTP Handler that uses a ReverseProxy and allows
// to modify the requests and answers based on the routers
type RoutingProxy struct {
	Proxy *httputil.ReverseProxy

	requestModifiers []RequestModifier
}

// AddRequestModifier adds a RequestModifier function to the RoutingProxy
func (rp *RoutingProxy) AddRequestModifier(rm RequestModifier) error {
	regex, err := regexp.Compile(rm.MatchingPath)

	if err == nil {
		rm.pathRegex = regex
		rp.requestModifiers = append(rp.requestModifiers, rm)
	}

	return err
}

// NewRoutingProxy returns a new RoutingProxy with an embedded httputil.ReverseProxy
func NewRoutingProxy(target *url.URL) *RoutingProxy {
	targetQuery := target.RawQuery

	routingProxy := RoutingProxy{}

	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}

		for _, rm := range routingProxy.requestModifiers {
			rm.modifyRequest(req)
		}
	}

	responseModifier := func(resp *http.Response) (err error) {
		for _, rm := range routingProxy.requestModifiers {
			if err := rm.modifyResponse(resp); err != nil {
				return err
			}
		}

		return nil
	}

	routingProxy.Proxy = &httputil.ReverseProxy{Director: director}
	routingProxy.Proxy.ModifyResponse = responseModifier

	return &routingProxy
}
