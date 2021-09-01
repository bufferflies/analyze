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
package config

import "flag"

type Config struct {
	ListenPort        int    `json:"listen_port" toml:"listen_port"`
	PrometheusAddress string `json:"prometheus_address" toml:"prometheus_address"`
	flagSet           *flag.FlagSet
}

func NewConfig() *Config {
	cfg := &Config{}
	cfg.flagSet = flag.NewFlagSet("pd", flag.ContinueOnError)
	fs := cfg.flagSet
	fs.StringVar(&cfg.PrometheusAddress, "prometheus_address", "http://172.16.4.3:22815", "prometheus address")
	fs.IntVar(&cfg.ListenPort, "port", 8080, "listen port")
	return cfg
}
