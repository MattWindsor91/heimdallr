package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/UniversityRadioYork/baps3-go"
	"github.com/docopt/docopt-go"
)

// service contains options for configuring and running a baps3 service
type service struct {
	Name         string // Service name as given in a service definition title.
	Device       int    // Audio device number
	Host, Dshost string // Host to listen on and downstream host to connect to (if applicable)
	Port, Dsport int    // Port to listen on and downstream port to connect to (if applicable)
}

// httpConf is the section of config containing options for the bifrost HTTP server.
// A separate struct so that it can be passed around internally.
type httpconf struct {
	Listen string // Listen address (host:port) for http server
}

// svcdef is a service definition
// A separate struct so that it can be passed around internally.
type svcdef struct {
	// A command 'template' - a command with specifiers
	// representing possible configurations - e.g. `playd %d %h %p`.
	Command string
}

// Config is the parent struct containing the configuration, parsed from a file.
type Config struct {
	Svcdefs map[string]svcdef
	Groups  []struct {
		Name     string
		Services []service
	}

	HTTP httpconf
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

// commandFromService generates a command ready to run given a service definition
// (a string with command and arguments with specifiers e.g. %p) and a service
func commandFromService(svcdef svcdef, svc service) *exec.Cmd {
	// %d        device          audio device number
	// %h, %p    host, port      listen host and port
	// %H, %P    dshost, dsport  Downstream host and port
	r := strings.NewReplacer(
		`%d`, strconv.Itoa(svc.Device),
		`%h`, svc.Host,
		`%p`, strconv.Itoa(svc.Port),
		`%H`, svc.Dshost,
		`%P`, strconv.Itoa(svc.Dsport),
	)
	cmdslice := strings.Split(r.Replace(svcdef.Command), " ")
	return exec.Command(cmdslice[0], cmdslice[1:]...)
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

	connectors := []*bfConnector{}

	wg := new(sync.WaitGroup)

	serviceprocs := []*os.Process{}

	logger.Printf("%+v\n", conf)

	for _, group := range conf.Groups {
		for _, svc := range group.Services {
			svcdef, ok := conf.Svcdefs[svc.Name]
			if !ok {
				logger.Fatalf("Unknown service: %s", svc.Name)
			}
			svccmd := commandFromService(svcdef, svc)
			logger.Printf("Spawning command: %q", svccmd.Args)
			go func() {
				stdout, err := svccmd.StderrPipe()
				if err != nil {
					logger.Fatal(err)
				}
				if err := svccmd.Start(); err != nil {
					logger.Fatal(err)
				}
				wg.Add(1)
				serviceprocs = append(serviceprocs, svccmd.Process)
				scanner := bufio.NewScanner(stdout)
				for scanner.Scan() {
					logger.Printf("[%s] %s\n", svc.Name, scanner.Text())
				}
				if err := svccmd.Wait(); err != nil {
					logger.Println(err)
				}
				wg.Done()
			}()
			time.Sleep(1 * time.Second) // Give service time to start up, otherwise upstream connects will fail

		}
		c := initBfConnector(group.Name, resCh, wg, logger)
		connectors = append(connectors, c)
		lastsvc := group.Services[len(group.Services)-1] // Connects to the last svc in a group (top of stack)
		c.conn.Connect(fmt.Sprintf("%s:%d", lastsvc.Host, lastsvc.Port))
		go c.Run()
	}

	// Goroutine for the bifrost connector, and the lower-level
	// baps3-go connector.
	wg.Add(len(connectors) * 2)
	wspool := NewWspool(wg)
	initAndStartHTTP(conf.HTTP, connectors, wspool, logger)
	go wspool.run()

	for {
		select {
		case data := <-resCh:
			logger.Println(data.String())
			wspool.broadcast <- []byte(data.String())
		case <-sigs:
			logger.Println("Interrupt.")
			logger.Println("Killing connectors...")
			for _, c := range connectors {
				close(c.reqCh)
			}
			logger.Println("Asking services to quit...")
			// Quit here
			close(wspool.broadcast)
			wg.Wait()
			logger.Println("Sayonara")
			os.Exit(0)
		}
	}
}

func initAndStartHTTP(conf httpconf, connectors []*bfConnector, wspool *Wspool, logger *log.Logger) {
	mux := initHTTP(connectors, wspool, logger)
	go func() {
		err := http.ListenAndServe(conf.Listen, mux)
		if err != nil {
			logger.Println(err)
		}
	}()
}
