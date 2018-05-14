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
	Get(ctx context.Context, urlStr string, queryParam url.Values, opt ...RequestOption) (Responser, error)
	Post(ctx context.Context, urlStr string, body io.Reader, opt ...RequestOption) (Responser, error)
	PostJSON(ctx context.Context, urlStr string, body interface{}, opt ...RequestOption) (Responser, error)
	PostForm(ctx context.Context, urlStr string, body url.Values, opt ...RequestOption) (Responser, error)
	Do(ctx context.Context, urlStr, method string, body io.Reader, opt ...RequestOption) (Responser, error)
}

// Responser HTTP response interface
type Responser interface {
	String() (string, error)
	Bytes() ([]byte, error)
	JSON(v interface{}) error
	Response() *http.Response
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

func (r *request) Get(ctx context.Context, urlStr string, queryParam url.Values, opt ...RequestOption) (Responser, error) {
	if queryParam != nil {
		c := '?'
		if strings.IndexByte(urlStr, '?') != -1 {
			c = '&'
		}
		urlStr = fmt.Sprintf("%s%c%s", urlStr, c, queryParam.Encode())
	}
	return r.Do(ctx, urlStr, http.MethodGet, nil, opt...)
}

func (r *request) Post(ctx context.Context, urlStr string, body io.Reader, opt ...RequestOption) (Responser, error) {
	return r.Do(ctx, urlStr, http.MethodPost, body, opt...)
}

func (r *request) PostJSON(ctx context.Context, urlStr string, body interface{}, opt ...RequestOption) (Responser, error) {
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(body)
	if err != nil {
		return nil, err
	}

	ro := r.setContentType("application/json; charset=UTF-8", opt...)
	return r.Do(ctx, urlStr, http.MethodPost, buf, ro...)
}

func (r *request) PostForm(ctx context.Context, urlStr string, body url.Values, opt ...RequestOption) (Responser, error) {
	var s string
	if body != nil {
		s = body.Encode()
	}

	ro := r.setContentType("application/x-www-form-urlencoded", opt...)
	return r.Do(ctx, urlStr, http.MethodPost, strings.NewReader(s), ro...)
}

func (r *request) Do(ctx context.Context, urlStr, method string, body io.Reader, opt ...RequestOption) (Responser, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	var buf bytes.Buffer
	baseURL := r.opts.baseURL
	if l := len(baseURL); l > 0 {
		if baseURL[l-1] == '/' {
			baseURL = baseURL[:l-1]
		}
		buf.WriteString(baseURL)

		if rl := len(urlStr); rl > 0 {
			if urlStr[0] != '/' {
				buf.WriteByte('/')
			}
		}
	}
	buf.WriteString(urlStr)

	req, err := http.NewRequest(method, buf.String(), body)
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
