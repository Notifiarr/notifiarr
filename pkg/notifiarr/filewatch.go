package notifiarr

import (
	"io"
	"reflect"
	"regexp"
	"strings"

	"github.com/nxadm/tail"
)

type WatchFile struct {
	//	Cooldown  cnfg.Duration `json:"cooldown" toml:"cooldown" xml:"cooldown" yaml:"cooldown"`
	Path      string `json:"path" toml:"path" xml:"path" yaml:"path"`
	Match     string `json:"match" toml:"match" xml:"match" yaml:"match"`
	Regexp    string `json:"regex" toml:"regex" xml:"regex" yaml:"regex"`
	Poll      bool   `json:"poll" toml:"poll" xml:"poll" yaml:"poll"`
	Pipe      bool   `json:"pipe" toml:"pipe" xml:"pipe" yaml:"pipe"`
	MustExist bool   `json:"mustExist" toml:"must_exist" xml:"must_exist" yaml:"mustExist"`
	LogMatch  bool   `json:"logMatch" toml:"log_match" xml:"log_match" yaml:"logMatch"`
	re        *regexp.Regexp
	tail      *tail.Tail
}

// Match is what we send to the website.
type Match struct {
	File    string `json:"file"`
	Matches string `json:"matches"`
	Line    string `json:"line"`
}

// runFileWatcher compiles any regexp's and opens a tail -f on provided watch files.
func (c *Config) runFileWatcher() {
	var (
		err        error
		validTails = []*WatchFile{}
	)

	for _, item := range c.WatchFiles {
		if item.Regexp != "" {
			item.re, err = regexp.Compile(item.Regexp)
			if err != nil {
				if item.Match == "" {
					c.Errorf("Regexp compile failed, not watching file %s: %v", item.Path, err)
					continue
				}

				c.Errorf("Regexp compile failed, watching file with match-only %s: %v", item.Path, err)
			}
		} else if item.Match == "" {
			c.Errorf("Watch File has no regexp or string match. Ignored: %s", item.Path)
			continue
		}

		item.tail, err = tail.TailFile(item.Path, tail.Config{
			Follow:    true,
			ReOpen:    true,
			MustExist: item.MustExist,
			Poll:      item.Poll,
			Pipe:      item.Pipe,
			Location:  &tail.SeekInfo{Whence: io.SeekEnd},
			Logger:    c.client.Logger,
		})
		if err != nil {
			c.Errorf("Unable to watch file %s: %v", item.Path, err)
			continue
		}

		validTails = append(validTails, item)
	}

	if len(validTails) != 0 {
		go c.collectFileTails(validTails)
	}
}

// collectFileTails uses reflection to watch a dynamic list of files in one go routine.
func (c *Config) collectFileTails(tails []*WatchFile) {
	cases := make([]reflect.SelectCase, len(tails))

	for idx, item := range tails {
		cases[idx] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(item.tail.Lines)}

		c.Printf("==> Watching: %s (regexp: '%s' match: '%s')", item.Path, item.Regexp, item.Match)
	}

	defer c.Printf("==> All file watchers stopped.")

	for {
		idx, data, ok := reflect.Select(cases)
		if ok {
			if data.Elem().CanInterface() {
				line, _ := data.Elem().Interface().(tail.Line)
				c.checkMatch(line, tails[idx])
			} else {
				c.Errorf("Got non-addressable file watcher data from %s", tails[idx].Path)
			}

			continue
		}

		c.Printf("==> No longer watching file (channel closed): %s", tails[idx].Path)
		// The channel was closed? okay, remove it.
		tails = append(tails[:idx], tails[idx+1:]...)

		cases = append(cases[:idx], cases[idx+1:]...)
		if len(cases) < 1 {
			return
		}
	}
}

// checkMatch runs when a watched file has a new line written.
// If a match is found a notification is sent.
func (c *Config) checkMatch(line tail.Line, tail *WatchFile) {
	var (
		match = &Match{File: tail.Path}
		event EventType
	)

	switch {
	default:
		return
	case tail.Match != "" && strings.Contains(line.Text, tail.Match):
		event = EventFileMa
		match.Line = strings.TrimSpace(line.Text)
		match.Matches = tail.Match
	case tail.re != nil && tail.re.MatchString(line.Text):
		match.Matches = tail.Regexp
		match.Line = strings.TrimSpace(line.Text)
		event = EventFileRe
	}

	if resp, err := c.SendData(LogLineRoute.Path(event), match, true); err != nil {
		c.Errorf("[%s requested] Sending Watched-File Line Match to Notifiarr: %s: %s => %s",
			event, tail.Path, line.Text, err)
	} else if tail.LogMatch {
		c.Printf("[%s requested] Sent Watched-File Line Match to Notifiarr: %s: %s => %s",
			event, tail.Path, line.Text, resp)
	}
}

func (c *Config) stopFileWatcher() {
	for _, tail := range c.WatchFiles {
		if tail.tail != nil {
			if err := tail.tail.Stop(); err != nil {
				c.Errorf("Stopping File Watcher: %s: %v", tail.Path, err)
			}
		}
	}
}
