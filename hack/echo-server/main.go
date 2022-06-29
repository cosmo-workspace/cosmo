package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"

	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func main() {
	var port int
	flag.IntVar(&port, "port", 8888, "listen port")

	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	s := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: echoHandler(port),
	}
	log.Fatalln(s.ListenAndServe())
}

func echoHandler(port int) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		dump, err := httputil.DumpRequest(r, true)
		if err != nil {
			log.Println(err)
		} else {
			log.Println(string(dump))
		}
		buf := bytes.NewBuffer(dump)

		buf.WriteString(fmt.Sprintf("port: %d\n", port))
		buf.WriteString(fmt.Sprintf("remoteaddr: %s\n", r.RemoteAddr))

		_, err = io.Copy(rw, buf)
		if err != nil {
			log.Println(err)
		}
	})
}
