package checkapp

import (
	"context"
	"sync"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/logs"
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
	Status  int              `json:"status"`
	Msg     string           `json:"message"`
	Elapsed string           `json:"elapsed"`
	Config  apps.ExtraConfig `json:"config"`
}

// CheckAllOutput is the output from a check all instances test.
// The JSON keys are used for human display, so ya.
//
//nolint:tagliatelle
type CheckAllOutput struct {
	Sonarr    []TestResult `json:"Sonarr"`
	Radarr    []TestResult `json:"Radarr"`
	Readarr   []TestResult `json:"Readarr"`
	Lidarr    []TestResult `json:"Lidarr"`
	Prowlarr  []TestResult `json:"Prowlarr"`
	Plex      []TestResult `json:"Plex"`
	Tautulli  []TestResult `json:"Tautulli"`
	NZBGet    []TestResult `json:"NZBGet"`
	Deluge    []TestResult `json:"Deluge"`
	Qbit      []TestResult `json:"Qbittorrent"`
	Rtorrent  []TestResult `json:"Rtorrent"`
	Xmiss     []TestResult `json:"Transmission"`
	SabNZB    []TestResult `json:"SabNZB"`
	TimeMS    int64        `json:"timeMS"`
	Elapsed   int64        `json:"elapsed"`
	Workers   int          `json:"workers"`
	Instances int          `json:"instances"`
}

type checkAll struct {
	input  *CheckAllInput
	output *CheckAllOutput
	ch     chan *job
}

const (
	timeout = 5 * time.Second
	divider = 2
	cBuffer = 100
)

// CheckAll checks all the starr, media and downloader apps and returns the results.
func CheckAll(ctx context.Context, input *CheckAllInput) *CheckAllOutput {
	return newCheckAll(input).run(ctx)
}

func newCheckAll(input *CheckAllInput) *checkAll {
	output := &CheckAllOutput{
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
		Instances: len(input.Sonarr) + len(input.Radarr) + len(input.Readarr) +
			len(input.Lidarr) + len(input.Prowlarr) + len(input.Plex) +
			len(input.Tautulli) + len(input.NZBGet) + len(input.Deluge) +
			len(input.Qbit) + len(input.Rtorrent) + len(input.Transmission) +
			len(input.SabNZB),
	}

	return &checkAll{input: input, ch: make(chan *job, cBuffer), output: output}
}

func (c *checkAll) run(ctx context.Context) *CheckAllOutput {
	c.output.Workers = c.output.Instances/divider + 1
	wait := c.workPool(c.output.Workers)

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	start := time.Now()
	c.checkAllStarr(ctx)
	c.checkAllMedia(ctx)
	c.checkAllDownloaders(ctx)
	close(c.ch)
	wait()

	c.output.Elapsed = time.Since(start).Milliseconds()

	return c.output
}

func chk[D any](ctx context.Context, d D, fn func(context.Context, D) (string, int)) func() (string, int) {
	return func() (string, int) { return fn(ctx, d) }
}

func (c *checkAll) checkAllStarr(ctx context.Context) {
	for i, sonarr := range c.input.Sonarr {
		c.output.Sonarr[i].Config = sonarr.ExtraConfig
		c.ch <- &job{res: &c.output.Sonarr[i], fn: chk(ctx, sonarr, Sonarr)}
	}

	for i, radarr := range c.input.Radarr {
		c.output.Radarr[i].Config = radarr.ExtraConfig
		c.ch <- &job{res: &c.output.Radarr[i], fn: chk(ctx, radarr, Radarr)}
	}

	for i, readarr := range c.input.Readarr {
		c.output.Readarr[i].Config = readarr.ExtraConfig
		c.ch <- &job{res: &c.output.Readarr[i], fn: chk(ctx, readarr, Readarr)}
	}

	for i, lidarr := range c.input.Lidarr {
		c.output.Lidarr[i].Config = lidarr.ExtraConfig
		c.ch <- &job{res: &c.output.Lidarr[i], fn: chk(ctx, lidarr, Lidarr)}
	}

	for i, prowlarr := range c.input.Prowlarr {
		c.output.Prowlarr[i].Config = prowlarr.ExtraConfig
		c.ch <- &job{res: &c.output.Prowlarr[i], fn: chk(ctx, prowlarr, Prowlarr)}
	}
}

func (c *checkAll) checkAllMedia(ctx context.Context) {
	for i, plex := range c.input.Plex {
		c.output.Plex[i].Config = plex.ExtraConfig
		c.ch <- &job{res: &c.output.Plex[i], fn: chk(ctx, plex, Plex)}
	}

	for i, tautulli := range c.input.Tautulli {
		c.output.Tautulli[i].Config = tautulli.ExtraConfig
		c.ch <- &job{res: &c.output.Tautulli[i], fn: chk(ctx, tautulli, Tautulli)}
	}
}

func (c *checkAll) checkAllDownloaders(ctx context.Context) {
	for i, nzbget := range c.input.NZBGet {
		c.output.NZBGet[i].Config = nzbget.ExtraConfig
		c.ch <- &job{res: &c.output.NZBGet[i], fn: chk(ctx, nzbget, NZBGet)}
	}

	for i, deluge := range c.input.Deluge {
		c.output.Deluge[i].Config = deluge.ExtraConfig
		c.ch <- &job{res: &c.output.Deluge[i], fn: chk(ctx, deluge, Deluge)}
	}

	for i, qbit := range c.input.Qbit {
		c.output.Qbit[i].Config = qbit.ExtraConfig
		c.ch <- &job{res: &c.output.Qbit[i], fn: chk(ctx, qbit, Qbit)}
	}

	for i, rtorrent := range c.input.Rtorrent {
		c.output.Rtorrent[i].Config = rtorrent.ExtraConfig
		c.ch <- &job{res: &c.output.Rtorrent[i], fn: chk(ctx, rtorrent, Rtorrent)}
	}

	for i, transmission := range c.input.Transmission {
		c.output.Xmiss[i].Config = transmission.ExtraConfig
		c.ch <- &job{res: &c.output.Xmiss[i], fn: chk(ctx, transmission, Transmission)}
	}

	for i, sabnzb := range c.input.SabNZB {
		c.output.SabNZB[i].Config = sabnzb.ExtraConfig
		c.ch <- &job{res: &c.output.SabNZB[i], fn: chk(ctx, sabnzb, SabNZB)}
	}
}

type job struct {
	res *TestResult
	fn  func() (string, int)
}

func (c *checkAll) workPool(workers int) func() {
	var (
		wtgrp sync.WaitGroup
		start time.Time
	)

	logs.Log.Debugf("[gui requested] Checking %d instances with %d workers.", c.output.Instances, workers)

	for range workers {
		wtgrp.Go(func() {
			for work := range c.ch {
				start = time.Now()
				work.res.Msg, work.res.Status = work.fn()
				work.res.Elapsed = time.Since(start).Round(time.Millisecond).String()
			}
		})
	}

	return wtgrp.Wait
}
