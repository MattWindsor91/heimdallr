package main

import "time"
import "fmt"
import "os"
import "net"
import "bufio"

import "github.com/UniversityRadioYork/ury-rapid-go/baps3protocol"

type Connector struct {
	state     string
	time      time.Duration
	tokeniser *baps3protocol.Tokeniser
	conn      net.Conn
	buf       *bufio.Reader
	resCh     chan<- string
	reqCh     <-chan string
}

func InitConnector(reqCh <-chan string, resCh chan<- string) *Connector {
	c := new(Connector)
	c.tokeniser = baps3protocol.NewTokeniser()
	c.resCh = resCh
	c.reqCh = reqCh
	return c
}

func (c *Connector) Connect(hostport string) {
	conn, err := net.Dial("tcp", hostport)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	c.conn = conn
	c.buf = bufio.NewReader(c.conn)
}

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
						fmt.Println(err)
						os.Exit(1)
					}
					c.time = time
					c.resCh <- PrettyDuration(time)
				case "STATE":
					c.state = line[1]
					c.resCh <- line[1]
				}
			}
		case err := <-errCh:
			fmt.Println(err)
			os.Exit(1)
		}
	}
}
