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

	"github.com/bufferflies/pd-analyze/config"
	"github.com/bufferflies/pd-analyze/core"
	"github.com/bufferflies/pd-analyze/repository"
	"github.com/gorilla/mux"
)

type Server struct {
	config          *config.Config
	source          core.Source
	checker         core.Parser
	projectStorage  repository.ProjectStorage
	workloadStorage repository.WorkloadStorage
}

func NewServer(config *config.Config) *Server {
	source := core.NewPrometheus(config.PrometheusAddress)
	checker := core.NewChecker(source)
	db, err := repository.NewMysqlManager(config.StorageAddress, "tinker")
	if err != nil {
		log.Fatal("storage init failed", err)
	}
	projectStorage := repository.NewProjectDao(db)
	workloadStorage := repository.NewWorkload(db, projectStorage)
	return &Server{
		config:          config,
		source:          source,
		checker:         checker,
		projectStorage:  projectStorage,
		workloadStorage: workloadStorage,
	}
}

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Headers", "X-Requested-With, Content-Type,Origin, Authorization, Accept, Client-Security-Token, Accept-Encoding, X-Auth-Token, content-type,Target")
		w.Header().Set("content-type", "application/json")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE,PUT")
		if r.Method == http.MethodOptions {
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (server *Server) CreateRoute() *mux.Router {
	router := mux.NewRouter()

	projectRouter := router.PathPrefix("/project").Subrouter()
	projectServer := NewProjectServer(server.projectStorage)
	projectRouter.HandleFunc("/", projectServer.GetProjects).Methods(http.MethodGet)
	projectRouter.HandleFunc("/new", projectServer.NewProject).Methods(http.MethodPost)
	projectRouter.HandleFunc("/session/new", projectServer.NewSession).Methods(http.MethodPost)
	projectRouter.HandleFunc("/session/{session_id}", projectServer.UpdateSession).Methods(http.MethodPost)
	projectRouter.HandleFunc("/sessions/{project_id}", projectServer.GetSessions).Methods(http.MethodGet)
	projectRouter.HandleFunc("/session/{session_id}", projectServer.GetSession).Methods(http.MethodGet)
	projectRouter.HandleFunc("/session/{session_id}", projectServer.DeleteSession).Methods(http.MethodDelete, http.MethodOptions)

	analyzeRouters := router.PathPrefix("/analyze").Subrouter()
	analyze := NewPromAnalyze(server)
	analyzeRouters.HandleFunc("/metrics/{session_id}", analyze.GetMetrics).Methods(http.MethodGet)
	analyzeRouters.HandleFunc("/config/{session_id}", analyze.GetWorkloadNames).Methods(http.MethodGet)
	analyzeRouters.HandleFunc("/workload/{session_id}", analyze.GetWorkloads).Methods(http.MethodGet)
	analyzeRouters.HandleFunc("/bench/{session_id}/{name}", analyze.GetBench).Methods(http.MethodGet)
	analyzeRouters.HandleFunc("/workload/{workload_id}", analyze.DeleteWorkloads).Methods(http.MethodDelete, http.MethodOptions)
	analyzeRouters.HandleFunc("/session/{session_id}", analyze.DeleteWorkloadByName).Methods(http.MethodDelete, http.MethodOptions)

	toolRouters := router.PathPrefix("/tools").Subrouter()
	tools := NewTools(server)
	toolRouters.HandleFunc("/{session_id}/{bench_name}", tools.AnalyzeSchedule).Methods(http.MethodPost)

	router.PathPrefix("/proxy/{path}").Handler(NewProxy())
	router.Use(CORSMiddleware)
	return router
}
