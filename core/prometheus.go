package core

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/bufferflies/pd-analyze/errs"
)

// prometheus api prefix
var (
	prefix   = "/api/v1/query_range"
	duration = -2 * time.Minute
	step     = "30s"
)

const (
	fail = "error"
)

type Prometheus struct {
	Address string
	Step    string
	client  http.Client
}

func NewPrometheus(address string) *Prometheus {
	client := http.Client{}
	return &Prometheus{
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
func (p *Prometheus) Get(metrics, _, end string) (values *PrometheusData, err error) {
	req, err := http.NewRequest(http.MethodGet, p.Address+prefix, nil)
	if err != nil {
		return nil, err
	}

	start, err := addDuration(end, duration)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Add("query", metrics)
	q.Add("start", start)
	q.Add("end", end)
	q.Add("step", step)
	req.URL.RawQuery = q.Encode()

	rsp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	if rsp.StatusCode != http.StatusOK {
		return nil, errs.Result_Not_Match
	}
	var data PrometheusResponse
	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(body, &data); err != nil {
		return
	}
	if data.Status == fail {
		return
	}
	return &data.Data, nil
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

func addDuration(timestamp string, duration time.Duration) (string, error) {
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return "", err
	}
	t := time.Unix(ts, 0)
	then := t.Add(duration)
	return strconv.FormatInt(then.Unix(), 10), err
}
