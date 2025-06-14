// Package services provides service-checks to the notifiarr client application.
// This package spins up go routines to check http endpoints, running processes,
// tcp ports, etc. The configuration comes directly from the config file.
package services

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/logs"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"github.com/gorilla/mux"
	"golift.io/version"
)

func (c *Config) Fix() {
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
}

func (s *Services) Add(services []ServiceConfig) error {
	for _, svc := range services {
		if !svc.validated {
			if err := svc.Validate(); err != nil {
				return err
			}
		}

		// Add this validated service to our service map.
		s.add(&svc)
	}

	return nil
}

func (s *Services) add(svc *ServiceConfig) {
	mnd.ServiceChecks.Add(svc.Name+"&&Total", 0)
	mnd.ServiceChecks.Add(svc.Name+"&&"+StateUnknown.String(), 0)
	mnd.ServiceChecks.Add(svc.Name+"&&"+StateOK.String(), 0)
	mnd.ServiceChecks.Add(svc.Name+"&&"+StateWarning.String(), 0)
	mnd.ServiceChecks.Add(svc.Name+"&&"+StateCritical.String(), 0)

	// Add this validated service to our service map.
	s.services[svc.Name] = &Service{
		ServiceConfig: svc,
	}
}

// Start begins the service check routines.
// Runs Parallel checkers and the check reporter.
func (s *Services) Start(ctx context.Context, plexName string) {
	if len(s.services) == 0 {
		mnd.Log.Printf("==> Service Checker Disabled! No services to check.")
		return
	}

	s.stopLock.Lock()
	defer s.stopLock.Unlock()

	s.checks = make(chan *Service, DefaultBuffer)
	s.done = make(chan bool)
	s.stopChan = make(chan struct{})
	s.triggerChan = make(chan website.EventType)
	s.checkChan = make(chan triggerCheck)

	logger := mnd.Log
	if s.LogFile != "" {
		logger = logs.CustomLog(s.LogFile, "Services")
	}

	for name := range s.services {
		s.services[name].log = logger
	}

	s.applyLocalOverrides(plexName)
	s.loadServiceStates(ctx)
	for range s.Parallel {
		go func() {
			defer mnd.Log.CapturePanic()

			for check := range s.checks {
				if s.done == nil {
					return
				} else if check == nil {
					s.done <- false
					return
				}

				s.done <- check.check(ctx)
			}
		}()
	}

	go s.runServiceChecker()

	word := "Started"
	if s.Disabled {
		word = "Disabled"
	}

	mnd.Log.Printf("==> Service Checker %s! %d services, interval: %s, parallel: %d",
		word, len(s.services), s.Interval, s.Parallel)
}

func (s *Services) applyLocalOverrides(plexName string) {
	if plexName == "" {
		return
	}

	// This is how we shoehorn the plex servr name into the service check.
	// We do this because we don't have the name when the config file is parsed.
	for _, svc := range s.services {
		if svc.Name == PlexServerName {
			if svc.Tags == nil {
				svc.Tags = map[string]any{}
			}

			svc.Tags["name"] = plexName

			return
		}
	}
}

// loadServiceStates brings service states from the website into the fold.
// In other words, states are stored in the website's database.
func (s *Services) loadServiceStates(ctx context.Context) {
	names := []string{}
	for name := range s.services {
		names = append(names, valuePrefix+name)
	}

	if len(names) == 0 {
		return
	}

	values, err := website.Site.GetState(ctx, names...)
	if err != nil {
		mnd.Log.ErrorfNoShare("Getting initial service states from website: %v", err)
		return
	}

	for name := range s.services {
		for siteDataName := range values {
			if name == strings.TrimPrefix(siteDataName, valuePrefix) {
				var svc Service
				if err := json.Unmarshal(values[siteDataName], &svc); err != nil {
					mnd.Log.ErrorfNoShare("Service check data for '%s' returned from site is invalid: %v", name, err)
					break
				}

				if time.Since(svc.LastCheck) < 2*time.Hour {
					mnd.Log.Printf("==> Set service state with website-saved data: %s, %s for %s",
						name, svc.State, time.Since(svc.Since).Round(time.Second))

					s.services[name].Output = svc.Output
					s.services[name].State = svc.State
					s.services[name].Since = svc.Since
					s.services[name].LastCheck = svc.LastCheck
				}

				break
			}
		}
	}
}

