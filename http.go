package main

import (
	"bufio"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var wsClients []*websocket.Conn

func wsbroadcast(msg string) {
	for _, ws := range wsClients {
		err := ws.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			fmt.Println(err)
			continue
		}
	}
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	wsClients = append(wsClients, conn)
	err = conn.WriteMessage(websocket.TextMessage, []byte("Connected"))
	if err != nil {
		fmt.Println(err)
		return
	}
}

func initHTTP(connectors []*bfConnector) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("static")))
	mux.HandleFunc("/ws", wsHandler)

	for i := range connectors {
		connector := connectors[i]
		mux.HandleFunc("/"+connector.name, func(w http.ResponseWriter, r *http.Request) {
			bio := bufio.NewWriter(w)

			resCh := make(chan string)

			fmt.Printf("sending request to %s\n", connector.name)

			connector.reqCh <- httpRequest{
				r,
				resCh,
			}

			select {
			case res := <-resCh:
				bio.WriteString(res + "\n")
				bio.Flush()
			}
		})
	}

	return mux
}
