package core

import (
	"github.com/Knetic/govaluate"
	"github.com/bufferflies/pd-analyze/errs"
	"gonum.org/v1/gonum/floats"
	"gonum.org/v1/gonum/stat"
)

func init() {
	RegisterFunction("mean", Base(convertWeight(stat.Mean)))
	RegisterFunction("max", Base(floats.Max))
	RegisterFunction("min", Base(floats.Min))
	RegisterFunction("std", Base(convertWeight(stat.StdDev)))
}

type fn func(nums []float64) float64

type weightFn func(nums []float64, weight []float64) float64

func convertWeight(fn2 weightFn) fn {
	return func(nums []float64) float64 {
		return fn2(nums, nil)
	}
}

func Base(f fn) (ex govaluate.ExpressionFunction) {
	return func(args ...interface{}) (interface{}, error) {
		switch args[0].(type) {
		case [][]float64:
			values := args[0].([][]float64)
			max := make([]float64, len(values[0]))
			for j := range values[0] {
				series := make([]float64, len(values))
				for i := range values {
					series[i] = values[i][j]
				}
				max[j] = f(series)
			}
			return max, nil
		case []float64:
			values := args[0].([]float64)
			rst := f(values)
			return rst, nil
		default:
			return nil, errs.Argument_Not_Match
		}
	}
}
