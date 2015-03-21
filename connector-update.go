package main

// The part of the BAPS3 connector code responsible for updating the
// internal state.

import (
	"fmt"
	"strconv"
	"time"

	"github.com/UniversityRadioYork/baps3-go"
)

func (c *bfConnector) update(res baps3.Message) {
	var err error

	switch res.Word() {
	case baps3.RsFeatures:
		err = c.updateFeaturesFromMessage(res)
	case baps3.RsFile:
		err = c.updateFileFromMessage(res)
	case baps3.RsState:
		err = c.updateStateFromMessage(res)
	case baps3.RsTime:
		err = c.updateTimeFromMessage(res)
	}
	if err != nil {
		fmt.Println(err)
	}
	c.updateCh <- res
}

func (c *bfConnector) updateFeaturesFromMessage(res baps3.Message) (err error) {
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

	c.features = feats
	return
}

func (c *bfConnector) updateFileFromMessage(res baps3.Message) (err error) {
	file, err := res.Arg(0)
	if err != nil {
		c.file = ""
		return
	}

	c.file = file

	return
}

func (c *bfConnector) updateStateFromMessage(res baps3.Message) (err error) {
	state, err := res.Arg(0)
	if err != nil {
		c.state = "???"
		return
	}

	c.state = state

	return
}

func (c *bfConnector) updateTimeFromMessage(res baps3.Message) (err error) {
	usecs, err := res.Arg(0)
	if err != nil {
		return
	}

	usec, err := strconv.Atoi(usecs)
	if err != nil {
		return
	}

	c.time = time.Duration(usec) * time.Microsecond

	return
}
