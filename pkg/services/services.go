// Package services provides service-checks to the notifiarr client application.
// This package spins up go routines to check http endpoints, running processes,
// tcp ports, etc. The configuration comes directly from the config file.
package services

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/exp"
	"github.com/Notifiarr/notifiarr/pkg/logs"
	"github.com/Notifiarr/notifiarr/pkg/notifiarr"
)

func (c *Config) Setup(services []*Service) (*notifiarr.ServiceConfig, error) {
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

	services = append(services, c.collectApps()...)

	return c.setup(services)
}

func (c *Config) setup(services []*Service) (*notifiarr.ServiceConfig, error) {
	c.services = make(map[string]*Service)
	scnfg := &notifiarr.ServiceConfig{
		Interval: c.Interval,
		Parallel: c.Parallel,
		Disabled: c.Disabled,
		Checks:   make([]*notifiarr.ServiceCheck, len(services)),
	}

	for idx, check := range services {
		if err := services[idx].Validate(); err != nil {
			return nil, err
		}

		exp.ServiceChecks.Add(check.Name+"&&Total", 0)
		exp.ServiceChecks.Add(check.Name+"&&"+StateUnknown.String(), 0)
		exp.ServiceChecks.Add(check.Name+"&&"+StateOK.String(), 0)
		exp.ServiceChecks.Add(check.Name+"&&"+StateWarning.String(), 0)
		exp.ServiceChecks.Add(check.Name+"&&"+StateCritical.String(), 0)

		// Add this validated service to our service map.
		c.services[services[idx].Name] = services[idx]
		scnfg.Checks[idx] = &notifiarr.ServiceCheck{
			Name:     check.Name,
			Type:     string(check.Type),
			Expect:   check.Expect,
			Timeout:  check.Timeout,
			Interval: check.Interval,
		}
	}

	return scnfg, nil
}

// Start begins the service check routines.
// Runs Parallel checkers and the check reporter.
func (c *Config) Start() {
	c.stopLock.Lock()
	defer c.stopLock.Unlock()

	if c.LogFile != "" {
		c.Logger = logs.CustomLog(c.LogFile, "Services")
	}

	for name := range c.services {
		c.services[name].svc.log = c.Logger
	}

	c.loadServiceStates()
	c.checks = make(chan *Service, DefaultBuffer)
	c.done = make(chan bool)
	c.stopChan = make(chan struct{})
	c.triggerChan = make(chan notifiarr.EventType)
	c.checkChan = make(chan triggerCheck)

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

	word := "Started"
	if c.Disabled {
		word = "Disabled"
	}

	c.Printf("==> Service Checker %s! %d services, interval: %s, parallel: %d",
		word, len(c.services), c.Interval, c.Parallel)
}

// loadServiceStates brings service states from the website into the fold.
// In other words, states are stored in the website's database.
func (c *Config) loadServiceStates() {
	names := []string{}
	for name := range c.services {
		names = append(names, valuePrefix+name)
	}

	values, err := c.Notifiarr.GetValue(names...)
	if err != nil {
		c.Errorf("Getting initial service states from Notifiarr.com: %v", err)
		return
	}

	for name := range c.services {
		for siteDataName := range values {
			if name == strings.TrimPrefix(siteDataName, valuePrefix) {
				var svc service
				if err := json.Unmarshal(values[siteDataName], &svc); err != nil {
					c.Errorf("Service check data for '%s' returned from site is invalid: %v", name, err)
					break
				}

				if time.Since(svc.LastCheck) < time.Hour {
					c.services[name].svc.Output = svc.Output
					c.services[name].svc.State = svc.State
					c.services[name].svc.Since = svc.Since
					c.services[name].svc.LastCheck = svc.LastCheck
				}

				break
			}
		}
	}
}

func (c *Config) runServiceChecker() {
	defer func() {
		defer c.CapturePanic()
		c.Printf("==> Service Checker Stopped!")
		c.stopChan <- struct{}{} // signal we're finished.
	}()

	ticker := &time.Ticker{C: make(<-chan time.Time)}
	second := &time.Ticker{C: make(<-chan time.Time)}

	if !c.Disabled {
		ticker = time.NewTicker(c.Interval.Duration)
		defer ticker.Stop()

		second = time.NewTicker(10 * time.Second) //nolint:gomnd
		defer second.Stop()

		c.runChecks(true)
		c.SendResults(&Results{What: notifiarr.EventStart, Svcs: c.GetResults()})
	}

	for {
		select {
		case <-c.stopChan:
			for i := uint(0); i < c.Parallel; i++ {
				c.checks <- nil
				<-c.done
			}

			return
		case <-ticker.C:
			c.SendResults(&Results{What: notifiarr.EventCron, Svcs: c.GetResults()})
		case event := <-c.checkChan:
			c.Debugf("Running service check '%s' via event: %s, buffer: %d/%d",
				event.Service.Name, event.Source, len(c.checks), cap(c.checks))
			c.runCheck(event.Service, true)
		case event := <-c.triggerChan:
			c.Debugf("Running all service checks via event: %s, buffer: %d/%d", event, len(c.checks), cap(c.checks))
			c.runChecks(true)

			if event != "log" {
				c.SendResults(&Results{What: event, Svcs: c.GetResults()})
			} else {
				// nolint:errchkjson
				data, _ := json.MarshalIndent(&Results{Svcs: c.GetResults(), Interval: c.Interval.Seconds()}, "", " ")
				c.Debug("Service Checks Payload (log only):", string(data))
			}
		case <-second.C:
			c.runChecks(false)
		}
	}
}

func (c *Config) Running() bool {
	c.stopLock.Lock()
	defer c.stopLock.Unlock()

	return c.stopChan != nil
}

// Stop ends all service checker routines.
func (c *Config) Stop() {
	c.stopLock.Lock()
	defer c.stopLock.Unlock()

	if c.stopChan == nil {
		return
	}

	defer close(c.stopChan)
	c.stopChan <- struct{}{}
	<-c.stopChan // wait for all go routines to die off.

	close(c.triggerChan)
	close(c.checkChan)
	close(c.checks)
	close(c.done)

	c.triggerChan = nil
	c.checkChan = nil
	c.checks = nil
	c.done = nil
	c.stopChan = nil
}
