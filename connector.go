package main

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/UniversityRadioYork/baps3-go"
)

type GetResponse struct {
	Status string
	Value  interface{}
}

type bfConnector struct {
	conn   *baps3.Connector
	name   string
	wg     *sync.WaitGroup
	logger *log.Logger

	// Cache of BAPS3 service internal state
	state  string
	time   time.Duration
	file   string

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
			fmt.Printf("connector %s response\n", c.name)
			rq.resCh <- GetResponse{
				Status: "ok",
				Value: struct {
					State string
					Time  int64
					File  string
				}{
					c.state,
					c.time.Nanoseconds() / 1000,
					c.file,
				},
			}
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
