package gzip

import (
	"github.com/go-martini/martini"
	"bufio"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_GzipAll(t *testing.T) {
	// Set up
	recorder := httptest.NewRecorder()
	before := false

	m := martini.New()
	m.Use(All())
	m.Use(func(r http.ResponseWriter) {
		r.(martini.ResponseWriter).Before(func(rw martini.ResponseWriter) {
			before = true
		})
	})

	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Error(err)
	}

	m.ServeHTTP(recorder, r)

	// Make our assertions
	_, ok := recorder.HeaderMap[HeaderContentEncoding]
	if ok {
		t.Error(HeaderContentEncoding + " present")
	}

	ce := recorder.Header().Get(HeaderContentEncoding)
	if strings.EqualFold(ce, "gzip") {
		t.Error(HeaderContentEncoding + " is 'gzip'")
	}

	recorder = httptest.NewRecorder()
	r.Header.Set(HeaderAcceptEncoding, "gzip")
	m.ServeHTTP(recorder, r)

	// Make our assertions
	_, ok = recorder.HeaderMap[HeaderContentEncoding]
	if !ok {
		t.Error(HeaderContentEncoding + " not present")
	}

	ce = recorder.Header().Get(HeaderContentEncoding)
	if !strings.EqualFold(ce, "gzip") {
		t.Error(HeaderContentEncoding + " is not 'gzip'")
	}

	if before == false {
		t.Error("Before hook was not called")
	}
}

type hijackableResponse struct {
	Hijacked bool
	header   http.Header
}

func newHijackableResponse() *hijackableResponse {
	return &hijackableResponse{header: make(http.Header)}
}

func (h *hijackableResponse) Header() http.Header           { return h.header }
func (h *hijackableResponse) Write(buf []byte) (int, error) { return 0, nil }
func (h *hijackableResponse) WriteHeader(code int)          {}
func (h *hijackableResponse) Flush()                        {}
func (h *hijackableResponse) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h.Hijacked = true
	return nil, nil, nil
}

func Test_ResponseWriter_Hijack(t *testing.T) {
	hijackable := newHijackableResponse()

	m := martini.New()
	m.Use(All())
	m.Use(func(rw http.ResponseWriter) {
		if hj, ok := rw.(http.Hijacker); !ok {
			t.Error("Unable to hijack")
		} else {
			hj.Hijack()
		}
	})

	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Error(err)
	}

	r.Header.Set(HeaderAcceptEncoding, "gzip")
	m.ServeHTTP(hijackable, r)

	if !hijackable.Hijacked {
		t.Error("Hijack was not called")
	}
}
