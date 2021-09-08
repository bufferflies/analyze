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
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func mockRecord() []Record {
	metrics := make(map[string]map[string]Index)
	index := make(map[string]Index)
	index["avg"] = Index{
		Data: []float64{1, 2, 3},
		Max:  1,
		Min:  2,
		Mean: 3,
		Std:  0,
	}
	metrics["test"] = index
	record := Record{Cmd: "sysbench", Start: "111", End: "222", Metrics: metrics}
	return []Record{record}

}

func TestSaveAndGet(t *testing.T) {
	te := assert.New(t)
	s, err := NewSqlite("./test.db")
	defer os.Remove("./test.db")
	te.Nil(err)
	r := mockRecord()
	err = s.Save("111", r)
	te.Nil(err)
	r1, err := s.Get("111")
	te.Nil(err)
	te.NotNil(r1)
	te.Equal(r, r1)

}
