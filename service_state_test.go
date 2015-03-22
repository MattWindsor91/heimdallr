package main

import (
	"fmt"
	"testing"

	"github.com/UniversityRadioYork/baps3-go"
)

// TestServiceStateUpdate tests the updating of a ServiceState by messages.
func TestServiceStateUpdate(t *testing.T) {
	// TODO(CaptainHayashi): test failure states as well as successes

	cases := []struct {
		msg *baps3.Message
		cmp func(*serviceState) error
	}{
		{
			baps3.NewMessage(baps3.RsFeatures).AddArg("End").AddArg("FileLoad"),
			func(s *serviceState) (err error) {
				_, endIn := s.features[FtEnd]
				_, flIn := s.features[FtFileLoad]

				if !endIn || !flIn {
					err = fmt.Errorf(
						"features should contain End and Fileload, got %d",
						s.features,
					)
				}

				return
			},
		},
		{
			baps3.NewMessage(baps3.RsFile).AddArg("/home/foo/bar.mp3"),
			func(s *serviceState) (err error) {
				if s.file != "/home/foo/bar.mp3" {
					err = fmt.Errorf(
						"file should be %d, got %d",
						"/home/foo/bar.mp3",
						s.file,
					)
				}

				return
			},
		},
		{
			baps3.NewMessage(baps3.RsState).AddArg("Ejected"),
			func(s *serviceState) (err error) {
				if s.state != "Ejected" {
					err = fmt.Errorf(
						"state should be %d, got %d",
						"Ejected",
						s.state,
					)
				}

				return
			},
		},
		{
			baps3.NewMessage(baps3.RsTime).AddArg("1337000000"),
			func(s *serviceState) (err error) {
				if s.time.Seconds() != 1337 {
					err = fmt.Errorf(
						"time should be %i secs, got %i",
						1337,
						s.time,
					)
				}

				return
			},
		},
	}

	for _, c := range cases {
		st := initServiceState()
		if err := st.update(*c.msg); err != nil {
			t.Errorf("error when sending %d: %s", c.msg, err)
		}
		if err := c.cmp(st); err != nil {
			t.Errorf("sent %d, but got error: %s", c.msg, err)
		}
	}

}
