package main

// The part of the BAPS3 connector code responsible for updating the
// internal state.

import (
	"fmt"
	"strconv"
	"time"

	"github.com/UniversityRadioYork/baps3-go"
)

// serviceState is the struct of all known state for a BAPS3 service.
// TODO(CaptainHayashi): possibly segregate more by feature, so elements not
// relevant to the current feature set aren't allocated?
type serviceState struct {
	// Core
	features map[Feature]struct{}
	state    string

	// TimeReport
	time time.Duration

	// FileLoad
	file string
}

// initServiceState creates a new, blank, serviceState.
func initServiceState() (s *serviceState) {
	s = new(serviceState)
	s.features = make(map[Feature]struct{})

	return
}

// update updates a serviceState according to an incoming service response.
func (s *serviceState) update(res baps3.Message) (err error) {
	switch res.Word() {
	case baps3.RsFeatures:
		err = s.updateFeaturesFromMessage(res)
	case baps3.RsFile:
		err = s.updateFileFromMessage(res)
	case baps3.RsState:
		err = s.updateStateFromMessage(res)
	case baps3.RsTime:
		err = s.updateTimeFromMessage(res)
	}

	return
}

func (s *serviceState) updateFeaturesFromMessage(res baps3.Message) (err error) {
	feats := make(map[Feature]struct{})

	for i := 0; ; i++ {
		if fstring, e := res.Arg(i); e == nil {
			feat := LookupFeature(fstring)
			if feat == FtUnknown {
				err = fmt.Errorf("unknown feature: %q", fstring)
				break
			}
			feats[feat] = struct{}{}
		} else {
			// e != nil means we've run out of arguments.
			break
		}
	}

	s.features = feats
	return
}

func (s *serviceState) updateFileFromMessage(res baps3.Message) (err error) {
	file, err := res.Arg(0)
	if err != nil {
		s.file = ""
		return
	}

	s.file = file

	return
}

func (s *serviceState) updateStateFromMessage(res baps3.Message) (err error) {
	state, err := res.Arg(0)
	if err != nil {
		s.state = "???"
		return
	}

	s.state = state

	return
}

func (s *serviceState) updateTimeFromMessage(res baps3.Message) (err error) {
	usecs, err := res.Arg(0)
	if err != nil {
		return
	}

	usec, err := strconv.Atoi(usecs)
	if err != nil {
		return
	}

	s.time = time.Duration(usec) * time.Microsecond

	return
}

// hasFeature returns whether the connected server advertises the given feature.
func (s *serviceState) hasFeature(f Feature) bool {
	_, ok := s.features[f]
	return ok
}
