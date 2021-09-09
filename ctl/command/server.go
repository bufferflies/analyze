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
	"net/http"

	config2 "github.com/bufferflies/pd-analyze/config"
	"github.com/bufferflies/pd-analyze/server"
	"github.com/spf13/cobra"
)

// NewServerCommand return a server subcommand of rootCmd
func NewServerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "start analyze ",
		Run:   RunServer,
	}
	cmd.PersistentFlags().StringP("listen_address", "a", "localhost:8080", "analyze listen address")
	cmd.PersistentFlags().StringP("prometheus_address", "p", "http://172.16.4.3:22815/", "address of prometheus")
	cmd.PersistentFlags().StringP("storage_path", "s", "./test.db", "storage of path")
	return cmd
}

func RunServer(cmd *cobra.Command, args []string) {
	config := GetConfig(cmd)
	server := server.NewServer(config)
	router := server.CreateRoute()
	cmd.Printf("server start %v", config)
	if err := http.ListenAndServe(config.ListenAddress, router); err != nil {
		cmd.Print("server run failed ")
	}
}

func GetConfig(cmd *cobra.Command) *config2.Config {
	var config config2.Config
	addr, err := cmd.Flags().GetString("listen_address")
	if err != nil {
		cmd.Printf("address failed, err:%v", err)
	}
	config.ListenAddress = addr
	if config.StorageDir, err = cmd.Flags().GetString("storage_path"); err != nil {
		cmd.Printf("address failed, err:%v", err)
	}
	if config.PrometheusAddress, err = cmd.Flags().GetString("prometheus_address"); err != nil {
		cmd.Printf("address failed, err:%v", err)
	}
	return &config
}
