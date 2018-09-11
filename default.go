package req

import (
	"context"
	"io"
	"net/url"
	"sync"
)

var (
	internalReq Requester
	once        sync.Once
)

func req(opt ...Option) Requester {
	once.Do(func() {
		internalReq = New(opt...)
	})
	return internalReq
}

// SetOptions set the parameter options
func SetOptions(opt ...Option) {
	req(opt...)
}

// Head head request
func Head(ctx context.Context, urlStr string, queryParam url.Values, opt ...RequestOption) (Responser, error) {
	return req().Head(ctx, urlStr, queryParam, opt...)
}

// Get get request
func Get(ctx context.Context, urlStr string, queryParam url.Values, opt ...RequestOption) (Responser, error) {
	return req().Get(ctx, urlStr, queryParam, opt...)
}

// Delete delete request
func Delete(ctx context.Context, urlStr string, queryParam url.Values, opt ...RequestOption) (Responser, error) {
	return req().Delete(ctx, urlStr, queryParam, opt...)
}

// Patch patch request
func Patch(ctx context.Context, urlStr string, queryParam url.Values, opt ...RequestOption) (Responser, error) {
	return req().Patch(ctx, urlStr, queryParam, opt...)
}

// Post post request
func Post(ctx context.Context, urlStr string, body io.Reader, opt ...RequestOption) (Responser, error) {
	return req().Post(ctx, urlStr, body, opt...)
}

// PostJSON post json request
func PostJSON(ctx context.Context, urlStr string, body interface{}, opt ...RequestOption) (Responser, error) {
	return req().PostJSON(ctx, urlStr, body, opt...)
}

// PostForm post form request
func PostForm(ctx context.Context, urlStr string, body url.Values, opt ...RequestOption) (Responser, error) {
	return req().PostForm(ctx, urlStr, body, opt...)
}

// Put put request
func Put(ctx context.Context, urlStr string, body io.Reader, opt ...RequestOption) (Responser, error) {
	return req().Put(ctx, urlStr, body, opt...)
}

// PutJSON put json request
func PutJSON(ctx context.Context, urlStr string, body interface{}, opt ...RequestOption) (Responser, error) {
	return req().PutJSON(ctx, urlStr, body, opt...)
}

// PutForm put form request
func PutForm(ctx context.Context, urlStr string, body url.Values, opt ...RequestOption) (Responser, error) {
	return req().PutForm(ctx, urlStr, body, opt...)
}

// Do http request
func Do(ctx context.Context, urlStr, method string, body io.Reader, opt ...RequestOption) (Responser, error) {
	return req().Do(ctx, urlStr, method, body, opt...)
}
