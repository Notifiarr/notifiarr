package dnclient

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// RunWebServer starts the web server.
func (c *Client) RunWebServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/add", c.Handle)
	mux.HandleFunc("/", c.notFound)

	c.server = &http.Server{
		Handler:      mux,
		Addr:         c.Config.BindAddr,
		IdleTimeout:  time.Second,
		WriteTimeout: time.Second,
		ReadTimeout:  time.Second,
		ErrorLog:     c.Logger.Logger,
	}
	if err := c.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		c.Printf("[ERROR] HTTP Server: %v", err)
	}
}

// Handle is our main web server handler.
func (c *Client) Handle(w http.ResponseWriter, r *http.Request) {
	payload := &IncomingPayload{}

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		c.Printf("HTTP [%s] %s %s: %s", r.RemoteAddr, r.Method, r.RequestURI, http.StatusText(http.StatusMethodNotAllowed))

		return
	} else if err := json.NewDecoder(r.Body).Decode(payload); err != nil {
		c.Printf("HTTP [%s] %s %s: %s: %v", r.RemoteAddr, r.Method, r.RequestURI, http.StatusText(http.StatusBadRequest), err)
		w.WriteHeader(http.StatusBadRequest)

		return
	} else if c.Config.APIKey != payload.Key {
		c.Printf("HTTP [%s] %s %s: %s", r.RemoteAddr, r.Method, r.RequestURI, http.StatusText(http.StatusUnauthorized))
		w.WriteHeader(http.StatusUnauthorized)

		return
	}

	switch s := strings.ToLower(payload.App); s {
	case "radarr":
		msg, err := c.handleRadarr(payload)
		c.respond(w, r, payload, msg, err)
	case "sonarr":
		msg, err := c.handleSonarr(payload)
		c.respond(w, r, payload, msg, err)
	case "readarr":
		msg, err := c.handleReadarr(payload)
		c.respond(w, r, payload, msg, err)
	case "lidarr":
		msg, err := c.handleLidarr(payload)
		c.respond(w, r, payload, msg, err)
	default:
		c.Printf("HTTP [%s] %s %s: %s: %s", r.RemoteAddr, r.Method, r.RequestURI, http.StatusText(http.StatusNotFound), s)
		w.WriteHeader(http.StatusUnprocessableEntity)
	}
}

func (c *Client) respond(w http.ResponseWriter, r *http.Request, p *IncomingPayload, msg string, err error) {
	if msg = strings.Join([]string{p.Title, msg}, ": "); err != nil {
		msg = strings.Join([]string{msg, err.Error()}, ": ")
	}

	c.Printf("HTTP [%s] %s %s: %s: %s", r.RemoteAddr, r.Method, r.RequestURI, http.StatusText(http.StatusOK), msg)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("content-type", "application/json")

	b, _ := json.Marshal(&OutgoingPayload{Status: err == nil, Message: msg})
	_, _ = w.Write(b)
	_, _ = w.Write([]byte("\n"))
}

// notFound is the handler for paths that are not found: 404s.
func (c *Client) notFound(w http.ResponseWriter, r *http.Request) {
	c.Printf("HTTP [%s] %s %s: %s", r.RemoteAddr, r.Method, r.RequestURI, http.StatusText(http.StatusNotFound))
	w.WriteHeader(http.StatusNotFound)
}
