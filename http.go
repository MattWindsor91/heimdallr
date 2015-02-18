package main

import "net/http"
import "fmt"

import "github.com/gorilla/websocket"

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var wsClients []*websocket.Conn

func wsbroadcast(msg string) {
	for _, ws := range wsClients {
		ws.WriteMessage(websocket.TextMessage, []byte(msg))
	}
}

func WSHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	wsClients = append(wsClients, conn)
	conn.WriteMessage(websocket.TextMessage, []byte("Connected"))
}

func initHTTP() *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("static")))
	mux.HandleFunc("/ws", WSHandler)
	return mux
}
