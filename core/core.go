package core

type Source interface {
	Source(metrics, start, end string) (data [][]float64, err error)
}

type Parser interface {
	Apply(start, end string, name, metrics, cmd string) (v interface{}, err error)
}
