package dnclient

import (
	"sync"

	"golift.io/starr"
)

/*
$series                                                        = findSonarrSeries('', $id);
$sonarrPayload['id']                                         = $series['id'];
$sonarrPayload['tvdbId']                                     = $series['tvdbId'];
$sonarrPayload['monitored']                                 = $series['monitored'];
$sonarrPayload['qualityProfileId']                             = $series['qualityProfileId'];
$sonarrPayload['path']                                        = $destination .'\\'. end(explode('\\', $series['path']));
$sonarrPayload['title']                                        = $series['title'];
$sonarrPayload['seasonFolder']                                = $series['seasonFolder'];
$sonarrPayload['seriesType']                                = $series['seriesType'];
$sonarrPayload['languageProfileId']                            = $series['languageProfileId'];
$sonarrPayload['seasons']                                    = $series['seasons'];

curl_setopt($ch, CURLOPT_URL, 'http://10.1.0.63:8989/api/v3/series/'. $id .'?moveFiles=true');

*/

// SonarrAddSeries is the data we expect to get from discord notifier.
type SonarrAddSeries struct {
	Root string `json:"root_folder"` // optional
	ID   int    `json:"id"`          // default: 0 (configured instance)
	TVDB int    `json:"tvdb"`        // required if App = sonarr
}

func (c *Config) fixSonarrConfig() {
	for i := range c.Sonarr {
		if c.Sonarr[i].Timeout.Duration == 0 {
			c.Sonarr[i].Timeout.Duration = c.Timeout.Duration
		}
	}
}

// SonarrConfig represents the input data for a Sonarr server.
type SonarrConfig struct {
	Name string `json:"name" toml:"name" xml:"name" yaml:"name"`
	*starr.Config
	sync.RWMutex `json:"-" toml:"-" xml:"-" yaml:"-"`
}

func (c *Client) logSonarr() {
	if count := len(c.Sonarr); count == 1 {
		c.Printf(" => Sonarr Config: 1 server: %s / %s, apikey:%v, timeout:%v, verify ssl:%v",
			c.Sonarr[0].Name, c.Sonarr[0].URL, c.Sonarr[0].APIKey != "", c.Sonarr[0].Timeout, c.Sonarr[0].ValidSSL)
	} else {
		c.Print(" => Sonarr Config:", count, "servers")

		for _, f := range c.Sonarr {
			c.Printf(" =>    Server: %s / %s, apikey:%v, timeout:%v, verify ssl:%v",
				f.Name, f.URL, f.APIKey != "", f.Timeout, f.ValidSSL)
		}
	}
}

// getSonarr finds a Sonarr based on the passed-in ID.
// Every Sonarr handler calls this.
func (c *Client) getSonarr(id string) *LidarrConfig {
	j, _ := strconv.Atoi(id)

	for i, app := range c.Sonarr {
		if i != j-1 { // discordnotifier wants 1-indexes
			continue
		}

		return app
	}

	return nil
}

func (c *Client) handleSonarr(p *SonarrAddSeries) (string, error) {
	return "", nil
}
