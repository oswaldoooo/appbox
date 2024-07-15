package linux

//#define _GNU_SOURCE
//#include <sched.h>
//#include <unistd.h>
import "C"
import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"syscall"
)

type Hook struct {
	Err   error
	is_do bool //if true do hook function
	data  any
}

func (h *Hook) Data() any {
	return h.data
}
func (h *Hook) Then(do func(*Hook, any), arg any) *Hook {
	if h.Err == nil && h.is_do {
		do(h, arg)
	}
	return h
}
func (h *Hook) End(code int) error {
	if h.is_do {
		if h.Err != nil {
			fmt.Fprintln(os.Stderr, h.Err)
			os.Exit(code)
		}
		os.Exit(0)
	}
	return h.Err
}
func SetNsWithFile(file string, flags int) error {
	fd, err := syscall.Open(file, syscall.O_RDONLY, 0)
	if fd < 0 {
		return err
	}
	defer syscall.Close(fd)
	ok := int(C.setns(C.int(fd), C.int(flags)))
	// ok, _, err := syscall.Syscall(syscall.SYS_SETNS, uintptr(fd), uintptr(flags), 0)
	if int(ok) < 0 {
		err = errors.New("setns error")
		return err
	}
	return nil
}

func UnMountProc(rpath string) error {
	return syscall.Unmount(rpath, 0)
}

func Unshare(flags int, mount_proc string) *Hook {
	if len(mount_proc) > 0 && flags&syscall.CLONE_NEWNS == 0 {
		flags |= syscall.CLONE_NEWNS
	}
	err := syscall.Unshare(flags)
	if err != nil {
		return &Hook{Err: errors.New("unshare error " + err.Error()), is_do: true}
	}
	if flags&syscall.CLONE_NEWPID > 0 {
		pid_ := C.fork()
		if pid_ != 0 {
			var hook = Hook{is_do: true, data: int(pid_)}
			if int(pid_) < 0 {
				println("fork error")
				os.Exit(-1)
			}
			return &hook
		}
	}
	var hook = Hook{is_do: false}
	f, err := os.OpenFile("./appbox-error.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err == nil {
		os.Stderr = f
		os.Stdout = f
	} else {
		println("open appbox-error.log error ", err.Error())
	}
	if flags&syscall.CLONE_NEWNS > 0 {
		err = syscall.Mount("none", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")
		if err != nil {
			hook.Err = errors.New("mount root directory error " + err.Error())
			return &hook
		}
		if len(mount_proc) == 0 {
			return &hook
		}
		err = syscall.Mount("tmpfs", mount_proc, "tmpfs", 0, "")
		if err != nil {
			hook.Err = errors.New("mount dir " + mount_proc + " error " + err.Error())
			return &hook
		}
	}
	return &hook
}

func MountBind(flags int, src, dst string) error {
	var _flag = []string{
		"--bind",
	}
	if flags&syscall.MS_RDONLY > 0 {
		_flag = append(_flag, "-r")
	}
	_flag = append(_flag, src, dst)
	cmd := exec.Command("mount", _flag...)
	cmd.Stderr, cmd.Stdout = os.Stderr, os.Stdout
	return cmd.Run()
}

// change root to current namespace
func Chroot(rpath string) error {
	return _pivot_root(rpath)
}

func _pivot_root(rpath string) error {
	old_home_repo := path.Join(rpath, "/oldroot.app")
	err := os.Mkdir(old_home_repo, 0755)
	if err != nil {
		return err
	}
	err = syscall.PivotRoot(rpath, old_home_repo)
	if err != nil {
		return err
	}
	err = os.Chdir("/")
	if err != nil {
		return err
	}
	err = syscall.Unmount("/oldroot.app", syscall.MNT_DETACH)
	os.Remove("/oldroot.app")
	return err
}

const (
	CLONE_NS = 1 << iota
	CLONE_NET
	CLONE_PID
	CLONE_UTS
)

func NsExec(_flag int, pid string) *Hook {
	_pid := int(C.fork())
	if _pid != 0 {
		if _pid < 0 {
			return &Hook{Err: errors.New("fork error")}
		}
		return &Hook{data: _pid}
	}
	var (
		err  error
		hook = Hook{is_do: true}
	)
	if _flag&CLONE_NS > 0 {
		err = SetNsWithFile("/proc/"+pid+"/ns/mnt", 0)
		hook.Err = err
		return &hook
	}
	if _flag&CLONE_NET > 0 {
		err = SetNsWithFile("/proc/"+pid+"/ns/net", 0)
		hook.Err = err
		return &hook
	}
	if _flag&CLONE_PID > 0 {
		err = SetNsWithFile("/proc/"+pid+"/ns/pid", 0)
		hook.Err = err
		return &hook
	}
	if _flag&CLONE_UTS > 0 {
		err = SetNsWithFile("/proc/"+pid+"/ns/uts", 0)
		hook.Err = err
		return &hook
	}
	return &hook
}

// func ExecuteWithNs(nsfile string, name string, args ...string) error {

// 	e := Execute(context.Background(), name, args...)

// }
