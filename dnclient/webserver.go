package dnclient

import (
	"encoding/json"
	"net/http"
	"strings"
)

// RunWebServer starts the web server.
func (c *Client) RunWebServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/add", c.Handle)

	if err := http.ListenAndServe(c.Config.BindAddr, mux); err != nil && err != http.ErrServerClosed {
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

	switch payload.App {
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
		w.WriteHeader(http.StatusNotFound)
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
