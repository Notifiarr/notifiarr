package website

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/shirou/gopsutil/v4/host"
	"golift.io/cnfg"
)

const (
	// DefaultRetries is the number of times to attempt a request to notifiarr.com.
	// 4 means 5 total tries: 1 try + 4 retries.
	DefaultRetries = 4
	// RetryDelay is how long to Sleep between retries.
	RetryDelay = 222 * time.Millisecond
	// APIKeyLength is the string length of a valid notifiarr API key.
	APIKeyLength = 36
	// MaxTimeout is the maximum timeout for a request to notifiarr.com.
	MaxTimeout = 3 * time.Minute
	// MinTimeout is the minimum timeout for a request to notifiarr.com.
	MinTimeout = 10 * time.Second
)

// Errors returned by this library.
var (
	ErrNon200          = errors.New("return code was not 200")
	ErrInvalidResponse = errors.New("invalid response")
	ErrNoChannel       = errors.New("the website send-data channel is closed")
	ErrInvalidAPIKey   = errors.New("configured notifiarr API key is invalid")
)

// Config is the input data needed to send payloads to notifiarr.
type Config struct {
	Apps     *apps.Apps
	Retries  int
	Timeout  cnfg.Duration
	HostID   string
	BindAddr string
}

// Server is what you get for providing a Config to New().
type Server struct {
	config *Config
	// Internal cruft.
	client    *httpClient
	hostInfo  *host.InfoStat
	sendData  chan *Request // in (buffered)
	reconfig  chan *Config  // in+out (unbuffered bidirectional)
	getConfig chan struct{} // in (buffered)
}

func New(ctx context.Context, config *Config) {
	if config.Retries < 0 {
		config.Retries = 0
	} else if config.Retries == 0 {
		config.Retries = DefaultRetries
	}

	if Site != nil {
		Site.reconfig <- config
		return
	}

	Site = &Server{
		config: config,
		client: &httpClient{
			Retries: config.Retries,
			Client:  &http.Client{},
		},
		hostInfo:  nil, // must start nil
		sendData:  make(chan *Request, mnd.Base8),
		reconfig:  make(chan *Config), // do not buffer.
		getConfig: make(chan struct{}, 1),
	}

	go Site.watchSendDataChan(ctx)
}

// SendData puts a POST request to notifiarr.com into a channel queue.
func (s *Server) SendData(req *Request) {
	s.sendData <- req
}

// GetData sends data to a notifiarr URL as JSON and returns a response.
func (s *Server) GetData(req *Request) (*Response, error) {
	req.respChan = make(chan *chResponse)
	defer close(req.respChan)

	s.sendData <- req
	resp := <-req.respChan

	return resp.Response, resp.Error
}

// GetConfig returns the current website config.
func GetConfig() *Config {
	Site.getConfig <- struct{}{}
	return <-Site.reconfig
}

// RawGetData sends a request to the website without using a channel.
// Avoid this method, it can trigger data races. Use it only in cli.go.
func RawGetData(ctx context.Context, req *Request) (*Response, time.Duration, error) {
	return Site.sendRequest(ctx, req)
}

func (s *Server) watchSendDataChan(ctx context.Context) {
	for {
		select {
		case config := <-s.reconfig:
			s.client.Retries = config.Retries
			s.config = config
		case <-s.getConfig:
			s.reconfig <- s.config
		case data := <-s.sendData:
			s.sendAndLogRequest(ctx, data)
		}
	}
}

// ValidAPIKey checks if the API key is valid.
func ValidAPIKey() error {
	if len(GetConfig().Apps.APIKey) != APIKeyLength {
		return fmt.Errorf("%w: length must be %d characters", ErrInvalidAPIKey, APIKeyLength)
	}

	return nil
}
