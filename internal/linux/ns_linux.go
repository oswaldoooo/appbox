package linux

//#include <unistd.h>
import "C"
import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
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
		}
		os.Exit(code)
	}
	return h.Err
}
func SetNsWithFile(file string, flags int) error {
	fd, err := syscall.Open(file, syscall.O_RDONLY, 0)
	if fd < 0 {
		return err
	}
	defer syscall.Close(fd)
	ok, _, err := syscall.Syscall(syscall.SYS_SETNS, uintptr(fd), uintptr(flags), 0)
	if int(ok) < 0 {
		return err
	}
	return nil
}

func UnMountProc(rpath string) error {
	return syscall.Unmount(rpath, 0)
}

var (
	Parent_NETNS_ID = "1"
	Parent_MNTNS_ID = "1"
	Parent_PIDNS_ID = "1"
	Parent_UTSNS_ID = "1"
)

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
			} else if pid_ > 0 {
				err = SetNsWithFile("/proc/"+Parent_NETNS_ID+"/ns/net", 0)
				if err != nil {
					fmt.Fprintln(os.Stderr, "set parent ns error", err)
				}
				err = SetNsWithFile("/proc/"+Parent_MNTNS_ID+"/ns/mnt", 0)
				if err != nil {
					fmt.Fprintln(os.Stderr, "set parent ns error", err)
				}
				if flags&syscall.CLONE_NEWNET > 0 {
					//config new ns to ip netns
					pidstr := strconv.Itoa(int(pid_))
					err = os.Symlink("/proc/"+pidstr+"/ns/net", "/var/run/netns/"+pidstr)
					if err != nil {
						hook.Err = errors.New("link network to ip netns error" + err.Error())
					}
				}
			}
			return &hook
		}
	}
	var hook = Hook{is_do: false}
	f, err := os.OpenFile("./appbox-error.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err == nil {
		os.Stderr = f
		os.Stdout = f
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

// func ExecuteWithNs(nsfile string, name string, args ...string) error {

// 	e := Execute(context.Background(), name, args...)

// }
