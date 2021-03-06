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

	"gonum.org/v1/gonum/stat"

	"github.com/bufferflies/pd-analyze/core"

	"github.com/bufferflies/pd-analyze/repository"
	"github.com/spf13/cobra"
)

var (
	dialClient = &http.Client{}
	metrics    = map[string]string{
		// tikv metrics
		"tikv_cpu": "sum(rate(tikv_thread_cpu_seconds_total{}[1m])) by (instance)",

		// pd metrics
		"store_write_rate_bytes": "pd_scheduler_store_status{ type=\"store_write_rate_bytes\"}",
		"store_write_rate_keys":  "pd_scheduler_store_status{type=\"store_write_rate_keys\"}",
		"store_write_query":      "pd_scheduler_store_status{type=\"store_write_query_rate\"}",

		"store_read_rate_bytes": "pd_scheduler_store_status{ type=\"store_read_rate_bytes\"}",
		"store_read_rate_keys":  "pd_scheduler_store_status{type=\"store_read_rate_keys\"}",
		"store_read_query":      "pd_scheduler_store_status{type=\"store_read_query_rate\"}",

		//tidb
		"tidb_duration_P999": "histogram_quantile(0.999, sum(rate(tidb_server_handle_query_duration_seconds_bucket{}[1m])) by (le))*1000",
		"tidb_duration_P99":  "histogram_quantile(0.99, sum(rate(tidb_server_handle_query_duration_seconds_bucket{}[1m])) by (le))*1000",
		"tidb_duration_P95":  "histogram_quantile(0.95, sum(rate(tidb_server_handle_query_duration_seconds_bucket{}[1m])) by (le))*1000",
		"tidb_duration_P80":  "histogram_quantile(0.80, sum(rate(tidb_server_handle_query_duration_seconds_bucket{}[1m])) by (le))*1000",

		"tidb_command_per_second": "sum(rate(tidb_server_query_total{}[1m])) by (result)",
	}
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
		d, err := checker.Apply(records.Start, records.End, name, metrics, fmt.Sprintf("mean(%s)", name))
		if err != nil {
			return err
		}
		data := d.([]float64)

		mean, std := stat.MeanStdDev(data, nil)
		if strings.HasPrefix(name, "tidb") {
			records.Metrics[name] = mean
		} else {
			records.Metrics[strings.Join([]string{name, "avg"}, "_")] = mean
			records.Metrics[strings.Join([]string{name, "std"}, "_")] = std
			if mean != 0 {
				records.Metrics[strings.Join([]string{name, "std/avg"}, "_")] = std / mean
			}
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
