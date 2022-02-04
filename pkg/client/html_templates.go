package client

import (
	"fmt"
	"html/template"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Notifiarr/notifiarr/pkg/bindata"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/fsnotify/fsnotify"
	"github.com/mitchellh/go-homedir"
)

// loadAssetsTemplates watches for changs to template files, and loads them.
func (c *Client) loadAssetsTemplates() error {
	if c.Flags.Assets == "" {
		return nil
	}

	if err := c.ParseGUITemplates(); err != nil {
		return err
	}

	fsn, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("fsnotify.NewWatcher: %w", err)
	}

	templates := filepath.Join(c.Flags.Assets, "templates")
	if err := fsn.Add(templates); err != nil {
		return fmt.Errorf("cannot watch '%s' templates path: %w", templates, err)
	}

	go c.watchAssetsTemplates(fsn)

	return nil
}

func (c *Client) watchAssetsTemplates(fsn *fsnotify.Watcher) {
	for {
		select {
		case err := <-fsn.Errors:
			c.Errorf("fsnotify: %v", err)
		case event, ok := <-fsn.Events:
			if !ok {
				return
			}

			if (event.Op&fsnotify.Write != fsnotify.Write && event.Op&fsnotify.Create != fsnotify.Create) ||
				!strings.HasSuffix(event.Name, ".html") {
				continue
			}

			c.Debugf("Got event: %s on %s, reloading HTML templates!", event.Op, event.Name)

			if err := c.StopWebServer(); err != nil {
				panic("Stopping web server: " + err.Error())
			}

			if err := c.ParseGUITemplates(); err != nil {
				c.Errorf("fsnotify/parsing templates: %v", err)
			}

			c.StartWebServer()
		}
	}
}

// ParseGUITemplates parses the baked-in templates, and overrides them if a template directory is provided.
func (c *Client) ParseGUITemplates() (err error) {
	// Index and 404 do not have template files, but they can be customized.
	index := "<p>" + c.Flags.Name() + `: <strong>working</strong></p> <p>(<a href="login">login</a>)</p>`
	c.templat = template.Must(template.New("index.html").Parse(index))
	c.templat = template.Must(c.templat.New("404.html").Parse("NOT FOUND! Check your request parameters and try again."))
	c.templat = c.templat.Funcs(template.FuncMap{
		"megabyte": func(size int64) int64 { return size / mnd.Megabyte },
		"base":     func() string { return strings.TrimSuffix(c.Config.URLBase, "/") },
		"files":    func() string { return path.Join(c.Config.URLBase, "files") },
		"instance": func(idx int) int { return idx + 1 },
	})

	// Parse all our compiled-in templates.
	for _, name := range bindata.AssetNames() {
		if strings.HasPrefix(name, "templates/") {
			c.templat = template.Must(c.templat.New(path.Base(name)).Parse(bindata.MustAssetString(name)))
		}
	}

	if c.Flags.Assets == "" {
		return nil
	}

	templates := filepath.Join(c.Flags.Assets, "templates", "*.html")
	c.Printf("==> Parsing and watching HTML templates @ %s", templates)

	c.templat, err = c.templat.ParseGlob(templates)
	if err != nil {
		return fmt.Errorf("parsing custom template: %w", err)
	}

	return nil
}

// haveCustomFile searches known locatinos for a file. Returns the file's path.
func (c *Client) haveCustomFile(fileName string) string {
	cwd, _ := os.Getwd()
	exe, _ := os.Executable()

	paths := map[string][]string{
		mnd.Windows: {
			`~/notifiarr`,
			cwd,
			filepath.Dir(exe),
			`C:\ProgramData\notifiarr`,
		},
		"darwin": {
			"~/.notifiarr",
			"/usr/local/etc/notifiarr",
		},
		"default": {
			`~/notifiarr`,
			"/config",
			"/etc/notifiarr",
		},
	}

	findIn := paths[runtime.GOOS]
	if len(findIn) == 0 {
		findIn = paths["default"]
	}

	for _, find := range findIn {
		if find == "" {
			continue
		}

		custom, err := homedir.Expand(filepath.Join(find, fileName))
		if err != nil {
			custom = filepath.Join(find, fileName)
		}

		custom2, err := filepath.Abs(custom)
		if err == nil {
			custom = custom2
		}

		if _, err = os.Stat(custom); err == nil {
			return custom
		}
	}

	return ""
}

func getLastLinesInFile(filepath string, count, skip int) ([]byte, error) { //nolint:cyclop
	fileHandle, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}
	defer fileHandle.Close()

	stat, err := fileHandle.Stat()
	if err != nil {
		return nil, fmt.Errorf("stating open file: %w", err)
	}

	var (
		output   = []byte{}
		location int64
		filesize = stat.Size()
		char     = make([]byte, 1)
		found    int
	)

	for {
		location-- // read 1 byte
		if _, err = fileHandle.Seek(location, io.SeekEnd); err != nil {
			return output, fmt.Errorf("seeking open file: %w", err)
		}

		if _, err := fileHandle.Read(char); err != nil {
			return output, fmt.Errorf("reading open file: %w", err)
		}

		if location != -1 && (char[0] == 10) { // nolint:gomnd
			found++ // we found a line
		}

		if skip == 0 || found > skip {
			output = append(char, output...)
		}

		if location == -filesize {
			return output, nil // beginning of file,
		}

		if found >= count+skip {
			// we found enough lines.
			// remove the first character, a newline.
			return output[1:], nil
		}
	}
}
