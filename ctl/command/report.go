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
package command

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/bufferflies/pd-analyze/core"

	"gonum.org/v1/gonum/floats"
	"gonum.org/v1/gonum/stat"

	"github.com/bufferflies/pd-analyze/repository"
	"github.com/spf13/cobra"
)

var (
	dialClient = &http.Client{}
	metrics    = map[string]string{
		"tikv_cpu":   "sum(rate(tikv_thread_cpu_seconds_total{}[1m])) by (instance)",
		"tikv_write": "sum(rate(tikv_engine_flow_bytes{ db=\"kv\", type=\"wal_file_bytes\"}[1m])) by (instance)",
		"tikv_read":  "sum(rate(tikv_engine_flow_bytes{ db=\"kv\", type=~\"bytes_read|iter_bytes_read\"}[1m])) by (instance)",
	}
	operators = map[string]string{"mean": "mean(%s)", "std": "std(%s)"}
)

type ReportConfig struct {
	prometheus string
	server     string
	sessionId  uint32
	name       string
	checker    *core.Checker
	records    []repository.Record
}

func newReportConfig(prometheus string) *ReportConfig {
	source := core.NewPrometheus(prometheus)
	checker := core.NewChecker(source)
	return &ReportConfig{
		prometheus: prometheus,
		checker:    checker,
	}
}

// NewConfigCommand return a config subcommand of rootCmd
func NewReportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report",
		Short: "report record to analyze",
		Run:   Report,
	}
	cmd.Flags().StringP("server", "s", "http://localhost:8080", "analyze server address")
	cmd.Flags().StringP("name", "n", "test", "bench name")
	cmd.Flags().Uint32P("session_id", "i", 1, "session id")
	cmd.Flags().StringP("data", "d", "time.log", "record log path")
	cmd.Flags().StringP("prometheus", "p", "localhost:9090", "prometheus address ")
	return cmd
}

func Report(cmd *cobra.Command, args []string) {
	prometheus, err := cmd.Flags().GetString("prometheus")
	if err != nil {
		cmd.Printf("get session id  err:%v", err)
		return
	}
	config := newReportConfig(prometheus)
	path, err := cmd.Flags().GetString("data")
	if err != nil {
		cmd.Printf("data can not nil:%v", err)
		return
	}
	if config.records, err = ReadFile(path); err != nil {
		cmd.Printf("read file failed err:%v", err)
		return
	}

	if config.server, err = cmd.Flags().GetString("server"); err != nil {
		cmd.Printf("get analyze address failed err:%v", err)
		return
	}

	if config.name, err = cmd.Flags().GetString("name"); err != nil {
		cmd.Printf("get bench name  err:%v", err)
		return
	}

	if config.sessionId, err = cmd.Flags().GetUint32("session_id"); err != nil {
		cmd.Printf("get session id  err:%v", err)
		return
	}

	url := fmt.Sprintf("%s/tools/%d/%s", config.server, config.sessionId, config.name)
	cmd.Println(url)
	for i := range config.records {
		config.records[i].Metrics = make(map[string]float64)
		if err := config.check(&config.records[i]); err != nil {
			cmd.Println(err.Error())
			return
		}
	}
	cmd.Println(config.records)
	body, err := json.Marshal(config.records)
	if err != nil {
		cmd.Printf("json marshal failed err:%v", err)
		return
	}
	rsp, err := dialClient.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		cmd.Print("request send failed err:%v", err)
		return
	}
	if rsp.StatusCode != http.StatusOK {
		cmd.Printf("response is not ok code:%d", rsp.StatusCode)
	}
	return

}

func (config *ReportConfig) check(records *repository.Record) error {
	source := core.NewPrometheus(config.prometheus)
	checker := core.NewChecker(source)
	records.Metrics = make(map[string]float64)
	for name, metrics := range metrics {
		for opName, m := range operators {
			prefix := strings.Join([]string{name, opName}, "_")
			d, err := checker.Apply(records.Start, records.End, name, metrics, fmt.Sprintf(m, name))
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

func ReadFile(path string) ([]repository.Record, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	result := make([]repository.Record, 0, 10)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Bytes()
		var r repository.Record
		if err := json.Unmarshal(line, &r); err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, nil
}
