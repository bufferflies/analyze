package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/bufferflies/pd-analyze/server"

	"github.com/bufferflies/pd-analyze/config"
)

func main() {
	config := config.NewConfig()
	server := server.NewServer(config)
	router := server.CreateRoute()
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.ListenPort), router))
}
