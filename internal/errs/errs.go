package errs

import "fmt"

var CODE_ERR string = "err"
var CODE_NOT_FOUND string = "not_found_err"

type ErrData struct {
	msg  string
	code string
}

type Err interface {
	Error() string
	Msg() string
	Code() string
}

func (e *ErrData) Error() string {
	return fmt.Sprintf("%s:: %s", e.code, e.msg)
}

func (e *ErrData) Msg() string {
	return e.msg
}

func (e *ErrData) Code() string {
	return e.code
}

func New(msg string, code string) *ErrData {
	if code == "" {
		code = CODE_ERR
	}
	return &ErrData{msg, code}
}

func FromBase(e error) Err {
	return New(e.Error(), CODE_ERR)
}

func Unwrap(err error) {
	if err != nil {
		panic(err)
	}
}
