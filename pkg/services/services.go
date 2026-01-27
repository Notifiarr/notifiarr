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

const (
	svcsPerThread = 10
	maxParallel   = 10
)

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
	s.stopLock.Lock()
	defer s.stopLock.Unlock()

	s.checks = make(chan *Service, DefaultBuffer)
	s.done = make(chan bool)
	s.actionChan = make(chan action)
	s.replyChan = make(chan bool)
	s.triggerChan = make(chan website.EventType)
	s.checkChan = make(chan triggerCheck)

	if s.parallel = uint(len(s.services) / svcsPerThread); s.parallel < 1 {
		s.parallel = 1
	} else if s.parallel > maxParallel {
		s.parallel = maxParallel
	}

	if s.log = mnd.Log; s.LogFile != "" {
		s.log = logs.CustomLog(s.LogFile, "Services")
	}

	for name := range s.services {
		s.services[name].log = s.log
	}

	s.applyLocalOverrides(plexName)
	s.loadServiceStates()

	for range s.parallel {
		go s.watchServiceChan(ctx)
	}

	go s.runServiceChecker()

	word := "Started"
	if s.Disabled || len(s.services) == 0 {
		word = "Disabled"
	}

	mnd.Log.Printf("==> Service Checker %s! %d services, parallel: %d",
		word, len(s.services), s.parallel)

	if s.log != mnd.Log {
		s.log.Printf("==> Service Checker %s! %d services, parallel: %d",
			word, len(s.services), s.parallel)
	}
}

func (s *Services) watchServiceChan(ctx context.Context) {
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
func (s *Services) loadServiceStates() {
	names := []string{}
	for name := range s.services {
		names = append(names, valuePrefix+name)
	}

	if len(names) == 0 {
		return
	}

	values, err := website.GetState(names...)
	if err != nil {
		s.log.ErrorfNoShare("Getting initial service states from website: %v", err)
		return
	}

	for name := range s.services {
		for siteDataName := range values {
			if name == strings.TrimPrefix(siteDataName, valuePrefix) {
				var svc Service
				if err := json.Unmarshal(values[siteDataName], &svc); err != nil {
					s.log.ErrorfNoShare("Service check data for '%s' returned from site is invalid: %v", name, err)
					break
				}

				if time.Since(svc.LastCheck) < 2*time.Hour {
					s.log.Printf("==> Set service state with website-saved data: %s, %s for %s",
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

// action is what we use to send actions to the service checker for loop.
type action int

const (
	actionStop   action = iota // happens on reload.
	actionPause                // user controlled pause.
	actionResume               // user controlled resume.
	actionCheck                // check if service checks are running.
)

func (s *Services) runServiceChecker() { //nolint:cyclop,funlen
	checker := time.NewTicker(time.Second)
	running := true

	defer func() {
		defer s.log.CapturePanic()
		checker.Stop()
		s.log.Printf("==> Service Checker Stopped!")
		s.actionChan <- actionStop // signal we're finished.
	}()

	if !s.Disabled {
		s.runChecks(true, version.Started)
		s.SendResults(&Results{What: website.EventStart, Svcs: s.GetResults()})
	} else {
		running = false
		checker.Stop()
	}

	for {
		select {
		case action := <-s.actionChan:
			switch action {
			case actionCheck:
				s.replyChan <- running
			case actionResume:
				s.log.Printf("==> Service Checker Resumed!")
				checker.Reset(time.Second)
				running = true
			case actionPause:
				s.log.Printf("==> Service Checker Paused!")
				checker.Stop()
				running = false
			case actionStop:
				// Stop all the checkers.
				for range s.parallel {
					s.checks <- nil
					<-s.done
				}

				return
			}
		case event := <-s.checkChan:
			s.log.Printf("Running service check '%s' via event: %s, buffer: %d/%d",
				event.Service.Name, event.Source, len(s.checks), cap(s.checks))

			if s.runCheck(event.Service, true, time.Now()) {
				s.SendResults(&Results{What: event.Source, Svcs: s.GetResults()})
			}
		case event := <-s.triggerChan:
			s.log.Debugf("Running all service checks via event: %s, buffer: %d/%d", event, len(s.checks), cap(s.checks))
			s.runChecks(true, time.Now())

			if event != "log" {
				s.SendResults(&Results{What: event, Svcs: s.GetResults()})
				continue
			}

			data, err := json.MarshalIndent(&Results{Svcs: s.GetResults()}, "", " ")
			if err != nil {
				s.log.Errorf("Marshalling Service Checks: %v; payload: %s", err, string(data))
				continue
			}

			s.log.Debug("Service Checks Payload (log only):", string(data))
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

	s.actionChan <- actionCheck
	return <-s.replyChan
}

func (s *Services) Pause() {
	s.stopLock.Lock()
	defer s.stopLock.Unlock()

	if s.actionChan != nil {
		s.actionChan <- actionPause
	}
}

func (s *Services) Resume() {
	s.stopLock.Lock()
	defer s.stopLock.Unlock()

	if s.actionChan != nil {
		s.actionChan <- actionResume
	}
}

// Stop sends current states to the website and ends all service checker routines.
func (s *Services) Stop() {
	defer func() {
		logs.Log.CapturePanic()
		logs.Log.Printf("==> Service Checker Stopped!")
	}()

	s.stopLock.Lock()
	defer s.stopLock.Unlock()

	if s.actionChan == nil {
		return
	}

	defer close(s.actionChan)
	s.actionChan <- actionStop
	<-s.actionChan // wait for all go routines to die off.

	close(s.triggerChan)
	close(s.checkChan)
	close(s.checks)
	close(s.done)

	s.triggerChan = nil
	s.checkChan = nil
	s.checks = nil
	s.done = nil
	s.actionChan = nil
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
	s.log.Debugf("[%s requested] Incoming Service Action: %s (%s)", event, action)

	switch action {
	case "list":
		return s.returnServiceList()
	default:
		return http.StatusBadRequest, "unknown service action: " + action
	}
}

// @Description	Returns a list of service check results.
// @Summary		Get service check results
// @Tags			Triggers
// @Produce		json
// @Success		200	{object}	apps.ApiResponse{message=[]CheckResult}	"list check results"
// @Failure		404	{object}	string									"bad token or api key"
// @Router			/services/list [get]
// @Security		ApiKeyAuth
func (s *Services) returnServiceList() (int, any) {
	return http.StatusOK, s.GetResults()
}
