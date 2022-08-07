package website

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/plex"
	"github.com/shirou/gopsutil/v3/host"
	"golift.io/cnfg"
)

const (
	// DefaultRetries is the number of times to attempt a request to notifiarr.com.
	// 4 means 5 total tries: 1 try + 4 retries.
	DefaultRetries = 4
	// RetryDelay is how long to Sleep between retries.
	RetryDelay = 222 * time.Millisecond
)

// Errors returned by this library.
var (
	ErrNon200          = fmt.Errorf("return code was not 200")
	ErrInvalidResponse = fmt.Errorf("invalid response")
	ErrNoChannel       = fmt.Errorf("the website send-data channel is closed")
)

// Config is the input data needed to send payloads to notifiarr.
type Config struct {
	Apps       *apps.Apps
	Plex       *plex.Server // plex sessions
	Serial     bool
	Retries    int
	BaseURL    string
	Timeout    cnfg.Duration
	MaxBody    int
	Sighup     chan os.Signal
	HostID     string
	mnd.Logger // log file writer
}

// Server is what you get for providing a Config to New().
type Server struct {
	config *Config
	// Internal cruft.
	sdMutex      sync.RWMutex // senddata/queuedata
	ciMutex      sync.RWMutex // clientinfo
	clientInfo   *ClientInfo
	client       *httpClient
	hostInfo     *host.InfoStat
	sendData     chan *Request
	stopSendData chan struct{}
}

func New(c *Config) *Server {
	c.BaseURL = BaseURL

	if c.Retries < 0 {
		c.Retries = 0
	} else if c.Retries == 0 {
		c.Retries = DefaultRetries
	}

	return &Server{
		config: c,
		// clientInfo:   &ClientInfo{},
		client: &httpClient{
			Retries: c.Retries,
			Logger:  c.Logger,
			Client:  &http.Client{},
		},
		hostInfo:     nil, // must start nil
		sendData:     make(chan *Request, mnd.Kilobyte),
		stopSendData: make(chan struct{}),
	}
}

func (s *Server) ReloadCh(sighup chan os.Signal) {
	s.config.Sighup = sighup
}

// Start runs the website go routine.
func (s *Server) Start() {
	go s.watchSendDataChan()
}

// Stop stops the website go routine.
func (s *Server) Stop() {
	s.sdMutex.Lock()
	defer s.sdMutex.Unlock()

	if s.sendData != nil {
		close(s.sendData)
	}

	<-s.stopSendData // wait for done signal.
	s.stopSendData = nil
	s.sendData = nil
}

// GetData sends data to a notifiarr URL as JSON.
func (s *Server) GetData(req *Request) (*Response, error) {
	s.sdMutex.RLock()
	defer s.sdMutex.RUnlock()

	if s.sendData == nil {
		return nil, ErrNoChannel
	}

	req.respChan = make(chan *chResponse)
	defer close(req.respChan)

	s.sendData <- req

	resp := <-req.respChan

	return resp.Response, resp.Error
}

// RawGetData sends a request to the website without using a channel.
// Avoid this method.
func (s *Server) RawGetData(req *Request) (*Response, time.Duration, error) {
	return s.sendRequest(req)
}

func (s *Server) sendPayload(uri string, payload interface{}, log bool) (*Response, error) {
	data, err := json.Marshal(payload)
	if err == nil {
		var torn map[string]interface{}
		if err := json.Unmarshal(data, &torn); err == nil {
			if torn["host"], err = s.GetHostInfo(); err != nil {
				s.config.Errorf("Host Info Unknown: %v", err)
			}

			payload = torn
		}
	}

	var post []byte

	if log {
		post, err = json.MarshalIndent(payload, "", " ")
	} else {
		post, err = json.Marshal(payload)
	}

	if err != nil {
		return nil, fmt.Errorf("encoding data to JSON (report this bug please): %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.config.Timeout.Duration)
	defer cancel()

	code, body, err := s.sendJSON(ctx, s.config.BaseURL+uri, post, log)
	if err != nil {
		return nil, err
	}

	return unmarshalResponse(s.config.BaseURL+uri, code, body)
}

// SendData puts a send-data request to notifiarr.com into a channel queue.
func (s *Server) SendData(req *Request) {
	s.sdMutex.RLock()
	defer s.sdMutex.RUnlock()

	if s.sendData != nil {
		s.sendData <- req
	}
}
