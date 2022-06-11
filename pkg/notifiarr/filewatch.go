package notifiarr

import (
	"fmt"
	"io"
	"log"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/exp"
	"github.com/nxadm/tail"
)

var ErrInvalidRegexp = fmt.Errorf("invalid regexp")

const (
	maxRetries    = 6
	retryInterval = 10 * time.Second
)

type WatchFile struct {
	//	Cooldown  cnfg.Duration `json:"cooldown" toml:"cooldown" xml:"cooldown" yaml:"cooldown"`
	Path      string `json:"path" toml:"path" xml:"path" yaml:"path"`
	Regexp    string `json:"regex" toml:"regex" xml:"regex" yaml:"regex"`
	Skip      string `json:"skip" toml:"skip" xml:"skip" yaml:"skip"`
	Poll      bool   `json:"poll" toml:"poll" xml:"poll" yaml:"poll"`
	Pipe      bool   `json:"pipe" toml:"pipe" xml:"pipe" yaml:"pipe"`
	MustExist bool   `json:"mustExist" toml:"must_exist" xml:"must_exist" yaml:"mustExist"`
	LogMatch  bool   `json:"logMatch" toml:"log_match" xml:"log_match" yaml:"logMatch"`
	re        *regexp.Regexp
	skip      *regexp.Regexp
	tail      *tail.Tail
	mu        sync.RWMutex
	retries   uint
}

// Match is what we send to the website.
type Match struct {
	File    string   `json:"file"`
	Matches []string `json:"matches"`
	Line    string   `json:"line"`
}

// runFileWatcher compiles any regexp's and opens a tail -f on provided watch files.
func (c *Config) runFileWatcher() {
	// two fake tails for internal channels.
	validTails := []*WatchFile{{Path: "/add watcher channel/"}, {Path: "/retry ticker/"}}

	for _, item := range c.WatchFiles {
		if err := item.setup(c.Logger.GetInfoLog()); err != nil {
			c.Errorf("Unable to watch file %v", err)
			continue
		}

		validTails = append(validTails, item)
	}

	if len(validTails) != 0 {
		cases, ticker := c.collectFileTails(validTails)
		go c.tailFiles(cases, validTails, ticker)
	}
}

func (w *WatchFile) setup(logger *log.Logger) error {
	var err error

	w.retries = maxRetries // so it will not get "restarted" unless it passes validation.

	if w.Regexp == "" {
		return fmt.Errorf("%w: no regexp match provided, ignored: %s", ErrInvalidRegexp, w.Path)
	} else if w.re, err = regexp.Compile(w.Regexp); err != nil {
		return fmt.Errorf("%w: regexp match compile failed, ignored: %s", ErrInvalidRegexp, w.Path)
	} else if w.skip, err = regexp.Compile(w.Skip); err != nil {
		return fmt.Errorf("%w: regexp skip compile failed, ignored: %s", ErrInvalidRegexp, w.Path)
	}

	w.tail, err = tail.TailFile(w.Path, tail.Config{
		Follow:        true,
		ReOpen:        true,
		MustExist:     w.MustExist,
		Poll:          w.Poll,
		Pipe:          w.Pipe,
		CompleteLines: true,
		Location:      &tail.SeekInfo{Whence: io.SeekEnd},
		Logger:        logger,
	})
	if err != nil {
		exp.FileWatcher.Add(w.Path+" Errors", 1)
		return fmt.Errorf("watching file %s: %w", w.Path, err)
	}

	w.retries = 0

	return nil
}

// collectFileTails uses reflection to watch a dynamic list of files in one go routine.
func (c *Config) collectFileTails(tails []*WatchFile) ([]reflect.SelectCase, *time.Ticker) {
	c.extras.addWatcher = make(chan *WatchFile, 1)
	ticker := time.NewTicker(retryInterval)
	cases := make([]reflect.SelectCase, len(tails))

	for idx, item := range tails {
		if idx == 0 { // 0 is skipped (see above), and used as an internal I/O channel
			cases[idx] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(c.addWatcher)}
			continue
		} else if idx == 1 {
			cases[idx] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ticker.C)}
			continue
		}

		cases[idx] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(item.tail.Lines)}

		c.Printf("==> Watching: %s, regexp: '%s' skip: '%s' poll:%v pipe:%v must:%v log:%v",
			item.Path, item.Regexp, item.Skip, item.Poll, item.Pipe, item.MustExist, item.LogMatch)

		if exp.FileWatcher.Get(item.Path+" Matched") == nil {
			// so it shows up on the Metrics page if no lines have been read.
			exp.FileWatcher.Add(item.Path+" Matched", 0)
		}
	}

	return cases, ticker
}

