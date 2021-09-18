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
	"net/http"
	"os"

	"github.com/bufferflies/pd-analyze/repository"
	"github.com/spf13/cobra"
)

var (
	dialClient = &http.Client{}
)

// NewConfigCommand return a config subcommand of rootCmd
func NewReportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report",
		Short: "report record to analyze",
		Run:   Report,
	}
	cmd.Flags().StringP("server", "s", "http://localhost:8080", "analyze server address")
	cmd.Flags().StringP("id", "i", "test", "bench id")
	cmd.Flags().StringP("path", "p", "time.log", "record log path")
	return cmd
}

func Report(cmd *cobra.Command, args []string) {
	path, err := cmd.Flags().GetString("path")
	if err != nil {
		cmd.Printf("path can not nil:%v", err)
		return
	}
	records, err := ReadFile(path)
	if err != nil {
		cmd.Printf("read file fialed err:%v", err)
		return
	}

	address, err := cmd.Flags().GetString("server")
	if err != nil {
		cmd.Printf("get analyze address failed err:%v", err)
		return
	}
	id, err := cmd.Flags().GetString("id")
	if err != nil {
		cmd.Printf("get bench id  err:%v", err)
		return
	}
	url := address + "/tools/" + id
	body, err := json.Marshal(records)
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
