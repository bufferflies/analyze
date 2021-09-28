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
	"strconv"
	"time"

	"gorm.io/gorm"
)

type WorkloadStorage interface {
	SaveRecords(sessionID uint, benchName string, record []Record) error

	GetWorkloadsByName(sessionID uint, benchName string) ([]Workload, error)
	GetWorkloadsBySessionID(sessionID uint) ([]Workload, error)
	GetWorkloadNameAndVersion(sessionID uint) ([]Workload, error)

	GetWorkload(workload, version string, sessionID uint, page, size int) (int64, []Workload, error)
	GetWorkloadByID(id uint) (Workload, error)
	DeleteWorkload(wID uint) error
	DeleteWorkloadByName(sID uint, name string) error

	GetMetrics(workload uint, limit int, metrics []string) (map[string][]Metrics, error)
	GetMetricsBySid(sid uint, workload string, limit int, metrics []string) (map[string][]Metrics, error)
	GetMetricsByLoads(wIDs uint) ([]Metrics, error)
}

type WorkloadDao struct {
	db      *gorm.DB
	project ProjectStorage
}

func NewWorkload(db *gorm.DB, project ProjectStorage) WorkloadStorage {
	db.AutoMigrate(&Workload{}, &Metrics{})
	return &WorkloadDao{db: db, project: project}
}

func (p *WorkloadDao) GetWorkload(workload string, version string, sessionID uint, page, size int) (int64, []Workload, error) {
	m := p.db.Model(&Workload{}).Where(&Workload{SessionID: sessionID, Name: workload})
	var total int64
	if me := m.Count(&total); me.Error != nil {
		return 0, nil, me.Error
	}
	var workloads []Workload
	offset := (page - 1) * size
	if me := m.Offset(offset).Limit(size).Find(&workloads); me.Error != nil {
		return 0, nil, me.Error
	}
	return total, workloads, nil
}

func (p *WorkloadDao) SaveRecords(sessionID uint, benchName string, records []Record) error {
	workloads := make([]*Workload, len(records))
	metrics := make([]*Metrics, 0)
	session, err := p.project.GetSession(sessionID)
	if err != nil {
		return err
	}
	for i, v := range records {
		workloads[i] = &Workload{
			SessionID:    sessionID,
			Name:         v.Workload,
			Start:        timeStampToTime(v.Start),
			End:          timeStampToTime(v.End),
			Cmd:          v.Cmd,
			TargetObject: getTarget(session.TargetObject, v),
			BenchName:    benchName,
		}
	}
	m := p.db.Save(workloads)
	if m.Error != nil {
		return m.Error
	}
	for i, r := range records {
		for k, v := range r.Metrics {
			m := &Metrics{
				WID:       workloads[i].ID,
				Key:       k,
				Value:     v,
				Start:     workloads[i].Start,
				SessionID: sessionID,
				Name:      r.Workload,
			}
			metrics = append(metrics, m)
		}
	}
	m = p.db.Save(metrics)
	return m.Error
}

func (p *WorkloadDao) GetWorkloadsByName(sessionID uint, benchName string) ([]Workload, error) {
	var workloads []Workload
	m := p.db.Where(&Workload{SessionID: sessionID, BenchName: benchName}).Find(&workloads)
	return workloads, m.Error
}

func (p *WorkloadDao) GetWorkloadsBySessionID(sessionID uint) ([]Workload, error) {
	var workloads []Workload
	m := p.db.Where(&Workload{SessionID: sessionID}).Find(&workloads)
	return workloads, m.Error
}

func (p *WorkloadDao) GetWorkloadByID(id uint) (Workload, error) {
	var workload Workload
	m := p.db.Where(&Workload{ID: id}).Find(&workload)
	return workload, m.Error
}

func (p *WorkloadDao) DeleteWorkload(wID uint) error {
	if m := p.db.Where(&Workload{ID: wID}).Delete(&Workload{}); m.Error != nil {
		return m.Error
	}
	if m := p.db.Where(&Metrics{WID: wID}).Delete(&Metrics{}); m.Error != nil {
		return m.Error
	}
	return nil
}

func (p *WorkloadDao) DeleteWorkloadByName(sID uint, name string) error {
	if m := p.db.Where(&Workload{SessionID: sID, Name: name}).Delete(&Workload{}); m.Error != nil {
		return m.Error
	}
	if m := p.db.Where(&Metrics{SessionID: sID, Name: name}).Delete(&Metrics{}); m.Error != nil {
		return m.Error
	}
	return nil
}

func (p *WorkloadDao) DeleteSession(sID uint) error {
	if m := p.db.Delete(&Session{ID: sID}); m.Error != nil {
		return m.Error
	}
	if m := p.db.Delete(&Workload{SessionID: sID}); m.Error != nil {
		return m.Error
	}
	if m := p.db.Delete(&Metrics{SessionID: sID}); m.Error != nil {
		return m.Error
	}
	return nil
}

func (p *WorkloadDao) GetMetrics(wid uint, limit int, metrics []string) (map[string][]Metrics, error) {
	rst := make(map[string][]Metrics, len(metrics))
	for _, m := range metrics {
		var res []Metrics
		if m := p.db.Where(&Metrics{Key: m, WID: wid}).Order("w_id").Limit(limit).Find(&res); m.Error != nil {
			return nil, m.Error
		}
	}
	return rst, nil
}

func (p *WorkloadDao) GetMetricsBySid(sid uint, name string, limit int, metrics []string) (map[string][]Metrics, error) {
	rst := make(map[string][]Metrics)
	for _, v := range metrics {
		var ms []Metrics
		m := p.db.Where(&Metrics{Key: v, Name: name}).Order("start").Limit(limit).Find(&ms)
		if m.Error != nil {
			return nil, m.Error
		}
		rst[v] = ms
	}
	return rst, nil
}

func (p *WorkloadDao) GetWorkloadNameAndVersion(sessionID uint) ([]Workload, error) {
	var workloads []Workload
	m := p.db.Distinct("name").Where(&Workload{SessionID: sessionID}).Find(&workloads)
	return workloads, m.Error
}

func (p *WorkloadDao) GetMetricsByLoads(wID uint) ([]Metrics, error) {
	var metrics []Metrics
	m := p.db.Where(&Metrics{WID: wID}).Find(&metrics)
	return metrics, m.Error
}

func getTarget(target string, record Record) float64 {
	for k, v := range record.Metrics {
		if k == target {
			return v
		}
	}
	return -1
}
func timeStampToTime(stamp string) time.Time {
	r, _ := strconv.ParseInt(stamp, 10, 64)
	return time.Unix(r, 0)
}
