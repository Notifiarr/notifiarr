package frontend

//go:generate sh generate.sh

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
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
	handler = http.FileServerFS(root)
}

type responseWriter struct {
	http.ResponseWriter
	Asset     bool
	SendIndex bool
}

// Languages is a map of the available languages and their display names localized to the parent language.
// The key is the parent language, and the value is a map of the available languages and their display names localized to the parent language.
type Languages map[string]map[string]string

// Translations returns all the configured frontend languages.
// The frontend uses this to populate the language dropdown localized to the current selected language.
func Translations() Languages {
	output := make(Languages)

	for _, parent := range langs {
		output[parent] = map[string]string{}
		curTag := language.MustParse(parent)

		for _, name := range langs {
			lang := language.MustParse(name)
			cur := display.Languages(curTag)
			title := cases.Title(curTag)
			selfTitle := cases.Title(lang)
			output[parent][name] = fmt.Sprintf("%s (%s)", title.String(cur.Name(lang)), selfTitle.String(display.Self.Name(lang)))
		}
	}

	return output
}

// IndexHandler returns an asset from the file system if it exists, otherwise the index page.
// Useful for a single page app.
func IndexHandler(resp http.ResponseWriter, req *http.Request) {
	// We serve assets from any parent path.
	asset, path := stripBefore(req.URL.Path, "/assets/")
	if asset {
		req.URL.Path = path
	}

	response := &responseWriter{ResponseWriter: resp, Asset: asset}
	handler.ServeHTTP(response, req)

	if !response.SendIndex {
		return
	}

	// The frontend uses this cookie to know what path to send API requests to.
	http.SetCookie(resp, &http.Cookie{Name: "urlbase", Value: URLBase})
	resp.Header().Set("Content-Type", "text/html")
	http.ServeFileFS(resp, req, root, "index.html")
}

func (w *responseWriter) WriteHeader(status int) {
	w.ResponseWriter.Header().Set("Cache-Control", "no-cache")

	// if it's found, return 200
	if status != http.StatusNotFound {
		w.ResponseWriter.Header().Set("Cache-Control", "max-age=31536000")
		w.ResponseWriter.WriteHeader(status)

		return
	}

	// if the request was for an asset and it's 404, return 404.
	if w.Asset {
		w.ResponseWriter.WriteHeader(http.StatusNotFound)
		return
	}

	// if it's not found and doesn't contain "assets", return index
	w.ResponseWriter.WriteHeader(http.StatusOK)
	w.SendIndex = true
}

func (w *responseWriter) Write(p []byte) (int, error) {
	if !w.SendIndex {
		return w.ResponseWriter.Write(p) //nolint:wrapcheck
	}

	return len(p), nil
}

// stripBefore strips any prefix from a string if the sub-string exists.
func stripBefore(s, sub string) (bool, string) {
	if index := strings.Index(s, sub); index != -1 {
		return true, s[index:]
	}

	return false, s
}
