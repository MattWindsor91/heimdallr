package main

import "time"
import "os"
import "os/signal"
import "syscall"
import "fmt"
import "math"

func main() {
	ticker := time.NewTicker(time.Second)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT)
	c1 := InitChannel("localhost:1350")
	c2 := InitChannel("localhost:1351")
	go c1.Run()
	go c2.Run()
	for {
		select {
		case <-ticker.C:
			fmt.Printf("C1: %s: %02d:%02d\n", c1.state, int(c1.time.Minutes()), int(math.Mod(c1.time.Seconds(), 60)))
			fmt.Printf("C2: %s: %02d:%02d\n", c2.state, int(c2.time.Minutes()), int(math.Mod(c2.time.Seconds(), 60)))
		case <-sigs:
			ticker.Stop()
			fmt.Println("Exiting...")
			os.Exit(0)
		}
	}
}
