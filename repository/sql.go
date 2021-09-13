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

import "time"

type Bench struct {
	ID      uint   `gorm:"AUTO_INCREMENT"`
	BenchID string `gorm:"size:10"`
	Start   time.Time
	End     time.Time
}

func (Bench) TableName() string {
	return "bench"
}

type Workload struct {
	ID     uint `gorm:"AUTO_INCREMENT"`
	BenID  uint
	Name   string
	Start  time.Time
	End    time.Time
	Config string
	Cmd    string
}

func (Workload) TableName() string {
	return "workload"
}

type Metrics struct {
	ID    uint `gorm:"AUTO_INCREMENT"`
	WID   uint
	Key   string
	Value float64
}

func (Metrics) TableName() string {
	return "metrics"
}
