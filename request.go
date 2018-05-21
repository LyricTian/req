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
	"sync"

	"github.com/LyricTian/queue"
)

var (
	_ Requester = &request{}
	_ Responser = &response{}
)

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

// Responser HTTP response interface
type Responser interface {
	String() (string, error)
	Bytes() ([]byte, error)
	JSON(v interface{}) error
	Response() *http.Response
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

	cli := &http.Client{
		Transport:     opts.transport,
		CheckRedirect: opts.checkRedirect,
		Jar:           opts.cookieJar,
		Timeout:       opts.timeout,
	}

	req := &request{
		opts:   opts,
		q:      queue.NewQueue(opts.maxQueue, opts.maxWorker),
		client: cli,
		pool: sync.Pool{
			New: func() interface{} {
				return &job{
					tr:  opts.transport,
					cli: cli,
				}
			},
		},
	}
	req.q.Run()

	return req
}

type request struct {
	opts   options
	q      *queue.Queue
	client *http.Client
	pool   sync.Pool
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

func (r *request) Do(ctx context.Context, urlStr, method string, body io.Reader, opt ...RequestOption) (Responser, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	url := RequestURL(r.opts.baseURL, urlStr)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req = r.fillRequest(req, opt...)

	job := r.pool.Get().(*job)
	job.Reset(ctx, req)
	r.q.Push(job)

	result := <-job.Result()
	if result.err != nil {
		return nil, result.err
	}

	return newResponse(result.resp), nil
}

func (r *request) setContentType(contentType string, opt ...RequestOption) []RequestOption {
	var ro []RequestOption
	ro = append(ro, SetHeader("Content-Type", contentType))
	if len(opt) > 0 {
		ro = append(ro, opt...)
	}
	return ro
}

func (r *request) fillRequest(req *http.Request, opt ...RequestOption) *http.Request {
	if len(r.opts.header) > 0 {
		for k, v := range r.opts.header {
			for _, vv := range v {
				req.Header.Add(k, vv)
			}
		}
	}

	ro := requestOptions{
		header: make(http.Header),
	}
	for _, o := range opt {
		o(&ro)
	}

	if len(ro.header) > 0 {
		for k, v := range ro.header {
			for _, vv := range v {
				req.Header.Add(k, vv)
			}
		}
	}

	if fn := ro.request; fn != nil {
		return fn(req)
	}

	return req
}
