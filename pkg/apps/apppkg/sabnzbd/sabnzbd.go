package sabnzbd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
)

var ErrUnknownByteType = errors.New("unknown byte type")

type Config struct {
	URL          string `toml:"url"     xml:"url"`
	APIKey       string `toml:"api_key" xml:"api_key"`
	*http.Client `json:"-"       toml:"-"      xml:"-"`
}

// QueueSlots has the following data structure.
/*
{
  "index": 1,
  "nzo_id": "SABnzbd_nzo_xnfbbdbh",
  "unpackopts": "3",
  "priority": "Normal",
  "script": "wtfnzb-renamer.py",
  "filename": "Rick Astley - Never Gonna Give You Up (1987)(24bit flac vinyl)",
  "labels": [],
  "password": "",
  "cat": "prowlarr",
  "mbleft": "593.67",
  "mb": "701.34",
  "size": "701.3 MB",
  "sizeleft": "593.7 MB",
  "percentage": "15",
  "mbmissing": "0.00",
  "direct_unpack": 0,
  "status": "Downloading",
  "timeleft": "0:03:57",
  "eta": "13:42 Sun 17 Oct",
  "avg_age": "2537d",
  "has_rating": false
}
Payload for this structure. */
//nolint:tagliatelle
type QueueSlots struct {
	Status     string     `json:"status"`
	Index      int        `json:"index"`
	Password   string     `json:"password"`
	AvgAge     string     `json:"avg_age"`
	Script     string     `json:"script"`
	HasRating  bool       `json:"has_rating"`
	Mb         string     `json:"mb"`
	Mbleft     float64    `json:"mbleft,string"`
	Mbmissing  float64    `json:"mbmissing,string"`
	Size       SabNZBSize `json:"size"`
	Sizeleft   SabNZBSize `json:"sizeleft"`
	Filename   string     `json:"filename"`
	Labels     []string   `json:"labels"`
	Priority   string     `json:"priority"`
	Cat        string     `json:"cat"`
	Eta        SabNZBDate `json:"eta"`
	Timeleft   string     `json:"timeleft"`
	Percentage int        `json:"percentage,string"`
	NzoID      string     `json:"nzo_id"`
	Unpackopts string     `json:"unpackopts"`
}

// StageLog is part of the json response from SABnzbd.
type StageLog struct {
	Name    string   `json:"name"`
	Actions []string `json:"actions"`
}

// HistorySlots is part of the json response from SABnzbd.
//
//nolint:tagliatelle
type HistorySlots struct {
	ID           int64       `json:"id"`
	Completed    int64       `json:"completed"`
	Name         string      `json:"name"`
	NzbName      string      `json:"nzb_name"`
	Category     string      `json:"category"`
	Pp           string      `json:"pp"`
	Script       string      `json:"script"`
	Report       string      `json:"report"`
	URL          string      `json:"url"`
	Status       string      `json:"status"`
	NzoID        string      `json:"nzo_id"`
	Storage      string      `json:"storage"`
	Path         string      `json:"path"`
	ScriptLog    string      `json:"script_log"`
	ScriptLine   string      `json:"script_line"`
	DownloadTime int64       `json:"download_time"`
	PostprocTime int64       `json:"postproc_time"`
	StageLog     []*StageLog `json:"stage_log"`
	Downloaded   int64       `json:"downloaded"`
	Completeness interface{} `json:"completeness"`
	FailMessage  string      `json:"fail_message"`
	URLInfo      string      `json:"url_info"`
	Bytes        int64       `json:"bytes"`
	Meta         interface{} `json:"meta"`
	Series       string      `json:"series"`
	Md5Sum       string      `json:"md5sum"`
	Password     string      `json:"password"`
	ActionLine   string      `json:"action_line"`
	Size         string      `json:"size"`
	Loaded       bool        `json:"loaded"`
	Retry        Bool        `json:"retry"`
}

// Bool exists because once upon a time sab changed the retry value from int to bool.
// https://github.com/sabnzbd/sabnzbd/issues/2911
type Bool bool

func (f *Bool) UnmarshalJSON(b []byte) error {
	txt := string(bytes.Trim(b, `"`))
	*f = Bool(txt == "true" || (txt != "0" && txt != "false"))

	return nil
}

//nolint:tagliatelle
type History struct {
	TotalSize         SabNZBSize     `json:"total_size"`
	MonthSize         SabNZBSize     `json:"month_size"`
	WeekSize          SabNZBSize     `json:"week_size"`
	DaySize           SabNZBSize     `json:"day_size"`
	Slots             []HistorySlots `json:"slots"`
	Noofslots         int            `json:"noofslots"`
	LastHistoryUpdate int64          `json:"last_history_update"`
	Version           string         `json:"version"`
}

