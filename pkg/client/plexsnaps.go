package client

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Go-Lift-TV/notifiarr/pkg/notifiarr"
	"github.com/Go-Lift-TV/notifiarr/pkg/plex"
)

func (c *Client) plexIncoming(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(1000 * 100); err != nil { // nolint:gomnd // 100kbyte memory usage
		c.Errorf("Parsing Multipart Form (plex): %v", err)
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	var v plex.Webhook

	payload := r.Form.Get("payload")
	c.Debugf("Plex Webhook Payload: %s", payload)

	switch err := json.Unmarshal([]byte(payload), &v); {
	case err != nil:
		c.Errorf("Unmarshalling Plex payload: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	case !strings.HasPrefix(v.Event, "media"):
		w.WriteHeader(http.StatusNoContent)
		_, _ = w.Write([]byte("ignored, non-media\n"))
	case c.plex.Active():
		c.Printf("Plex Incoming Webhook IGNORED (cooldown): %s, %s '%s' => %s",
			v.Server.Title, v.Account.Title, v.Event, v.Metadata.Title)
		w.WriteHeader(http.StatusNoContent)
		_, _ = w.Write([]byte("ignored, cooldown\n"))
	default:
		go c.collectSessions(&v)
		w.WriteHeader(http.StatusAccepted)
	}
}

func (c *Client) collectSessions(v *plex.Webhook) {
	defer c.plex.Done()
	c.Printf("Plex Incoming Webhook: %s, %s '%s' => %s (collecting sessions)",
		v.Server.Title, v.Account.Title, v.Event, v.Metadata.Title)

	reply, err := c.notify.SendMeta(notifiarr.PlexHook, c.notify.URL, v, plex.WaitTime)
	if err != nil {
		c.Errorf("Sending Plex Session to Notifiarr: %v", err)
		return
	}

	if fields := strings.Split(string(reply), `"`); len(fields) > 3 { // nolint:gomnd
		c.Printf("Plex => Notifiarr: %s '%s' => %s (%s)", v.Account.Title, v.Event, v.Metadata.Title, fields[3])
	} else {
		c.Printf("Plex => Notifiarr: %s '%s' => %s", v.Account.Title, v.Event, v.Metadata.Title)
	}
}

// logSnaps writes a full snapshot payload to the log file.
func (c *Client) logSnaps() {
	c.Printf("[user requested] Collecting Snapshot from Plex and the System (for log file).")

	snaps, errs, debug := c.Config.Snapshot.GetSnapshot()
	for _, err := range errs {
		if err != nil {
			c.Errorf("[user requested] %v", err)
		}
	}

	for _, err := range debug {
		if err != nil {
			c.Errorf("[user requested] %v", err)
		}
	}

	var (
		plex *plex.Sessions
		err  error
	)

	if c.Config.Plex != nil {
		if plex, err = c.Config.Plex.GetXMLSessions(); err != nil {
			c.Errorf("[user requested] %v", err)
		}
	}

	b, _ := json.MarshalIndent(&notifiarr.Payload{
		Type: notifiarr.LogLocal,
		Snap: snaps,
		Plex: plex,
	}, "", "  ")
	c.Printf("[user requested] Snapshot Data:\n%s", string(b))
}
