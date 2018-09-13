package req

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

var _ Responser = &response{}

// Responser HTTP response interface
type Responser interface {
	StatusCode() int
	Response() *http.Response
	String() (string, error)
	Bytes() ([]byte, error)
	JSON(v interface{}) error
	Close()
}

func newResponse(resp *http.Response) *response {
	return &response{resp}
}

type response struct {
	resp *http.Response
}

func (r *response) StatusCode() int {
	return r.resp.StatusCode
}

func (r *response) Response() *http.Response {
	return r.resp
}

func (r *response) String() (string, error) {
	b, err := r.Bytes()
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (r *response) Bytes() ([]byte, error) {
	defer r.resp.Body.Close()

	buf, err := ioutil.ReadAll(r.resp.Body)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func (r *response) JSON(v interface{}) error {
	defer r.resp.Body.Close()

	return json.NewDecoder(r.resp.Body).Decode(v)
}

func (r *response) Close() {
	if !r.resp.Close {
		r.resp.Body.Close()
	}
}
