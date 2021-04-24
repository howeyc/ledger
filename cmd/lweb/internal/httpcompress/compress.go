package httpcompress

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/andybalholm/brotli"
	"github.com/julienschmidt/httprouter"
)

// CompressResponseWriter is a Struct for manipulating io writer
type CompressResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (res CompressResponseWriter) Write(b []byte) (int, error) {
	if "" == res.Header().Get("Content-Type") {
		// If no content type, apply sniffing algorithm to un-gzipped body.
		res.Header().Set("Content-Type", http.DetectContentType(b))
	}
	return res.Writer.Write(b)
}

// Middleware force - bool, whether or not to force Compression regardless of the sent headers.
func Middleware(fn httprouter.Handle, force bool) httprouter.Handle {
	return func(res http.ResponseWriter, req *http.Request, pm httprouter.Params) {
		if !strings.Contains(req.Header.Get("Accept-Encoding"), "br") {
			if !strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") && !force {
				fn(res, req, pm)
				return
			}
			res.Header().Set("Vary", "Accept-Encoding")
			res.Header().Set("Content-Encoding", "gzip")
			gz := gzip.NewWriter(res)
			defer gz.Close()
			cw := CompressResponseWriter{Writer: gz, ResponseWriter: res}
			fn(cw, req, pm)
			return
		}
		res.Header().Set("Vary", "Accept-Encoding")
		res.Header().Set("Content-Encoding", "br")
		br := brotli.NewWriter(res)
		defer br.Close()
		cw := CompressResponseWriter{Writer: br, ResponseWriter: res}
		fn(cw, req, pm)
	}
}
