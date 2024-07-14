package linux_test

import (
	"context"
	"os"
	"os/exec"
	"path"
	"syscall"
	"testing"
	"time"

	"github.com/oswaldoooo/app/internal/linux"
)

const (
	rwflags = syscall.MS_PRIVATE | syscall.MS_REC
)

var flags = []uintptr{
	syscall.MS_ACTIVE,
	syscall.MS_ASYNC,
	syscall.MS_BIND,
	syscall.MS_DIRSYNC,
	syscall.MS_INVALIDATE,
	syscall.MS_I_VERSION,
	syscall.MS_KERNMOUNT,
	syscall.MS_MANDLOCK,
	syscall.MS_MGC_MSK,
	syscall.MS_MGC_VAL,
	syscall.MS_MOVE,
	syscall.MS_NOATIME,
	syscall.MS_NODEV,
	syscall.MS_NODIRATIME,
	syscall.MS_NOEXEC,
	syscall.MS_NOSUID,
}

func TestProc(t *testing.T) {
	err := syscall.Unshare(syscall.CLONE_NEWNS)
	if err != nil {
		t.Fatal("unshare error", err)
	}
	err = syscall.Mount("none", "/", "", rwflags, "")
	if err != nil {
		t.Fatal("mount error", err)
	}
	t.Log("start success", os.Getpid())
	time.Sleep(time.Second * 100)
}

func TestMountBind(t *testing.T) {
	err := linux.Unshare(syscall.CLONE_NEWNS, "").End(1)
	if err != nil {
		t.Fatal("unshare error", err)
	}
	t.Log("pid", os.Getpid())
	cmd := exec.CommandContext(context.Background(), "mount", "--bind", "-r", "/app", "/myapp")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		t.Fatal("mount bind error", err)
	}
	//test read only mount bind
	// err = syscall.Mount("/app", "/myapp", "", syscall.MS_BIND|syscall.MS_REMOUNT|syscall.MS_RDONLY, "")
	// if err != nil {
	// 	t.Fatal("mount bind as read only error", err)
	// }
	time.Sleep(time.Second * 100)

}

func TestMountBindPrivate(t *testing.T) {
	err := linux.Unshare(syscall.CLONE_NEWNS, "").End(1)
	if err != nil {
		t.Fatal("unshare error", err)
	}
	err = syscall.Mount("tmpfs", "/app", "tmpfs", 0, "")
	if err != nil {
		t.Fatal(err)
	}
	err = mount_bind_sys_dir(true, "/lib")
	if err != nil {
		t.Fatal(err)
	}
	err = mount_bind_sys_dir(true, "/bin")
	if err != nil {
		t.Fatal(err)
	}
	err = os.Mkdir("/app/proc", 0555)
	if err != nil {
		t.Fatal(err)
	}
	err = syscall.Mount("none", "/app/proc", "proc", 0, "")
	if err != nil {
		t.Fatal(err)
	}
	err = os.Mkdir("/app/oldroot.app", 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = syscall.PivotRoot("/app", "/app/oldroot.app")
	if err != nil {
		t.Fatal("pivot root error", err)
	}
	err = syscall.Unmount("/oldroot.app", syscall.MNT_DETACH)
	if err != nil {
		t.Fatal("unmount oldroot error", err)
	}
	t.Log("pid", os.Getpid())
	time.Sleep(time.Second * 100)

}
func mount_bind_sys_dir(rdonly bool, name string) error {
	info, err := os.Stat(name)
	if err != nil {
		return err
	}
	to := path.Join("/app", name)
	err = os.MkdirAll(to, info.Mode())
	if err != nil {
		return err
	}
	_flag := 0
	if rdonly {
		_flag |= syscall.MS_RDONLY
	}
	return linux.MountBind(_flag, name, to)
}

func TestSetNs(t *testing.T) {
	err := linux.SetNsWithFile("/proc/86674/ns/net", 0)
	if err != nil {
		t.Fatal(err)
	}
	err = linux.Execute(context.Background(), "ip", "a").Run().Err
	if err != nil {
		t.Fatal(err)
	}
}

func TestNsExec(t *testing.T) {
	linux.NsExec(linux.CLONE_NET|linux.CLONE_NS, "6350").Then(func(h *linux.Hook, a any) {
		os.Chdir("/")
		f, err := os.OpenFile("./test.txt", os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return
		}
		cmd := linux.Execute(context.Background(), "ip", "a")
		cmd.Stdout = f
		cmd.Stderr = f
		cmd.Run()

	}, nil).End(1)
}
