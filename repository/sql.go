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
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewMysqlManager(address string, database string) (*gorm.DB, error) {
	dsn := fmt.Sprintf("root@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", address, database)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	return db, err
}

type Workload struct {
	ID           uint `gorm:"AUTO_INCREMENT"`
	SessionID    uint
	Name         string
	Start        time.Time
	End          time.Time
	Config       string
	Cmd          string
	TargetObject float64
	BenchName    string
	Version      string
}

func (Workload) TableName() string {
	return "workload"
}

type Metrics struct {
	ID        uint `gorm:"AUTO_INCREMENT"`
	WID       uint
	Name      string
	Key       string
	Value     float64
	Start     time.Time
	SessionID uint
}

func (Metrics) TableName() string {
	return "metrics"
}

type Project struct {
	ID          uint `gorm:"AUTO_INCREMENT"`
	Name        string
	Description string
}

func (Project) TableName() string {
	return "project"
}

type Session struct {
	ID                   uint   `json:"id" gorm:"AUTO_INCREMENT"`
	PID                  uint   `json:"pid"`
	Name                 string `json:"name"`
	Description          string `json:"descript"`
	Object               string `json:"object"`
	TargetObject         string `json:"target_object"`
	PdAddress            string `json:"pd_address"`
	TidbAddress          string `json:"tidb_address"`
	PromAddress          string `json:"prom_address"`
	GrafanaAddress       string `json:"grafana_address"`
	GrafanaAuthorization string `json:"grafana_authorization"`
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

func (Session) TableName() string {
	return "session"
}
