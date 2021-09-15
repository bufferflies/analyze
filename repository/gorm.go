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

import (
	"fmt"
	"strconv"
	"time"

	"github.com/bufferflies/pd-analyze/errs"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Manager struct {
	db *gorm.DB
}

func NewMysqlManager(address string, database string) (*Manager, error) {
	dsn := fmt.Sprintf("root@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", address, database)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&Bench{}, &Workload{}, &Metrics{})
	return &Manager{
		db: db,
	}, err
}

func (m *Manager) Save(id string, records []Record) error {
	start := timeStampToTime(records[0].Start)
	end := timeStampToTime(records[len(records)-1].Start)
	bench := Bench{BenchID: id, Start: start, End: end}
	m.SaveBench(&bench)
	workloads := make([]*Workload, len(records))
	metrics := make([]*Metrics, 0)
	for i, v := range records {
		workloads[i] = &Workload{
			BenID: bench.ID,
			Name:  v.Workload,
			Start: timeStampToTime(v.Start),
			End:   timeStampToTime(v.End),
			Cmd:   v.Cmd,
		}
	}
	m.SaveWokload(workloads)
	for i, r := range records {
		for k, v := range r.Metrics {
			m := &Metrics{
				WID:   workloads[i].ID,
				Key:   k,
				Value: v,
			}
			metrics = append(metrics, m)
		}
	}
	m.SaveMetrics(metrics)
	return nil
}

func (m *Manager) Get(benchID string) (records []Record, err error) {
	var bench Bench
	m.db.Where(&Bench{BenchID: benchID}).First(&bench)
	if bench.ID == 0 {
		return nil, errs.Argument_Not_Match
	}
	var workloads []*Workload
	m.db.Where(&Workload{BenID: bench.ID}).Find(&workloads)
	result := make([]Record, len(workloads))
	for i, w := range workloads {
		var metrics []*Metrics
		m.db.Where(&Metrics{WID: w.ID}).Find(&metrics)
		ms := make(map[string]float64, len(metrics))
		for _, v := range metrics {
			ms[v.Key] = v.Value
		}
		result[i] = Record{
			Metrics:  ms,
			Start:    strconv.FormatInt(w.Start.Unix(), 10),
			End:      strconv.FormatInt(w.End.Unix(), 10),
			Cmd:      w.Cmd,
			Workload: w.Name,
		}
	}
	return result, err
}

func (m *Manager) SaveBench(bench *Bench) {
	m.db.Create(bench)

}

func (m *Manager) SaveWokload(workload []*Workload) {
	m.db.Create(workload)
}

func (m *Manager) SaveMetrics(metrics []*Metrics) {
	m.db.Create(metrics)
}

func timeStampToTime(stamp string) time.Time {
	r, _ := strconv.ParseInt(stamp, 10, 64)
	return time.Unix(r, 0)
}
