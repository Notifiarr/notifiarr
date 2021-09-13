// Package services provides service-checks to the notifiarr client application.
// This package spins up go routines to check http endpoints, running processes,
// tcp ports, etc. The configuration comes directly from the config file.
package services

import (
	"encoding/json"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/logs"
	"github.com/Notifiarr/notifiarr/pkg/notifiarr"
)

func (c *Config) Setup(services []*Service) (*notifiarr.ServiceConfig, error) {
	services = append(services, c.collectApps()...)
	if c.Disabled || len(services) == 0 {
		c.Disabled = true
		return &notifiarr.ServiceConfig{Disabled: true}, nil
	}

	if err := c.setup(services); err != nil {
		return nil, err
	}

	if c.LogFile != "" {
		c.Logger = logs.CustomLog(c.LogFile, "Services")
	}

	checks := make([]*notifiarr.ServiceCheck, len(services))
	for i, check := range services {
		checks[i] = &notifiarr.ServiceCheck{
			Name:     check.Name,
			Type:     string(check.Type),
			Expect:   check.Expect,
			Timeout:  check.Timeout,
			Interval: check.Interval,
		}

		if c.LogFile != "" && c.services[check.Name] != nil {
			c.services[check.Name].log = c.Logger
		}
	}

	return &notifiarr.ServiceConfig{
		Disabled: false,
		Interval: c.Interval,
		Parallel: int(c.Parallel),
		Checks:   checks,
	}, nil
}

// Start begins the service check routines.
// Runs Parallel checkers and the check reporter.
func (c *Config) Start() {
	if c.LogFile != "" {
		c.Printf("==> Service Checks Log File: %s", c.LogFile)
	}

	c.checks = make(chan *Service, DefaultBuffer)
	c.done = make(chan bool)
	c.stopChan = make(chan struct{})
	c.triggerChan = make(chan notifiarr.EventType)

	for i := uint(0); i < c.Parallel; i++ {
		go func() {
			defer c.CapturePanic()

			for check := range c.checks {
				if c.done == nil {
					return
				} else if check == nil {
					c.done <- false
					return
				}

				c.done <- check.check()
			}
		}()
	}

	go c.runServiceChecker()
	c.Printf("==> Service Checker Started! %d services, interval: %s, parallel: %d",
		len(c.services), c.Interval, c.Parallel)
}

func (c *Config) runServiceChecker() {
	ticker := time.NewTicker(c.Interval.Duration)
	second := time.NewTicker(10 * time.Second) //nolint:gomnd

	defer func() {
		c.CapturePanic()
		second.Stop()
		ticker.Stop()
	}()

	c.runChecks(true)
	c.SendResults(&Results{What: notifiarr.EventStart, Svcs: c.getResults()})

	for {
		select {
		case <-c.stopChan:
			for i := uint(0); i < c.Parallel; i++ {
				c.checks <- nil
				<-c.done
			}

			c.stopChan <- struct{}{}
			c.Printf("==> Service Checker Stopped!")

			return
		case <-ticker.C:
			c.SendResults(&Results{What: notifiarr.EventCron, Svcs: c.getResults()})
		case event := <-c.triggerChan:
			c.Debugf("Running all service checks via event: %s, buffer: %d/%d", event, len(c.checks), cap(c.checks))
			c.runChecks(true)

			if event != "log" {
				c.SendResults(&Results{What: event, Svcs: c.getResults()})
			} else {
				data, _ := json.MarshalIndent(&Results{Svcs: c.getResults(), Interval: c.Interval.Seconds()}, "", " ")
				c.Debug("Service Checks Payload (log only):", string(data))
			}
		case <-second.C:
			c.runChecks(false)
		}
	}
}

func (c *Config) setup(services []*Service) error {
	c.services = make(map[string]*Service)

	for i := range services {
		services[i].log = c.Logger
		if err := services[i].validate(); err != nil {
			return err
		}

		// Add this validated service to our service map.
		c.services[services[i].Name] = services[i]
	}

	if c.Parallel > MaximumParallel {
		c.Parallel = MaximumParallel
	} else if c.Parallel == 0 {
		c.Parallel = 1
	}

	if c.Interval.Duration == 0 {
		c.Interval.Duration = DefaultSendInterval
	} else if c.Interval.Duration < MinimumSendInterval {
		c.Interval.Duration = MinimumSendInterval
	}

	return nil
}

// Stop ends all service checker routines.
func (c *Config) Stop() {
	if c.stopChan == nil {
		return
	}

	defer close(c.triggerChan)
	defer close(c.stopChan)
	defer close(c.checks)
	defer close(c.done)

	c.triggerChan = nil
	c.stopChan <- struct{}{}
	<-c.stopChan
	c.checks = nil
	c.done = nil
	c.stopChan = nil
}
