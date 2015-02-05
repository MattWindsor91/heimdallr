package main

import "os"
import "os/signal"
import "syscall"
import "fmt"

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT)
	c1ReqCh := make(chan string)
	c1ResCh := make(chan string)
	c2ReqCh := make(chan string)
	c2ResCh := make(chan string)
	c1 := InitChannel("localhost:1350", c1ReqCh, c1ResCh)
	c2 := InitChannel("localhost:1351", c2ReqCh, c2ResCh)
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
