package services

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Custom errors.
var (
	ErrProcExpect = fmt.Errorf("invalid process expect type")
	ErrNoProcVal  = fmt.Errorf("process 'check' must not be empty")
	ErrCountZero  = fmt.Errorf("process 'count' may not be used with 'running'")
)

/* These all run once at startup to fill our check data. */

func (s *Service) checkProcValues() error {
	if s.Value == "" {
		return ErrNoProcVal
	} else if err := s.fillExpect(); err != nil {
		return err
	} else if err = s.fillExpectRegex(); err != nil {
		return err
	} else if (s.proc.countMin != 0 || s.proc.countMax != 0) && s.proc.running {
		return ErrCountZero
	}

	return nil
}

func (s *Service) fillExpect() (err error) {
	s.proc = &procExpect{}

	splitStr := strings.Split(s.Expect, ",")
	for _, str := range splitStr {
		switch {
		case strings.HasPrefix(str, "count:"): // "count:min:max" .. ie.  "count:1:3"
			if err := s.fillExpectCounts(str); err != nil {
				return err
			}
		case strings.EqualFold(str, "restart"):
			s.proc.restarts = true
		case strings.EqualFold(str, "running"):
			s.proc.running = true
		default:
			return fmt.Errorf("%w: %s", ErrProcExpect, str)
		}
	}

	return nil
}

// check Value for regex and attempt to compile it for later use.
func (s *Service) fillExpectRegex() (err error) {
	// Denote a regex by providing a string with slahes at each end.
	if s.Value[0] == '/' && len(s.Value) > 2 && s.Value[len(s.Value)-1] == '/' {
		s.proc.checkRE, err = regexp.Compile(s.Value[1 : len(s.Value)-1]) // strip slashes.
		if err != nil {
			return fmt.Errorf("invalid regex %s: %w", s.Value[1:len(s.Value)-1], err)
		}
	}

	return nil
}

func (s *Service) fillExpectCounts(str string) (err error) {
	countSplit := strings.Split(str, ":")
	if len(countSplit) > 1 {
		if s.proc.countMin, err = strconv.Atoi(countSplit[1]); err != nil {
			return fmt.Errorf("invalid minimum count: %s: %w", countSplit[1], err)
		}
	}

	if len(countSplit) > 2 { // nolint:gomnd
		if s.proc.countMax, err = strconv.Atoi(countSplit[2]); err != nil {
			return fmt.Errorf("invalid maximum count: %s: %w", countSplit[2], err)
		}
	}

	return nil
}
