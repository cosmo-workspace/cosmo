package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
)

func main() {
	s := http.Server{
		Addr:    ":8888",
		Handler: echoHandler(),
	}
	log.Fatalln(s.ListenAndServe())
}

func echoHandler() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		dump, err := httputil.DumpRequest(r, true)
		if err != nil {
			log.Println(err)
		} else {
			log.Println(string(dump))
		}
		buf := bytes.NewBuffer(dump)
		_, err = io.Copy(rw, buf)
		if err != nil {
			log.Println(err)
		}
	})
}
