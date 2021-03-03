package client

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Go-Lift-TV/discordnotifier-client/pkg/notifiarr"
	"github.com/Go-Lift-TV/discordnotifier-client/pkg/plex"
)

func (c *Client) plexIncoming(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(1000 * 100); err != nil { // nolint:gomnd // 100kbyte memory usage
		c.Errorf("Parsing Multipart Form (plex): %v", err)
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	var v plex.Webhook

	switch err := json.Unmarshal([]byte(r.Form.Get("payload")), &v); {
	case err != nil:
		c.Errorf("Unmarshalling Plex payload: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	case !strings.HasPrefix(v.Event, "media"):
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ignored, non-media\n"))
	case c.plex.Active():
		c.Printf("Plex Incoming Webhook IGNORED (cooldown): %s, %s '%s' => %s",
			v.Server.Title, v.Account.Title, v.Event, v.Metadata.Title)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ignored, cooldown\n"))
	default:
		go c.collectSessions(&v)
	}
}

func (c *Client) collectSessions(v *plex.Webhook) {
	defer c.plex.Done()
	c.Printf("Plex Incoming Webhook: %s, %s '%s' => %s (collecting sessions)",
		v.Server.Title, v.Account.Title, v.Event, v.Metadata.Title)

	body, err := c.notify.SendMeta(v, c.notify.URL, plex.WaitTime)
	if err != nil {
		c.Errorf("Sending Plex Session to Notifiarr: %v", err)
		return
	}

	if fields := strings.Split(string(body), `"`); len(fields) > 3 { // nolint:gomnd
		c.Printf("Plex => Notifiarr: %s '%s' => %s (%s)", v.Account.Title, v.Event, v.Metadata.Title, fields[3])
	} else {
		c.Printf("Plex => Notifiarr: %s '%s' => %s", v.Account.Title, v.Event, v.Metadata.Title)
	}
}

// sendPlexSessions is triggered from a menu-bar item.
func (c *Client) sendPlexSessions(url string) {
	c.Printf("[user requested] Sending Plex Sessions to %s", url)

	if body, err := c.notify.SendMeta(nil, url, 0); err != nil {
		c.Errorf("[user requested] Sending Plex Sessions to %s: %v: %v", url, err, string(body))
	} else if fields := strings.Split(string(body), `"`); len(fields) > 3 { //nolint:gomnd
		c.Printf("[user requested] Sent Plex Sessions to %s, reply: %s", url, fields[3])
	} else {
		c.Printf("[user requested] Sent Plex Sessions to %s, reply: %s", url, string(body))
	}
}

// sendSystemSnapshot is triggered from a menu-bar item.
func (c *Client) sendSystemSnapshot(url string) {
	c.Printf("[user requested] Sending System Snapshot to %s", url)

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

	b, _ := json.Marshal(&notifiarr.Payload{Snap: snaps})
	if body, err := c.notify.SendJSON(url, b); err != nil {
		c.Errorf("[user requested] Sending System Snapshot to %s: %v: %s", url, err, string(body))
	} else if fields := strings.Split(string(body), `"`); len(fields) > 3 { //nolint:gomnd
		c.Printf("[user requested] Sent System Snapshot to %s, reply: %s", url, fields[3])
	} else {
		c.Printf("[user requested] Sent System Snapshot to %s, reply: %s", url, string(body))
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
		if plex, err = c.Config.Plex.GetSessions(); err != nil {
			c.Errorf("[user requested] %v", err)
		}
	}

	b, _ := json.Marshal(&notifiarr.Payload{Snap: snaps, Plex: plex})
	c.Printf("[user requested] Snapshot Data:\n%s", string(b))
}
