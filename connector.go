package main

import (
	"bufio"
	"github.com/UniversityRadioYork/bifrost/baps3protocol"
	"github.com/UniversityRadioYork/bifrost/util"
	"log"
	"net"
	"sync"
	"time"
)

// Connector is a struct containing the internal state of a BAPS3 connector.
type Connector struct {
	state     string
	time      time.Duration
	tokeniser *baps3protocol.Tokeniser
	conn      net.Conn
	buf       *bufio.Reader
	resCh     chan<- string
	ReqCh     chan string
	name      string
	wg        *sync.WaitGroup
	logger    *log.Logger
}

// InitConnector creates and returns a Connector.
// The returned Connector shall have the given name, send responses through the
// response channel resCh, report termination via the wait group waitGroup, and
// log to logger.
func InitConnector(name string, resCh chan string, waitGroup *sync.WaitGroup, logger *log.Logger) *Connector {
	c := new(Connector)
	c.tokeniser = baps3protocol.NewTokeniser()
	c.resCh = resCh
	c.ReqCh = make(chan string)
	c.name = name
	c.wg = waitGroup
	c.logger = logger
	return c
}

// Connect connects an existing Connector to the BAPS3 server at hostport.
func (c *Connector) Connect(hostport string) {
	conn, err := net.Dial("tcp", hostport)
	if err != nil {
		c.logger.Fatal(err)
	}
	c.conn = conn
	c.buf = bufio.NewReader(c.conn)
}

// Run sets the given Connector off running.
func (c *Connector) Run() {
	lineCh := make(chan [][]string, 3)
	errCh := make(chan error)

	// Spin up a goroutine to accept and tokenise incoming bytes, and spit them
	// out in a channel
	go func(lineCh chan [][]string, eCh chan error) {
		for {
			data, err := c.buf.ReadBytes('\n')
			if err != nil {
				errCh <- err
			}
			lineCh <- c.tokeniser.Tokenise(data)
		}
	}(lineCh, errCh)

	// Main run loop, select on new received lines, errors or incoming requests
	for {
		select {
		case lines := <-lineCh:
			for _, line := range lines {
				switch line[0] {
				case "TIME":
					time, err := time.ParseDuration(line[1] + `us`)
					if err != nil {
						c.logger.Println(err)
					} else {
						c.time = time
						c.resCh <- c.name + ": " + util.PrettyDuration(time)
					}
				case "STATE":
					c.state = line[1]
					c.resCh <- c.name + ": " + line[1]
				}
			}
		case err := <-errCh:
			c.logger.Fatal(err)
		case _, ok := <-c.ReqCh:
			if !ok {
				c.logger.Println(c.name + " Connector shutting down")
				err := c.conn.Close()
				if err != nil {
					c.logger.Println(err)
				}
				c.wg.Done()
				return
			}
		}
	}
}
