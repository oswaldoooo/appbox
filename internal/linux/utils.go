package linux

import (
	"context"
	"io"
	"os"
	"os/exec"
	"syscall"
)

type Executor struct {
	Err error
	*exec.Cmd
}

var Stdout, Stderr = os.Stdout, os.Stderr

func Execute(ctx context.Context, name string, args ...string) *Executor {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = Stdout
	cmd.Stderr = Stderr
	return &Executor{Cmd: cmd}
}
func (e *Executor) Run() *Executor {
	e.Err = e.Cmd.Run()
	return e
}
func (e *Executor) Start() *Executor {
	e.Err = e.Cmd.Start()
	return e
}
func (e *Executor) Wait() *Executor {
	e.Err = e.Cmd.Wait()
	return e
}
func (e *Executor) SetIO(stdout, stderr io.Writer) *Executor {
	if stdout != nil {
		e.Stdout = stdout
	}
	if stderr != nil {
		e.Stderr = stderr
	}
	return e
}
func (e *Executor) SetStdin(stdin io.Reader) *Executor {
	if e != nil {
		e.Stdin = stdin
	}
	return e
}

func Mknod(rpath string, mode uint32) error {
	return syscall.Mknod(rpath, mode, syscall.S_IFCHR)
}
