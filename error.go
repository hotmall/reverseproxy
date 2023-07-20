package reverseproxy

import (
	"fmt"
	"sync"
)

type Error struct {
	Code    int
	Message string
}

func (e *Error) Reset() {
	e.Code = 0
	e.Message = ""
}

func (e Error) Error() string {
	return fmt.Sprintf("%d %s", e.Code, e.Message)
}

var (
	errorPool sync.Pool
)

func acquireError() *Error {
	e := errorPool.Get()
	if e == nil {
		return &Error{}
	}
	return e.(*Error)
}

func releaseError(e *Error) {
	e.Reset()
	errorPool.Put(e)
}
