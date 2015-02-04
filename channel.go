package main

import "time"
import "fmt"
import "os"
import "net"
import "bufio"

import "github.com/UniversityRadioYork/ury-rapid-go/baps3protocol"

type Channel struct {
	state     string
	time      time.Duration
	tokeniser *baps3protocol.Tokeniser
	conn      net.Conn
	buf       *bufio.Reader
}

func InitChannel(hostport string) *Channel {
	c := new(Channel)
	c.tokeniser = baps3protocol.NewTokeniser()
	conn, err := net.Dial("tcp", hostport)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	c.conn = conn
	c.buf = bufio.NewReader(c.conn)
	return c
}

func (c *Channel) Run() {
	for {
		data, err := c.buf.ReadBytes('\n')
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		lines := c.tokeniser.Tokenise(data)
		for _, line := range lines {
			switch line[0] {
			case "TIME":
				time, err := time.ParseDuration(line[1] + "us")
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				c.time = time
			case "STATE":
				c.state = line[1]
			}
		}
	}
}
