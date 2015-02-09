package main

import "os"
import "os/signal"
import "syscall"
import "fmt"
import "log"
import "io/ioutil"

import "github.com/BurntSushi/toml"

type server struct {
	Hostport string
}

type Config struct {
	Servers map[string]server
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

	for name, s := range conf.Servers {
		c := InitConnector(name, resCh)
		c.Connect(s.Hostport)
		go c.Run()
	}
	for {
		select {
		case data := <-resCh:
			fmt.Println(data)
		case <-sigs:
			fmt.Println("Exiting...")
			os.Exit(0)
		}
	}
}