//nolint:tagliatelle
type Queue struct {
	Version           string       `json:"version"`
	PauseInt          string       `json:"pause_int"`
	Diskspace1        float64      `json:"diskspace1,string"`
	Diskspace2        float64      `json:"diskspace2,string"`
	Diskspace1Norm    SabNZBSize   `json:"diskspace1_norm"`
	Diskspace2Norm    SabNZBSize   `json:"diskspace2_norm"`
	Diskspacetotal1   float64      `json:"diskspacetotal1,string"`
	Diskspacetotal2   float64      `json:"diskspacetotal2,string"`
	Loadavg           string       `json:"loadavg"`
	Speedlimit        float64      `json:"speedlimit,string"`
	SpeedlimitAbs     string       `json:"speedlimit_abs"`
	HaveWarnings      string       `json:"have_warnings"`
	Finishaction      interface{}  `json:"finishaction"`
	Quota             string       `json:"quota"`
	LeftQuota         string       `json:"left_quota"`
	CacheArt          string       `json:"cache_art"`
	CacheSize         SabNZBSize   `json:"cache_size"`
	CacheMax          int64        `json:"cache_max,string"`
	Kbpersec          float64      `json:"kbpersec,string"`
	Speed             SabNZBSize   `json:"speed"`
	Mbleft            float64      `json:"mbleft,string"`
	Mb                float64      `json:"mb,string"`
	Sizeleft          SabNZBSize   `json:"sizeleft"`
	Size              SabNZBSize   `json:"size"`
	NoofslotsTotal    int          `json:"noofslots_total"`
	Status            string       `json:"status"`
	Timeleft          string       `json:"timeleft"`
	Eta               string       `json:"eta"`
	RefreshRate       string       `json:"refresh_rate"`
	InterfaceSettings string       `json:"interface_settings"`
	Scripts           []string     `json:"scripts"`
	Categories        []string     `json:"categories"`
	Noofslots         int          `json:"noofslots"`
	Start             int64        `json:"start"`
	Limit             int64        `json:"limit"`
	Finish            int64        `json:"finish"`
	Slots             []QueueSlots `json:"slots"`
	PausedAll         bool         `json:"paused_all"`
	RatingEnable      bool         `json:"rating_enable"`
	Paused            bool         `json:"paused"`
	HaveQuota         bool         `json:"have_quota"`
}

// GetHistory returns the history items in SABnzbd.
func (s *Config) GetHistory(ctx context.Context) (*History, error) {
	if s == nil || s.URL == "" {
		return &History{}, nil
	}

	params := url.Values{}
	params.Add("output", "json")
	params.Add("mode", "history")
	params.Add("apikey", s.APIKey)

	var hist struct {
		History *History `json:"history"`
	}

	err := s.GetURLInto(ctx, params, &hist)
	if err != nil {
		return nil, err
	}

	return hist.History, nil
}

// GetQueue returns the active queued items in SABnzbd.
func (s *Config) GetQueue(ctx context.Context) (*Queue, error) {
	if s == nil || s.URL == "" {
		return &Queue{}, nil
	}

	params := url.Values{}
	params.Add("output", "json")
	params.Add("mode", "queue")
	params.Add("apikey", s.APIKey)

	var que struct {
		Queue *Queue `json:"queue"`
	}

	err := s.GetURLInto(ctx, params, &que)
	if err != nil {
		return nil, err
	}

	return que.Queue, nil
}

// GetURLInto gets a url and unmarshals the contents into the provided interface pointer.
func (s *Config) GetURLInto(ctx context.Context, params url.Values, into interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.URL+"/api", nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.URL.RawQuery = params.Encode()

	resp, err := s.Client.Do(req)
	if err != nil {
		return fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response (%s): %w: %s", resp.Status, err, string(body))
	}

	if err := json.Unmarshal(body, into); err != nil {
		return fmt.Errorf("decoding response (%s): %w: %s", resp.Status, err, string(body))
	}

	return nil
}

// SabNZBSize deals with bytes encoded as strings.
type SabNZBSize struct {
	Bytes  int64
	String string
}

// SabNZBDate is used to parse a custm date format from the json api.
type SabNZBDate struct {
	String string
	time.Time
}

// UnmarshalJSON exists because weird date formats and "unknown" seem sane in json output.
func (s *SabNZBDate) UnmarshalJSON(b []byte) error {
	s.String = strings.Trim(string(b), `"`)

	if s.String == "unknown" {
		s.Time = time.Now().Add(time.Hour * 24 * 366) //nolint:mnd
		return nil
	}

	var err error

	s.Time, err = time.Parse("15:04 Mon 02 Jan 2006", s.String+" "+strconv.Itoa(time.Now().Year()))
	if err != nil {
		return fmt.Errorf("invalid time: %w", err)
	}

	return nil
}

// UnmarshalJSON exists because someone decided that bytes should be strings with letters.
func (s *SabNZBSize) UnmarshalJSON(b []byte) error {
	s.String = strings.Trim(string(b), `"`)
	split := strings.Split(s.String, " ")

	bytes, err := strconv.ParseFloat(split[0], mnd.Bits64)
	if err != nil {
		return fmt.Errorf("could not convert to number: %s: %w", split[0], err)
	}

	if len(split) < 2 { //nolint:mnd
		s.Bytes = int64(bytes)
		return nil
	}

	switch strings.ToLower(split[1]) {
	case "", "b":
		s.Bytes = int64(bytes)
	case "k", "kb":
		s.Bytes = int64(bytes * mnd.Kilobyte)
	case "m", "mb":
		s.Bytes = int64(bytes * mnd.Megabyte)
	case "g", "gb":
		s.Bytes = int64(bytes * mnd.Megabyte * mnd.Kilobyte)
	case "t", "tb":
		s.Bytes = int64(bytes * mnd.Megabyte * mnd.Megabyte)
	case "p", "pb":
		s.Bytes = int64(bytes * mnd.Megabyte * mnd.Megabyte * mnd.Kilobyte)
	default:
		return fmt.Errorf("%w: %s", ErrUnknownByteType, split[1])
	}

	return nil
}
