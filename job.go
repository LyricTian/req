package req

import (
	"context"
	"net/http"
)

type jobResult struct {
	err  error
	resp *http.Response
}

type job struct {
	ctx    context.Context
	req    *http.Request
	tr     *http.Transport
	cli    *http.Client
	result chan *jobResult
}

func (j *job) Reset(ctx context.Context, req *http.Request) {
	j.ctx = ctx
	j.req = req
	j.result = make(chan *jobResult)
}

func (j *job) Result() <-chan *jobResult {
	return j.result
}

func (j *job) Job() {
	var result *jobResult
	j.httpDo(j.ctx, j.req, func(resp *http.Response, err error) error {
		result = &jobResult{
			err:  err,
			resp: resp,
		}
		return nil
	})
	j.result <- result
	close(j.result)
	return
}

func (j *job) httpDo(ctx context.Context, req *http.Request, f func(*http.Response, error) error) error {
	c := make(chan error, 1)
	go func() { c <- f(j.cli.Do(req)) }()
	select {
	case <-ctx.Done():
		j.tr.CancelRequest(req)
		<-c
		return ctx.Err()
	case err := <-c:
		return err
	}
}
