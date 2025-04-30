//nolint:gochecknoglobals,wrapcheck
package frontend

//go:generate npm install
//go:generate npm run build
import (
	"embed"
	"io/fs"
	"net/http"
)

var (
	handler http.Handler
	root    fs.FS
	//go:embed dist
	embedded embed.FS
)

func init() {
	root, _ = fs.Sub(embedded, "dist")
	handler = http.FileServer(http.FS(root))
}

type responseWriter struct {
	http.ResponseWriter
	Status int
}

func (w *responseWriter) WriteHeader(status int) {
	if w.Status = status; status != http.StatusNotFound {
		w.ResponseWriter.WriteHeader(status)
	}
}

func (w *responseWriter) Write(p []byte) (int, error) {
	if w.Status != http.StatusNotFound {
		return w.ResponseWriter.Write(p)
	}

	return len(p), nil
}

func IndexHandler(resp http.ResponseWriter, req *http.Request) {
	response := &responseWriter{ResponseWriter: resp}
	handler.ServeHTTP(response, req)

	if response.Status == http.StatusNotFound {
		resp.Header().Set("Content-Type", "text/html")
		http.ServeFileFS(resp, req, root, "index.html")
	}
}
