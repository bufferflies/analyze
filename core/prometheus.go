package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

// prometheus api prefix
var prefix = "/api/v1/query_range"

// prometheus result status
type Status string

const (
	success Status = "success"
	fail           = "error"
)

type Prometheus struct {
	Address string
	Step    string
	client  http.Client
}

func NewPrometheus(address string) Prometheus {
	client := http.Client{}
	return Prometheus{
		Step:    "15s",
		client:  client,
		Address: address,
	}
}

type PrometheusResponse struct {
	Status string         `json:"status"`
	Data   PrometheusData `json:"data"`
}
type PrometheusData struct {
	ResultType string             `json:"resultType"`
	Result     []PrometheusResult `json:"result"`
}

type PrometheusResult struct {
	Metric map[string]string `json:"metric"`
	Values [][]interface{}   `json:"values"`
}

func (p *Prometheus) Source(metrics, start, end string) (data [][]float64, err error) {
	values, err := p.Get(metrics, start, end)
	if err != nil {
		return nil, err
	}
	data = values.ToArray()
	return data, nil
}

// Get returns values from prometheus
func (p *Prometheus) Get(metrics, start, end string) (values PrometheusData, err error) {
	req, err := http.NewRequest(http.MethodGet, p.Address+prefix, nil)
	if err != nil {
		fmt.Printf("request init err:%v", err)
		return
	}
	q := req.URL.Query()
	q.Add("query", metrics)
	q.Add("start", start)
	q.Add("end", end)
	q.Add("step", "30s")
	req.URL.RawQuery = q.Encode()

	rsp, err := p.client.Do(req)
	defer rsp.Body.Close()
	if err != nil {
		return
	}
	var data PrometheusResponse
	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return
	}
	if err = json.Unmarshal(body, &data); err != nil {
		return
	}
	if data.Status == fail {
		return
	}
	return data.Data, nil
}

// ToArray convert prometheus to array
func (values PrometheusData) ToArray() (stat [][]float64) {
	stat = make([][]float64, len(values.Result))
	for k, r := range values.Result {
		arr := make([]float64, len(r.Values))
		for i, v := range r.Values {
			arr[i], _ = strconv.ParseFloat(v[1].(string), 64)
		}
		stat[k] = arr
	}
	return
}
