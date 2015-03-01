package main

import (
	"encoding/json"
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

type httpRequest struct {
	raw *http.Request

	resCh chan<- interface{}
}

func initHTTP(connectors []*bfConnector) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("static")))
	mux.HandleFunc("/ws", wsHandler)

	for i := range connectors {
		installConnector(mux, connectors[i])
	}

	return mux
}

func installConnector(mux *http.ServeMux, connector *bfConnector) {
	mux.HandleFunc("/"+connector.name, func(w http.ResponseWriter, r *http.Request) {
		resCh := make(chan interface{})

		fmt.Printf("sending request to %s\n", connector.name)

		w.Header().Add("Content-Type", "application/json")
		connector.reqCh <- httpRequest{
			r,
			resCh,
		}

		select {
		case res := <-resCh:
			j, err := json.Marshal(res)
			if err != nil {
				fmt.Println(err)
				break
			}
			_, err = w.Write(j)
			if err != nil {
				fmt.Println(err)
				break
			}
		}
	})
}