//nolint:cyclop
func (c *Config) tailFiles(cases []reflect.SelectCase, tails []*WatchFile, ticker *time.Ticker) {
	defer c.Printf("==> All file watchers stopped.")
	defer ticker.Stop()

	var died bool

	for {
		idx, data, running := reflect.Select(cases)
		item := tails[idx]

		switch {
		case !running && idx == 0:
			return // main channel closed, bail out.
		case !running:
			tails = append(tails[:idx], tails[idx+1:]...) // The channel was closed? okay, remove it.
			cases = append(cases[:idx], cases[idx+1:]...)

			if err := item.deactivate(); err != nil {
				c.Errorf("No longer watching file (channel closed): %s: %v", item.Path, err)
				exp.FileWatcher.Add(item.Path+" Errors", 1)
				died = true //nolint:wsl
			} else {
				c.Printf("==> No longer watching file (channel closed): %s", item.Path)
			}

			if len(cases) < 1 {
				return
			}
		case idx == 1:
			died = c.fileWatcherTicker(died)
		case data.IsNil(), data.IsZero(), !data.Elem().CanInterface():
			c.Errorf("Got non-addressable file watcher data from %s", item.Path)
			exp.FileWatcher.Add(item.Path+" Errors", 1)
		case idx == 0:
			item, _ = data.Elem().Addr().Interface().(*WatchFile)
			tails = append(tails, item)
			cases = append(cases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(item.tail.Lines)})
		default:
			exp.FileWatcher.Add(item.Path+" Lines", 1)

			line, _ := data.Elem().Addr().Interface().(*tail.Line)
			c.checkLineMatch(line, item)
			exp.FileWatcher.Add(item.Path+" Bytes", int64(len(line.Text)))
		}
	}
}

func (c *Config) fileWatcherTicker(died bool) bool {
	if c.Logger.DebugEnabled() { // remove this.
		c.Debugf("File Watcher Ticker. Dead files: %v", died)
	}

	if !died {
		return false
	}

	var stilldead bool

	for _, item := range c.WatchFiles {
		if item.Active() || item.retries >= maxRetries {
			continue
		}

		item.retries++
		exp.FileWatcher.Add(item.Path+" Retries", 1)

		// move this back to debug.
		c.Printf("Restarting File Watcher (retries: %d): %s", item.retries, item.Path)

		if err := c.AddFileWatcher(item); err != nil {
			c.Errorf("Restarting File Watcher (retries: %d): %s: %v", item.retries, item.Path, err)
			exp.FileWatcher.Add(item.Path+" Errors", 1)

			stilldead = true
		} else {
			item.retries = 0
			exp.FileWatcher.Add(item.Path+" Restarts", 1)
		}
	}

	return stilldead
}

// checkLineMatch runs when a watched file has a new line written.
// If a match is found a notification is sent.
func (c *Config) checkLineMatch(line *tail.Line, tail *WatchFile) {
	if tail.re == nil || line.Text == "" || !tail.re.MatchString(line.Text) {
		return // no match
	}

	if tail.skip != nil && tail.Skip != "" && tail.skip.MatchString(line.Text) {
		exp.FileWatcher.Add(tail.Path+" Skipped", 1)
		return // skip matches
	}

	exp.FileWatcher.Add(tail.Path+" Matched", 1)

	match := &Match{
		File:    tail.Path,
		Line:    strings.TrimSpace(line.Text),
		Matches: tail.re.FindAllString(line.Text, -1),
	}

	// this can be removed before release.
	c.Debugf("[%s requested] Sending Watched-File Line Match to Notifiarr: %s: %s",
		EventFile, tail.Path, match.Line)
	c.QueueData(&SendRequest{
		Route:      LogLineRoute,
		Event:      EventFile,
		LogPayload: true,
		LogMsg:     fmt.Sprintf("Watched-File Line Match: %s: %s", tail.Path, match.Line),
		Payload:    match,
	})
}

func (c *Config) AddFileWatcher(file *WatchFile) error {
	if err := file.setup(c.Logger.GetInfoLog()); err != nil {
		return err
	}

	c.Printf("Watching File: %s, regexp: '%s' skip: '%s' poll:%v pipe:%v must:%v log:%v",
		file.Path, file.Regexp, file.Skip, file.Poll, file.Pipe, file.MustExist, file.LogMatch)

	c.extras.addWatcher <- file

	return nil
}

func (c *Config) StopFileWatcher(file *WatchFile) error {
	if file.Active() {
		file.retries = maxRetries // so it will not get "restarted" after manually being stopped.

		// move this back to debug.
		c.Printf("Stopping File Watcher: %s", file.Path)

		if err := file.stop(); err != nil {
			c.Errorf("Stopping File Watcher: %s: %v", file.Path, err)
			return err
		}
	}

	return nil
}

func (c *Config) stopFileWatchers() {
	defer close(c.extras.addWatcher)
	c.extras.addWatcher = nil

	for _, tail := range c.WatchFiles {
		_ = c.StopFileWatcher(tail)
	}

	// The following code waits for all the watchers to die before returning.
	repeat := func() bool {
		for _, tail := range c.WatchFiles {
			if tail.Active() {
				return true
			}
		}

		return false
	}

	for repeat() {
		time.Sleep(70 * time.Millisecond) // nolint:gomnd
	}
}

// this runs when a channel dies from the main go routine loop.
func (w *WatchFile) deactivate() error {
	defer func() {
		w.mu.Lock()
		defer w.mu.Unlock()

		w.tail = nil
	}()

	return w.stop()
}

// Active returns true if the tail channel is still open.
func (w *WatchFile) Active() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return w.tail != nil
}

// stop stops a file watcher.
func (w *WatchFile) stop() error {
	w.mu.RLock()
	defer w.mu.RUnlock()
	// defer w.tail.Cleanup()

	if err := w.tail.Stop(); err != nil {
		return fmt.Errorf("stopping watcher: %w", err)
	}

	return nil
}
