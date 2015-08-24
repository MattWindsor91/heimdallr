package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/BurntSushi/toml"
	"github.com/UniversityRadioYork/bifrost-go"
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
	HTTP    httpServer
}

func parseArgs() (args map[string]interface{}, err error) {
	usage := `heimdallr.

Usage:
  heimdallr [-c <configfile>]
  heimdallr -h
  heimdallr -v

Options:
  -c --config=<configfile>    Path to heimdallr config file [default: config.toml].
  -h --help                   Show this help message.
  -v --version                Show version.`

	args, err = docopt.Parse(usage, nil, true, "heimdallr 0.0", false)
	return
}

type ChannelWorker struct {
	name        string
	connector   *bifrost.Connector
	channelRoot bifrost.ResourceNoder
	resCh       chan bifrost.Message
	quit        chan struct{}
	logger      *log.Logger
}

func NewChannelWorker(
	channelName string,
	logger *log.Logger,
	channelRoot bifrost.ResourceNoder,
) *ChannelWorker {
	resCh := make(chan bifrost.Message)
	return &ChannelWorker{
		channelName,
		bifrost.InitConnector(channelName, resCh, logger),
		channelRoot,
		resCh,
		make(chan struct{}),
		logger,
	}
}

func (w *ChannelWorker) Run(hostport string) {
	w.connector.Connect(hostport)
	go w.connector.Run()
	for {
		select {
		case msg := <-w.resCh: // A wild message from downstream appears!
			w.processMsg(&msg)
		case <-w.quit:
			return
		}
	}
}

func (w *ChannelWorker) processMsg(msg *bifrost.Message) {
	w.logger.Println(w.name + ": " + msg.String())
	switch msg.Word() {
	case bifrost.RsRes:
		if args := msg.Args(); len(args) == 3 { // Nasty hack until update is implemented
			if args[1] == "Entry" {
				err := bifrost.Add(w.channelRoot, args[0], bifrost.NewEntryResourceNode(bifrost.BifrostTypeString(args[2])))
				if err != nil {
					panic(err)
				}
			}
		}
	case bifrost.RsOhai:
		//
	}
}

// Goodbye cruel world
func (w *ChannelWorker) Die() {
	w.quit <- struct{}{}
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

	resCh := make(chan bifrost.Message)

	workers := []*ChannelWorker{}

	resourceTree := bifrost.NewDirectoryResourceNode()

	for name, s := range conf.Servers {
		channelNode := bifrost.NewDirectoryResourceNode()
		if err := bifrost.Add(resourceTree, "/"+name, channelNode); err != nil { // Add the root channel node
			panic(err)
		}
		worker := NewChannelWorker(name, logger, &channelNode)
		workers = append(workers, worker)
		go worker.Run(s.Hostport)
	}
	r := initHTTP(resourceTree, logger)
	go http.ListenAndServe(conf.HTTP.Hostport, r)
	for {
		select {
		case data := <-resCh:
			fmt.Println(data.String())
		case <-sigs:
			logger.Println("Quitting...")
			for _, w := range workers {
				w.Die()
			}
			os.Exit(0)
		}
	}
}
