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
	"strings"

	"github.com/bufferflies/pd-analyze/errs"
	"github.com/bufferflies/pd-analyze/repository"
	"github.com/gorilla/mux"
	"gonum.org/v1/gonum/floats"
	"gonum.org/v1/gonum/stat"
)

var (
	metrics = map[string]string{
		"tikv_cpu":   "sum(rate(tikv_thread_cpu_seconds_total{}[1m])) by (instance)",
		"tikv_write": "sum(rate(tikv_engine_flow_bytes{ db=\"kv\", type=\"wal_file_bytes\"}[1m])) by (instance)",
		"tikv_read":  "sum(rate(tikv_engine_flow_bytes{ db=\"kv\", type=~\"bytes_read|iter_bytes_read\"}[1m])) by (instance)",
	}
	operators = map[string]string{"mean": "mean(%s)", "std": "std(%s)"}
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
// @Success 200 {object} pdpb.GetMembersResponse
// @Failure 500 {string} string "PD server failed to proceed the request."
// @Router /analyze/{id} [Post]
func (analyze *PromAnalyze) AnalyzeSchedule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
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
	for i := range records {
		records[i].Metrics = make(map[string]float64)
		if err := analyze.check(&records[i]); err != nil {
			fmt.Fprint(w, err.Error())
			return
		}
	}
	err = analyze.server.storage.Save(id, records)
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	fmt.Fprint(w, "ok")
}

// @Tags analyze
// @Summary analyze hot scheduler
// @Produce json
// @Success 200 {object} pdpb.GetMembersResponse
// @Failure 500 {string} string "PD server failed to proceed the request."
// @Router /analyze/{id} [get]
func (analyze *PromAnalyze) GetResult(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if len(id) == 0 {
		fmt.Fprint(w, errs.Argument_Not_Match.Error())
		return
	}
	records, err := analyze.server.storage.Get(id)
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
// @Summary analyze hot scheduler
// @Produce json
// @Success 200 {object} pdpb.GetMembersResponse
// @Failure 500 {string} string "PD server failed to proceed the request."
// @Router /analyze/getAll [get]
func (analyze *PromAnalyze) GetAll(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	workload := query.Get("workload")
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
	loads, count, err := analyze.server.storage.GetAll(workload, version, page, size)
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
// @Summary analyze hot scheduler
// @Produce json
// @Success 200 {object} pdpb.GetMembersResponse
// @Failure 500 {string} string "PD server failed to proceed the request."
// @Router /analyze/getMetrics [get]
func (analyze *PromAnalyze) GetMetrics(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	query := r.URL.Query()
	workload := query.Get("workload")
	limit, err := strconv.Atoi(query.Get("limit"))
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}

	metrics := query["metrics"]
	rst, err := analyze.server.storage.GetMetrics(workload, limit, metrics)
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

func (analyze *PromAnalyze) check(records *repository.Record) error {
	for name, metrics := range metrics {
		records.Metrics = make(map[string]float64)
		for opName, m := range operators {
			prefix := strings.Join([]string{name, opName}, "_")
			d, err := analyze.server.checker.Apply(records.Start, records.End, name, metrics, fmt.Sprintf(m, name))
			if err != nil {
				return err
			}
			data := d.([]float64)
			records.Metrics[strings.Join([]string{prefix, "min"}, "_")] = floats.Min(data)
			records.Metrics[strings.Join([]string{prefix, "max"}, "_")] = floats.Max(data)
			records.Metrics[strings.Join([]string{prefix, "mean"}, "_")] = stat.Mean(data, nil)
			records.Metrics[strings.Join([]string{prefix, "std"}, "_")] = stat.StdDev(data, nil)
		}
	}
	return nil
}
