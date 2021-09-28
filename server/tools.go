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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/bufferflies/pd-analyze/repository"
)

var (
	metrics = map[string]string{
		"tikv_cpu":   "sum(rate(tikv_thread_cpu_seconds_total{}[1m])) by (instance)",
		"tikv_write": "sum(rate(tikv_engine_flow_bytes{ db=\"kv\", type=\"wal_file_bytes\"}[1m])) by (instance)",
		"tikv_read":  "sum(rate(tikv_engine_flow_bytes{ db=\"kv\", type=~\"bytes_read|iter_bytes_read\"}[1m])) by (instance)",
	}
	operators = map[string]string{"mean": "mean(%s)", "std": "std(%s)"}
)

type Tools struct {
	server *Server
}

func NewTools(server *Server) *Tools {
	return &Tools{
		server: server,
	}
}

// @Tags analyze
// @Summary analyze  scheduler
// @Produce json
// @Success 200 {object}
// @Failure 500 {string} string "PD server failed to proceed the request."
// @Router /tools/{session_id}/{bench_name} [Post]
func (analyze *Tools) AnalyzeSchedule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sid, err := strconv.ParseInt(vars["session_id"], 10, 32)
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	benchName := vars["bench_name"]

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	var records []repository.Record

	if err := json.Unmarshal(body, &records); err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	err = analyze.server.workloadStorage.SaveRecords(uint(sid), benchName, records)
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	fmt.Fprint(w, "ok")
}
