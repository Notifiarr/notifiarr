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

	body, err := c.notifiarr.SendMeta(v, plex.WaitTime)
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

// Temporary code? sure is ugly and should probably go in notifiarr.
func (c *Client) testSnaps(url string) {
	snaps, errs, debug := c.Config.Snapshot.GetSnapshot()
	for _, err := range errs {
		if err != nil {
			c.Errorf("%v", err)
		}
	}

	for _, err := range debug {
		if err != nil {
			c.Errorf("%v", err)
		}
	}

	var (
		plex *plex.Sessions
		err  error
	)

	if c.Config.Plex != nil {
		if plex, err = c.Config.Plex.GetSessions(); err != nil {
			c.Errorf("%v", err)
		}
	}

	b, _ := json.Marshal(&notifiarr.Payload{
		Snap: snaps,
		Plex: plex,
	})

	c.Printf("Snapshot Data:\n%s", string(b))

	if url == "" {
		return
	}

	if body, err := c.notifiarr.SendJSON(url, b); err != nil {
		c.Errorf("POSTING: %v: %s", err, string(body))
	} else if fields := strings.Split(string(body), `"`); len(fields) > 3 { //nolint:gomnd
		c.Printf("Sent Test Snapshot to %s, reply: %s", url, fields[3])
	} else {
		c.Printf("Sent Test Snapshot to %s, reply: %s", url, string(body))
	}
}
