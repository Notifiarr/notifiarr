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
	Regexp    string `json:"regexp" toml:"regexp" xml:"regexp" yaml:"regexp"`
	Poll      bool   `json:"poll" toml:"poll" xml:"poll" yaml:"poll"`
	Pipe      bool   `json:"pipe" toml:"pipe" xml:"pipe" yaml:"pipe"`
	MustExist bool   `json:"mustExist" toml:"must_exist" xml:"must_exist" yaml:"mustExist"`
	re        *regexp.Regexp
	tail      *tail.Tail
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
			c.Errorf("Watch FIle has no regexp or string match. Ignored: %s", item.Path)
			continue
		}

		item.tail, err = tail.TailFile(item.Path, tail.Config{
			Follow:    true,
			ReOpen:    true,
			MustExist: item.MustExist,
			Poll:      item.Poll,
			Pipe:      item.Pipe,
			Location:  &tail.SeekInfo{Whence: io.SeekEnd},
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

		c.Printf("==> Watching: %s (regexp: %s match: %s)", item.Path, item.Regexp, item.Match)
	}

	for {
		idx, data, ok := reflect.Select(cases)
		if ok {
			if data.Elem().CanInterface() {
				c.checkMatch(data.Elem().Interface().(tail.Line), tails[idx])
			} else {
				c.Errorf("Got non-addressable file watcher data from %s", tails[idx].Path)
			}

			continue
		}

		c.Errorf("No longer watching file (channel closed): %s", tails[idx].Path)
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
	switch {
	case tail.Match != "" && strings.Contains(line.Text, tail.Match):
		c.Printf("Found Match in %s: %s", tail.Path, line.Text)
	case tail.re != nil && tail.re.MatchString(line.Text):
		c.Printf("Found Regexp in %s: %s", tail.Path, line.Text)
	}
}
