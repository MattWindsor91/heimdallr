package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/BurntSushi/toml"
	"github.com/UniversityRadioYork/baps3-go"
	"github.com/docopt/docopt-go"
)

type server struct {
	Hostport string
}

type httpServer struct {
	Hostport string
}

// Config is a struct containing the configuration for an instance of Bifrost.
type Config struct {
	Servers map[string]server
	Http httpServer
}

func killConnectors(connectors []*baps3.Connector) {
	for _, c := range connectors {
		close(c.ReqCh)
	}
}
func parseArgs() (args map[string]interface{}, err error) {
	usage := `bifrost.

Usage:
  bifrost [-c <configfile>]
  bifrost -h
  bifrost -v

Options:
  -c --config=<configfile>    Path to bifrost config file [default: config.toml].
  -h --help                   Show this help message.
  -v --version                Show version.`

	args, err = docopt.Parse(usage, nil, true, "bifrost 0.0", false)
	return
}

func main() {
	logger := log.New(os.Stdout, "[-] ", log.Lshortfile)
	args, err := parseArgs()
	if err != nil {
		logger.Fatal("Error parsing args: " + err.Error())
	}
	conffile, err := ioutil.ReadFile(args["--config"].(string))
	if err != nil {
		logger.Fatal(err)
	}
	var conf Config
	if _, err := toml.Decode(string(conffile), &conf); err != nil {
		logger.Fatal(err)
	}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT)

	resCh := make(chan baps3.Message)

	connectors := []*baps3.Connector{}

	wg := new(sync.WaitGroup)

	for name, s := range conf.Servers {
		c := baps3.InitConnector(name, resCh, wg, logger)
		connectors = append(connectors, c)
		c.Connect(s.Hostport)
		go c.Run()
	}
	wg.Add(len(connectors))

	mux := initHTTP()
	go func() {
		err := http.ListenAndServe(conf.Http.Hostport, mux)
		if err != nil {
			logger.Println(err)
		}
	}()

	for {
		select {
		case data := <-resCh:
			fmt.Println(data.String())
			wsbroadcast(data.String())
		case <-sigs:
			killConnectors(connectors)
			wg.Wait()
			logger.Println("Exiting...")
			os.Exit(0)
		}
	}
}
