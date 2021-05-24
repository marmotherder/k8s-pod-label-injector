package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/jessevdk/go-flags"
)

var sOpts ServerOptions

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/hook", hook).Methods(http.MethodPost)
	parseArgs(&sOpts)
	log.Printf("Running server on port: %d\n", sOpts.ServerPort)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", sOpts.ServerPort),
		Handler: r,
	}

	log.Println(srv.ListenAndServeTLS(sOpts.TLSCertPath, sOpts.TLSKeyPath).Error())
}

// parseArgs parses the cli flags, allowing a common point to parse later downstream options when
// building config from multiple structs
func parseArgs(opts interface{}) {
	parser := flags.NewParser(opts, flags.IgnoreUnknown)
	_, err := parser.ParseArgs(os.Args[1:])
	if err != nil {
		log.Fatalln(err.Error())
	}
}
