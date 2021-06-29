package notifiarr

import (
	"strings"
)

/* This file contains the procedures to send stuck download queue items to notifiarr. */

type ItemList map[int][]interface{}

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
		count += len(v)
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

	_, _, body, err := c.SendData(url+"/api/v1/user/client", q)
	if err != nil {
		c.Errorf("Sending Stuck Queue Items: %v: %v", err, string(body))
		return
	}

	c.Printf("Sent Stuck Items: Lidarr: %d, Radarr: %d, Readarr: %d, Sonarr: %d",
		q.Lidarr.Len(), q.Radarr.Len(), q.Readarr.Len(), q.Sonarr.Len())
}

func (c *Config) getFinishedItemsLidarr() ItemList {
	stuck := make(ItemList)

	for i, l := range c.Apps.Lidarr {
		if !l.StuckItem {
			continue
		}

		queue, err := l.GetQueue(getItemsMax)
		if err != nil {
			c.Errorf("Getting Lidarr Queue (%d): %v", i, err)
			continue
		}

		for _, item := range queue.Records {
			if strings.EqualFold(item.Status, "completed") || len(item.StatusMessages) > 0 {
				item.Quality = nil
				stuck[i+1] = append(stuck[i+1], item)
			}
		}

		c.Debugf("Checking Lidarr (%d) Queue for Stuck Items, queue size: %d, stuck: %d",
			i+1, len(queue.Records), len(stuck[i+1]))
	}

	return stuck
}

func (c *Config) getFinishedItemsRadarr() ItemList {
	stuck := make(ItemList)

	for i, l := range c.Apps.Radarr {
		if !l.StuckItem {
			continue
		}

		queue, err := l.GetQueue(getItemsMax, 1)
		if err != nil {
			c.Errorf("Getting Radarr Queue (%d): %v", i, err)
			continue
		}

		for _, item := range queue.Records {
			if strings.EqualFold(item.Status, "completed") || len(item.StatusMessages) > 0 {
				item.Quality = nil
				item.CustomFormats = nil
				item.Languages = nil
				stuck[i+1] = append(stuck[i+1], item)
			}
		}

		c.Debugf("Checking Radarr (%d) Queue for Stuck Items, queue size: %d, stuck: %d",
			i+1, len(queue.Records), len(stuck[i+1]))
	}

	return stuck
}

func (c *Config) getFinishedItemsReadarr() ItemList {
	stuck := make(ItemList)

	for i, l := range c.Apps.Readarr {
		if !l.StuckItem {
			continue
		}

		queue, err := l.GetQueue(getItemsMax)
		if err != nil {
			c.Errorf("Getting Readarr Queue (%d): %v", i, err)
			continue
		}

		for j, item := range queue.Records {
			if strings.EqualFold(item.Status, "completed") || len(item.StatusMessages) > 0 {
				queue.Records[j].Quality = nil
				stuck[i+1] = append(stuck[i+1], &queue.Records[j])
			}
		}

		c.Debugf("Checking Readarr (%d) Queue for Stuck Items, queue size: %d, stuck: %d",
			i+1, len(queue.Records), len(stuck[i+1]))
	}

	return stuck
}

func (c *Config) getFinishedItemsSonarr() ItemList {
	stuck := make(ItemList)

	for i, l := range c.Apps.Sonarr {
		if !l.StuckItem {
			continue
		}

		queue, err := l.GetQueue(getItemsMax, 1)
		if err != nil {
			c.Errorf("Getting Sonarr Queue (%d): %v", i, err)
			continue
		}

		for _, item := range queue.Records {
			if strings.EqualFold(item.Status, "completed") || len(item.StatusMessages) > 0 {
				item.Quality = nil
				item.Language = nil
				stuck[i+1] = append(stuck[i+1], item)
			}
		}

		c.Debugf("Checking Sonarr (%d) Queue for Stuck Items, queue size: %d, stuck: %d",
			i+1, len(queue.Records), len(stuck[i+1]))
	}

	return stuck
}
