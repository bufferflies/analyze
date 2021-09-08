// Copyright 2021 TiKV Project Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package server

import (
	"log"
	"net/http"

	"github.com/bufferflies/pd-analyze/repository"

	"github.com/bufferflies/pd-analyze/core"

	"github.com/bufferflies/pd-analyze/config"
	"github.com/gorilla/mux"
)

type Server struct {
	config  *config.Config
	source  core.Source
	checker core.Parser
	storage repository.Storage
}

func NewServer(config *config.Config) *Server {
	source := core.NewPrometheus(config.PrometheusAddress)
	checker := core.NewChecker(source)
	storage, err := repository.NewSqlite(config.StorageDir)
	if err != nil {
		log.Fatal("storage init failed", err)
	}
	return &Server{
		config:  config,
		source:  source,
		checker: checker,
		storage: storage,
	}
}

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Headers", "X-Requested-With, Content-Type,Origin, Authorization, Accept, Client-Security-Token, Accept-Encoding, X-Auth-Token, content-type")
		w.Header().Set("content-type", "application/json")
		if r.Method == http.MethodOptions {
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (server *Server) CreateRoute() *mux.Router {
	router := mux.NewRouter()
	analyze := NewPromAnalyze(server)
	router.HandleFunc("/analyze/{id}", analyze.AnalyzeSchedule).Methods(http.MethodPost)
	router.HandleFunc("/analyze/{id}", analyze.GetResult).Methods(http.MethodGet)
	router.Use(CORSMiddleware)

	return router
}
