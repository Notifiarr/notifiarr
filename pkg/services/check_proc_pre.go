package services

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
)

// Custom errors.
var (
	ErrProcExpect = errors.New("invalid process expect type")
	ErrNoProcVal  = errors.New("process 'check' must not be empty")
	ErrCountZero  = errors.New("process 'count' may not be used with 'running'")
	ErrBSDRestart = errors.New("process 'restart' check does not work on FreeBSD") // one day.
)

/*
 * These all run once at startup to fill our check data.
 * The service Lock is acquired before running any of this code.
 */

func (s *Service) checkProcValues() error {
	if s.Value == "" {
		return ErrNoProcVal
	} else if err := s.fillExpect(); err != nil {
		return err
	} else if err = s.fillExpectRegex(); err != nil {
		return err
	} else if (s.svc.proc.countMin != 0 || s.svc.proc.countMax != 0) && s.svc.proc.running {
		return ErrCountZero
	}

	return nil
}

func (s *Service) fillExpect() error {
	s.svc.proc = &procExpect{}

	splitStr := strings.Split(s.Expect, ",")
	for _, str := range splitStr {
		switch {
		case strings.HasPrefix(str, "count:"): // "count:min:max" .. ie.  "count:1:3"
			if err := s.fillExpectCounts(str); err != nil {
				return err
			}
		case strings.EqualFold(str, "restart"):
			s.svc.proc.restarts = true

			if mnd.IsFreeBSD {
				return ErrBSDRestart
			}
		case strings.EqualFold(str, "running"):
			s.svc.proc.running = true
		case str == "":
			continue
		default:
			return fmt.Errorf("%s: %w: %s", s.Name, ErrProcExpect, str)
		}
	}

	return nil
}

// check Value for regex and attempt to compile it for later use.
func (s *Service) fillExpectRegex() error {
	// Denote a regex by providing a string with slahes at each end.
	if s.Value[0] == '/' && len(s.Value) > 2 && s.Value[len(s.Value)-1] == '/' {
		var err error

		s.svc.proc.checkRE, err = regexp.Compile(s.Value[1 : len(s.Value)-1]) // strip slashes.
		if err != nil {
			return fmt.Errorf("invalid regex %s: %w", s.Value[1:len(s.Value)-1], err)
		}
	}

	return nil
}

func (s *Service) fillExpectCounts(str string) error {
	var err error

	countSplit := strings.Split(str, ":")
	if len(countSplit) > 1 {
		if s.svc.proc.countMin, err = strconv.Atoi(countSplit[1]); err != nil {
			return fmt.Errorf("invalid minimum count: %s: %w", countSplit[1], err)
		}
	}

	if len(countSplit) > 2 { //nolint:mnd
		if s.svc.proc.countMax, err = strconv.Atoi(countSplit[2]); err != nil {
			return fmt.Errorf("invalid maximum count: %s: %w", countSplit[2], err)
		}
	}

	return nil
}
