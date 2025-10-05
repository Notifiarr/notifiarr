package checkapp

import (
	"context"
	"sync"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
)

type CheckAllInput struct {
	Sonarr       []apps.StarrConfig    `json:"sonarr"`
	Radarr       []apps.StarrConfig    `json:"radarr"`
	Readarr      []apps.StarrConfig    `json:"readarr"`
	Lidarr       []apps.StarrConfig    `json:"lidarr"`
	Prowlarr     []apps.StarrConfig    `json:"prowlarr"`
	Plex         []apps.PlexConfig     `json:"plex"`
	Tautulli     []apps.TautulliConfig `json:"tautulli"`
	NZBGet       []apps.NZBGetConfig   `json:"nzbget"`
	Deluge       []apps.DelugeConfig   `json:"deluge"`
	Qbit         []apps.QbitConfig     `json:"qbit"`
	Rtorrent     []apps.RtorrentConfig `json:"rtorrent"`
	Transmission []apps.XmissionConfig `json:"transmission"`
	SabNZB       []apps.SabNZBConfig   `json:"sabnzb"`
}

// TestResult is the result from an instance test.
type TestResult struct {
	Status int    `json:"status"`
	Msg    string `json:"message"`
}

// CheckAllOutput is the output from a check all instances test.
// The JSON keys are used for human display, so ya.
//
//nolint:tagliatelle
type CheckAllOutput struct {
	Sonarr   []TestResult `json:"Sonarr"`
	Radarr   []TestResult `json:"Radarr"`
	Readarr  []TestResult `json:"Readarr"`
	Lidarr   []TestResult `json:"Lidarr"`
	Prowlarr []TestResult `json:"Prowlarr"`
	Plex     []TestResult `json:"Plex"`
	Tautulli []TestResult `json:"Tautulli"`
	NZBGet   []TestResult `json:"NZBGet"`
	Deluge   []TestResult `json:"Deluge"`
	Qbit     []TestResult `json:"Qbittorrent"`
	Rtorrent []TestResult `json:"Rtorrent"`
	Xmiss    []TestResult `json:"Transmission"`
	SabNZB   []TestResult `json:"SabNZB"`
	TimeMS   int64        `json:"timeMS"`
}

type checkAll struct {
	input  *CheckAllInput
	output *CheckAllOutput
	ch     chan *job
}

const (
	timeout = 12 * time.Second
	workers = 5
	buffer  = 100
)

// CheckAll checks all the starr, media and downloader apps and returns the results.
func CheckAll(ctx context.Context, input *CheckAllInput) *CheckAllOutput {
	return (&checkAll{
		input:  input,
		output: initOutput(input),
		ch:     make(chan *job, buffer),
	}).run(ctx)
}

func (c *checkAll) run(ctx context.Context) *CheckAllOutput {
	wait := c.workPool()

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	c.checkAllStarr(ctx)
	c.checkAllMedia(ctx)
	c.checkAllDownloaders(ctx)
	close(c.ch)
	wait()

	return c.output
}

func initOutput(input *CheckAllInput) *CheckAllOutput {
	return &CheckAllOutput{
		Sonarr:   make([]TestResult, len(input.Sonarr)),
		Radarr:   make([]TestResult, len(input.Radarr)),
		Readarr:  make([]TestResult, len(input.Readarr)),
		Lidarr:   make([]TestResult, len(input.Lidarr)),
		Prowlarr: make([]TestResult, len(input.Prowlarr)),
		Plex:     make([]TestResult, len(input.Plex)),
		Tautulli: make([]TestResult, len(input.Tautulli)),
		NZBGet:   make([]TestResult, len(input.NZBGet)),
		Deluge:   make([]TestResult, len(input.Deluge)),
		Qbit:     make([]TestResult, len(input.Qbit)),
		Rtorrent: make([]TestResult, len(input.Rtorrent)),
		Xmiss:    make([]TestResult, len(input.Transmission)),
		SabNZB:   make([]TestResult, len(input.SabNZB)),
		TimeMS:   time.Now().UnixMilli(),
	}
}

func chk[D any](ctx context.Context, i D, fn func(context.Context, D) (string, int)) func() (string, int) {
	return func() (string, int) { return fn(ctx, i) }
}

func (c *checkAll) checkAllStarr(ctx context.Context) {
	for i, sonarr := range c.input.Sonarr {
		c.ch <- &job{res: &c.output.Sonarr[i], fn: chk(ctx, sonarr, Sonarr)}
	}

	for i, radarr := range c.input.Radarr {
		c.ch <- &job{res: &c.output.Radarr[i], fn: chk(ctx, radarr, Radarr)}
	}

	for i, readarr := range c.input.Readarr {
		c.ch <- &job{res: &c.output.Readarr[i], fn: chk(ctx, readarr, Readarr)}
	}

	for i, lidarr := range c.input.Lidarr {
		c.ch <- &job{res: &c.output.Lidarr[i], fn: chk(ctx, lidarr, Lidarr)}
	}

	for i, prowlarr := range c.input.Prowlarr {
		c.ch <- &job{res: &c.output.Prowlarr[i], fn: chk(ctx, prowlarr, Prowlarr)}
	}
}

func (c *checkAll) checkAllMedia(ctx context.Context) {
	for i, plex := range c.input.Plex {
		c.ch <- &job{res: &c.output.Plex[i], fn: chk(ctx, plex, Plex)}
	}

	for i, tautulli := range c.input.Tautulli {
		c.ch <- &job{res: &c.output.Tautulli[i], fn: chk(ctx, tautulli, Tautulli)}
	}
}

func (c *checkAll) checkAllDownloaders(ctx context.Context) {
	for i, nzbget := range c.input.NZBGet {
		c.ch <- &job{res: &c.output.NZBGet[i], fn: chk(ctx, nzbget, NZBGet)}
	}

	for i, deluge := range c.input.Deluge {
		c.ch <- &job{res: &c.output.Deluge[i], fn: chk(ctx, deluge, Deluge)}
	}

	for i, qbit := range c.input.Qbit {
		c.ch <- &job{res: &c.output.Qbit[i], fn: chk(ctx, qbit, Qbit)}
	}

	for i, rtorrent := range c.input.Rtorrent {
		c.ch <- &job{res: &c.output.Rtorrent[i], fn: chk(ctx, rtorrent, Rtorrent)}
	}

	for i, transmission := range c.input.Transmission {
		c.ch <- &job{res: &c.output.Xmiss[i], fn: chk(ctx, transmission, Transmission)}
	}

	for i, sabnzb := range c.input.SabNZB {
		c.ch <- &job{res: &c.output.SabNZB[i], fn: chk(ctx, sabnzb, SabNZB)}
	}
}

type job struct {
	res *TestResult
	fn  func() (string, int)
}

func (c *checkAll) workPool() func() {
	var wtgrp sync.WaitGroup

	for range workers {
		wtgrp.Add(1)
		go func() {
			defer wtgrp.Done()
			for work := range c.ch {
				work.res.Msg, work.res.Status = work.fn()
			}
		}()
	}

	return wtgrp.Wait
}
