package core

import (
	"regexp"
	"strings"

	"github.com/bufferflies/pd-analyze/errs"

	"github.com/Knetic/govaluate"
)

type Operator string

type Checker struct {
	source Source
	re     *regexp.Regexp
}

func NewChecker(source Source) *Checker {
	r, _ := regexp.Compile("\\(([a-zA-Z0-9_,={}']+?)\\)")
	return &Checker{
		source: source,
		re:     r,
	}
}

//func NewChecker(address string) *Checker {
//	s := NewPrometheus(address)
//	r, _ := regexp.Compile("\\(([a-zA-Z0-9_,={}']+?)\\)")
//	return &Checker{
//		source: &s,
//		re:     r,
//	}
//}

var ExpressionMap = make(map[string]govaluate.ExpressionFunction)

func RegisterFunction(name string, ex govaluate.ExpressionFunction) {
	if _, ok := ExpressionMap[name]; ok {
		return
	}
	ExpressionMap[name] = ex
}

func (c *Checker) Apply(start, end string, name, metrics, cmd string) (v interface{}, err error) {
	parameters := make(map[string]interface{}, 2)
	data, err := c.source.Source(metrics, start, end)
	if err != nil {
		return false, err
	}
	parameters[name] = data

	expression, err := govaluate.NewEvaluableExpressionWithFunctions(cmd, ExpressionMap)
	if err != nil {
		return nil, err
	}

	result, err := expression.Evaluate(parameters)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Checker) extractMetrics(cmd string) ([]string, error) {
	s := c.re.FindString(cmd)
	if len(s) == 0 {
		return nil, errs.Argument_Not_Match
	}
	metrics := strings.Split(s[1:len(s)-1], ",")
	return metrics, nil
}
