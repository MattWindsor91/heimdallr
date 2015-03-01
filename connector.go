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
			switch res.Word() {
			case baps3.RsState:
				state, err := res.Arg(0)
				if err != nil {
					c.state = "???"
				} else {
					c.state = state
				}
			case baps3.RsTime:
				usecs, err := res.Arg(0)
				if err != nil {
					break
				}

				usec, err := strconv.Atoi(usecs)
				if err != nil {
					break
				}

				c.time = time.Duration(usec) * time.Microsecond
			case baps3.RsFile:
				file, err := res.Arg(0)
				if err != nil {
					c.file = ""
					break
				}

				c.file = file
			}
			c.updateCh <- res
		}
	}

	return
}
