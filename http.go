package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type httpRequest struct {
	// TODO(CaptainHayashi): method, payload
	resource string
	resCh    chan<- interface{}
}

func initHTTP(connectors []*bfConnector, wspool *Wspool, log *log.Logger) http.Handler {
	r := mux.NewRouter()
	r.Handle("/", http.FileServer(http.Dir("static")))
	r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", 405)
			return
		}
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		c := &wsConn{send: make(chan []byte, 256), ws: ws}
		wspool.register <- c
		c.writeLoop()
	})

	for i := range connectors {
		installConnector(r, connectors[i])
	}

	return r
}

func installConnector(router *mux.Router, connector *bfConnector) {
	fn := func(w http.ResponseWriter, r *http.Request) {
		resCh := make(chan interface{})

		fmt.Printf("sending request to %s\n", connector.name)

		resource := r.URL.Path

		w.Header().Add("Content-Type", "application/json")
		connector.reqCh <- httpRequest{
			resource,
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
	}

	router.HandleFunc("/"+connector.name, fn)
	router.PathPrefix("/" + connector.name + "/").HandlerFunc(fn)
}
