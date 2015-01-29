package main

import "os"
import "fmt"
import "net"
import "bufio"
import "github.com/UniversityRadioYork/ury-rapid-go/rapid"

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:1350")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for {
		data, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf(data)
	}
}
