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
	var (
		err        error
		validTails = []*WatchFile{}
	)

	for _, item := range c.WatchFiles {
		if item.Regexp == "" {
			c.Errorf("Watch File has no regexp match. Ignored: %s", item.Path)
			continue
		} else if item.re, err = regexp.Compile(item.Regexp); err != nil {
			c.Errorf("Regexp compile failed, not watching file %s: %v", item.Path, err)
			continue
		}

		if item.skip, err = regexp.Compile(item.Skip); err != nil {
			c.Errorf("Skip Regexp compile failed for %s: %v", item.Path, err)
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

		c.Printf("==> Watching: %s (regexp: '%s' skip: '%s')", item.Path, item.Regexp, item.Skip)
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
	if tail.re == nil || !tail.re.MatchString(line.Text) {
		return // no match
	}

	if tail.skip != nil && tail.Skip != "" && tail.skip.MatchString(line.Text) {
		return // skip matches
	}

	match := &Match{
		File:    tail.Path,
		Line:    strings.TrimSpace(line.Text),
		Matches: tail.re.FindAllString(line.Text, 0),
	}

	if resp, err := c.SendData(LogLineRoute.Path(EventFile), match, true); err != nil {
		c.Errorf("[%s requested] Sending Watched-File Line Match to Notifiarr: %s: %s => %s",
			EventFile, tail.Path, line.Text, err)
	} else if tail.LogMatch {
		c.Printf("[%s requested] Sent Watched-File Line Match to Notifiarr: %s: %s => %s",
			EventFile, tail.Path, line.Text, resp)
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
