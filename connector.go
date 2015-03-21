package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/UniversityRadioYork/baps3-go"
)

type GetResponse struct {
	Status string
	Value  interface{}
}

func GetOk(value interface{}) *GetResponse {
	r := new(GetResponse)
	r.Status = "ok"
	r.Value = value
	return r
}

type bfConnector struct {
	conn   *baps3.Connector
	name   string
	wg     *sync.WaitGroup
	logger *log.Logger

	// Cache of BAPS3 service internal state
	state string
	time  time.Duration
	file  string

	reqCh chan httpRequest
	resCh <-chan baps3.Message

	// TODO(CaptainHayashi): move this away from baps3.Message to
	// something generic.
	updateCh chan<- baps3.Message
}

func initBfConnector(name string, updateCh chan baps3.Message, waitGroup *sync.WaitGroup, logger *log.Logger) (c *bfConnector) {
	resCh := make(chan baps3.Message)

	c = new(bfConnector)
	c.resCh = resCh
	c.conn = baps3.InitConnector(name, resCh, waitGroup, logger)
	c.name = name
	c.wg = waitGroup
	c.logger = logger
	c.reqCh = make(chan httpRequest)
	c.updateCh = updateCh

	return
}

func (c *bfConnector) Run() {
	defer c.wg.Done()
	defer close(c.conn.ReqCh)

	go c.conn.Run()

	fmt.Printf("connector %s now listening for requests\n", c.name)

	for {
		select {
		case rq, ok := <-c.reqCh:
			if !ok {
				return
			}

			// TODO(CaptainHayashi): probably make this more robust
			resource := strings.Replace(rq.resource, "/"+c.name, "", 1)
			fmt.Printf("connector %s response %s\n", c.name, resource)

			// TODO(CaptainHayashi): other methods
			rq.resCh <- c.get(resource)
		case res := <-c.resCh:
			var err error
			switch res.Word() {
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
	}

	return
}

// hasFeature returns whether the connected server advertises the given feature.
func (c *bfConnector) hasFeature(f Feature) bool {
	// TODO(CaptainHayashi): Actually check this
	return true
}

func splitResource(resource string) []string {
	res := strings.Split(strings.Trim(resource, "/"), "/")

	// The empty resource is returned as {""}: let's fix that
	if len(res) == 1 && res[0] == "" {
		res = []string{}
	}

	return res
}

func (c *bfConnector) get(resource string) interface{} {
	// TODO(CaptainHayashi): HTTP status codes

	resourcePath := splitResource(resource)

	if len(resourcePath) == 0 {
		return GetOk(c.rootGet())
	}

	var r interface{}

	switch resourcePath[0] {
	case "control":
		r = c.control(resourcePath[1:])
	case "player":
		r = c.player(resourcePath[1:])
		//case "playlist":
		//	r = c.playlist(resourcePath[1:])
	}

	if r == nil {
		// TODO(CaptainHayashi): more errors
		return GetResponse{
			Status: "what",
			Value:  "resource not found: " + resource,
		}
	}

	return GetOk(r)
}

// control is the main handler for the /control resource.
func (c *bfConnector) control(resourcePath []string) interface{} {
	if len(resourcePath) == 0 {
		return c.controlGet()
	}

	if len(resourcePath) == 1 {
		switch resourcePath[0] {
		case "state":
			return c.stateGet()
		}
	}

	return nil
}

// player is the main handler for the /player resource.
func (c *bfConnector) player(resourcePath []string) interface{} {
	if len(resourcePath) == 0 {
		return c.playerGet()
	}

	if len(resourcePath) == 1 {
		switch resourcePath[0] {
		case "time":
			return c.timeGet()
		case "file":
			return c.fileGet()
		}
	}

	return nil
}

// GET value for /
func (c *bfConnector) rootGet() interface{} {
	return struct {
		Control  interface{} `json:"control,omitempty"`
		Player   interface{} `json:"player,omitempty"`
		Playlist interface{} `json:"playlist,omitempty"`
	}{
		c.controlGet(),
		c.playerGet(),
		[]string{},
	}
}

// GET value for /control
func (c *bfConnector) controlGet() interface{} {
	return struct {
		State string `json:"state"`
	}{
		c.stateGet(),
	}
}

// GET value for /control/state
func (c *bfConnector) stateGet() string {
	return c.state
}

// GET value for /player
func (c *bfConnector) playerGet() interface{} {
	// TODO(CaptainHayashi): Probably a spec change, but the fact that this
	// resource is guarded by more than one feature is iffy.  Do we need a
	// Player feature?
	if !(c.hasFeature(FileLoad) || c.hasFeature(TimeReport)) {
		return nil
	}

	return struct {
		Time interface{} `json:"time"`
		File interface{} `json:"file,omitempty"`
	}{
		c.timeGet(),
		c.fileGet(),
	}
}

// GET value for /player/time
func (c *bfConnector) timeGet() interface{} {
	if !c.hasFeature(TimeReport) {
		return nil
	}

	// Time is reported in _micro_seconds
	return c.time.Nanoseconds() / 1000
}

// GET value for /player/file
func (c *bfConnector) fileGet() interface{} {
	if !c.hasFeature(FileLoad) {
		return nil
	}

	return c.file
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
