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
	"net/http"

	"github.com/bufferflies/pd-analyze/config"
	"github.com/gorilla/mux"
)

type Server struct {
	config *config.Config
}

func NewServer(config *config.Config) *Server {
	return &Server{
		config: config,
	}
}
func (server *Server) CreateRoute() *mux.Router {
	router := mux.NewRouter()
	analyze := NewPromAnalyze(server.config.PrometheusAddress)
	router.HandleFunc("/scheduler/hot", analyze.AnalyzeHotSchedule).Methods(http.MethodPost)

	return router
}
