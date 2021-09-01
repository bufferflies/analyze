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

	"gonum.org/v1/gonum/floats"
	"gonum.org/v1/gonum/stat"

	"github.com/bufferflies/pd-analyze/core"
)

var (
	metrics = map[string]string{
		"tikv_cpu":   "sum(rate(tikv_thread_cpu_seconds_total{}[1m])) by (instance)",
		"tikv_write": "sum(rate(tikv_engine_flow_bytes{ db=\"kv\", type=\"wal_file_bytes\"}[1m])) by (instance)",
		"tikv_read":  "sum(rate(tikv_engine_flow_bytes{ db=\"kv\", type=~\"bytes_read|iter_bytes_read\"}[1m])) by (instance)",
	}
	operators = map[string]string{"mean": "mean(%s)", "std": "std(%s)"}
)

// Record log record
type Record struct {
	Start   string                      `json:"start_ts"`
	End     string                      `json:"end_ts"`
	Cmd     string                      `json:"bench_cmd"`
	Metrics map[string]map[string]index `json:"metrics"`
}

type index struct {
	Min  float64   `json:"min"`
	Max  float64   `json:"max"`
	Mean float64   `json:"mean"`
	Std  float64   `json:"std"`
	Data []float64 `json:"data"`
}

type PromAnalyze struct {
	checker *core.Checker
}

func NewPromAnalyze(address string) *PromAnalyze {
	source := core.NewPrometheus(address)

	return &PromAnalyze{
		checker: core.NewChecker(&source),
	}
}

// @Tags analyze
// @Summary analyze hot scheduler
// @Produce json
// @Success 200 {object} pdpb.GetMembersResponse
// @Failure 500 {string} string "PD server failed to proceed the request."
// @Router /scheduler/hot [get]
func (analyze *PromAnalyze) AnalyzeHotSchedule(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	w.Header().Set("content-type", "text/json")
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	var records []Record

	if err := json.Unmarshal(body, &records); err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}
	for i := range records {
		records[i].Metrics = make(map[string]map[string]index)
		if err := analyze.check(&records[i]); err != nil {
			fmt.Fprintf(w, err.Error())
			return
		}
	}
	rsp, err := json.Marshal(records)
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}
	fmt.Fprintf(w, string(rsp))
}

func (analyze *PromAnalyze) check(records *Record) error {
	for name, metrics := range metrics {
		records.Metrics[name] = make(map[string]index)
		for opName, m := range operators {
			index := records.Metrics[name][opName]
			d, err := analyze.checker.Apply(records.Start, records.End, name, metrics, fmt.Sprintf(m, name))
			if err != nil {
				return err
			}
			data := d.([]float64)
			index.Data = data
			index.Min = floats.Min(data)
			index.Max = floats.Max(data)
			index.Mean = stat.Mean(data, nil)
			index.Std = stat.StdDev(data, nil)
			records.Metrics[name][opName] = index
		}
	}
	return nil
}
