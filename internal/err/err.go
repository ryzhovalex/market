package err

import "fmt"

type ErrData struct {
	msg  string
	code string
}

type E interface {
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
		code = "err"
	}
	return &ErrData{msg, code}
}

func FromBase(e error) E {
	return New(e.Error(), "err")
}

func Unwrap(err error) {
	if err != nil {
		panic(err)
	}
}
