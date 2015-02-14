package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type server struct {
	Hostport string
}

type Config struct {
	Servers map[string]server
}

func killConnectors(connectors []*Connector) {
	for _, c := range connectors {
		close(c.ReqCh)
	}
}

func main() {
	logger := log.New(os.Stdout, "[-] ", log.Lshortfile)
	var conf Config
	conffile, err := ioutil.ReadFile("conf_example.toml")
	if err != nil {
		logger.Fatal(err)
	}
	if _, err := toml.Decode(string(conffile), &conf); err != nil {
		logger.Fatal(err)
	}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT)

	resCh := make(chan string)

	connectors := []*Connector{}

	wg := new(sync.WaitGroup)

	for name, s := range conf.Servers {
		c := InitConnector(name, resCh, wg, logger)
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
