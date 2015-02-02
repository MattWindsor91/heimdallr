package main

import "os"
import "fmt"
import "net"
import "bufio"
import "math"

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:1350")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	c := InitChannel()
	for {
		data, err := bufio.NewReader(conn).ReadBytes('\n')
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		c.Update(data)
		fmt.Printf("%s: %02d:%02d\n", c.state, int(c.time.Minutes()), int(math.Mod(c.time.Seconds(), 60)))
	}
}
