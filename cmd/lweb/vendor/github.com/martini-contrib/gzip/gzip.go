package gzip

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/go-martini/martini"
)

const (
	HeaderAcceptEncoding  = "Accept-Encoding"
	HeaderContentEncoding = "Content-Encoding"
	HeaderContentLength   = "Content-Length"
	HeaderContentType     = "Content-Type"
	HeaderVary            = "Vary"
)

var serveGzip = func(w http.ResponseWriter, r *http.Request, c martini.Context) {
	if !strings.Contains(r.Header.Get(HeaderAcceptEncoding), "gzip") {
		return
	}

	headers := w.Header()
	headers.Set(HeaderContentEncoding, "gzip")
	headers.Set(HeaderVary, HeaderAcceptEncoding)

	gz := gzip.NewWriter(w)
	defer gz.Close()

	gzw := gzipResponseWriter{gz, w.(martini.ResponseWriter)}
	c.MapTo(gzw, (*http.ResponseWriter)(nil))

	c.Next()

	// delete content length after we know we have been written to
	gzw.Header().Del("Content-Length")
}

// All returns a Handler that adds gzip compression to all requests
func All() martini.Handler {
	return serveGzip
}

type gzipResponseWriter struct {
	w *gzip.Writer
	martini.ResponseWriter
}

func (grw gzipResponseWriter) Write(p []byte) (int, error) {
	if len(grw.Header().Get(HeaderContentType)) == 0 {
		grw.Header().Set(HeaderContentType, http.DetectContentType(p))
	}

	return grw.w.Write(p)
}

func (grw gzipResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := grw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("the ResponseWriter doesn't support the Hijacker interface")
	}
	return hijacker.Hijack()
}
