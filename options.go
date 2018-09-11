package req

import (
	"net"
	"net/http"
	"time"
)

var defaultOptions = options{
	transport: &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	},
}

type options struct {
	transport     *http.Transport
	cookieJar     http.CookieJar
	checkRedirect func(req *http.Request, via []*http.Request) error
	timeout       time.Duration
	baseURL       string
	header        http.Header
}

// Option parameter options
type Option func(*options)

// SetBaseURL set the requested base url
func SetBaseURL(base string) Option {
	return func(o *options) {
		o.baseURL = base
	}
}

// SetBaseHeader set the requested base header
func SetBaseHeader(key, value string) Option {
	return func(o *options) {
		if o.header == nil {
			o.header = make(http.Header)
		}
		o.header.Set(key, value)
	}
}

// SetTransport specifies the mechanism by which individual
// HTTP requests are made.
// If nil, DefaultTransport is used.
func SetTransport(tr *http.Transport) Option {
	return func(o *options) {
		o.transport = tr
	}
}

// SetCookieJar specifies the cookie jar
func SetCookieJar(jar http.CookieJar) Option {
	return func(o *options) {
		o.cookieJar = jar
	}
}

// SetCheckRedirect specifies the policy for handling redirects
func SetCheckRedirect(redirect func(req *http.Request, via []*http.Request) error) Option {
	return func(o *options) {
		o.checkRedirect = redirect
	}
}

// SetTimeout specifies a time limit for requests made by this Client
func SetTimeout(d time.Duration) Option {
	return func(o *options) {
		o.timeout = d
	}
}

type requestOptions struct {
	request *http.Request
	handle  func(req *http.Request) (*http.Request, error)
}

// RequestOption request parameter options
type RequestOption func(*requestOptions)

// SetHeader set the request header
func SetHeader(key, value string) RequestOption {
	return func(o *requestOptions) {
		o.request.Header.Set(key, value)
	}
}

// SetBasicAuth sets the request's Authorization header to use HTTP
// Basic Authentication with the provided username and password.
func SetBasicAuth(username, password string) RequestOption {
	return func(o *requestOptions) {
		o.request.SetBasicAuth(username, password)
	}
}

// SetRequest set the request handle
func SetRequest(handle func(req *http.Request) (*http.Request, error)) RequestOption {
	return func(o *requestOptions) {
		o.handle = handle
	}
}
