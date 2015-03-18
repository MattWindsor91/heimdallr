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

func (c *bfConnector) get(resource string) interface{} {
	// TODO(CaptainHayashi): HTTP status codes

	// TODO(CaptainHayashi): probably peel off one layer at a
	// time (or use regexes), because this approach won't scale to
	// playlists due to the dynamic nature of the resources there.
	// Possibly https://github.com/ryanuber/go-glob
	switch resource {
	case "", "/":
		return GetOk(c.rootGet())
	case "/control", "/control/":
		return GetOk(c.controlGet())
	case "/control/state", "/control/state/":
		return GetOk(c.stateGet())
	case "/player", "/player/":
		return GetOk(c.playerGet())
	case "/player/time", "/player/time/":
		return GetOk(c.timeGet())
	case "/player/file", "/player/file/":
		return GetOk(c.fileGet())
	}

	return GetResponse{
		Status: "what",
		Value:  "resource not found: " + resource,
	}
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
	return struct {
		Time int64  `json:"time"`
		File string `json:"file,omitempty"`
	}{
		c.timeGet(),
		c.fileGet(),
	}
}

// GET value for /player/time
func (c *bfConnector) timeGet() int64 {
	// Time is reported in _micro_seconds
	return c.time.Nanoseconds() / 1000
}

// GET value for /player/file
func (c *bfConnector) fileGet() string {
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
