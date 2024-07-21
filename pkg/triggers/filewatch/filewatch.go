package filewatch

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"github.com/nxadm/tail"
	"github.com/nxadm/tail/ratelimiter"
)

var (
	ErrInvalidRegexp = errors.New("invalid regexp")
	ErrIgnoredLog    = errors.New("the requested path is internally ignored")
)

const (
	Errors        = " Errors"
	Matched       = " Matched"
	maxRetries    = 12                                 // how many times to retry watching a file.
	retryInterval = 10 * time.Second                   // how often channels are checked for being closed.
	specialCase   = 2                                  // We have two special channels in our select cases.
	burstRate     = 6                                  // burst to this many 'matches' before throttling.
	requestPer    = time.Second + 500*time.Millisecond // 1 request per this time period allowed + burst rate.
)

type cmd struct {
	*common.Config
	addWatcher  chan *WatchFile
	stopWatcher chan struct{}
	awMutex     sync.RWMutex
	files       []*WatchFile
	limiter     *ratelimiter.LeakyBucket
	ignored     []string
}

// Action contains the exported methods for this package.
type Action struct {
	cmd *cmd
}

// WatchFile is the input data needed to watch files.
type WatchFile struct {
	Path      string `json:"path"      toml:"path"       xml:"path"       yaml:"path"`
	Regexp    string `json:"regex"     toml:"regex"      xml:"regex"      yaml:"regex"`
	Skip      string `json:"skip"      toml:"skip"       xml:"skip"       yaml:"skip"`
	Poll      bool   `json:"poll"      toml:"poll"       xml:"poll"       yaml:"poll"`
	Pipe      bool   `json:"pipe"      toml:"pipe"       xml:"pipe"       yaml:"pipe"`
	MustExist bool   `json:"mustExist" toml:"must_exist" xml:"must_exist" yaml:"mustExist"`
	LogMatch  bool   `json:"logMatch"  toml:"log_match"  xml:"log_match"  yaml:"logMatch"`
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
func New(config *common.Config, files []*WatchFile, ignored []string) *Action {
	return &Action{
		cmd: &cmd{
			Config:  config,
			files:   files,
			limiter: ratelimiter.NewLeakyBucket(burstRate, requestPer),
			ignored: checkIgnored(ignored),
		},
	}
}

// We ignore our own log files.
func (i ignored) isIgnored(filePath string) bool {
	if abs, err := filepath.Abs(filePath); err == nil {
		filePath = abs
	}

	for _, name := range i {
		if filePath == name {
			return true
		}
	}

	return false
}

type ignored []string

// We ignore our own log files.
func checkIgnored(ignored []string) ignored {
	output := []string{}

	for _, item := range ignored {
		if item == "" {
			continue
		}

		if abs, err := filepath.Abs(item); err == nil {
			item = abs // err is swallowed
		}

		output = append(output, item)
	}

	return output
}

// Verify the interfaces are satisfied.
var _ = common.Run(&Action{nil})

// Run compiles any regexp's and opens a tail -f on provided watch files.
func (a *Action) Run(_ context.Context) {
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
		if err := item.setup(&logger{Logger: c.Config.Logger}, c.ignored); err != nil {
			c.Errorf("Unable to watch file: %v", err)
			continue
		}

		validTails = append(validTails, item)
	}

	if len(validTails) == 0 {
		return
	}

	cases, ticker := c.collectFileTails(validTails)
	c.tailFiles(cases, validTails, ticker)
}

func (w *WatchFile) setup(logger *logger, ignored ignored) error {
	var err error

	w.retries = maxRetries // so it will not get "restarted" unless it passes validation.

	if w.Regexp == "" {
		return fmt.Errorf("%w: no regexp match provided, ignored: %s", ErrInvalidRegexp, w.Path)
	} else if w.re, err = regexp.Compile(w.Regexp); err != nil {
		return fmt.Errorf("%w: regexp match compile failed, ignored: %s", ErrInvalidRegexp, w.Path)
	} else if w.skip, err = regexp.Compile(w.Skip); err != nil {
		return fmt.Errorf("%w: regexp skip compile failed, ignored: %s", ErrInvalidRegexp, w.Path)
	} else if ignored.isIgnored(w.Path) {
		return fmt.Errorf("%w: %s", ErrIgnoredLog, w.Path)
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
		mnd.FileWatcher.Add(w.Path+Errors, 1)
		return fmt.Errorf("watching file %s: %w", w.Path, err)
	}

	w.retries = 0

	return nil
}

