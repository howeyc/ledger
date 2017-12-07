package staticbin

import (
	"bytes"
	"log"
	"net/http"
	"path"
	"time"

	"github.com/go-martini/martini"
)

// Static returns a middleware handler that serves static files in the given directory.
func Static(dir string, asset func(string) ([]byte, error), options ...Options) martini.Handler {
	if asset == nil {
		return func() {}
	}

	opt := retrieveOptions(options)

	modtime := time.Now()

	return func(res http.ResponseWriter, req *http.Request, log *log.Logger) {
		if req.Method != "GET" && req.Method != "HEAD" {
			return
		}

		url := req.URL.Path

		b, err := asset(dir + url)

		if err != nil {
			// Try to serve the index file.
			b, err = asset(path.Join(dir+url, opt.IndexFile))

			if err != nil {
				// Exit if the asset could not be found.
				return
			}
		}

		if !opt.SkipLogging {
			log.Println("[Static] Serving " + url)
		}

		http.ServeContent(res, req, url, modtime, bytes.NewReader(b))
	}
}