func (s *Services) runServiceChecker() { //nolint:cyclop
	defer func() {
		defer mnd.Log.CapturePanic()
		mnd.Log.Printf("==> Service Checker Stopped!")
		s.stopChan <- struct{}{} // signal we're finished.
	}()

	checker := &time.Ticker{C: make(<-chan time.Time)}

	if !s.Disabled {
		checker = time.NewTicker(time.Second)
		defer checker.Stop()

		s.runChecks(true, version.Started)
		s.SendResults(&Results{What: website.EventStart, Svcs: s.GetResults()})
	}

	for {
		select {
		case <-s.stopChan:
			for range s.Parallel {
				s.checks <- nil
				<-s.done
			}

			return
		case event := <-s.checkChan:
			mnd.Log.Printf("Running service check '%s' via event: %s, buffer: %d/%d",
				event.Service.Name, event.Source, len(s.checks), cap(s.checks))

			if s.runCheck(event.Service, true, time.Now()) {
				s.SendResults(&Results{What: event.Source, Svcs: s.GetResults()})
			}
		case event := <-s.triggerChan:
			mnd.Log.Debugf("Running all service checks via event: %s, buffer: %d/%d", event, len(s.checks), cap(s.checks))
			s.runChecks(true, time.Now())

			if event != "log" {
				s.SendResults(&Results{What: event, Svcs: s.GetResults()})
				continue
			}

			data, err := json.MarshalIndent(&Results{Svcs: s.GetResults(), Interval: s.Interval.Seconds()}, "", " ")
			if err != nil {
				mnd.Log.Errorf("Marshalling Service Checks: %v; payload: %s", err, string(data))
				continue
			}

			mnd.Log.Debug("Service Checks Payload (log only):", string(data))
		case now := <-checker.C:
			if s.runChecks(false, now) {
				s.SendResults(&Results{What: website.EventCron, Svcs: s.GetResults()})
			}
		}
	}
}

func (s *Services) Running() bool {
	s.stopLock.Lock()
	defer s.stopLock.Unlock()

	return s.stopChan != nil
}

// Stop sends current states to the website and ends all service checker routines.
func (s *Services) Stop() {
	s.stopLock.Lock()
	defer s.stopLock.Unlock()

	if s.stopChan == nil {
		return
	}

	defer close(s.stopChan)
	s.stopChan <- struct{}{}
	<-s.stopChan // wait for all go routines to die off.

	close(s.triggerChan)
	close(s.checkChan)
	close(s.checks)
	close(s.done)

	s.triggerChan = nil
	s.checkChan = nil
	s.checks = nil
	s.done = nil
	s.stopChan = nil
}

// SvcCount returns the count of services being monitored.
func (s *Services) SvcCount() int {
	return len(s.services)
}

// APIHandler is passed into the webserver so services can be accessed by the API.
func (s *Services) APIHandler(req *http.Request) (int, any) {
	return s.handleTrigger(req, website.EventAPI)
}

func (s *Services) handleTrigger(req *http.Request, event website.EventType) (int, any) {
	action := mux.Vars(req)["action"]
	mnd.Log.Debugf("[%s requested] Incoming Service Action: %s (%s)", event, action)

	switch action {
	case "list":
		return s.returnServiceList()
	default:
		return http.StatusBadRequest, "unknown service action: " + action
	}
}

// @Description  Returns a list of service check results.
// @Summary      Get service check results
// @Tags         Triggers
// @Produce      json
// @Success      200  {object} apps.Respond.apiResponse{message=[]CheckResult} "list check results"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/services/list [get]
// @Security     ApiKeyAuth
func (s *Services) returnServiceList() (int, any) {
	return http.StatusOK, s.GetResults()
}
