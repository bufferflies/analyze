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
package repository

// save
type Storage interface {
	Save(id string, records []Record) error
	Get(id string) (records []Record, err error)
}

// Record log record
type Record struct {
	Workload string                      `json:"workload"`
	Start    string                      `json:"start_ts"`
	End      string                      `json:"end_ts"`
	Cmd      string                      `json:"bench_cmd"`
	Metrics  map[string]map[string]Index `json:"metrics"` //struct [metrics][operator]value
}

type Index struct {
	Min   float64   `json:"min"`
	Max   float64   `json:"max"`
	Mean  float64   `json:"mean"`
	Std   float64   `json:"std"`
	Data  []float64 `json:"data"`
	Start string    `json:"start_ts"`
	End   string    `json:"end_ts"`
}
