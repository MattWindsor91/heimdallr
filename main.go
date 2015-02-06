package main

import "os"
import "os/signal"
import "syscall"
import "fmt"
import "log"
import "io/ioutil"

import "github.com/BurntSushi/toml"

type Config struct {
	servers []string
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
	c1ReqCh := make(chan string)
	c1ResCh := make(chan string)
	c2ReqCh := make(chan string)
	c2ResCh := make(chan string)
	c1 := InitConnector(c1ReqCh, c1ResCh)
	c2 := InitConnector(c2ReqCh, c2ResCh)
	c1.Connect("localhost:1350")
	c2.Connect("localhost:1351")
	go c1.Run()
	go c2.Run()
	for {
		select {
		case data := <-c1ResCh:
			fmt.Println("C1: " + data)
		case data := <-c2ResCh:
			fmt.Println("C2: " + data)
		case <-sigs:
			fmt.Println("Exiting...")
			os.Exit(0)
		}
	}
}
