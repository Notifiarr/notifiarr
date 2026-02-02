package website

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
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

// server is what you get for providing a Config to New().
type server struct {
	config *Config
	// Internal cruft.
	client    *httpClient
	hostInfo  *host.InfoStat
	sendData  chan *Request // in (buffered)
	reconfig  chan *Config  // in+out (unbuffered bidirectional)
	getConfig chan struct{} // in (buffered)
	mu        sync.RWMutex
}

func New(ctx context.Context, config *Config) {
	if config.Retries < 0 {
		config.Retries = 0
	} else if config.Retries == 0 {
		config.Retries = DefaultRetries
	}

	if site != nil {
		site.reconfig <- config
		return
	}

	site = &server{
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

	go site.watchSendDataChan(ctx)
}

// SendData puts a POST request to notifiarr.com into a channel queue.
func SendData(req *Request) {
	site.sendData <- req
}

// GetData sends data to a notifiarr URL as JSON and returns a response.
func GetData(req *Request) (*Response, error) {
	req.respChan = make(chan *chResponse)
	defer close(req.respChan)

	site.sendData <- req
	resp := <-req.respChan

	return resp.Response, resp.Error
}

// GetConfig returns the current website config.
func GetConfig() *Config {
	site.mu.RLock()
	defer site.mu.RUnlock()

	return site.config
}

// RawGetData sends a request to the website without using a channel.
// Avoid this method, it can trigger data races. Use it only in cli.go.
func RawGetData(ctx context.Context, req *Request) (*Response, time.Duration, error) {
	return site.sendRequest(ctx, req)
}

// ValidAPIKey checks if the API key is valid.
func ValidAPIKey() error {
	if len(GetConfig().Apps.APIKey) != APIKeyLength {
		return fmt.Errorf("%w: length must be %d characters", ErrInvalidAPIKey, APIKeyLength)
	}

	return nil
}

func (s *server) watchSendDataChan(ctx context.Context) {
	for {
		select {
		case config := <-s.reconfig:
			s.mu.Lock()
			s.client.Retries = config.Retries
			s.config = config
			s.mu.Unlock()
		case data := <-s.sendData:
			ctx := mnd.WithID(ctx, data.ReqID)
			data.ReqID = mnd.GetID(ctx)
			s.sendAndLogRequest(ctx, data)
		}
	}
}
