package filewatch

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
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"github.com/nxadm/tail"
)

var ErrInvalidRegexp = fmt.Errorf("invalid regexp")

const (
	maxRetries    = 6
	retryInterval = 10 * time.Second
)

type cmd struct {
	*common.Config
	addWatcher  chan *WatchFile
	stopWatcher chan struct{}
	awMutex     sync.RWMutex
	files       []*WatchFile
}

// Action contains the exported methods for this package.
type Action struct {
	cmd *cmd
}

// WatchFile is the input data needed to watch files.
type WatchFile struct {
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

// New configures the library.
func New(config *common.Config, files []*WatchFile) *Action {
	return &Action{
		cmd: &cmd{
			Config: config,
			files:  files,
		},
	}
}

// Run compiles any regexp's and opens a tail -f on provided watch files.
func (a *Action) Run() {
	a.cmd.run()
}

// Files returns the list of files configured.
func (a *Action) Files() []*WatchFile {
	return a.cmd.files
}

// Stop all file watcher routines.
func (a *Action) Stop() {
	a.cmd.stop()
}

func (c *cmd) run() {
	// two fake tails for internal channels.
	validTails := []*WatchFile{{Path: "/add watcher channel/"}, {Path: "/retry ticker/"}}

	for _, item := range c.files {
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
func (c *cmd) collectFileTails(tails []*WatchFile) ([]reflect.SelectCase, *time.Ticker) {
	c.addWatcher = make(chan *WatchFile, 1)
	c.stopWatcher = make(chan struct{})
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
func (c *cmd) tailFiles(cases []reflect.SelectCase, tails []*WatchFile, ticker *time.Ticker) {
	limitChan, limitTicker := getLimiter()
	defer c.closeWatcher(limitChan, limitTicker, ticker)

	var died bool

	for {
		<-limitChan

		idx, data, running := reflect.Select(cases)
		item := tails[idx]

		switch {
		case !running && idx == 0:
			return // main channel closed, bail out.
		case !running:
			tails = append(tails[:idx], tails[idx+1:]...) // The channel was closed? okay, remove it.
			cases = append(cases[:idx], cases[idx+1:]...)

			// This routines runs the Stop method on the tail.
			// If that returns an error, it means it died.
			// If that does not return an error, it means Stop was already called.
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

// closeWatcher is defered when the watcher dies.
func (c *cmd) closeWatcher(limiter chan time.Time, tickers ...*time.Ticker) {
	defer c.CapturePanic()

	for _, ticker := range tickers {
		ticker.Stop()
	}

	c.Printf("==> All file watchers stopped.")
	close(limiter)
	close(c.stopWatcher) // signal we're done.
}

// getLimiter returns a channel that limits requests to 1 per second, with a burst rate of 10.
func getLimiter() (chan time.Time, *time.Ticker) {
	const (
		burstRate  = 10
		requestPer = time.Second
	)

	limiter := time.NewTicker(requestPer)

	burstyLimiter := make(chan time.Time, burstRate)
	for i := 0; i < 3; i++ {
		burstyLimiter <- time.Now()
	}

	go func() {
		for t := range limiter.C {
			burstyLimiter <- t
		}
	}()

	return burstyLimiter, limiter
}

func (c *cmd) fileWatcherTicker(died bool) bool {
	if !died {
		return false
	}

	var stilldead bool

	for _, item := range c.files {
		if item.Active() || item.retries >= maxRetries {
			continue
		}

		item.retries++
		exp.FileWatcher.Add(item.Path+" Retries", 1)

		// move this back to debug.
		c.Printf("Restarting File Watcher (retries: %d): %s", item.retries, item.Path)

		if err := c.addFileWatcher(item); err != nil {
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
func (c *cmd) checkLineMatch(line *tail.Line, tail *WatchFile) {
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

	c.SendData(&website.Request{
		Route:      website.LogLineRoute,
		Event:      website.EventFile,
		LogPayload: tail.LogMatch,
		LogMsg:     fmt.Sprintf("Watched-File Line Match: %s: %s", tail.Path, match.Line),
		Payload:    match,
	})
}

func (a *Action) AddFileWatcher(file *WatchFile) error {
	return a.cmd.addFileWatcher(file)
}

func (c *cmd) addFileWatcher(file *WatchFile) error {
	c.awMutex.RLock()
	defer c.awMutex.RUnlock()

	if c.addWatcher == nil {
		return common.ErrNoChannel
	}

	if err := file.setup(c.Logger.GetInfoLog()); err != nil {
		return err
	}

	c.Printf("Watching File: %s, regexp: '%s' skip: '%s' poll:%v pipe:%v must:%v log:%v",
		file.Path, file.Regexp, file.Skip, file.Poll, file.Pipe, file.MustExist, file.LogMatch)

	c.addWatcher <- file

	return nil
}

func (w *WatchFile) Stop() error {
	if !w.Active() {
		return nil
	}

	w.retries = maxRetries // so it will not get "restarted" after manually being stopped.

	if err := w.stop(); err != nil {
		return err
	}

	return nil
}

func (c *cmd) stop() {
	c.awMutex.Lock()
	defer c.awMutex.Unlock()

	for _, tail := range c.files {
		if err := tail.Stop(); err != nil {
			c.Errorf("Stopping File Watcher: %s: %v", tail.Path, err)
		}
	}

	// The following code might wait for all the watchers to die before returning.
	close(c.addWatcher)
	<-c.stopWatcher
	c.addWatcher = nil
	c.stopWatcher = nil
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

	if err := w.tail.Stop(); err != nil {
		return fmt.Errorf("stop failed: %w", err)
	}

	return nil
}
