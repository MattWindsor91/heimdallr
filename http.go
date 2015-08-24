package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/UniversityRadioYork/bifrost-go"
)

func handleHTTPReq(w http.ResponseWriter, r *http.Request, resTree *bifrost.ResourceNoder) {
	if r.Method == "GET" {
		response := bifrost.Read(*resTree, r.URL.Path)
		if response.Status.Code != bifrost.StatusOk {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, response.Status.Message)
		} else {
			json, err := json.Marshal(response.Node)
			if err != nil {
				panic(err)
			}
			w.Write(json)
		}

	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func initHTTP(resourceTree bifrost.ResourceNoder, log *log.Logger) http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handleHTTPReq(w, r, &resourceTree)
	})
	return m
}

// dumpJSON dumps the JSON marshalling of res into w.
func dumpJSON(w io.Writer, res interface{}) (err error) {
	j, err := json.Marshal(res)

	if err == nil {
		_, err = w.Write(j)
	}

	return
}
