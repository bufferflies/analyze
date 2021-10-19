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
package server

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	url2 "net/url"
	"strings"
)

type Proxy struct {
	client *http.Client
}

func NewProxy() *Proxy {
	return &Proxy{
		client: &http.Client{},
	}
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	uri := strings.Trim(req.RequestURI, "/proxy")
	target := fmt.Sprintf("%s/%s", req.Header.Get("target"), uri)
	url, err := url2.Parse(target)
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}

	req.URL = url
	req.RequestURI = req.URL.Path

	t := req.Header.Get("Target")
	url, err = url2.Parse(t)
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.ServeHTTP(w, req)

}
