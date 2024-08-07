package unix

//#include <unistd.h>
import "C"
import (
	"errors"
	"fmt"
	"os"
)

type Hook struct {
	Err   error
	is_do bool
	data  any
}

func (h *Hook) Then(f func(any), a any) *Hook {
	if h.is_do && h.Err == nil {
		f(a)
	}
	return h
}
func (h *Hook) End(code int) error {
	if h.is_do {
		if h.Err != nil {
			fmt.Fprintln(os.Stderr, h.Err)
		} else {
			code = 0
		}
		os.Exit(code)
	}
	return h.Err
}
func Fork() *Hook {
	var hook Hook
	_pid := int(C.fork())
	hook.data = _pid
	if _pid != 0 {
		println("pid", _pid)
		hook.is_do = true
		if _pid < 0 {
			hook.Err = errors.New("fork error")
		}
		return &hook
	}
	return &hook
}
