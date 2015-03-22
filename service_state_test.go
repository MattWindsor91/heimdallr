package main

import (
	"fmt"
	"testing"

	"github.com/UniversityRadioYork/baps3-go"
)

// TestServiceStateUpdateFail tests the behaviour of a serviceState when it
// receives a malformed message.
func TestServiceStateUpdateFail(t *testing.T) {
	// TODO(CaptainHayashi): maybe test what the error actually is
	cases := []struct {
		msg    *baps3.Message
		hasErr bool
	}{
		// Request where response was expected
		{
			baps3.NewMessage(baps3.RqLoad).AddArg("/quux"),
			false, // TODO(CaptainHayashi): error on requests?
		},
		// Too few arguments
		{
			baps3.NewMessage(baps3.RsFile),
			true,
		},
		// Too many arguments
		{
			baps3.NewMessage(baps3.RsTime).AddArg("3003").AddArg("lol"),
			true,
		},
		// Unknown request (should be ignored)
		{
			baps3.NewMessage(baps3.RsUnknown).AddArg("heh"),
			false,
		},
	}

	for _, c := range cases {
		err := initServiceState().update(*c.msg)
		if c.hasErr && (err == nil) {
			t.Errorf("expected %q to produce error, none produced", c.msg)
		} else if !c.hasErr && (err != nil) {
			t.Errorf("expected %q not to produce error, one produced", c.msg)
		}
	}
}

// TestServiceStateUpdate tests the updating of a serviceState by messages.
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
