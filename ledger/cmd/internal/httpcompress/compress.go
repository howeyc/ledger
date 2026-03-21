package httpcompress

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/andybalholm/brotli"
)

// CompressResponseWriter is a Struct for manipulating io writer
type CompressResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (res CompressResponseWriter) Write(b []byte) (int, error) {
	if "" == res.ResponseWriter.Header().Get("Content-Type") {
		// If no content type, apply sniffing algorithm to un-gzipped body.
		res.ResponseWriter.Header().Set("Content-Type", http.DetectContentType(b))
	}
	return res.Writer.Write(b)
}

// Middleware force - bool, whether or not to force Compression regardless of the sent headers.
func Middleware(fn http.HandlerFunc, force bool) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if !strings.Contains(req.Header.Get("Accept-Encoding"), "br") {
			if !strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") && !force {
				fn(res, req)
				return
			}
			res.Header().Set("Vary", "Accept-Encoding")
			res.Header().Set("Content-Encoding", "gzip")
			gz := gzip.NewWriter(res)
			defer gz.Close()
			cw := CompressResponseWriter{Writer: gz, ResponseWriter: res}
			fn(cw, req)
			return
		}
		res.Header().Set("Vary", "Accept-Encoding")
		res.Header().Set("Content-Encoding", "br")
		br := brotli.NewWriter(res)
		defer br.Close()
		cw := CompressResponseWriter{Writer: br, ResponseWriter: res}
		fn(cw, req)
	}
}
