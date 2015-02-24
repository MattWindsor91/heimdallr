package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/BurntSushi/toml"
	baps3 "github.com/UniversityRadioYork/baps3-go"
)

type server struct {
	Hostport string
}

// Config is a struct containing the configuration for an instance of Bifrost.
type Config struct {
	Servers map[string]server
}

func killConnectors(connectors []*baps3.Connector) {
	for _, c := range connectors {
		close(c.ReqCh)
	}
}

func main() {
	logger := log.New(os.Stdout, "[-] ", log.Lshortfile)
	conffile, err := ioutil.ReadFile("conf_example.toml")
	if err != nil {
		logger.Fatal(err)
	}
	var conf Config
	if _, err := toml.Decode(string(conffile), &conf); err != nil {
		logger.Fatal(err)
	}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT)

	resCh := make(chan string)

	connectors := []*baps3.Connector{}

	wg := new(sync.WaitGroup)

	for name, s := range conf.Servers {
		c := baps3.InitConnector(name, resCh, wg, logger)
		connectors = append(connectors, c)
		c.Connect(s.Hostport)
		go c.Run()
	}
	wg.Add(len(connectors))

	for {
		select {
		case data := <-resCh:
			fmt.Println(data)
		case <-sigs:
			killConnectors(connectors)
			wg.Wait()
			logger.Println("Exiting...")
			os.Exit(0)
		}
	}
}
