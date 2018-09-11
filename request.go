package req

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

var _ Requester = &request{}

// Requester HTTP request interface
type Requester interface {
	Head(ctx context.Context, urlStr string, queryParam url.Values, opt ...RequestOption) (Responser, error)
	Get(ctx context.Context, urlStr string, queryParam url.Values, opt ...RequestOption) (Responser, error)
	Delete(ctx context.Context, urlStr string, queryParam url.Values, opt ...RequestOption) (Responser, error)
	Patch(ctx context.Context, urlStr string, queryParam url.Values, opt ...RequestOption) (Responser, error)
	Post(ctx context.Context, urlStr string, body io.Reader, opt ...RequestOption) (Responser, error)
	PostJSON(ctx context.Context, urlStr string, body interface{}, opt ...RequestOption) (Responser, error)
	PostForm(ctx context.Context, urlStr string, body url.Values, opt ...RequestOption) (Responser, error)
	Put(ctx context.Context, urlStr string, body io.Reader, opt ...RequestOption) (Responser, error)
	PutJSON(ctx context.Context, urlStr string, body interface{}, opt ...RequestOption) (Responser, error)
	PutForm(ctx context.Context, urlStr string, body url.Values, opt ...RequestOption) (Responser, error)
	Do(ctx context.Context, urlStr, method string, body io.Reader, opt ...RequestOption) (Responser, error)
}

// RequestURL get request url
func RequestURL(base, router string) string {
	var buf bytes.Buffer
	if l := len(base); l > 0 {
		if base[l-1] == '/' {
			base = base[:l-1]
		}
		buf.WriteString(base)

		if rl := len(router); rl > 0 {
			if router[0] != '/' {
				buf.WriteByte('/')
			}
		}
	}
	buf.WriteString(router)
	return buf.String()
}

// New create a request instance
func New(opt ...Option) Requester {
	opts := defaultOptions
	for _, o := range opt {
		o(&opts)
	}

	req := &request{
		opts: opts,
		cli: &http.Client{
			Transport:     opts.transport,
			CheckRedirect: opts.checkRedirect,
			Jar:           opts.cookieJar,
			Timeout:       opts.timeout,
		},
	}

	return req
}

type request struct {
	opts options
	cli  *http.Client
}

func (r *request) parseQueryParam(urlStr string, param url.Values) string {
	if param != nil {
		c := '?'
		if strings.IndexByte(urlStr, '?') != -1 {
			c = '&'
		}
		urlStr = fmt.Sprintf("%s%c%s", urlStr, c, param.Encode())
	}
	return urlStr
}

func (r *request) setContentType(contentType string, opt ...RequestOption) []RequestOption {
	var ro []RequestOption
	ro = append(ro, SetHeader("Content-Type", contentType))
	if len(opt) > 0 {
		ro = append(ro, opt...)
	}
	return ro
}

func (r *request) fillRequest(req *http.Request, opts ...RequestOption) (*http.Request, error) {
	if req.Header == nil {
		req.Header = make(http.Header)
	}

	for k := range r.opts.header {
		req.Header.Set(k, r.opts.header.Get(k))
	}

	ro := &requestOptions{
		request: req,
	}
	for _, opt := range opts {
		opt(ro)
	}

	if fn := ro.handle; fn != nil {
		return fn(req)
	}

	return req, nil
}

func (r *request) doForm(ctx context.Context, urlStr, method string, body url.Values, opt ...RequestOption) (Responser, error) {
	var s string
	if body != nil {
		s = body.Encode()
	}

	ro := r.setContentType("application/x-www-form-urlencoded", opt...)
	return r.Do(ctx, urlStr, method, strings.NewReader(s), ro...)
}

func (r *request) doJSON(ctx context.Context, urlStr, method string, body interface{}, opt ...RequestOption) (Responser, error) {
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(body)
	if err != nil {
		return nil, err
	}

	ro := r.setContentType("application/json; charset=UTF-8", opt...)
	return r.Do(ctx, urlStr, method, buf, ro...)
}

func (r *request) httpDo(ctx context.Context, req *http.Request, f func(*http.Response, error) error) error {
	c := make(chan error, 1)
	go func() { c <- f(r.cli.Do(req)) }()
	select {
	case <-ctx.Done():
		r.opts.transport.CancelRequest(req)
		<-c
		return ctx.Err()
	case err := <-c:
		return err
	}
}

func (r *request) Head(ctx context.Context, urlStr string, queryParam url.Values, opt ...RequestOption) (Responser, error) {
	return r.Do(ctx, r.parseQueryParam(urlStr, queryParam), http.MethodHead, nil, opt...)
}

func (r *request) Get(ctx context.Context, urlStr string, queryParam url.Values, opt ...RequestOption) (Responser, error) {
	return r.Do(ctx, r.parseQueryParam(urlStr, queryParam), http.MethodGet, nil, opt...)
}

func (r *request) Delete(ctx context.Context, urlStr string, queryParam url.Values, opt ...RequestOption) (Responser, error) {
	return r.Do(ctx, r.parseQueryParam(urlStr, queryParam), http.MethodDelete, nil, opt...)
}

func (r *request) Patch(ctx context.Context, urlStr string, queryParam url.Values, opt ...RequestOption) (Responser, error) {
	return r.Do(ctx, r.parseQueryParam(urlStr, queryParam), http.MethodPatch, nil, opt...)
}

func (r *request) Post(ctx context.Context, urlStr string, body io.Reader, opt ...RequestOption) (Responser, error) {
	return r.Do(ctx, urlStr, http.MethodPost, body, opt...)
}

func (r *request) PostJSON(ctx context.Context, urlStr string, body interface{}, opt ...RequestOption) (Responser, error) {
	return r.doJSON(ctx, urlStr, http.MethodPost, body, opt...)
}

func (r *request) PostForm(ctx context.Context, urlStr string, body url.Values, opt ...RequestOption) (Responser, error) {
	return r.doForm(ctx, urlStr, http.MethodPost, body, opt...)
}

func (r *request) Put(ctx context.Context, urlStr string, body io.Reader, opt ...RequestOption) (Responser, error) {
	return r.Do(ctx, urlStr, http.MethodPut, body, opt...)
}

func (r *request) PutJSON(ctx context.Context, urlStr string, body interface{}, opt ...RequestOption) (Responser, error) {
	return r.doJSON(ctx, urlStr, http.MethodPut, body, opt...)
}

func (r *request) PutForm(ctx context.Context, urlStr string, body url.Values, opt ...RequestOption) (Responser, error) {
	return r.doForm(ctx, urlStr, http.MethodPut, body, opt...)
}

func (r *request) Do(ctx context.Context, urlStr, method string, body io.Reader, opt ...RequestOption) (Responser, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	url := RequestURL(r.opts.baseURL, urlStr)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req, err = r.fillRequest(req, opt...)
	if err != nil {
		return nil, err
	}

	var resp Responser
	err = r.httpDo(ctx, req, func(res *http.Response, err error) error {
		if err != nil {
			return err
		}
		resp = newResponse(res)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}
