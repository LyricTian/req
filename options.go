package req

import (
	"net"
	"net/http"
	"time"
)

type requestOptions struct {
	header  http.Header
	request func(req *http.Request) *http.Request
}

// RequestOption request parameter options
type RequestOption func(*requestOptions)

// SetHeader set the request header
func SetHeader(key, value string) RequestOption {
	return func(o *requestOptions) {
		o.header.Add(key, value)
	}
}

// SetRequest set the request handle
func SetRequest(handle func(req *http.Request) *http.Request) RequestOption {
	return func(o *requestOptions) {
		o.request = handle
	}
}

var defaultOptions = options{
	maxQueue:  64,
	maxWorker: 8,
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
	header: make(http.Header),
}

type options struct {
	maxQueue      int
	maxWorker     int
	transport     *http.Transport
	cookieJar     http.CookieJar
	checkRedirect func(req *http.Request, via []*http.Request) error
	timeout       time.Duration
	baseURL       string
	header        http.Header
}

// Option parameter options
type Option func(*options)

// SetMaxQueue set the maximum number of queues in the buffer
func SetMaxQueue(n int) Option {
	return func(o *options) {
		o.maxQueue = n
	}
}

// SetMaxWorker set the maximum number of worker goroutines
func SetMaxWorker(n int) Option {
	return func(o *options) {
		o.maxWorker = n
	}
}

// SetBaseURL set the requested base url
func SetBaseURL(base string) Option {
	return func(o *options) {
		o.baseURL = base
	}
}

// SetBaseHeader set the requested base header
func SetBaseHeader(key, value string) Option {
	return func(o *options) {
		o.header.Add(key, value)
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
