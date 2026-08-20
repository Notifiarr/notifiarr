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
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"github.com/gorilla/mux"
	"golift.io/version"
)

const (
	svcsPerThread = 10
	maxParallel   = 10
	actionBuf     = 2
	controlBuf    = 1
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
		State:         StateUnknown,
		Since:         time.Now(),
	}
}

func (s *Services) setParallel() {
	count := len(s.services)
	if count == 0 {
		s.parallel = 0
		return
	}

	s.parallel = uint(count / svcsPerThread)
	if s.parallel < 1 {
		s.parallel = 1
	} else if s.parallel > maxParallel {
		s.parallel = maxParallel
	}
}

// Start begins the service check routines.
// Runs Parallel checkers and the check reporter.
func (s *Services) Start(ctx context.Context, plexName string) {
	s.setParallel()

	if s.log = mnd.Log; s.LogFile != "" {
		s.log = logs.CustomLog(s.LogFile, "Services")
	}

	for name := range s.services {
		s.services[name].log = s.log
	}

	s.applyLocalOverrides(plexName)
	s.loadServiceStates(mnd.GetID(ctx))

	ctx, cancel := context.WithCancel(ctx)

	s.stopLock.Lock()
	s.cancel = cancel
	s.stopped = make(chan struct{})
	s.checks = make(chan *Service, DefaultBuffer)
	s.done = make(chan bool, s.parallel)
	s.actionChan = make(chan action, actionBuf)
	s.replyChan = make(chan bool, controlBuf)
	s.triggerChan = make(chan *common.ActionInput, controlBuf)
	s.checkChan = make(chan triggerCheck, controlBuf)

	for range s.parallel {
		go s.watchServiceChan(ctx)
	}

	go s.runServiceChecker(ctx)
	s.stopLock.Unlock()

	word := "Started"
	if s.Disabled || len(s.services) == 0 {
		word = "Disabled"
	}

	mnd.Log.Printf(mnd.GetID(ctx), "==> Service Checker %s! %d services, parallel: %d",
		word, len(s.services), s.parallel)

	if s.log != mnd.Log {
		s.log.Printf(mnd.GetID(ctx), "==> Service Checker %s! %d services, parallel: %d",
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
func (s *Services) loadServiceStates(reqID string) {
	names := make([]string, 0, len(s.services))
	for name := range s.services {
		names = append(names, valuePrefix+name)
	}

	if len(names) == 0 {
		return
	}

	values, err := website.GetState(reqID, names...)
	if err != nil {
		s.log.ErrorfNoShare(reqID, "Getting initial service states from website: %v", err)
		return
	}

	saved := make(map[string][]byte, len(values))
	for siteDataName, data := range values {
		saved[strings.TrimPrefix(siteDataName, valuePrefix)] = data
	}

	for name := range s.services {
		data, ok := saved[name]
		if !ok {
			continue
		}

		var svc Service
		if err := json.Unmarshal(data, &svc); err != nil {
			s.log.ErrorfNoShare(reqID, "Service check data for '%s' returned from site is invalid: %v", name, err)
			continue
		}

		if time.Since(svc.LastCheck) < 2*time.Hour {
			s.log.Printf(reqID, "==> Set service state with website-saved data: %s, %s for %s",
				name, svc.State, time.Since(svc.Since).Round(time.Second))

			s.services[name].Output = svc.Output
			s.services[name].State = svc.State
			s.services[name].Since = svc.Since
			s.services[name].LastCheck = svc.LastCheck
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

func (s *Services) stopWorkers() {
	for range s.parallel {
		s.checks <- nil
		<-s.done
	}
}

func (s *Services) runServiceChecker(ctx context.Context) { //nolint:cyclop,funlen
	checker := time.NewTicker(time.Second)
	running := true
	reqID := mnd.ReqID()

	defer func() {
		defer s.log.CapturePanic()
		checker.Stop()
		s.log.Printf(reqID, "==> Service Checker Stopped!")
		close(s.stopped)
	}()

	if !s.Disabled {
		s.runChecks(true, version.Started)

		if ctx.Err() == nil {
			s.SendResults(&Results{What: website.EventStart, Svcs: s.GetResults()}, reqID)
		}
	} else {
		running = false
		checker.Stop()
	}

	for {
		select {
		case <-ctx.Done():
			s.stopWorkers()
			return
		case action := <-s.actionChan:
			switch action {
			case actionCheck:
				s.replyChan <- running
			case actionResume:
				s.log.Printf(reqID, "==> Service Checker Resumed!")
				checker.Reset(time.Second)
				running = true
			case actionPause:
				s.log.Printf(reqID, "==> Service Checker Paused!")
				checker.Stop()
				running = false
			case actionStop:
				s.stopWorkers()
				return
			}
		case event := <-s.checkChan:
			s.log.Printf(event.ReqID, "Running service check '%s' via event: %s, buffer: %d/%d",
				event.Service.Name, event.Source, len(s.checks), cap(s.checks))

			if s.runCheck(event.Service, true, time.Now()) {
				s.SendResults(&Results{What: event.Source, Svcs: s.GetResults()}, event.ReqID)
			}
		case input := <-s.triggerChan:
			s.log.Printf(input.ReqID, "Running all service checks via event: %s, buffer: %d/%d",
				input.Type, len(s.checks), cap(s.checks))
			s.runChecks(true, time.Now())

			if input.Type != "log" {
				s.SendResults(&Results{What: input.Type, Svcs: s.GetResults()}, input.ReqID)
				continue
			}

			data, err := json.MarshalIndent(&Results{Svcs: s.GetResults()}, "", " ")
			if err != nil {
				s.log.Errorf(input.ReqID, "Marshalling Service Checks: %v; payload: %s", err, string(data))
				continue
			}

			s.log.Printf(input.ReqID, "Service Checks Payload (log only): %s", string(data))
		case now := <-checker.C:
			if s.runChecks(false, now) {
				s.SendResults(&Results{What: website.EventCron, Svcs: s.GetResults()}, reqID)
			}
		}
	}
}

func (s *Services) Running() bool {
	s.stopLock.Lock()
	actionChan := s.actionChan
	replyChan := s.replyChan
	stopping := s.stopping
	s.stopLock.Unlock()

	if actionChan == nil || stopping {
		return false
	}

	actionChan <- actionCheck

	return <-replyChan
}

func (s *Services) Pause() {
	s.sendAction(actionPause)
}

func (s *Services) Resume() {
	s.sendAction(actionResume)
}

func (s *Services) sendAction(act action) {
	s.stopLock.Lock()
	actionChan := s.actionChan
	stopping := s.stopping
	s.stopLock.Unlock()

	if actionChan != nil && !stopping {
		actionChan <- act
	}
}

// Stop ends all service checker routines.
func (s *Services) Stop() {
	defer logs.Log.CapturePanic()

	s.stopLock.Lock()
	if s.actionChan == nil || s.stopping {
		s.stopLock.Unlock()
		return
	}

	s.stopping = true
	cancel := s.cancel
	stopped := s.stopped
	s.stopLock.Unlock()

	if cancel != nil {
		cancel()
	}

	if stopped != nil {
		<-stopped
	}

	s.stopLock.Lock()
	defer s.stopLock.Unlock()

	s.actionChan = nil
	s.triggerChan = nil
	s.checkChan = nil
	s.cancel = nil
	s.stopped = nil
	s.stopping = false

	if s.checks != nil {
		close(s.checks)
		s.checks = nil
	}

	if s.done != nil {
		close(s.done)
		s.done = nil
	}
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
	s.log.Debugf(mnd.GetID(req.Context()), "[%s requested] Incoming Service Action: %s (%s)", event, action)

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
