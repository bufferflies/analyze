package main

import (
	"fmt"
	"net/http"

	"github.com/bufferflies/pd-analyze/config"
	"github.com/bufferflies/pd-analyze/server"
	"github.com/pingcap/log"
)

func main() {
	config := config.NewConfig()
	server := server.NewServer(config)
	router := server.CreateRoute()
	log.Info("server start")
	if err := http.ListenAndServe(fmt.Sprintf(":%d", config.ListenPort), router); err != nil {
		log.Fatal("server run failed ")
	}
}
