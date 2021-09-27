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
	"net/http"
	"strconv"

	"github.com/bufferflies/pd-analyze/repository"

	"github.com/bufferflies/pd-analyze/errs"
	"github.com/gorilla/mux"
)

type PromAnalyze struct {
	server *Server
}

func NewPromAnalyze(server *Server) *PromAnalyze {
	return &PromAnalyze{
		server: server,
	}
}

// @Tags analyze
// @Summary analyze hot scheduler
// @Produce json
// @Success 200 {object}
// @Failure 500 {string} string "PD server failed to proceed the request."
// @Router /analyze/{session_id}/{bench_name} [get]
func (analyze *PromAnalyze) GetWorkloadsByBenchName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["session_id"], 10, 10)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	benchName := vars["bench_name"]
	records, err := analyze.server.workloadStorage.GetWorkloadsByName(uint(id), benchName)
	if err != nil {
		fmt.Fprint(w, errs.Argument_Not_Match.Error())
		return
	}
	rsp, err := json.Marshal(records)
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	fmt.Fprint(w, string(rsp))
}

// @Tags analyze
// @Summary get analyze workload names and type
// @Produce json
// @Success 200 {object}
// @Failure 500 {string} string "PD server failed to proceed the request."
// @Router /analyze/config/{session_id} [get]
func (analyze *PromAnalyze) GetWorkloadNames(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sid, err := strconv.ParseUint(vars["session_id"], 10, 10)
	if err != nil {
		fmt.Fprint(w, errs.Argument_Not_Match.Error())
		return
	}

	result, err := analyze.server.workloadStorage.GetWorkloadNameAndVersion(uint(sid))
	if err != nil {
		fmt.Fprint(w, errs.Argument_Not_Match.Error())
		return
	}
	rsp, err := json.Marshal(result)
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	fmt.Fprint(w, string(rsp))
}

// @Tags analyze
// @Summary get analyze
// @Produce json
// @Success 200 {object}
// @Failure 500 {string} string "PD server failed to proceed the request."
// @Router /analyze//workload{session_id} [get]
func (analyze *PromAnalyze) GetWorkloads(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sid, err := strconv.ParseUint(vars["session_id"], 10, 10)
	if err != nil {
		fmt.Fprint(w, errs.Argument_Not_Match.Error())
		return
	}

	query := r.URL.Query()
	version := query.Get("version")
	page, err := strconv.Atoi(query.Get("page"))
	if err != nil {
		fmt.Fprint(w, errs.Argument_Not_Match.Error())
		return
	}
	size, err := strconv.Atoi(query.Get("size"))
	if err != nil {
		fmt.Fprint(w, errs.Argument_Not_Match.Error())
		return
	}

	workload := query.Get("workload")
	count, loads, err := analyze.server.workloadStorage.GetWorkload(workload, version, uint(sid), page, size)
	if err != nil {
		fmt.Fprint(w, errs.Argument_Not_Match.Error())
		return
	}
	result := make(map[string]interface{})
	result["workloads"] = loads
	result["count"] = count
	rsp, err := json.Marshal(result)
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	fmt.Fprint(w, string(rsp))
}

// @Tags analyze
// @Summary get analyze
// @Produce json
// @Success 200 {object}
// @Failure 500 {string} string "PD server failed to proceed the request."
// @Router /analyze/bench/{session_id}/{name}/ [get]
func (analyze *PromAnalyze) GetBench(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	sid, err := strconv.ParseUint(vars["session_id"], 10, 10)
	if err != nil {
		fmt.Fprint(w, errs.Argument_Not_Match.Error())
		return
	}
	loads, err := analyze.server.workloadStorage.GetWorkloadsByName(uint(sid), name)
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	result := make([]WorkloadMetrics, len(loads))
	for i, l := range loads {
		metrics, err := analyze.server.workloadStorage.GetMetricsByLoads(l.ID)
		if err != nil {
			fmt.Fprint(w, err.Error())
			return
		}
		result[i] = WorkloadMetrics{Workload: l, Metrics: metrics}
	}

	rsp, err := json.Marshal(result)
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	fmt.Fprint(w, string(rsp))
}

// @Tags analyze
// @Summary analyze hot scheduler
// @Produce json
// @Success 200 {object}
// @Failure 500 {string} string "PD server failed to proceed the request."
// @Router /analyze/getMetrics [get]
func (analyze *PromAnalyze) GetMetrics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sid, err := strconv.ParseUint(vars["session_id"], 10, 10)
	if err != nil {
		fmt.Fprint(w, errs.Argument_Not_Match.Error())
		return
	}
	err = r.ParseForm()
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	query := r.URL.Query()
	name := query.Get("workload")
	limit, err := strconv.Atoi(query.Get("limit"))
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}

	metrics := query["metrics"]
	rst, err := analyze.server.workloadStorage.GetMetricsBySid(uint(sid), name, limit, metrics)
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	body, err := json.Marshal(rst)
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	fmt.Fprint(w, string(body))
	return
}

type WorkloadMetrics struct {
	repository.Workload
	Metrics []repository.Metrics
}
