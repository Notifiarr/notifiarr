// Package services provides service-checks to the notifiarr client application.
// This package spins up go routines to check http endpoints, running processes,
// tcp ports, etc. The configuration comes directly from the config file.
package services

import (
	"time"

	"github.com/Notifiarr/notifiarr/pkg/logs"
	"github.com/Notifiarr/notifiarr/pkg/notifiarr"
)

// Start begins the service check routines.
func (c *Config) Start(services []*Service) error {
	services = append(services, c.collectApps()...)
	if c.Disabled || len(services) == 0 {
		c.Disabled = true
		return nil
	} else if err := c.setup(services); err != nil {
		return err
	}

	c.start()

	return nil
}

// start runs Parallel checkers and the check reporter.
func (c *Config) start() {
	if c.LogFile != "" {
		c.Logger = logs.CustomLog(c.LogFile, "Services")
		c.Printf("==> Service Checks Log File: %s", c.LogFile)

		for i := range c.services {
			c.services[i].log = c.Logger
		}
	}

	for i := uint(0); i < c.Parallel; i++ {
		go func() {
			defer c.CapturePanic()

			for check := range c.checks {
				c.done <- check.check()
			}

			c.done <- false
		}()
	}

	go c.runServiceChecker()
	c.Printf("==> Service Checker Started! %d services, interval: %s", len(c.services), c.Interval)
}

func (c *Config) runServiceChecker() {
	ticker := time.NewTicker(c.Interval.Duration)
	second := time.NewTicker(time.Millisecond * 4159) //nolint:gomnd

	defer func() {
		c.CapturePanic()
		second.Stop()
		ticker.Stop()
		c.done <- false
	}()

	c.RunChecks(true)
	c.SendResults(notifiarr.ProdURL, &Results{
		What: "start",
		Svcs: c.GetResults(),
	})

	for {
		select {
		case <-second.C:
			c.RunChecks(false)
		case <-ticker.C:
			c.SendResults(notifiarr.ProdURL, &Results{
				What: "timer",
				Svcs: c.GetResults(),
			})
		case source := <-c.triggerChan:
			c.RunChecks(false)
			c.SendResults(notifiarr.ProdURL, &Results{
				What: source,
				Svcs: c.GetResults(),
			})
		case <-c.stopChan:
			return
		}
	}
}

func (c *Config) setup(services []*Service) error {
	c.services = make(map[string]*Service)
	c.checks = make(chan *Service, DefaultBuffer)
	c.done = make(chan bool)
	c.stopChan = make(chan struct{})
	c.triggerChan = make(chan string)

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

	close(c.triggerChan)
	c.triggerChan = nil

	close(c.stopChan)
	c.stopChan = nil
	<-c.done

	close(c.checks)
	c.checks = nil

	for i := uint(0); i < c.Parallel; i++ {
		<-c.done
	}

	close(c.done)
	c.done = nil
}
