package notifiarr

import (
	"strings"

	"golift.io/starr/sonarr"
)

/* This file contains the procedures to send stuck download queue items to notifiarr. */

const (
	errorstr  = "error"
	failed    = "failed"
	warning   = "warning"
	completed = "completed"
)

type custom struct {
	Repeat uint          `json:"repeat"`
	Queue  []interface{} `json:"queue"`
}

type ItemList map[int]custom

type QueuePayload struct {
	Type    string   `json:"type"`
	Lidarr  ItemList `json:"lidarr,omitempty"`
	Radarr  ItemList `json:"radarr,omitempty"`
	Readarr ItemList `json:"readarr,omitempty"`
	Sonarr  ItemList `json:"sonarr,omitempty"`
}

const getItemsMax = 100

func (i ItemList) Len() (count int) {
	for _, v := range i {
		count += len(v.Queue)
	}

	return count
}

func (i ItemList) Empty() bool {
	return i.Len() < 1
}

func (c *Config) SendFinishedQueueItems(url string) {
	q := &QueuePayload{
		Type:    "queue",
		Lidarr:  c.getFinishedItemsLidarr(),
		Radarr:  c.getFinishedItemsRadarr(),
		Readarr: c.getFinishedItemsReadarr(),
		Sonarr:  c.getFinishedItemsSonarr(),
	}

	if q.Lidarr.Empty() && q.Radarr.Empty() && q.Readarr.Empty() && q.Sonarr.Empty() {
		return
	}

	_, _, err := c.SendData(url+"/api/v1/user/client", q, true)
	if err != nil {
		c.Errorf("Sending Stuck Queue Items: %v", err)
		return
	}

	c.Printf("Sent Stuck Items: Lidarr: %d, Radarr: %d, Readarr: %d, Sonarr: %d",
		q.Lidarr.Len(), q.Radarr.Len(), q.Readarr.Len(), q.Sonarr.Len())
}

func (c *Config) getFinishedItemsLidarr() ItemList {
	stuck := make(ItemList)

	for i, l := range c.Apps.Lidarr {
		if l.CheckQ == nil {
			continue
		}

		queue, err := l.GetQueue(getItemsMax)
		if err != nil {
			c.Errorf("Getting Lidarr Queue (%d): %v", i, err)
			continue
		}

		for _, item := range queue.Records {
			if s := strings.ToLower(item.Status); s != completed && s != warning &&
				s != failed && s != errorstr && item.ErrorMessage == "" && len(item.StatusMessages) == 0 {
				continue
			}

			item.Quality = nil
			instance := stuck[i+1]
			instance.Repeat = *l.CheckQ
			instance.Queue = append(instance.Queue, item)
			stuck[i+1] = instance
		}

		c.Debugf("Checking Lidarr (%d) Queue for Stuck Items, queue size: %d, stuck: %d",
			i+1, len(queue.Records), len(stuck[i+1].Queue))
	}

	return stuck
}

func (c *Config) getFinishedItemsRadarr() ItemList {
	stuck := make(ItemList)

	for i, l := range c.Apps.Radarr {
		if l.CheckQ == nil {
			continue
		}

		queue, err := l.GetQueue(getItemsMax, 1)
		if err != nil {
			c.Errorf("Getting Radarr Queue (%d): %v", i, err)
			continue
		}

		for _, item := range queue.Records {
			if s := strings.ToLower(item.Status); s != completed && s != warning &&
				s != failed && s != errorstr && item.ErrorMessage == "" && len(item.StatusMessages) == 0 {
				continue
			}

			item.Quality = nil
			item.CustomFormats = nil
			item.Languages = nil
			instance := stuck[i+1]
			instance.Repeat = *l.CheckQ
			instance.Queue = append(instance.Queue, item)
			stuck[i+1] = instance
		}

		c.Debugf("Checking Radarr (%d) Queue for Stuck Items, queue size: %d, stuck: %d",
			i+1, len(queue.Records), len(stuck[i+1].Queue))
	}

	return stuck
}

func (c *Config) getFinishedItemsReadarr() ItemList {
	stuck := make(ItemList)

	for i, l := range c.Apps.Readarr {
		if l.CheckQ == nil {
			continue
		}

		queue, err := l.GetQueue(getItemsMax)
		if err != nil {
			c.Errorf("Getting Readarr Queue (%d): %v", i, err)
			continue
		}

		for _, item := range queue.Records {
			if s := strings.ToLower(item.Status); s != completed && s != warning &&
				s != failed && s != errorstr && item.ErrorMessage == "" && len(item.StatusMessages) == 0 {
				continue
			}

			item.Quality = nil
			instance := stuck[i+1]
			instance.Repeat = *l.CheckQ
			instance.Queue = append(instance.Queue, item)
			stuck[i+1] = instance
		}

		c.Debugf("Checking Readarr (%d) Queue for Stuck Items, queue size: %d, stuck: %d",
			i+1, len(queue.Records), len(stuck[i+1].Queue))
	}

	return stuck
}

func (c *Config) getFinishedItemsSonarr() ItemList {
	stuck := make(ItemList)

	for i, l := range c.Apps.Sonarr {
		if l.CheckQ == nil {
			continue
		}

		queue, err := l.GetQueue(getItemsMax, 1)
		if err != nil {
			c.Errorf("Getting Sonarr Queue (%d): %v", i, err)
			continue
		}

		// repeatStomper is used to collapse duplicate download IDs.
		repeatStomper := make(map[string]*sonarr.QueueRecord)

		for _, item := range queue.Records {
			if s := strings.ToLower(item.Status); s != completed && s != warning &&
				s != failed && s != errorstr && item.ErrorMessage == "" && len(item.StatusMessages) == 0 {
				continue
			} else if repeatStomper[item.DownloadID] != nil {
				continue
			}

			item.Quality = nil
			item.Language = nil
			repeatStomper[item.DownloadID] = item
			instance := stuck[i+1]
			instance.Repeat = *l.CheckQ
			instance.Queue = append(instance.Queue, item)
			stuck[i+1] = instance
		}

		c.Debugf("Checking Sonarr (%d) Queue for Stuck Items, queue size: %d, stuck: %d",
			i+1, len(queue.Records), len(stuck[i+1].Queue))
	}

	return stuck
}
