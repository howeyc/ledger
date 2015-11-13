package staticbin

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStatic(t *testing.T) {
	// Case when Asset is nil.
	h := Static("public", nil)
	if h == nil {
		t.Errorf("returned value should not be nil")
	}

	// Case when req.Method != "GET" && req.Method != "HEAD".
	fnc := func(s string) ([]byte, error) {
		return []byte("test"), nil
	}
	res := httptest.NewRecorder()
	m := Classic(fnc)
	req, err := http.NewRequest("POST", "http://localhost:3000/static.go", nil)
	if err != nil {
		t.Error(err)
	}
	m.ServeHTTP(res, req)
	if body := strings.TrimSpace(res.Body.String()); body != "404 page not found" {
		t.Errorf(
			"returned value is invalid [actual: %s][expected: %s]",
			body,
			"404 page not found",
		)
	}

	// Case when Asset(dir + path) returns an error.
	fnc = func(s string) ([]byte, error) {
		return nil, fmt.Errorf("test error")
	}
	res = httptest.NewRecorder()
	m = Classic(fnc)
	req, err = http.NewRequest("GET", "http://localhost:3000/static.go", nil)
	if err != nil {
		t.Error(err)
	}
	m.ServeHTTP(res, req)
	if body := strings.TrimSpace(res.Body.String()); body != "404 page not found" {
		t.Errorf(
			"returned value is invalid [actual: %s][expected: %s]",
			body,
			"404 page not found",
		)
	}

	// Case when Static serves a file.
	fnc = func(s string) ([]byte, error) {
		return []byte("test"), nil
	}
	res = httptest.NewRecorder()
	m = Classic(fnc)
	req, err = http.NewRequest("GET", "http://localhost:3000/static.go", nil)
	if err != nil {
		t.Error(err)
	}
	m.ServeHTTP(res, req)
	if body := strings.TrimSpace(res.Body.String()); body != "test" {
		t.Errorf(
			"returned value is invalid [actual: %s][expected: %s]",
			body,
			"test",
		)
	}
}
