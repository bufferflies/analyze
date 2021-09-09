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

import (
	"flag"
)

type Config struct {
	ListenAddress     string `json:"listen_address" toml:"listen_port"`
	PrometheusAddress string `json:"prometheus_address" toml:"prometheus_address"`
	StorageDir        string `json:"storage_dir" toml:"storage_dir"`
	flagSet           *flag.FlagSet
}
