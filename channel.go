package main

import "time"
import "fmt"
import "os"

import "github.com/UniversityRadioYork/ury-rapid-go/baps3protocol"

type Channel struct {
	state     string
	time      time.Duration
	tokeniser *baps3protocol.Tokeniser
}

func InitChannel() *Channel {
	c := new(Channel)
	c.tokeniser = baps3protocol.NewTokeniser()
	return c
}

func (c *Channel) Update(data []byte) {
	lines := c.tokeniser.Tokenise(data)
	for _, line := range lines {
		switch line[0] {
		case "TIME":
			time, err := time.ParseDuration(line[1] + `us`)
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
