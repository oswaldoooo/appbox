package boxaction_test

import (
	"os"
	"strconv"
	"syscall"
	"testing"
	"time"

	"github.com/oswaldoooo/app/boxaction"
	"github.com/oswaldoooo/app/internal/linux"
)

func TestSetns(t *testing.T) {
	pid := strconv.Itoa(os.Getpid())
	println("pid", pid)
	fd, err := syscall.Open("/proc/"+pid+"/ns/net", syscall.O_RDONLY, 0644)
	if fd < 0 {
		t.Fatal("open fd error", err)
	}
	defer syscall.Close(fd)
	status, _, err := syscall.Syscall(syscall.SYS_UNSHARE, syscall.CLONE_NEWNET|syscall.CLONE_FS, 0, 0)
	if int(status) < 0 {
		t.Log("setns error", err)
		return
	}
	time.Sleep(time.Second * 100)
}

func TestBuildBox(t *testing.T) {
	//todo: unshare mount-proc 未找到如何在代码中实现
	err := boxaction.BuildBoxAction("123456")
	if err != nil {
		t.Fatal("build app error", err)
	}
	t.Log("pid", os.Getpid())
	time.Sleep(time.Second * 100)
}
func TestUnshare(t *testing.T) {
	err := linux.Unshare(syscall.CLONE_NEWNET|syscall.CLONE_NEWUTS, "/app")
	if err != nil {
		t.Fatal("unshare error", err)
	}
	println("pid", os.Getpid())
	time.Sleep(time.Second * 100)
}
