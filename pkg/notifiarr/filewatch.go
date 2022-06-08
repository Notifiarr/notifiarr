package notifiarr

import (
	"fmt"
	"io"
	"log"
	"reflect"
	"regexp"
	"strings"

	"github.com/Notifiarr/notifiarr/pkg/exp"
	"github.com/nxadm/tail"
)

var ErrInvalidRegexp = fmt.Errorf("invalid regexp")

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
}

// Match is what we send to the website.
type Match struct {
	File    string   `json:"file"`
	Matches []string `json:"matches"`
	Line    string   `json:"line"`
}

// runFileWatcher compiles any regexp's and opens a tail -f on provided watch files.
func (c *Config) runFileWatcher() {
	validTails := []*WatchFile{{Path: "/add watcher channel/"}}

	for _, item := range c.WatchFiles {
		if err := item.Setup(c.Logger.InfoLog); err != nil {
			c.Errorf("Unable to watch file %v", err)
			continue
		}

		validTails = append(validTails, item)
	}

	if len(validTails) != 0 {
		go c.collectFileTails(validTails)
	}
}

func (w *WatchFile) Setup(logger *log.Logger) error {
	var err error

	if w.Regexp == "" {
		return fmt.Errorf("%w: no regexp match provided, ignored: %s", ErrInvalidRegexp, w.Path)
	} else if w.re, err = regexp.Compile(w.Regexp); err != nil {
		return fmt.Errorf("%w: regexp match compile failed, ignored: %s", ErrInvalidRegexp, w.Path)
	} else if w.skip, err = regexp.Compile(w.Skip); err != nil {
		return fmt.Errorf("%w: regexp skip compile failed, ignored: %s", ErrInvalidRegexp, w.Path)
	}

	w.tail, err = tail.TailFile(w.Path, tail.Config{
		Follow:    true,
		ReOpen:    true,
		MustExist: w.MustExist,
		Poll:      w.Poll,
		Pipe:      w.Pipe,
		Location:  &tail.SeekInfo{Whence: io.SeekEnd},
		Logger:    logger,
	})
	if err != nil {
		exp.FileWatcher.Add(w.Path+" Errors", 1)
		return fmt.Errorf("watching file %s: %w", w.Path, err)
	}

	return nil
}

// collectFileTails uses reflection to watch a dynamic list of files in one go routine.
func (c *Config) collectFileTails(tails []*WatchFile) {
	c.extras.addWatcher = make(chan *WatchFile)
	cases := make([]reflect.SelectCase, len(tails))

	for idx, item := range tails {
		if idx == 0 { // 0 is skipped (see above), and used as an internal I/O channel
			cases[idx] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(c.addWatcher)}
			continue
		} else {
			cases[idx] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(item.tail.Lines)}
		}

		c.Printf("==> Watching: %s, regexp: '%s' skip: '%s' poll:%v pipe:%v must:%v log:%v",
			item.Path, item.Regexp, item.Skip, item.Poll, item.Pipe, item.MustExist, item.LogMatch)

		if exp.FileWatcher.Get(item.Path+" Matched") == nil {
			// so it shows up on the Metrics page if no lines have been read.
			exp.FileWatcher.Add(item.Path+" Matched", 0)
		}
	}

	defer c.Printf("==> All file watchers stopped.")

	for {
		switch idx, data, ok := reflect.Select(cases); {
		case !ok:
			tails[idx].tail = nil // so we do not try to Stop() it.

			c.Printf("==> No longer watching file (channel closed): %s", tails[idx].Path)
			// The channel was closed? okay, remove it.
			tails = append(tails[:idx], tails[idx+1:]...)

			cases = append(cases[:idx], cases[idx+1:]...)
			if len(cases) < 1 || idx == 0 {
				return
			}
		case !data.Elem().CanInterface():
			c.Errorf("Got non-addressable file watcher data from %s", tails[idx].Path)
			exp.FileWatcher.Add(tails[idx].Path+" Errors", 1)
		case idx == 0:
			item, _ := data.Elem().Interface().(*WatchFile)
			tails = append(tails, item)
			cases = append(cases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(item.tail.Lines)})
		default:
			line, _ := data.Elem().Interface().(tail.Line)
			c.checkLineMatch(line, tails[idx])
			exp.FileWatcher.Add(tails[idx].Path+" Lines", 1)
			exp.FileWatcher.Add(tails[idx].Path+" Bytes", int64(len(line.Text)+1))
		}
	}
}

// checkLineMatch runs when a watched file has a new line written.
// If a match is found a notification is sent.
func (c *Config) checkLineMatch(line tail.Line, tail *WatchFile) {
	if tail.re == nil || !tail.re.MatchString(line.Text) {
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

	if resp, err := c.SendData(LogLineRoute.Path(EventFile), match, true); err != nil {
		c.Errorf("[%s requested] Sending Watched-File Line Match to Notifiarr: %s: %s => %s",
			EventFile, tail.Path, line.Text, err)
		exp.FileWatcher.Add(tail.Path+" Error", 1)
	} else if tail.LogMatch {
		c.Printf("[%s requested] Sent Watched-File Line Match to Notifiarr: %s: %s => %s",
			EventFile, tail.Path, line.Text, resp)
	}
}

func (c *Config) AddFileWatcher(file *WatchFile) error {
	if err := file.Setup(c.Logger.InfoLog); err != nil {
		return err
	}

	c.Printf("Watching File: %s, regexp: '%s' skip: '%s' poll:%v pipe:%v must:%v log:%v",
		file.Path, file.Regexp, file.Skip, file.Poll, file.Pipe, file.MustExist, file.LogMatch)

	c.extras.addWatcher <- file

	return nil
}

func (c *Config) StopFileWatcher(file *WatchFile) {
	if file.tail != nil {
		c.Debugf("Stopping File Watcher: %s", file.Path)

		if err := file.tail.Stop(); err != nil {
			c.Errorf("Stopping File Watcher: %s: %v", file.Path, err)
		}
	}
}

func (c *Config) stopFileWatchers() {
	defer close(c.extras.addWatcher)
	c.extras.addWatcher = nil

	for _, tail := range c.WatchFiles {
		c.StopFileWatcher(tail)
	}
}

// Active returns true if the tail channel is still open.
func (w *WatchFile) Active() bool {
	return w.tail != nil
}
