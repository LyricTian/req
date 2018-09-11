package req

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGet(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(500)
			fmt.Fprintf(w, "Not expected value,Method:%s", r.Method)
			return
		}

		if r.Header.Get("BFoo") != "bar" {
			w.WriteHeader(500)
			fmt.Fprintf(w, "Not expected value,Header - BHFoo:%s", r.Header.Get("Foo"))
			return
		}

		if r.Header.Get("HFoo") != "bar" {
			w.WriteHeader(500)
			fmt.Fprintf(w, "Not expected value,Header - HFoo:%s", r.Header.Get("Foo"))
			return
		}

		if r.URL.Query().Get("QFoo") != "bar" {
			w.WriteHeader(500)
			fmt.Fprintf(w, "Not expected value,Query - QFoo:%s", r.Header.Get("Foo"))
			return
		}
		fmt.Fprint(w, "ok")
	}))
	defer ts.Close()

	Convey("Test Get Request", t, func() {
		r := New(SetBaseHeader("BFoo", "bar"))

		queryParam := make(url.Values)
		queryParam.Add("QFoo", "bar")
		resp, err := r.Get(context.Background(), ts.URL, queryParam, SetHeader("HFoo", "bar"))
		So(err, ShouldBeNil)
		So(resp, ShouldNotBeNil)

		body, err := resp.String()
		So(err, ShouldBeNil)
		So(body, ShouldEqual, "ok")
		So(resp.Response().StatusCode, ShouldEqual, 200)
	})
}

func TestPostJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(500)
			fmt.Fprintf(w, "Not expected value,Method:%s", r.Method)
			return
		}

		if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
			w.WriteHeader(500)
			fmt.Fprintf(w, "Not expected value,Header - ContentType:%s", r.Header.Get("Content-Type"))
			return
		}

		var body struct {
			Foo string `json:"foo"`
		}

		defer r.Body.Close()
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintf(w, "Not expected value,Body - Error:%s", err.Error())
			return
		}

		if body.Foo != "bar" {
			w.WriteHeader(500)
			fmt.Fprintf(w, "Not expected value,Body - foo:%s", body.Foo)
			return
		}

		fmt.Fprint(w, "ok")
	}))
	defer ts.Close()

	Convey("Test Post JSON Request", t, func() {
		r := New()

		resp, err := r.PostJSON(context.Background(), ts.URL, map[string]string{"foo": "bar"})
		So(err, ShouldBeNil)
		So(resp, ShouldNotBeNil)

		body, err := resp.String()
		So(err, ShouldBeNil)
		So(body, ShouldEqual, "ok")
		So(resp.Response().StatusCode, ShouldEqual, 200)
	})
}

func BenchmarkGet(b *testing.B) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	}))
	defer ts.Close()

	r := New()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		resp, err := r.Get(context.Background(), ts.URL, nil)
		if err != nil {
			b.Error(err.Error())
			return
		}
		body, err := resp.String()
		if err != nil {
			b.Error(err.Error())
			return
		}
		if body != "ok" {
			b.Errorf("Not expected value:%s", body)
		}
	}
}
