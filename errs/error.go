package errs

import "errors"

var (
	Argument_Not_Match = errors.New("argument not match")
	Result_Not_Match   = errors.New("result not match")
)
