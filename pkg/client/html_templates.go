package client

import (
	"fmt"
	"html/template"
	"path"
	"path/filepath"
	"strings"

	"github.com/Notifiarr/notifiarr/pkg/bindata"
	"github.com/fsnotify/fsnotify"
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
