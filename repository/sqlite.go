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
	"database/sql"
	"encoding/json"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type SqlLit struct {
	db *sql.DB
}

func NewSqlite(path string) (*SqlLit, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	sql := `create table if not exists bench(id varchar not null, create_time  varchar(32));
			create table if not exists  record(id varchar not null, benchID varchar not null,start_time string, end_time varchar,data varchar );
	`
	_, err = db.Exec(sql)
	if err != nil {
		return nil, err
	}
	return &SqlLit{
		db: db,
	}, nil

}

func (sqlite *SqlLit) Save(id string, records []Record) error {
	stmt, err := sqlite.db.Prepare("insert into bench(id,create_time) values (?,?)")
	if err != nil {
		return err
	}
	if _, err = stmt.Exec(id, time.Now().String()); err != nil {
		return err
	}
	sql := `insert into record(id,benchID,start_time,end_time,data) values `
	vals := make([]interface{}, 0, len(records)*4)
	for _, v := range records {
		sql += "(?,?,?,?,?),"
		date, err := json.Marshal(v.Metrics)
		if err != nil {
			return err
		}
		vals = append(vals, v.Cmd, id, v.Start, v.End, string(date))
	}
	sql = sql[0 : len(sql)-1]
	stmt, err = sqlite.db.Prepare(sql)
	if err != nil {
		return err
	}
	if _, err = stmt.Exec(vals...); err != nil {
		return err
	}
	return nil
}

func (sqlLit *SqlLit) Get(id string) (records []Record, err error) {
	sql := "select id,start_time,end_time,data from record where benchID=?"
	stmt, err := sqlLit.db.Prepare(sql)
	if err != nil {
		return nil, err
	}
	r, err := stmt.Query(id)
	if err != nil {
		return nil, err
	}
	for r.Next() {
		var data string
		record := Record{}
		if err = r.Scan(&record.Cmd, &record.Start, &record.End, &data); err != nil {
			return nil, err
		}
		if err = json.Unmarshal([]byte(data), &record.Metrics); err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, err
}
