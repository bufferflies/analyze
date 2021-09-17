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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSave(t *testing.T) {
	te := assert.New(t)
	manager, err := NewMysqlManager("172.16.4.3:25831", "test1")
	te.Nil(err)
	bench := &Bench{
		BenchID: "test",
		Start:   time.Now(),
		End:     time.Now().Add(time.Minute),
	}
	manager.SaveBench(bench)
	fmt.Sprintf("id:%d", bench.ID)

}

func TestTimeStamp(t *testing.T) {
	te := assert.New(t)
	s := "1631184147"
	r, err := strconv.ParseInt(s, 10, 64)
	te.Nil(err)
	//format := "2006-01-01 12:33:36"
	fmt.Println(time.Unix(r, 0))
}
