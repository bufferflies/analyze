package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockSource struct {
}

func (m *mockSource) Source(metrics, start, end string) (data [][]float64, err error) {
	result := make([][]float64, 5)
	for i := range result {
		result[i] = []float64{1, 2, 3, 4}
	}
	return result, nil
}

func TestMean(t *testing.T) {
	as := assert.New(t)
	c := NewChecker(&mockSource{})
	r, err := c.Apply("1630381080", "1630386080", "store_available", "pd_scheduler_store_status{type='store_available'}", "max(mean(%s))")
	as.Nil(err)
	as.Equal(float64(4), r)
}
