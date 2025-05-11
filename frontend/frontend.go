package frontend

//go:generate npm install
//go:generate npm run build
import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// URLBase is the base URL of the application.
// It is set by the server and injected into the frontend using a cookie.
//
//nolint:gochecknoglobals
var URLBase = "/"

//nolint:gochecknoglobals
var (
	handler http.Handler
	root    fs.FS
	//go:embed dist
	embedded embed.FS
)

//nolint:gochecknoinits
func init() {
	root, _ = fs.Sub(embedded, "dist")
	handler = http.FileServer(http.FS(root))
}

type responseWriter struct {
	http.ResponseWriter
	Path      string
	SendIndex bool
}

// We do not return index for files that have period in them. Only folders.
func (w *responseWriter) hasDot() bool {
	return strings.Contains(path.Base(w.Path), ".")
}

func (w *responseWriter) WriteHeader(status int) {
	w.ResponseWriter.Header().Set("Cache-Control", "no-cache")

	// if it's found, return 200
	if status != http.StatusNotFound {
		w.ResponseWriter.Header().Set("Cache-Control", "max-age=31536000")
		w.ResponseWriter.WriteHeader(status)

		return
	}

	// if it's not found and has a dot, return 404.
	if w.hasDot() {
		w.ResponseWriter.WriteHeader(http.StatusNotFound)
		return
	}

	// if it's not found and doesn't contain a dot, return index
	w.ResponseWriter.WriteHeader(http.StatusOK)
	w.SendIndex = true
}

func (w *responseWriter) Write(p []byte) (int, error) {
	if !w.SendIndex {
		return w.ResponseWriter.Write(p) //nolint:wrapcheck
	}

	return len(p), nil
}

// IndexHandler returns an asset from the file system if it exists, otherwise the index page.
// Useful for a single page app.
func IndexHandler(resp http.ResponseWriter, req *http.Request) {
	response := &responseWriter{ResponseWriter: resp, Path: req.URL.Path}
	handler.ServeHTTP(response, req)

	if response.SendIndex {
		http.SetCookie(resp, &http.Cookie{
			Name:  "urlbase",
			Value: URLBase,
		})
		resp.Header().Set("Content-Type", "text/html")
		http.ServeFileFS(resp, req, root, "index.html")
	}
}
