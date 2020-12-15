package dnclient

import (
	"sync"

	"golift.io/starr"
)

/*
$readarrPayload['foreignBookId']    = $work['id'];
$readarrPayload['monitored']         = true;
$readarrPayload['addOptions']        = array('searchForNewBook' => true);
$readarrPayload['author']            = array('rootFolderPath' => $paths['books']['everyone'], 'qualityProfileId' => 1, 'metadataProfileId' => 2, 'foreignAuthorId' => $author['id'], 'monitored' => true);
$readarrPayload['editions'][]        = array('foreignEditionId' => $grid, 'monitored' => true, 'manualAdd' => true);
[3:23 PM] nitsua:
curl_setopt($ch, CURLOPT_URL, 'http://10.1.0.63:8787/api/v1/book');
*/

// ReadarrAddBook is the data we expect to get from discord notifier.
type ReadarrAddBook struct {
	Root string `json:"root_folder"` // optional
	ID   int    `json:"id"`          // default: 0 (configured instance)
	GRID int    `json:"grid"`        // required if App = readarr
}

func (c *Config) fixReadarrConfig() {
	for i := range c.Readarr {
		if c.Readarr[i].Timeout.Duration == 0 {
			c.Readarr[i].Timeout.Duration = c.Timeout.Duration
		}
	}
}

// ReadarrConfig represents the input data for a Readarr server.
type ReadarrConfig struct {
	Name string `json:"name" toml:"name" xml:"name" yaml:"name"`
	*starr.Config
	sync.RWMutex `json:"-" toml:"-" xml:"-" yaml:"-"`
}

func (c *Client) logReadarr() {
	if count := len(c.Readarr); count == 1 {
		c.Printf(" => Readarr Config: 1 server: %s / %s, apikey:%v, timeout:%v, verify ssl:%v",
			c.Readarr[0].Name, c.Readarr[0].URL, c.Readarr[0].APIKey != "", c.Readarr[0].Timeout, c.Readarr[0].ValidSSL)
	} else {
		c.Print(" => Readarr Config:", count, "servers")

		for _, f := range c.Readarr {
			c.Printf(" =>    Server: %s / %s, apikey:%v, timeout:%v, verify ssl:%v",
				f.Name, f.URL, f.APIKey != "", f.Timeout, f.ValidSSL)
		}
	}
}

func (c *Client) handleReadarr(p *ReadarrAddBook) (string, error) {
	return "", nil
}