// collectFileTails uses reflection to watch a dynamic list of files in one go routine.
func (c *cmd) collectFileTails(tails []*WatchFile) ([]reflect.SelectCase, *time.Ticker) {
	c.addWatcher = make(chan *WatchFile, len(tails)+1)
	c.stopWatcher = make(chan struct{})
	ticker := time.NewTicker(retryInterval)
	cases := make([]reflect.SelectCase, len(tails))

	for idx, item := range tails {
		// If you add more special cases here, increment specialCase.
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

		if mnd.FileWatcher.Get(item.Path+Matched) == nil {
			// so it shows up on the Metrics page if no lines have been read.
			mnd.FileWatcher.Add(item.Path+Matched, 0)
		}
	}

	return cases, ticker
}

func (c *cmd) tailFiles(cases []reflect.SelectCase, tails []*WatchFile, ticker *time.Ticker) {
	defer func() {
		defer c.CapturePanic()
		ticker.Stop()
		c.Printf("==> All file watchers stopped.")
		close(c.stopWatcher) // signal we're done.
	}()

	var died bool

	for {
		idx, data, running := reflect.Select(cases)
		item := tails[idx]

		switch {
		case !running && idx == 0:
			if len(cases) <= specialCase {
				return // all channels are now closed, bail out.
			}
		case !running:
			tails = append(tails[:idx], tails[idx+1:]...) // The channel was closed? okay, remove it.
			cases = append(cases[:idx], cases[idx+1:]...)
			died = c.killWatcher(item)
		case idx == 1:
			died = c.fileWatcherTicker(died)
		case data.IsNil(), data.IsZero(), !data.Elem().CanInterface():
			c.Errorf("Got non-addressable file watcher data from %s", item.Path)
			mnd.FileWatcher.Add(item.Path+Errors, 1)
		case idx == 0:
			item, _ = data.Elem().Addr().Interface().(*WatchFile)
			tails = append(tails, item)
			cases = append(cases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(item.tail.Lines)})
		default:
			mnd.FileWatcher.Add(item.Path+" Lines", 1)

			line, _ := data.Elem().Addr().Interface().(*tail.Line)
			c.checkLineMatch(line, item)
			mnd.FileWatcher.Add(item.Path+" Bytes", int64(len(line.Text)))
		}
	}
}

// killWatcher runs the Stop method on the tail.
// If that returns an error, it means it died.
// If that does not return an error, it means Stop was already called.
func (c *cmd) killWatcher(item *WatchFile) bool {
	if err := item.deactivate(); err != nil {
		c.Errorf("No longer watching file (channel closed): %s: %v", item.Path, err)
		mnd.FileWatcher.Add(item.Path+Errors, 1)

		return true
	}

	c.Printf("==> No longer watching file (channel closed): %s", item.Path)

	return false
}

// fileWatcherTicker checks if a file watcher died and needs to be restarted.
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
		retries := item.retries
		mnd.FileWatcher.Add(item.Path+" Retries", 1)

		// move this back to debug.
		c.Printf("Restarting File Watcher (retries: %d/%d): %s", item.retries, maxRetries, item.Path)

		if err := c.addFileWatcher(item); err != nil {
			c.Errorf("Restarting File Watcher (retries: %d/%d): %s: %v", item.retries, maxRetries, item.Path, err)
			mnd.FileWatcher.Add(item.Path+Errors, 1)

			stilldead = true
		} else {
			mnd.FileWatcher.Add(item.Path+" Restarts", 1)
		}

		// Calling addFileWatcher will reset the retries to 0, so undo that.
		item.retries = retries
	}

	return stilldead
}

// checkLineMatch runs when a watched file has a new line written.
// If a match is found a notification is sent.
func (c *cmd) checkLineMatch(line *tail.Line, tail *WatchFile) {
	tail.retries = 0 // reset retries once we get a line from the file.

	if tail.re == nil || line.Text == "" || !tail.re.MatchString(line.Text) {
		return // no match
	}

	if tail.skip != nil && tail.Skip != "" && tail.skip.MatchString(line.Text) {
		mnd.FileWatcher.Add(tail.Path+" Skipped", 1)
		return // skip matches
	}

	mnd.FileWatcher.Add(tail.Path+Matched, 1)

	match := &Match{
		File:    tail.Path,
		Line:    strings.TrimSpace(line.Text),
		Matches: tail.re.FindAllString(line.Text, -1),
	}

	if !c.limiter.Pour(1) {
		mnd.FileWatcher.Add(tail.Path+" Dropped", 1)
		return // rate limited.
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

	err := file.setup(&logger{Logger: c.Config.Logger}, c.ignored)
	if err != nil {
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

	return w.stop()
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
