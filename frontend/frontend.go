package frontend

//go:generate sh generate.sh

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
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
	Root    fs.FS
	Locale  fs.FS
	//go:embed dist
	embedded embed.FS
	//go:embed src/includes/locale/*.json
	locale embed.FS
)

//nolint:gochecknoinits
func init() {
	Root, _ = fs.Sub(embedded, "dist")
	handler = http.FileServerFS(Root)
	Locale, _ = fs.Sub(locale, "src/includes/locale")
}

type responseWriter struct {
	http.ResponseWriter
	Asset     bool
	Asset2    bool
	SendIndex bool
}

// Languages is a map of the available languages and their display names localized to the parent language.
// The key is the parent language, and the value is a map of the available
// languages and their display names localized to the parent language.
type Languages map[string]map[string]LocalizedLanguage

// LocalizedLanguage is a language and its display name localized to itself and another (parent) language.
type LocalizedLanguage struct {
	// Lang is the parent language code this language Name is localized to.
	Lang string `json:"lang"`
	// Code is the language code of the language.
	Code string `json:"code"`
	// Name is the display name of the language localized to the parent (Lang) language.
	Name string `json:"name"`
	// Self is the display name of the language localized in its own language.
	Self string `json:"self"`
}

// Translations returns all the configured frontend languages.
// The frontend uses this to populate the language dropdown localized to the currently selected language.
func Translations() Languages {
	output := make(Languages)

	for _, parent := range langs {
		output[parent] = map[string]LocalizedLanguage{}
		curTag := language.MustParse(parent)

		for _, name := range langs {
			lang := language.MustParse(name)
			cur := display.Languages(curTag)
			title := cases.Title(curTag)
			selfTitle := cases.Title(lang)
			output[parent][name] = LocalizedLanguage{
				Code: name,
				Name: title.String(cur.Name(lang)),
				Self: selfTitle.String(display.Self.Name(lang)),
				Lang: parent,
			}
		}
	}

	return output
}

// IndexHandler returns an asset from the file system if it exists, otherwise the index page.
// Useful for a single page app.
func IndexHandler(resp http.ResponseWriter, req *http.Request) {
	// Remove the URLBase from the file's path.
	req.URL.Path = stripBefore(strings.TrimPrefix(
		strings.TrimPrefix(req.URL.Path, URLBase), "/"), "/assets/")

	// The frontend uses this cookie to know what path to send API requests to.
	http.SetCookie(resp, &http.Cookie{Name: "urlbase", Value: URLBase})

	response := &responseWriter{
		ResponseWriter: resp,
		Asset:          strings.HasPrefix(req.URL.Path, "assets/"),
		Asset2:         isAsset(req.URL.Path),
	}
	handler.ServeHTTP(response, req)

	if !response.SendIndex {
		return
	}

	resp.Header().Set("Content-Type", "text/html; charset=utf-8")
	http.ServeFileFS(resp, req, Root, "index.html")
}

// isAsset returns true if the file is an asset or an allowed file type in the public/ directory.
func isAsset(filepath string) bool {
	return strings.HasPrefix(filepath, "assets/") || map[string]bool{
		".json": true,
		".svg":  true,
		".png":  true,
		".jpg":  true,
		".ico":  true,
	}[strings.ToLower(path.Ext(filepath))]
}

func (w *responseWriter) WriteHeader(status int) {
	w.ResponseWriter.Header().Set("Cache-Control", "no-cache")

	// if it's found, return 200
	if status != http.StatusNotFound {
		if w.Asset {
			w.ResponseWriter.Header().Set("Cache-Control", "max-age=31536000")
		} else if w.Asset2 {
			w.ResponseWriter.Header().Set("Cache-Control", "max-age=300")
		}
		w.ResponseWriter.WriteHeader(status)

		return
	}

	// if the request was for an asset and it's 404, return 404.
	if w.Asset || w.Asset2 {
		w.ResponseWriter.WriteHeader(http.StatusNotFound)
		return
	}

	// if it's not found and doesn't contain "assets", return index
	w.SendIndex = true
}

func (w *responseWriter) Write(p []byte) (int, error) {
	if !w.SendIndex {
		return w.ResponseWriter.Write(p) //nolint:wrapcheck
	}

	return len(p), nil
}

// stripBefore strips any prefix from a string if the sub-string exists.
func stripBefore(s, sub string) string {
	if index := strings.Index(s, sub); index != -1 {
		return s[index:]
	}

	return s
}
